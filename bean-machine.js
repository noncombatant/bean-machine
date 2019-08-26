// Copyright 2016 by Chris Palmer (https://noncombatant.org), and released under
// the terms of the GNU GPL3. See help.html for more information.

"use strict";

let catalog
const itemIDs = []
const recordSeparator = "\n"
const fieldSeparator = "\t"

let player = audioPlayer
let searchHits = []
let searchWorker

const setAudioVideoControls = function(item) {
  if (isAudioPathname(item.pathname)) {
    player = audioPlayer
    audioPlayer.className = ""
    videoPlayer.className = "hidden"
  } else if (isVideoPathname(item.pathname)) {
    player = videoPlayer
    audioPlayer.className = "hidden"
    videoPlayer.className = ""
  }
  player.className = "normal"
}

const preparePlay = function(itemID) {
  player.pause()
  const item = getItem(catalog, itemID)
  setAudioVideoControls(item)
  player.src = blobCache[itemID] || item.pathname
  player.itemID = itemID
  localStorage.setItem("itemID", itemID)
  player.currentTime = getTimeupdateForItemID(itemID)
  displayNowPlaying(item, nowPlayingTitle)
  populateArt(artSpan, dirname(item.pathname))
  searchCatalogFetchBudget++
  searchCatalogFetchIndex = searchHits.indexOf(itemID) + 1
}

