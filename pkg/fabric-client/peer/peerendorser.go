/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package peer

import (
	"time"

	"golang.org/x/net/context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
	grpcstatus "google.golang.org/grpc/status"

	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	"github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/pkg/errors/status"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"

	"crypto/x509"

	"github.com/hyperledger/fabric-sdk-go/pkg/config/comm"
	"github.com/hyperledger/fabric-sdk-go/pkg/config/urlutil"
	"github.com/pkg/errors"
)

// peerEndorser enables access to a GRPC-based endorser for running transaction proposal simulations
type peerEndorser struct {
	grpcDialOption       []grpc.DialOption
	target               string
	dialTimeout          time.Duration
	failFast             bool
	transportCredentials credentials.TransportCredentials
	secured              bool
	allowInsecure        bool
}

type peerEndorserRequest struct {
	target             string
	certificate        *x509.Certificate
	serverHostOverride string
	dialBlocking       bool
	config             apiconfig.Config
	kap                keepalive.ClientParameters
	failFast           bool
	allowInsecure      bool
}

func newPeerEndorser(endorseReq *peerEndorserRequest) (*peerEndorser, error) {
	if len(endorseReq.target) == 0 {
		return nil, errors.New("target is required")
	}

	// Construct dialer options for the connection
	var opts []grpc.DialOption
	if endorseReq.kap.Time > 0 || endorseReq.kap.Timeout > 0 {
		opts = append(opts, grpc.WithKeepaliveParams(endorseReq.kap))
	}
	opts = append(opts, grpc.WithDefaultCallOptions(grpc.FailFast(endorseReq.failFast)))

	timeout := endorseReq.config.TimeoutOrDefault(apiconfig.Endorser)

	if endorseReq.dialBlocking { // TODO: configurable?
		opts = append(opts, grpc.WithBlock())
	}

	tlsConfig, err := comm.TLSConfig(endorseReq.certificate, endorseReq.serverHostOverride, endorseReq.config)
	if err != nil {
		return nil, err
	}

	pc := &peerEndorser{grpcDialOption: opts, target: urlutil.ToAddress(endorseReq.target), dialTimeout: timeout,
		transportCredentials: credentials.NewTLS(tlsConfig), secured: urlutil.AttemptSecured(endorseReq.target),
		allowInsecure: endorseReq.allowInsecure}

	return pc, nil
}

// ProcessTransactionProposal sends the transaction proposal to a peer and returns the response.
func (p *peerEndorser) ProcessTransactionProposal(proposal apifabclient.TransactionProposal) (apifabclient.TransactionProposalResponse, error) {
	logger.Debugf("Processing proposal using endorser :%s", p.target)

	proposalResponse, err := p.sendProposal(proposal, p.secured)
	if err != nil {
		return apifabclient.TransactionProposalResponse{
				Proposal: proposal,
				Endorser: p.target,
			}, errors.Wrapf(err, "Transaction processor (%s) returned error for txID '%s'",
				p.target, proposal.TxnID.ID)
	}

	return apifabclient.TransactionProposalResponse{
		Proposal:         proposal,
		ProposalResponse: proposalResponse,
		Endorser:         p.target, // TODO: what format is expected for Endorser? Just target? URL?
		Status:           proposalResponse.GetResponse().Status,
	}, nil
}

func (p *peerEndorser) conn(secured bool) (*grpc.ClientConn, error) {
	// Establish connection to Ordering Service
	var grpcOpts []grpc.DialOption
	if secured {
		grpcOpts = append(p.grpcDialOption, grpc.WithTransportCredentials(p.transportCredentials))
	} else {
		grpcOpts = append(p.grpcDialOption, grpc.WithInsecure())
	}

	ctx := context.Background()
	ctx, _ = context.WithTimeout(ctx, p.dialTimeout)

	return grpc.DialContext(ctx, p.target, grpcOpts...)
}

func (p *peerEndorser) releaseConn(conn *grpc.ClientConn) {
	conn.Close()
}

func (p *peerEndorser) sendProposal(proposal apifabclient.TransactionProposal, secured bool) (*pb.ProposalResponse, error) {
	conn, err := p.conn(secured)
	if err != nil {
		if secured && p.allowInsecure {
			//If secured mode failed and allow in-secure is enabled then retry in in-secure mode
			logger.Debug("Secured NewEndorserClient failed, attempting in-secured")
			return p.sendProposal(proposal, false)
		}
		return nil, status.New(status.EndorserClientStatus, status.ConnectionFailed.ToInt32(), err.Error(), []interface{}{p.target})
	}
	defer p.releaseConn(conn)

	endorserClient := pb.NewEndorserClient(conn)
	resp, err := endorserClient.ProcessProposal(context.Background(), proposal.SignedProposal)
	if err != nil {
		logger.Error("NewEndorserClient failed, cause : ", err)
		if secured && p.allowInsecure {
			//If secured mode failed and allow in-secure is enabled then retry in in-secure mode
			logger.Debug("Secured NewEndorserClient failed, attempting in-secured")
			return p.sendProposal(proposal, false)
		}

		rpcStatus, ok := grpcstatus.FromError(err)
		if ok {
			err = status.NewFromGRPCStatus(rpcStatus)
		}
	}
	return resp, err
}
