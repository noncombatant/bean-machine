// Copyright 2019 by Chris Palmer (https://noncombatant.org), and released under
// the terms of the GNU GPL3. See help.html for more information.

package main

import (
	"fmt"
	"id3"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
)

func buildCatalog(root string) {
	assertValidRootPathname(root)

	if os.PathSeparator == root[len(root)-1] {
		root = root[:len(root)-1]
	}
	pathname := path.Join(root, "catalog.tsv")
	output, e := os.Create(pathname)
	if e != nil {
		fmt.Fprintf(os.Stderr, "Could not create %q: %s\n", pathname, e)
		os.Exit(1)
	}
	defer func() {
		e := output.Close()
		if e != nil {
			fmt.Fprintf(os.Stderr, "%q: %v\n", pathname, e)
		}
	}()

	e = filepath.Walk(root,
		func(pathname string, info os.FileInfo, e error) error {
			if e != nil {
				fmt.Fprintf(os.Stderr, "%q: %s\n", pathname, e)
				return e
			}
			if shouldSkipFile(pathname, info) {
				fmt.Fprintf(os.Stderr, "Skipping %q\n", pathname)
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
				fmt.Fprintf(os.Stderr, "%q: %s\n", pathname, e)
				return nil
			}
			defer input.Close()

			webPathname := pathname[len(root)+1:]
			if isAudioPathname(pathname) || isVideoPathname(pathname) {
				itemInfo := ItemInfo{Pathname: webPathname}
				if isAudioPathname(pathname) {
					itemInfo.File, e = id3.Read(input)
					if e != nil {
						//fmt.Fprintf(os.Stderr, "%q: %v\n", pathname, e)
					}
				}
				fileInfo, e := os.Stat(pathname)
				if e == nil {
					itemInfo.ModTime = fileInfo.ModTime()
				}
				fmt.Fprintf(output, "%s\n", itemInfo.ToTSV())
			}

			return nil
		})

	if e != nil {
		fmt.Fprintf(os.Stderr, "Problem walking %q: %s\n", root, e)
	}
}

func buildMediaIndex(pathname string) {
	infos, e := ioutil.ReadDir(pathname)
	if e != nil {
		log.Fatal(e)
	}

	index, e := os.OpenFile(pathname+"/media.html", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
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
