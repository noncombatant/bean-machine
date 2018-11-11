"use strict";

const zeroOrMoreSpaces = /^\s*$/
const pushTerm = function(terms, term) {
  if (term.match(zeroOrMoreSpaces)) {
    return
  }
  terms.push(normalizeStringForSearch(term))
}

const parseTerms = function(string) {
  const terms = []
  let in_quotes = false
  let in_word = false
  let word_start = 0

  for (let i = 0; i < string.length; ++i) {
    const c = string[i]
    if ('"' === c) {
      if (in_quotes) {
        in_quotes = in_word = false
        pushTerm(terms, string.substring(word_start, i))
        word_start = i + 1
      } else {
        if (-1 !== word_start) {
          pushTerm(terms, string.substring(word_start, i))
        }
        in_quotes = in_word = true
        word_start = i + 1
      }
    } else if (c.match(/^\s/)) {
      if (in_quotes) {
        // do nothing
      } else if (in_word) {
        pushTerm(terms, string.substring(word_start, i))
        in_word = in_quotes = false
        word_start = i + 1
      } else {
        // do nothing
      }
    } else {
      if (in_word || in_quotes) {
        // do nothing
      } else {
        word_start = i
        in_word = true
      }
    }
  }
  if (-1 !== word_start) {
    const t = string.substring(word_start, string.length).trim()
    if (t.length > 0) {
      pushTerm(terms, t)
    }
  }
  return terms
}

const itemMatches = function(terms, item) {
  const delimiter = "\x00"
  const all = normalizeStringForSearch(item[Pathname] + delimiter + item[Artist] + delimiter + item[Album] + delimiter + item[Name] + delimiter + item[Genre] + delimiter + item[Year] + delimiter + item[Mtime])
  for (let i = 0; i < terms.length; ++i) {
    let t = terms[i]
    const negated = "-" === t[0]
    if (negated) {
      t = t.substring(1)
    }
    const matched = all.indexOf(t) >= 0
    if (negated === matched) {
      return false
    }
  }
  return true
}

const getMatchingItems = function(catalog, query) {
  const hits = []
  const terms = parseTerms(query)
  for (let i = 0; i < catalog.length; ++i) {
    const item = catalog[i]
    if (itemMatches(terms, item)) {
      hits.push(i)
    }
  }
  return hits
}

const searchCatalog = function(query, forceSearch) {
  query = query.trim()
  const previousQuery = localStorage.getItem("query")
  if (!forceSearch && previousQuery === query) {
    return
  }
  localStorage.setItem("query", query)
  searchHits = getMatchingItems(catalog, query)
  previousLastItem = buildCatalog(0)
}

const executeSearch = function(e) {
  searchCatalog(searchInput.value, false)
}

const searchInputOnKeyUp = function(e) {
  e.stopPropagation()
  if ("Enter" === e.code) {
    searchCatalog(this.value, false)
  }
}
