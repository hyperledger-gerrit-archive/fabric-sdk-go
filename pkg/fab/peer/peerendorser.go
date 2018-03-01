/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package peer

import (
	grpccontext "context"
	"crypto/x509"
	"time"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
	grpcstatus "google.golang.org/grpc/status"

	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config/comm"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config/urlutil"
	"github.com/hyperledger/fabric-sdk-go/pkg/errors/status"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
)

// peerEndorser enables access to a GRPC-based endorser for running transaction proposal simulations
type peerEndorser struct {
	grpcDialOption []grpc.DialOption
	target         string
	dialTimeout    time.Duration
	failFast       bool
}

type peerEndorserRequest struct {
	target             string
	certificate        *x509.Certificate
	serverHostOverride string
	dialBlocking       bool
	config             core.Config
	kap                keepalive.ClientParameters
	failFast           bool
	allowInsecure      bool
}

func newPeerEndorser(endorseReq *peerEndorserRequest) (*peerEndorser, error) {
	if len(endorseReq.target) == 0 {
		return nil, errors.New("target is required")
	}

	// Construct dialer options for the connection
	var grpcOpts []grpc.DialOption
	if endorseReq.kap.Time > 0 || endorseReq.kap.Timeout > 0 {
		grpcOpts = append(grpcOpts, grpc.WithKeepaliveParams(endorseReq.kap))
	}
	grpcOpts = append(grpcOpts, grpc.WithDefaultCallOptions(grpc.FailFast(endorseReq.failFast)))

	if endorseReq.dialBlocking { // TODO: configurable?
		grpcOpts = append(grpcOpts, grpc.WithBlock())
	}

	if urlutil.AttemptSecured(endorseReq.target, endorseReq.allowInsecure) {
		tlsConfig, err := comm.TLSConfig(endorseReq.certificate, endorseReq.serverHostOverride, endorseReq.config)
		if err != nil {
			return nil, err
		}
		grpcOpts = append(grpcOpts, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	} else {
		grpcOpts = append(grpcOpts, grpc.WithInsecure())
	}

	timeout := endorseReq.config.TimeoutOrDefault(core.Endorser)

	pc := &peerEndorser{
		grpcDialOption: grpcOpts,
		target:         urlutil.ToAddress(endorseReq.target),
		dialTimeout:    timeout,
	}

	return pc, nil
}

// ProcessTransactionProposal sends the transaction proposal to a peer and returns the response.
func (p *peerEndorser) ProcessTransactionProposal(request fab.ProcessProposalRequest) (*fab.TransactionProposalResponse, error) {
	logger.Debugf("Processing proposal using endorser: %s", p.target)

	proposalResponse, err := p.sendProposal(request)
	if err != nil {
		tpr := fab.TransactionProposalResponse{Endorser: p.target}
		return &tpr, errors.Wrapf(err, "Transaction processing for endorser [%s]", p.target)
	}

	tpr := fab.TransactionProposalResponse{
		ProposalResponse: proposalResponse,
		Endorser:         p.target,
		Status:           proposalResponse.GetResponse().Status,
	}
	return &tpr, nil
}

func (p *peerEndorser) conn() (*grpc.ClientConn, error) {
	// Establish connection to Ordering Service
	ctx := grpccontext.Background()
	ctx, cancel := grpccontext.WithTimeout(ctx, p.dialTimeout)
	defer cancel()

	return grpc.DialContext(ctx, p.target, p.grpcDialOption...)
}

func (p *peerEndorser) releaseConn(conn *grpc.ClientConn) {
	conn.Close()
}

func (p *peerEndorser) sendProposal(proposal fab.ProcessProposalRequest) (*pb.ProposalResponse, error) {
	conn, err := p.conn()
	if err != nil {
		return nil, status.New(status.EndorserClientStatus, status.ConnectionFailed.ToInt32(), err.Error(), []interface{}{p.target})
	}
	defer p.releaseConn(conn)

	endorserClient := pb.NewEndorserClient(conn)
	resp, err := endorserClient.ProcessProposal(grpccontext.Background(), proposal.SignedProposal)
	if err != nil {
		rpcStatus, ok := grpcstatus.FromError(err)
		if ok {
			err = status.NewFromGRPCStatus(rpcStatus)
		}
	}
	return resp, err
}
