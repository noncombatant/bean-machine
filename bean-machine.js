// Copyright 2016 by Chris Palmer (https://noncombatant.org), and released under
// the terms of the GNU GPL3. See help.html for more information.

"use strict";

const catalog = []
const buildCatalogLimit = 50

let player = audioPlayer
let searchHits = []
let searchWorker

const setAudioVideoControls = function(itemID) {
  const pathname = catalog[itemID].pathname
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
  player.src = item.blobURL || item.pathname
  player.itemID = itemID
  if (shouldStartPlaying) {
    player.play()
  }
  localStorage.setItem("itemID", itemID)

  displayNowPlaying(item, nowPlayingTitle)
  populateArt(artSpan, dirname(item.pathname))
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
  if (item.blobURL) {
    searchCatalogFetchIndex++
    return
  }

  fetchSearchHitsInProgress = true
  fetch(item.pathname)
  .then(function(response) {
    return response.blob()
  })
  .then(function(blob) {
    item.blobURL = URL.createObjectURL(blob)
    searchCatalogFetchIndex++
    searchCatalogFetchBudget--
    fetchSearchHitsInProgress = false
  })
}

// TODO: Memoize this to save network traffic.
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

  const trackSpan = createElement("span", "itemDivCell trackNumber", (item.disc || "1") + "-" + (item.track || "1"))
  div.appendChild(trackSpan)

  const nameSpan = createElement("span", "itemDivCell", item.name)
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

  const albumSpan = createElement("span", "itemDivCell albumTitle", item.album)
  div.appendChild(albumSpan)

  const artistSpan = createElement("span", "itemDivCell artistName", item.artist)
  div.appendChild(artistSpan)

  if (item.year) {
    const yearSpan = createElement("span", "itemDivCell year", item.year)
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
    const albumPathname = dirname(item.pathname)
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

let haveRequestedExtendCatalog = false
const extendCatalog = function() {
  if (isElementInViewport($("bottom"))) {
    previousLastItem = buildCatalog(previousLastItem)
  }
  haveRequestedExtendCatalog = false
}

const albumTitleDivOnClick = function(e) {
  const itemID = this.itemID
  randomCheckbox.checked = false
  if (player.paused || player.itemID != itemID) {
    if (undefined !== itemID) {
      if (itemID === player.itemID) {
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
  if (!haveRequestedExtendCatalog) {
    window.requestAnimationFrame(extendCatalog)
  }
  haveRequestedExtendCatalog = true
}

const displayNowPlaying = function(item, element) {
  removeAllChildren(element)
  const trackName = item.name || basename(item.pathname)
  element.appendChild(createElement("span", "", item.disc + "-" + item.track + " “" + trackName + "”\u200A—\u200A"))
  element.appendChild(createElement("strong", "", item.artist))
  element.appendChild(createElement("span", "", "\u200A—\u200A"))
  element.appendChild(createElement("em", "", item.album))
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
  console.log("Could not load", catalog[player.itemID].pathname, e)
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

  searchCatalog(localStorage.getItem("query"), true)
}

const parseTSVRecords = function(tsvs, array) {
  const pathname = 0
  const album = 1
  const artist = 2
  const name = 3
  const disc = 4
  const track = 5
  const year = 6
  const genre = 7
  const mtime = 8

  let start = 0
  for (let i = 0; i < tsvs.length; ++i) {
    if ('\n' === tsvs[i]) {
      const record = tsvs.substring(start, i)
      const fields = record.split("\t")
      array.push({ pathname: fields[pathname],
                   album:    fields[album],
                   artist:   fields[artist],
                   name:     fields[name],
                   disc:     fields[disc],
                   track:    fields[track],
                   year:     fields[year],
                   genre:    fields[genre],
                   mtime:    fields[mtime] })
      start = i + 1
    }
  }
}

let searchCatalogFetchIndex = 0
let searchCatalogFetchBudget = 0
const searchCatalog = function(query, forceSearch) {
  query = query.trim() || getRandomWord()
  const previousQuery = localStorage.getItem("query")
  if (!forceSearch && previousQuery === query) {
    return
  }
  searchInput.value = query
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

const stopwords = new Set(["a", "an", "the", "le", "la"])
const getRandomWord = function() {
  while (true) {
    const item = catalog[getRandomIndex(catalog)]
    const words = item.pathname.split("/").join(" ").split(" ")
    const word = words[getRandomIndex(words)].toLowerCase()
    if (/\w/.test(word) && !/^[0-9_-]+$/.test(word) && !stopwords.has(word)) {
      return word
    }
  }
}

const main = function() {
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
