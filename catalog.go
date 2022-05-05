// Copyright 2019 by Chris Palmer (https://noncombatant.org), and released under
// the terms of the GNU GPL3. See web/index.html for more information.

package main

import (
	"compress/gzip"
	"encoding/gob"
	"fmt"
	"id3"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type Catalog struct {
	ItemInfos
}

const (
	catalogFile = "catalog.gobs.gz"
	eraseLine   = "\033[2K\r"
)

func WriteCatalog(w io.Writer, c *Catalog) error {
	e := gob.NewEncoder(w)
	return e.Encode(c)
}

func WriteCatalogByPathname(pathname string, c *Catalog) error {
	w, e := os.OpenFile(pathname, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if e != nil {
		return e
	}
	zw, e := gzip.NewWriterLevel(w, 9)
	if e != nil {
		return e
	}
	if e := WriteCatalog(zw, c); e != nil {
		return e
	}
	if e := zw.Close(); e != nil {
		return e
	}
	return w.Close()
}

func shouldSkipFile(pathname string, info os.FileInfo) bool {
	basename := path.Base(pathname)
	return basename == "" || basename[0] == '.' || info.Size() == 0
}

func BuildCatalog(root string) (*Catalog, error) {
	var c Catalog
	previousDir := ""
	e := filepath.Walk(root,
		func(pathname string, info os.FileInfo, e error) error {
			if e != nil {
				log.Print(e)
				return e
			}
			if shouldSkipFile(pathname, info) || info.Mode().IsDir() || !info.Mode().IsRegular() {
				return nil
			}

			input, e := os.Open(pathname)
			if e != nil {
				log.Print(e)
				return e
			}
			defer input.Close()

			// If we have progressed to a new directory, print progress indicator.
			dir := path.Dir(path.Dir(pathname[len(root)+1:]))
			if dir != previousDir {
				fmt.Fprintf(os.Stdout, "%s%s", eraseLine, dir)
				previousDir = dir
			}

			if IsAudioPathname(pathname) || IsVideoPathname(pathname) {
				webPathname := pathname[len(root)+1:]
				itemInfo := ItemInfo{Pathname: webPathname}
				itemInfo.File, _ = id3.Read(input)
				time := info.ModTime()
				itemInfo.ModTime = fmt.Sprintf("%04d-%02d-%02d", time.Year(), time.Month(), time.Day())
				itemInfo.fillMetadata()
				c.ItemInfos = append(c.ItemInfos, itemInfo)
			}
			return nil
		})

	fmt.Fprintf(os.Stdout, "%s\n", eraseLine)
	return &c, e
}

func ReadCatalog(r io.Reader) (*Catalog, error) {
	var c Catalog
	d := gob.NewDecoder(r)
	if e := d.Decode(&c); e != nil && e != io.EOF {
		return nil, e
	}
	return &c, nil
}

func ReadCatalogByPathname(pathname string) (*Catalog, error) {
	f, e := os.Open(pathname)
	if e != nil {
		return nil, e
	}
	zr, e := gzip.NewReader(f)
	if e != nil {
		return nil, e
	}
	c, e := ReadCatalog(zr)
	if e != nil {
		return nil, e
	}
	if e = zr.Close(); e != nil {
		return nil, e
	}
	if e = f.Close(); e != nil {
		return nil, e
	}
	return c, nil
}

func buildMediaIndex(pathname string) string {
	header := `<!DOCTYPE html>
<meta charset="UTF-8"/>
<meta name="viewport" content="width=device-width, initial-scale=1"/>
<link rel="stylesheet" href="/media.css"/>
<title>` + path.Base(pathname) + `</title>
`

	var builder strings.Builder
	builder.WriteString(header)

	infos, e := ioutil.ReadDir(pathname)
	if e != nil {
		log.Print(e)
		return builder.String()
	}

	for _, info := range infos {
		name := info.Name()
		if IsImagePathname(name) {
			builder.WriteString(fmt.Sprintf("<img src=\"%s\"/>\n", EscapeDoubleQuotes(name)))
		} else if IsDocumentPathname(name) {
			name = EscapeDoubleQuotes(name)
			builder.WriteString(fmt.Sprintf("<li><a href=\"%s\">%s</a></li>\n", name, name))
		}
	}
	return builder.String()
}
