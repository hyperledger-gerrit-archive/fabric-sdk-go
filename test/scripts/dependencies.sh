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
GOPATH="${GOPATH:-$HOME/go}"

# Automatically install go tools (particularly for CI)
if [ "$FABRIC_SDKGO_DEPEND_INSTALL" = "true" ]; then
    echo "Installing dependencies ..."
    TMP=`mktemp -d 2>/dev/null || mktemp -d -t 'mytmpdir'`
    GOPATH=$TMP $GO_CMD get -u github.com/axw/gocov/...
    GOPATH=$TMP $GO_CMD get -u github.com/AlekSi/gocov-xml
    GOPATH=$TMP $GO_CMD get -u github.com/client9/misspell/cmd/misspell
    GOPATH=$TMP $GO_CMD get -u github.com/golang/lint/golint
    GOPATH=$TMP $GO_CMD get -u golang.org/x/tools/cmd/goimports
    GOPATH=$TMP $GO_CMD get -u github.com/golang/mock/mockgen
    GOPATH=$TMP $GO_CMD get -u github.com/alecthomas/gometalinter
    mkdir -p $GOPATH/bin
    cp $TMP/bin/* $GOPATH/bin
    gometalinter --install
fi

# Install specific version of go dep (particularly for CI)
if [ "$FABRIC_SDKGO_DEPEND_INSTALL" = "true" ] && [ -n "$GO_DEP_COMMIT" ]; then
    echo "Installing dep@$GO_DEP_COMMIT to $GOPATH/bin ..."
    TMP=`mktemp -d 2>/dev/null || mktemp -d -t 'mytmpdir'`

    GOPATH=$TMP go get -d github.com/golang/dep
    (cd $TMP/src/github.com/golang/dep && git reset --hard $GO_DEP_COMMIT)
    GOPATH=$TMP go install github.com/golang/dep/cmd/dep
    cp $TMP/bin/dep $GOPATH/bin

    rm -Rf $TMP
fi

# Check that Go tools are installed and help the user if they are missing
type gocov >/dev/null 2>&1 || { echo >& 2 "gocov is not installed (go get -u github.com/axw/gocov/...)"; ABORT=1; }
type gocov-xml >/dev/null 2>&1 || { echo >& 2 "gocov-xml is not installed (go get -u github.com/AlekSi/gocov-xml)"; ABORT=1; }
type misspell >/dev/null 2>&1 || { echo >& 2 "misspell is not installed (go get -u github.com/client9/misspell/cmd/misspell)"; ABORT=1; }
type golint >/dev/null 2>&1 || { echo >& 2 "golint is not installed (go get -u github.com/golang/lint/golint)"; ABORT=1; }
type goimports >/dev/null 2>&1 || { echo >& 2 "goimports is not installed (go get -u golang.org/x/tools/cmd/goimports)"; ABORT=1; }
type mockgen >/dev/null 2>&1 || { echo >& 2 "mockgen is not installed (go get -u github.com/golang/mock/mockgen)"; ABORT=1; }
type $GO_DEP_CMD >/dev/null 2>&1 || { echo >& 2 "dep is not installed (go get -u github.com/golang/dep/cmd/dep)"; ABORT=1; }
type gometalinter >/dev/null 2>&1 || { echo >& 2 "gometalinter is not installed (go get -u github.com/alecthomas/gometalinter)"; ABORT=1; }


if [ -n "$ABORT" ]; then
    echo "Missing dependency. Aborting. You can fix by installing the tool listed above or running make depend-install."
    exit 1
fi
