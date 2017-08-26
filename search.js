// Copyright 2016 by Chris Palmer (https://noncombatant.org), and released under
// the terms of the GNU GPL3. See help.html for more information.

// P R E D I C A T E S

//let editDistanceCache = {}
//const editDistanceThreshold = 1
//const allDigits = new RegExp(/^\d+$/)
//const spaces = new RegExp(/\s+/)
let reCache = {}

const match = function(property, term) {
  term = term.toLocaleLowerCase()

  if (!(term in reCache)) {
    reCache[term] = new RegExp(term, "i")
  }
  if (reCache[term].test(property)) {
    return true
  }

  //if (allDigits.test(property)) {
  //  return false
  //}

  //const words = property.toLocaleLowerCase().split(spaces)
  //for (let i = 0; i < words.length; ++i) {
  //  const word = words[i]

  //  const shorter = Math.min(word.length, term.length)
  //  const longer = Math.max(word.length, term.length)
  //  if (longer - shorter > editDistanceThreshold) {
  //    continue
  //  }

  //  const key = word + "\x00" + term
  //  if (!(key in editDistanceCache)) {
  //    editDistanceCache[key] = levenshteinEditDistance(word, term)
  //  }
  //  if (editDistanceCache[key] <= editDistanceThreshold) {
  //    return true
  //  }
  //}

  return false
}

const matchAny = function(object, term) {
  // NOTE: This is the global function `any` defined in util.js, not in
  // `searchFilters` below.
  return any(object, function(property) { return match(property, term) })
}

const twoMonthsAgo = get2MonthsAgo()

const searchFilters = {
  and: function() {
    return all(arguments, function(term) { return term })
  },

  or: function() {
    return any(arguments, function(term) { return term })
  },

  not: function() {
    return all(arguments, function(term) { return !term })
  },

  any: function() {
    let x = this
    return all(arguments, function(term) { return matchAny(x, term) })
  },

  audio: function() {
    if (arguments.length > 0) {
      let x = this
      return isAudioPathname(this[Pathname]) &&
             all(arguments, function(term) { return matchAny(x, term) })
    }
    return isAudioPathname(this[Pathname])
  },

  video: function() {
    if (arguments.length > 0) {
      let x = this
      return isVideoPathname(this[Pathname]) &&
             all(arguments, function(term) { return matchAny(x, term) })
    }
    return isVideoPathname(this[Pathname])
  },

  path: function() {
    let p = this[Pathname]
    return all(arguments, function(term) { return match(p, term) })
  },

  album: function() {
    let a = this[Album]
    return all(arguments, function(term) { return match(a, term) })
  },

  artist: function() {
    let a = this[Artist]
    return all(arguments, function(term) { return match(a, term) })
  },

  name: function() {
    let n = this[Name]
    return all(arguments, function(term) { return match(n, term) })
  },

  disc: function() {
    let d = parseIntOr(this[Disc], 1)
    return all(arguments, function(term) { return d === parseIntOr(term, 1) })
  },

  track: function() {
    let t = parseIntOr(this[Track], 1)
    return all(arguments, function(term) { return t === parseIntOr(term, 1) })
  },

  year: function() {
    let y = parseIntOr(this[Year], 1970)
    return all(arguments, function(term) { return y === parseIntOr(term, 1970) })
  },

  genre: function() {
    let g = this[Genre]
    return all(arguments, function(term) { return match(g, term) })
  },

  recent: function() {
    return this[Mtime] > twoMonthsAgo
  },

  all: function() {
    return true
  },
}

// P A R S E R
//
// Based on the Little Lisp interpreter by Mary Rose Cook:
// https://www.recurse.com/blog/21-little-lisp-interpreter
// https://github.com/maryrosecook/littlelisp/

const tokenize = function(string) {
  return string
    .trim()
    .replace(/\s+/, " ")
    .toLowerCase()
    .replace(/\(/g, " ( ")
    .replace(/\)/g, " ) ")
    .trim()
    .split(/\s+/)
}

const categorize = function(token, is_first) {
  return { type: (is_first ? "symbol" : "string"), value: token }
}

const parenthesize = function(tokens, list) {
  let token = tokens.shift()
  if (undefined === token) {
    return list.pop()
  }
  if ("(" == token) {
    list.push(parenthesize(tokens, []))
    return parenthesize(tokens, list)
  }
  if (")" == token) {
    return list
  }
  return parenthesize(tokens, list.concat(categorize(
      token, list.length == 0 ? true : false)))
}

const parse = function(query) {
  let tokens = tokenize(query)

  if (0 == tokens.length) {
    return parse("(all)")
  }

  if ("(" != tokens[0]) {
    return parse("(any " + query + ")")
  }

  return parenthesize(tokenize(query), [])
}

const Context = function(scope, parent) {
  this.scope = scope
  this.parent = parent
}

Context.prototype.get = function(symbol) {
  return this.scope[symbol] || (this.parent && this.parent.get(symbol))
}

const interpret = function(input, context) {
  if (input instanceof Array) {
    return interpretList(input, context)
  }
  if ("symbol" === input.type) {
    return context.get(input.value) || input.value
  }
  return input.value
}

const interpretList = function(input, context) {
  let list = input.map(function(x) { return interpret(x, context) })
  if (list[0] instanceof Function) {
    return list[0].apply(context.get("item"), list.slice(1))
  }
  return list
}
