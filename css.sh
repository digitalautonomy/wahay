#!/usr/bin/env bash

WATCH_ARG=$1

if [ ! "$(which sass 2> /dev/null)" ]; then
  echo sass needs to be installed to generate the css.
  exit 1
fi

if [ ! -d ./gui/styles ]; then
  mkdir -p ./gui/styles
fi

sass $WATCH_ARG ./sass/gui.scss:./gui/styles/gui.css
