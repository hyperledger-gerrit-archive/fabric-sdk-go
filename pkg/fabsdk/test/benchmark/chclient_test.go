/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package benchmark

import (
	"fmt"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	contextImpl "github.com/hyperledger/fabric-sdk-go/pkg/context"
	fcmocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/provider/chpvdr"
	"github.com/hyperledger/fabric-sdk-go/pkg/util/pathvar"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/common/cauthdsl"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/core/common/ccprovider"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/testdata"
)

const (
	channelID         = "myChannel"
	peerTLSServerCert = "${GOPATH}/src/github.com/hyperledger/fabric-sdk-go/test/fixtures/fabric/v1/crypto-config/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/server.crt"
	peerTLSServerKey  = "${GOPATH}/src/github.com/hyperledger/fabric-sdk-go/test/fixtures/fabric/v1/crypto-config/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/server.key"

	ordererTLSServerCert = "${GOPATH}/src/github.com/hyperledger/fabric-sdk-go/test/fixtures/fabric/v1/crypto-config/ordererOrganizations/example.com/orderers/orderer.example.com/tls/server.crt"
	ordererTLSServerKey  = "${GOPATH}/src/github.com/hyperledger/fabric-sdk-go/test/fixtures/fabric/v1/crypto-config/ordererOrganizations/example.com/orderers/orderer.example.com/tls/server.key"

	testhost          = "peer0.org1.example.com"
	testport          = 7051
	testBroadcastport = 7050
)

func BenchmarkExecuteTx(b *testing.B) {
	// report memory allocations for this benchmark
	b.ReportAllocs()

	tlsServerCertFile := testdata.Path(pathvar.Subst(peerTLSServerCert))
	tlsServerKeyFile := testdata.Path(pathvar.Subst(peerTLSServerKey))

	creds, err := credentials.NewServerTLSFromFile(tlsServerCertFile, tlsServerKeyFile)
	if err != nil {
		b.Fatalf("Failed to create new peer tls creds from file: %s", err)
	}
	payloadMap := make(map[string][]byte, 2)
	payloadMap["GetConfigBlock"] = getConfigBlockPayload()
	payloadMap["getccdata"] = getCCDataPayload()
	payloadMap["invoke"] = []byte("moved 'b' bytes")
	payloadMap["default"] = []byte("value")

	// setup mocked peer
	mockEndorserServer := &MockEndorserServer{Creds: creds}
	mockEndorserServer.SetMockPeer(&MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil, MockMSP: "Org1MSP", Status: 200,
		Payload: payloadMap})
	fmt.Println("***************** Mocked Peer Started: ", mockEndorserServer.Start(fmt.Sprintf("%s:%d", testhost, testport)), " ******************************")
	defer mockEndorserServer.Stop()

	// setup mocked CA
	f := testFixture{}
	_, ctx := f.setup()
	defer f.close()

	// setup mocked broadcast server with tls credentials (mocking orderer requests)
	tlsServerCertFile = testdata.Path(pathvar.Subst(ordererTLSServerCert))
	tlsServerKeyFile = testdata.Path(pathvar.Subst(ordererTLSServerKey))

	creds, err = credentials.NewServerTLSFromFile(tlsServerCertFile, tlsServerKeyFile)
	if err != nil {
		b.Fatalf("Failed to create new orderer tls creds from file: %s", err)
	}
	ordererMockSrv := &fcmocks.MockBroadcastServer{Creds: creds}
	fmt.Println("***************** Mocked Orderer Started: ", ordererMockSrv.Start(fmt.Sprintf("%s:%d", testhost, testBroadcastport)), " ******************************")
	defer ordererMockSrv.Stop()

	chClient := setupChannelClient(b, f.endpointConfig, ctx)

	// using channel Client, let's start the benchmark
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		chClient.Execute(channel.Request{ChaincodeID: "testCC", Fcn: "invoke", Args: [][]byte{[]byte("move"), []byte("b")}})
		//assert.NoError(b, err, "expected no error for valid channel client invoke")

		//b.Logf("Execute Responses: %s", resp.Responses)
	}
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

	payload, err := proto.Marshal(builder.Build())
	if err != nil {
		fmt.Println("Error marching ConfigBlockPayload: ", err)
	}

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

func setupChannelClient(b *testing.B, endpointConfig fab.EndpointConfig, ctx context.Client) *channel.Client {

	clntPvdr := setupCustomTestContext(b, endpointConfig, ctx)

	chPvdr := createChannelContext(clntPvdr, channelID)

	ch, err := channel.New(chPvdr)

	if err != nil {
		b.Fatalf("Failed to create new channel client: %s", err)
	}

	return ch
}

func setupCustomTestContext(b *testing.B, endpointConfig fab.EndpointConfig, ctx context.Client) context.ClientProvider {
	_, err := setupTestChannelService(ctx, endpointConfig)
	assert.Nil(b, err, "Got error %s", err)

	return createClientContext(ctx)
}

func setupTestChannelService(ctx context.Client, endpointConfig fab.EndpointConfig) (fab.ChannelService, error) {
	chProvider, err := chpvdr.New(endpointConfig)
	if err != nil {
		return nil, errors.WithMessage(err, "channel provider creation failed")
	}

	chService, err := chProvider.ChannelService(ctx, channelID)
	if err != nil {
		return nil, errors.WithMessage(err, "channel service creation failed")
	}

	return chService, nil
}

func createChannelContext(clientContext context.ClientProvider, channelID string) context.ChannelProvider {

	channelProvider := func() (context.Channel, error) {
		return contextImpl.NewChannel(clientContext, channelID)
	}

	return channelProvider
}

func createClientContext(client context.Client) context.ClientProvider {
	return func() (context.Client, error) {
		return client, nil
	}
}

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
