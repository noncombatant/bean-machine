// Copyright 2016 by Chris Palmer (https://noncombatant.org)
// SPDX-License-Identifier: GPL-3.0

package main

import (
	"archive/zip"
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"io"
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

//go:embed web
var frontend embed.FS

var (
	coverExtensions = []string{
		".gif",
		".jpeg",
		".jpg",
		".png",
	}

	wordSplitter = regexp.MustCompile(`\s+`)
)

type httpHandler struct {
	Root                  string
	ConfigurationPathname string
	*Catalog
	*log.Logger
}

func (h *httpHandler) isAuthenticated(r *http.Request) bool {
	cookie, e := r.Cookie("token")
	if e != nil {
		return false
	}
	return checkToken(cookie.Value, path.Join(h.ConfigurationPathname, sessionsDirectoryName))
}

func (h *httpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.Logger.Printf("%q,%q,%q,%q,%q", r.RemoteAddr, r.Proto, r.Method, r.Host, r.RequestURI)
	if r.URL.Path == "/login.html" && r.Method == http.MethodPost {
		h.handleLogIn(w, r)
		return
	}

	// All the front-end files can and should be served to anonymous clients.
	if r.URL.Path == "/" {
		r.URL.Path = "/index.html"
	}
	// ...but the UI will break if we serve the index to an anonymous client.
	if r.URL.Path == "/index.html" && !h.isAuthenticated(r) {
		redirectToLogin(w, r)
	}

	if f, info, e := openFileAndInfoFS("web"+r.URL.Path, frontend); e == nil {
		data, _ := io.ReadAll(f)
		h.serveContent(w, r, r.URL.Path, info.ModTime(), bytes.NewReader(data))
		if e := f.Close(); e != nil {
			h.Logger.Print(e)
		}
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
			return
		}

		w.WriteHeader(http.StatusOK)
		page := h.buildMediaIndex(pathname)
		w.Write([]byte(page))
		return
	} else if r.URL.RawQuery == "download" {
		h.serveZip(w, r)
	}

	h.serveFile(w, r)
}

func (h *httpHandler) handleLogIn(w http.ResponseWriter, r *http.Request) {
	username := normalizeUsername(r.FormValue("name"))
	password := r.FormValue("password")
	credentials, e := readCredentials(path.Join(h.ConfigurationPathname, passwordsBasename))
	if e != nil {
		h.Logger.Print(e)
		redirectToLogin(w, r)
		return
	}

	ok, e := checkPassword(credentials, username, password)
	if e != nil {
		h.Logger.Print(e)
		redirectToLogin(w, r)
		return
	}
	if !ok {
		h.Logger.Printf("%q unsuccessful", username)
		redirectToLogin(w, r)
		return
	}

	h.Logger.Printf("%q successful", username)
	token, e := createToken(path.Join(h.ConfigurationPathname, sessionsDirectoryName))
	if e != nil {
		h.Logger.Print(e)
		redirectToLogin(w, r)
		return
	}
	http.SetCookie(w, getCookie(token))
	http.Redirect(w, r, "/index.html", http.StatusFound)
}

func (h *httpHandler) handleSearch(w http.ResponseWriter, r *http.Request) {
	queries := r.URL.Query()["q"]
	if len(queries) == 0 {
		h.Logger.Print("Ignoring empty search.")
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
			matches = matchItems(h.Catalog.ItemInfos, query)
			if len(matches) > 0 {
				goto done
			}
		}

		// If we get here, there were no new items in the last 6 months. Just make
		// a random query, then.
		query = "?"
	}

	if query == "?" {
		item := h.Catalog.ItemInfos[rand.Intn(len(h.Catalog.ItemInfos))]
		words := wordSplitter.Split(path.Dir(item.Pathname), -1)
		query = words[len(words)-1]
	}

	matches = matchItems(h.Catalog.ItemInfos, query)

done:
	json, e := json.Marshal(matches)
	if e != nil {
		h.Logger.Print(e)
		http.Error(w, "", 500)
	} else {
		w.Header().Set("Content-Type", "text/json")
		w.Write(json)
	}
}

func (h *httpHandler) normalizePathname(pathname string) string {
	if pathname == "/" {
		pathname = "/index.html"
	}
	pathname = path.Join(h.Root, filepath.Clean(pathname))
	if !strings.HasPrefix(pathname, h.Root) {
		return h.Root + "/404.html"
	}
	return pathname
}

func (h *httpHandler) serveContent(w http.ResponseWriter, r *http.Request, pathname string, modified time.Time, content io.ReadSeeker) {
	http.ServeContent(w, r, pathname, modified, content)
}

