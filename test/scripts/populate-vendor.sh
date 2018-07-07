#!/bin/bash
#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#
# This script populates the vendor folder.

set -e

GO_DEP_CMD="${GO_DEP_CMD:-dep}"
LASTRUN_INFO_FILENAME="populate-vendor.txt"
VENDOR_TREE_FILENAME="vendor-tree.json"
SCRIPT_REVISION=$(git log -1 --pretty=format:"%h" test/scripts/populate-vendor.sh)
LOCK_REVISION=$(git log -1 --pretty=format:"%h" Gopkg.lock)
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
    mkdir -p ${CACHE_PATH}
    echo ${SCRIPT_REVISION} ${LOCK_REVISION} ${DATE} > "${CACHE_PATH}/${LASTRUN_INFO_FILENAME}"
    tree -J vendor > "${CACHE_PATH}/${VENDOR_TREE_FILENAME}"
}

function isLastPopulateCurrent {
    declare filesModified=$(git status | grep -E 'test/scripts/populate-vendor.sh|Gopkg.lock')

    if [ ! -z "${filesModified}" ]; then
        echo "Vendor script or Gopkg.lock modified - repopulating vendor"
        #return 1
    fi

    if [ -f "${CACHE_PATH}/${LASTRUN_INFO_FILENAME}" ]; then
        declare -a lastScriptUsage=($(cat < "${CACHE_PATH}/${LASTRUN_INFO_FILENAME}"))
        echo "Dependency script last ran ${lastScriptUsage[2]} on revision ${lastScriptUsage[0]} with Gopkg.lock revision ${lastScriptUsage[1]}"

        if [ "${lastScriptUsage[0]}" = "${SCRIPT_REVISION}" ] && [ "${lastScriptUsage[1]}" = "${LOCK_REVISION}" ] && [ "${lastScriptUsage[2]}" = "${DATE}" ]; then
            return 0
        fi
    fi

    return 1
}

function isForceMode {
    if [ "${BASH_ARGV[0]}" = "-f" ]; then
        return 0
    fi

    return 1
}

function populateVendor {
    echo "Populating vendor ..."
	${GO_DEP_CMD} ensure -vendor-only
}

setCachePath

if ! isLastPopulateCurrent || isForceMode; then
    populateVendor
else
    echo "No need to populate vendor"
fi

recordCacheResult