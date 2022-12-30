#!/usr/bin/env sh

set -e

if [ "$1" = 'levant' ]; then
  shift
fi

exec levant "$@"
