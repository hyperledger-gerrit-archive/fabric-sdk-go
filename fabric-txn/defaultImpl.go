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

// GetDefaultImplClient ...
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

// GetDefaultImplClientWithUser ...
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

// GetDefaultImplUser ...
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

// GetDefaultImplPreEnrolledUser ...
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

// GetDefaultImplChannel ...
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

// GetDefaultImplEventHub ...
func GetDefaultImplEventHub(client api.FabricClient) (api.EventHub, error) {
	return eventsImpl.NewEventHub(client)
}

// GetDefaultImplOrderer ...
func GetDefaultImplOrderer(url string, certificate string, serverHostOverride string, config api.Config) (api.Orderer, error) {
	return ordererImpl.NewOrderer(url, certificate, serverHostOverride, config)
}

// GetDefaultImplPeer ...
func GetDefaultImplPeer(url string, certificate string, serverHostOverride string, config api.Config) (api.Peer, error) {
	return peerImpl.NewPeerTLSFromCert(url, certificate, serverHostOverride, config)
}

// GetDefaultImplConfig ...
func GetDefaultImplConfig(configFile string) (api.Config, error) {
	return configImpl.InitConfig(configFile)
}

// GetDefaultImplMspClient ...
func GetDefaultImplMspClient(config api.Config) (api.Services, error) {
	mspClient, err := fabricCAClient.NewFabricCAClient(config)
	if err != nil {
		return nil, fmt.Errorf("NewFabricCAClient returned error: %v", err)
	}

	return mspClient, nil
}
