// Copyright 2020 by Chris Palmer (https://noncombatant.org), and released under
// the terms of the GNU GPL3. See web/index.html for more information.

package main

import (
	"testing"
)

type StringToStringExpectation struct {
	Input  string
	Output string
}

func TestExtractDigits(t *testing.T) {
	expectations := []StringToStringExpectation{
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

func TestGetBasenameExtension(t *testing.T) {
	expectations := []StringToStringExpectation{
		{"/usr/bin/goat.leg", ".leg"},
		{"/usr/bin/goat.lEg", ".leg"},
		{"/usr/bin/goat.BAT", ".bat"},
		{"c:\\win32\\goatpad.EXE", ".exe"},
		{"zip", ""},
		{"~/bin/zip.sh", ".sh"},
		{"zip.pdf.sh", ".sh"},
		{"/usr/local/goat.beard/thing.stuff.txt", ".txt"},
		{".hidden", ".hidden"},
		{"~/whatever/.hidden", ".hidden"},
	}

	for _, e := range expectations {
		r := GetBasenameExtension(e.Input)
		if e.Output != r {
			t.Errorf("%q != GetBasenameExtension(%q) (%q)\n", e.Output, e.Input, r)
		}
	}
}

func TestRemoveBasenameExtension(t *testing.T) {
	expectations := []StringToStringExpectation{
		{"/usr/bin/goat.leg", "/usr/bin/goat"},
		{"/usr/bin/goat.lEg", "/usr/bin/goat"},
		{"/usr/BIN/goat.BAT", "/usr/BIN/goat"},
		{"c:\\win32\\goatpad.EXE", "c:\\win32\\goatpad"},
		{"zip", "zip"},
		{"~/bin/zip.sh", "~/bin/zip"},
		{"zip.pdf.sh", "zip.pdf"},
		{"/usr/local/goat.beard/thing.stuff.txt", "/usr/local/goat.beard/thing.stuff"},
		{"/usr/local/goat.beard/thing", "/usr/local/goat.beard/thing"},
	}

	for _, e := range expectations {
		r := RemoveBasenameExtension(e.Input)
		if e.Output != r {
			t.Errorf("%q != RemoveBasenameExtension(%q) (%q)\n", e.Output, e.Input, r)
		}
	}
}

func TestEscapeDoubleQuotes(t *testing.T) {
	expectations := []StringToStringExpectation{
		{"\"Hello,\" they said. \"Blorp!\"",
			"\\\"Hello,\\\" they said. \\\"Blorp!\\\""},
	}

	for _, e := range expectations {
		r := EscapeDoubleQuotes(e.Input)
		if e.Output != r {
			t.Errorf("%q != EscapeDoubleQuotes(%q) (%q)\n", e.Output, e.Input, r)
		}
	}
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
	for i := range bytes {
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
		Input  string
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
