/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package apifabclient

import (
	config "github.com/hyperledger/fabric-sdk-go/api/apiconfig" // TODO: Think about package hierarchy
	"github.com/hyperledger/fabric-sdk-go/api/apicryptosuite"
	txn "github.com/hyperledger/fabric-sdk-go/api/apitxn"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
)

// Resource is a client that provides access to fabric resources such as chaincode.
type Resource interface {
	Context
	client

	// TODO refactor into channel provider and remove from Resource (upcoming CR)
	NewChannel(name string) (Channel, error)
	Channel(name string) Channel
}

type client interface {
	CreateChannel(request CreateChannelRequest) (txn.TransactionID, error)
	InstallChaincode(request InstallChaincodeRequest) ([]*txn.TransactionProposalResponse, string, error)
	QueryInstalledChaincodes(peer Peer) (*pb.ChaincodeQueryResponse, error)
	QueryChannels(peer Peer) (*pb.ChannelQueryResponse, error)

	ExtractChannelConfig(configEnvelope []byte) ([]byte, error)
	SignChannelConfig(config []byte, signer IdentityContext) (*common.ConfigSignature, error)
}

// Context supplies the configuration and signing identity to client objects.
type Context interface {
	SigningManager() SigningManager
	Config() config.Config
	CryptoSuite() apicryptosuite.CryptoSuite
	UserContext() IdentityContext
}

// FabricClient provides access to infrastructure functionality.
//
// Deprecated: this interface has been renamed.
/*
 * Main interaction handler with end user. A client instance provides a handler to interact
 * with a network of peers, orderers and optionally member services. An application using the
 * SDK may need to interact with multiple networks, each through a separate instance of the Client.
 *
 * Each client when initially created should be initialized with configuration data from the
 * consensus service, which includes a list of trusted roots, orderer certificates and IP addresses,
 * and a list of peer certificates and IP addresses that it can access. This must be done out of band
 * as part of bootstrapping the application environment. It is also the responsibility of the application
 * to maintain the configuration of a client as the SDK does not persist this object.
 *
 * Each Client instance can maintain several {@link Channel} instances representing channels and the associated
 * private ledgers.
 *
 *
 */
type FabricClient interface {
	Resource

	QueryChannelInfo(name string, peers []Peer) (Channel, error)

	SetUserContext(user IdentityContext)
	SaveUserToStateStore(user User) error
	LoadUserFromStateStore(name string) (User, error)
	StateStore() KeyValueStore
}

// CreateChannelRequest requests channel creation on the network
type CreateChannelRequest struct {
	// required - The name of the new channel
	Name string
	// required - The Orderer to send the update request
	Orderer Orderer
	// optional - the envelope object containing all
	// required settings and signatures to initialize this channel.
	// This envelope would have been created by the command
	// line tool "configtx"
	Envelope []byte
	// optional - ConfigUpdate object built by the
	// buildChannelConfig() method of this package
	Config []byte
	// optional - the list of collected signatures
	// required by the channel create policy when using the `config` parameter.
	// see signChannelConfig() method of this package
	Signatures []*common.ConfigSignature

	// TODO: InvokeChannelRequest allows the TransactionID to be passed in.
	// This request struct also has the field for consistency but perhaps it should be removed.
	TxnID txn.TransactionID
}

// InstallChaincodeRequest requests chaincode installation on the network
type InstallChaincodeRequest struct {
	// required - name of the chaincode
	Name string
	// required - path to the location of chaincode sources (path from GOPATH/src folder)
	Path string
	// chaincodeVersion: required - version of the chaincode
	Version string
	// required - package (chaincode package type and bytes)
	Package *CCPackage
	// required - proposal processor list
	Targets []txn.ProposalProcessor
}

// CCPackage contains package type and bytes required to create CDS
type CCPackage struct {
	Type pb.ChaincodeSpec_Type
	Code []byte
}
