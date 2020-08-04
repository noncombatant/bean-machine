// Copyright 2019 by Chris Palmer (https://noncombatant.org), and released under
// the terms of the GNU GPL3. See help.html for more information.

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

func (c *Catalog) buildCatalogFromGobs(gobs *os.File, modified time.Time) {
	decoder := gob.NewDecoder(gobs)
	newInfos := ItemInfos{}
	for {
		var info ItemInfo
		e := decoder.Decode(&info)
		if e != nil {
			if e == io.EOF {
				break
			} else {
				Logger.Fatal("decode error:", e)
			}
		}
		newInfos = append(newInfos, &info)
	}
	c.ItemInfos = newInfos
	c.Modified = modified
}

func shouldSkipFile(pathname string, info os.FileInfo) bool {
	basename := path.Base(pathname)
	return "" == basename || '.' == basename[0] || 0 == info.Size()
}

func (c *Catalog) buildCatalogFromWalk(root string) {
	if c.buildCatalogFromWalkInProgress {
		return
	}
	c.buildCatalogFromWalkInProgress = true
	defer func() {
		c.buildCatalogFromWalkInProgress = false
	}()

	catalogFileTempPath := path.Join(root, catalogFileTemp)
	catalogFilePath := path.Join(root, catalogFile)
	gobs, e := os.OpenFile(catalogFileTempPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if e != nil {
		Logger.Fatal(e)
	}
	// We do `gobs.Close()` explicitly at the end.

	status, e := gobs.Stat()
	if e != nil {
		Logger.Printf("Can't Stat %q: %v", catalogFileTemp, e)
	} else {
		c.Modified = status.ModTime()
	}

	encoder := gob.NewEncoder(gobs)

	// Log the walk progress periodically so that the operator knows what’s going
	// on.
	count := 0
	timerFrequency := 1 * time.Second
	timer := time.NewTimer(timerFrequency)

	newItems := ItemInfos{}
	e = filepath.Walk(root,
		func(pathname string, info os.FileInfo, e error) error {
			if e != nil {
				Logger.Printf("%q: %s", pathname, e)
				return nil
			}
			if shouldSkipFile(pathname, info) || info.Mode().IsDir() || !info.Mode().IsRegular() {
				return nil
			}

			input, e := os.Open(pathname)
			if e != nil {
				Logger.Printf("%q: %s", pathname, e)
				return nil
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
				e := encoder.Encode(itemInfo)
				if e != nil {
					Logger.Fatal(e)
				}

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
		Logger.Printf("Problem walking %q: %s", root, e)
	}
	c.buildCatalogFromWalkInProgress = false
	gobs.Close()
	e = os.Rename(catalogFileTempPath, catalogFilePath)
	if e != nil {
		Logger.Fatal(e)
	}
	c.ItemInfos = newItems
	Logger.Printf("Completed. %v items.", len(c.ItemInfos))
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
		c.buildCatalogFromGobs(gobs, modified)
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
		builder.String()
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
