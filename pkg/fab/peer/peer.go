/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package peer

import (
	reqContext "context"

	"crypto/x509"

	"github.com/spf13/cast"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"

	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/errors/status"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging"
)

var logger = logging.NewLogger("fabsdk/fab")

// Peer represents a node in the target blockchain network to which
// HFC sends endorsement proposals, transaction ordering or query requests.
type Peer struct {
	config      core.Config
	certificate *x509.Certificate
	serverName  string
	processor   fab.ProposalProcessor
	mspID       string
	url         string
	kap         keepalive.ClientParameters
	failFast    bool
	inSecure    bool
	commManager fab.CommManager
}

// Option describes a functional parameter for the New constructor
type Option func(*Peer) error

// New Returns a new Peer instance
func New(config core.Config, opts ...Option) (*Peer, error) {
	peer := &Peer{
		config:      config,
		commManager: &defCommManager{},
	}

	for _, opt := range opts {
		err := opt(peer)

		if err != nil {
			return nil, err
		}
	}

	if peer.processor == nil {
		// TODO: config is declaring TLS but cert & serverHostOverride is being passed-in...
		endorseRequest := peerEndorserRequest{
			target:             peer.url,
			certificate:        peer.certificate,
			serverHostOverride: peer.serverName,
			config:             peer.config,
			kap:                peer.kap,
			failFast:           peer.failFast,
			allowInsecure:      peer.inSecure,
			commManager:        peer.commManager,
		}
		processor, err := newPeerEndorser(&endorseRequest)

		if err != nil {
			return nil, err
		}
		peer.processor = processor
	}

	return peer, nil
}

// WithURL is a functional option for the peer.New constructor that configures the peer's URL
func WithURL(url string) Option {
	return func(p *Peer) error {
		p.url = url

		return nil
	}
}

// WithTLSCert is a functional option for the peer.New constructor that configures the peer's TLS certificate
func WithTLSCert(certificate *x509.Certificate) Option {
	return func(p *Peer) error {
		p.certificate = certificate

		return nil
	}
}

// WithServerName is a functional option for the peer.New constructor that configures the peer's server name
func WithServerName(serverName string) Option {
	return func(p *Peer) error {
		p.serverName = serverName

		return nil
	}
}

// WithInsecure is a functional option for the peer.New constructor that configures the peer's grpc insecure option
func WithInsecure() Option {
	return func(p *Peer) error {
		p.inSecure = true

		return nil
	}
}

// WithMSPID is a functional option for the peer.New constructor that configures the peer's msp ID
func WithMSPID(mspID string) Option {
	return func(p *Peer) error {
		p.mspID = mspID

		return nil
	}
}

// FromPeerConfig is a functional option for the peer.New constructor that configures a new peer
// from a apiconfig.NetworkPeer struct
func FromPeerConfig(peerCfg *core.NetworkPeer) Option {
	return func(p *Peer) error {

		p.url = peerCfg.URL
		p.serverName = getServerNameOverride(peerCfg)
		p.inSecure = isInsecureConnectionAllowed(peerCfg)

		var err error
		p.certificate, err = peerCfg.TLSCACerts.TLSCert()

		if err != nil {
			//Ignore empty cert errors,
			errStatus, ok := err.(*status.Status)
			if !ok || errStatus.Code != status.EmptyCert.ToInt32() {
				return err
			}
		}

		// TODO: Remove upon making peer interface immutable
		p.mspID = peerCfg.MspID
		p.kap = getKeepAliveOptions(peerCfg)
		p.failFast = getFailFast(peerCfg)
		return nil
	}
}

func getServerNameOverride(peerCfg *core.NetworkPeer) string {
	serverHostOverride := ""
	if str, ok := peerCfg.GRPCOptions["ssl-target-name-override"].(string); ok {
		serverHostOverride = str
	}

	return serverHostOverride
}

func getFailFast(peerCfg *core.NetworkPeer) bool {
	var failFast = true
	if ff, ok := peerCfg.GRPCOptions["fail-fast"].(bool); ok {
		failFast = cast.ToBool(ff)
	}

	return failFast
}

func getKeepAliveOptions(peerCfg *core.NetworkPeer) keepalive.ClientParameters {

	var kap keepalive.ClientParameters
	if kaTime, ok := peerCfg.GRPCOptions["keep-alive-time"]; ok {
		kap.Time = cast.ToDuration(kaTime)
	}
	if kaTimeout, ok := peerCfg.GRPCOptions["keep-alive-timeout"]; ok {
		kap.Timeout = cast.ToDuration(kaTimeout)
	}
	if kaPermit, ok := peerCfg.GRPCOptions["keep-alive-permit"]; ok {
		kap.PermitWithoutStream = cast.ToBool(kaPermit)
	}
	return kap
}

func isInsecureConnectionAllowed(peerCfg *core.NetworkPeer) bool {
	allowInsecure, ok := peerCfg.GRPCOptions["allow-insecure"].(bool)
	if ok {
		return allowInsecure
	}
	return false
}

// WithPeerProcessor is a functional option for the peer.New constructor that configures the peer's proposal processor
func WithPeerProcessor(processor fab.ProposalProcessor) Option {
	return func(p *Peer) error {
		p.processor = processor

		return nil
	}
}

// MSPID gets the Peer mspID.
func (p *Peer) MSPID() string {
	return p.mspID
}

// URL gets the Peer URL. Required property for the instance objects.
// It returns the address of the Peer.
func (p *Peer) URL() string {
	return p.url
}

// ProcessTransactionProposal sends the created proposal to peer for endorsement.
func (p *Peer) ProcessTransactionProposal(ctx reqContext.Context, proposal fab.ProcessProposalRequest) (*fab.TransactionProposalResponse, error) {
	return p.processor.ProcessTransactionProposal(ctx, proposal)
}

func (p *Peer) String() string {
	return p.url
}

// PeersToTxnProcessors converts a slice of Peers to a slice of TxnProposalProcessors
func PeersToTxnProcessors(peers []fab.Peer) []fab.ProposalProcessor {
	tpp := make([]fab.ProposalProcessor, len(peers))

	for i := range peers {
		tpp[i] = peers[i]
	}
	return tpp
}

type defCommManager struct{}

func (*defCommManager) DialContext(ctx reqContext.Context, target string, opts ...grpc.DialOption) (conn *grpc.ClientConn, err error) {
	logger.Debugf("DialContext [%s]", target)
	opts = append(opts, grpc.WithBlock())
	return grpc.DialContext(ctx, target, opts...)
}

func (*defCommManager) ReleaseConn(conn *grpc.ClientConn) {
	logger.Debugf("ReleaseConn [%p]", conn)
	if err := conn.Close(); err != nil {
		logger.Debugf("unable to close connection [%s]", err)
	}
}
