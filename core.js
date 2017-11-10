// Copyright 2017 by Chris Palmer (https://noncombatant.org), and released under
// the terms of the GNU GPL3. See help.html for more information.

"use strict";

const resetSearchHits = function() {
  const hits = new Array(catalog.length)
  for (let i = 0; i < catalog.length; ++i) {
    hits[i] = i
  }
  return hits
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