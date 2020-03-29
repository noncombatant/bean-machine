// Copyright 2019 by Chris Palmer (https://noncombatant.org), and released under
// the terms of the GNU GPL3. See help.html for more information.

package main

import (
	"encoding/gob"
	"fmt"
	"id3"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
)

var (
	catalog = []*ItemInfo{}
)

const (
	catalogFile = "catalog.gobs"
)

func buildCatalogFromGobs(gobs *os.File) {
	log.Print("buildCatalogFromGobs: start.")
	decoder := gob.NewDecoder(gobs)
	for {
		var info ItemInfo
		e := decoder.Decode(&info)
		if e != nil {
			if e == io.EOF {
				log.Printf("buildCatalogFromGobs: Completed. %v items.", len(catalog))
				return
			} else {
				log.Fatal("decode error 1:", e)
			}
		}
		catalog = append(catalog, &info)
	}
}

func buildCatalogFromWalk(root string) {
	log.Print("buildCatalogFromWalk: Start.")
	log.Print("buildCatalogFromWalk: This might take a while.")

	gobs, e := os.OpenFile(path.Join(root, string(os.PathSeparator), catalogFile), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if e != nil {
		log.Fatal(e)
	}
	defer gobs.Close()
	encoder := gob.NewEncoder(gobs)

	count := 0
	e = filepath.Walk(root,
		func(pathname string, info os.FileInfo, e error) error {
			if e != nil {
				log.Printf("buildCatalogFromWalk: %q: %s", pathname, e)
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
				log.Printf("buildCatalogFromWalk: %q: %s", pathname, e)
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
				catalog = append(catalog, &itemInfo)
				e := encoder.Encode(itemInfo)
				if e != nil {
					log.Fatal(e)
				}

				count++
				if count%1000 == 0 {
					log.Printf("buildCatalogFromWalk: Processed %v items", count)
				}
			}

			return nil
		})

	if e != nil {
		log.Printf("buildCatalogFromWalk: Problem walking %q: %s", root, e)
	}
	log.Printf("buildCatalogFromWalk: Completed. %v items.", len(catalog))
}

func buildCatalog(root string) {
	gobs, e := os.Open(path.Join(root, string(os.PathSeparator), catalogFile))
	if e != nil {
		buildCatalogFromWalk(root)
		return
	}
	defer gobs.Close()
	buildCatalogFromGobs(gobs)
}

func shouldBuildMediaIndex(pathname string, infos []os.FileInfo) bool {
	index, e := os.Open(path.Join(pathname, string(os.PathSeparator), "media.html"))
	if e != nil {
		return true
	}
	defer index.Close()

	indexStatus, e := index.Stat()
	if e != nil {
		log.Fatal(e)
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
		log.Fatal(e)
	}

	if !shouldBuildMediaIndex(pathname, infos) {
		return
	}

	index, e := os.OpenFile(path.Join(pathname, string(os.PathSeparator), "media.html"), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if e != nil {
		log.Fatal(e)
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
