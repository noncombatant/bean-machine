package main

import (
	"strings"
	"text/scanner"
)

type Query struct {
	Keyword string
	Term    string
	Negated bool
}

func ParseTerms(query string) []string {
	var s scanner.Scanner
	s.Init(strings.NewReader(query))
	terms := make([]string, 0)
	for t := s.Scan(); t != scanner.EOF; t = s.Scan() {
		terms = append(terms, strings.Trim(s.TokenText(), "\""))
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
