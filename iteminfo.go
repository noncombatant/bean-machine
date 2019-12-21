// Copyright 2019 by Chris Palmer (https://noncombatant.org), and released under
// the terms of the GNU GPL3. See help.html for more information.

package main

import (
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
	ModTime            time.Time
	File               *id3.File
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

	// TODO: Do this first, then overlay ID3 on top. Simplify logic.
	i.fillMetadataFromPathname()

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

	i.NormalizedPathname = normalizeStringForSearch(i.Pathname)
	i.NormalizedAlbum = normalizeStringForSearch(i.Album)
	i.NormalizedArtist = normalizeStringForSearch(i.Artist)
	i.NormalizedName = normalizeStringForSearch(i.Name)
	i.Disc = extractNumericString(i.Disc)
	i.Track = extractNumericString(i.Track)
	i.Year = extractNumericString(i.Year)
	i.NormalizedGenre = normalizeStringForSearch(i.Genre)
}
