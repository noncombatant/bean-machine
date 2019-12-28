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
		//log.Printf("normalizeStringForSearch: %q: %v", s, e)
		normalized = s
	}
	return strings.ToLower(normalized)
}

func matchItem(info *ItemInfo, query []string) bool {
	for _, term := range query {
		if strings.Contains(info.NormalizedPathname, term) ||
			strings.Contains(info.NormalizedAlbum, term) ||
			strings.Contains(info.NormalizedArtist, term) ||
			strings.Contains(info.NormalizedName, term) ||
			strings.Contains(info.NormalizedDisc, term) ||
			strings.Contains(info.NormalizedTrack, term) ||
			strings.Contains(info.NormalizedYear, term) ||
			strings.Contains(info.NormalizedGenre, term) ||
			strings.Contains(info.ModTime, term) {
			continue
		}
		return false
	}
	return true
}

func matchItems(infos []*ItemInfo, query []string) []*ItemInfo {
	for i, term := range query {
		query[i] = normalizeStringForSearch(term)
	}

	results := []*ItemInfo{}
	for _, info := range infos {
		if matchItem(info, query) {
			results = append(results, info)
		}
	}

	return results
}
