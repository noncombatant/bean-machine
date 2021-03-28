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
		"favicon.ico",
		"help.html",
		"help.png",
		"index.css",
		"login.html",
		"manifest.json",
		"pause.png",
		"play.png",
		"readme.html",
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

type HTTPHandler struct {
	Root                  string
	ConfigurationPathname string
}

// Creates a gzipped version of the uncompressed file named by pathname, and
// returns an open File and FileInfo. On error, it logs the error and returns
// nil, nil.
//
// pathname, file, and info all refer to the uncompressed file. (We need
// pathname because `info.Name()` gives us only the basename.)
func createGzipped(pathname string, file *os.File, info os.FileInfo) (*os.File, os.FileInfo) {
	gzPathname := pathname + ".gz"
	// Remove any old one.
	os.Remove(gzPathname)
	e := GzipStream(gzPathname, file)
	if e != nil {
		Logger.Printf("Could not create gzipped file %q: %v", gzPathname, e)
		return nil, nil
	}

	gzFile, gzInfo, e := OpenFileAndInfo(gzPathname)
	if e != nil {
		Logger.Printf("Could not open just-created gzipped file %q: %v", gzPathname, e)
		return nil, nil
	}
	return gzFile, gzInfo
}

func getCookieLifetime() time.Time {
	return (time.Now()).Add(cookieLifetime)
}

func (h *HTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
	if e != nil {
		Logger.Printf("Refusing %q to client with missing cookie (%v)", r.URL.Path, e)
		redirectToLogin(w, r)
		return
	}

	username, decodedToken, e = parseCookie(cookie.Value)
	if e != nil {
		Logger.Printf("Refusing %q to client with invalid cookie (%v)", r.URL.Path, e)
		redirectToLogin(w, r)
		return
	}

	if username == "" {
		Logger.Printf("Refusing %q to client with blank username", r.URL.Path)
		redirectToLogin(w, r)
		return
	}

	if !h.checkToken(username, decodedToken) {
		Logger.Printf("Refusing %q to %q with invalid token", r.URL.Path, username)
		redirectToLogin(w, r)
		return
	}

	if r.URL.Path == "/search" {
		h.handleSearch(w, r)
		return
	} else if strings.HasSuffix(r.URL.Path, "/media.html") {
		pathname := path.Clean(h.Root + path.Dir(r.URL.Path))
		// In addition to `path.Clean`, handle this special case:
		if pathname == h.Root {
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte("Forbidden"))
			return
		}

		w.WriteHeader(http.StatusOK)
		page := buildMediaIndex(pathname)
		w.Write([]byte(page))
		return
	}

	h.serveFile(w, r)
}

func (h *HTTPHandler) checkToken(username string, receivedToken []byte) bool {
	passwords := readPasswordDatabase(path.Join(h.ConfigurationPathname, passwordsBasename))
	username = normalizeUsername(username)
	storedCredential, ok := passwords[username]
	if !ok {
		Logger.Printf("No such username %q", username)
		return false
	}

	expected := h.generateToken(username, storedCredential)
	return hmac.Equal(receivedToken, expected)
}

