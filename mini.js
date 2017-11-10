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

const deserializeState = function(string) {
  return parseQueryString(string)
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

const itemMatchesQuery = interpret

const doSearchCatalog = function(query) {
  if ("" === query) {
    searchHits = resetSearchHits(catalog)
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
  searchHits = resetSearchHits(catalog)
}
main()
