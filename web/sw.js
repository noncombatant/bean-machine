// Originally adapted from a Service Worker example from Google, which carried
// this license:
//
// Copyright 2016 Google Inc. All Rights Reserved. Licensed under the Apache
// License, Version 2.0 (the "License"); you may not use this file except in
// compliance with the License. You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0 Unless required by applicable law
// or agreed to in writing, software distributed under the License is
// distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied. See the License for the specific language
// governing permissions and limitations under the License.

"use strict";

const PreCache = "precache-v2"
const RunTimeCache = "runtime"

const PreCacheURLs = [
  "./", // Alias for index.html
  "clef-512.png",
  "close.png",
  "help.png",
  "index.css",
  "index.html",
  "index.js",
  "pause.png",
  "play.png",
  "repeat.png",
  "search.png",
  "skip.png",
  "shuffle.png",
]

self.addEventListener("install", event => {
console.log("install")
  event.waitUntil(
    caches.open(PreCache)
      .then(cache => cache.addAll(PreCacheURLs))
      .then(self.skipWaiting())
  )
})

self.addEventListener("activate", event => {
//  const currentCaches = [PreCache, RunTimeCache]
//  event.waitUntil(
//    caches.keys().then(cacheNames => {
//      return cacheNames.filter(cacheName => !currentCaches.includes(cacheName))
//    }).then(cachesToDelete => {
//      return Promise.all(cachesToDelete.map(cacheToDelete => {
//        return caches.delete(cacheToDelete)
//      }))
//    }).then(() => self.clients.claim())
//  )
})

self.addEventListener("fetch", event => {
//  if (!event.request.url.startsWith(self.location.origin)) {
//    return
//  }
//
//  // Just do a stupid pass-through, no caching.
//  return fetch(event.request, {"credentials": "include"}).then(response => {
//    return response
//  })
//
//  event.respondWith(
//    caches.match(event.request).then(cachedResponse => {
//      if (cachedResponse) {
//        return cachedResponse
//      }
//
//      return caches.open(RunTimeCache).then(cache => {
//        return fetch(event.request).then(response => {
//          // Put a copy of the response in the runtime cache.
//          return cache.put(event.request, response.clone()).then(() => {
//            return response
//          })
//        })
//      })
//    })
//  )
})
