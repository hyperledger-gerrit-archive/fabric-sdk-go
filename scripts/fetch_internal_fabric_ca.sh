#!/bin/bash
#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

# This script fetches code used in the SDK originating from other Hyperledger Fabric projects
# These files are checked into internal paths.
# Note: This script must be adjusted as upstream makes adjustments

INTERNAL_PATH="internal/fabric-ca/"
REMOTE_URL="https://raw.githubusercontent.com/hyperledger/fabric-ca/$FABRIC_CA_COMMIT"

IMPORT_FABRIC_SUBST='s/github.com\/hyperledger\/fabric/github.com\/hyperledger\/fabric-sdk-go\/internal\/fabric/g'
IMPORT_FABRICCA_SUBST='s/github.com\/hyperledger\/fabric-ca/github.com\/hyperledger\/fabric-sdk-go\/internal\/fabric-ca/g'

rm -Rf $INTERNAL_PATH
mkdir -p $INTERNAL_PATH

declare -a PKGS=(
    "api"
    "lib"
    "lib/tls"
    "lib/tcert"
    "lib/spi"
    "util"
)

# TODO: selective removal of files
declare -a FILES=(
    "api/client.go"
    "api/net.go"

    "lib/client.go"
    "lib/identity.go"
    "lib/signer.go"
    "lib/clientconfig.go"
    "lib/util.go"

    "lib/tls/tls.go"

    "lib/tcert/api.go"
    "lib/tcert/util.go"
    "lib/tcert/tcert.go"
    "lib/tcert/keytree.go"

    "lib/spi/affiliation.go"
    "lib/spi/userregistry.go"

    "util/util.go"
    "util/args.go"
    "util/csp.go"
    "util/struct.go"
    "util/flag.go"
)

for i in "${PKGS[@]}"
do
    mkdir -p $INTERNAL_PATH/${i}
done

for i in "${FILES[@]}"
do
    # Alt: clone local copy and copy individual files?
    curl -o $INTERNAL_PATH/${i} $REMOTE_URL/${i}

    # Apply global patching of import paths
    sed -i '' -e $IMPORT_FABRIC_SUBST $INTERNAL_PATH/${i}
    sed -i '' -e $IMPORT_FABRICCA_SUBST $INTERNAL_PATH/${i}
done

# Apply targeted patches