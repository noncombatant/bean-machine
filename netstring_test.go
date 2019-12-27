// Copyright 2019 by Chris Palmer (https://noncombatant.org), and released under
// the terms of the GNU GPL3. See help.html for more information.

package main

import (
	"bytes"
  "testing"
)

func TestEncodeAndDecode(t *testing.T) {
	messages := [][]byte{
		[]byte("Hello, noodle-monsters!"),
		[]byte("—wow—"),
		[]byte("Boing\t\nBonk woop"),
		[]byte("*@#%@$^\x00**;[]yay"),
	}

	for _, message := range messages {
		netstring := ToNetstring(message)
		m, comma, e := FromNetstring(netstring)
		if comma != len(netstring)-1 {
			t.Fatal("Netstring did not end in comma")
		}
		if e != nil {
			t.Fatal(e)
		} else {
			if bytes.Compare(m, message) != 0 {
				t.Fatal(m, message)
			}
		}
	}

	all := make([]byte, 0)
	for _, message := range messages {
		all = append(all, ToNetstring(message)...)
	}
	netstring := ToNetstring(all)
	decoded, e := ArrayFromNetstring(netstring)
	if e != nil {
		t.Fatal(e)
	}
	if len(messages) != len(decoded) {
		t.Fatal(len(messages), len(decoded))
	}
	for i, _ := range decoded {
		if bytes.Compare(messages[i], decoded[i]) != 0 {
			t.Fatal(i, messages[i], decoded[i])
		}
	}
}
