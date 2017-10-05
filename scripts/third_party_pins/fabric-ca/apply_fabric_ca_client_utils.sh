#!/bin/bash
#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

# This script pins the BCCSP package family from Hyperledger Fabric into the SDK
# These files are checked into internal paths.
# Note: This script must be adjusted as upstream makes adjustments

IMPORT_SUBSTS=($IMPORT_SUBSTS)

GOIMPORTS_CMD=goimports
GOFILTER_CMD="go run scripts/_go/cmd/gofilter/gofilter.go"

declare -a PKGS=(
    "api"
    "lib"
    "lib/tls"
    "lib/logbridge"
    "util"
)

declare -a FILES=(
    "api/client.go"
    "api/net.go"

    "lib/client.go"
    "lib/identity.go"
    "lib/signer.go"
    "lib/clientconfig.go"
    "lib/util.go"
    "lib/serverstruct.go"

    "lib/tls/tls.go"

    "lib/logbridge/logbridge.go"
    "lib/logbridge/syslogwriter.go"

    "util/util.go"
    "util/csp.go"
    "util/args.go"
)

echo 'Removing current upstream project from working directory ...'
rm -Rf "${INTERNAL_PATH}"
mkdir -p "${INTERNAL_PATH}"

# Create directory structure for packages
for i in "${PKGS[@]}"
do
    mkdir -p $INTERNAL_PATH/${i}
done

# Apply fine-grained patching
gofilter() {
    echo "Filtering: ${FILTER_FILENAME}"
    cp ${TMP_PROJECT_PATH}/${FILTER_FILENAME} ${TMP_PROJECT_PATH}/${FILTER_FILENAME}.bak
    $GOFILTER_CMD -filename "${TMP_PROJECT_PATH}/${FILTER_FILENAME}.bak" \
        -fn "$FILTER_FN" \
        > "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
} 

echo "Filtering Go sources for allowed functions ..."
FILTER_FILENAME="lib/client.go"
FILTER_FN=Enroll,GenCSR,SendReq,Init,newPost,newEnrollmentResponse,newCertificateRequest,getURL,StoreMyIdentity,NormalizeURL,initHTTPClient,net2LocalServerInfo,NewIdentity
gofilter

FILTER_FILENAME="lib/util.go"
FILTER_FN=GetCertID,BytesToX509Cert
gofilter

FILTER_FILENAME="util/args.go"
FILTER_FN=GetServerPort,GetCommandLineOptValue
gofilter

FILTER_FILENAME="util/csp.go"
FILTER_FN=GetDefaultBCCSP,InitBCCSP,ConfigureBCCSP,GetBCCSP,makeFileNamesAbsolute,getBCCSPKeyOpts,ImportBCCSPKeyFromPEM,LoadX509KeyPair,GetSignerFromCert,BCCSPKeyRequestGenerate,GetSignerFromCertFile
gofilter
sed -i '' -e '/_.\"time\"/d' $FILTER_FILENAME
sed -i '' -e '/\"github.com\/cloudflare\/cfssl\/cli\"/d' $FILTER_FILENAME
sed -i '' -e '/\"github.com\/cloudflare\/cfssl\/ocsp\"/d' $FILTER_FILENAME

FILTER_FILENAME="util/util.go"
FILTER_FN=ReadFile,WriteFile,GenECDSAToken,B64Encode,B64Decode,HTTPRequestToString,HTTPResponseToString,GetX509CertificateFromPEM,GetEnrollmentIDFromPEM,GetSerialAsHex,MakeFileAbs,GetEnrollmentIDFromX509Certificate,Marshal,Unmarshal,StructToString,FileExists,GetServerPort,LoadX509KeyPair,CreateToken
gofilter

# Apply patching
echo "Patching import paths on upstream project ..."
WORKING_DIR=$TMP_PROJECT_PATH FILES="${FILES[@]}" IMPORT_SUBSTS="${IMPORT_SUBSTS[@]}" scripts/third_party_pins/common/apply_import_patching.sh

echo "Inserting modification notice ..."
WORKING_DIR=$TMP_PROJECT_PATH FILES="${FILES[@]}" scripts/third_party_pins/common/apply_header_notice.sh

# Copy patched project into internal paths
echo "Copying patched upstream project into working directory ..."
for i in "${FILES[@]}"
do
    TARGET_PATH=`dirname $INTERNAL_PATH/${i}`
    cp $TMP_PROJECT_PATH/${i} $TARGET_PATH
done
