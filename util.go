// Copyright 2016 by Chris Palmer (https://noncombatant.org), and released under
// the terms of the GNU GPL3. See help.html for more information.

// Assorted utility functions.
//
// TODO: Maybe make this a separate module.

package main

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

	digitsFinder = regexp.MustCompile(`(\d+)`)
)

// TODO: These are not application-generic; move them out.
func IsAudioPathname(pathname string) bool {
	return IsStringInStrings(GetBasenameExtension(pathname), audioFormatExtensions)
}

func IsDocumentPathname(pathname string) bool {
	return IsStringInStrings(GetBasenameExtension(pathname), documentFormatExtensions)
}

func IsFileNewestInDirectory(directoryName, baseName string) bool {
	file, e := os.Open(path.Join(directoryName, baseName))
	if e != nil {
		return false
	}
	defer file.Close()

	status, e := file.Stat()
	if e != nil {
		return false
	}
	modified := status.ModTime()

	infos, e := ioutil.ReadDir(directoryName)
	if e != nil {
		return false
	}

	for _, info := range infos {
		if info.IsDir() && modified.Before(info.ModTime()) {
			return false
		}
	}
	return true
}

func IsImagePathname(pathname string) bool {
	return IsStringInStrings(GetBasenameExtension(pathname), imageFormatExtensions)
}

func IsVideoPathname(pathname string) bool {
	return IsStringInStrings(GetBasenameExtension(pathname), videoFormatExtensions)
}

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

// Copies the file named by `source` into the file named by `destination`. If
// an error occurs, logs fatal.
//
// See also `CopyFileByName`.
func MustCopyFileByName(destination, source string) {
	e := CopyFileByName(destination, source)
	if e != nil {
		Logger.Fatalf("Could not CopyFileByName(%q, %q): %v\n", destination, source, e)
	}
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

// Returns a copy of `s`, with all double quotes escaped with a backslash.
func EscapeDoubleQuotes(s string) string {
	return strings.ReplaceAll(s, "\"", "\\\"")
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

// Fills `bytes` with cryptographically random data. If an error occurs, logs
// fatal.
//
// See also `MustMakeRandomBytes`.
func MustGetRandomBytes(bytes []byte) {
	_, e := rand.Read(bytes)
	if e != nil {
		Logger.Fatalf("Could not get random bytes: %v", e)
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

func IsFileWorldReadable(info os.FileInfo) bool {
	return info.Mode()&0004 == 04
}
