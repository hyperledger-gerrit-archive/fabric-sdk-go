#!/bin/bash
#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#
# Environment variables that affect this script:
# GO_TESTFLAGS: Flags are added to the go test command.
# GO_LDFLAGS: Flags are added to the go test command (example: -s).
# TEST_CHANGED_ONLY: Boolean on whether to only run tests on changed packages.
# TEST_RACE_CONDITIONS: Boolean on whether to test for race conditions.
# FABRIC_SDKGO_CODELEVEL_TAG: Go tag that represents the fabric code target
# FABRIC_SDKGO_CODELEVEL_VER: Version that represents the fabric code target (primarily for fixture lookup)
# FABRIC_CRYPTOCONFIG_VERSION: Version of cryptoconfig fixture to use

set -e

GO_CMD="${GO_CMD:-go}"
FABRIC_SDKGO_CODELEVEL_TAG="${FABRIC_SDKGO_CODELEVEL_TAG:-devstable}"
FABRIC_CRYPTOCONFIG_VERSION="${FABRIC_CRYPTOCONFIG_VERSION:-v1}"
TEST_CHANGED_ONLY="${TEST_CHANGED_ONLY:-false}"
TEST_RACE_CONDITIONS="${TEST_RACE_CONDITIONS:-true}"
SCRIPT_DIR="$(dirname "$0")"

REPO="github.com/hyperledger/fabric-sdk-go"

source ${SCRIPT_DIR}/lib/find_packages.sh


# Find all packages that should be tested.
declare -a PKG_SRC=(
"./pkg"
"./test"
)
findPackages

# Reduce unit tests to changed packages.
if [ "$TEST_CHANGED_ONLY" = true ]; then
    findChangedPackages
    filterExcludedPackages
    appendDepPackages
    PKGS=(${DEP_PKGS})
fi

RACEFLAG=""
if [ "$TEST_RACE_CONDITIONS" = true ]; then
    ARCH=$(uname -m)

    if [ "$ARCH" == "x86_64" ]; then
        echo "Enabling race condition flag for upcoming unit test run"
        RACEFLAG="-race"
    else
        echo "Race condition flag not supported on $ARCH for upcoming unit test run"
    fi
fi

if [ ${#PKGS[@]} -eq 0 ]; then
    echo "Skipping unit tests since no packages were changed"
    exit 0
fi

echo "Running unit tests..."
echo "Testing with code level $FABRIC_SDKGO_CODELEVEL_TAG (Fabric ${FABRIC_SDKGO_CODELEVEL_VER}) ..."
GO_TAGS="$GO_TAGS $FABRIC_SDKGO_CODELEVEL_TAG"

GO_LDFLAGS="$GO_LDFLAGS -X github.com/hyperledger/fabric-sdk-go/test/metadata.ChannelConfigPath=test/fixtures/fabric/${FABRIC_SDKGO_CODELEVEL_VER}/channel -X github.com/hyperledger/fabric-sdk-go/test/metadata.CryptoConfigPath=test/fixtures/fabric/${FABRIC_CRYPTOCONFIG_VERSION}/crypto-config"
$GO_CMD test $RACEFLAG -cover -tags "testing $GO_TAGS" $GO_TESTFLAGS -ldflags="$GO_LDFLAGS" ${PKGS[@]} -p 1 -timeout=40m

echo "Unit tests finished successfully"