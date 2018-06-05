/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package benchmark

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/asn1"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"sync"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/protos/discovery"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	discmocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/discovery/mocks"
	eventmocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/events/mocks"
	fcmocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/util/test"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/common/cauthdsl"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/core/common/ccprovider"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/core/ledger/kvledger/txmgmt/rwsetutil"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/ledger/rwset/kvrwset"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/msp"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	testhost          = "peer0.org1.example.com"
	testport          = 7051
	testBroadcastport = 7050
)

// MockEndorserServer mock endorser server to process endorsement proposals
type MockEndorserServer struct {
	mockPeer      *MockPeer
	ProposalError error
	AddkvWrite    bool
	Creds         credentials.TransportCredentials
	srv           *grpc.Server
	wg            sync.WaitGroup
}

// ProcessProposal mock implementation that returns success (through mockPeer) if error is not set
// error if it is
func (m *MockEndorserServer) ProcessProposal(context context.Context,
	proposal *pb.SignedProposal) (*pb.ProposalResponse, error) {
	fcn, err := m.getFuncNameFromProposal(proposal)
	if err != nil {
		return nil, errors.Wrap(err, "Error getting function name from mockPeer")
	}
	if m.ProposalError == nil {

		tp, err := m.GetMockPeer().ProcessTransactionProposal(fab.TransactionProposal{}, fcn)
		if err != nil {
			return &pb.ProposalResponse{Response: &pb.Response{
				Status:  500,
				Message: err.Error(),
			}}, err
		}
		return tp.ProposalResponse, nil

	}

	return &pb.ProposalResponse{Response: &pb.Response{
		Status:  500,
		Message: m.ProposalError.Error(),
	}}, m.ProposalError
}

func (m *MockEndorserServer) getFuncNameFromProposal(proposal *pb.SignedProposal) ([]byte, error) {
	pr := &pb.Proposal{}
	err := proto.Unmarshal(proposal.GetProposalBytes(), pr)
	if err != nil {
		return nil, err
	}
	cpp := &pb.ChaincodeProposalPayload{}
	err = proto.Unmarshal(pr.Payload, cpp)
	if err != nil {
		return nil, err
	}

	cic := &pb.ChaincodeInvocationSpec{}
	err = proto.Unmarshal(cpp.Input, cic)
	if err != nil {
		return nil, err
	}
	return cic.ChaincodeSpec.Input.Args[0], nil
}

func (m *MockEndorserServer) createProposalResponsePayload() []byte {

	prp := &pb.ProposalResponsePayload{}
	ccAction := &pb.ChaincodeAction{}
	txRwSet := &rwsetutil.TxRwSet{}

	if m.AddkvWrite {
		txRwSet.NsRwSets = []*rwsetutil.NsRwSet{
			{NameSpace: "ns1", KvRwSet: &kvrwset.KVRWSet{
				Reads:  []*kvrwset.KVRead{{Key: "key1", Version: &kvrwset.Version{BlockNum: 1, TxNum: 1}}},
				Writes: []*kvrwset.KVWrite{{Key: "key2", IsDelete: false, Value: []byte("value2")}},
			}}}
	}

	txRWSetBytes, err := txRwSet.ToProtoBytes()
	if err != nil {
		return nil
	}
	ccAction.Results = txRWSetBytes
	ccActionBytes, err := proto.Marshal(ccAction)
	if err != nil {
		return nil
	}
	prp.Extension = ccActionBytes
	prpBytes, err := proto.Marshal(prp)
	if err != nil {
		return nil
	}
	return prpBytes
}

// Start the mock endorser server
func (m *MockEndorserServer) Start(address string) string {
	if m.srv != nil {
		panic("MockEndorserServer already started")
	}

	// pass in TLS creds if present
	if m.Creds != nil {
		m.srv = grpc.NewServer(grpc.Creds(m.Creds))
	} else {
		m.srv = grpc.NewServer()
	}

	lis, err := net.Listen("tcp", address)
	if err != nil {
		panic(fmt.Sprintf("Error starting EndorserServer %s", err))
	}
	addr := lis.Addr().String()

	test.Logf("Starting MockEndorserServer [%s]", addr)
	pb.RegisterEndorserServer(m.srv, m)

	m.registerDiscoveryAndDeliveryServers(address)

	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		if err := m.srv.Serve(lis); err != nil {
			test.Logf("StartEndorserServer failed [%s]", err)
		}
	}()

	return addr
}

// Stop the mock broadcast server and wait for completion.
func (m *MockEndorserServer) Stop() {
	if m.srv == nil {
		panic("MockEndorserServer not started")
	}

	m.srv.Stop()
	m.wg.Wait()
	m.srv = nil
}

