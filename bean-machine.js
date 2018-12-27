// Copyright 2016 by Chris Palmer (https://noncombatant.org), and released under
// the terms of the GNU GPL3. See help.html for more information.

"use strict";

let tsvs
const tsvOffsets = []
const buildCatalogLimit = 50

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
  const item = getItem(tsvs, itemID)
  setAudioVideoControls(item)
  player.src = item.blobURL || item.pathname
  player.itemID = itemID
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
  const item = getItem(tsvs, itemID)
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
    const item = getItem(tsvs, itemID)
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
  if (player.itemID !== this.itemID) {
    preparePlay(this.itemID)
    player.play()
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

  let index = 0
  if (randomCheckbox.checked) {
    index = getRandomIndex(searchHits)
  } else {
    const i = searchHits.indexOf(player.itemID)
    index = -1 === i ? 0 : (i + 1) % searchHits.length
  }
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
  const item = getItem(tsvs, player.itemID)
  speechSynthesis.speak(new SpeechSynthesisUtterance(`Could not play ${item.name} by ${item.artist}`))
}

const randomCheckboxOnClick = function(e) {
  localStorage.setItem("random", randomCheckbox.checked)
}

const restoreState = function() {
  randomCheckbox.checked = "true" === localStorage.getItem("random")

  let itemID = parseInt(localStorage.getItem("itemID"))
  if (itemID > tsvs.length || itemID < 0 || (itemID > 0 && "\n" !== tsvs[itemID - 1])) {
    itemID = 0
  }
  if (!Number.isNaN(itemID)) {
    preparePlay(itemID)
  }

  searchCatalog(localStorage.getItem("query") || "")
}

const getItem = function(tsvs, itemID) {
  const end = tsvs.indexOf("\n", itemID)
  const record = tsvs.substring(itemID, end === -1 ? undefined : end)
  const fields = record.split("\t")
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

const parseTSVRecords = function(tsvs, array) {
  array.push(0)
  for (let i = 0; i < tsvs.length; ++i) {
    if ("\n" === tsvs[i]) {
      array.push(i + 1)
    }
  }
}

let searchCatalogFetchIndex = 0
let searchCatalogFetchBudget = 0
let haveSentTsvsToWorker = false
const searchCatalog = function(query) {
  query = query.trim()
  if ("?" === query) {
    query = getRandomWord()
  }
  searchInput.value = query
  localStorage.setItem("query", query)
  const maybeTsvs = haveSentTsvsToWorker ? undefined : tsvs
  searchWorker.postMessage({tsvs: maybeTsvs, tsvOffsets: tsvOffsets, query: query})
  haveSentTsvsToWorker = true
}

const onMessageFromSearchWorker = function(e) {
  searchHits = e.data
  previousLastItem = buildCatalog(0)
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
    const item = getItem(tsvs, tsvOffsets[getRandomIndex(tsvOffsets)])
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
    tsvs = text
    parseTSVRecords(tsvs, tsvOffsets)
    restoreState()
  })

  setInterval(fetchSearchHits, 2000)
}
main()
