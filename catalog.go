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
	"time"
)

type Catalog struct {
	ItemInfos
}

var (
	catalog = Catalog{}
)

const (
	catalogFile     = "catalog.gobs.gz"
)

func (c *Catalog) writeCatalog(root string) {
	catalogFilePath := path.Join(root, catalogFile)
	gobs, e := os.OpenFile(catalogFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if e != nil {
		log.Fatal(e)
	}

	gz, e := gzip.NewWriterLevel(gobs, 9)
	if e != nil {
		log.Fatal(e)
	}

	encoder := gob.NewEncoder(gz)
	if e := encoder.Encode(c); e != nil {
		log.Fatal(e)
	}

	if e := gz.Close(); e != nil {
		log.Fatal(e)
	}

	if e := gobs.Close(); e != nil {
		log.Fatal(e)
	}
}

func shouldSkipFile(pathname string, info os.FileInfo) bool {
	basename := path.Base(pathname)
	return basename == "" || basename[0] == '.' || info.Size() == 0
}

func (c *Catalog) BuildCatalog(root string) {
	// Log the walk progress periodically so that the operator knows what’s going
	// on.
	count := 0
	timerFrequency := 1 * time.Second
	timer := time.NewTimer(timerFrequency)

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

			if IsAudioPathname(pathname) || IsVideoPathname(pathname) {
				webPathname := pathname[len(root)+1:]
				itemInfo := ItemInfo{Pathname: webPathname}
				itemInfo.File, _ = id3.Read(input)
				time := info.ModTime()
				itemInfo.ModTime = fmt.Sprintf("%04d-%02d-%02d", time.Year(), time.Month(), time.Day())
				itemInfo.fillMetadata()
				c.ItemInfos = append(c.ItemInfos, itemInfo)

				count++
				select {
				case _, ok := <-timer.C:
					if ok {
						log.Print(count)
						timer.Reset(timerFrequency)
					} else {
						log.Fatal("Channel closed?!")
					}
				default:
					// Do nothing.
				}
			}

			return nil
		})

	if e != nil {
		log.Print(e)
	}
	c.writeCatalog(root)
}

func (c *Catalog) ReadCatalog(root string) {
	gobs, e := os.Open(path.Join(root, catalogFile))
	if e != nil {
		log.Fatal(e)
	}
	gz, e := gzip.NewReader(gobs)
	if e != nil {
		log.Fatal(e)
	}
	decoder := gob.NewDecoder(gz)
	if e := decoder.Decode(c); e != nil && e != io.EOF {
		log.Fatal(e)
	}
	if e = gz.Close(); e != nil {
		log.Fatal(e)
	}
	if e = gobs.Close(); e != nil {
		log.Fatal(e)
	}
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
