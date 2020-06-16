#!/bin/sh
TAG=$(git describe --tags --abbrev=0 2>/dev/null)
SUFFIX=$(git show -s --format=%ct.%h HEAD)

if [ -n "${TAG}" ]; then
  COMMITS_SINCE_TAG=$(git rev-list ${TAG}.. --count)
  if [ "${COMMITS_SINCE_TAG}" -gt 0 ]; then
    SUFFIX="dev${COMMITS_SINCE_TAG}.${SUFFIX}"
  fi
else
  TAG="0"
fi

echo "${TAG}+git.${SUFFIX}"
