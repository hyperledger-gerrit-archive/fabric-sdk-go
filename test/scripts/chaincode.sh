#!/bin/bash
#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

set -e

GO_CMD="${GO_CMD:-go}"
GOPATH="${GOPATH:-$HOME/go}"

cp -R ${GOPATH}/src/chaincoded/vendor/* ${GOPATH}/src

echo "socat path" $(which socat)

#go install github.com/events_cc
go install github.com/example_cc
go install github.com/example_pvt_cc
go install chaincoded/cmd/chaincoded

PEERS=(
    peer0.org1.example.com:7052
    peer1.org1.example.com:7152
    peer0.org2.example.com:8052
    peer1.org2.example.com:9052
)

chaincoded ":2375" ${PEERS[@]}