// Copyright 2019 by Chris Palmer (https://noncombatant.org), and released under
// the terms of the GNU GPL3. See web/index.html for more information.

package main

import (
	"id3"
	"path/filepath"
	"strings"
)

type ItemInfo struct {
	Pathname           string    `json:"pathname"`
	Album              string    `json:"album"`
	Artist             string    `json:"artist"`
	Name               string    `json:"name"`
	Disc               string    `json:"disc"`
	Track              string    `json:"track"`
	Year               string    `json:"year"`
	Genre              string    `json:"genre"`
	NormalizedPathname string    `json:"-"`
	NormalizedAlbum    string    `json:"-"`
	NormalizedArtist   string    `json:"-"`
	NormalizedName     string    `json:"-"`
	NormalizedDisc     string    `json:"-"`
	NormalizedTrack    string    `json:"-"`
	NormalizedYear     string    `json:"-"`
	NormalizedGenre    string    `json:"-"`
	ModTime            string    `json:"-"`
	File               *id3.File `json:"-"`
}

type ItemInfos []*ItemInfo

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
		i.Name = RemoveBasenameExtension(i.Name)
	}
}

func (i *ItemInfo) fillMetadata() {
	i.fillMetadataFromPathname()

	if i.File != nil {
		// Quirk: I prefer to get the album and artist from `Pathname`. This makes .../Compilations/Jock Jams/... work.
		//i.File.Album = strings.TrimSpace(i.File.Album)
		//if i.File.Album != "" {
		//	i.Album = i.File.Album
		//}
		//i.File.Artist = strings.TrimSpace(i.File.Artist)
		//if i.File.Artist != "" {
		//	i.Artist = i.File.Artist
		//}
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
	i.NormalizedDisc = ExtractDigits(i.Disc)
	i.NormalizedTrack = ExtractDigits(i.Track)
	i.NormalizedYear = ExtractDigits(i.Year)
	i.NormalizedGenre = normalizeStringForSearch(i.Genre)
}