func (h *httpHandler) serveCover(pathname string, w http.ResponseWriter, r *http.Request) {
	for _, extension := range coverExtensions {
		file, info, e := h.openFileIfPublic(pathname + extension)
		if e != nil {
			continue
		}
		h.serveContent(w, r, pathname, info.ModTime(), file)
		if e := file.Close(); e != nil {
			h.Logger.Print(e)
		}
		return
	}

	f, info, e := openFileAndInfoFS("web/unknown-album.png", frontend)
	if e != nil {
		h.Logger.Fatal(e)
	}
	data, e := io.ReadAll(f)
	if e != nil {
		h.Logger.Print(e)
	} else {
		h.serveContent(w, r, r.URL.Path, info.ModTime(), bytes.NewReader(data))
	}
	if e := f.Close(); e != nil {
		h.Logger.Print(e)
	}
}

func zipDirectory(log *log.Logger, pathname string) (*os.File, error) {
	file, e := os.CreateTemp("", "*album.zip")
	if e != nil {
		return nil, e
	}
	defer func() {
		e := os.Remove(file.Name())
		if e != nil {
			log.Print(e)
		}
	}()

	entries, e := os.ReadDir(pathname)
	if e != nil {
		return nil, e
	}

	zipWriter := zip.NewWriter(file)
	for _, entry := range entries {
		f, e := zipWriter.Create(entry.Name())
		if e != nil {
			return nil, e
		}
		contents, e := os.ReadFile(pathname + "/" + entry.Name())
		if e != nil {
			return nil, e
		}
		_, e = f.Write(contents)
		if e != nil {
			return nil, e
		}
	}

	if e := zipWriter.Close(); e != nil {
		return nil, e
	}
	return file, nil
}

func (h *httpHandler) serveZip(w http.ResponseWriter, r *http.Request) {
	pathname := h.normalizePathname(r.URL.Path)
	info, e := os.Stat(pathname)
	if e != nil {
		h.Logger.Print("stat", e)
		return
	}

	zipFile, e := zipDirectory(h.Logger, pathname)
	if e != nil {
		h.Logger.Print(e)
		return
	}
	defer func() {
		if e := zipFile.Close(); e != nil {
			h.Logger.Print(e)
		}
	}()

	if _, e := zipFile.Seek(0, 0); e != nil {
		h.Logger.Print(e)
		return
	}

	// This may be redundant: server.go:3160: http: superfluous response.WriteHeader call from main.(*HTTPHandler).serveContent (authentication.go:250)
	w.Header().Set("Content-Type", "application/zip, application/octet-stream")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filepath.Base(filepath.Dir(pathname))+" - "+filepath.Base(pathname)+".zip"))
	h.serveContent(w, r, pathname, info.ModTime(), zipFile)
}

func (h *httpHandler) serveFile(w http.ResponseWriter, r *http.Request) {
	pathname := h.normalizePathname(r.URL.Path)
	if strings.HasSuffix(pathname, "/cover") {
		h.serveCover(pathname, w, r)
		return
	}

	h.serveFileContents(pathname, w, r)
}

func (h *httpHandler) serveFileContents(pathname string, w http.ResponseWriter, r *http.Request) {
	file, info, e := h.openFileIfPublic(pathname)
	if e != nil || file == nil || info == nil {
		h.Logger.Print(e)
		http.NotFound(w, r)
		return
	}
	h.serveContent(w, r, pathname, info.ModTime(), file)
	if e := file.Close(); e != nil {
		h.Logger.Print(e)
	}
}

func (h *httpHandler) openFileIfPublic(pathname string) (*os.File, os.FileInfo, error) {
	file, info, e := openFileAndInfo(pathname)
	if e != nil {
		return nil, nil, e
	}
	if !isFileWorldReadable(info) {
		_ = file.Close()
		return nil, nil, fmt.Errorf("openFileIfPublic: %q not public", pathname)
	}
	return file, info, nil
}

func redirectToLogin(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/login.html", http.StatusFound)
}

func (h *httpHandler) buildMediaIndex(pathname string) string {
	header := `<!DOCTYPE html>
<meta charset="UTF-8"/>
<meta name="viewport" content="width=device-width, initial-scale=1"/>
<link rel="stylesheet" href="/index.css"/>
<title>` + path.Base(pathname) + `</title>
`

	var builder strings.Builder
	builder.WriteString(header)

	entries, e := os.ReadDir(pathname)
	if e != nil {
		h.Logger.Print(e)
		return builder.String()
	}

	for _, entry := range entries {
		info, e := entry.Info()
		if e != nil {
			log.Print(e)
			continue
		}
		name := info.Name()
		if isImagePathname(name) {
			builder.WriteString(fmt.Sprintf("<img src=\"%s\"/>\n", escapeDoubleQuotes(name)))
		} else if isDocumentPathname(name) {
			name = escapeDoubleQuotes(name)
			builder.WriteString(fmt.Sprintf("<li><a href=\"%s\">%s</a></li>\n", name, name))
		}
	}
	return builder.String()
}
