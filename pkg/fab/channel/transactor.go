/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package channel

import (
	"strings"

	"github.com/pkg/errors"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/fab"

	reqContext "context"

	contextImpl "github.com/hyperledger/fabric-sdk-go/pkg/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config/endpoint"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/orderer"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/txn"
)

// Transactor enables sending transactions and transaction proposals on the channel.
type Transactor struct {
	reqCtx    reqContext.Context
	ChannelID string
	orderers  []fab.Orderer
}

// NewTransactor returns a Transactor for the current context and channel config.
func NewTransactor(reqCtx reqContext.Context, cfg fab.ChannelCfg) (*Transactor, error) {

	ctx, ok := contextImpl.RequestClientContext(reqCtx)
	if !ok {
		return nil, errors.New("failed get client context from reqContext for create new transactor")
	}

	orderers, err := orderersFromChannelCfg(ctx, cfg)
	if err != nil {
		return nil, errors.WithMessage(err, "reading orderers from channel config failed")
	}
	// TODO: adjust integration tests to always have valid orderers (even when doing only queries)
	//if len(orderers) == 0 {
	//	return nil, errors.New("orderers are not configured")
	//}

	t := Transactor{
		reqCtx:    reqCtx,
		ChannelID: cfg.ID(),
		orderers:  orderers,
	}
	return &t, nil
}

func orderersFromChannelCfg(ctx context.Client, cfg fab.ChannelCfg) ([]fab.Orderer, error) {
	orderers := []fab.Orderer{}
	ordererDict, err := orderersByTarget(ctx)
	if err != nil {
		return nil, err
	}

	// Add orderer if specified in config
	for _, target := range cfg.Orderers() {

		// Figure out orderer configuration
		oCfg, ok := ordererDict[target]

		if !ok {
			// Try to find a match from entityMatchers config
			matchingOrdererConfig, matchErr := ctx.Config().OrdererConfig(strings.ToLower(target))
			if matchErr == nil && matchingOrdererConfig != nil {
				oCfg = *matchingOrdererConfig
				ok = true
			}

		}
		if !ok {
			o, err := orderer.New(ctx.Config(), orderer.WithURL(target))
			// TODO: should we fail hard if we cannot configure a default orderer?
			//if err != nil {
			//	return nil, errors.WithMessage(err, "failed to create orderer from defaults")
			//}
			if err == nil {
				orderers = append(orderers, o)
			}
		} else {
			o, err := orderer.New(ctx.Config(), orderer.FromOrdererConfig(&oCfg))
			if err != nil {
				return nil, errors.WithMessage(err, "failed to create orderer from config")
			}
			orderers = append(orderers, o)
		}
	}
	return orderers, nil
}

func orderersByTarget(ctx context.Client) (map[string]core.OrdererConfig, error) {
	ordererDict := map[string]core.OrdererConfig{}
	orderersConfig, err := ctx.Config().OrderersConfig()
	if err != nil {
		return nil, errors.WithMessage(err, "loading orderers config failed")
	}

	for _, oc := range orderersConfig {
		address := endpoint.ToAddress(oc.URL)
		ordererDict[address] = oc
	}
	return ordererDict, nil
}

// CreateTransactionHeader creates a Transaction Header based on the current context.
func (t *Transactor) CreateTransactionHeader() (fab.TransactionHeader, error) {

	ctx, ok := contextImpl.RequestClientContext(t.reqCtx)
	if !ok {
		return nil, errors.New("failed get client context from reqContext for txn Header")
	}

	txh, err := txn.NewHeader(ctx, t.ChannelID)
	if err != nil {
		return nil, errors.WithMessage(err, "new transaction ID failed")
	}

	return txh, nil
}

// SendTransactionProposal sends a TransactionProposal to the target peers.
func (t *Transactor) SendTransactionProposal(proposal *fab.TransactionProposal, targets []fab.ProposalProcessor) ([]*fab.TransactionProposalResponse, error) {
	ctx, ok := contextImpl.RequestClientContext(t.reqCtx)
	if !ok {
		return nil, errors.New("failed get client context from reqContext for SendTransactionProposal")
	}

	reqCtx, cancel := contextImpl.NewRequest(ctx, contextImpl.WithTimeoutKey(core.PeerResponse), contextImpl.WithReqContext(t.reqCtx))
	defer cancel()

	return txn.SendProposal(reqCtx, proposal, targets)
}

// CreateTransaction create a transaction with proposal response.
// TODO: should this be removed as it is purely a wrapper?
func (t *Transactor) CreateTransaction(request fab.TransactionRequest) (*fab.Transaction, error) {
	return txn.New(request)
}

// SendTransaction send a transaction to the chain’s orderer service (one or more orderer endpoints) for consensus and committing to the ledger.
func (t *Transactor) SendTransaction(tx *fab.Transaction) (*fab.TransactionResponse, error) {
	ctx, ok := contextImpl.RequestClientContext(t.reqCtx)
	if !ok {
		return nil, errors.New("failed get client context from reqContext for SendTransaction")
	}

	reqCtx, cancel := contextImpl.NewRequest(ctx, contextImpl.WithTimeoutKey(core.OrdererResponse), contextImpl.WithReqContext(t.reqCtx))
	defer cancel()

	return txn.Send(reqCtx, tx, t.orderers)
}
