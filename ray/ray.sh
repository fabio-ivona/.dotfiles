#!/bin/bash

LATEST_BINARY=$(ls -1  $HOME/.dotfiles/ray/Ray-*.*.AppImage | sort -r | head -n 1)

$LATEST_BINARY
