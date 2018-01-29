// Copyright 2016 by Chris Palmer (https://noncombatant.org), and released under
// the terms of the GNU GPL3. See help.html for more information.

package main

import (
	"crypto/md5"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"regexp"
	"strings"
)

func copyFile(source, destination string) {
	s, e := os.Open(source)
	defer s.Close()
	if e != nil {
		log.Fatalf("Could not read %q: %s\n", source, e)
	}

	d, e := os.Create(destination)
	defer d.Close()
	if e != nil {
		log.Fatalf("Could not write %q: %s\n", destination, e)
	}

	_, e = io.Copy(d, s)
	if e != nil {
		log.Fatalf("Could not copy %q to %q: %s\n", source, destination, e)
	}
}

func computeMD5(pathname string) (string, error) {
	file, e := os.Open(pathname)
	if e != nil {
		return "", e
	}
	defer file.Close()

	hash := md5.New()
	if _, e := io.Copy(hash, file); e != nil {
		return "", e
	}

	var result []byte
	return string(hash.Sum(result)), nil
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

func printStringArray(strings []string) {
	fmt.Println()
	for _, s := range strings {
		fmt.Printf("    %q\n", s)
	}
	fmt.Println()
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
	dot := strings.LastIndex(pathname, ".")
	if -1 == dot {
		return ""
	}
	return strings.ToLower(pathname[dot:])
}

func isStringInStrings(needle string, haystack []string) bool {
	for _, s := range haystack {
		if needle == s {
			return true
		}
	}
	return false
}

func isStringAllDigits(s string) bool {
	matched, e := regexp.MatchString("^\\d+$", s)
	if e != nil {
		log.Fatal(e)
	}
	return matched
}

func maybeQuote(s string) string {
	if isStringAllDigits(s) {
		return s
	}
	return fmt.Sprintf("%q", s)
}
