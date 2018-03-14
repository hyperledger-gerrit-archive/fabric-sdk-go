/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package channel

import (
	reqContext "context"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/errors/retry"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
	"github.com/pkg/errors"
)

// opts allows the user to specify more advanced options
type requestOptions struct {
	Targets    []fab.Peer // targets
	Retry      retry.Opts
	Timeouts   map[core.TimeoutType]time.Duration //timeout options for channel client operations
	Background reqContext.Context                 //background grpc context for channel client operations (query, execute, invokehandler)
}

// RequestOption func for each Opts argument
type RequestOption func(ctx context.Client, opts *requestOptions) error

// Request contains the parameters to query and execute an invocation transaction
type Request struct {
	ChaincodeID  string
	Fcn          string
	Args         [][]byte
	TransientMap map[string][]byte
}

//Response contains response parameters for query and execute an invocation transaction
type Response struct {
	Payload          []byte
	TransactionID    fab.TransactionID
	TxValidationCode pb.TxValidationCode
	Proposal         *fab.TransactionProposal
	Responses        []*fab.TransactionProposalResponse
}

//WithTargets encapsulates ProposalProcessors to Option
func WithTargets(targets ...fab.Peer) RequestOption {
	return func(ctx context.Client, o *requestOptions) error {
		o.Targets = targets
		return nil
	}
}

// WithTargetURLs allows overriding of the target peers for the request.
// Targets are specified by URL, and the SDK will create the underlying peer
// objects.
func WithTargetURLs(urls ...string) RequestOption {
	return func(ctx context.Client, opts *requestOptions) error {

		var targets []fab.Peer

		for _, url := range urls {

			peerCfg, err := config.NetworkPeerConfigFromURL(ctx.Config(), url)
			if err != nil {
				return err
			}

			peer, err := ctx.InfraProvider().CreatePeerFromConfig(peerCfg)
			if err != nil {
				return errors.WithMessage(err, "creating peer from config failed")
			}

			targets = append(targets, peer)
		}

		return WithTargets(targets...)(ctx, opts)
	}
}

// WithRetry option to configure retries
func WithRetry(retryOpt retry.Opts) RequestOption {
	return func(ctx context.Client, o *requestOptions) error {
		o.Retry = retryOpt
		return nil
	}
}

//WithTimeout encapsulates key value pairs of timeout type, timeout duration to Options
func WithTimeout(timeoutType core.TimeoutType, timeout time.Duration) RequestOption {
	return func(ctx context.Client, o *requestOptions) error {
		if o.Timeouts == nil {
			o.Timeouts = make(map[core.TimeoutType]time.Duration)
		}
		o.Timeouts[timeoutType] = timeout
		return nil
	}
}

//WithBackground encapsulates grpc context background to Options
func WithBackground(background reqContext.Context) RequestOption {
	return func(ctx context.Client, o *requestOptions) error {
		o.Background = background
		return nil
	}
}
