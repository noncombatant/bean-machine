// Copyright 2016 by Chris Palmer (https://noncombatant.org), and released under
// the terms of the GNU GPL3. See help.html for more information.

package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

const (
	hmacKeyLength             = 32
	hmacBasename              = "hmac.key"
	authenticationTokenLength = 32
)

var (
	anonymousFiles = []string{
		"bean-machine.css",
		"favicon.ico",
		"help.html",
		"login.html",
		"manifest.json",
		"readme.html",
	}

	gzippableExtensions = map[string]bool{
		".css":  true,
		".html": true,
		".js":   true,
		".json": true,
		".svg":  true,
		".txt":  true,
	}
)

func generateAndSaveHmacKey(pathname string) {
	key := makeRandomBytes(hmacKeyLength)
	if e := ioutil.WriteFile(pathname, key, 0600); e != nil {
		log.Fatalf("Could not save HMAC key to %q: %v", pathname, e)
	}
	log.Print("Generated and saved new HMAC key")
}

func getHmacKey() []byte {
	pathname := path.Join(configurationPathname, hmacBasename)

	if _, e := os.Stat(pathname); os.IsNotExist(e) {
		generateAndSaveHmacKey(pathname)
	}

	key, e := ioutil.ReadFile(pathname)
	if e != nil {
		log.Fatalf("Could not open %q: %v", pathname, e)
	}

	if len(key) != hmacKeyLength {
		log.Fatalf("No valid key in %q: %v", pathname, e)
	}
	return key
}

// The token is the HMAC_SHA256 of the credential as stored in the password
// database (salt and scrypted password). This means that a token is valid as
// long as the HMAC key and the stored credential remain constant. No additional
// session state storage (beyond the password database and the HMAC key) is
// necessary.
//
// The security implications of this are:
//
//   * If the storedCredential changes (new scrypt parameters, new salt, new
//     password), existing sessions are invalidated.
//   * The token for all live sessions for the same (username, storedCredential)
//     pair is the same.
//   * Deducing the storedCredential from the token is as hard as 'reversing'
//     HMAC_SHA256 with a `hmacKeyLength`-length key.
//   * Users with with the same password will still get different token values,
//     both because the username is an input, and because each storedCredential
//     is created with a different random salt.
//
func generateToken(username string, storedCredential string) []byte {
	mac := hmac.New(sha256.New, getHmacKey())
	mac.Write([]byte(username))
	mac.Write([]byte("\x00"))
	mac.Write([]byte(storedCredential))
	return mac.Sum(nil)
}

func checkToken(username string, receivedToken []byte) bool {
	passwords := readPasswordDatabase(path.Join(configurationPathname, passwordsBasename))
	storedCredential, ok := passwords[username]
	if !ok {
		log.Printf("checkToken: No such username: %q", username)
		return false
	}

	expected := generateToken(username, storedCredential)
	return hmac.Equal(receivedToken, expected)
}

func splitCookie(cookie string) (string, []byte, error) {
	parts := strings.Split(cookie, ":")
	if len(parts) != 2 {
		return "", nil, errors.New(fmt.Sprintf("Could not parse cookie %q", cookie))
	}

	username := parts[0]
	token := parts[1]

	decodedToken, e := hex.DecodeString(token)
	if e != nil {
		return "", nil, errors.New(fmt.Sprintf("Invalid token %q (%v)", cookie, e))
	}

	if len(decodedToken) != authenticationTokenLength {
		return "", nil, errors.New(fmt.Sprintf("Invalid token length %q", cookie))
	}

	return username, decodedToken, nil
}

type AuthenticatingFileHandler struct {
	Root string
}

func (h AuthenticatingFileHandler) handleLogIn(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("name")
	password := r.FormValue("password")
	stored := readPasswordDatabase(path.Join(configurationPathname, passwordsBasename))

	if checkPassword(stored, username, password) {
		log.Printf("Successful authentication for user %q", username)
		token := username + ":" + hex.EncodeToString(generateToken(username, stored[username]))
		cookie := &http.Cookie{Name: "token", Value: token, Secure: true, HttpOnly: true}
		http.SetCookie(w, cookie)
		http.Redirect(w, r, "/index.html", http.StatusFound)
	} else {
		log.Printf("Failed authentication for user %q", username)
		cookie := &http.Cookie{Name: "token", Value: "", Secure: true, HttpOnly: true}
		http.SetCookie(w, cookie)
		redirectToLogin(w, r)
	}
}

func redirectToLogin(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/login.html", http.StatusFound)
}

func openFileIfPublic(pathname string) (*os.File, os.FileInfo) {
	file, e := os.Open(pathname)
	if e != nil {
		log.Printf("Couldn't open %q: %v", pathname, e)
		return nil, nil
	}

	info, e := file.Stat()
	if e != nil {
		log.Printf("Couldn't stat %q: %v", pathname, e)
		file.Close()
		return nil, nil
	}

	if info.Mode()&0004 == 0 {
		log.Printf("Couldn't open %q: not public", pathname)
		file.Close()
		return nil, nil
	}

	return file, info
}

func (h AuthenticatingFileHandler) serveFile(w http.ResponseWriter, r *http.Request) {
	pathname := h.normalizePathname(r.URL.Path)
	acceptsGzip := strings.Contains(r.Header.Get("Accept-Encoding"), "gzip")
	_, gzippable := gzippableExtensions[path.Ext(pathname)]

	var file *os.File
	var info os.FileInfo

	if gzippable && acceptsGzip {
		gzPathname := pathname + ".gz"
		gzFile, gzFileInfo := openFileIfPublic(gzPathname)
		if gzFile != nil && gzFileInfo != nil {
			file = gzFile
			info = gzFileInfo
			w.Header().Set("Content-Encoding", "gzip")
		} else {
			f, _ := openFileIfPublic(pathname)
			if f != nil {
				e := compressFile(gzPathname, f)
				defer f.Close()
				if e != nil {
					log.Printf("Could not create %q: %v", gzPathname, e)
				}
			}
			file, info = openFileIfPublic(gzPathname)
			if file != nil {
				w.Header().Set("Content-Encoding", "gzip")
			}
		}
	} else {
		file, info = openFileIfPublic(pathname)
	}

	if file == nil || info == nil {
		http.NotFound(w, r)
		return
	}
	defer file.Close()
	http.ServeContent(w, r, pathname, info.ModTime(), file)
}

func (h AuthenticatingFileHandler) normalizePathname(pathname string) string {
	if "/" == pathname {
		pathname = "/index.html"
	}
	pathname = filepath.Clean(h.Root + pathname)
	if !strings.HasPrefix(pathname, h.Root) {
		return h.Root + "/404.html"
	}
	return pathname
}

func shouldServeFileToAnonymousClients(pathname string) bool {
	return isStringInStrings(path.Base(pathname), anonymousFiles)
}

func (h AuthenticatingFileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/login.html" && r.Method == http.MethodPost {
		h.handleLogIn(w, r)
		return
	}

	var username string
	var decodedToken []byte

	cookie, e := r.Cookie("token")
	if e == nil {
		username, decodedToken, e = splitCookie(cookie.Value)
		if e != nil {
			log.Printf("Refusing %q to client with invalid cookie (%v)", r.URL.Path, e)
			redirectToLogin(w, r)
			return
		}
	}

	if !shouldServeFileToAnonymousClients(r.URL.Path) && !checkToken(username, decodedToken) {
		log.Printf("Refusing %q to %q with invalid token", r.URL.Path, username)
		redirectToLogin(w, r)
		return
	}

	log.Printf("Serving %q to %q", r.URL.Path, username)
	h.serveFile(w, r)
}
