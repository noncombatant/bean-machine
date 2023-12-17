#!/usr/bin/env bash

help="Updates the catalog and synchronizes with the server."

source "$HOME/bin/script.sh"

source ~/.go_profile
make
fslint -E -J -P /Users/Shared/music
./bean-machine -m /Users/Shared/music catalog
backup file
ssh file