// GetMockPeer will return the mock endorser's mock peer in a thread safe way
func (m *MockEndorserServer) GetMockPeer() *MockPeer {
	var v = func() *MockPeer {
		m.wg.Add(1)
		defer m.wg.Done()
		return m.mockPeer
	}

	return v()
}

// SetMockPeer will write the mock endorser's mock peer in a thread safe way
func (m *MockEndorserServer) SetMockPeer(mPeer *MockPeer) {
	func(p *MockPeer) {
		m.wg.Add(1)
		defer m.wg.Done()
		m.mockPeer = p
	}(mPeer)
}

func (m *MockEndorserServer) registerDiscoveryAndDeliveryServers(peerAddress string) {
	//register DiscoverService and DeliveryService
	discoverServer := discmocks.NewServer(
		discmocks.WithLocalPeers(
			&discmocks.MockDiscoveryPeerEndpoint{
				MSPID:        "Org1MSP",
				Endpoint:     peerAddress,
				LedgerHeight: 26,
			},
		),
		discmocks.WithPeers(
			&discmocks.MockDiscoveryPeerEndpoint{
				MSPID:        "Org1MSP",
				Endpoint:     peerAddress,
				LedgerHeight: 26,
			},
		),
	)
	deliverServer := eventmocks.NewMockDeliverServer()

	discovery.RegisterDiscoveryServer(m.srv, discoverServer)
	pb.RegisterDeliverServer(m.srv, deliverServer)
}

// MockPeer is a mock fabricsdk.Peer.
type MockPeer struct {
	RWLock               *sync.RWMutex
	Error                error
	MockName             string
	MockURL              string
	MockRoles            []string
	MockCert             *pem.Block
	Payload              map[string][]byte
	ResponseMessage      string
	MockMSP              string
	Status               int32
	ProcessProposalCalls int
	Endorser             []byte
	KVWrite              bool
}

// NewMockPeer creates basic mock peer
func NewMockPeer(name string, url string) *MockPeer {
	mp := &MockPeer{MockName: name, MockURL: url, Status: 200, RWLock: &sync.RWMutex{}}
	return mp
}

// Name returns the mock peer's mock name
func (p MockPeer) Name() string {
	return p.MockName
}

// SetName sets the mock peer's mock name
func (p *MockPeer) SetName(name string) {
	p.MockName = name
}

// MSPID gets the Peer mspID.
func (p *MockPeer) MSPID() string {
	return p.MockMSP
}

// SetMSPID sets the Peer mspID.
func (p *MockPeer) SetMSPID(mspID string) {
	p.MockMSP = mspID
}

// Roles returns the mock peer's mock roles
func (p *MockPeer) Roles() []string {
	return p.MockRoles
}

// SetRoles sets the mock peer's mock roles
func (p *MockPeer) SetRoles(roles []string) {
	p.MockRoles = roles
}

// EnrollmentCertificate returns the mock peer's mock enrollment certificate
func (p *MockPeer) EnrollmentCertificate() *pem.Block {
	return p.MockCert
}

// SetEnrollmentCertificate sets the mock peer's mock enrollment certificate
func (p *MockPeer) SetEnrollmentCertificate(pem *pem.Block) {
	p.MockCert = pem
}

// URL returns the mock peer's mock URL
func (p *MockPeer) URL() string {
	return p.MockURL
}

// ProcessTransactionProposal does not send anything anywhere but returns an empty mock ProposalResponse
func (p *MockPeer) ProcessTransactionProposal(tp fab.TransactionProposal, funcName []byte) (*fab.TransactionProposalResponse, error) {
	if p.RWLock != nil {
		p.RWLock.Lock()
		defer p.RWLock.Unlock()
	}
	p.ProcessProposalCalls++

	if p.Endorser == nil {
		// We serialize identities by prepending the MSPID and appending the ASN.1 DER content of the cert
		sID := &msp.SerializedIdentity{Mspid: "Org1MSP", IdBytes: []byte(CertPem)}
		endorser, err := proto.Marshal(sID)
		if err != nil {
			return nil, err
		}
		p.Endorser = endorser
	}

	block, _ := pem.Decode(KeyPem)
	lowLevelKey, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		return nil, errors.Wrap(err, "Error received while parsing EC Key")
	}
	proposalResponsePayload, err := p.createProposalResponsePayload()
	if err != nil {
		return nil, errors.Wrap(err, "Error received while creating proposal response")
	}
	sigma, err := SignECDSA(lowLevelKey, append(proposalResponsePayload, p.Endorser...))
	if err != nil {
		return nil, errors.Wrap(err, "Error received while signing proposal for endorser with EC key")
	}

	payload, ok := p.Payload[string(funcName)]
	if !ok {
		payload, ok = p.Payload[string("default")]
		if !ok {
			fmt.Printf("payload for func(%s) not found\n", funcName)
		}
	}
	return &fab.TransactionProposalResponse{
		Endorser: p.MockURL,
		Status:   p.Status,
		ProposalResponse: &pb.ProposalResponse{Response: &pb.Response{
			Message: p.ResponseMessage, Status: p.Status, Payload: payload}, Payload: proposalResponsePayload,
			Endorsement: &pb.Endorsement{Endorser: p.Endorser, Signature: sigma}},
	}, p.Error

}

