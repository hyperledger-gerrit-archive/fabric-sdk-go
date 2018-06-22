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
GOMETALINT_CMD="gometalinter"
SCRIPT_DIR="$(dirname "$0")"

REPO="github.com/hyperledger/fabric-sdk-go"

source ${SCRIPT_DIR}/lib/find_packages.sh

findPackages

# Reduce Linter checks to changed packages.
if [ "$LINT_CHANGED_ONLY" = true ]; then
    findChangedPackages
    filterExcludedPackages
    appendDepPackages
fi

packagesToDirs

if [ ${#DIRS[@]} -eq 0 ]; then
    echo "Skipping linter since no packages were changed"
    exit 0
fi

if [ "$LINT_CHANGED_ONLY" = true ]; then
    echo "Changed directories to lint: ${DIRS[@]}"
fi

echo "Running metalinters..."
$GOMETALINT_CMD --config=./gometalinter.json "${DIRS[@]}"
echo "Metalinters finished successfully"
