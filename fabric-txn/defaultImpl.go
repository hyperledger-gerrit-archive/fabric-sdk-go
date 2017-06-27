/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabrictxn

import (
	"fmt"
	"io/ioutil"

	fabricCaUtil "github.com/hyperledger/fabric-ca/util"
	api "github.com/hyperledger/fabric-sdk-go/api"
	configImpl "github.com/hyperledger/fabric-sdk-go/pkg/config"
	fabricCAClient "github.com/hyperledger/fabric-sdk-go/pkg/fabric-ca-client"
	clientImpl "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client"
	eventsImpl "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/events"
	kvs "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/keyvaluestore"
	ordererImpl "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/orderer"
	peerImpl "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/peer"
	userImpl "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/user"
	bccspFactory "github.com/hyperledger/fabric/bccsp/factory"
)

// GetDefaultImplClient returns a new default implementation of the Client interface using the config provided.
// It will save the provided user if requested into the state store.
func GetDefaultImplClient(user api.User, skipUserPersistence bool, stateStorePath string, config api.Config) (api.FabricClient, error) {
	client := clientImpl.NewClient(config)

	cryptoSuite := bccspFactory.GetDefault()

	client.SetCryptoSuite(cryptoSuite)
	stateStore, err := kvs.CreateNewFileKeyValueStore(stateStorePath)
	if err != nil {
		return nil, fmt.Errorf("CreateNewFileKeyValueStore returned error[%s]", err)
	}
	client.SetStateStore(stateStore)
	client.SaveUserToStateStore(user, skipUserPersistence)

	return client, nil
}

// GetDefaultImplClientWithUser returns a new default implementation of the Client interface.
// It creates a default implementation of User, enrolls the user, and saves it to the state store.
func GetDefaultImplClientWithUser(name string, pwd string, stateStorePath string, config api.Config, msp api.Services) (api.FabricClient, error) {
	client := clientImpl.NewClient(config)

	cryptoSuite := bccspFactory.GetDefault()

	client.SetCryptoSuite(cryptoSuite)
	stateStore, err := kvs.CreateNewFileKeyValueStore(stateStorePath)
	if err != nil {
		return nil, fmt.Errorf("CreateNewFileKeyValueStore returned error[%s]", err)
	}
	client.SetStateStore(stateStore)

	user, err := GetDefaultImplUser(client, msp, name, pwd)
	if err != nil {
		return nil, fmt.Errorf("GetDefaultImplUser returned error: %v", err)
	}
	client.SetUserContext(user)

	return client, nil
}

// GetDefaultImplClientWithPreEnrolledUser returns a new default Client implementation
// by using a the default implementation of a pre-enrolled user.
func GetDefaultImplClientWithPreEnrolledUser(config api.Config, stateStorePath string, skipUserPersistence bool, username string, keyDir string, certDir string) (api.FabricClient, error) {
	client := clientImpl.NewClient(config)

	cryptoSuite := bccspFactory.GetDefault()

	client.SetCryptoSuite(cryptoSuite)
	if stateStorePath != "" {
		stateStore, err := kvs.CreateNewFileKeyValueStore(stateStorePath)
		if err != nil {
			return nil, fmt.Errorf("CreateNewFileKeyValueStore returned error[%s]", err)
		}
		client.SetStateStore(stateStore)
	}
	user, err := GetDefaultImplPreEnrolledUser(client, keyDir, certDir, username)
	if err != nil {
		return nil, fmt.Errorf("GetDefaultImplPreEnrolledUser returned error: %v", err)
	}
	client.SetUserContext(user)
	client.SaveUserToStateStore(user, skipUserPersistence)

	return client, nil
}

// GetDefaultImplUser returns a new default implementation of a User.
func GetDefaultImplUser(client api.FabricClient, msp api.Services, name string, pwd string) (api.User, error) {
	user, err := client.LoadUserFromStateStore(name)
	if err != nil {
		return nil, fmt.Errorf("client.LoadUserFromStateStore returned error: %v", err)
	}

	if user == nil {
		key, cert, err := msp.Enroll(name, pwd)
		if err != nil {
			return nil, fmt.Errorf("Enroll returned error: %v", err)
		}
		user = userImpl.NewUser(name)
		user.SetPrivateKey(key)
		user.SetEnrollmentCertificate(cert)
		err = client.SaveUserToStateStore(user, false)
		if err != nil {
			return nil, fmt.Errorf("client.SaveUserToStateStore returned error: %v", err)
		}
	}

	return user, nil
}

// GetDefaultImplPreEnrolledUser returns a new default implementation of User.
// The user should already be pre-enrolled.
func GetDefaultImplPreEnrolledUser(client api.FabricClient, privateKeyPath string, enrollmentCertPath string, username string) (api.User, error) {

	privateKey, err := fabricCaUtil.ImportBCCSPKeyFromPEM(privateKeyPath, client.GetCryptoSuite(), true)
	if err != nil {
		return nil, fmt.Errorf("Error importing private key: %v", err)
	}
	enrollmentCert, err := ioutil.ReadFile(enrollmentCertPath)
	if err != nil {
		return nil, fmt.Errorf("Error reading from the enrollment cert path: %v", err)
	}

	user := userImpl.NewUser(username)
	user.SetEnrollmentCertificate(enrollmentCert)
	user.SetPrivateKey(privateKey)

	return user, nil
}

// GetDefaultImplChannel returns a new default implementation of Channel
func GetDefaultImplChannel(client api.FabricClient, orderer api.Orderer, peers []api.Peer, channelID string) (api.Channel, error) {

	channel, err := client.NewChannel(channelID)
	if err != nil {
		return nil, fmt.Errorf("NewChannel returned error: %v", err)
	}

	err = channel.AddOrderer(orderer)
	if err != nil {
		return nil, fmt.Errorf("Error adding orderer: %v", err)
	}

	for _, p := range peers {
		err = channel.AddPeer(p)
		if err != nil {
			return nil, fmt.Errorf("Error adding peer: %v", err)
		}
	}

	return channel, nil
}

// GetDefaultImplEventHub returns a new default implementation of Event Hub
func GetDefaultImplEventHub(client api.FabricClient) (api.EventHub, error) {
	return eventsImpl.NewEventHub(client)
}

// GetDefaultImplOrderer returns a new default implementation of Orderer
func GetDefaultImplOrderer(url string, certificate string, serverHostOverride string, config api.Config) (api.Orderer, error) {
	return ordererImpl.NewOrderer(url, certificate, serverHostOverride, config)
}

// GetDefaultImplPeer returns a new default implementation of Peer
func GetDefaultImplPeer(url string, certificate string, serverHostOverride string, config api.Config) (api.Peer, error) {
	return peerImpl.NewPeerTLSFromCert(url, certificate, serverHostOverride, config)
}

// GetDefaultImplConfig returns a new default implementation of the Config interface
func GetDefaultImplConfig(configFile string) (api.Config, error) {
	return configImpl.InitConfig(configFile)
}

// GetDefaultImplMspClient returns a new default implmentation of the MSP client
func GetDefaultImplMspClient(config api.Config) (api.Services, error) {
	mspClient, err := fabricCAClient.NewFabricCAClient(config)
	if err != nil {
		return nil, fmt.Errorf("NewFabricCAClient returned error: %v", err)
	}

	return mspClient, nil
}
