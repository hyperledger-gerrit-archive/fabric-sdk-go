/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package integration

import (
	"fmt"
	"math/rand"
	"path/filepath"
	"time"

	ca "github.com/hyperledger/fabric-sdk-go/api/apifabca"
	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	deffab "github.com/hyperledger/fabric-sdk-go/def/fabapi"
)

// GetOrdererAdmin returns a pre-enrolled orderer admin user
func GetOrdererAdmin(sdk *deffab.FabricSDK, orgID string) (ca.User, error) {

	// Orderer Admin Credentials
	privateKeyPath := filepath.Join(sdk.ConfigProvider().CryptoConfigPath(), "ordererOrganizations/example.com/users/Admin@example.com/msp/keystore/f4aa194b12d13d7c2b7b275a7115af5e6f728e11710716f2c754df4587891511_sk")
	enrollmentCertPath := filepath.Join(sdk.ConfigProvider().CryptoConfigPath(), "ordererOrganizations/example.com/users/Admin@example.com/msp/signcerts/Admin@example.com-cert.pem")

	credentialMgr, err := sdk.ContextFactory.NewCredentialManager(orgID, sdk.ConfigProvider(), sdk.CryptoSuiteProvider())
	if err != nil {
		return nil, fmt.Errorf("Error getting credential manager: %s ", err)
	}

	signingIdentity, err := credentialMgr.GetSigningIdentityFromPath(privateKeyPath, enrollmentCertPath)
	if err != nil {
		return nil, fmt.Errorf("Error getting signing identity: %s ", err)
	}

	user, err := deffab.NewPreEnrolledUser(sdk.ConfigProvider(), "ordererAdmin", signingIdentity)
	if err != nil {
		return nil, fmt.Errorf("NewUser returned error: %v", err)
	}

	return user, nil
}

// GenerateRandomID generates random ID
func GenerateRandomID() string {
	rand.Seed(time.Now().UnixNano())
	return randomString(10)
}

// Utility to create random string of strlen length
func randomString(strlen int) string {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, strlen)
	for i := 0; i < strlen; i++ {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
}

// HasPrimaryPeerJoinedChannel checks whether the primary peer of a channel
// has already joined the channel. It returns true if it has, false otherwise,
// or an error
func HasPrimaryPeerJoinedChannel(client fab.FabricClient, channel fab.Channel) (bool, error) {
	foundChannel := false
	primaryPeer := channel.PrimaryPeer()
	response, err := client.QueryChannels(primaryPeer)
	if err != nil {
		return false, fmt.Errorf("Error querying channel for primary peer: %s", err)
	}
	for _, responseChannel := range response.Channels {
		if responseChannel.ChannelId == channel.Name() {
			foundChannel = true
		}
	}

	return foundChannel, nil
}
