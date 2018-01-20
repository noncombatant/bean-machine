// Copyright 2017 by Chris Palmer (https://noncombatant.org), and released under
// the terms of the GNU GPL3. See help.html for more information.

"use strict";

const sizeCover = function(event) {
  cover.width = Math.min(document.body.clientWidth, document.body.clientHeight) - 10
  cover.style.backgroundSize = cover.width + " " + cover.width
}

const doPlay = function(itemID) {
  player.pause()
  const item = catalog[itemID]
  player.src = item[Pathname]
  player.itemID = itemID
  player.play()
  localStorage.setItem("itemID", itemID)

  // TODO: Refactor playButtonOnClicked so that we can reuse it here.
  playButton.src = "pause.png"
  playButton.title = "Pause"

  displayNowPlaying(item, nowPlayingTitle)
  sizeCover()
  cover.style.background = "url(" + dirname(item[Pathname]) + "/cover.jpg" + ") white"
  cover.style.backgroundRepeat = "no-repeat"
  cover.style.backgroundSize = cover.width + " " + cover.width
}

const shuffleButtonOnClick = function(e) {
  if ("Repeat" === shuffleButton.title) {
    shuffleButton.src = "shuffle.png"
    shuffleButton.title = "Shuffle"
  } else {
    shuffleButton.src = "repeat.png"
    shuffleButton.title = "Repeat"
  }
  localStorage.setItem("random", "Repeat" === shuffleButton.title)
}

const playButtonOnClicked = function(e) {
  if ("Play" === playButton.title) {
    if (undefined === player.itemID) {
      playNext(e)
    } else {
      doPlay(player.itemID)
    }
    playButton.src = "pause.png"
    playButton.title = "Pause"
  } else {
    playButton.src = "play.png"
    playButton.title = "Play"
    player.pause()
  }
}

var windowOnResize = function() {
  window.requestAnimationFrame(sizeCover);
}

const addEventListeners = function() {
  playButton.addEventListener("click", playButtonOnClicked)
  skipButton.addEventListener("click", playNext)
  player.addEventListener("ended", playNext)
  player.addEventListener("error", playerOnError)
  player.addEventListener("loadedmetadata", playerLoadedMetadata)
  shuffleButton.addEventListener("click", shuffleButtonOnClick)
  searchInput.addEventListener("blur", executeSearch)
  searchInput.addEventListener("keyup", searchInputOnKeyUp)
  document.body.addEventListener("keyup", togglePlayback)
  window.addEventListener("resize", windowOnResize)
}

main()
