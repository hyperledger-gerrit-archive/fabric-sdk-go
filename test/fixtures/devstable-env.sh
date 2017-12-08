#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

# This file contains environment overrides to enable testing
# against the latest dev-stable target.

# Uses nexus registry (see https://nexus3.hyperledger.org/#browse/browse/components:docker.snapshot)
export ARCH=""
export ARCH_SEP=""

export FABRIC_FIXTURE_VERSION="v1.1"

export FABRIC_CA_FIXTURE_TAG="DEV_STABLE"
export FABRIC_ORDERER_FIXTURE_TAG="DEV_STABLE"
export FABRIC_PEER_FIXTURE_TAG="DEV_STABLE"
export FABRIC_COUCHDB_FIXTURE_TAG="DEV_STABLE"
export FABRIC_BUILDER_FIXTURE_TAG="DEV_STABLE"

# override configuration that loads crypto-config
FABRIC_SDK_CLIENT_CRYPTOCONFIG_PATH='$GOPATH'"/src/github.com/hyperledger/fabric-sdk-go/test/fixtures/fabric-${FABRIC_FIXTURE_VERSION}/config/crypto-config"
FABRIC_SDK_CLIENT_ORDERERS_TLSCACERTS_PATH='$GOPATH'"/src/github.com/hyperledger/fabric-sdk-go/test/fixtures/fabric-${FABRIC_FIXTURE_VERSION}/config/crypto-config/ordererOrganizations/example.com/tlsca/tlsca.example.com-cert.pem"
FABRIC_SDK_CLIENT_PEERS_PEER0_ORG1_EXAMPLE_COM_TLSCACERTS_PATH='$GOPATH'"/src/github.com/hyperledger/fabric-sdk-go/test/fixtures/fabric-${FABRIC_FIXTURE_VERSION}/config/crypto-config/peerOrganizations/org1.example.com/tlsca/tlsca.org1.example.com-cert.pem"
FABRIC_SDK_CLIENT_PEERS_PEER0_ORG2_EXAMPLE_COM_TLSCACERTS_PATH='$GOPATH'"/src/github.com/hyperledger/fabric-sdk-go/test/fixtures/fabric-${FABRIC_FIXTURE_VERSION}/config/crypto-config/peerOrganizations/org2.example.com/tlsca/tlsca.org2.example.com-cert.pem"
