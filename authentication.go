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
	"math/rand"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
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
		"index.css",
		"favicon.ico",
		"help.html",
		"login.html",
		"manifest.json",
		"readme.html",

		"clef-512.png",
		"favicon.ico",
		"help.png",
		"play.png",
		"pause.png",
		"repeat.png",
		"shuffle.png",
		"skip.png",
		"unknown-album.png",
	}

	gzippableExtensions = []string{
		".css",
		".html",
		".js",
		".json",
		".svg",
		".tsv",
		".txt",
	}

	coverExtensions = []string{
		".gif",
		".jpeg",
		".jpg",
		".png",
	}

	cookieLifetime, _ = time.ParseDuration("24000h")

	wordSplitter = regexp.MustCompile(`\s+`)
)

func writeString(w http.ResponseWriter, s string) (int, error) {
	return w.Write([]byte(s))
}

func writeItemInfos(w http.ResponseWriter, infos ItemInfos) {
	writeString(w, "[\n")
	for i, info := range infos {
		writeString(w, info.ToJSON())
		if i == len(infos)-1 {
			writeString(w, "\n")
		} else {
			writeString(w, ",\n")
		}
	}
	writeString(w, "]")
}

func getCookieLifetime() time.Time {
	return (time.Now()).Add(cookieLifetime)
}

func generateAndSaveHmacKey(pathname string) {
	key := MustMakeRandomBytes(hmacKeyLength)
	if e := ioutil.WriteFile(pathname, key, 0600); e != nil {
		Logger.Fatalf("Could not save HMAC key to %q: %v", pathname, e)
	}
}

