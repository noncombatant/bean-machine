// Copyright 2016 by Chris Palmer (https://noncombatant.org), and released under
// the terms of the GNU GPL3. See help.html for more information.

package main

import (
	"bufio"
	"compress/gzip"
	"crypto/rand"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func copyFile(source, destination string) {
	source = filepath.Clean(source)
	destination = filepath.Clean(destination)
	if source == destination {
		return
	}

	s, e := os.Open(source)
	if e != nil {
		log.Fatalf("copyFile: Could not read %q: %s\n", source, e)
	}
	defer s.Close()

	d, e := os.Create(destination)
	if e != nil {
		log.Fatalf("copyFile: Could not write %q: %s\n", destination, e)
	}
	defer d.Close()

	_, e = io.Copy(d, s)
	if e != nil {
		log.Fatalf("copyFile: Could not copy %q to %q: %s\n", source, destination, e)
	}
}

func normalizeNumericString(numeric string) string {
	numeric = strings.TrimSpace(numeric)
	if len(numeric) == 0 {
		return ""
	}

	i := strings.Index(numeric, "/")
	if i > 0 {
		numeric = numeric[:i]
	}

	for i = 0; numeric[i] == '0' && i < len(numeric)-1; i++ {
	}
	return numeric[i:]
}

func escapePathname(pathname string) string {
	pathname = strings.Replace(pathname, "%", "%25", -1)
	pathname = strings.Replace(pathname, "#", "%23", -1)
	pathname = strings.Replace(pathname, "?", "%3f", -1)
	return pathname
}

func shouldSkipFile(pathname string, info os.FileInfo) bool {
	basename := path.Base(pathname)
	return "" == basename || '.' == basename[0] || 0 == info.Size()
}

func getFileExtension(pathname string) string {
	return strings.ToLower(filepath.Ext(pathname))
}

func removeFileExtension(pathname string) string {
	dot := strings.LastIndex(pathname, ".")
	if -1 == dot {
		return pathname
	}
	return pathname[:dot]
}

func isStringInStrings(needle string, haystack []string) bool {
	for _, s := range haystack {
		if needle == s {
			return true
		}
	}
	return false
}

func makeRandomBytes(length int) []byte {
	bytes := make([]byte, length)
	_, e := rand.Read(bytes)
	if e != nil {
		log.Fatalf("makeRandomBytes: Could not get random bytes: %v", e)
	}
	return bytes
}

func compressFile(gzPathname string, file io.Reader) error {
	bytes, e := ioutil.ReadAll(file)
	if e != nil {
		return e
	}

	gzFile, e := os.OpenFile(gzPathname, os.O_WRONLY|os.O_CREATE, 0666)
	if e != nil {
		return e
	}
	defer gzFile.Close()

	gzWriter, e := gzip.NewWriterLevel(gzFile, gzip.BestCompression)
	if e != nil {
		return e
	}
	defer gzWriter.Close()

	bufferedWriter := bufio.NewWriter(gzWriter)
	defer bufferedWriter.Flush()

	_, e = bufferedWriter.Write(bytes)
	if e != nil {
		os.Remove(gzPathname)
	}
	return e
}

type FileAndInfoResult struct {
	File  *os.File
	Info  os.FileInfo
	Error error
}

func openFileAndGetInfo(pathname string) FileAndInfoResult {
	file, e := os.Open(pathname)
	if e != nil {
		return FileAndInfoResult{File: nil, Error: e}
	}

	info, e := file.Stat()
	if e != nil {
		file.Close()
		return FileAndInfoResult{File: nil, Error: e}
	}

	return FileAndInfoResult{File: file, Info: info, Error: nil}
}

func replaceStringAndLog(s, old, new, description string) string {
	if -1 != strings.Index(s, old) {
		log.Printf("replaceStringAndLog: %q contains a %s", s, description)
		s = strings.Replace(s, old, new, -1)
	}
	return s
}

func replaceTSVMetacharacters(s string) string {
	s = replaceStringAndLog(s, "\t", " ", "tab")
	s = replaceStringAndLog(s, "\n", " ", "newline")
	return s
}
