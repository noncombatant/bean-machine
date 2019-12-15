// Copyright 2019 by Chris Palmer (https://noncombatant.org), and released under
// the terms of the GNU GPL3. See help.html for more information.

package main

import (
	"fmt"
	"id3"
	"path"
	"path/filepath"
	"strings"
	"time"
)

type ItemInfo struct {
	Pathname string
	Album    string
	Artist   string
	Name     string
	Disc     string
	Track    string
	Year     string
	Genre    string
	ModTime  time.Time
	File     *id3.File
}

func (i *ItemInfo) normalize() {
	if i.File != nil {
		i.File.Album = strings.TrimSpace(i.File.Album)
		if i.File.Album != "" {
			i.Album = i.File.Album
		}
		i.File.Artist = strings.TrimSpace(i.File.Artist)
		if i.File.Artist != "" {
			i.Artist = i.File.Artist
		}
		i.File.Name = strings.TrimSpace(i.File.Name)
		if i.File.Name != "" {
			i.Name = i.File.Name
		}
		if i.File.Disc != "" {
			i.Disc = i.File.Disc
		}
		if i.File.Track != "" {
			i.Track = i.File.Track
		}
		if i.File.Year != "" {
			i.Year = i.File.Year
		}
		i.File.Genre = strings.TrimSpace(i.File.Genre)
		if i.File.Genre != "" {
			i.Genre = i.File.Genre
		}
	}

	if i.Artist == "" || i.Album == "" || i.Name == "" {
		// Get info from pathname, assuming format:
		// "AC_DC/Back In Black/1-01 Hells Bells.m4a"
		//     performer/album/disc#-track# name
		parts := strings.Split(i.Pathname, string(filepath.Separator))
		length := len(parts)
		if i.Artist == "" && length > 2 {
			i.Artist = parts[length-3]
		}
		if i.Album == "" && length > 1 {
			i.Album = parts[length-2]
		}
		if i.Name == "" && length > 0 {
			i.Name = removeFileExtension(parts[length-1])
		}
	}

	if i.Artist == "" {
		i.Artist = "Unknown Artist"
	}
	if i.Album == "" {
		i.Album = "Unknown Album"
	}
	if i.Name == "" {
		i.Name = "Unknown Item"
	}
	if i.Disc == "" {
		i.Disc = "1"
	}
	if i.Track == "" {
		i.Track = "1"
	}

	i.Disc = normalizeNumericString(i.Disc)
	i.Track = normalizeNumericString(i.Track)
	i.Year = normalizeNumericString(i.Year)

	// Now, remove redundant information: If the ID3 info is identical to the
	// info contained in the pathname, remove it (to save file size).
	x := path.Base(i.Pathname)
	mp3Name := i.Name + ".mp3"
	track := fmt.Sprintf("%02d", parseInt(i.Track))
	if mp3Name == x || track+" "+mp3Name == x || i.Disc+"-"+track+" "+mp3Name == x {
		i.Name = ""
	}
	x = path.Dir(i.Pathname)
	if i.Album == path.Base(x) {
		i.Album = ""
	}
	if i.Artist == path.Base(path.Dir(x)) {
		i.Artist = ""
	}
}

// TODO: Delete this once the new server-side search thing is committed.
func (i *ItemInfo) ToTSV() string {
	i.normalize()
	year, month, day := i.ModTime.Date()
	modTime := fmt.Sprintf("%04d-%02d-%02d", year, month, day)
	return fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s",
		replaceTSVMetacharacters(escapePathname(i.Pathname)),
		replaceTSVMetacharacters(i.Album),
		replaceTSVMetacharacters(i.Artist),
		replaceTSVMetacharacters(i.Name),
		replaceTSVMetacharacters(i.Disc),
		replaceTSVMetacharacters(i.Track),
		replaceTSVMetacharacters(i.Year),
		replaceTSVMetacharacters(i.Genre),
		modTime)
}
