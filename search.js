// Copyright 2018 by Chris Palmer (https://noncombatant.org), and released under
// the terms of the GNU GPL3. See help.html for more information.

"use strict";

// TODO: It's goaty to have to repeat these.
const Pathname = 0
const Album = 1
const Artist = 2
const Name = 3
const Disc = 4
const Track = 5
const Year = 6
const Genre = 7
const Mtime = 8

const zeroOrMoreSpaces = /^\s*$/
const pushTerm = function(terms, term) {
  if (term.match(zeroOrMoreSpaces)) {
    return
  }
  terms.push(normalizeStringForSearch(term))
}

const parseTerms = function(string) {
  const terms = []
  // TODO: Use jsStyle:
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

// Borrowed from
// https://github.com/mathiasbynens/strip-combining-marks/blob/master/strip-combining-marks.js
// by Mathias Bynens <https://mathiasbynens.be/>.
//
// "héllo".normalize("NFD").replace(regexSymbolWithCombiningMarks, '$1') -> "hello"

const regexSymbolWithCombiningMarks = new RegExp(/([\0-\u02FF\u0370-\u1AAF\u1B00-\u1DBF\u1E00-\u20CF\u2100-\uD7FF\uE000-\uFE1F\uFE30-\uFFFF]|[\uD800-\uDBFF][\uDC00-\uDFFF]|[\uD800-\uDBFF](?![\uDC00-\uDFFF])|(?:[^\uD800-\uDBFF]|^)[\uDC00-\uDFFF])([\u0300-\u036F\u1AB0-\u1AFF\u1DC0-\u1DFF\u20D0-\u20FF\uFE20-\uFE2F]+)/g)

const memoize = function(f) {
  const memo = {}
  return function() {
    const a = Array.prototype.slice.call(arguments)
    if (!memo.hasOwnProperty(a)) {
      memo[a] = f.apply(null, a)
    }
    return memo[a]
  }
}

const normalizeStringForSearch = memoize(function(string) {
  return string.toString().normalize("NFD").replace(regexSymbolWithCombiningMarks, '$1').toLocaleLowerCase()
})

addEventListener("message", function(e) {
  postMessage(getMatchingItems(e.data.catalog, e.data.query))
})
