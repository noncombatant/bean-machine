// Copyright 2019 by Chris Palmer (https://noncombatant.org), and released under
// the terms of the GNU GPL3. See help.html for more information.

// https://cr.yp.to/proto/netstrings.txt

package main

import (
  "errors"
  "fmt"
)

func ToNetstring(bytes []byte) []byte {
  length := []byte(fmt.Sprintf("%d:", len(bytes)))
  netstring := make([]byte, 0, len(length) + len(bytes) + 1)
  netstring = append(netstring, length...)
  netstring = append(netstring, bytes...)
  netstring = append(netstring, ',')
  return netstring
}

func FromNetstring(bytes []byte) ([]byte, error) {
  colon := -1
  for i, c := range bytes {
    if c == ':' {
      colon = i
      break
    }
  }
  if colon == -1 {
    return nil, errors.New("Malformed netstring: no colon")
  }

  length := -1
  n, e := fmt.Sscanf(string(bytes[:colon]), "%d", &length)
  if n != 1 || e != nil || length == -1 {
    return nil, errors.New("Malformed netstring: bad length")
  }

  if bytes[colon+1+length] != ',' {
    return nil, errors.New("Malformed netstring: no comma")
  }

  return bytes[colon+1:colon+1+length], nil
}
