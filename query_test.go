// Copyright 2020 by Chris Palmer (https://noncombatant.org), and released under
// the terms of the GNU GPL3. See web/index.html for more information.

package main

import (
	"testing"
)

var (
	rawTerms = `Foo bar kw:term kw2 : term2 -greeb graggle kw3: -"term 3"`

	expectedParsedTerms = []string{
		"Foo",
		"bar",
		"kw",
		":",
		"term",
		"kw2",
		":",
		"term2",
		"-",
		"greeb",
		"graggle",
		"kw3",
		":",
		"-",
		"term 3",
	}

	expectedQueries = []Query{
		{"", "Foo", false},
		{"", "bar", false},
		{"kw", "term", false},
		{"kw2", "term2", false},
		{"", "greeb", true},
		{"", "graggle", false},
		{"kw3", "term 3", true},
	}
)

func stringSlicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func testParseTermsHelper(t *testing.T) []string {
	terms := ParseTerms(rawTerms)
	if !stringSlicesEqual(expectedParsedTerms, terms) {
		t.Error(expectedParsedTerms, terms)
	}
	return terms
}

func TestParseTerms(t *testing.T) {
	testParseTermsHelper(t)
}

func TestReconstructQueries(t *testing.T) {
	terms := testParseTermsHelper(t)
	queries := ReconstructQueries(terms)
	if len(expectedQueries) != len(queries) {
		t.Errorf("mismatched lengths")
	}
	for i := range expectedQueries {
		if expectedQueries[i] != queries[i] {
			t.Errorf("%v != %v\n", expectedQueries[i], queries[i])
		}
	}
}

func TestNormalizeStringForSearch(t *testing.T) {
	input := []string{
		"Monáe",
		"monÁe",
		"MONÁE",
		"gürg",
		"ALLCAPS",
	}
	expected := []string{
		"monae",
		"monae",
		"monae",
		"gurg",
		"allcaps",
	}

	for i, v := range input {
		received := normalizeStringForSearch(v)
		ex := expected[i]
		if received != ex {
			t.Errorf("%q != %q", ex, received)
		}
		t.Logf("%q == %q", ex, received)
	}
}
