// Copyright 2016 by Chris Palmer (https://noncombatant.org), and released under
// the terms of the GNU GPL3. See help.html for more information.

"use strict";

// G L O B A L   V A R I A B L E S
//
// TODO: Move globals to inside `main`. This requires explicitly passing and
// returning them to and from functions.

let searchHits

const sortingProperties = [ Album, Disc, Track, Pathname, Name ]

// C O R E   F U N C T I O N A L I T Y

const displayNowPlaying = function(item, element) {
  removeAllChildren(element)

  const trackName = item[Name] || basename(item[Pathname])
  element.appendChild(createElement("span", "", item[Disc] + "-" + item[Track] + " “" + trackName + "”\u200A—\u200A"))
  element.appendChild(createElement("strong", "", item[Artist]))
  element.appendChild(createElement("span", "", "\u200A—\u200A"))
  element.appendChild(createElement("em", "", item[Album]))

  document.title = element.textContent

  // TODO: Re-enable this when fully supported. As of December 2016, Firefox
  // ignores `silent`, and Chrome for Android throws an "illegal constructor"
  // exception.
  //const icon = dirname(player.src) + "/cover.jpg"
  //showNotification(document.title, { silent: true, icon: icon, badge: icon })
}

const leadingJunk = new RegExp("^(the\\s+|a\\s+|an\\s+|les?\\s+|las?\\s+|\"|'|\\.+\\s*)", "i")
const normalizeTitle = function(title) {
  const match = title.match(leadingJunk)
  return match ? title.substr(match[0].length) : title
}

const compareNormalizedStrings = function(a, b) {
  const aa = normalizeTitle(a)
  const bb = normalizeTitle(b)
  if (aa === bb) {
    return 0
  }
  if (aa < bb) {
    return -1
  }
  return 1
}

const itemComparator = function(a, b) {
  a = catalog[a]
  b = catalog[b]
  for (let p of sortingProperties) {
    const c =
      (Disc == p || Track == p || Year == p)
        ? parseIntOr(a[p], 1) - parseIntOr(b[p], 1)
        : compareNormalizedStrings(a[p], b[p])
    if (0 !== c) {
      return c
    }
  }
  return 0
}

const deserializeState = function(string) {
  return parseQueryString(string)
}

const setLocationHash = function() {
  const state = { "itemID": player.itemID, "query": searchInput.value }
  document.location.hash = constructQueryString(state)
}

const doPlay = function(itemID) {
  player.pause()
  // TODO: Potentially not necessary.
  setSingleTextChild(playButton, "Play")
  const item = catalog[itemID]
  player.src = item[Pathname]
  player.itemID = itemID
  player.play()
  setSingleTextChild(playButton, "Pause")

  displayNowPlaying(item, nowPlayingTitle)
  setLocationHash()
}

const playNext = function(e) {
  if ("Sequential" === randomToggle.innerText) {
    let i
    while (true) {
      i = getRandomIndex(searchHits)
      if (i !== undefined) {
        break
      }
    }
    doPlay(searchHits[i])
  } else {
    for (let i = 0; i < searchHits.length; ++i) {
      if (player.itemID === searchHits[i]) {
        doPlay(searchHits[(i + 1) % searchHits.length])
        return
      }
    }
    doPlay(searchHits[0])
  }
}

const buildItemDiv = function(itemID) {
  const item = catalog[itemID]
  const div = createElement("div", "itemDiv")
  div.itemID = itemID
  div.addEventListener("dblclick", itemDivOnClick)
  div.addEventListener("click", itemDivOnClick)

  const trackSpan = createElement("span", "itemDivCell trackNumber", (item[Disc] || "1") + "-" + (item[Track] || "1"))
  div.appendChild(trackSpan)

  const nameSpan = createElement("span", "itemDivCell", item[Name])
  div.appendChild(nameSpan)

  return div
}

const itemMatchesQuery = interpret

const doSearchCatalog = function(query) {
  if ("" === query) {
    searchHits = resetSearchHits()
  } else {
    const ast = parse(searchInput.value)
    const context = new Context(searchFilters)
    searchHits = []
    for (let i = 0; i < catalog.length; ++i) {
      context.scope.item = catalog[i]
      if (itemMatchesQuery(ast, context)) {
        searchHits.push(i)
      }
    }
  }

  setLocationHash()
  // TODO: Is this necessary (here or in bean-machine.js)?
  searchHits.sort(itemComparator)
}

const searchCatalog = function(query, forceSearch) {
  query = query.trim()
  const previousQuery = deserializeState(document.location.hash).query
  if (!forceSearch && previousQuery === query) {
    return
  }
  doSearchCatalog(query)
}

// E V E N T   H A N D L E R S

const togglePlayback = function(e) {
  e.stopPropagation()
  if ("p" !== e.key) {
    return
  }
  if (player.paused) {
    player.play()
  } else {
    player.pause()
  }
}

let errorCount = 0
const playerLoadedMetadata = function(e) {
  errorCount = 0
}

const playerOnError = function(e) {
  console.log("Could not load", catalog[player.itemID][Pathname], e)
  if (errorCount < 10) {
    this.dispatchEvent(new Event("ended"))
  }
  ++errorCount
}

const searchInputOnKeyUp = function(e) {
  e.stopPropagation()
  const enterKeyCode = 13
  enterKeyCode == e.keyCode && searchCatalog(this.value, false)
}

const executeSearch = function(e) {
  searchCatalog(searchInput.value, false)
}

const randomToggleOnClick = function(e) {
  if ("Sequential" === randomToggle.innerText) {
    setSingleTextChild(randomToggle, "Random")
  } else {
    setSingleTextChild(randomToggle, "Sequential")
  }
}

const playButtonOnClicked = function(e) {
  if ("Play" === playButton.innerText) {
    if (undefined === player.itemID) {
      playNext(e)
    } else {
      doPlay(player.itemID)
    }
    setSingleTextChild(playButton, "Pause")
  } else {
    setSingleTextChild(playButton, "Play")
    player.pause()
  }
}

// M A I N

const addEventListeners = function() {
  playButton.addEventListener("click", playButtonOnClicked)
  nextButton.addEventListener("click", playNext)
  player.addEventListener("ended", playNext)
  player.addEventListener("error", playerOnError)
  player.addEventListener("loadedmetadata", playerLoadedMetadata)
  randomToggle.addEventListener("click", randomToggleOnClick)
  searchInput.addEventListener("blur", executeSearch)
  searchInput.addEventListener("keyup", searchInputOnKeyUp)
  searchButton.addEventListener("click", executeSearch)
  document.body.addEventListener("keyup", togglePlayback)
}

const main = function() {
  getFormatExtensions()
  addEventListeners()
  searchHits = resetSearchHits()
}
main()
