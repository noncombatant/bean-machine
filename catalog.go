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
	"time"
)

var (
	catalog         = []*ItemInfo{}
	lastReadCatalog = time.Time{}
	// `buildCatalogFromWalk` could be invoked (indirectly) through either
	// `buildCatalog` or `serveApp`. To avoid having multiple invocations walk
	// twice and stomp on the `catalogFile`, we maintain this sentinel.
	buildCatalogFromWalkInProgress = false
)

const (
	catalogFile = "catalog.gobs"
)

func buildCatalogFromGobs(gobs *os.File, modTime time.Time) {
	Logger.Print("running")
	decoder := gob.NewDecoder(gobs)
	newCatalog := []*ItemInfo{}
	for {
		var info ItemInfo
		e := decoder.Decode(&info)
		if e != nil {
			if e == io.EOF {
				break
			} else {
				Logger.Fatal("decode error 1:", e)
			}
		}
		newCatalog = append(newCatalog, &info)
	}
	catalog = newCatalog
	lastReadCatalog = modTime
}

func buildCatalogFromWalk(root string) {
	if buildCatalogFromWalkInProgress {
		return
	}
	buildCatalogFromWalkInProgress = true
	defer func() {
		buildCatalogFromWalkInProgress = false
	}()
	Logger.Print("Start. This might take a while.")

	gobs, e := os.OpenFile(path.Join(root, catalogFile), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if e != nil {
		Logger.Fatal(e)
	}
	defer gobs.Close()

	status, e := gobs.Stat()
	if e != nil {
		Logger.Printf("Can't Stat %q: %v", catalogFile, e)
	} else {
		lastReadCatalog = status.ModTime()
	}

	encoder := gob.NewEncoder(gobs)

	// Log the walk progress periodically so that the person knows what’s going
	// on.
	count := 0
	timerFrequency := 1 * time.Second
	timer := time.NewTimer(timerFrequency)

	newCatalog := []*ItemInfo{}
	e = filepath.Walk(root,
		func(pathname string, info os.FileInfo, e error) error {
			if e != nil {
				Logger.Printf("%q: %s", pathname, e)
				return nil
			}
			if shouldSkipFile(pathname, info) {
				return nil
			}
			if info.Mode().IsDir() {
				buildMediaIndex(pathname)
				return nil
			}
			if !info.Mode().IsRegular() {
				return nil
			}

			input, e := os.Open(pathname)
			if e != nil {
				Logger.Printf("%q: %s", pathname, e)
				return nil
			}
			defer input.Close()

			if isAudioPathname(pathname) || isVideoPathname(pathname) {
				webPathname := pathname[len(root)+1:]
				itemInfo := ItemInfo{Pathname: webPathname}
				itemInfo.File, _ = id3.Read(input)
				time := info.ModTime()
				itemInfo.ModTime = fmt.Sprintf("%04d-%02d-%02d", time.Year(), time.Month(), time.Day())
				itemInfo.fillMetadata()
				newCatalog = append(newCatalog, &itemInfo)
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
	catalog = newCatalog
	Logger.Printf("Completed. %v items.", len(catalog))
}

func isFileNewestInDirectory(directoryName, baseName string) bool {
	file, e := os.Open(path.Join(directoryName, baseName))
	if e != nil {
		return false
	}
	defer file.Close()

	status, e := file.Stat()
	if e != nil {
		return false
	}
	modTime := status.ModTime()

	infos, e := ioutil.ReadDir(directoryName)
	if e != nil {
		return false
	}

	for _, info := range infos {
		if info.IsDir() && modTime.Before(info.ModTime()) {
			return false
		}
	}
	return true
}

func buildCatalog(root string) {
	if !isFileNewestInDirectory(root, catalogFile) {
		buildCatalogFromWalk(root)
		return
	}

	gobs, e := os.Open(path.Join(root, catalogFile))
	if e != nil {
		buildCatalogFromWalk(root)
		return
	}
	defer gobs.Close()

	status, e := gobs.Stat()
	if e != nil {
		buildCatalogFromWalk(root)
		return
	}

	modTime := status.ModTime()
	if lastReadCatalog.IsZero() || lastReadCatalog.Before(modTime) {
		buildCatalogFromGobs(gobs, modTime)
		return
	}
}

func shouldBuildMediaIndex(pathname string, infos []os.FileInfo) bool {
	index, e := os.Open(path.Join(pathname, "media.html"))
	if e != nil {
		return true
	}
	defer index.Close()

	indexStatus, e := index.Stat()
	if e != nil {
		Logger.Fatal(e)
	}

	time := indexStatus.ModTime()
	for _, info := range infos {
		name := info.Name()
		if isImagePathname(name) || isDocumentPathname(name) {
			if time.Before(info.ModTime()) {
				return true
			}
		}
	}
	return false
}

func buildMediaIndex(pathname string) {
	infos, e := ioutil.ReadDir(pathname)
	if e != nil {
		Logger.Fatal(e)
	}

	if !shouldBuildMediaIndex(pathname, infos) {
		return
	}

	index, e := os.OpenFile(path.Join(pathname, "media.html"), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if e != nil {
		Logger.Fatal(e)
	}
	defer index.Close()

	header := `
<!DOCTYPE html>
<meta charset="UTF-8"/>
<meta name="viewport" content="width=device-width, initial-scale=1"/>
<title>%s</title>
<style>
body {
  line-height: 1.6;
  font-size: 16px;
  color: #222;
  font-family: system-ui;
}
img {
  border: 1px solid black;
  max-width: 100%%;
  height: auto;
}
</style>
`
	fmt.Fprintf(index, header, path.Base(pathname))

	for _, info := range infos {
		name := info.Name()
		if isImagePathname(name) {
			fmt.Fprintf(index, "<img src=\"%s\"/>\n", escapeDoubleQuotes(name))
		} else if isDocumentPathname(name) {
			name = escapeDoubleQuotes(name)
			fmt.Fprintf(index, "<li><a href=\"%s\">%s</a></li>\n", name, name)
		}
	}
}
