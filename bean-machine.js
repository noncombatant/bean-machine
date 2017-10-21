// Copyright 2016 by Chris Palmer (https://noncombatant.org), and released under
// the terms of the GNU GPL3. See help.html for more information.

"use strict";

// G L O B A L   V A R I A B L E S
//
// TODO: Move globals to inside `main`. This requires explicitly passing and
// returning them to and from functions.

let player = audioPlayer
let searchHits
let playHistory = []
let randomHistory = {}

const buildCatalogLimit = 50
const sortingProperties = [ Album, Disc, Track, Pathname, Name ]

// C O R E   F U N C T I O N A L I T Y

const resetSearchHits = function() {
  const hits = new Array(catalog.length)
  for (let i = 0; i < catalog.length; ++i) {
    hits[i] = i
  }
  return hits
}

const setAudioVideoControls = function(itemID) {
  const pathname = catalog[itemID][Pathname]
  const volume = player.volume
  if (isAudioPathname(pathname)) {
    player = audioPlayer
    videoPlayerDiv.style.display = videoPlayerBackground.style.display = "none"
  } else if (isVideoPathname(pathname)) {
    player = videoPlayer
    videoPlayerDiv.style.display = videoPlayerBackground.style.display = "block"
  }
  player.className = "normal"
  player.volume = volume
}

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

const assertStateDefaults = function(state) {
  state.itemID = parseIntOr(idOrLast(state.itemID), 0)
  state.query = idOrLast(state.query) || ""
}

const deserializeState = function(string) {
  const state = parseQueryString(string)
  assertStateDefaults(state)
  return state
}

const setLocationHash = function() {
  const state = { "itemID": player.itemID, "query": searchInput.value }
  assertStateDefaults(state)
  document.location.hash = constructQueryString(state)
}

const updateShareLink = function() {
  shareLink.href = ""
  setSingleTextChild(shareLink, "")
  const item = catalog[player.itemID]
  if (!item) {
    return
  }
  const pathname = item[Pathname]

  const xhr = new XMLHttpRequest()
  xhr.open("GET", "/get-cap?n=" + pathname)
  xhr.addEventListener("load", function() {
    const l = document.location
    const link = l.protocol + "//" + l.host + "/" + pathname + "?cap=" + this.responseText
    shareLink.href = link
    setSingleTextChild(shareLink, "Shareable Link")
  })
  xhr.addEventListener("error", function() {
    console.log("Could not get cap for " + pathname, this.statusText)
  })
  xhr.send()
}

const doPlay = function(itemID) {
  player.pause()
  setAudioVideoControls(itemID)
  const item = catalog[itemID]
  player.src = item[Pathname]
  player.itemID = itemID
  player.play()
  playHistory.unshift(itemID)
  if (randomCheckbox.checked) {
    randomHistory[itemID] = true
  }

  displayNowPlaying(item, nowPlayingTitle)
  setLocationHash()
  // TODO: Re-enable this when we've deployed the Go server in 'production'.
  //updateShareLink()
}

