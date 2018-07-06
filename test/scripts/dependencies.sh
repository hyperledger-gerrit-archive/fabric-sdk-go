#!/bin/bash
#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#
# This script installs dependencies for testing tools
# Environment variables that affect this script:
# FABRIC_SDKGO_DEPEND_INSTALL: Installs dependencies
# GO_DEP_COMMIT: Tag or commit level of the go dep tool to install (if FABRIC_SDKGO_DEPEND_INSTALL=true)

GO_CMD="${GO_CMD:-go}"
GO_DEP_CMD="${GO_DEP_CMD:-dep}"
GO_DEP_REPO="github.com/golang/dep"
GO_METALINTER_CMD="${GO_METALINTER_CMD:-gometalinter.v2}"
GO_METALINTER_REPO="gopkg.in/alecthomas/gometalinter.v2"
GOPATH="${GOPATH:-${HOME}/go}"
CACHE_PATH="${HOME}/.cache/fabric-sdk-go"

DEPEND_SCRIPT_REVISION=$(git log -1 --pretty=format:"%h" test/scripts/dependencies.sh)
DATE=$(date +"%m-%d-%Y")

LASTRUN_INFO_PATH="${CACHE_PATH}/dependencies.txt"

function installGoDep {
    declare repo=$1
    declare revision=$2

    installGoPkg "${repo}" "${revision}" "/cmd/dep" "dep"
}

function installGoMetalinter {
    echo "Installing ${GO_METALINTER_PKG} to $GOPATH/bin ..."
    declare GO_METALINTER_PKG="github.com/alecthomas/gometalinter"

    GOPATH=${BUILD_TMP} ${GO_CMD} get -u ${GO_METALINTER_REPO}

    mkdir -p ${GOPATH}/bin
    cp ${BUILD_TMP}/bin/* ${GOPATH}/bin
    rm -Rf ${GOPATH}/src/${GO_METALINTER_PKG}
    mkdir -p ${GOPATH}/src/${GO_METALINTER_PKG}
    cp -Rf ${BUILD_TMP}/src/${GO_METALINTER_REPO}/* ${GOPATH}/src/${GO_METALINTER_PKG}
    ${GO_METALINTER_CMD} --install --force
}

function installGoGas {
    declare repo="github.com/GoASTScanner/gas"
    declare revision="4ae8c95"

    GOPATH=${BUILD_TMP} ${GO_CMD} get -u github.com/kisielk/gotool
    GOPATH=${BUILD_TMP} ${GO_CMD} get -u github.com/nbutton23/zxcvbn-go
    GOPATH=${BUILD_TMP} ${GO_CMD} get -u github.com/ryanuber/go-glob
    GOPATH=${BUILD_TMP} ${GO_CMD} get -u gopkg.in/yaml.v2

    installGoPkg "${repo}" "${revision}" "/cmd/gas/..." "gas"
}

function installGoPkg {
    declare repo=$1
    declare revision=$2
    declare pkgPath=$3
    shift 3
    declare -a cmds=$@

    echo "Installing ${repo}@${revision} to $GOPATH/bin ..."

    GOPATH=${BUILD_TMP} go get -d ${repo}
    (cd ${BUILD_TMP}/src/${repo} && git reset --hard ${revision})
    GOPATH=${BUILD_TMP} go install -i ${repo}/${pkgPath}

    for cmd in ${cmds[@]}
    do
        echo "Copying ${cmd} to ${GOPATH}/bin"
        cp ${BUILD_TMP}/bin/${cmd} ${GOPATH}/bin
    done
}

function isLastInstallCurrent {
    if [ -f ${LASTRUN_INFO_PATH} ]; then
        declare -a lastScriptUsage=($(cat < ${LASTRUN_INFO_PATH}))
        echo "Dependency script last ran ${lastScriptUsage[1]} on revision ${lastScriptUsage[0]}"

        if [ "${lastScriptUsage[0]}" = "${DEPEND_SCRIPT_REVISION}" ] && [ "${lastScriptUsage[1]}" = "${DATE}" ]; then
            return 0
        fi
    fi

    return 1
}

# isDependenciesInstalled checks that Go tools are installed and help the user if they are missing
function isDependenciesInstalled {
    declare printMsgs=$1
    declare -a msgs=()

    # Check that Go tools are installed and help the user if they are missing
    type gocov >/dev/null 2>&1 || msgs+=("gocov is not installed (go get -u github.com/axw/gocov/...)")
    type gocov-xml >/dev/null 2>&1 || msgs+=("gocov-xml is not installed (go get -u github.com/AlekSi/gocov-xml)")
    type mockgen >/dev/null 2>&1 || msgs+=("mockgen is not installed (go get -u github.com/golang/mock/mockgen)")
    type ${GO_DEP_CMD} >/dev/null 2>&1 || msgs+=("dep is not installed (go get -u github.com/golang/dep/cmd/dep)")
    type ${GO_METALINTER_CMD} >/dev/null 2>&1 || msgs+=("gometalinter is not installed (go get -u ${GO_METALINTER_PKG})")

    if [ ${#msgs[@]} -gt 0 ]; then
        if [ ${printMsgs} = true ]; then
            echo >& 2 $(echo ${msgs[@]} | tr ' ' '\n')
        fi

        return 1
    fi
}

function installDependencies {
    echo "Installing dependencies ..."
    BUILD_TMP=`mktemp -d 2>/dev/null || mktemp -d -t 'fabricsdkgo'`
    GOPATH=${BUILD_TMP} ${GO_CMD} get -u github.com/axw/gocov/...
    GOPATH=${BUILD_TMP} ${GO_CMD} get -u github.com/AlekSi/gocov-xml
    GOPATH=${BUILD_TMP} ${GO_CMD} get -u github.com/golang/mock/mockgen

    installGoMetalinter

    # gas in gometalinter is out of date.
    installGoGas

    # Install specific version of go dep (particularly for CI)
    if [ -n "${GO_DEP_COMMIT}" ]; then
        installGoDep ${GO_DEP_REPO} ${GO_DEP_COMMIT}
    fi

    rm -Rf ${BUILD_TMP}
}

# Automatically install go tools (particularly for CI)
if [ "${FABRIC_SDKGO_DEPEND_INSTALL}" = "true" ]; then
    if ! isLastInstallCurrent || ! isDependenciesInstalled false; then
        installDependencies
    else
        echo "No need to install dependencies"
    fi
fi

if ! isDependenciesInstalled true; then
    echo "Missing dependency. Aborting. You can fix by installing the tool listed above or running make depend-install."
    exit 1
fi

# Record date and revision of successful script runs, to preempt unnecessary installs.
if [ "${FABRIC_SDKGO_DEPEND_INSTALL}" = "true" ]; then
    mkdir -p ${CACHE_PATH}
    echo ${DEPEND_SCRIPT_REVISION} ${DATE} > ${LASTRUN_INFO_PATH}
fi
