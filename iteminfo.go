// Copyright 2019 by Chris Palmer (https://noncombatant.org), and released under
// the terms of the GNU GPL3. See help.html for more information.

package main

import (
	"fmt"
	"id3"
	"path/filepath"
	"strings"
	"time"
)

type ItemInfo struct {
	Pathname           string
	Album              string
	Artist             string
	Name               string
	Disc               string
	Track              string
	Year               string
	Genre              string
	NormalizedPathname string
	NormalizedAlbum    string
	NormalizedArtist   string
	NormalizedName     string
	NormalizedDisc     string
	NormalizedTrack    string
	NormalizedYear     string
	NormalizedGenre    string
	// TODO: Decide on using this in search, or not. If so, make it a YYYY-MM-DD string; if not, delete it.
	ModTime time.Time
	File    *id3.File
}

func escape(s string) string {
	s = strings.ReplaceAll(s, "\t", "\\t")
	s = strings.ReplaceAll(s, "\n", "\\n")
	return s
}

// TODO: Replace ToTSV, FromTSV, and escape with something like netstrings, so
// that no escaping is necessary.
func (i *ItemInfo) ToTSV() string {
	return fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
		escape(i.Pathname), escape(i.Album), escape(i.Artist), escape(i.Name), escape(i.Disc), escape(i.Track), escape(i.Year), escape(i.Genre))
}

func ItemInfoFromTSV(tsv string) *ItemInfo {
	if len(tsv) == 0 {
		return nil
	}

	if tsv[len(tsv)-1] == '\n' {
		tsv = tsv[:len(tsv)-1]
	}
	fields := strings.Split(tsv, "\t")
	if len(fields) != 8 {
		return nil
	}

	info := ItemInfo{
		Pathname: fields[0],
		Album:    fields[1],
		Artist:   fields[2],
		Name:     fields[3],
		Disc:     fields[4],
		Track:    fields[5],
		Year:     fields[6],
		Genre:    fields[7],
	}
	info.Normalize()
	return &info
}

func (i *ItemInfo) ToJSON() string {
	return fmt.Sprintf(`{"pathname":%q,
"album":%q,
"artist":%q,
"name":%q,
"disc":%q,
"track":%q,
"year":%q,
"genre":%q}`,
		i.Pathname, i.Album, i.Artist, i.Name, i.Disc, i.Track, i.Year, i.Genre)
}

// Get info from pathname, assuming format:
// ".../AC_DC/Back In Black/1-01 Hells Bells.m4a"
//     performer/album/disc#-track# name
func (i *ItemInfo) fillMetadataFromPathname() {
	if i.Artist == "" || i.Album == "" || i.Name == "" {
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
	// TODO: Fill in disc and track, too.
}

func (i *ItemInfo) fillMetadata() {
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

	// TODO: Do this second, then overlay ID3 on top. Simplify logic.
	i.fillMetadataFromPathname()

	// TODO: Do this first, then overlay pathname info on top. Simplify logic.
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

	i.Normalize()
}

func (i *ItemInfo) Normalize() {
	i.NormalizedPathname = normalizeStringForSearch(i.Pathname)
	i.NormalizedAlbum = normalizeStringForSearch(i.Album)
	i.NormalizedArtist = normalizeStringForSearch(i.Artist)
	i.NormalizedName = normalizeStringForSearch(i.Name)
	i.Disc = extractNumericString(i.Disc)
	i.Track = extractNumericString(i.Track)
	i.Year = extractNumericString(i.Year)
	i.NormalizedGenre = normalizeStringForSearch(i.Genre)
}
