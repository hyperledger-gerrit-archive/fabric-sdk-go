/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package proposer

import (
	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	txn "github.com/hyperledger/fabric-sdk-go/api/apitxn"

	"github.com/hyperledger/fabric-sdk-go/pkg/errors"
)

const (
	systemChannel = ""
)

// QueryByChaincode sends a proposal to one or more endorsing peers that will be handled by the chaincode.
// This request will be presented to the chaincode 'invoke' and must understand
// from the arguments that this is a query request. The chaincode must also return
// results in the byte array format and the caller will have to be able to decode.
// these results.
func QueryByChaincode(ctx context, channelID string, request txn.ChaincodeInvokeRequest) ([][]byte, error) {
	if err := validateChaincodeInvokeRequest(request); err != nil {
		return nil, err
	}

	transactionProposalResponses, _, err := SendTransactionProposal(ctx, channelID, request)
	if err != nil {
		return nil, errors.WithMessage(err, "SendTransactionProposalWithChannelID failed")
	}

	return filterProposalResponses(transactionProposalResponses)
}

func filterProposalResponses(tpr []*txn.TransactionProposalResponse) ([][]byte, error) {
	var responses [][]byte
	errMsg := ""
	for _, response := range tpr {
		if response.Err != nil {
			errMsg = errMsg + response.Err.Error() + "\n"
		} else {
			responses = append(responses, response.ProposalResponse.GetResponse().Payload)
		}
	}

	if len(errMsg) > 0 {
		return responses, errors.New(errMsg)
	}
	return responses, nil
}

// QueryBySystemChannel invokes a chaincode on the system channel.
func QueryBySystemChannel(ctx fab.Context, request txn.ChaincodeInvokeRequest) ([][]byte, error) {
	return QueryByChaincode(ctx, systemChannel, request)
}
