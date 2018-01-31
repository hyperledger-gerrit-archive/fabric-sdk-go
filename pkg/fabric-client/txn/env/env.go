/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package env

import (
	"math/rand"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/hyperledger/fabric-sdk-go/api/apicryptosuite"
	"github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/api/apitxn"
	"github.com/hyperledger/fabric-sdk-go/pkg/errors"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
	protos_utils "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/utils"
	"github.com/hyperledger/fabric/common/crypto"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type context interface {
	SigningManager() apifabclient.SigningManager
	PrivateKey() apicryptosuite.Key
}

// NewTxnID computes a TransactionID for the current user context
func NewTxnID(signingIdentity apifabclient.IdentityContext) (apitxn.TransactionID, error) {
	// generate a random nonce
	nonce, err := crypto.GetRandomNonce()
	if err != nil {
		return apitxn.TransactionID{}, err
	}

	creator, err := signingIdentity.Identity()
	if err != nil {
		return apitxn.TransactionID{}, err
	}

	id, err := protos_utils.ComputeProposalTxID(nonce, creator)
	if err != nil {
		return apitxn.TransactionID{}, err
	}

	txnID := apitxn.TransactionID{
		ID:    id,
		Nonce: nonce,
	}

	return txnID, nil
}

// SignPayload signs payload
func SignPayload(ctx context, payload []byte) (*apifabclient.SignedEnvelope, error) {
	signingMgr := ctx.SigningManager()
	signature, err := signingMgr.Sign(payload, ctx.PrivateKey())
	if err != nil {
		return nil, err
	}
	return &apifabclient.SignedEnvelope{Payload: payload, Signature: signature}, nil
}

// BuildChannelHeader is a utility method to build a common chain header (TODO refactor)
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
