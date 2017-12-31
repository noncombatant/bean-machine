"use strict";

const splitIntoWordSetMemo = new Set()
const splitIntoWordSet = function(string) {
  if (splitIntoWordSetMemo.has(string)) {
    return splitIntoWordSetMemo[string]
  }
  const words = new Set(string.split(/\W+/))
  splitIntoWordSetMemo[string] = words
  return words
}

const itemMatches = function(terms, item) {
  const all = normalizeStringForSearch(item[Pathname] + item[Artist] + item[Album] + item[Name] + item[Genre])
  for (let t = 0 ; t < terms.length; ++t) {
    if (-1 === all.indexOf(terms[t])) {
      return false
    }
  }
  return true
}

const getMatchingItems = function(catalog, query) {
  const hits = []
  for (let i = 0; i < catalog.length; ++i) {
    const item = catalog[i]
    const terms = [...splitIntoWordSet(normalizeStringForSearch(query))]
    if (itemMatches(terms, item)) {
      hits.push(i)
    }
  }
  return hits
}
