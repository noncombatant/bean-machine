// Copyright 2019 by Chris Palmer (https://noncombatant.org)
// SPDX-License-Identifier: GPL-3.0

package main

import (
	"strings"
)

func normalizeStringForSearch(s string) string {
	normalized := removeAccents(s)
	return strings.ToLower(normalized)
}

func matchItem(info *itemInfo, queries []Query) bool {
	for _, query := range queries {
		matched := false
		if query.Keyword == "path" || query.Keyword == "pathname" {
			matched = strings.Contains(info.NormalizedPathname, query.Term)
		} else if query.Keyword == "album" {
			matched = strings.Contains(info.NormalizedAlbum, query.Term)
		} else if query.Keyword == "artist" {
			matched = strings.Contains(info.NormalizedArtist, query.Term)
		} else if query.Keyword == "name" {
			matched = strings.Contains(info.NormalizedName, query.Term)
		} else if query.Keyword == "disc" {
			matched = strings.Contains(info.NormalizedDisc, query.Term)
		} else if query.Keyword == "track" {
			matched = strings.Contains(info.NormalizedTrack, query.Term)
		} else if query.Keyword == "year" {
			matched = strings.Contains(info.NormalizedYear, query.Term)
		} else if query.Keyword == "genre" {
			matched = strings.Contains(info.NormalizedGenre, query.Term)
		} else if query.Keyword == "mtime" || query.Keyword == "added" {
			matched = strings.Contains(info.ModTime, query.Term)
		} else {
			if strings.Contains(info.NormalizedPathname, query.Term) ||
				strings.Contains(info.NormalizedAlbum, query.Term) ||
				strings.Contains(info.NormalizedArtist, query.Term) ||
				strings.Contains(info.NormalizedName, query.Term) ||
				strings.Contains(info.NormalizedDisc, query.Term) ||
				strings.Contains(info.NormalizedTrack, query.Term) ||
				strings.Contains(info.NormalizedYear, query.Term) ||
				strings.Contains(info.NormalizedGenre, query.Term) ||
				strings.Contains(info.ModTime, query.Term) {
				matched = true
			}
		}
		if (matched && query.Negated) || (!matched && !query.Negated) {
			return false
		}
	}
	return true
}

func matchItems(infos itemInfos, rawQuery string) itemInfos {
	query := strings.TrimSpace(normalizeStringForSearch(rawQuery))
	queries := reconstructQueries(parseTerms(query))
	results := itemInfos{}
	for _, info := range infos {
		if matchItem(&info, queries) {
			results = append(results, info)
		}
	}
	return results
}
