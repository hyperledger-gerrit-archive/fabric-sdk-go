#!/bin/bash
#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#
# This script runs Go linting and vetting tools

set -e
LINT_CHANGED_ONLY="${LINT_CHANGED_ONLY:-false}"
GO_CMD="${GO_CMD:-go}"
GO_PROJECT="github.com/hyperledger/fabric-sdk-go"
GOMETALINT_CMD=gometalinter

# Find all packages that should be linted.
declare -a src=(
"./pkg"
"./test"
)

PKGS=()
for i in "${src[@]}"
do
   PKG_LIST=`$GO_CMD list $i/...`
   while read -r line; do
      PKGS+=("$line")
   done <<< "$PKG_LIST"
done

# Determine which directories have changes.
CHANGED=$(git diff --name-only --diff-filter=ACMRTUXB HEAD)
CHANGED_PKGS=()

while read -r line; do
    DIR=`dirname $line`
    CHANGED_PKGS+=("$GO_PROJECT/$DIR")
done <<< "$CHANGED"
CHANGED_PKGS=($(printf "%s\n" "${CHANGED_PKGS[@]}" | sort -u | tr '\n' ' '))

function filterExcludedPackages {
    FILTERED_PKGS=()

    for pkg in "${PKGS[@]}"
    do
        for i in "${CHANGED_PKGS[@]}"
        do
            if [ "$pkg" = "$i" ]; then
              FILTERED_PKGS+=("$pkg")
            fi
        done
    done

    PKGS=("${FILTERED_PKGS[@]}")
}

# Reduce Linter checks to changed packages.
if [ "$LINT_CHANGED_ONLY" = true ]; then
    filterExcludedPackages
fi

# Convert packages to directories
echo "Paths to lint:"
DIRS=()
for i in "${PKGS[@]}"
do
    dirname=${i#$GO_PROJECT/}
    DIRS+=($dirname)
    echo "  $dirname"
done

echo "Running metalinters..."
$GOMETALINT_CMD --config=./gometalinter.json "${DIRS[@]}"
echo "Metalinters finished successfully"
