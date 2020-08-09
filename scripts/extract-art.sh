#!/bin/sh

file="$1"
directory=$(dirname "$file")

# find . -iname '*.mp3' -exec ~/src/bean-machine/extract-art.sh "{}" ';' | less

eyeD3 --quiet --write-images "$directory" "$file"

# --remove-all-images
# --remove-all-lyrics

# Then find all the OTHER_*.JPG and FRONT_COVER_*.JPG files and delete them.
