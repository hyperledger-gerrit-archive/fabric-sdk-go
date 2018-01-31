/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package chledger

import (
	"time"

	"github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/txn/env"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/txn/sender"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging"

	"github.com/golang/protobuf/proto"

	ab "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/protos/orderer"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	protos_utils "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/utils"

	ccomm "github.com/hyperledger/fabric-sdk-go/pkg/config/comm"
	"github.com/hyperledger/fabric-sdk-go/pkg/errors"
	fc "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/internal"
)

var logger = logging.NewLogger("fabric_sdk_go")

// OrdererLedger holds the context for interacting with a Channel's ledger via the Orderer.
type OrdererLedger struct {
	ctx apifabclient.Context
	cfg apifabclient.ChannelCfg
}

// FromOrderer allows interaction with a Channel's ledger based on the context and channel config.
func FromOrderer(ctx apifabclient.Context, cfg apifabclient.ChannelCfg) *OrdererLedger {
	l := OrdererLedger{ctx, cfg}
	return &l
}

// GenesisBlock returns the genesis block from the defined orderer that may be
// used in a join request
// request: An object containing the following fields:
//          `txId` : required - String of the transaction id
//          `nonce` : required - Integer of the once time number
//
// See /protos/peer/proposal_response.proto
func (l *OrdererLedger) GenesisBlock() (*common.Block, error) {
	logger.Debug("GenesisBlock - start")

	// verify that we have an orderer configured
	if len(l.cfg.Orderers()) == 0 {
		return nil, errors.New("GenesisBlock missing orderer assigned to this channel for the GenesisBlock request")
	}
	creator, err := l.ctx.Identity()
	if err != nil {
		return nil, errors.WithMessage(err, "failed to get creator identity")
	}

	txnID, err := env.NewTxnID(l.ctx)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to calculate transaction id")
	}

	// now build the seek info , will be used once the channel is created
	// to get the genesis block back
	seekStart := fc.NewSpecificSeekPosition(0)
	seekStop := fc.NewSpecificSeekPosition(0)
	seekInfo := &ab.SeekInfo{
		Start:    seekStart,
		Stop:     seekStop,
		Behavior: ab.SeekInfo_BLOCK_UNTIL_READY,
	}
	protos_utils.MakeChannelHeader(common.HeaderType_DELIVER_SEEK_INFO, 1, l.cfg.Name(), 0)
	tlsCertHash := ccomm.TLSCertHash(l.ctx.Config())
	seekInfoHeader, err := env.BuildChannelHeader(common.HeaderType_DELIVER_SEEK_INFO, l.cfg.Name(), txnID.ID, 0, "", time.Now(), tlsCertHash)
	if err != nil {
		return nil, errors.Wrap(err, "BuildChannelHeader failed")
	}
	seekHeader, err := fc.BuildHeader(creator, seekInfoHeader, txnID.Nonce)
	if err != nil {
		return nil, errors.Wrap(err, "BuildHeader failed")
	}
	seekPayload := &common.Payload{
		Header: seekHeader,
		Data:   fc.MarshalOrPanic(seekInfo),
	}
	seekPayloadBytes := fc.MarshalOrPanic(seekPayload)

	signedEnvelope, err := env.SignPayload(l.ctx, seekPayloadBytes)
	if err != nil {
		return nil, errors.WithMessage(err, "SignPayload failed")
	}

	block, err := sender.SendEnvelope(l.ctx, signedEnvelope, l.cfg.Orderers())
	if err != nil {
		return nil, errors.WithMessage(err, "SendEnvelope failed")
	}
	return block, nil
}

