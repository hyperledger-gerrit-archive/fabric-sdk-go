package test

import (
	"errors"
	"fmt"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/orderer"
)

func ModifyMaxMessageCount(block *common.Block) (uint32, error) {

	// Extract Config from Block
	blockPayload := block.Data.Data[0]

	envelope := &common.Envelope{}
	if err := proto.Unmarshal(blockPayload, envelope); err != nil {
		return 0, err
	}
	payload := &common.Payload{}
	if err := proto.Unmarshal(envelope.Payload, payload); err != nil {
		return 0, err
	}

	cfgEnv := &common.ConfigEnvelope{}
	if err := proto.Unmarshal(payload.Data, cfgEnv); err != nil {
		return 0, err
	}
	config := cfgEnv.Config

	// Modify Config
	batchSizeBytes := config.ChannelGroup.Groups["Orderer"].Values["BatchSize"].Value
	batchSize := &orderer.BatchSize{}
	if err := proto.Unmarshal(batchSizeBytes, batchSize); err != nil {
		return 0, err
	}
	batchSize.MaxMessageCount = batchSize.MaxMessageCount + 1
	newMatchSizeBytes, err := proto.Marshal(batchSize)
	if err != nil {
		return 0, err
	}
	config.ChannelGroup.Groups["Orderer"].Values["BatchSize"].Value = newMatchSizeBytes

	// Repackage Block
	newCfgEnv, err := proto.Marshal(cfgEnv)
	if err != nil {
		return 0, err
	}
	payload.Data = newCfgEnv
	newPayload, err := proto.Marshal(payload)
	if err != nil {
		return 0, err
	}
	envelope.Payload = newPayload
	newEnvelope, err := proto.Marshal(envelope)
	if err != nil {
		return 0, err
	}
	block.Data.Data[0] = newEnvelope

	return batchSize.MaxMessageCount, nil
}

func VerifyMaxMessageCount(block *common.Block, expected uint32) error {

	// Extract Config from Block
	blockPayload := block.Data.Data[0]

	envelope := &common.Envelope{}
	if err := proto.Unmarshal(blockPayload, envelope); err != nil {
		return err
	}
	payload := &common.Payload{}
	if err := proto.Unmarshal(envelope.Payload, payload); err != nil {
		return err
	}

	cfgEnv := &common.ConfigEnvelope{}
	if err := proto.Unmarshal(payload.Data, cfgEnv); err != nil {
		return err
	}
	config := cfgEnv.Config

	// Modify Config
	batchSizeBytes := config.ChannelGroup.Groups["Orderer"].Values["BatchSize"].Value
	batchSize := &orderer.BatchSize{}
	if err := proto.Unmarshal(batchSizeBytes, batchSize); err != nil {
		return err
	}

	if batchSize.MaxMessageCount != expected {
		return errors.New(fmt.Sprintf("Unexpected MaxMessageCount. actual: %d, expected: %d", batchSize.MaxMessageCount, expected))
	}
	return nil
}