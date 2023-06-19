// Copyright 2020 by Chris Palmer (https://noncombatant.org)
// SPDX-License-Identifier: GPL-3.0

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

func TestContains(t *testing.T) {
	for _, s := range audioFormatExtensions {
		if !Contains[string](s, audioFormatExtensions) {
			t.Errorf("Could not find %q", s)
		}
	}

	if Contains[string](".FLAC", audioFormatExtensions) {
		t.Errorf("Found \".FLAC\" even though it's upper-case")
	}

	if Contains[string]("goat", audioFormatExtensions) {
		t.Errorf("Found \"goat\" even though it's not in `audioFormatExtensions`")
	}
}
