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