const playNext = function(e) {
  if (randomCheckbox.checked) {
    let i
    while (true) {
      i = getRandomIndexWithoutRepeating(searchHits, randomHistory)
      if (i !== undefined) {
        break
      }
      randomHistory = {}
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

const buildAlbumTitleDiv = function(itemID) {
  const item = catalog[itemID]
  const div = createElement("div", "albumTitleDiv")
  div.itemID = itemID
  div.addEventListener("dblclick", albumTitleDivOnClick)
  div.addEventListener("click", albumTitleDivOnClick)

  const albumSpan = createElement("span", "itemDivCell albumTitle", item[Album])
  div.appendChild(albumSpan)

  const artist = item[Pathname].startsWith("Compilations/") ? "Various Artists" : item[Artist]
  const artistSpan = createElement("span", "itemDivCell artistName", artist)
  div.appendChild(artistSpan)

  if (item[Year]) {
    const yearSpan = createElement("span", "itemDivCell year", item[Year])
    div.appendChild(yearSpan)
  }

  return div
}

let previousLastItem = 0
let currentAlbumPathname = ""
const buildCatalog = function(start) {
  if (0 === start) {
    removeAllChildren(itemListDiv)
  } else {
    itemListDiv.removeChild($("bottom"))
  }

  const limit = Math.min(searchHits.length, buildCatalogLimit)
  let i
  for (i = 0; i < limit && start + i < searchHits.length; ++i) {
    const itemID = searchHits[start + i]
    const item = catalog[itemID]
    const albumPathname = dirname(item[Pathname])
    if (albumPathname !== currentAlbumPathname) {
      itemListDiv.appendChild(buildAlbumTitleDiv(itemID))
      currentAlbumPathname = albumPathname
    }
    itemListDiv.appendChild(buildItemDiv(itemID))
  }

  const bottom = createElement("div")
  bottom.id = "bottom"
  itemListDiv.appendChild(bottom)

  return start + i
}

let extendCatalogRequested = false
const extendCatalog = function() {
  if (isElementInViewport($("bottom"))) {
    previousLastItem = buildCatalog(previousLastItem)
  }
  extendCatalogRequested = false
}

const itemMatchesQuery = interpret

const doSearchCatalog = function(query) {
  const start = performance.now()
  setSingleTextChild(messageSpan, "Loading media. Please wait...")

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

  const end = performance.now()

  setLocationHash()
  setSingleTextChild(messageSpan, "Found " + searchHits.length.toLocaleString() + " items in " + Math.round(end - start) + " ms")
  searchHits.sort(itemComparator)
  previousLastItem = buildCatalog(0)
}

const searchCatalog = function(query, forceSearch) {
  query = query.trim()
  const previousQuery = deserializeState(document.location.hash).query
  if (!forceSearch && previousQuery === query) {
    return
  }
  doSearchCatalog(query)
  randomHistory = {}
}

const showPlayHistory = function() {
  searchHits = []
  for (let i = 0; i < playHistory.length; ++i) {
    searchHits.push(playHistory[i])
  }
  previousLastItem = buildCatalog(0)
}

// E V E N T   H A N D L E R S

const closeVideo = function(e) {
  videoPlayer.pause()
  videoPlayerDiv.style.display = videoPlayerBackground.style.display = "none"
}

const albumTitleDivOnClick = function(e) {
  const itemID = this.itemID
  randomCheckbox.checked = false
  setLocationHash()
  if (player.paused || player.itemID != itemID) {
    if (undefined !== itemID) {
      if (itemID == player.itemID) {
        player.play()
      } else {
        doPlay(itemID)
      }
    } else {
      playNext()
    }
  } else {
    player.pause()
  }
}
const itemDivOnClick = albumTitleDivOnClick

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
  setSingleTextChild(messageSpan, catalog[player.itemID][Pathname])
  if (errorCount < 10) {
    this.dispatchEvent(new Event("ended"))
  }
  ++errorCount
}

const searchInputOnKeyUp = function(e) {
  e.stopPropagation()
  showHistoryButton.className = ""
  const enterKeyCode = 13
  enterKeyCode == e.keyCode && searchCatalog(this.value, false)
}

const executeSearch = function(e) {
  searchCatalog(searchInput.value, false)
}

const showHistoryButtonOnClick = function(e) {
  if ("Show History" === showHistoryButton.innerText) {
    showHistoryButton.innerText = "Show Search Results"
    showPlayHistory()
  } else {
    showHistoryButton.innerText = "Show History"
    doSearchCatalog(searchInput.value)
  }
}

const randomCheckboxOnClick = function(e) {
  randomHistory = {}
}

const windowOnScroll = function(e) {
  if (!extendCatalogRequested) {
    window.requestAnimationFrame(extendCatalog)
  }
  extendCatalogRequested = true
}

// TODO: This can be removed when we remove the letter links.
const letterLinkOnClick = function(e) {
  const letter = e.target.id.substring("letter_".length)
  console.log(letter)
  if ("japanese" === letter) {
    // http://stackoverflow.com/questions/15033196/using-javascript-to-check-whether-a-string-contains-japanese-characters-includi
    searchInput.value = "[\\u3000-\\u303f\\u3040-\\u309f\\u30a0-\\u30ff\\uff00-\\uff9f\\u4e00-\\u9faf\\u3400-\\u4dbf]"
  } else {
    searchInput.value = "(artist ^" + letter + ")"
  }
  searchCatalog(searchInput.value, true)
}

const windowOnResize = function(e) {
  const windowHeight = window.innerHeight
  const quickSearchHeight = quickSearchDiv.getBoundingClientRect().height
  const controlsHeight = controlsDiv.getBoundingClientRect().height
  if (windowHeight > (controlsHeight - quickSearchHeight) * 3.3) {
    quickSearchDiv.style.display = "block"
  } else {
    quickSearchDiv.style.display = "none"
  }
}

const hashChange = function(event) {
  searchInput.value = deserializeState(document.location.hash).query
  doSearchCatalog(searchInput.value)
}

// M A I N

const addEventListeners = function() {
  nextButton.addEventListener("click", playNext)
  player.addEventListener("ended", playNext)
  player.addEventListener("error", playerOnError)
  player.addEventListener("loadedmetadata", playerLoadedMetadata)
  randomCheckbox.addEventListener("click", randomCheckboxOnClick)
  searchInput.addEventListener("blur", executeSearch)
  searchInput.addEventListener("keyup", searchInputOnKeyUp)
  searchButton.addEventListener("click", executeSearch)
  showHistoryButton.addEventListener("click", showHistoryButtonOnClick)
  videoCloseButton.addEventListener("click", closeVideo)
  window.addEventListener("resize", windowOnResize)
  window.addEventListener("scroll", windowOnScroll)
  document.body.addEventListener("keyup", togglePlayback)
  window.addEventListener("hashchange", hashChange, false);

  for (let i = 0; i < 26; i++) {
    $("letter_" + String.fromCharCode(97 + i)).addEventListener("click", letterLinkOnClick)
  }
  for (let i = 0; i < 10; i++) {
    $("letter_" + String.fromCharCode(48 + i)).addEventListener("click", letterLinkOnClick)
  }
  $("letter_japanese").addEventListener("click", letterLinkOnClick)
}

const applyState = function(serialized) {
  const state = deserializeState(serialized)
  searchInput.value = state.query
  const item = catalog[state.itemID]
  if (item) {
    player.itemID = state.itemID
    player.src = item[Pathname]
    displayNowPlaying(item, nowPlayingTitle)
  }

  searchHits = resetSearchHits()
  searchCatalog(state.query, true)
}

const main = function() {
  windowOnResize()
  getFormatExtensions()
  addEventListeners()
  applyState(document.location.hash.substring(1))
}
main()
