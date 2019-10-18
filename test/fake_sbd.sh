#!/usr/bin/env bash

if [[ "$2" == "/dev/vdc" ]]; then
  exit 0
fi

exit 1