func getHmacKey() []byte {
	pathname := path.Join(configurationPathname, hmacBasename)
	if _, e := os.Stat(pathname); os.IsNotExist(e) {
		generateAndSaveHmacKey(pathname)
	}

	key, e := ioutil.ReadFile(pathname)
	if e != nil {
		Logger.Fatalf("Could not open %q: %v", pathname, e)
	}

	if len(key) != hmacKeyLength {
		Logger.Fatalf("No valid key in %q: %v", pathname, e)
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
	mac.Write([]byte(strings.ToLower(username)))
	mac.Write([]byte("\x00"))
	mac.Write([]byte(storedCredential))
	return mac.Sum(nil)
}

func checkToken(username string, receivedToken []byte) bool {
	passwords := readPasswordDatabase(path.Join(configurationPathname, passwordsBasename))
	username = strings.ToLower(username)
	storedCredential, ok := passwords[username]
	if !ok {
		Logger.Printf("No such username %q", username)
		return false
	}

	expected := generateToken(username, storedCredential)
	return hmac.Equal(receivedToken, expected)
}

func parseCookie(cookie string) (string, []byte, error) {
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
	// TODO: Factor the username normalization in a minimal, robust way. In this
	// file and in passwords.go, we're just sprinkling `ToLower` everywhere.
	username := strings.ToLower(r.FormValue("name"))
	password := r.FormValue("password")
	stored := readPasswordDatabase(path.Join(configurationPathname, passwordsBasename))

	cookie := &http.Cookie{Name: "token", Value: "", Secure: true, HttpOnly: true, Expires: getCookieLifetime(), Path: "/"}
	if checkPassword(stored, username, password) {
		Logger.Printf("%q successful", username)
		token := username + ":" + hex.EncodeToString(generateToken(username, stored[username]))
		cookie.Value = token
		http.SetCookie(w, cookie)
		http.Redirect(w, r, "/index.html", http.StatusFound)
	} else {
		Logger.Printf("%q unsuccessful", username)
		http.SetCookie(w, cookie)
		redirectToLogin(w, r)
	}
}

func (h AuthenticatingFileHandler) handleSearch(w http.ResponseWriter, r *http.Request) {
	queries := r.URL.Query()["q"]
	if queries == nil || len(queries) == 0 {
		Logger.Print("Ignoring empty search.")
		return
	}

	// TODO: If empty query, show new entries for this year or month.
	query := strings.TrimSpace(queries[0])
	if len(query) == 0 || "?" == query {
		rand.Seed(time.Now().Unix())
		item := catalog.ItemInfos[rand.Intn(len(catalog.ItemInfos))]
		words := wordSplitter.Split(path.Dir(item.Pathname), -1)
		query = words[len(words)-1]
	}

	matches := matchItems(catalog.ItemInfos, query)
	w.Header().Set("Content-Type", "text/json")
	writeItemInfos(w, matches)
}

func redirectToLogin(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/login.html", http.StatusFound)
}

// TODO: This function is just too complex, since it combines gzipping and
// if-public. We need a transparent `MaybeGzippedFile` type, or similar.
//
// Returns an open File, a FileInfo, any error, and a bool indicating whether
// or not the file contains gzipped contents.
func openFileIfPublic(pathname string, shouldTryGzip bool) (*os.File, os.FileInfo, error, bool) {
	file, info, e := OpenFileAndInfo(pathname)
	if e != nil {
		return file, info, e, false
	}

	if !IsFileWorldReadable(info) {
		file.Close()
		Logger.Printf("NOTE: %q not world-readable", pathname)
		return nil, nil, errors.New(fmt.Sprintf("openFileIfPublic: %q not public", pathname)), false
	}

	if shouldTryGzip {
		gzPathname := pathname + ".gz"
		gzFile, gzInfo, e := OpenFileAndInfo(gzPathname)

		if IsFileWorldReadable(gzInfo) {
			// Handle the common case first.
			if e == nil && gzInfo.ModTime().After(info.ModTime()) {
				file.Close()
				return gzFile, gzInfo, e, true
			}

			// Clean up, remove the old gzPathname if it exists, create a new one,
			// and return it.
			if gzFile != nil {
				gzFile.Close()
			}

			e = os.Remove(gzPathname)
			if e == nil {
				e = GzipFile(gzPathname, file)
				if e != nil {
					Logger.Printf("Could not create new gz file: %q, %v", gzPathname, e)
					file.Seek(0, os.SEEK_SET)
					return file, info, e, false
				}

				gzFile, gzInfo, e = OpenFileAndInfo(gzPathname)
				if e == nil {
					file.Close()
					return gzFile, gzInfo, e, true
				} else {
					gzFile.Close()
				}
			} else {
				Logger.Printf("Could not remove %q: %v", gzPathname, e)
			}
		} else {
			Logger.Printf("NOTE: %q not world-readable", gzPathname)
		}
	}

	return file, info, e, false
}

func (h AuthenticatingFileHandler) serveFileContents(pathname string, w http.ResponseWriter, r *http.Request) {
	acceptsGzip := strings.Contains(r.Header.Get("Accept-Encoding"), "gzip")
	gzippable := IsStringInStrings(path.Ext(pathname), gzippableExtensions)

	file, info, e, isGzipped := openFileIfPublic(pathname, gzippable && acceptsGzip)
	if e != nil || file == nil || info == nil {
		Logger.Print(e)
		http.NotFound(w, r)
		return
	}
	defer file.Close()

	if isGzipped {
		w.Header().Set("Content-Encoding", "gzip")
	}

	// TODO: Do something to help the browser cache resources. ETag seems most likely?
	http.ServeContent(w, r, pathname, info.ModTime(), file)
	Logger.Printf("%q", pathname)
}

func (h AuthenticatingFileHandler) serveFile(w http.ResponseWriter, r *http.Request) {
	pathname := h.normalizePathname(r.URL.Path)
	if strings.HasSuffix(pathname, "/cover") {
		h.serveCover(pathname, w, r)
		return
	}

	h.serveFileContents(pathname, w, r)
}

func (h AuthenticatingFileHandler) serveCover(pathname string, w http.ResponseWriter, r *http.Request) {
	for _, extension := range coverExtensions {
		file, info, e, _ := openFileIfPublic(pathname+extension, false)
		// TODO: Reshape this!
		if file != nil {
			defer file.Close()
		}
		if e != nil || file == nil || info == nil {
			continue
		}

		// TODO: Unify this in serveFileContents.
		http.ServeContent(w, r, pathname, info.ModTime(), file)
		Logger.Printf("%q", pathname)
		return
	}

	http.Redirect(w, r, "/unknown-album.png", http.StatusFound)
}

func (h AuthenticatingFileHandler) normalizePathname(pathname string) string {
	if "/" == pathname {
		pathname = "/index.html"
	}
	pathname = path.Join(h.Root, filepath.Clean(pathname))
	if !strings.HasPrefix(pathname, h.Root) {
		return h.Root + "/404.html"
	}
	return pathname
}

func shouldServeFileToAnonymousClients(pathname string) bool {
	return IsStringInStrings(path.Base(pathname), anonymousFiles)
}

func (h AuthenticatingFileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/login.html" && r.Method == http.MethodPost {
		h.handleLogIn(w, r)
		return
	}

	if shouldServeFileToAnonymousClients(r.URL.Path) {
		h.serveFile(w, r)
		return
	}

	var username string
	var decodedToken []byte

	cookie, e := r.Cookie("token")
	if e == nil {
		username, decodedToken, e = parseCookie(cookie.Value)
		if e != nil {
			Logger.Printf("Refusing %q to client with invalid cookie (%v)", r.URL.Path, e)
			redirectToLogin(w, r)
			return
		}
	}

	if "" == username {
		redirectToLogin(w, r)
		return
	}

	if !checkToken(username, decodedToken) {
		Logger.Printf("Refusing %q to %q with invalid token", r.URL.Path, username)
		redirectToLogin(w, r)
		return
	}

	if r.URL.Path == "/search" {
		h.handleSearch(w, r)
		return
	}

	h.serveFile(w, r)
}
