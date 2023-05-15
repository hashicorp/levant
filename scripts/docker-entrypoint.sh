#!/usr/bin/env sh
# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0


set -e

if [ "$1" = 'levant' ]; then
  shift
fi

exec levant "$@"
