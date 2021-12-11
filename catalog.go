// Copyright 2019 by Chris Palmer (https://noncombatant.org), and released under
// the terms of the GNU GPL3. See web/index.html for more information.

package main

import (
	"encoding/gob"
	"fmt"
	"id3"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

type Catalog struct {
	ItemInfos

	// Time at which this Catalog's ItemInfos were last updated from the
	// filesystem.
	Modified time.Time

	// `buildCatalogFromWalk` could be invoked through either `buildCatalog` or
	// `serveApp`. This sentinel avoids having multiple invocations walk twice
	// and stomp on the `catalogFile`.
	buildCatalogFromWalkInProgress bool
}

var (
	catalog = Catalog{}
)

const (
	catalogFile     = "catalog.gobs"
	catalogFileTemp = "catalog.gobs.tmp"
)

func (c *Catalog) readCatalog(gobs *os.File) {
	decoder := gob.NewDecoder(gobs)
	infos := ItemInfos{}
	e := decoder.Decode(&infos)
	if e != nil && e != io.EOF {
		Logger.Fatal(e)
	}
	c.ItemInfos = infos
}

func (c *Catalog) writeCatalog(root string, infos ItemInfos) {
	catalogFileTempPath := path.Join(root, catalogFileTemp)
	catalogFilePath := path.Join(root, catalogFile)
	gobs, e := os.OpenFile(catalogFileTempPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if e != nil {
		Logger.Fatal(e)
	}

	status, e := gobs.Stat()
	if e != nil {
		Logger.Fatal(e)
	} else {
		c.Modified = status.ModTime()
	}

	encoder := gob.NewEncoder(gobs)
	e = encoder.Encode(infos)
	if e != nil {
		Logger.Fatal(e)
	}

	e = gobs.Close()
	if e != nil {
		Logger.Fatal(e)
	}

	e = os.Rename(catalogFileTempPath, catalogFilePath)
	if e != nil {
		Logger.Fatal(e)
	}
}

func shouldSkipFile(pathname string, info os.FileInfo) bool {
	basename := path.Base(pathname)
	return basename == "" || basename[0] == '.' || info.Size() == 0
}

func (c *Catalog) buildCatalogFromWalk(root string) {
	if c.buildCatalogFromWalkInProgress {
		return
	}
	c.buildCatalogFromWalkInProgress = true
	defer func() {
		c.buildCatalogFromWalkInProgress = false
	}()

	// Log the walk progress periodically so that the operator knows what’s going
	// on.
	count := 0
	timerFrequency := 1 * time.Second
	timer := time.NewTimer(timerFrequency)

	newItems := ItemInfos{}
	e := filepath.Walk(root,
		func(pathname string, info os.FileInfo, e error) error {
			if e != nil {
				Logger.Printf("%q: %s", pathname, e)
				return e
			}
			if shouldSkipFile(pathname, info) || info.Mode().IsDir() || !info.Mode().IsRegular() {
				return nil
			}

			input, e := os.Open(pathname)
			if e != nil {
				Logger.Printf("%q: %s", pathname, e)
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
				newItems = append(newItems, &itemInfo)

				count++
				select {
				case _, ok := <-timer.C:
					if ok {
						Logger.Printf("Processed %v items", count)
						timer.Reset(timerFrequency)
					} else {
						Logger.Printf("Channel closed?!")
					}
				default:
					// Do nothing.
				}
			}

			return nil
		})

	if e != nil {
		Logger.Print(e)
	}
	c.buildCatalogFromWalkInProgress = false
	c.writeCatalog(root, newItems)
	c.ItemInfos = newItems
}

func (c *Catalog) BuildCatalog(root string) {
	if !IsFileNewestInDirectory(root, catalogFile) {
		c.buildCatalogFromWalk(root)
		return
	}

	gobs, e := os.Open(path.Join(root, catalogFile))
	if e != nil {
		c.buildCatalogFromWalk(root)
		return
	}
	defer gobs.Close()

	status, e := gobs.Stat()
	if e != nil {
		c.buildCatalogFromWalk(root)
		return
	}

	modified := status.ModTime()
	if c.Modified.IsZero() || c.Modified.Before(modified) {
		c.readCatalog(gobs)
		c.Modified = modified
		return
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
		Logger.Print(e)
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
