/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package endpoint

import (
	"crypto/x509"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/status"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/options"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/comm"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/spf13/cast"
	"google.golang.org/grpc/keepalive"
)

// EventEndpoint extends a Peer endpoint and provides the
// event URL, which, in the case of Eventhub, is different
// from the peer endpoint
type EventEndpoint struct {
	Certificate *x509.Certificate
	fab.Peer
	EvtURL          string
	HostOverride    string
	KeepAliveParams keepalive.ClientParameters
	ConnectTimeout  time.Duration
	FailFast        bool
	AllowInsecure   bool
}

// EventURL returns the event URL
func (e *EventEndpoint) EventURL() string {
	return e.EvtURL
}

// Opts returns additional options for the event connection
func (e *EventEndpoint) Opts() []options.Opt {
	opts := []options.Opt{
		comm.WithHostOverride(e.HostOverride),
		comm.WithFailFast(e.FailFast),
		comm.WithKeepAliveParams(e.KeepAliveParams),
		comm.WithCertificate(e.Certificate),
		comm.WithConnectTimeout(e.ConnectTimeout),
	}
	if e.AllowInsecure {
		opts = append(opts, comm.WithInsecure())
	}
	return opts
}

// FromPeerConfig creates a new EventEndpoint from the given config
func FromPeerConfig(config fab.EndpointConfig, peer fab.Peer, peerCfg *fab.PeerConfig) (*EventEndpoint, error) {
	certificate, err := peerCfg.TLSCACerts.TLSCert()
	if err != nil {
		//Ignore empty cert errors,
		errStatus, ok := err.(*status.Status)
		if !ok || errStatus.Code != status.EmptyCert.ToInt32() {
			return nil, err
		}
	}

	return &EventEndpoint{
		Peer:            peer,
		EvtURL:          peerCfg.EventURL,
		HostOverride:    getServerNameOverride(peerCfg),
		Certificate:     certificate,
		KeepAliveParams: getKeepAliveOptions(peerCfg),
		FailFast:        getFailFast(peerCfg),
		ConnectTimeout:  config.Timeout(fab.EventHubConnection),
		AllowInsecure:   isInsecureAllowed(peerCfg),
	}, nil
}

func getServerNameOverride(peerCfg *fab.PeerConfig) string {
	if str, ok := peerCfg.GRPCOptions["ssl-target-name-override"].(string); ok {
		return str
	}
	return ""
}

func getFailFast(peerCfg *fab.PeerConfig) bool {
	if ff, ok := peerCfg.GRPCOptions["fail-fast"].(bool); ok {
		return cast.ToBool(ff)
	}
	return false
}

func getKeepAliveOptions(peerCfg *fab.PeerConfig) keepalive.ClientParameters {
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

func isInsecureAllowed(peerCfg *fab.PeerConfig) bool {
	allowInsecure, ok := peerCfg.GRPCOptions["allow-insecure"].(bool)
	if ok {
		return allowInsecure
	}
	return false
}
