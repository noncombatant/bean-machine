// Copyright 2019 by Chris Palmer (https://noncombatant.org), and released under
// the terms of the GNU GPL3. See help.html for more information.

package main

import (
	"strings"
	"unicode"

	"golang.org/x/text/secure/precis"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

var (
	// Borrowed from
	// https://stackoverflow.com/questions/26722450/remove-diacritics-using-go.
	loosecompare = precis.NewIdentifier(
		precis.AdditionalMapping(func() transform.Transformer {
			return transform.Chain(norm.NFD, transform.RemoveFunc(func(r rune) bool {
				return unicode.Is(unicode.Mn, r)
			}))
		}),
		precis.Norm(norm.NFC), // This is the default; be explicit though.
	)
)

func normalizeStringForSearch(s string) string {
	normalized, e := loosecompare.String(s)
	if e != nil {
		// TODO: I'm probably not using precis right... "precis: disallowed rune
		// encountered" for every string.
		//Logger.Printf("%q: %v", s, e)
		normalized = s
	}
	return strings.ToLower(normalized)
}

func matchItem(info *ItemInfo, queries []Query) bool {
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

func matchItems(infos []*ItemInfo, rawQuery string) []*ItemInfo {
	rawQuery = strings.TrimSpace(normalizeStringForSearch(rawQuery))
	queries := ReconstructQueries(ParseTerms(rawQuery))
	Logger.Print(queries)

	results := []*ItemInfo{}
	for _, info := range infos {
		if matchItem(info, queries) {
			results = append(results, info)
		}
	}

	return results
}
