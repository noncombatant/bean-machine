// Copyright 2016 by Chris Palmer (https://noncombatant.org), and released under
// the terms of the GNU GPL3. See help.html for more information.

"use strict";

const Pathname = 0
const Album = 1
const Artist = 2
const Name = 3
const Disc = 4
const Track = 5
const Year = 6
const Genre = 7
const Mtime = 8
const catalog = []
const buildCatalogLimit = 50

let player = audioPlayer
let searchHits = []
let searchWorker

const setAudioVideoControls = function(itemID) {
  const pathname = catalog[itemID][Pathname]
  if (isAudioPathname(pathname)) {
    player = audioPlayer
    audioPlayer.className = ""
    videoPlayer.className = "hidden"
  } else if (isVideoPathname(pathname)) {
    player = videoPlayer
    audioPlayer.className = "hidden"
    videoPlayer.className = ""
  }
  player.className = "normal"
}

const doPlay = function(itemID, shouldStartPlaying) {
  player.pause()
  setAudioVideoControls(itemID)
  const item = catalog[itemID]
  player.src = item[Pathname]
  player.itemID = itemID
  if (shouldStartPlaying) {
    player.play()
  }
  localStorage.setItem("itemID", itemID)

  displayNowPlaying(item, nowPlayingTitle)
  populateArt(artSpan, dirname(item[Pathname]))
  searchCatalogFetchBudget++
}

let fetchSearchHitsInProgress = false
const fetchSearchHits = function() {
  if (randomCheckbox.checked ||
      fetchSearchHitsInProgress ||
      searchCatalogFetchIndex >= searchHits.length ||
      0 === searchCatalogFetchBudget)
  {
    return
  }

  const itemID = searchHits[searchCatalogFetchIndex]
  const item = catalog[itemID]
  if (item[Pathname].startsWith("blob:")) {
    searchCatalogFetchIndex++
    return
  }

  fetchSearchHitsInProgress = true
  fetch(item[Pathname])
  .then(function(response) {
    return response.blob()
  })
  .then(function(blob) {
    const blobURL = URL.createObjectURL(blob)
    item[Pathname] = blobURL
    searchCatalogFetchIndex++
    searchCatalogFetchBudget--
    fetchSearchHitsInProgress = false
  })
}

const populateArt = function(parentElement, directory) {
  removeAllChildren(parentElement)

  fetch("/getArt?d=" + encodeURIComponent(directory), {"credentials": "include"})
  .then(function(response) {
    return response.text()
  })
  .then(function(arts) {
    arts = arts.split("\n")
    let haveDoneFirst = false
    for (let art of arts) {
      if (0 == art.length) {
        continue
      }
      const a = document.createElement("a")
      a.href = directory + "/" + art
      a.target = "_blank"
      a.appendChild(document.createTextNode(stripFileExtension(art)))
      if (haveDoneFirst) {
        parentElement.appendChild(document.createTextNode(" "))
      }
      parentElement.appendChild(a)
      haveDoneFirst = true
    }
  })
}

const requireLongPress = /android/i.test(navigator.userAgent)

