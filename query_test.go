package main

import (
	"testing"
)

// TODO: add test for "-noodle" and -thing:"noodle" and thing:"-noodle"

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

// TODO: Make this more complete.
func TestNormalizeStringForSearch(t *testing.T) {
	input := "Monáe"
	expected := "monae"
	received := normalizeStringForSearch(input)
	if received != expected {
		t.Errorf("%q != %q", expected, received)
	}
	t.Logf("%q == %q", expected, received)
}
