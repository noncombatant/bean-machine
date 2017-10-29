"use strict";

// query := terms
// terms := term
//     | terms
// term := and-term
//     | and-not-term
// and-term: property ":" value
//     | value
// and-not-term: "-" and-term
// property := string
// value := string
// string := bareword
//     | quoted-string
// bareword := /^[-\w]+$/
// quoted-string := /^"[^"]+"$/

const spaceMatcher = new RegExp(/^\s+$/)
const isSpace = function(character) {
  return spaceMatcher.test(character)
}

const pushIf = function(tokens, token) {
  token = token.trim()
  if (0 === token.length) {
    return
  }
  tokens.push(token)
}

const getTokens = function(string) {
  const tokens = [] 
  let inQuotes = false
  let currentToken = ""
  for (let i = 0; i < string.length; ++i) {
    const c = string[i]
    const next = i < (string.length - 1) ? string[i + 1] : ""

    // Allow escaped quotes.
    if ('\\' === c && '"' === next) {
      currentToken += '"'
      ++i
      continue
    }

    if (inQuotes) {
      if ('"' === c) {
        if (":" === next) {
          currentToken += ":"
          ++i
        }
        pushIf(tokens, currentToken)
        currentToken = ""
        inQuotes = false
      } else {
        currentToken += c
      }
    } else {
      if ('"' === c) {
        inQuotes = true
      } else if (isSpace(c)) {
        pushIf(tokens, currentToken)
        currentToken = ""
      } else if (":" === c) {
        pushIf(tokens, currentToken + ":")
        currentToken = ""
      } else {
        currentToken += c
      }
    }
  }
  pushIf(tokens, currentToken)
  return tokens
}

const getTerms = function(tokens) {
  const terms = []
  for (let i = 0; i < tokens.length; ++i) {
    let token = tokens[i]
    if (0 === token.length) {
      throw "Invalid token. Bug in `getTokens`."
    }
    const next = i < (tokens.length - 1) ? tokens[i + 1] : ""
    const term = { property: "", value: "", negated: false }

    if ("-" === token[0]) {
      if (1 === token.length) {
        continue
      }
      token = token.substring(1)
      term.negated = true
    }

    if (token.endsWith(":")) {
      if (1 === token.length) {
        continue
      }
      token = token.substring(0, token.length - 1)
      term.property = normalizeStringForSearch(token)
      token = next
      ++i
    }

    // Allow a 0-length value if there is a property, to enable tests for
    // property existence. Otherwise, require a non-empty token.
    if (0 === term.property.length && 0 === token.length) {
      throw "Invalid token. Bug in `getTokens`."
    }
    term.value = token
    terms.push(term)
  }
  return terms
}

// catalog.js must have been included first. TODO: Might be good to define this
// in catalog.js instead. TODO: Remove this testing stuff.
const Pathname = 0
const Album = 1
const Artist = 2
const Name = 3
const Disc = 4
const Track = 5
const Year = 6
const Genre = 7
const Mtime = 8

const termPropertiesMap = {
  "path": Pathname,
  "album": Album,
  "artist": Artist,
  "name": Name,
  "disc": Disc,
  "track": Track,
  "year": Year,
  "before": Year,
  "after": Year,
  "genre": Genre,
  "mtime": Mtime,
  "mbefore": Mtime,
  "mafter": Mtime,
}

// TODO: Implement before and after as well. Requires a 3rd argument.
//
// TODO: If itemPropertyValue is numeric, match numerically. Otherwise, debate
// indexOf or startsWith? indexOf for Pathname, startsWith for other string
// properties.
const matchPropertyValue = function(itemPropertyValue, termValue) {
  if ("number" === typeof(itemPropertyValue)) {
    return itemPropertyValue == Number(termValue)
  } else {
    const normalized = normalizeStringForSearch(itemPropertyValue)
    return normalized.startsWith(termValue)
  }
}

// Returns true if the catalog `item` matches the `terms`, false if not.
const matchItem = function(terms, item) {
  for (let term of terms) {
    const itemPropertyIndex = termPropertiesMap[term.property]
    term.value

    // Test the presence of the property.
    if (term.property) {
      if (!itemPropertyIndex) {
        if (term.negated) {
          continue
        }
        return false
      }

      // Test if the property matches.
      const itemPropertyValue = item[itemPropertyIndex]
      if (matchPropertyValue(itemPropertyValue, term.value) === term.negated) {
        return false
      }
    } else {
      // Otherwise, test all properties.
      if (term.negated) {
        // TODO: Should probably abstract these matchPropertyValue if-blocks
        // into a matchAnyPropertyValue function.
        if (matchPropertyValue(item[Pathname], term.value) ||
            matchPropertyValue(item[Album], term.value) ||
            matchPropertyValue(item[Artist], term.value) ||
            matchPropertyValue(item[Name], term.value) ||
            matchPropertyValue(item[Disc], term.value) ||
            matchPropertyValue(item[Track], term.value) ||
            matchPropertyValue(item[Year], term.value) ||
            matchPropertyValue(item[Genre], term.value) ||
            matchPropertyValue(item[Mtime], term.value)) {
          return false
        }
      } else {
        if (!matchPropertyValue(item[Pathname], term.value) &&
            !matchPropertyValue(item[Album], term.value) &&
            !matchPropertyValue(item[Artist], term.value) &&
            !matchPropertyValue(item[Name], term.value) &&
            !matchPropertyValue(item[Disc], term.value) &&
            !matchPropertyValue(item[Track], term.value) &&
            !matchPropertyValue(item[Year], term.value) &&
            !matchPropertyValue(item[Genre], term.value) &&
            !matchPropertyValue(item[Mtime], term.value)) {
          return false
        }
      }
    }
  }

  return true
}

const matchItems = function(terms, items) {
  const matches = []
  for (let i = 0; i < items.length; ++i) {
    if (matchItem(terms, items[i])) {
      matches.push(i)
    }
  }
  return matches
}
