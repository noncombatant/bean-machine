// Copyright 2022 by Chris Palmer (https://noncombatant.org), and released under
// the terms of the GNU GPL3. See web/index.html for more information.

package main

import (
	"testing"
)

func TestGetDiscAndTrackFromBasename(t *testing.T) {
	testCases := [][]string{
		{"1-01 Hells Bells.m4a", "1", "01", "Hells Bells.m4a"},
		{"1-1 Hells Bells.m4a", "1", "1", "Hells Bells.m4a"},
		{"1 Hells Bells.m4a", "", "1", "Hells Bells.m4a"},
		{"01 Hells Bells.m4a", "", "01", "Hells Bells.m4a"},
		{"-1 Hells Bells.m4a", "", "1", "Hells Bells.m4a"},
		{"-01 Hells Bells.m4a", "", "01", "Hells Bells.m4a"},
		{"Hells Bells.m4a", "", "", "Hells Bells.m4a"},
	}
	for i, c := range testCases {
		disc, track, name := getDiscTrackAndNameFromBasename(c[0])
		if c[1] != disc || c[2] != track || c[3] != name {
			t.Errorf("%d: %q → %q, %q, %q", i, c[0], disc, track, name)
		}
	}
}
