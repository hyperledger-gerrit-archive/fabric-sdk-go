#!/bin/bash
#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

# This script fetches code used in the SDK originating from other Hyperledger Fabric projects
# These files are checked into internal paths.
# Note: This script must be adjusted as upstream makes adjustments

INTERNAL_PATH="internal/fabric/"
REMOTE_URL="https://raw.githubusercontent.com/hyperledger/fabric/$FABRIC_COMMIT"

IMPORT_SUBST='s/github.com\/hyperledger\/fabric/github.com\/hyperledger\/fabric-sdk-go\/internal\/fabric/g'

rm -Rf $INTERNAL_PATH
mkdir -p $INTERNAL_PATH

declare -a PKGS=(
    "bccsp"
    "bccsp/factory"
    "bccsp/pkcs11"
    "bccsp/signer"
    "bccsp/sw"
    "bccsp/utils"
    "protos/utils"
    "protos/common"
    "protos/ledger/rwset"
    "protos/ledger/rwset/kvrwset"
    "protos/msp"
    "protos/orderer"
    "protos/peer"

    "common/crypto"
    "common/errors"
    "common/flogging"
    "common/util"
    "common/metadata"
    "common/channelconfig"
    "common/cauthdsl"
    "common/configtx"
    "common/configtx/api"
    "common/policies"
    "common/ledger"
    "common/ledger/util"

    "core/comm"
    "core/config"
    "core/ledger"
    "core/ledger/kvledger/txmgmt/rwsetutil"
    "core/ledger/kvledger/txmgmt/version"
    "core/ledger/util"

    "events/consumer"

    "msp"
    "msp/cache"
    "msp/mgmt"
)

