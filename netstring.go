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

func FromNetstring(bytes []byte) ([]byte, int, error) {
	if bytes == nil || len(bytes) == 0 {
		return nil, -1, errors.New("Empty netstring")
	}

  colon := -1
  for i, c := range bytes {
    if c == ':' {
      colon = i
      break
    }
  }
  if colon == -1 {
    return nil, -1, errors.New("Malformed netstring: no colon")
  }

  length := -1
  n, e := fmt.Sscanf(string(bytes[:colon]), "%d", &length)
  if n != 1 || e != nil || length == -1 {
    return nil, -1, errors.New(fmt.Sprintf("Malformed netstring: bad length (%v, %v, %v)", n, e, length))
  }

  if bytes[colon+1+length] != ',' {
    return nil, -1, errors.New("Malformed netstring: no comma")
  }

	comma := colon+1+length
  return bytes[colon+1:comma], comma, nil
}

func ArrayFromNetstring(bytes []byte) ([][]byte, error) {
	decoded, final_comma, e := FromNetstring(bytes)
	if e != nil {
		return nil, e
	}
	if final_comma != len(bytes) - 1 {
    return nil, errors.New("Malformed netstring array: no final comma")
	}

	result := make([][]byte, 0)
	for {
		if len(decoded) == 0 {
			break
		}
		d, comma, e := FromNetstring(decoded)
		if e != nil {
			return nil, e
		}
		result = append(result, d)
		decoded = decoded[comma+1:]
	}
	return result, nil
}
