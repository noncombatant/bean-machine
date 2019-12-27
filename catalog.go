// Copyright 2019 by Chris Palmer (https://noncombatant.org), and released under
// the terms of the GNU GPL3. See help.html for more information.

package main

import (
	"bufio"
	"fmt"
	"id3"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
)

var (
	catalog = []*ItemInfo{}
)

func buildCatalogFromTSV(tsv *os.File) {
	log.Print("buildCatalogFromTSV: start.")
	reader := bufio.NewReader(tsv)
	for {
		line, e := reader.ReadString('\n')
		if line == "" || e != nil {
			return
		}
		info := ItemInfoFromTSV(line)
		if info == nil {
			log.Printf("Bad record: %q", line)
			continue
		}
		catalog = append(catalog, info)
	}
}

func buildCatalogFromNetstrings(netstrings *os.File) {
	log.Print("buildCatalogFromNetstringNetstrings: start.")
	reader := bufio.NewReader(netstrings)
	for {
		line, e := reader.ReadString('\n')
		if line == "" || e != nil {
			return
		}
		info := ItemInfoFromTSV(line)
		if info == nil {
			log.Printf("Bad record: %q", line)
			continue
		}
		catalog = append(catalog, info)
	}
}

func buildCatalogFromWalk(root string) {
	log.Print("buildCatalogFromWalk: start.")

	tsv, e := os.OpenFile(path.Join(root, string(os.PathSeparator), "catalog.tsv"), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if e != nil {
		log.Fatal(e)
	}
	defer tsv.Close()

	e = filepath.Walk(root,
		func(pathname string, info os.FileInfo, e error) error {
			if e != nil {
				log.Printf("buildCatalog: %q: %s\n", pathname, e)
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
				log.Printf("buildCatalog: %q: %s\n", pathname, e)
				return nil
			}
			defer input.Close()

			if isAudioPathname(pathname) || isVideoPathname(pathname) {
				webPathname := pathname[len(root)+1:]
				itemInfo := ItemInfo{Pathname: webPathname}
				itemInfo.File, _ = id3.Read(input)
				itemInfo.ModTime = info.ModTime()
				itemInfo.fillMetadata()
				catalog = append(catalog, &itemInfo)
				_, e := tsv.WriteString(itemInfo.ToTSV())
				if e != nil {
					log.Fatal(e)
				}
			}

			return nil
		})

	if e != nil {
		log.Printf("buildCatalog: Problem walking %q: %s\n", root, e)
	}
	log.Printf("buildCatalog: completed. %v items.", len(catalog))
}

func buildCatalog(root string) {
	tsv, e := os.Open(path.Join(root, string(os.PathSeparator), "catalog.tsv"))
	if e != nil {
		buildCatalogFromWalk(root)
		return
	}
	defer tsv.Close()
	buildCatalogFromTSV(tsv)
}

func buildMediaIndex(pathname string) {
	infos, e := ioutil.ReadDir(pathname)
	if e != nil {
		log.Fatal(e)
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

	for _, f := range infos {
		name := f.Name()
		if isImagePathname(name) {
			fmt.Fprintf(index, "<img src=\"%s\"/>\n", escapeDoubleQuotes(name))
		}
	}
	for _, f := range infos {
		name := f.Name()
		if isDocumentPathname(name) {
			name = escapeDoubleQuotes(name)
			fmt.Fprintf(index, "<li><a href=\"%s\">%s</a></li>\n", name, name)
		}
	}
}
