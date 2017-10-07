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

    // Allow escaped quotes.
    if ('\\' === c && i < (string.length - 1) && '"' === string[i + 1]) {
      currentToken += '"'
      ++i
      continue
    }

    if (inQuotes) {
      if ('"' === c) {
        if (i < (string.length - 1) && ":" === string[i + 1]) {
          currentToken += ":"
          ++i
        }
        pushIf(tokens, currentToken)
        currentToken = ""
        inQuotes = true
      } else {
        currentToken += c
      }
    } else {
      if ('"' === c) {
        pushIf(tokens, currentToken)
        currentToken = ""
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
