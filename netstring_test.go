// Copyright 2019 by Chris Palmer (https://noncombatant.org), and released under
// the terms of the GNU GPL3. See help.html for more information.

package main

import (
  "testing"
)

func TestEncodeAndDecode(t *testing.T) {
  message := "Hello, noodle-monsters!"
  netstring := ToNetstring([]byte(message))
  m, e := FromNetstring(netstring)
  if e != nil {
    t.Error(e)
  } else {
    s := string(m)
    if s != message {
      t.Error(s, message)
    }
  }
}
