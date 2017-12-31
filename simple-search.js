"use strict";

const splitIntoWordSet = memoize(function(string) {
  return new Set(string.split(/\W+/))
})

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

const searchCatalog = function(query, forceSearch) {
  query = query.trim()
  const previousQuery = deserializeState(document.location.hash).query
  if (!forceSearch && previousQuery === query) {
    return
  }
  searchHits = getMatchingItems(catalog, query)
  setLocationHash()
  if (typeof(buildCatalog) !== "undefined") {
    // TODO: BUG: The presence of buildCatalog is a symptom of this code running
    // in the context of index.html, not mini.html. But it's not a direct
    // indicator; it would be better to parameterize this function.
    previousLastItem = buildCatalog(0)
    randomHistory = {}
  }
}