func (p *MockPeer) createProposalResponsePayload() ([]byte, error) {

	prp := &pb.ProposalResponsePayload{}
	ccAction := &pb.ChaincodeAction{}
	txRwSet := &rwsetutil.TxRwSet{}
	var kvWrite []*kvrwset.KVWrite
	if p.KVWrite {
		kvWrite = []*kvrwset.KVWrite{&kvrwset.KVWrite{Key: "key2", IsDelete: false, Value: []byte("value2")}}
	}
	txRwSet.NsRwSets = []*rwsetutil.NsRwSet{
		&rwsetutil.NsRwSet{NameSpace: "ns1", KvRwSet: &kvrwset.KVRWSet{
			Reads:  []*kvrwset.KVRead{&kvrwset.KVRead{Key: "key1", Version: &kvrwset.Version{BlockNum: 1, TxNum: 1}}},
			Writes: kvWrite,
		}}}

	txRWSetBytes, err := txRwSet.ToProtoBytes()
	if err != nil {
		return nil, err
	}

	ccAction.Results = txRWSetBytes
	ccActionBytes, err := proto.Marshal(ccAction)
	if err != nil {
		return nil, err
	}
	prp.Extension = ccActionBytes
	prpBytes, err := proto.Marshal(prp)
	if err != nil {
		return nil, err
	}
	return prpBytes, nil
}

// SignECDSA sign with ec key
func SignECDSA(k *ecdsa.PrivateKey, digest []byte) (signature []byte, err error) {
	hash := sha256.New()
	hash.Write(digest)

	r, s, err := ecdsa.Sign(rand.Reader, k, hash.Sum(nil))
	if err != nil {
		return nil, err
	}

	s, _, err = ToLowS(&k.PublicKey, s)
	if err != nil {
		return nil, err
	}

	return MarshalECDSASignature(r, s)
}

func getConfigBlockPayload() []byte {
	// create config block builder in order to create valid payload
	builder := &fcmocks.MockConfigBlockBuilder{
		MockConfigGroupBuilder: fcmocks.MockConfigGroupBuilder{
			ModPolicy: "Admins",
			MSPNames: []string{
				"Org1MSP",
			},
			OrdererAddress:          fmt.Sprintf("grpc://%s:%d", testhost, testBroadcastport),
			RootCA:                  rootCA,
			ApplicationCapabilities: []string{fab.V1_1Capability},
		},
		Index:           0,
		LastConfigIndex: 0,
	}

	payload, _ := proto.Marshal(builder.Build())

	return payload
}

func getCCDataPayload() []byte {
	ccPolicy := cauthdsl.SignedByMspMember("Org1MSP")
	pp, err := proto.Marshal(ccPolicy)
	if err != nil {
		panic(fmt.Sprintf("failed to build mock CC Policy: %s", err))
	}

	ccData := &ccprovider.ChaincodeData{
		Name:   "lscc",
		Policy: pp,
	}

	pd, err := proto.Marshal(ccData)
	if err != nil {
		panic(fmt.Sprintf("failed to build mock CC Data: %s", err))
	}

	return pd
}

// CertPem certificate
var CertPem = `-----BEGIN CERTIFICATE-----
MIICCjCCAbGgAwIBAgIQOcq9Om9VwUe9hGN0TTGw1DAKBggqhkjOPQQDAjBYMQsw
CQYDVQQGEwJVUzETMBEGA1UECBMKQ2FsaWZvcm5pYTEWMBQGA1UEBxMNU2FuIEZy
YW5jaXNjbzENMAsGA1UEChMET3JnMTENMAsGA1UEAxMET3JnMTAeFw0xNzA1MDgw
OTMwMzRaFw0yNzA1MDYwOTMwMzRaMGUxCzAJBgNVBAYTAlVTMRMwEQYDVQQIEwpD
YWxpZm9ybmlhMRYwFAYDVQQHEw1TYW4gRnJhbmNpc2NvMRUwEwYDVQQKEwxPcmcx
LXNlcnZlcjExEjAQBgNVBAMTCWxvY2FsaG9zdDBZMBMGByqGSM49AgEGCCqGSM49
AwEHA0IABAm+2CZhbmsnA+HKQynXKz7fVZvvwlv/DdNg3Mdg7lIcP2z0b07/eAZ5
0chdJNcjNAd/QAj/mmGG4dObeo4oTKGjUDBOMA4GA1UdDwEB/wQEAwIFoDAdBgNV
HSUEFjAUBggrBgEFBQcDAQYIKwYBBQUHAwIwDAYDVR0TAQH/BAIwADAPBgNVHSME
CDAGgAQBAgMEMAoGCCqGSM49BAMCA0cAMEQCIG55RvN4Boa0WS9UcIb/tI2YrAT8
EZd/oNnZYlbxxyvdAiB6sU9xAn4oYIW9xtrrOISv3YRg8rkCEATsagQfH8SiLg==
-----END CERTIFICATE-----`

