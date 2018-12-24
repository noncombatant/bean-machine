// Copyright 2018 by Chris Palmer (https://noncombatant.org), and released under
// the terms of the GNU GPL3. See help.html for more information.

"use strict";

let tsvs = undefined

const zeroOrMoreSpaces = /^\s*$/
const pushTerm = function(terms, term) {
  if (term.match(zeroOrMoreSpaces)) {
    return
  }
  terms.push(normalizeStringForSearch(term))
}

const parseTerms = function(string) {
  const terms = []
  let inQuotes = false
  let inWord = false
  let wordStart = 0

  for (let i = 0; i < string.length; ++i) {
    const c = string[i]
    if ('"' === c) {
      if (inQuotes) {
        inQuotes = inWord = false
        pushTerm(terms, string.substring(wordStart, i))
        wordStart = i + 1
      } else {
        if (-1 !== wordStart) {
          pushTerm(terms, string.substring(wordStart, i))
        }
        inQuotes = inWord = true
        wordStart = i + 1
      }
    } else if (c.match(/^\s/)) {
      if (inQuotes) {
        // do nothing
      } else if (inWord) {
        pushTerm(terms, string.substring(wordStart, i))
        inWord = inQuotes = false
        wordStart = i + 1
      } else {
        // do nothing
      }
    } else {
      if (inWord || inQuotes) {
        // do nothing
      } else {
        wordStart = i
        inWord = true
      }
    }
  }
  if (-1 !== wordStart) {
    const t = string.substring(wordStart, string.length).trim()
    if (t.length > 0) {
      pushTerm(terms, t)
    }
  }
  return terms
}

const getItemString = function(tsvs, itemID) {
  const end = tsvs.indexOf("\n", itemID)
  return tsvs.substring(itemID, end === -1 ? undefined : end)
}

const getMatchingItems = function(tsvs, tsvOffsets, query) {
  const start = performance.now()
  const hits = []
  const terms = parseTerms(query)
  for (let i = 0; i < tsvOffsets.length; ++i) {
    const item = getItemString(tsvs, tsvOffsets[i])
    if (itemMatches(terms, item)) {
      hits.push(tsvOffsets[i])
    }
  }
  console.log("getMatchingItems: " + Math.round(performance.now() - start))
  return hits
}

const itemMatches = function(terms, itemString) {
  itemString = normalizeStringForSearch(itemString)
  for (let i = 0; i < terms.length; ++i) {
    let t = terms[i]
    const negated = "-" === t[0]
    if (negated) {
      t = t.substring(1)
    }
    const matched = itemString.indexOf(t) >= 0
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
  if (!tsvs) {
    tsvs = e.data.tsvs
  }
  postMessage(getMatchingItems(tsvs, e.data.tsvOffsets, e.data.query))
})
