// Copyright 2022 by Chris Palmer (https://noncombatant.org), and released under
// the terms of the GNU GPL3. See web/index.html for more information.

"use strict";

const cacheName = "cache"

self.addEventListener("install", event => {
  console.log("install")
  // If we wanted to pre-cache anything:
  //event.waitUntil((async () => {
  //  const cache = await caches.open(cacheName);
  //  await cache.addAll(contentToCache);
  //})());
})

self.addEventListener("register", event => {
  console.log("register")
})

//self.addEventListener("activate", event => {
//  console.log("activate")
//})
//
//self.addEventListener("fetch", event => {
//  console.log("fetch")
//  event.respondWith((async () => {
//    console.log("responding...")
//    let r = await caches.match(event.request)
//    console.log("cached response?", r)
//    if (r) {
//      console.log("returning cached response")
//      return r
//    }
//
//    r = await fetch(event.request)
//    console.log("fetched response", r)
//    const c = await caches.open(cacheName)
//    c.put(event.request, r.clone())
//    console.log("put cloned response, returning r")
//    return r
//  })())
//})