// KeyPem ec private key
var KeyPem = []byte(`-----BEGIN EC PRIVATE KEY-----
MHcCAQEEICfXQtVmdQAlp/l9umWJqCXNTDurmciDNmGHPpxHwUK/oAoGCCqGSM49
AwEHoUQDQgAECb7YJmFuaycD4cpDKdcrPt9Vm+/CW/8N02Dcx2DuUhw/bPRvTv94
BnnRyF0k1yM0B39ACP+aYYbh05t6jihMoQ==
-----END EC PRIVATE KEY-----`)

// RootCA ca
var rootCA = `-----BEGIN CERTIFICATE-----
MIIB8TCCAZegAwIBAgIQU59imQ+xl+FmwuiFyUgFezAKBggqhkjOPQQDAjBYMQsw
CQYDVQQGEwJVUzETMBEGA1UECBMKQ2FsaWZvcm5pYTEWMBQGA1UEBxMNU2FuIEZy
YW5jaXNjbzENMAsGA1UEChMET3JnMTENMAsGA1UEAxMET3JnMTAeFw0xNzA1MDgw
OTMwMzRaFw0yNzA1MDYwOTMwMzRaMFgxCzAJBgNVBAYTAlVTMRMwEQYDVQQIEwpD
YWxpZm9ybmlhMRYwFAYDVQQHEw1TYW4gRnJhbmNpc2NvMQ0wCwYDVQQKEwRPcmcx
MQ0wCwYDVQQDEwRPcmcxMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEFkpP6EqE
87ghFi25UWLvgPatxDiYKYaVSPvpo/XDJ0+9uUmK/C2r5Bvvxx1t8eTROwN77tEK
r+jbJIxX3ZYQMKNDMEEwDgYDVR0PAQH/BAQDAgGmMA8GA1UdJQQIMAYGBFUdJQAw
DwYDVR0TAQH/BAUwAwEB/zANBgNVHQ4EBgQEAQIDBDAKBggqhkjOPQQDAgNIADBF
AiEA1Xkrpq+wrmfVVuY12dJfMQlSx+v0Q3cYce9BE1i2mioCIAzqyduK/lHPI81b
nWiU9JF9dRQ69dEV9dxd/gzamfFU
-----END CERTIFICATE-----`

/****************************************************************************************************************************************/
// copied code from "github.com/hyperledger/fabric/bccsp/utils/ecdsa.go" to avoid importing Fabric dependency in this benchamark test
// IsLow checks that s is a low-S
func IsLowS(k *ecdsa.PublicKey, s *big.Int) (bool, error) {
	halfOrder, ok := curveHalfOrders[k.Curve]
	if !ok {
		return false, fmt.Errorf("curve not recognized [%s]", k.Curve)
	}

	return s.Cmp(halfOrder) != 1, nil

}

// ToLowS will config s to a low-S
func ToLowS(k *ecdsa.PublicKey, s *big.Int) (*big.Int, bool, error) {
	lowS, err := IsLowS(k, s)
	if err != nil {
		return nil, false, err
	}

	if !lowS {
		// Set s to N - s that will be then in the lower part of signature space
		// less or equal to half order
		s.Sub(k.Params().N, s)

		return s, true, nil
	}

	return s, false, nil
}

// curveHalfOrders contains the precomputed curve group orders halved.
// It is used to ensure that signature' S value is lower or equal to the
// curve group order halved. We accept only low-S signatures.
// They are precomputed for efficiency reasons.
var curveHalfOrders = map[elliptic.Curve]*big.Int{
	elliptic.P224(): new(big.Int).Rsh(elliptic.P224().Params().N, 1),
	elliptic.P256(): new(big.Int).Rsh(elliptic.P256().Params().N, 1),
	elliptic.P384(): new(big.Int).Rsh(elliptic.P384().Params().N, 1),
	elliptic.P521(): new(big.Int).Rsh(elliptic.P521().Params().N, 1),
}

// ECDSASignature struct
type ECDSASignature struct {
	R, S *big.Int
}

// MarshalECDSASignature will marshal the ECDSA signature
func MarshalECDSASignature(r, s *big.Int) ([]byte, error) {
	return asn1.Marshal(ECDSASignature{r, s})
}

// end of copy code from "github.com/hyperledger/fabric/bccsp/utils/ecdsa.go"
/****************************************************************************************************************************************/