# TODO: selective removal of files
declare -a FILES=(
"bccsp/aesopts.go"
"bccsp/bccsp.go"
"bccsp/ecdsaopts.go"
"bccsp/hashopts.go"
"bccsp/keystore.go"
"bccsp/opts.go"
"bccsp/rsaopts.go"
"bccsp/rsaopts.go"
"bccsp/rsaopts.go"
"bccsp/factory/factory.go"
"bccsp/factory/nopkcs11.go"
"bccsp/factory/opts.go"
"bccsp/factory/pkcs11.go"
"bccsp/factory/pkcs11factory.go"
"bccsp/factory/swfactory.go"
"bccsp/pkcs11/conf.go"
"bccsp/pkcs11/ecdsa.go"
"bccsp/pkcs11/ecdsakey.go"
"bccsp/pkcs11/impl.go"
"bccsp/pkcs11/pkcs11.go"
"bccsp/signer/signer.go"
"bccsp/sw/aes.go"
"bccsp/sw/aeskey.go"
"bccsp/sw/conf.go"
"bccsp/sw/dummyks.go"
"bccsp/sw/ecdsa.go"
"bccsp/sw/ecdsakey.go"
"bccsp/sw/fileks.go"
"bccsp/sw/hash.go"
"bccsp/sw/impl.go"
"bccsp/sw/internals.go"
"bccsp/sw/keyderiv.go"
"bccsp/sw/keygen.go"
"bccsp/sw/keyimport.go"
"bccsp/sw/rsa.go"
"bccsp/sw/rsakey.go"
"bccsp/utils/errs.go"
"bccsp/utils/io.go"
"bccsp/utils/keys.go"
"bccsp/utils/slice.go"
"bccsp/utils/x509.go"

"common/crypto/random.go"
"common/crypto/signer.go"

"common/errors/codes.go"
"common/errors/errors.go"

"common/flogging/grpclogger.go"
"common/flogging/logging.go"

"common/util/utils.go"

"common/metadata/metadata.go"

"common/channelconfig/api.go"
"common/channelconfig/application.go"
"common/channelconfig/application_util.go"
"common/channelconfig/applicationorg.go"
"common/channelconfig/bundle.go"
"common/channelconfig/bundlesource.go"
"common/channelconfig/channel.go"
"common/channelconfig/channel_util.go"
"common/channelconfig/consortium.go"
"common/channelconfig/consortiums.go"
"common/channelconfig/logsanitychecks.go"
"common/channelconfig/msp.go"
"common/channelconfig/msp_util.go"
"common/channelconfig/orderer.go"
"common/channelconfig/orderer_util.go"
"common/channelconfig/organization.go"
"common/channelconfig/standardvalues.go"
"common/channelconfig/template.go"

"common/cauthdsl/cauthdsl.go"
"common/cauthdsl/cauthdsl_builder.go"
"common/cauthdsl/policy.go"
"common/cauthdsl/policy_util.go"
"common/cauthdsl/policyparser.go"

"common/configtx/compare.go"
"common/configtx/configmap.go"
"common/configtx/manager.go"
"common/configtx/template.go"
"common/configtx/update.go"
"common/configtx/util.go"
"common/configtx/api/api.go"

"common/policies/implicitmeta.go"
"common/policies/implicitmeta_util.go"
"common/policies/policy.go"

"core/ledger/kvledger/txmgmt/rwsetutil/query_results_helper.go"
"core/ledger/kvledger/txmgmt/rwsetutil/rwset_builder.go"
"core/ledger/kvledger/txmgmt/rwsetutil/rwset_proto_util.go"

"core/ledger/kvledger/txmgmt/version/version.go"

"core/ledger/util/txvalidationflags.go"
"core/ledger/util/util.go"

"events/consumer/adapter.go"
"events/consumer/consumer.go"

"msp/cert.go"
"msp/configbuilder.go"
"msp/identities.go"
"msp/msp.go"
"msp/mspimpl.go"
"msp/mspmgrimpl.go"

"msp/cache/cache.go"
"msp/mgmt/deserializer.go"
"msp/mgmt/mgmt.go"
"msp/mgmt/principal.go"

"core/comm/config.go"
"core/comm/connection.go"
"core/comm/creds.go"
"core/comm/producer.go"
"core/comm/server.go"

"core/config/config.go"

"core/ledger/ledger_interface.go"

"core/ledger/kvledger/txmgmt/rwsetutil/query_results_helper.go"
"core/ledger/kvledger/txmgmt/rwsetutil/rwset_builder.go"
"core/ledger/kvledger/txmgmt/rwsetutil/rwset_proto_util.go"

"common/ledger/ledger_interface.go"

"common/ledger/util/ioutil.go"
"common/ledger/util/protobuf_util.go"
"common/ledger/util/util.go"

"protos/utils/blockutils.go"
"protos/utils/commonutils.go"
"protos/utils/proputils.go"
"protos/utils/txutils.go"

"protos/common/block.go"
"protos/common/common.go"
"protos/common/common.pb.go"
"protos/common/configtx.go"
"protos/common/configtx.pb.go"
"protos/common/configuration.go"
"protos/common/configuration.pb.go"
"protos/common/ledger.pb.go"
"protos/common/policies.go"
"protos/common/policies.pb.go"
"protos/common/signed_data.go"

"protos/ledger/rwset/rwset.pb.go"
"protos/ledger/rwset/kvrwset/kv_rwset.pb.go"
"protos/ledger/rwset/kvrwset/helper.go"

"protos/msp/identities.pb.go"
"protos/msp/msp_config.go"
"protos/msp/msp_config.pb.go"
"protos/msp/msp_principal.go"
"protos/msp/msp_principal.pb.go"

"protos/orderer/ab.pb.go"
"protos/orderer/configuration.go"
"protos/orderer/configuration.pb.go"
"protos/orderer/kafka.pb.go"

"protos/peer/admin.pb.go"
"protos/peer/chaincode.pb.go"
"protos/peer/chaincode_event.pb.go"
"protos/peer/chaincode_shim.pb.go"
"protos/peer/chaincodeunmarshall.go"
"protos/peer/configuration.go"
"protos/peer/configuration.pb.go"
"protos/peer/events.pb.go"
"protos/peer/init.go"
"protos/peer/peer.pb.go"
"protos/peer/proposal.go"
"protos/peer/proposal.pb.go"
"protos/peer/proposal_response.go"
"protos/peer/proposal_response.pb.go"
"protos/peer/query.pb.go"
"protos/peer/resources.go"
"protos/peer/resources.pb.go"
"protos/peer/signed_cc_dep_spec.pb.go"
"protos/peer/transaction.go"
"protos/peer/transaction.pb.go"
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
    sed -i '' -e $IMPORT_SUBST $INTERNAL_PATH/${i}
done

# Apply targetted patches