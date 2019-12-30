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

type Query struct {
	Term    string
	Negated bool
}

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

func matchItem(info *ItemInfo, queries []Query) bool {
	for _, query := range queries {
		matched := false
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
		if (matched && query.Negated) || (!matched && !query.Negated) {
			return false
		}
	}
	return true
}

func matchItems(infos []*ItemInfo, query []string) []*ItemInfo {
	queries := make([]Query, len(query))
	for i, term := range query {
		if term[0] == '-' {
			term = term[1:]
			queries[i] = Query{Term: normalizeStringForSearch(term), Negated: true}
		} else {
			queries[i] = Query{Term: normalizeStringForSearch(term), Negated: false}
		}
	}

	results := []*ItemInfo{}
	for _, info := range infos {
		if matchItem(info, queries) {
			results = append(results, info)
		}
	}

	return results
}
