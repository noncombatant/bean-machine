// Copyright 2016 by Chris Palmer (https://noncombatant.org), and released under
// the terms of the GNU GPL3. See web/index.html for more information.

package main

import (
	"archive/zip"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"log"
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
	tokenLength        = 16
	encodedTokenLength = 24
)

var (
	anonymousFiles = []string{
		"favicon.ico",
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
		log.Print(e)
		return nil, nil
	}

	gzFile, gzInfo, e := OpenFileAndInfo(gzPathname)
	if e != nil {
		log.Print(e)
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

	cookie, e := r.Cookie("token")
	if e != nil {
		log.Printf("Refusing %q (missing cookie)", r.URL.Path)
		redirectToLogin(w, r)
		return
	}

	if !h.checkToken(cookie.Value) {
		log.Printf("Refusing %q (invalid token)", r.URL.Path)
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
	} else if r.URL.RawQuery == "download" {
		h.serveZip(w, r)
	}

	h.serveFile(w, r)
}

func (h *HTTPHandler) checkToken(token string) bool {
	if len(token) != encodedTokenLength {
		return false
	}

	data, e := base64.URLEncoding.DecodeString(token)
	if e != nil || len(data) != tokenLength {
		return false
	}

	_, e = os.Stat(path.Join(h.ConfigurationPathname, sessionsDirectoryName, token))
	return e == nil
}

func (h *HTTPHandler) generateToken() string {
	bytes := MustMakeRandomBytes(tokenLength)
	token := base64.URLEncoding.EncodeToString(bytes)
	pathname := path.Join(h.ConfigurationPathname, sessionsDirectoryName, token)

	file, e := os.Create(pathname)
	if e != nil {
		log.Fatal(e)
	}

	e = file.Close()
	if e != nil {
		log.Fatal(e)
	}

	return token
}

func (h *HTTPHandler) handleLogIn(w http.ResponseWriter, r *http.Request) {
	username := normalizeUsername(r.FormValue("name"))
	password := r.FormValue("password")
	stored := readPasswordDatabase(path.Join(h.ConfigurationPathname, passwordsBasename))

	cookie := &http.Cookie{Name: "token", Value: "", Secure: true, HttpOnly: true, Expires: getCookieLifetime(), Path: "/"}
	if checkPassword(stored, username, password) {
		log.Printf("%q successful", username)
		cookie.Value = h.generateToken()
		http.SetCookie(w, cookie)
		http.Redirect(w, r, "/index.html", http.StatusFound)
	} else {
		log.Printf("%q unsuccessful", username)
		http.SetCookie(w, cookie)
		redirectToLogin(w, r)
	}
}

func (h *HTTPHandler) handleSearch(w http.ResponseWriter, r *http.Request) {
	queries := r.URL.Query()["q"]
	if len(queries) == 0 {
		log.Print("Ignoring empty search.")
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

func (h *HTTPHandler) serveContent(w http.ResponseWriter, r *http.Request, pathname string, modified time.Time, content io.ReadSeeker) {
	http.ServeContent(w, r, pathname, modified, content)
	log.Printf("%v %v %v %v", r.RemoteAddr, r.Method, r.Host, r.URL)
}

func (h *HTTPHandler) serveCover(pathname string, w http.ResponseWriter, r *http.Request) {
	for _, extension := range coverExtensions {
		file, info, _, e := openFileIfPublic(pathname+extension, false)
		if e != nil {
			continue
		}
		defer file.Close()

		// TODO: Unify this in serveFileContents.
		h.serveContent(w, r, pathname, info.ModTime(), file)
		return
	}

	file, info, _, e := openFileIfPublic(h.normalizePathname("/unknown-album.png"), false)
	if e != nil {
		log.Fatal(e)
	}
	defer file.Close()

	// TODO: Unify this in serveFileContents. Find a way to not copy and paste this.
	h.serveContent(w, r, pathname, info.ModTime(), file)
}

func zipDirectory(pathname string) (*os.File, error) {
	os.Mkdir("/tmp/beanzip", 0700)
	file, e := ioutil.TempFile("/tmp/beanzip", "album.zip")
	if e != nil {
		return nil, e
	}
	defer os.Remove(file.Name())

	zipWriter := zip.NewWriter(file)

	infos, e := ioutil.ReadDir(pathname)
	if e != nil {
		return nil, e
	}

	for _, info := range infos {
		f, e := zipWriter.Create(info.Name())
		if e != nil {
			return nil, e
		}
		contents, e := ioutil.ReadFile(pathname + "/" + info.Name())
		if e != nil {
			return nil, e
		}
		_, e = f.Write(contents)
		if e != nil {
			return nil, e
		}
	}

	e = zipWriter.Close()
	if e != nil {
		return nil, e
	}

	return file, nil
}

func (h *HTTPHandler) serveZip(w http.ResponseWriter, r *http.Request) {
	pathname := h.normalizePathname(r.URL.Path)
	info, e := os.Stat(pathname)
	if e != nil {
		log.Print("stat", e)
		return
	}

	zipFile, e := zipDirectory(pathname)
	if e != nil {
		log.Print(e)
		return
	}

	_, e = zipFile.Seek(0, 0)
	if e != nil {
		log.Print(e)
		zipFile.Close()
		return
	}

	w.Header().Set("Content-Type", "application/zip, application/octet-stream")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filepath.Base(filepath.Dir(pathname))+" - "+filepath.Base(pathname)+".zip"))
	h.serveContent(w, r, pathname, info.ModTime(), zipFile)

	e = zipFile.Close()
	if e != nil {
		log.Print(e)
	}
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

	file, info, isGzipped, e := openFileIfPublic(pathname, gzippable && acceptsGzip)
	if e != nil || file == nil || info == nil {
		log.Print(e)
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

	h.serveContent(w, r, pathname, info.ModTime(), file)
}

// Returns an open File, a FileInfo, any error, and a bool indicating whether
// or not the file contains gzipped contents.
func openFileIfPublic(pathname string, shouldTryGzip bool) (*os.File, os.FileInfo, bool, error) {
	file, info, e := OpenFileAndInfo(pathname)
	if e != nil {
		return nil, nil, false, e
	}

	if !IsFileWorldReadable(info) {
		file.Close()
		return nil, nil, false, fmt.Errorf("openFileIfPublic: %q not public", pathname)
	}

	if shouldTryGzip {
		gzFile, gzInfo := openOrCreateGzipped(pathname, file, info)
		if gzFile == nil {
			log.Print(e)
			file.Seek(0, io.SeekStart)
			return file, info, false, nil
		}
		return gzFile, gzInfo, true, nil
	}

	return file, info, false, nil
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
		log.Print(e)
		return createGzipped(pathname, file, info)
	}

	if gzInfo.ModTime().After(info.ModTime()) {
		return gzFile, gzInfo
	}
	return createGzipped(pathname, file, info)
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
