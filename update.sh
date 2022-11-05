#!/usr/bin/env bash

help="Updates the catalog and synchronizes with the server."

source "$HOME"/bin/script.sh

source ~/.go_profile
make
fslint -E -J -P -X ~/muzak
./bean-machine -m ~/muzak catalog
~/web/server-control music
