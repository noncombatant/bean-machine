// Copyright 2016 by Chris Palmer (https://noncombatant.org), and released under
// the terms of the GNU GPL3. See web/index.html for more information.

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

const preparePlay = function(itemID) {
  if ("undefined" !== typeof(player.itemID)) {
    // Clear possible `nowPlayingItemDiv` style:
    $("itemDiv" + player.itemID).className = "itemDiv"
  }

  player.pause()
  const item = searchHits[itemID]
  setAudioVideoControls(item)
  player.src = blobCache[item.pathname] || item.pathname
  player.itemID = itemID
  displayNowPlaying(item, nowPlayingTitle)
  searchCatalogFetchIndex = itemID + 1
  searchCatalogFetchBudget++
  searchHitsUpdated = false
  const itemDiv = $("itemDiv" + itemID)
  itemDiv.scrollIntoView({behavior: "smooth", block: "center"})
  itemDiv.className = "itemDiv nowPlayingItemDiv"
  if (getArtist(item).toLocaleLowerCase().includes("prince")) {
    purpulate(themeTopColor, themeBottomColor)
  }
}

let fetchSearchHitsInProgress = false
const blobCache = {}
const fetchSearchHits = function() {
  if (fetchSearchHitsInProgress || 0 === searchCatalogFetchBudget || 0 === searchHits.length) {
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
  div.title = decodeURI(item.pathname)
  div.itemID = itemID
  div.id = "itemDiv" + itemID
  if (requireLongPress) {
    div.addEventListener("contextmenu", itemDivOnClick)
  } else {
    div.addEventListener("dblclick", itemDivOnClick)
    div.addEventListener("click", itemDivOnClick)
  }

  const trackSpan = createElement("span", "itemDivCell secondaryMetadata", getDiscAndTrack(item))
  div.appendChild(trackSpan)

  const nameSpan = createElement("span", "itemDivCell", getName(item))
  div.appendChild(nameSpan)

  const genreSpan = createElement("span", "itemDivCell secondaryMetadata", getGenre(item))
  div.appendChild(genreSpan)

  return div
}

const buildAlbumTitleDiv = function(item, itemID) {
  const div = createElement("div", "albumTitleDiv")

  const directory = dirname(item.pathname)

  const coverA = createElement("a")
  const coverImg = createElement("img")
  coverA.href = directory + "/media.html"
  coverImg.src = directory + "/cover"
  coverImg.title = "Album art and inserts"
  coverImg.height = coverImg.width = 32
  coverImg.loading = "lazy"
  coverA.target = "cover"
  coverA.appendChild(coverImg)
  div.appendChild(coverA)

  const downloadA = createElement("a")
  const downloadImg = createElement("img")
  downloadA.href = directory + "?download"
  downloadImg.src = "/download.png"
  downloadImg.title = "Download album"
  downloadImg.height = downloadImg.width = 32
  downloadImg.loading = "lazy"
  downloadA.target = "download"
  downloadA.appendChild(downloadImg)
  div.appendChild(downloadA)

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
const maxItemsPerDraw = 500

const buildCatalog = function(start) {
  if (0 === start) {
    removeAllChildren(itemListDiv)
    currentAlbumPathname = ""
    haveRequestedExtendCatalog = false
    if ("true" === localStorage.getItem("shuffle")) {
      shuffle(searchHits)
    } else {
      searchHits.sort((a, b) => a.pathname.localeCompare(b.pathname))
    }
  } else {
    itemListDiv.removeChild($("bottom"))
  }

  const limit = Math.min(searchHits.length, maxItemsPerDraw)
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

const itemDivOnClick = function(event) {
  preparePlay(this.itemID)
  player.play()
}

const windowOnScroll = function(event) {
  if (!haveRequestedExtendCatalog) {
    window.requestAnimationFrame(extendCatalog)
  }
  haveRequestedExtendCatalog = true
}

const closeHelpButtonOnClick = function(event) {
  helpDiv.style.display = "none"
  controlsDiv.style.display = itemListDiv.style.display = "block"
}

const displayNowPlaying = function(item, element) {
  removeAllChildren(element)
  element.appendChild(createElement("span", "", "“" + getName(item) + "” by "))
  element.appendChild(createElement("strong", "", getArtist(item)))
  element.appendChild(createElement("span", "", " from "))
  element.appendChild(createElement("em", "", getAlbum(item)))
  document.title = element.textContent
}

let searchHitsUpdated = false

const playNext = function(event) {
  if (0 === searchHits.length || undefined === player.itemID) {
    return
  }

  let itemID = 0
  if (searchHitsUpdated) {
    searchHitsUpdated = false
  } else {
    itemID = (player.itemID + 1) % searchHits.length
  }
  preparePlay(itemID)
  player.play()
}

window.onkeydown = function(event) {
  // Return false on space, so that we don't scroll down. We reserve Space for
  // `playButtonOnClick` in `bodyOnKeyup`. We have to set this as *the* event
  // listener, not use `addEventListener`.
  return !(" " === event.key && event.target === document.body)
}

const bodyOnKeyup = function(event) {
  event.stopPropagation()
  switch (event.key) {
    case "ArrowRight":
    case "n":
      if (undefined === player.itemID) {
        playButtonOnClick(event)
        break
      }
      playNext()
      break
    case "p":
    case " ":
      playButtonOnClick(event)
      break
    case "s":
      shuffleButtonOnClick(event)
      break
    case "/":
      searchInput.focus()
      searchInput.select()
      break
    case "?":
    case "h":
      helpDiv.style.display = helpDiv.style.display || "none"
      helpDiv.style.display = "none" === helpDiv.style.display ? "block" : "none"
      controlsDiv.style.display = itemListDiv.style.display = "none" === itemListDiv.style.display ? "block" : "none"
      break
    case "Escape":
      closeHelpButtonOnClick()
  }
}

const togglePlayback = function() {
  player[player.paused ? "play" : "pause"]()
}

const playerOnError = function(event) {
  this.dispatchEvent(new Event("ended"))
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

const restoreState = function() {
  const shuffleOn = "true" === localStorage.getItem("shuffle")
  shuffleButton.title = shuffleOn ? "Sort (s)" : "Shuffle (s)"
  shuffleButton.innerText = shuffleOn ? "Sort" : "Shuffle"
  searchCatalog(localStorage.getItem("query") || "")
}

let searchCatalogFetchIndex = 0
let searchCatalogFetchBudget = 0

const searchCatalog = function(query) {
  query = query.trim()
  searchInput.value = query
  localStorage.setItem("query", query)
  const queryURL = "search?q=" + searchInput.value

  const progressTimeout = setTimeout(function() {
    removeAllChildren(itemListDiv)
    setSingleTextChild(itemListDiv, "Loadin’ up yer tunez...")
  }, 250)
  fetch(queryURL, {"credentials": "include"})
  .then(r => r.json())
  .then(j => {
    searchHits = j
    searchHitsUpdated = true
    clearTimeout(progressTimeout)
    buildCatalog(0)
    searchCatalogFetchIndex = 0
    searchCatalogFetchBudget = 3
  })
}

const executeSearch = function(event) {
  searchCatalog(searchInput.value)
}

const searchInputOnKeyUp = function(event) {
  event.stopPropagation()
  if ("Enter" === event.code) {
    searchCatalog(this.value)
  }
}

const playButtonOnClick = function(event) {
  if (undefined === player.itemID) {
    preparePlay(0)
  }
  togglePlayback()
}

const shuffleButtonOnClick = function(event) {
  const shuffleOn = "true" === localStorage.getItem("shuffle")
  shuffleButton.title = shuffleOn ? "Shuffle (s)" : "Sort (s)"
  shuffleButton.innerText = shuffleOn ? "Shuffle" : "Sort"
  localStorage.setItem("shuffle", shuffleOn ? "false" : "true")
  buildCatalog(0)
}

const $ = function(id) {
  return document.getElementById(id)
}

// TODO: This doesn't seem to return true in all the cases that it should.
const isElementInViewport = function(e) {
  if (!e) {
    return false
  }

  let top = e.offsetTop
  let left = e.offsetLeft
  const width = e.offsetWidth
  const height = e.offsetHeight

  while (e.offsetParent) {
    e = e.offsetParent
    top += e.offsetTop
    left += e.offsetLeft
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

const getDiscAndTrack = function(item) {
  return item.track.replace(/^0*/, "").replace(/(\/\d*)/, "")
}

const getName = function(item) {
  return item.name || decodeURIComponent(stripLeadingTrack(stripFileExtension(basename(item.pathname)))) || "Unknown Track"
}

const getAlbum = function(item) {
  return item.album || decodeURIComponent(basename(dirname(item.pathname))) || "Unknown Album"
}

const getArtist = function(item) {
  return item.artist || decodeURIComponent(basename(dirname(dirname(item.pathname)))) || "Unknown Artist"
}

const getGenre = function(item) {
  return item.genre || ""
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
  ".m4v",
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

const randomRGB = function() {
  return Math.min(Math.floor(Math.random() * 255 + 1), 255)
}

const randomColor = function() {
  return [randomRGB(), randomRGB(), randomRGB()]
}

const colorToString = function(c) {
  return "rgb(" + c[0] + "," + c[1] + "," + c[2] + ", 1.0)"
}

const setThemeColors = function(topColor, bottomColor) {
  const t = colorToString(topColor)
  const b = colorToString(bottomColor)
  document.querySelector("meta[name=theme-color]").setAttribute("content", t)
  controlsDiv.style.background = itemListDiv.style.background = "linear-gradient(to bottom," + t + "," + b + ")"
}

// TODO: Get rid of all globals
let themeTopColor, themeBottomColor

const changeThemeColor = function() {
  setThemeColors(themeTopColor = randomColor(), themeBottomColor = randomColor())
}

const purpulate = function(topColor, bottomColor) {
  topColor[0] += 5
  topColor[1] -= 5
  topColor[2] += 5
  bottomColor[0] -= 5
  bottomColor[1] += 5
  bottomColor[2] -= 5
  setThemeColors(topColor, bottomColor)
}

const main = function() {
  if ("serviceWorker" in navigator) {
    navigator.serviceWorker.register("sw.js");
  }

  player.addEventListener("ended", playNext)
  player.addEventListener("error", playerOnError)
  shuffleButton.addEventListener("click", shuffleButtonOnClick)
  searchInput.addEventListener("keyup", searchInputOnKeyUp)
  searchButton.addEventListener("click", executeSearch)
  closeHelpButton.addEventListener("click", closeHelpButtonOnClick)
  window.addEventListener("scroll", windowOnScroll)
  document.body.addEventListener("keyup", bodyOnKeyup)

  if (requireLongPress) {
    nowPlayingTitle.innerText = "Long-press on any track to play."
  }
  restoreState()
  setInterval(fetchSearchHits, 2000)
  changeThemeColor()
  setInterval(changeThemeColor, 60 * 60 * 1000)
}

main()
