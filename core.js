// Copyright 2017 by Chris Palmer (https://noncombatant.org), and released under
// the terms of the GNU GPL3. See help.html for more information.

"use strict";

let searchHits

const resetSearchHits = function(catalog) {
  const hits = new Array(catalog.length)
  for (let i = 0; i < catalog.length; ++i) {
    hits[i] = i
  }
  return hits
}

const displayNowPlaying = function(item, element) {
  removeAllChildren(element)
  const trackName = item[Name] || basename(item[Pathname])
  element.appendChild(createElement("span", "", item[Disc] + "-" + item[Track] + " “" + trackName + "”\u200A—\u200A"))
  element.appendChild(createElement("strong", "", item[Artist]))
  element.appendChild(createElement("span", "", "\u200A—\u200A"))
  element.appendChild(createElement("em", "", item[Album]))
  document.title = element.textContent
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

const playNext = function(e) {
  const random = (typeof(randomCheckbox) !== "undefined" && randomCheckbox.checked) ||
                 (typeof(shuffleButton) !== "undefined" && "Repeat" === shuffleButton.title)
  if (random) {
    let i
    while (true) {
      i = getRandomIndex(searchHits)
      if (i !== undefined) {
        break
      }
    }
    doPlay(searchHits[i], true)
  } else {
    for (let i = 0; i < searchHits.length; ++i) {
      if (player.itemID === searchHits[i]) {
        doPlay(searchHits[(i + 1) % searchHits.length], true)
        return
      }
    }
    doPlay(searchHits[0], true)
  }
}

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

const randomCheckboxOnClick = function(e) {
  localStorage.setItem("random", randomCheckbox.checked)
}

const restoreState = function() {
  const itemID = localStorage.getItem("itemID")
  if (undefined !== typeof(itemID) && null !== itemID) {
    doPlay(itemID, false)
  }
  const query = localStorage.getItem("query")
  if (query) {
    searchInput.value = query
    searchCatalog(query, true)
  }
  searchInput.focus()
  searchInput.select()
  if ('undefined' !== typeof(randomCheckbox)) {
    const random = localStorage.getItem("random")
    randomCheckbox.checked = "true" === random ? true : false
  }
}

const main = function() {
  addEventListeners()
  searchHits = resetSearchHits(catalog)
  player.volume = 0.5
  restoreState()
}