let fetchSearchHitsInProgress = false
const blobCache = {}
const fetchSearchHits = function() {
  if (fetchSearchHitsInProgress || 0 === searchCatalogFetchBudget) {
    return
  }

  const itemID = searchHits[searchCatalogFetchIndex % searchHits.length]
  const item = getItem(catalog, itemID)
  if (blobCache[itemID]) {
    searchCatalogFetchIndex++
    return
  }

  fetchSearchHitsInProgress = true
  fetch(item.pathname)
  .then(function(response) {
    return response.blob()
  })
  .then(function(blob) {
    blobCache[itemID] = URL.createObjectURL(blob)
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
    arts = arts.split(recordSeparator)
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

const buildItemDiv = function(item, itemID) {
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

const buildAlbumTitleDiv = function(item, itemID) {
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
let haveRequestedExtendCatalog = false
const buildCatalog = function(start) {
  if (0 === start) {
    removeAllChildren(itemListDiv)
    currentAlbumPathname = ""
    haveRequestedExtendCatalog = false
    if (randomCheckbox.checked) {
      shuffle(searchHits)
    } else {
      searchHits.sort((a, b) => a - b)
    }
  } else {
    itemListDiv.removeChild($("bottom"))
  }

  const limit = Math.min(searchHits.length, 50)
  let i
  for (i = 0; i < limit && start + i < searchHits.length; ++i) {
    const itemID = searchHits[start + i]
    const item = getItem(catalog, itemID)
    const albumPathname = dirname(item.pathname)
    if (albumPathname !== currentAlbumPathname) {
      itemListDiv.appendChild(buildAlbumTitleDiv(item, itemID))
      currentAlbumPathname = albumPathname
    }
    itemListDiv.appendChild(buildItemDiv(item, itemID))
  }

  const bottom = createElement("div")
  bottom.id = "bottom"
  itemListDiv.appendChild(bottom)
  previousLastItem = start + i
}

const extendCatalog = function() {
  if (isElementInViewport($("bottom"))) {
    buildCatalog(previousLastItem)
  }
  haveRequestedExtendCatalog = false
}

const albumTitleDivOnClick = function(e) {
  if (player.itemID !== this.itemID) {
    preparePlay(this.itemID)
  }
  player.play()
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
  element.appendChild(createElement("span", "", "“" + trackName + "” by "))
  element.appendChild(createElement("strong", "", item.artist))
  element.appendChild(createElement("span", "", " from "))
  element.appendChild(createElement("em", "", item.album))
  document.title = element.textContent
}

const playNext = function(e) {
  if (0 === searchHits.length) {
    return
  }

  const i = searchHits.indexOf(player.itemID)
  const index = -1 === i ? 0 : (i + 1) % searchHits.length
  preparePlay(searchHits[index])
  player.play()
}

const togglePlayback = function(e) {
  e.stopPropagation()
  if ("p" !== e.key) {
    return
  }
  player[player.paused ? "play" : "pause"]()
}

const playerOnError = function(e) {
  this.dispatchEvent(new Event("ended"))
}

const getTimeupdateForItemID = function(itemID) {
  const key = "timeupdate" + itemID
  const time = parseInt(localStorage.getItem(key))
  return Number.isNaN(time) ? 0 : time
}

const playerOnTimeupdate = function(e) {
  const time = getTimeupdateForItemID(player.itemID)
  if (player.currentTime < time || (player.currentTime - time) > 5) {
    const key = "timeupdate" + player.itemID
    localStorage.setItem(key, player.currentTime)
  }
}

// https://stackoverflow.com/questions/2450954/how-to-randomize-shuffle-a-javascript-array
const shuffle = function(array) {
  let currentIndex = array.length
  while (0 !== currentIndex) {
    const randomIndex = Math.floor(Math.random() * currentIndex)
    currentIndex--
    const temp = array[currentIndex]
    array[currentIndex] = array[randomIndex]
    array[randomIndex] = temp
  }
}

const randomCheckboxOnClick = function(e) {
  buildCatalog(0)
  localStorage.setItem("random", randomCheckbox.checked)
}

const restoreState = function() {
  randomCheckbox.checked = "true" === localStorage.getItem("random")

  let itemID = parseInt(localStorage.getItem("itemID"))
  if (itemID > catalog.length || itemID < 0 || (itemID > 0 && recordSeparator !== catalog[itemID - 1])) {
    itemID = 0
  }
  if (!Number.isNaN(itemID)) {
    preparePlay(itemID)
  }

  searchCatalog(localStorage.getItem("query") || "")
}

const getItem = function(catalog, itemID) {
  const end = catalog.indexOf(recordSeparator, itemID)
  const record = catalog.substring(itemID, end === -1 ? undefined : end)
  const fields = record.split(fieldSeparator)
  return { pathname: fields[0],
           album:    fields[1],
           artist:   fields[2],
           name:     fields[3],
           disc:     fields[4],
           track:    fields[5],
           year:     fields[6],
           genre:    fields[7],
           mtime:    fields[8] }
}

const parseCatalogRecords = function(catalog, array) {
  array.push(0)
  for (let i = 0; i < catalog.length; ++i) {
    if (recordSeparator === catalog[i]) {
      array.push(i + 1)
    }
  }
}

let searchCatalogFetchIndex = 0
let searchCatalogFetchBudget = 0
let haveSentCatalogToWorker = false
const searchCatalog = function(query) {
  query = query.trim()
  if ("?" === query) {
    query = getRandomWord()
  }
  searchInput.value = query
  localStorage.setItem("query", query)
  const maybeCatalog = haveSentCatalogToWorker ? undefined : catalog
  searchWorker.postMessage({catalog: maybeCatalog, itemIDs: itemIDs, query: query})
  haveSentCatalogToWorker = true
}

const onMessageFromSearchWorker = function(e) {
  searchHits = e.data
  buildCatalog(0)
  searchCatalogFetchIndex = 0
  searchCatalogFetchBudget = 3
}

const executeSearch = function(e) {
  searchCatalog(searchInput.value)
}

const searchInputOnKeyUp = function(e) {
  e.stopPropagation()
  if ("Enter" === e.code) {
    searchCatalog(this.value)
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
  e.className = className
  setSingleTextChild(e, text)
  return e
}

const setSingleTextChild = function(element, text) {
  (element.childNodes[0] || element.appendChild(document.createTextNode("")))
      .data = text || ""
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
    const item = getItem(catalog, itemIDs[getRandomIndex(itemIDs)])
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
  player.addEventListener("timeupdate", playerOnTimeupdate)
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
  .then(function(text) {
    catalog = text
    parseCatalogRecords(catalog, itemIDs)
    restoreState()
  })

  setInterval(fetchSearchHits, 2000)
}
main()

// TODO: Turn everything into ES modules, and export only `main`. Or, at least
// create a utilities.js module, and put shared stuff in there (like `memoize`).
