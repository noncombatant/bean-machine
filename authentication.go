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

	cookieLifetime, _ = time.ParseDuration("2400h")
)

func getCookieLifetime() time.Time {
	return (time.Now()).Add(cookieLifetime)
}

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
		cookie := &http.Cookie{Name: "token", Value: token, Secure: true, HttpOnly: true, Expires: getCookieLifetime()}
		http.SetCookie(w, cookie)
		http.Redirect(w, r, "/index.html", http.StatusFound)
	} else {
		cookie := &http.Cookie{Name: "token", Value: "", Secure: true, HttpOnly: true}
		http.SetCookie(w, cookie)
		redirectToLogin(w, r)
	}
}

func redirectToLogin(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/login.html", http.StatusFound)
}

func openFileIfPublic(pathname string, shouldTryGzip bool) (FileAndInfoResult, bool) {
	nonGzResult := openFileAndGetInfo(pathname)
	if nonGzResult.Error != nil {
		log.Printf("Could not open %q: %v", pathname, nonGzResult.Error)
		return nonGzResult, false
	}

	if nonGzResult.Info.Mode()&0004 == 0 {
		nonGzResult.File.Close()
		return FileAndInfoResult{File: nil, Error: errors.New(fmt.Sprintf("Couldn't open %q: not public", pathname))}, false
	}

	if shouldTryGzip {
		gzPathname := pathname + ".gz"
		gzResult := openFileAndGetInfo(gzPathname)

		// Handle the common case first.
		if gzResult.Error == nil && gzResult.Info.ModTime().After(nonGzResult.Info.ModTime()) {
			nonGzResult.File.Close()
			return gzResult, true
		}

		// Clean up, remove the old gzPathname if it exists, create a new one, and
		// return it.
		if gzResult.File != nil {
			gzResult.File.Close()
		}
		_ = os.Remove(gzPathname)

		e := compressFile(gzPathname, nonGzResult.File)
		if e != nil {
			nonGzResult.File.Seek(0, os.SEEK_SET)
			return nonGzResult, false
		}

		gzResult = openFileAndGetInfo(gzPathname)
		if gzResult.Error == nil {
			return gzResult, true
		}
	}

	return nonGzResult, false
}

func (h AuthenticatingFileHandler) serveFile(w http.ResponseWriter, r *http.Request) {
	pathname := h.normalizePathname(r.URL.Path)
	acceptsGzip := strings.Contains(r.Header.Get("Accept-Encoding"), "gzip")
	_, gzippable := gzippableExtensions[path.Ext(pathname)]

	result, isGzipped := openFileIfPublic(pathname, gzippable && acceptsGzip)

	if result.Error != nil || result.File == nil || result.Info == nil {
		http.NotFound(w, r)
		return
	}

	if isGzipped {
		w.Header().Set("Content-Encoding", "gzip")
	}

	defer result.File.Close()
	http.ServeContent(w, r, pathname, result.Info.ModTime(), result.File)
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

	// TODO: Create /queryItems endpoint that lets clients get the metadata of 1
	// or more items by their unique ID numbers and the time of the client's last
	// view of the metadata. Should work by GET or POST (even though
	// non-state-changing on the server side, requests for many items may be too
	// large to fit into a URL query string). A call to /queryItems with no items
	// means "get all items' metadata".
	//
	// This will require the server-side data model to change: Have to keep track
	// of when the metadata last changed (most recent of: file mtime and parent
	// directory mtime). Also need to keep track of each item's unique ID, and
	// these must stay stable across re-builds of the catalog.
	//
	// This will enable the client to not need to request all of catalog.js every
	// time starting the app. The client can cache metadata in `localStorage`.
	// This should speed up loading (less data sent over the network) and parsing
	// (don't need to parse a huge blob of JS).
	//
	// Another, completely different idea, is to make catalog.js not JS(ON) but
	// some simpler format that is faster (?) to parse. (Records per line, fields
	// separated by tabs?) Such a format might also be marginally smaller.
	//
	// Additionally, the catalog request should be with the Fetch API, not a
	// blocking `<script src=...>` load.

	if shouldServeFileToAnonymousClients(r.URL.Path) {
		h.serveFile(w, r)
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

	if !checkToken(username, decodedToken) {
		log.Printf("Refusing %q to %q with invalid token", r.URL.Path, username)
		redirectToLogin(w, r)
		return
	}

	if cookie != nil {
		cookie.Expires = getCookieLifetime()
		http.SetCookie(w, cookie)
	}

	h.serveFile(w, r)
}
