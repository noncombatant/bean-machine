// Copyright 2016 by Chris Palmer (https://noncombatant.org), and released under
// the terms of the GNU GPL3. See help.html for more information.

"use strict";

// G L O B A L   V A R I A B L E S
//
// TODO: Move globals to inside `main`. This requires explicitly passing and
// returning them to and from functions.

let player = audioPlayer
let playHistory = []

const buildCatalogLimit = 50

// C O R E   F U N C T I O N A L I T Y

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

const doPlay = function(itemID) {
  player.pause()
  setAudioVideoControls(itemID)
  const item = catalog[itemID]
  player.src = item[Pathname]
  player.itemID = itemID
  player.play()
  playHistory.unshift(itemID)

  displayNowPlaying(item, nowPlayingTitle)
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

const windowOnScroll = function(e) {
  if (!extendCatalogRequested) {
    window.requestAnimationFrame(extendCatalog)
  }
  extendCatalogRequested = true
}

// M A I N

const addEventListeners = function() {
  nextButton.addEventListener("click", playNext)
  player.addEventListener("ended", playNext)
  player.addEventListener("error", playerOnError)
  player.addEventListener("loadedmetadata", playerLoadedMetadata)
  searchInput.addEventListener("blur", executeSearch)
  searchInput.addEventListener("keyup", searchInputOnKeyUp)
  searchButton.addEventListener("click", executeSearch)
  videoCloseButton.addEventListener("click", closeVideo)
  window.addEventListener("scroll", windowOnScroll)
  document.body.addEventListener("keyup", togglePlayback)
}

main()
previousLastItem = buildCatalog(0)
