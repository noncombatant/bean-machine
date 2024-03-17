// Copyright 2019 by Chris Palmer (https://noncombatant.org)
// SPDX-License-Identifier: GPL-3.0

package main

import (
	"compress/gzip"
	"encoding/gob"
	"fmt"
	"id3"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
)

type catalog struct {
	itemInfos
}

func (c *catalog) writeToFile(pathname string) error {
	w, e := os.Create(pathname)
	if e != nil {
		return e
	}
	zw, e := gzip.NewWriterLevel(w, 9)
	if e != nil {
		return e
	}
	if e := gob.NewEncoder(zw).Encode(c); e != nil {
		return e
	}
	if e := zw.Close(); e != nil {
		return e
	}
	return w.Close()
}

const (
	catalogBasename = "catalog.gobs.gz"
	eraseLine       = "\033[2K\r"
)

func shouldSkipFile(info os.FileInfo) bool {
	return info.Name() == "" || info.Name()[0] == '.' || info.Size() == 0 || info.Mode().IsDir() || !info.Mode().IsRegular()
}

func newCatalog(log *log.Logger, root string) (*catalog, error) {
	var c catalog
	previousDir := ""
	e := filepath.Walk(root,
		func(pathname string, info os.FileInfo, e error) error {
			if e != nil {
				log.Print(e)
				return e
			}
			if shouldSkipFile(info) {
				return nil
			}

			// If we have progressed to a new directory, print progress indicator.
			dir := path.Dir(path.Dir(pathname[len(root)+1:]))
			if dir != previousDir {
				fmt.Fprintf(os.Stdout, "%s%s", eraseLine, dir)
				previousDir = dir
			}

			if isAudioPathname(pathname) || isVideoPathname(pathname) {
				webPathname := pathname[len(root)+1:]
				itemInfo := itemInfo{Pathname: webPathname}

				input, e := os.Open(pathname)
				if e != nil {
					log.Print(e)
					return e
				}
				itemInfo.File, _ = id3.Read(input)
				if e := input.Close(); e != nil {
					log.Print(e)
					return e
				}

				time := info.ModTime()
				itemInfo.ModTime = fmt.Sprintf("%04d-%02d-%02d", time.Year(), time.Month(), time.Day())
				itemInfo.fillMetadata()
				c.itemInfos = append(c.itemInfos, itemInfo)
			}
			return nil
		})

	fmt.Fprintf(os.Stdout, "%s\n", eraseLine)
	return &c, e
}

func readCatalogFromFile(pathname string) (*catalog, error) {
	f, e := os.Open(pathname)
	if e != nil {
		return nil, e
	}
	zr, e := gzip.NewReader(f)
	if e != nil {
		return nil, e
	}
	var c catalog
	d := gob.NewDecoder(zr)
	if e := d.Decode(&c); e != nil && e != io.EOF {
		return nil, e
	}
	if e = zr.Close(); e != nil {
		return nil, e
	}
	if e = f.Close(); e != nil {
		return nil, e
	}
	return &c, nil
}
