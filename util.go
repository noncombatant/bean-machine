// Copyright 2016 by Chris Palmer (https://noncombatant.org)
// SPDX-License-Identifier: GPL-3.0

// Assorted utility functions.

package main

import (
	"crypto/rand"
	"embed"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
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
		".webp",
	}

	digitsFinder = regexp.MustCompile(`(\d+)`)
)

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

func IsAudioPathname(pathname string) bool {
	return slices.Contains(audioFormatExtensions, GetBasenameExtension(pathname))
}

func IsDocumentPathname(pathname string) bool {
	return slices.Contains(documentFormatExtensions, GetBasenameExtension(pathname))
}

func IsFileWorldReadable(info os.FileInfo) bool {
	return info.Mode()&0004 == 04
}

func IsImagePathname(pathname string) bool {
	return slices.Contains(imageFormatExtensions, GetBasenameExtension(pathname))
}

func IsVideoPathname(pathname string) bool {
	return slices.Contains(videoFormatExtensions, GetBasenameExtension(pathname))
}

func GetRandomBytes(count int) ([]byte, error) {
	bytes := make([]byte, count)
	n, e := rand.Read(bytes)
	if e != nil {
		return nil, e
	}
	if n != count {
		return nil, nil
	}
	return bytes, nil
}

// Returns the pathname with the basename's extension (including its '.')
// removed. If the basename has no extension, returns the pathname.
func RemoveBasenameExtension(pathname string) string {
	dot := strings.LastIndex(pathname, ".")
	if dot == -1 {
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
		_ = file.Close()
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
		_ = file.Close()
		return nil, nil, e
	}
	return file, info, nil
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