const buildItemDiv = function(itemID) {
  const item = catalog[itemID]
  const div = createElement("div", "itemDiv")
  div.itemID = itemID
  if (requireLongPress) {
    div.addEventListener("contextmenu", itemDivOnClick)
  } else {
    div.addEventListener("dblclick", itemDivOnClick)
    div.addEventListener("click", itemDivOnClick)
  }

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
  if (requireLongPress) {
    div.addEventListener("contextmenu", itemDivOnClick)
  } else {
    div.addEventListener("dblclick", itemDivOnClick)
    div.addEventListener("click", itemDivOnClick)
  }

  const albumSpan = createElement("span", "itemDivCell albumTitle", item[Album])
  div.appendChild(albumSpan)

  const artistSpan = createElement("span", "itemDivCell artistName", item[Artist])
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
    currentAlbumPathname = ""
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

const albumTitleDivOnClick = function(e) {
  const itemID = this.itemID
  randomCheckbox.checked = false
  if (player.paused || player.itemID != itemID) {
    if (undefined !== itemID) {
      if (itemID == player.itemID) {
        player.play()
      } else {
        doPlay(itemID, true)
      }
    } else {
      playNext()
    }
  } else {
    player.pause()
  }
}
const itemDivOnClick = albumTitleDivOnClick

const windowOnScroll = function(e) {
  if (!extendCatalogRequested) {
    window.requestAnimationFrame(extendCatalog)
  }
  extendCatalogRequested = true
}

const addEventListeners = function() {
  nextButton.addEventListener("click", playNext)
  player.addEventListener("ended", playNext)
  player.addEventListener("error", playerOnError)
  player.addEventListener("loadedmetadata", playerOnLoadedMetadata)
  searchInput.addEventListener("blur", executeSearch)
  searchInput.addEventListener("keyup", searchInputOnKeyUp)
  searchButton.addEventListener("click", executeSearch)
  window.addEventListener("scroll", windowOnScroll)
  document.body.addEventListener("keyup", togglePlayback)
  randomCheckbox.addEventListener("click", randomCheckboxOnClick)
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

const playNext = function(e) {
  if (randomCheckbox.checked) {
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
const playerOnLoadedMetadata = function(e) {
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
  randomCheckbox.checked = "true" === localStorage.getItem("random")
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
}

const parseTSVRecords = function(tsvs, array) {
  let start = 0
  for (let i = 0; i < tsvs.length; ++i) {
    if ('\n' === tsvs[i]) {
      const record = tsvs.substring(start, i)
      array.push(record.split("\t"))
      start = i + 1
    }
  }
}

let searchCatalogFetchIndex = 0
let searchCatalogFetchBudget = 0
const searchCatalog = function(query, forceSearch) {
  query = query.trim()
  const previousQuery = localStorage.getItem("query")
  if (!forceSearch && previousQuery === query) {
    return
  }
  localStorage.setItem("query", query)
  searchWorker.postMessage({catalog: catalog, query: query})
}

const onMessageFromSearchWorker = function(e) {
  searchHits = e.data
  previousLastItem = buildCatalog(0)
  searchCatalogFetchIndex = 0
  searchCatalogFetchBudget = 3
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

const $ = function(id) {
  return document.getElementById(id)
}

const isElementInViewport = function(element) {
  if (!element) {
    return false
  }

  let top = element.offsetTop
  let left = element.offsetLeft
  const width = element.offsetWidth
  const height = element.offsetHeight

  while (element.offsetParent) {
    element = element.offsetParent
    top += element.offsetTop
    left += element.offsetLeft
  }

  return top >= window.pageYOffset &&
      left >= window.pageXOffset &&
      (top + height) <= (window.pageYOffset + window.innerHeight) &&
      (left + width) <= (window.pageXOffset + window.innerWidth)
}

const createElement = function(type, className, text) {
  const e = document.createElement(type)
  if (className) {
    e.className = className
  }
  if (text) {
    setSingleTextChild(e, text)
  }
  return e
}

const setSingleTextChild = function(element, text) {
  (element.childNodes[0] || element.appendChild(document.createTextNode("")))
      .data = text
}

const removeAllChildren = function(element) {
  while (element.firstChild) {
    element.removeChild(element.firstChild)
  }
}

const all = function(array, predicate) {
  for (let i in array) {
    if (!predicate(array[i])) {
      return false
    }
  }
  return true
}

const any = function(array, predicate) {
  for (let i in array) {
    if (predicate(array[i])) {
      return true
    }
  }
  return false
}

const basename = function(pathname) {
  const i = pathname.lastIndexOf("/")
  return -1 == i ? pathname : pathname.substring(i + 1)
}

const dirname = function(pathname) {
  return pathname.substring(0, pathname.lastIndexOf("/"))
}

const fileExtension = function(pathname) {
  const i = pathname.lastIndexOf(".")
  return -1 == i ? "" : pathname.substring(i)
}

const stripFileExtension = function(pathname) {
  const i = pathname.lastIndexOf(".")
  return -1 == i ? pathname : pathname.substring(0, i)
}

const isPathnameInExtensions = function(pathname, extensions) {
  const e = fileExtension(pathname)
  return any(extensions, function(extension) { return e == extension })
}

// NOTE: These must be kept in sync with the format extensions arrays in the Go
// code.
const audioFormatExtensions = [
  ".flac",
  ".m4a",
  ".mid",
  ".midi",
  ".mp3",
  ".ogg",
  ".wav",
  ".wave",
]
const videoFormatExtensions = [
  ".avi",
  ".mkv",
  ".mov",
  ".mp4",
  ".mpeg",
  ".mpg",
  ".ogv",
  ".webm",
]

const isAudioPathname = function(pathname) {
  return isPathnameInExtensions(pathname, audioFormatExtensions)
}

const isVideoPathname = function(pathname) {
  return isPathnameInExtensions(pathname, videoFormatExtensions)
}

const getRandomIndex = function(array) {
  return Math.floor(Math.random() * array.length)
}

const main = function() {
  addEventListeners()
  searchWorker = new Worker("search.js")
  searchWorker.addEventListener("message", onMessageFromSearchWorker)

  fetch("catalog.tsv", {"credentials": "include"})
  .then(function(response) {
    return response.text()
  })
  .then(function(tsvs) {
    parseTSVRecords(tsvs, catalog)
    restoreState()
  })

  setInterval(fetchSearchHits, 2000)
}
main()
