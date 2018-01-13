/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package orderer

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	ab "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/protos/orderer"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"

	"crypto/x509"

	"github.com/hyperledger/fabric-sdk-go/pkg/config/comm"
	"github.com/hyperledger/fabric-sdk-go/pkg/config/urlutil"
	"github.com/hyperledger/fabric-sdk-go/pkg/errors"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging"
	"google.golang.org/grpc/credentials"
)

var logger = logging.NewLogger("fabric_sdk_go")

// Orderer allows a client to broadcast a transaction.
type Orderer struct {
	config         apiconfig.Config
	url            string
	tlsCACert      *x509.Certificate
	serverName     string
	grpcDialOption []grpc.DialOption
}

// New Returns a Orderer instance
func New(config apiconfig.Config, options ...func(*Orderer) error) (*Orderer, error) {
	orderer := &Orderer{config: config}

	err := applyOptions(orderer, options...)

	if err != nil {
		return nil, err
	}

	grpcOpts := append([]grpc.DialOption{}, grpc.WithTimeout(config.TimeoutOrDefault(apiconfig.OrdererConnection)))

	if urlutil.IsTLSEnabled(orderer.url) {
		tlsConfig, err := comm.TLSConfig(orderer.tlsCACert, orderer.serverName, config)

		if err != nil {
			return nil, err
		}

		grpcOpts = append(grpcOpts, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	} else {
		grpcOpts = append(grpcOpts, grpc.WithInsecure())
	}

	orderer.url = urlutil.ToAddress(orderer.url)
	orderer.grpcDialOption = grpcOpts

	return orderer, nil
}

func applyOptions(orderer *Orderer, options ...func(*Orderer) error) error {
	for _, option := range options {
		err := option(orderer)

		if err != nil {
			return err
		}
	}

	return nil
}

// WithURL is a functional option for the New constructor.
func WithURL(url string) func(*Orderer) error {
	return func(o *Orderer) error {
		o.url = url

		return nil
	}
}

// WithTLSCert is a functional option for the New constructor.
func WithTLSCert(tlsCACert *x509.Certificate) func(*Orderer) error {
	return func(o *Orderer) error {
		o.tlsCACert = tlsCACert

		return nil
	}
}

// WithServerName is a functional option for the New constructor.
func WithServerName(serverName string) func(*Orderer) error {
	return func(o *Orderer) error {
		o.serverName = serverName

		return nil
	}
}

// FromOrdererConfig is a functional option for the New constructor.
func FromOrdererConfig(ordererCfg *apiconfig.OrdererConfig) func(*Orderer) error {
	return func(o *Orderer) error {
		o.url = ordererCfg.URL

		var err error

		if urlutil.IsTLSEnabled(ordererCfg.URL) {
			o.tlsCACert, err = ordererCfg.TLSCACerts.TLSCert()

			if err != nil {
				return err
			}
		}

		o.serverName = getServerNameOverride(ordererCfg)

		return nil
	}
}

// FromOrdererName is a functional option for the New constructor.
func FromOrdererName(name string) func(*Orderer) error {
	return func(o *Orderer) error {
		ordererCfg, err := o.config.OrdererConfig(name)

		if err != nil {
			return err
		}

		return FromOrdererConfig(ordererCfg)(o)
	}
}

func getServerNameOverride(ordererCfg *apiconfig.OrdererConfig) string {
	serverNameOverride := ""
	if str, ok := ordererCfg.GRPCOptions["ssl-target-name-override"].(string); ok {
		serverNameOverride = str
	}
	return serverNameOverride
}

// URL Get the Orderer url. Required property for the instance objects.
// Returns the address of the Orderer.
func (o *Orderer) URL() string {
	return o.url
}

// SendBroadcast Send the created transaction to Orderer.
func (o *Orderer) SendBroadcast(envelope *fab.SignedEnvelope) (*common.Status, error) {
	conn, err := grpc.Dial(o.url, o.grpcDialOption...)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	broadcastStream, err := ab.NewAtomicBroadcastClient(conn).Broadcast(context.Background())
	if err != nil {
		return nil, errors.Wrap(err, "NewAtomicBroadcastClient failed")
	}
	done := make(chan bool)
	var broadcastErr error
	var broadcastStatus *common.Status

	go func() {
		for {
			broadcastResponse, err := broadcastStream.Recv()
			logger.Debugf("Orderer.broadcastStream - response:%v, error:%v\n", broadcastResponse, err)
			if err != nil {
				broadcastErr = errors.Wrap(err, "broadcast recv failed")
				done <- true
				return
			}
			broadcastStatus = &broadcastResponse.Status
			if broadcastResponse.Status == common.Status_SUCCESS {
				done <- true
				return
			}
			if broadcastResponse.Status != common.Status_SUCCESS {
				broadcastErr = errors.Errorf("broadcast response is not success %v", broadcastResponse.Status)
				done <- true
				return
			}
		}
	}()
	if err := broadcastStream.Send(&common.Envelope{
		Payload:   envelope.Payload,
		Signature: envelope.Signature,
	}); err != nil {
		return nil, errors.Wrap(err, "failed to send envelope to orderer")
	}
	broadcastStream.CloseSend()
	<-done
	return broadcastStatus, broadcastErr
}

// SendDeliver sends a deliver request to the ordering service and returns the
// blocks requested
// envelope: contains the seek request for blocks
func (o *Orderer) SendDeliver(envelope *fab.SignedEnvelope) (chan *common.Block,
	chan error) {
	responses := make(chan *common.Block)
	errs := make(chan error, 1)
	// Validate envelope
	if envelope == nil {
		errs <- errors.New("envelope is nil")
		return responses, errs
	}
	// Establish connection to Ordering Service
	conn, err := grpc.Dial(o.url, o.grpcDialOption...)
	if err != nil {
		errs <- err
		return responses, errs
	}
	// Create atomic broadcast client
	broadcastStream, err := ab.NewAtomicBroadcastClient(conn).
		Deliver(context.Background())
	if err != nil {
		errs <- errors.Wrap(err, "NewAtomicBroadcastClient failed")
		return responses, errs
	}
	// Send block request envolope
	logger.Debugf("Requesting blocks from ordering service")
	if err := broadcastStream.Send(&common.Envelope{
		Payload:   envelope.Payload,
		Signature: envelope.Signature,
	}); err != nil {
		errs <- errors.Wrap(err, "failed to send block request to orderer")
		return responses, errs
	}
	// Receive blocks from the GRPC stream and put them on the channel
	go func() {
		defer conn.Close()
		for {
			response, err := broadcastStream.Recv()
			if err != nil {
				errs <- errors.Wrap(err, "recv from ordering service failed")
				return
			}
			// Assert response type
			switch t := response.Type.(type) {
			// Seek operation success, no more resposes
			case *ab.DeliverResponse_Status:
				if t.Status == common.Status_SUCCESS {
					close(responses)
					return
				}
				if t.Status != common.Status_SUCCESS {
					errs <- errors.Errorf("error status from ordering service %s",
						t.Status)
					return
				}

			// Response is a requested block
			case *ab.DeliverResponse_Block:
				logger.Debug("Received block from ordering service")
				responses <- response.GetBlock()
			// Unknown response
			default:
				errs <- errors.Errorf("unknown response from ordering service %s", t)
				return
			}
		}
	}()

	return responses, errs
}
