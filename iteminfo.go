// Copyright 2019 by Chris Palmer (https://noncombatant.org), and released under
// the terms of the GNU GPL3. See help.html for more information.

package main

import (
	"fmt"
	"id3"
	"path/filepath"
	"strings"
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
	ModTime            string
	File               *id3.File
}

type Catalog []*ItemInfo

func (i *ItemInfo) ToJSON() string {
	return fmt.Sprintf(`{"pathname":%q,
"album":%q,
"artist":%q,
"name":%q,
"disc":%d,
"track":%d,
"year":%d,
"genre":%q}`,
		i.Pathname, i.Album, i.Artist, i.Name, atoi(i.NormalizedDisc), atoi(i.NormalizedTrack), atoi(i.NormalizedYear), i.Genre)
}

func getDiscAndTrackFromBasename(basename string) (string, string, string) {
	parts := strings.SplitN(basename, " ", 2)
	if len(parts) != 2 {
		return "", "", basename
	}

	rest := parts[1]

	parts = strings.Split(parts[0], "-")
	if len(parts) > 2 {
		return "", "", basename
	}
	if len(parts) == 2 {
		return parts[0], parts[1], rest
	}
	return "", parts[0], rest
}

// Get info from pathname, assuming format:
// ".../AC_DC/Back In Black/1-01 Hells Bells.m4a"
//     performer/album/disc#-track# name
func (i *ItemInfo) fillMetadataFromPathname() {
	parts := strings.Split(i.Pathname, string(filepath.Separator))
	length := len(parts)
	if length > 2 {
		i.Artist = parts[length-3]
	}
	if length > 1 {
		i.Album = parts[length-2]
	}
	if length > 0 {
		i.Disc, i.Track, i.Name = getDiscAndTrackFromBasename(parts[length-1])
		i.Name = removeFileExtension(i.Name)
	}
}

func (i *ItemInfo) fillMetadata() {
	i.fillMetadataFromPathname()

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
	i.NormalizedDisc = extractNumericString(i.Disc)
	i.NormalizedTrack = extractNumericString(i.Track)
	i.NormalizedYear = extractNumericString(i.Year)
	i.NormalizedGenre = normalizeStringForSearch(i.Genre)
}
