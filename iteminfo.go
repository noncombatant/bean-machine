// Copyright 2019 by Chris Palmer (https://noncombatant.org)
// SPDX-License-Identifier: GPL-3.0

package main

import (
	"id3"
	"net/url"
	"path/filepath"
	"regexp"
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

type ItemInfos []ItemInfo

// This terrible hack is an alternative to separately `url.PathEscape`ing each
// pathname component and then re-joining them. That would be conceptually
// better but this is expedient.
func pathnameEscape(pathname string) string {
	// `PathEscape` uses capital hex, hence "%2F".
	return strings.ReplaceAll(url.PathEscape(pathname), "%2F", "/")
}

var (
	discTrackAndNameMatcher = regexp.MustCompile(`^\s*(\d*)?-?(\d*)?\s+(.*)$`)
)

func getDiscTrackAndNameFromBasename(basename string) (string, string, string) {
	submatches := discTrackAndNameMatcher.FindSubmatch([]byte(basename))
	if len(submatches) != 4 {
		return "", "", basename
	}
	if len(submatches[1]) > 0 && len(submatches[2]) == 0 {
		return "", string(submatches[1]), string(submatches[3])
	}
	return string(submatches[1]), string(submatches[2]), string(submatches[3])
}

// Sets fields of `i` from `i.Pathname`, assuming the format:
//
//	".../AC_DC/Back In Black/1-01 Hells Bells.m4a"
//	     performer/album/disc#-track# name
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
		i.Disc, i.Track, i.Name = getDiscTrackAndNameFromBasename(parts[length-1])
		i.Name = removeBasenameExtension(i.Name)
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
		i.File.Disc = strings.TrimSpace(i.File.Disc)
		if i.File.Disc != "" {
			i.Disc = i.File.Disc
		}
		i.File.Track = strings.TrimSpace(i.File.Track)
		if i.File.Track != "" {
			i.Track = i.File.Track
		}
		i.File.Year = strings.TrimSpace(i.File.Year)
		if i.File.Year != "" {
			i.Year = i.File.Year
		}
		i.File.Genre = strings.TrimSpace(i.File.Genre)
		if i.File.Genre != "" {
			i.Genre = i.File.Genre
		}
	}

	i.Pathname = pathnameEscape(i.Pathname)
	i.normalize()
}

func (i *ItemInfo) normalize() {
	i.NormalizedPathname = normalizeStringForSearch(i.Pathname)
	i.NormalizedAlbum = normalizeStringForSearch(i.Album)
	i.NormalizedArtist = normalizeStringForSearch(i.Artist)
	i.NormalizedName = normalizeStringForSearch(i.Name)
	i.NormalizedDisc = extractDigits(i.Disc)
	i.NormalizedTrack = extractDigits(i.Track)
	i.NormalizedYear = extractDigits(i.Year)
	i.NormalizedGenre = normalizeStringForSearch(i.Genre)
}
