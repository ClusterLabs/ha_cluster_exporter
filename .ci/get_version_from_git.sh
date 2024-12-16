#!/bin/sh
TAG=$(git tag | grep -E "[0-9]\.[0-9]\.[0-9]" | sort -rn | head -n1)

if [ -z "${TAG}" ]; then
	echo "Could not find any tag" 1>&2
	exit 1
else
	COMMITS_SINCE_TAG=$(git rev-list "${TAG}".. --count)
	if [ "${COMMITS_SINCE_TAG}" -gt 0 ]; then
		COMMIT_INFO=$(git show -s --format=%ct.%h HEAD)
		SUFFIX="+git.${COMMITS_SINCE_TAG}.${COMMIT_INFO}"
	fi
fi

echo "${TAG}${SUFFIX}"
