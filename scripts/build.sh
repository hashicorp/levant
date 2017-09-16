#!/usr/bin/env bash
#
# This script builds the application from source for multiple platforms.
set -e

export CGO_ENABLED=0

# Determine the Arch/OS combos we're building for
echo "==> Determining Arch/OS Info..."
XC_ARCH=${XC_ARCH:-"386 amd64 arm"}
XC_OS=${XC_OS:-"darwin freebsd linux"}

# Get Git Commit information
echo "==> Determining Git Info..."
GIT_COMMIT=$(git rev-parse HEAD)
GIT_DIRTY="$(test -n "$(git status --porcelain)" && echo "+CHANGES" || true)"

# LDFlags for Runtime Variables
LDFLAGS="-s -w"
LDFLAGS="$LDFLAGS -X main.Name=${NAME}"
LDFLAGS="$LDFLAGS -X main.Version=${VERSION}"
LDFLAGS="$LDFLAGS -X main.GitCommit=${GIT_COMMIT}${GIT_DIRTY}"

# Delete the old dir
echo "==> Removing old directory..."
rm -rf pkg/*

# Build!
echo "==> Building..."
"`which gox`" \
    -os="${XC_OS}" \
    -arch="${XC_ARCH}" \
    -osarch="!darwin/arm" \
    -ldflags "-X ${GIT_IMPORT}.GitCommit='${GIT_COMMIT}${GIT_DIRTY}' -X ${GIT_IMPORT}.GitDescribe='${GIT_DESCRIBE}'" \
    -output "pkg/{{.OS}}-{{.Arch}}-levant" \
    -tags="${BUILD_TAGS}" \
    .
