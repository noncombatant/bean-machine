// Copyright 2020 by Chris Palmer (https://noncombatant.org), and released under
// the terms of the GNU GPL3. See help.html for more information.

package main

import (
	"testing"
)

func TestExtractDigits(t *testing.T) {
	type Expectation struct {
		Input string
		Output string
	}
	expectations := []Expectation{
		{"123409", "123409"},
		{"Hello 123409", "123409"},
		{"Hello 123409goat", "123409"},
		{"123409goat", "123409"},
		{"hello pumpkins?", ""},
		{"ABCDEF", ""},
		{"wow 3.14159", "3"},
	}

	for _, e := range expectations {
		r := ExtractDigits(e.Input)
		if e.Output != r {
			t.Errorf("%q != ExtractDigits(%q) (%q)\n", e.Output, e.Input, r)
		}
	}
}

func TestGetFileExtension(t *testing.T) {
	// TODO
}

func TestRemoveFileExtension(t *testing.T) {
	// TODO
}

func TestEscapeDoubleQuotes(t *testing.T) {
	// TODO
}

func TestIsStringInStrings(t *testing.T) {
	for _, s := range audioFormatExtensions {
		if !IsStringInStrings(s, audioFormatExtensions) {
			t.Errorf("Could not find %q", s)
		}
	}

	if IsStringInStrings(".FLAC", audioFormatExtensions) {
		t.Errorf("Found \".FLAC\" even though it's upper-case")
	}

	if IsStringInStrings("goat", audioFormatExtensions) {
		t.Errorf("Found \"goat\" even though it's not in `audioFormatExtensions`")
	}
}

func TestMustGetRandomBytes(t *testing.T) {
	bytes := make([]byte, 100)
	for i, _ := range bytes {
		bytes[i] = 42
	}

	MustGetRandomBytes(bytes)

	for _, b := range bytes {
		if b != 42 {
			return
		}
	}

	t.Errorf("All bytes were still 42")
}

func TestParseIntegerOr0(t *testing.T) {
	type Expectation struct {
		Input string
		Output int
	}
	expectations := []Expectation{
		{"0x00DADB0D", 0xdadb0d},
		{"0xdadb0d", 0xdadb0d},
		{"0xdeadbeef", 0},
		{"0x0", 0x0},
		{"42", 42},
		{"hello", 0},
		{"", 0},
		{"  -99", 0},
		{"-99", -99},
		{"2000123", 2000123},
	}

	for _, e := range expectations {
		r := ParseIntegerOr0(e.Input)
		if e.Output != r {
			t.Errorf("%d != ParseIntegerOr0(%q) (%d)\n", e.Output, e.Input, r)
		}
	}
}
