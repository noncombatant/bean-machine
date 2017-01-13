// Copyright 2016 by Chris Palmer (https://noncombatant.org), and released under
// the terms of the GNU GPL3. See help.html for more information.

const $ = function(id) {
  return document.getElementById(id)
}

const isElementInViewport = function(element) {
  let top = element.offsetTop
  let left = element.offsetLeft
  let width = element.offsetWidth
  let height = element.offsetHeight

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

const scrollElementIntoView = function(element) {
  if (isElementInViewport(element)) {
    return
  }
  element.scrollIntoView({behavior: "smooth"})
}

const createElement = function(type, className, text) {
  let e = document.createElement(type)
  if (className) {
    e.className = className
  }
  if (text) {
    setSingleTextChild(e, text)
  }
  return e
}

const setSingleTextChild = function(element, text) {
  (element.childNodes[0] || element.appendChild(document.createTextNode("")))
      .data = text
}

const removeAllChildren = function(element) {
  while (element.firstChild) {
    element.removeChild(element.firstChild)
  }
}

const all = function(array, predicate) {
  for (let i in array) {
    if (!predicate(array[i])) {
      return false
    }
  }
  return true
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
  let i = pathname.lastIndexOf("/")
  return -1 == i ? pathname : pathname.substring(i + 1)
}

const dirname = function(pathname) {
  return pathname.substring(0, pathname.lastIndexOf("/"))
}

const fileExtension = function(pathname) {
  let i = pathname.lastIndexOf(".")
  return -1 == i ? "" : pathname.substring(i)
}

const isPathnameInExtensions = function(pathname, extensions) {
  let e = fileExtension(pathname)
  return any(extensions, function(extension) { return e == extension })
}

let formatExtensions = { "started": false }
const getFormatExtensions = function() {
  if (!formatExtensions.started) {
    formatExtensions.started = true
    let xhr = new XMLHttpRequest()
    xhr.addEventListener("load", function() {
      formatExtensions = JSON.parse(this.responseText)
      formatExtensions.started = true
    })
    xhr.addEventListener("error", function() {
      console.log("Could not load formats.json", this.statusText)
    })
    xhr.open("GET", "formats.json")
    xhr.send()
  }
  return formatExtensions
}

const isAudioPathname = function(pathname) {
  return isPathnameInExtensions(pathname, getFormatExtensions().audio)
}

const isVideoPathname = function(pathname) {
  return isPathnameInExtensions(pathname, getFormatExtensions().video)
}

const getRandomIndex = function(array) {
  return Math.floor(Math.random() * array.length)
}

const getRandomIndexWithoutRepeating = function(array, historyObject) {
  if (countProperties(historyObject) === array.length) {
    return undefined
  }
  let i
  do {
    i = getRandomIndex(array)
  } while (historyObject.hasOwnProperty(i))
  return i
}

const get2MonthsAgo = function() {
  let now = (new Date()).getTime() / 1000
  return now - (2 * 30 * 24 * 60 * 60)
}

const showNotification = function(title, options) {
  if (!("Notification" in window) || "denied" === window.Notification.permission) {
    return
  }

  let n
  if ("granted" === window.Notification.permission) {
    n = new window.Notification(title, options)
  } else {
    Notification.requestPermission(function (permission) {
      if ("granted" === permission) {
        n = new window.Notification(title, options)
      }
    })
  }
  setTimeout(function() { n.close() }, 5000)
}

const countProperties = function(object) {
  // According to http://jsben.ch/#/oSt2p, this method is the fastest in Chrome,
  // Safari, and Firefox — at least for objects with relatively few properties.
  return Object.keys(object).length

  // If that ceases to be true, there's always this:
  //let i = 0
  //for (let p in object) {
  //  if (object.hasOwnProperty(p)) {
  //    ++i
  //  }
  //}
  //return i
}

const parseIntOr = function(string, fallback, base) {
  const n = parseInt(string, base || 10)
  if (Number.isNaN(n))
    return undefined === fallback ? 0 : fallback
  return n
}

const parseQueryString = function(string) {
  let result = {}
  for (let pair of string.split("&")) {
    let kv = pair.split("=", 2).map(decodeURIComponent)
    let key = kv[0], value = kv[1]
    if (result[key]) {
      result[key].push(value)
    } else {
      result[key] = [value]
    }
  }
  return result
}

const constructQueryString = function(object) {
  let result = []
  for (let key in object) {
    if (!object.hasOwnProperty(key)) {
      continue
    }

    let value = object[key]
    if (Array !== value.constructor) {
      result.push(encodeURIComponent(key.toString()) + "=" + encodeURIComponent(value.toString()))
    } else {
      for (let v of value) {
        result.push(encodeURIComponent(key.toString()) + "=" + encodeURIComponent(v.toString()))
      }
    }
  }
  return result.join("&")
}

const idOrLast = function(x) {
  return Array.isArray(x) ? x[x.length - 1] : x
}
