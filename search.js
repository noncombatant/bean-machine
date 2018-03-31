"use strict";

const splitIntoWordSet = memoize(function(string) {
  return new Set(string.split(/\W+/))
})

const itemMatches = function(terms, item) {
  const delimiter = "\x00"
  const all = normalizeStringForSearch(item[Pathname] + delimiter + item[Artist] + delimiter + item[Album] + delimiter + item[Name] + delimiter + item[Genre])
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
  const previousQuery = localStorage.getItem("query")
  if (!forceSearch && previousQuery === query) {
    return
  }
  localStorage.setItem("query", query)
  searchHits = getMatchingItems(catalog, query)
  if (typeof(buildCatalog) !== "undefined") {
    // TODO: BUG: The presence of buildCatalog is a symptom of this code running
    // in the context of index.html, not mini.html. But it's not a direct
    // indicator; it would be better to parameterize this function.
    previousLastItem = buildCatalog(0)
  }
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
