#!/bin/bash
# 
# Copyright IBM Corp, SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

echo "Checking Go version"

function isGoVersionValid {
    GO_MAJOR_VERSION=`go version |awk '{print $3}' | awk -F "." '{print substr($1,3)}'`
    GO_MINOR_VERSION=`go version |awk '{print $3}' | awk -F "." '{print $2}'`
    GO_RELEASE_NO=`go version |awk '{print $3}' | awk -F "." '{print $3}'`

    if [ $GO_MAJOR_VERSION -ne 1 ]; then
        return 1
    fi

    if [ $GO_MINOR_VERSION -ne 10 ]; then
        return 1
    fi

    if [ ! -z "$GO_RELEASE_NO" ] && [ $GO_RELEASE_NO -gt 3 ]; then
        return 1
    fi

    return 0
}

if ! isGoVersionValid; then
    echo "You should install go 1.10.X (X <= 3) to build and run hyperledger fabric sdk/tools"
    exit 1
fi
