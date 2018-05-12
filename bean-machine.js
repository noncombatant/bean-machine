// Copyright 2016 by Chris Palmer (https://noncombatant.org), and released under
// the terms of the GNU GPL3. See help.html for more information.

"use strict";

let player = audioPlayer

const buildCatalogLimit = 50

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
}

const shouldRequireLongPress = isAndroidDevice()

const buildItemDiv = function(itemID) {
  const item = catalog[itemID]
  const div = createElement("div", "itemDiv")
  div.itemID = itemID
  if (shouldRequireLongPress) {
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
  if (shouldRequireLongPress) {
    div.addEventListener("contextmenu", itemDivOnClick)
  } else {
    div.addEventListener("dblclick", itemDivOnClick)
    div.addEventListener("click", itemDivOnClick)
  }

  const albumSpan = createElement("span", "itemDivCell albumTitle", basename(dirname(item[Pathname])))
  div.appendChild(albumSpan)

  const artistSpan = createElement("span", "itemDivCell artistName", basename(dirname(dirname(item[Pathname]))))
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
  player.addEventListener("loadedmetadata", playerLoadedMetadata)
  searchInput.addEventListener("blur", executeSearch)
  searchInput.addEventListener("keyup", searchInputOnKeyUp)
  searchButton.addEventListener("click", executeSearch)
  videoCloseButton.addEventListener("click", closeVideo)
  window.addEventListener("scroll", windowOnScroll)
  document.body.addEventListener("keyup", togglePlayback)
  randomCheckbox.addEventListener("click", randomCheckboxOnClick)
}

main()
previousLastItem = buildCatalog(0)