// ChannelConfig queries for the current config block for this channel.
// This transaction will be made to the orderer.
// @returns {ConfigEnvelope} Object containing the configuration items.
// @see /protos/orderer/ab.proto
// @see /protos/common/configtx.proto
func (l *OrdererLedger) ChannelConfig() (*common.ConfigEnvelope, error) {
	logger.Debugf("channelConfig - start for channel %s", l.cfg.Name())

	// Get the newest block
	block, err := l.block(fc.NewNewestSeekPosition())
	if err != nil {
		return nil, err
	}
	logger.Debugf("channelConfig - Retrieved newest block number: %d\n", block.Header.Number)

	// Get the index of the last config block
	lastConfig, err := fc.GetLastConfigFromBlock(block)
	if err != nil {
		return nil, errors.Wrap(err, "GetLastConfigFromBlock failed")
	}
	logger.Debugf("channelConfig - Last config index: %d\n", lastConfig.Index)

	// Get the last config block
	block, err = l.block(fc.NewSpecificSeekPosition(lastConfig.Index))

	if err != nil {
		return nil, errors.WithMessage(err, "retrieve block failed")
	}
	logger.Debugf("channelConfig - Last config block number %d, Number of tx: %d", block.Header.Number, len(block.Data.Data))

	if len(block.Data.Data) != 1 {
		return nil, errors.New("config block must contain one transaction")
	}

	return l.createConfigEnvelope(block.Data.Data[0])

}

func (l *OrdererLedger) createConfigEnvelope(data []byte) (*common.ConfigEnvelope, error) {

	envelope := &common.Envelope{}
	if err := proto.Unmarshal(data, envelope); err != nil {
		return nil, errors.Wrap(err, "unmarshal envelope from config block failed")
	}
	payload := &common.Payload{}
	if err := proto.Unmarshal(envelope.Payload, payload); err != nil {
		return nil, errors.Wrap(err, "unmarshal payload from envelope failed")
	}
	channelHeader := &common.ChannelHeader{}
	if err := proto.Unmarshal(payload.Header.ChannelHeader, channelHeader); err != nil {
		return nil, errors.Wrap(err, "unmarshal payload from envelope failed")
	}
	if common.HeaderType(channelHeader.Type) != common.HeaderType_CONFIG {
		return nil, errors.New("block must be of type 'CONFIG'")
	}
	configEnvelope := &common.ConfigEnvelope{}
	if err := proto.Unmarshal(payload.Data, configEnvelope); err != nil {
		return nil, errors.Wrap(err, "unmarshal config envelope failed")
	}

	return configEnvelope, nil
}

// block retrieves the block at the given position
func (l *OrdererLedger) block(pos *ab.SeekPosition) (*common.Block, error) {
	nonce, err := fc.GenerateRandomNonce()
	if err != nil {
		return nil, errors.Wrap(err, "GenerateRandomNonce failed")
	}

	creator, err := l.ctx.Identity()
	if err != nil {
		return nil, errors.WithMessage(err, "serializing identity failed")
	}

	txID, err := protos_utils.ComputeProposalTxID(nonce, creator)
	if err != nil {
		return nil, errors.Wrap(err, "generating TX ID failed")
	}

	tlsCertHash := ccomm.TLSCertHash(l.ctx.Config())
	seekInfoHeader, err := env.BuildChannelHeader(common.HeaderType_DELIVER_SEEK_INFO, l.cfg.Name(), txID, 0, "", time.Now(), tlsCertHash)
	if err != nil {
		return nil, errors.Wrap(err, "BuildChannelHeader failed")
	}

	seekInfoHeaderBytes, err := proto.Marshal(seekInfoHeader)
	if err != nil {
		return nil, errors.Wrap(err, "marshal seek info failed")
	}

	signatureHeader := &common.SignatureHeader{
		Creator: creator,
		Nonce:   nonce,
	}

	signatureHeaderBytes, err := proto.Marshal(signatureHeader)
	if err != nil {
		return nil, errors.Wrap(err, "marshal signature header failed")
	}

	seekHeader := &common.Header{
		ChannelHeader:   seekInfoHeaderBytes,
		SignatureHeader: signatureHeaderBytes,
	}

	seekInfo := &ab.SeekInfo{
		Start:    pos,
		Stop:     pos,
		Behavior: ab.SeekInfo_BLOCK_UNTIL_READY,
	}

	seekInfoBytes, err := proto.Marshal(seekInfo)
	if err != nil {
		return nil, errors.Wrap(err, "marshal seek info failed")
	}

	seekPayload := &common.Payload{
		Header: seekHeader,
		Data:   seekInfoBytes,
	}

	seekPayloadBytes, err := proto.Marshal(seekPayload)
	if err != nil {
		return nil, err
	}

	signedEnvelope, err := env.SignPayload(l.ctx, seekPayloadBytes)
	if err != nil {
		return nil, errors.WithMessage(err, "SignPayload failed")
	}

	return sender.SendEnvelope(l.ctx, signedEnvelope)
}
