#!/bin/bash
#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#
# Environment variables that affect this script:
# GO_TESTFLAGS: Flags are added to the go test command.
# GO_LDFLAGS: Flags are added to the go test command (example: -s).
# FABRIC_SDKGO_CODELEVEL_TAG: Go tag that represents the fabric code target
# FABRIC_SDKGO_CODELEVEL_VER: Version that represents the fabric code target (primarily for fixture lookup)
# FABRIC_CRYPTOCONFIG_VERSION: Version of cryptoconfig fixture to use

set -e

GO_CMD="${GO_CMD:-go}"
FABRIC_SDKGO_CODELEVEL_TAG="${FABRIC_SDKGO_CODELEVEL_TAG:-devstable}"
FABRIC_CRYPTOCONFIG_VERSION="${FABRIC_CRYPTOCONFIG_VERSION:-v1}"
TEST_CHANGED_ONLY="${TEST_CHANGED_ONLY:-false}"

REPO="github.com/hyperledger/fabric-sdk-go"

function findPackages {
    # Packages to include in test run
    PKG_LIST=`$GO_CMD list $REPO... 2> /dev/null | \
          grep -v ^$REPO$ | \
          grep -v ^$REPO/api/ | grep -v ^$REPO/.*/api[^/]*$ | \
          grep -v ^$REPO/.*/mocks$ | \
          grep -v ^$REPO/internal/github.com/ | grep -v ^$REPO/third_party/ | \
          grep -v ^$REPO/pkg/core/cryptosuite/bccsp/pkcs11 | grep -v ^$REPO/pkg/core/cryptosuite/bccsp/multisuite | \
          grep -v ^$REPO/vendor/ | grep -v ^$REPO/test/`

    PKGS=()
    while read -r line; do
      PKGS+=("$line")
    done <<< "$PKG_LIST"
}

function findChangedPackages {
    # Determine which directories have changes.
    CHANGED=$(git diff --name-only --diff-filter=ACMRTUXB HEAD)

    if [[ "$CHANGED" != "" ]]; then
        CHANGED+=$'\n'
    fi

    LAST_COMMITS=($(git log -2 --pretty=format:"%h"))
    CHANGED+=$(git diff-tree --no-commit-id --name-only --diff-filter=ACMRTUXB -r ${LAST_COMMITS[1]} ${LAST_COMMITS[0]})

    CHANGED_PKGS=()
    while read -r line; do
        if [ "$line" != "" ]; then
            DIR=`dirname $line`
            if [ "$DIR" = "." ]; then
                CHANGED_PKGS+=("$REPO")
            else
                CHANGED_PKGS+=("$REPO/$DIR")
            fi
        fi
    done <<< "$CHANGED"
    CHANGED_PKGS=($(printf "%s\n" "${CHANGED_PKGS[@]}" | sort -u | tr '\n' ' '))
}

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

findPackages

# Reduce unit tests to changed packages.
if [ "$TEST_CHANGED_ONLY" = true ]; then
    findChangedPackages
    filterExcludedPackages
fi

echo "Running unit tests..."

RACEFLAG=""
ARCH=$(uname -m)

if [ "$ARCH" == "x86_64" ]; then
    RACEFLAG="-race"
fi

echo "Testing with code level $FABRIC_SDKGO_CODELEVEL_TAG (Fabric ${FABRIC_SDKGO_CODELEVEL_VER}) ..."
GO_TAGS="$GO_TAGS $FABRIC_SDKGO_CODELEVEL_TAG"

GO_LDFLAGS="$GO_LDFLAGS -X github.com/hyperledger/fabric-sdk-go/test/metadata.ChannelConfigPath=test/fixtures/fabric/${FABRIC_SDKGO_CODELEVEL_VER}/channel -X github.com/hyperledger/fabric-sdk-go/test/metadata.CryptoConfigPath=test/fixtures/fabric/${FABRIC_CRYPTOCONFIG_VERSION}/crypto-config"
$GO_CMD test $RACEFLAG -cover -tags "testing $GO_TAGS" $GO_TESTFLAGS -ldflags="$GO_LDFLAGS" ${PKGS[@]} -p 1 -timeout=40m

echo "Unit tests finished successfully"