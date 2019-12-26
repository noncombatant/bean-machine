// Copyright 2016 by Chris Palmer (https://noncombatant.org), and released under
// the terms of the GNU GPL3. See help.html for more information.

"use strict";

let player = audioPlayer
let searchHits = []

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

const preparePlay = function(item, itemID) {
  player.pause()
  setAudioVideoControls(item)
  player.src = blobCache[item.pathname] || item.pathname
  player.itemID = itemID
  localStorage.setItem("itemID", itemID)
  //player.currentTime = getTimeupdateForItemID(itemID)
  displayNowPlaying(item, nowPlayingTitle)
  searchCatalogFetchBudget++
}

let fetchSearchHitsInProgress = false
const blobCache = {}
const fetchSearchHits = function() {
  if (fetchSearchHitsInProgress || 0 === searchCatalogFetchBudget) {
    return
  }

  const item = searchHits[searchCatalogFetchIndex % searchHits.length]
  if (blobCache[item.pathname]) {
    searchCatalogFetchIndex++
    return
  }

  fetchSearchHitsInProgress = true
  fetch(item.pathname)
  .then(function(response) {
    return response.blob()
  })
  .then(function(blob) {
    blobCache[item.pathname] = URL.createObjectURL(blob)
    searchCatalogFetchIndex++
    searchCatalogFetchBudget--
    fetchSearchHitsInProgress = false
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

  const nameSpan = createElement("span", "itemDivCell", getName(item))
  div.appendChild(nameSpan)

  return div
}

const buildAlbumTitleDiv = function(item, itemID) {
  const div = createElement("div", "albumTitleDiv")

  const directory = dirname(item.pathname)

  const coverA = createElement("a")
  const coverImg = createElement("img")
  coverA.href = directory + "/media.html"
  coverImg.src = directory + "/cover"
  coverImg.height = coverImg.width = 64
  coverA.target = "cover"
  coverA.appendChild(coverImg)
  div.appendChild(coverA)

  const albumSpan = createElement("span", "itemDivCell albumTitle", getAlbum(item))
  div.appendChild(albumSpan)

  const artistSpan = createElement("span", "itemDivCell artistName", getArtist(item))
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
      searchHits.sort((a, b) => a.pathname.localeCompare(b.pathname))
    }
  } else {
    itemListDiv.removeChild($("bottom"))
  }

  const limit = Math.min(searchHits.length, 50)
  let i
  for (i = 0; i < limit && start + i < searchHits.length; ++i) {
    const itemID = start + i
    const item = searchHits[itemID]
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

const itemDivOnClick = function(e) {
  preparePlay(searchHits[this.itemID], this.itemID)
  player.play()
}

const windowOnScroll = function(e) {
  if (!haveRequestedExtendCatalog) {
    window.requestAnimationFrame(extendCatalog)
  }
  haveRequestedExtendCatalog = true
}

const displayNowPlaying = function(item, element) {
  removeAllChildren(element)
  element.appendChild(createElement("span", "", "“" + getName(item) + "” by "))
  element.appendChild(createElement("strong", "", getArtist(item)))
  element.appendChild(createElement("span", "", " from "))
  element.appendChild(createElement("em", "", getAlbum(item)))
  document.title = element.textContent
}

const playNext = function(e) {
  if (0 === searchHits.length) {
    return
  }
  const itemID = (player.itemID + 1) % searchHits.length
  preparePlay(searchHits[itemID], itemID)
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
  searchCatalog(localStorage.getItem("query") || "")
}

let searchCatalogFetchIndex = 0
let searchCatalogFetchBudget = 0
let haveSentCatalogToWorker = false

const searchCatalog = function(query) {
  query = query.trim()
  searchInput.value = query
  localStorage.setItem("query", query)
  const queryURL = "search?q=" + searchInput.value
  fetch(queryURL, {"credentials": "include"})
  .then(r => r.json())
  .then(j => {
    searchHits = j
    buildCatalog(0)
    searchCatalogFetchIndex = 0
    searchCatalogFetchBudget = 3
  })
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

const stripLeadingTrack = function(pathname) {
  return pathname.replace(/^(\d|-)+ /, "")
}

const getName = function(item) {
  return item.name || decodeURIComponent(stripLeadingTrack(stripFileExtension(basename(item.pathname))))
}

const getAlbum = function(item) {
  return item.album || decodeURIComponent(basename(dirname(item.pathname)))
}

const getArtist = function(item) {
  return item.artist || decodeURIComponent(basename(dirname(dirname(item.pathname))))
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

const main = function() {
  nextButton.addEventListener("click", playNext)
  player.addEventListener("ended", playNext)
  player.addEventListener("error", playerOnError)
  //player.addEventListener("timeupdate", playerOnTimeupdate)
  searchInput.addEventListener("keyup", searchInputOnKeyUp)
  searchButton.addEventListener("click", executeSearch)
  window.addEventListener("scroll", windowOnScroll)
  document.body.addEventListener("keyup", togglePlayback)
  randomCheckbox.addEventListener("click", randomCheckboxOnClick)
  restoreState()
  setInterval(fetchSearchHits, 2000)
}

main()