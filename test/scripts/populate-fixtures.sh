#!/bin/bash
#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#
# This script populates the fixtures folder.

set -e

GO_DEP_CMD="${GO_DEP_CMD:-dep}"
FABRIC_SDKGO_CODELEVEL_TAG="${FABRIC_SDKGO_CODELEVEL_TAG:-stable}"
FABRIC_CRYPTOCONFIG_VERSION="${FABRIC_CRYPTOCONFIG_VERSION:-v1}"
FABRIC_FIXTURE_VERSION="${FABRIC_FIXTURE_VERSION:-v1.2}"
LASTRUN_INFO_FILENAME="populate-fixtures-${FABRIC_SDKGO_CODELEVEL_TAG}.txt"
FIXTURES_CHANNEL_TREE_FILENAME="fixtures-channel-tree-${FABRIC_FIXTURE_VERSION}.txt"
FIXTURES_CRYPTO_TREE_FILENAME="fixtures-crypto-tree-${FABRIC_CRYPTOCONFIG_VERSION}.txt"
SCRIPT_REVISION=$(git log -1 --pretty=format:"%h" test/scripts/populate-fixtures.sh)
DATE=$(date +"%m-%d-%Y")

CACHE_PATH=""
function setCachePath {
    declare envOS=$(uname -s)
    declare pkgDir="fabric-sdk-go"

    if [ ${envOS} = 'Darwin' ]; then
        CACHE_PATH="${HOME}/Library/Caches/${pkgDir}"
    else
        CACHE_PATH="${HOME}/.cache/${pkgDir}"
    fi
}

# recordCacheResult writes the date and revision of successful script runs, to preempt unnecessary installs.
function recordCacheResult {
    declare FIXTURES_TREE_CHANNEL=$(ls -R test/fixtures/fabric/${FABRIC_FIXTURE_VERSION}/channel)
    declare FIXTURES_TREE_CRYPTO=$(ls -R test/fixtures/fabric/${FABRIC_CRYPTOCONFIG_VERSION}/crypto-config)

    mkdir -p ${CACHE_PATH}
    echo ${SCRIPT_REVISION} ${DATE} > "${CACHE_PATH}/${LASTRUN_INFO_FILENAME}"
    echo "${FIXTURES_TREE_CHANNEL}" > "${CACHE_PATH}/${FIXTURES_CHANNEL_TREE_FILENAME}"
    echo "${FIXTURES_TREE_CRYPTO}" > "${CACHE_PATH}/${FIXTURES_CRYPTO_TREE_FILENAME}"
}

function isScriptCurrent {
    declare filesModified=$(git diff --name-only --diff-filter=ACMRTUXBD HEAD | tr '\n' ' ' | xargs)
    if [[ "${filesModified}" =~ ( |^)(test/scripts/populate-fixtures.sh)( |$) ]]; then
        echo "Fixtures script modified - will need to repopulate fixtures"
        return 1
    fi
}

function isFixturesCurrent {
    if [ ! -d "test/fixtures/fabric/${FABRIC_FIXTURE_VERSION}/channel" ]; then
        echo "Channel directory does not exist - will need to populate fixture"
        return 1
    fi

    if [ ! -d "test/fixtures/fabric/${FABRIC_CRYPTOCONFIG_VERSION}/crypto-config" ]; then
        echo "Crypto config directory does not exist - will need to populate fixture"
        return 1
    fi

    if [ ! -f "${CACHE_PATH}/${FIXTURES_CHANNEL_TREE_FILENAME}" ]; then
        echo "Fixtures channel cache doesn't exist - populating fixtures"
        return 1
    fi

    if [ ! -f "${CACHE_PATH}/${FIXTURES_CRYPTO_TREE_FILENAME}" ]; then
        echo "Fixtures crypto cache doesn't exist - populating fixtures"
        return 1
    fi

    declare FIXTURES_TREE_CHANNEL=$(ls -R test/fixtures/fabric/${FABRIC_FIXTURE_VERSION}/channel)
    declare FIXTURES_TREE_CRYPTO=$(ls -R test/fixtures/fabric/${FABRIC_CRYPTOCONFIG_VERSION}/crypto-config)
    declare cachedChannelFixturesTree=$(< "${CACHE_PATH}/${FIXTURES_CHANNEL_TREE_FILENAME}")
    declare cachedCryptoFixturesTree=$(< "${CACHE_PATH}/${FIXTURES_CRYPTO_TREE_FILENAME}")
    if [ "${FIXTURES_TREE_CHANNEL}" != "${cachedChannelFixturesTree}" ]; then
        echo "Fixtures channel directory modified - will need to repopulate fixtures"
        return 1
    fi
    if [ "${FIXTURES_TREE_CRYPTO}" != "${cachedCryptoFixturesTree}" ]; then
        echo "Fixtures crypto directory modified - will need to repopulate fixtures"
        return 1
    fi
}

function isLastPopulateCurrent {

    if [ -f "${CACHE_PATH}/${LASTRUN_INFO_FILENAME}" ]; then
        declare -a lastScriptUsage=($(< "${CACHE_PATH}/${LASTRUN_INFO_FILENAME}"))
        echo "Populate script last ran ${lastScriptUsage[1]} on revision ${lastScriptUsage[0]}"

        if [ "${lastScriptUsage[0]}" = "${SCRIPT_REVISION}" ] && [ "${lastScriptUsage[1]}" = "${DATE}" ]; then
            return 0
        fi
    fi

    return 1
}

function isPopulateCurrent {
    if ! isScriptCurrent || ! isFixturesCurrent || ! isLastPopulateCurrent; then
        return 1
    fi
}


function isForceMode {
    if [ "${BASH_ARGV[0]}" != "-f" ]; then
        return 1
    fi
}

function generateCryptoConfig {
    echo "Generating crypto config ..."
    make crypto-gen
}

function generateChannelConfig {
    echo "Generating channel config ..."
    if [ "${FABRIC_SDKGO_CODELEVEL_TAG}" = "stable" ]; then
        make channel-config-stable-gen
    elif [ "${FABRIC_SDKGO_CODELEVEL_TAG}" = "prev" ]; then
        make channel-config-prev-gen
    elif [ "${FABRIC_SDKGO_CODELEVEL_TAG}" = "prerelease" ]; then
        make channel-config-prerelease-gen
    elif [ "${FABRIC_SDKGO_CODELEVEL_TAG}" = "devstable" ]; then
        make channel-config-devstable-gen
    fi
}

setCachePath

if ! isPopulateCurrent || isForceMode; then
    generateCryptoConfig
    generateChannelConfig
else
    echo "No need to populate fixtures"
fi

recordCacheResult