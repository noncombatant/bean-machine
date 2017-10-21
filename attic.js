const letterLinkOnClick = function(e) {
  const letter = e.target.id.substring("letter_".length)
  console.log(letter)
  if ("japanese" === letter) {
    // http://stackoverflow.com/questions/15033196/using-javascript-to-check-whether-a-string-contains-japanese-characters-includi
    searchInput.value = "[\\u3000-\\u303f\\u3040-\\u309f\\u30a0-\\u30ff\\uff00-\\uff9f\\u4e00-\\u9faf\\u3400-\\u4dbf]"
  } else {
    searchInput.value = "(artist ^" + letter + ")"
  }
  searchCatalog(searchInput.value, true)
}

const windowOnResize = function(e) {
  const windowHeight = window.innerHeight
  const quickSearchHeight = quickSearchDiv.getBoundingClientRect().height
  const controlsHeight = controlsDiv.getBoundingClientRect().height
  if (windowHeight > (controlsHeight - quickSearchHeight) * 3.3) {
    quickSearchDiv.style.display = "block"
  } else {
    quickSearchDiv.style.display = "none"
  }
}

  // in addEventListeners:
  for (let i = 0; i < 26; i++) {
    $("letter_" + String.fromCharCode(97 + i)).addEventListener("click", letterLinkOnClick)
  }
  for (let i = 0; i < 10; i++) {
    $("letter_" + String.fromCharCode(48 + i)).addEventListener("click", letterLinkOnClick)
  }
  $("letter_japanese").addEventListener("click", letterLinkOnClick)

const showHistoryButtonOnClick = function(e) {
  if ("Show History" === showHistoryButton.innerText) {
    showHistoryButton.innerText = "Show Search Results"
    showPlayHistory()
  } else {
    showHistoryButton.innerText = "Show History"
    doSearchCatalog(searchInput.value)
  }
}

