// Copyright 2020 by Chris Palmer (https://noncombatant.org)
// SPDX-License-Identifier: GPL-3.0

package main

import (
	"fmt"
	"unicode"
)

type Query struct {
	Keyword string
	Term    string
	Negated bool
}

func (q Query) String() string {
	return fmt.Sprintf("{Keyword: %q, Term: %q, Negated: %t}", q.Keyword, q.Term, q.Negated)
}

func ParseTerms(query string) []string {
	const (
		Start = iota
		Bareword
		Quoted
		Boundary
	)

	state := Start
	currentTerm := ""
	var terms []string

	for _, r := range query {
		if state == Bareword {
			if unicode.IsSpace(r) {
				state = Boundary
				if currentTerm != "" {
					terms = append(terms, currentTerm)
					currentTerm = ""
				}
			} else if r == ':' {
				state = Boundary
				if currentTerm != "" {
					terms = append(terms, currentTerm)
					currentTerm = ""
				}
				terms = append(terms, string(r))
			} else {
				currentTerm += string(r)
			}
		} else if state == Quoted {
			if r == '"' {
				state = Boundary
				if currentTerm != "" {
					terms = append(terms, currentTerm)
					currentTerm = ""
				}
			} else {
				currentTerm += string(r)
			}
		} else if state == Boundary {
			if r == '"' {
				state = Quoted
			} else if r == '-' || r == ':' {
				if currentTerm != "" {
					terms = append(terms, currentTerm)
					currentTerm = ""
				}
				terms = append(terms, string(r))
			} else if !unicode.IsSpace(r) {
				state = Bareword
				currentTerm += string(r)
			}
		} else {
			if unicode.IsSpace(r) {
				state = Boundary
				if currentTerm != "" {
					terms = append(terms, currentTerm)
					currentTerm = ""
				}
			} else if r == '"' {
				state = Quoted
				if currentTerm != "" {
					terms = append(terms, currentTerm)
					currentTerm = ""
				}
			} else if r == '-' || r == ':' {
				state = Boundary
				if currentTerm != "" {
					terms = append(terms, currentTerm)
					currentTerm = ""
				}
				terms = append(terms, string(r))
			} else {
				state = Bareword
				currentTerm += string(r)
			}
		}
	}

	if currentTerm != "" {
		terms = append(terms, currentTerm)
	}

	return terms
}

// Consumes some of `terms`; return a `Query` and remainder of `terms`.
func getQuery(terms []string) (Query, []string) {
	if terms[0] == "-" {
		if len(terms) > 1 {
			return Query{"", terms[1], true}, terms[2:]
		} else {
			// Don't return `Query{"", "-", ...}`, which is meaningless.
			return Query{}, []string{}
		}
	}

	if len(terms) > 1 {
		if terms[1] == ":" {
			if len(terms) > 3 {
				if terms[2] == "-" {
					return Query{terms[0], terms[3], true}, terms[4:]
				} else {
					return Query{terms[0], terms[2], false}, terms[3:]
				}
			} else {
				return Query{terms[0], terms[2], false}, terms[3:]
			}
		} else {
			return Query{"", terms[0], false}, terms[1:]
		}
	} else {
		return Query{"", terms[0], false}, []string{}
	}
}

func ReconstructQueries(terms []string) []Query {
	queries := make([]Query, 0)

	for {
		if len(terms) == 0 {
			break
		}

		var q Query
		q, terms = getQuery(terms)
		queries = append(queries, q)
	}

	return queries
}
