#!/usr/bin/env bash

version_file=$1
version=$(awk '$1 == "Version" && $2 == "=" { gsub(/"/, "", $3); print $3 }' < "${version_file}")
prerelease=$(awk '$1 == "Prerelease" && $2 == "=" { gsub(/"/, "", $3); print $3 }' < "${version_file}")
version=${version#v}

if [ -n "$prerelease" ]; then
    echo "${version}-${prerelease}"
else
    echo "${version}"
fi
