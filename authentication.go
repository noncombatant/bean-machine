// Copyright 2016 by Chris Palmer (https://noncombatant.org), and released under
// the terms of the GNU GPL3. See help.html for more information.

package main

import (
	"encoding/hex"
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
	authenticationTokenLength = 16
)

var (
	anonymousFiles = []string{"manifest.json", "bean-machine.css", "readme.html", "help.html"}
)

func generateToken() string {
	token := make([]byte, authenticationTokenLength)
	getRandomBytes(token)
	return hex.EncodeToString(token)
}

type AuthenticatingFileHandler struct {
	Root     string
	Sessions map[string]bool
}

func (h AuthenticatingFileHandler) isRequestAuthenticated(r *http.Request) bool {
	cookie, e := r.Cookie("token")
	if e == nil && len(cookie.Value) == 2*authenticationTokenLength {
		if _, e := hex.DecodeString(cookie.Value); e != nil {
			return false
		}
		if _, ok := h.Sessions[cookie.Value]; ok {
			return true
		}
	}

	return false
}

func (h AuthenticatingFileHandler) handleLogIn(w http.ResponseWriter, r *http.Request) {
	if CheckPassword(r.FormValue("name"), r.FormValue("password")) {
		token := generateToken()
		h.Sessions[token] = true
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

func stringInSlice(a string, slice []string) bool {
	for _, b := range slice {
		if b == a {
			return true
		}
	}
	return false
}

func shouldServeFileToAnonymousClients(pathname string) bool {
	return stringInSlice(path.Base(pathname), anonymousFiles)
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
