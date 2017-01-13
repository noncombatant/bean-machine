#!/bin/sh

if test -s cover.jpg; then
  exit 0
fi

find . -iname '*.mp3' -depth 1 -print0 | xargs -0 -n1 -Ix ffmpeg -v quiet -i x x.jpg
find . -iname '*.m4a' -depth 1 -print0 | xargs -0 -n1 -Ix ffmpeg -v quiet -i x x.jpg

shasum *.jpg | awk -F'  ' '
  {
    sums[$1] = $2
  }
  END {
    i = 0
    for (key in sums) {
      system(sprintf("mv -i \"%s\" cover%s.jpg", sums[key], (i ? i : "")))
      i++
    }
  }
'

find . -iname '*.mp3.jpg' -depth 1 -print0 | xargs -0 rm
find . -iname '*.m4a.jpg' -depth 1 -print0 | xargs -0 rm
