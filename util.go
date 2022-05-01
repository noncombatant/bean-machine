// Copyright 2016 by Chris Palmer (https://noncombatant.org), and released under
// the terms of the GNU GPL3. See web/index.html for more information.

// Assorted utility functions.

package main

import (
	"compress/gzip"
	"crypto/rand"
	"embed"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
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

	digitsFinder = regexp.MustCompile(`(\d+)`)
)

// Copies the file named by `source` into the file named by `destination`.
// Returns an error, if any.
//
// See also `MustCopyFileByName`.
func CopyFileByName(destination, source string) error {
	source = filepath.Clean(source)
	destination = filepath.Clean(destination)
	if source == destination {
		return nil
	}

	s, e := os.Open(source)
	if e != nil {
		return e
	}
	defer s.Close()

	d, e := os.Create(destination)
	if e != nil {
		return e
	}
	defer d.Close()

	_, e = io.Copy(d, s)
	return e
}

// Returns a copy of `s`, with all double quotes escaped with a backslash.
func EscapeDoubleQuotes(s string) string {
	return strings.ReplaceAll(s, "\"", "\\\"")
}

// Returns the first substring of decimal digits in `s`, or an empty string if
// there is no such substring.
func ExtractDigits(s string) string {
	results := digitsFinder.FindStringSubmatch(s)
	if len(results) > 0 {
		return results[0]
	}
	return ""
}

// Returns the basename's extension, including the '.', normalized to
// lowercase. If the basename has no extension, returns an empty string.
func GetBasenameExtension(pathname string) string {
	return strings.ToLower(filepath.Ext(pathname))
}

// Reads `input`, gzips it (with `gzip.BestCompression`), and stores the output
// in a file named by `outputPathname`. This function will clobber any previous
// file named by `outputPathname`.
func GzipStream(outputPathname string, input io.Reader) error {
	gzFile, e := os.OpenFile(outputPathname, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if e != nil {
		return e
	}
	defer gzFile.Close()

	gzWriter, e := gzip.NewWriterLevel(gzFile, gzip.BestCompression)
	if e != nil {
		return e
	}
	defer gzWriter.Close()

	buffer := make([]byte, 4096)
	for {
		count, e := input.Read(buffer)
		if count == 0 && io.EOF == e {
			break
		}
		if e != nil {
			return e
		}

		_, e = gzWriter.Write(buffer[:count])
		if e != nil {
			return e
		}
	}
	return nil
}

func IsAudioPathname(pathname string) bool {
	return IsStringInStrings(GetBasenameExtension(pathname), audioFormatExtensions)
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

func IsDocumentPathname(pathname string) bool {
	return IsStringInStrings(GetBasenameExtension(pathname), documentFormatExtensions)
}

func IsFileWorldReadable(info os.FileInfo) bool {
	return info.Mode()&0004 == 04
}

// Returns true if `haystack` contains `needle`.
func IsStringInStrings(needle string, haystack []string) bool {
	for _, s := range haystack {
		if needle == s {
			return true
		}
	}
	return false
}

func IsImagePathname(pathname string) bool {
	return IsStringInStrings(GetBasenameExtension(pathname), imageFormatExtensions)
}

func IsVideoPathname(pathname string) bool {
	return IsStringInStrings(GetBasenameExtension(pathname), videoFormatExtensions)
}

// Copies the file named by `source` into the file named by `destination`. If
// an error occurs, logs fatal.
//
// See also `CopyFileByName`.
func MustCopyFileByName(destination, source string) {
	e := CopyFileByName(destination, source)
	if e != nil {
		log.Fatal(e)
	}
}

// Fills `bytes` with cryptographically random data. If an error occurs, logs
// fatal.
//
// See also `MustMakeRandomBytes`.
func MustGetRandomBytes(bytes []byte) {
	_, e := rand.Read(bytes)
	if e != nil {
		log.Fatal(e)
	}
}

// Returns `count` cryptographically random bytes. If an error occurs, logs
// fatal.
//
// See also `MustGetRandomBytes`.
func MustMakeRandomBytes(count int) []byte {
	bytes := make([]byte, count)
	MustGetRandomBytes(bytes)
	return bytes
}

// Returns the pathname with the basename's extension (including its '.')
// removed. If the basename has no extension, returns the pathname.
func RemoveBasenameExtension(pathname string) string {
	dot := strings.LastIndex(pathname, ".")
	if -1 == dot {
		return pathname
	}

	slash := strings.LastIndex(pathname, string(os.PathSeparator))
	if slash > dot {
		// There may be a dot, but it's not in the basename. In that case, return
		// the whole pathname.
		return pathname
	}

	return pathname[:dot]
}

func OpenFileAndInfo(pathname string) (*os.File, os.FileInfo, error) {
	file, e := os.Open(pathname)
	if e != nil {
		return nil, nil, e
	}
	info, e := file.Stat()
	if e != nil {
		file.Close()
		return nil, nil, e
	}
	return file, info, nil
}

func OpenFileAndInfoFS(pathname string, fs embed.FS) (fs.File, os.FileInfo, error) {
	file, e := fs.Open(pathname)
	if e != nil {
		return nil, nil, e
	}
	info, e := file.Stat()
	if e != nil {
		file.Close()
		return nil, nil, e
	}
	return file, info, nil
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

// https://twinnation.org/articles/33/remove-accents-from-characters-in-go
func RemoveAccents(s string) string {
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	output, _, e := transform.String(t, s)
	if e != nil {
		panic(e)
	}
	return output
}
