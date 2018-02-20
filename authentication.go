// Copyright 2016 by Chris Palmer (https://noncombatant.org), and released under
// the terms of the GNU GPL3. See help.html for more information.

package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

const (
	hmacKeyLength             = 32
	hmacBasename              = "hmac.key"
	authenticationTokenLength = 32
)

var (
	anonymousFiles = []string{
		"bean-machine.css",
		"help.html",
		"manifest.json",
		"readme.html",
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

type AuthenticatingFileHandler struct {
	Root string
}

func (h AuthenticatingFileHandler) isRequestAuthenticated(r *http.Request) bool {
	cookie, e := r.Cookie("token")
	if e != nil {
		return false
	}

	parts := strings.Split(cookie.Value, ":")
	if len(parts) != 2 {
		log.Printf("Invalid cookie format")
		return false
	}
	username := parts[0]
	token := parts[1]

	decodedToken, e := hex.DecodeString(token)
	if e != nil {
		log.Printf("Invalid cookie format: %v", e)
		return false
	}

	if len(decodedToken) != authenticationTokenLength {
		log.Printf("Invalid cookie format: length %d", len(decodedToken))
		return false
	}

	return checkToken(username, decodedToken)
}

func (h AuthenticatingFileHandler) handleLogIn(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("name")
	password := r.FormValue("password")
	stored := readPasswordDatabase(path.Join(configurationPathname, passwordsBasename))

	if checkPassword(stored, username, password) {
		token := username + ":" + hex.EncodeToString(generateToken(username, stored[username]))
		cookie := &http.Cookie{Name: "token", Value: token, Secure: true, HttpOnly: true}
		http.SetCookie(w, cookie)
		http.Redirect(w, r, "/index.html", http.StatusFound)
	} else {
		cookie := &http.Cookie{Name: "token", Value: "", Secure: true, HttpOnly: true}
		http.SetCookie(w, cookie)
		h.redirectToLogin(w, r)
	}
}

func (h AuthenticatingFileHandler) redirectToLogin(w http.ResponseWriter, r *http.Request) {
	path := h.Root + "/login.html"
	file, e := os.Open(path)
	if e != nil {
		log.Fatalf("Could not open %q: %v", path, e)
	}
	defer file.Close()
	http.ServeContent(w, r, "login.html", time.Now(), file)
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
	path := h.normalizePathname(r.URL.Path)
	file, info := openFileIfPublic(path)
	if file == nil || info == nil {
		http.NotFound(w, r)
		return
	}
	defer file.Close()
	http.ServeContent(w, r, path, info.ModTime(), file)
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

	if !shouldServeFileToAnonymousClients(r.URL.Path) && !h.isRequestAuthenticated(r) {
		h.redirectToLogin(w, r)
		return
	}

	h.serveFile(w, r)
}
