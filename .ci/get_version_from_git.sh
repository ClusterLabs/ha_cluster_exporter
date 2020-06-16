#!/bin/sh
TAG=$(git describe --tags --abbrev=0 2>/dev/null)

if [ -z "${TAG}" ]; then
  TAG="0.0dev"
elif [ "$(git rev-list ${TAG}.. --count)" -gt 0 ]; then
  TAG="${TAG}dev"
fi

SUFFIX=$(git show -s --format=%ct.%h HEAD)

echo "${TAG}+git.${SUFFIX}"
