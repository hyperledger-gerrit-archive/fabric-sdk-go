/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package txn

import (
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/pkg/errors"

	"github.com/hyperledger/fabric-sdk-go/api/apifabclient"

	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
	protos_utils "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/utils"
	"github.com/hyperledger/fabric/common/crypto"
)

// NewID computes a TransactionID for the current user context
//
// TODO: Determine if this function should be exported after refactoring is completed.
func NewID(signingIdentity apifabclient.IdentityContext) (apifabclient.TransactionID, error) {
	// generate a random nonce
	nonce, err := crypto.GetRandomNonce()
	if err != nil {
		return apifabclient.TransactionID{}, err
	}

	creator, err := signingIdentity.Identity()
	if err != nil {
		return apifabclient.TransactionID{}, err
	}

	id, err := protos_utils.ComputeProposalTxID(nonce, creator)
	if err != nil {
		return apifabclient.TransactionID{}, err
	}

	txnID := apifabclient.TransactionID{
		ID:    id,
		Nonce: nonce,
	}

	return txnID, nil
}

// SignPayload signs payload
//
// TODO: Determine if this function should be exported after refactoring is completed.
func SignPayload(ctx context, payload []byte) (*apifabclient.SignedEnvelope, error) {
	signingMgr := ctx.SigningManager()
	signature, err := signingMgr.Sign(payload, ctx.PrivateKey())
	if err != nil {
		return nil, err
	}
	return &apifabclient.SignedEnvelope{Payload: payload, Signature: signature}, nil
}

// BuildChannelHeader is a utility method to build a common chain header (TODO refactor)
//
// TODO: Determine if this function should be exported after refactoring is completed.
func BuildChannelHeader(headerType common.HeaderType, channelID string, txID string, epoch uint64, chaincodeID string, timestamp time.Time, tlsCertHash []byte) (*common.ChannelHeader, error) {
	logger.Debugf("buildChannelHeader - headerType: %s channelID: %s txID: %d epoch: % chaincodeID: %s timestamp: %v", headerType, channelID, txID, epoch, chaincodeID, timestamp)
	channelHeader := &common.ChannelHeader{
		Type:        int32(headerType),
		Version:     1,
		ChannelId:   channelID,
		TxId:        txID,
		Epoch:       epoch,
		TlsCertHash: tlsCertHash,
	}

	ts, err := ptypes.TimestampProto(timestamp)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create timestamp in channel header")
	}
	channelHeader.Timestamp = ts

	if chaincodeID != "" {
		ccID := &pb.ChaincodeID{
			Name: chaincodeID,
		}
		headerExt := &pb.ChaincodeHeaderExtension{
			ChaincodeId: ccID,
		}
		headerExtBytes, err := proto.Marshal(headerExt)
		if err != nil {
			return nil, errors.Wrap(err, "marshal header extension failed")
		}
		channelHeader.Extension = headerExtBytes
	}
	return channelHeader, nil
}