func (h *HTTPHandler) generateAndSaveHmacKey() {
	pathname := path.Join(h.ConfigurationPathname, hmacBasename)
	key := MustMakeRandomBytes(hmacKeyLength)
	if e := ioutil.WriteFile(pathname, key, 0600); e != nil {
		Logger.Fatalf("Could not save HMAC key to %q: %v", pathname, e)
	}
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
func (h *HTTPHandler) generateToken(username string, storedCredential string) []byte {
	mac := hmac.New(sha256.New, h.getHmacKey())
	mac.Write([]byte(normalizeUsername(username)))
	mac.Write([]byte("\x00"))
	mac.Write([]byte(storedCredential))
	return mac.Sum(nil)
}

func (h *HTTPHandler) getHmacKey() []byte {
	pathname := path.Join(h.ConfigurationPathname, hmacBasename)
	if _, e := os.Stat(pathname); os.IsNotExist(e) {
		h.generateAndSaveHmacKey()
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

func (h *HTTPHandler) handleLogIn(w http.ResponseWriter, r *http.Request) {
	username := normalizeUsername(r.FormValue("name"))
	password := r.FormValue("password")
	stored := readPasswordDatabase(path.Join(h.ConfigurationPathname, passwordsBasename))

	cookie := &http.Cookie{Name: "token", Value: "", Secure: true, HttpOnly: true, Expires: getCookieLifetime(), Path: "/"}
	if checkPassword(stored, username, password) {
		Logger.Printf("%q successful", username)
		token := username + ":" + hex.EncodeToString(h.generateToken(username, stored[username]))
		cookie.Value = token
		http.SetCookie(w, cookie)
		http.Redirect(w, r, "/index.html", http.StatusFound)
	} else {
		Logger.Printf("%q unsuccessful", username)
		http.SetCookie(w, cookie)
		redirectToLogin(w, r)
	}
}

func (h *HTTPHandler) handleSearch(w http.ResponseWriter, r *http.Request) {
	queries := r.URL.Query()["q"]
	if len(queries) == 0 {
		Logger.Print("Ignoring empty search.")
		return
	}

	query := strings.TrimSpace(queries[0])
	var matches ItemInfos
	if len(query) == 0 {
		year, month, _ := time.Now().Date()
		for i := 0; i < 6; i++ {
			if i > 0 && month == 1 {
				month = 12
				year -= 1
			}
			query = fmt.Sprintf("mtime:%04d-%02d-", year, int(month)-i)
			matches = matchItems(catalog.ItemInfos, query)
			if len(matches) > 0 {
				goto done
			}
		}

		// If we get here, there were no new items in the last 6 months. Just make
		// a random query, then.
		query = "?"
	}

	if query == "?" {
		rand.Seed(time.Now().Unix())
		item := catalog.ItemInfos[rand.Intn(len(catalog.ItemInfos))]
		words := wordSplitter.Split(path.Dir(item.Pathname), -1)
		query = words[len(words)-1]
	}

	matches = matchItems(catalog.ItemInfos, query)

done:
	w.Header().Set("Content-Type", "text/json")
	writeItemInfos(w, matches)
}

func (h *HTTPHandler) normalizePathname(pathname string) string {
	if pathname == "/" {
		pathname = "/index.html"
	}
	pathname = path.Join(h.Root, filepath.Clean(pathname))
	if !strings.HasPrefix(pathname, h.Root) {
		return h.Root + "/404.html"
	}
	return pathname
}

func (h *HTTPHandler) serveCover(pathname string, w http.ResponseWriter, r *http.Request) {
	for _, extension := range coverExtensions {
		file, info, e, _ := openFileIfPublic(pathname+extension, false)
		if e != nil {
			continue
		}
		defer file.Close()

		// TODO: Unify this in serveFileContents.
		http.ServeContent(w, r, pathname, info.ModTime(), file)
		Logger.Printf("%q", pathname)
		return
	}

	file, info, e, _ := openFileIfPublic(h.normalizePathname("/unknown-album.png"), false)
	if e != nil {
		Logger.Fatal(e)
	}
	defer file.Close()

	// TODO: Unify this in serveFileContents. Find a way to not copy and paste this.
	http.ServeContent(w, r, pathname, info.ModTime(), file)
	Logger.Printf("%q", pathname)
}

func (h *HTTPHandler) serveFile(w http.ResponseWriter, r *http.Request) {
	pathname := h.normalizePathname(r.URL.Path)
	if strings.HasSuffix(pathname, "/cover") {
		h.serveCover(pathname, w, r)
		return
	}

	h.serveFileContents(pathname, w, r)
}

func (h *HTTPHandler) serveFileContents(pathname string, w http.ResponseWriter, r *http.Request) {
	acceptsGzip := strings.Contains(r.Header.Get("Accept-Encoding"), "gzip")
	gzippable := IsStringInStrings(path.Ext(pathname), gzippableExtensions)

	file, info, e, isGzipped := openFileIfPublic(pathname, gzippable && acceptsGzip)
	if e != nil || file == nil || info == nil {
		Logger.Printf("%q: %v", pathname, e)
		http.NotFound(w, r)
		return
	}
	defer file.Close()

	if isGzipped {
		w.Header().Set("Content-Encoding", "gzip")
	}

	// If not gzipped, that's because it's a big ol' audio, video, or image file.
	// Tell the client to cache thos beans.
	if !gzippable {
		w.Header().Set("Cache-Control", "max-age=604800")
		//expires := time.Now()
		//duration, _ := time.ParseDuration("604800s")
		//expires = expires.Add(duration)
		//w.Header().Set("Expires", expires.Format(time.RFC1123))
	}

	http.ServeContent(w, r, pathname, info.ModTime(), file)
	Logger.Printf("%q", pathname)
}

// Returns an open File, a FileInfo, any error, and a bool indicating whether
// or not the file contains gzipped contents.
func openFileIfPublic(pathname string, shouldTryGzip bool) (*os.File, os.FileInfo, error, bool) {
	file, info, e := OpenFileAndInfo(pathname)
	if e != nil {
		return nil, nil, e, false
	}

	if !IsFileWorldReadable(info) {
		file.Close()
		Logger.Printf("NOTE: %q not world-readable", pathname)
		return nil, nil, errors.New(fmt.Sprintf("openFileIfPublic: %q not public", pathname)), false
	}

	if shouldTryGzip {
		gzFile, gzInfo := openOrCreateGzipped(pathname, file, info)
		if gzFile == nil {
			Logger.Printf("Could not create new gz file for: %q, %v", pathname, e)
			file.Seek(0, os.SEEK_SET)
			return file, info, nil, false
		}
		return gzFile, gzInfo, nil, true
	}

	return file, info, nil, false
}

// Given a pathname to an uncompressed file, opens or creates an equivalent
// gzipped file and returns it. On error, it logs the error and returns nil,
// nil.
//
// pathname, file, and info all refer to the uncompressed file. (We need
// pathname because `info.Name()` gives us only the basename.)
func openOrCreateGzipped(pathname string, file *os.File, info os.FileInfo) (*os.File, os.FileInfo) {
	gzPathname := pathname + ".gz"
	gzFile, gzInfo, e := OpenFileAndInfo(gzPathname)
	if e != nil {
		Logger.Printf("Could not open %q: %v", gzPathname, e)
		return createGzipped(pathname, file, info)
	}

	if gzInfo.ModTime().After(info.ModTime()) {
		return gzFile, gzInfo
	}
	return createGzipped(pathname, file, info)
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

func redirectToLogin(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/login.html", http.StatusFound)
}

func shouldServeFileToAnonymousClients(pathname string) bool {
	return IsStringInStrings(path.Base(pathname), anonymousFiles)
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

func writeString(w http.ResponseWriter, s string) (int, error) {
	return w.Write([]byte(s))
}
