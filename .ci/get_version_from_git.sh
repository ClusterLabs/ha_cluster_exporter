#!/bin/sh
TAG=$(git describe --tags --abbrev=0 2>/dev/null)

if [ -n "${TAG}" ]; then
  COMMITS_SINCE_TAG=$(git rev-list ${TAG}.. --count)
  if [ "${COMMITS_SINCE_TAG}" -gt 0 ]; then
    TAG="${TAG}.dev${COMMITS_SINCE_TAG}"
  fi
else
  TAG="0"
fi

SUFFIX=$(git show -s --format=%ct.%h HEAD)

echo "${TAG}+git.${SUFFIX}"
