#!/usr/bin/env bash

help="Updates the catalog and synchronizes with the server."

source "$HOME"/bin/script.sh

source ~/.go_profile
make
fslint -E -J -P -X ~/d/muzak
./bean-machine -m ~/d/muzak catalog
~/d/web/server-control music
~/d/web/server-control ssh
