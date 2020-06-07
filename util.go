// Copyright 2016 by Chris Palmer (https://noncombatant.org), and released under
// the terms of the GNU GPL3. See help.html for more information.

package main

// TODO: Write tests for all of these.
// TOOD: Maybe make it a separate module.

import (
	"bufio"
	"compress/gzip"
	"crypto/rand"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var (
	// NOTE: These must be kept in sync with the format extensions arrays in the
	// JS code.
	audioFormatExtensions = []string{
		".flac",
		".m4a",
		".mid",
		".midi",
		".mp3",
		".ogg",
		".wav",
		".wave",
	}
	videoFormatExtensions = []string{
		".avi",
		".m4v",
		".mkv",
		".mov",
		".mp4",
		".mpeg",
		".mpg",
		".ogv",
		".webm",
	}
	documentFormatExtensions = []string{
		".pdf",
		".txt",
	}
	imageFormatExtensions = []string{
		".gif",
		".jpeg",
		".jpg",
		".png",
	}

	digitsFinder = regexp.MustCompile(`\D*(\d+)`)
)

func IsAudioPathname(pathname string) bool {
	return IsStringInStrings(GetFileExtension(pathname), audioFormatExtensions)
}

func IsVideoPathname(pathname string) bool {
	return IsStringInStrings(GetFileExtension(pathname), videoFormatExtensions)
}

func IsDocumentPathname(pathname string) bool {
	return IsStringInStrings(GetFileExtension(pathname), documentFormatExtensions)
}

func IsImagePathname(pathname string) bool {
	return IsStringInStrings(GetFileExtension(pathname), imageFormatExtensions)
}

// TODO: Rename to CopyFileByName
func CopyFile(source, destination string) {
	source = filepath.Clean(source)
	destination = filepath.Clean(destination)
	if source == destination {
		return
	}

	s, e := os.Open(source)
	if e != nil {
		Logger.Fatalf("Could not read %q: %s\n", source, e)
	}
	defer s.Close()

	d, e := os.Create(destination)
	if e != nil {
		Logger.Fatalf("Could not write %q: %s\n", destination, e)
	}
	defer d.Close()

	_, e = io.Copy(d, s)
	if e != nil {
		Logger.Fatalf("Could not copy %q to %q: %s\n", source, destination, e)
	}
}

func ExtractNumericString(numeric string) string {
	results := digitsFinder.FindStringSubmatch(numeric)
	if len(results) > 0 {
		return results[0]
	}
	return ""
}

func shouldSkipFile(pathname string, info os.FileInfo) bool {
	basename := path.Base(pathname)
	return "" == basename || '.' == basename[0] || 0 == info.Size()
}

func GetFileExtension(pathname string) string {
	return strings.ToLower(filepath.Ext(pathname))
}

func RemoveFileExtension(pathname string) string {
	dot := strings.LastIndex(pathname, ".")
	if -1 == dot {
		return pathname
	}
	return pathname[:dot]
}

func EscapeDoubleQuotes(s string) string {
	return strings.ReplaceAll(s, "\"", "\\\"")
}

func IsStringInStrings(needle string, haystack []string) bool {
	for _, s := range haystack {
		if needle == s {
			return true
		}
	}
	return false
}

func MustGetRandomBytes(bytes []byte) {
	_, e := rand.Read(bytes)
	if e != nil {
		Logger.Fatalf("Could not get random bytes: %v", e)
	}
}

func MustMakeRandomBytes(length int) []byte {
	bytes := make([]byte, length)
	MustGetRandomBytes(bytes)
	return bytes
}

func GzipFile(gzPathname string, file io.Reader) error {
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

func OpenFileAndGetInfo(pathname string) FileAndInfoResult {
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

func assertValidRootPathname(root string) {
	info, e := os.Stat(root)
	if e != nil || !info.IsDir() {
		Logger.Fatal("Cannot continue without a valid music-directory.")
	}
}

// Parses `s` as a signed 32-bit integer, represented in any base up to base
// 36, and returns the result. (The functionality is equivalent to
// `strconv.ParseInt(s, 0, 32)`.) If any error occurs, returns 0.
func ParseIntegerOr0(s string) int {
	i, e := strconv.ParseInt(s, 0, 32)
	if e != nil {
		return 0
	}
	return int(i)
}

func IsDirectoryEmpty(pathname string) (bool, error) {
	f, e := os.Open(pathname)
	if e != nil {
		return false, e
	}
	defer f.Close()
	_, e = f.Readdir(1)
	return e == io.EOF, e
}
