#!/bin/bash

# NOTE: This script needs be run from the root of the GD2 repository

# Find all Go source files in the repository, that are not vendored or generated
# and then run gofmt on them
GOFMT_FILE_LIST=$(gofmt -l . | grep -v vendor)
if [ -n "${GOFMT_FILE_LIST}" ]; then
  printf >&2 'gofmt failed for the following files:\n%s\n\nplease run "go fmt".\n' "${GOFMT_FILE_LIST}"
  exit 1
fi
