#!/bin/bash
#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

CRYPTOGEN_CMD="${CRYPTOGEN_CMD:-cryptogen}"
FIXTURES_PATH="${FIXTURES_PATH:-/opt/gopath/src/github.com/hyperledger/fabric-sdk-go/test/fixtures}"
CONFIG_DIR="${CONFIG_DIR:-config}"

if [ -z "$FABRIC_VERSION_DIR" ]; then
  echo "FABRIC_VERSION_DIR is required"
  exit 1
fi

declare -a peerOrgs=(
    "org1.example.com"
    "org2.example.com"
)

declare -a ordererOrgs=(
    "example.com"
)

echo Clearing old crypto directory ...
rm -Rf ${FIXTURES_PATH}/${FABRIC_VERSION_DIR}/crypto-config

echo Running cryptogen ...
${CRYPTOGEN_CMD} generate --config=${FIXTURES_PATH}/${FABRIC_VERSION_DIR}/config/cryptogen.yaml --output=${FIXTURES_PATH}/${FABRIC_VERSION_DIR}/crypto-config

# Remove unneeded ca MSP
for org in ${peerOrgs[@]}; do
    rm -Rf ${FIXTURES_PATH}/${FABRIC_VERSION_DIR}/crypto-config/peerOrganizations/${org}/peers/ca.${org}/msp
done

echo Updating config ...
keyPath=$(ls ${FIXTURES_PATH}/${FABRIC_VERSION_DIR}/crypto-config/peerOrganizations/org1.example.com/ca/*_sk)
sed -i'' -e "s/ORG1CA1_FABRIC_CA_SERVER_CA_KEYFILE=\/etc\/hyperledger\/fabric-ca-server-config\/.*/ORG1CA1_FABRIC_CA_SERVER_CA_KEYFILE=\/etc\/hyperledger\/fabric-ca-server-config\/${keyPath##*/}/" "${FIXTURES_PATH}/dockerenv/.env"
keyPath=$(ls ${FIXTURES_PATH}/${FABRIC_VERSION_DIR}/crypto-config/peerOrganizations/org2.example.com/ca/*_sk)
sed -i'' -e "s/ORG2CA1_FABRIC_CA_SERVER_CA_KEYFILE=\/etc\/hyperledger\/fabric-ca-server-config\/.*/ORG2CA1_FABRIC_CA_SERVER_CA_KEYFILE=\/etc\/hyperledger\/fabric-ca-server-config\/${keyPath##*/}/" "${FIXTURES_PATH}/dockerenv/.env"
