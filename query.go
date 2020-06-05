package main

import (
	"fmt"
	"strings"
	"text/scanner"
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

func NewParseTerms(query string) []string {
	inBareword := false
	inQuoted := false
	inBoundary := false
	currentTerm := ""
	var terms []string

	for _, r := range query {
		if inBareword {
			if unicode.IsSpace(r) {
				inBareword = false
				inBoundary = true
				if "" != currentTerm {
					terms = append(terms, currentTerm)
					currentTerm = ""
				}
			} else if ':' == r {
				inBareword = false
				inQuoted = false
				inBoundary = true
				if "" != currentTerm {
					terms = append(terms, currentTerm)
					currentTerm = ""
				}
				terms = append(terms, string(r))
			} else {
				currentTerm += string(r)
			}
		} else if inQuoted {
			if '"' == r {
				inQuoted = false
				inBoundary = false
				if "" != currentTerm {
					terms = append(terms, currentTerm)
					currentTerm = ""
				}
			} else {
				currentTerm += string(r)
			}
		} else if inBoundary {
			if '"' == r {
				inQuoted = true
				inBoundary = false
			} else if '-' == r || ':' == r {
				inBareword = false
				inQuoted = false
				inBoundary = true
				if "" != currentTerm {
					terms = append(terms, currentTerm)
					currentTerm = ""
				}
				terms = append(terms, string(r))
			} else {
				inBareword = true
				inBoundary = false
				if !unicode.IsSpace(r) {
					currentTerm += string(r)
				}
			}
		} else {
			if unicode.IsSpace(r) {
				inBareword = false
				inQuoted = false
				inBoundary = true
				if "" != currentTerm {
					terms = append(terms, currentTerm)
					currentTerm = ""
				}
			} else if '"' == r {
				inBareword = false
				inQuoted = true
				inBoundary = false
				if "" != currentTerm {
					terms = append(terms, currentTerm)
					currentTerm = ""
				}
			} else if '-' == r || ':' == r {
				inBareword = false
				inQuoted = false
				inBoundary = true
				if "" != currentTerm {
					terms = append(terms, currentTerm)
					currentTerm = ""
				}
				terms = append(terms, string(r))
			} else {
				inBareword = true
				inQuoted = false
				inBoundary = false
				currentTerm += string(r)
			}
		}
	}

	if "" != currentTerm {
		terms = append(terms, currentTerm)
	}

	return terms
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
