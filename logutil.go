// Copyright 2020 by Chris Palmer (https://noncombatant.org), and released under
// the terms of the GNU GPL3. See help.html for more information.
//
// Adapted from
// https://wycd.net/posts/2014-07-02-logging-function-names-in-go.html.Thanks
// wyc!

package main

import (
	"log"
	"runtime"
)

var (
	Logger = log.New(LogWriter{}, "", 0)
)

type LogWriter struct{}

func (f LogWriter) Write(bytes []byte) (int, error) {
	pc, _, _, _ := runtime.Caller(3)
	function := runtime.FuncForPC(pc)
	name := "unknown function"
	if function != nil {
		name = function.Name()
	}
	log.Printf("%s: %s", name, bytes)
	return len(bytes), nil
}
