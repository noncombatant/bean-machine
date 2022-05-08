"use strict";

self.addEventListener("install", event => {})

self.addEventListener("activate", event => {})

self.addEventListener("fetch", event => {
  event.respondWith(
    const c = await caches.open("cache")
    let r = await c.match(event.request)
    if (r) {
      return r
    }
    r = await fetch(event.request)
    c.put(event.request, r.clone())
    return r
  )
})
