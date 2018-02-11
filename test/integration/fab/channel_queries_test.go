/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fab

/*
func TestChannelQueries(t *testing.T) {

	testQueryConfigBlock(t, channel)

	testQueryChannels(t, channel, client)

	testInstalledChaincodes(t, channel, client)

	testQueryByChaincode(t, sdk.Config(), channel)
}
*/

/*
func testQueryConfigBlock(t *testing.T, channel fab.Channel, targets []fab.ProposalProcessor) {

	// Retrieve current channel configuration
	cfgEnvelope, err := channel.QueryConfigBlock(targets, 1)
	if err != nil {
		t.Fatalf("QueryConfig return error: %v", err)
	}

	if cfgEnvelope.Config == nil {
		t.Fatalf("QueryConfig config data is nil")
	}

}*/
/*
func testQueryChannels(t *testing.T, ledger fab.ChannelLedger, client fab.Resource, targets []fab.ProposalProcessor) {

	// Our target will be primary peer on this channel
	target := channel.PrimaryPeer()
	t.Logf("****QueryChannels for %s", target.URL())
	channelQueryResponse, err := client.QueryChannels(target)
	if err != nil {
		t.Fatalf("QueryChannels return error: %v", err)
	}

	for _, channel := range channelQueryResponse.Channels {
		t.Logf("**Channel: %s", channel)
	}

}
*/
/*
func testInstalledChaincodes(t *testing.T, channel fab.Channel, client fab.Resource, targets []fab.ProposalProcessor) {

	// Our target will be primary peer on this channel
	target := channel.PrimaryPeer()
	t.Logf("****QueryInstalledChaincodes for %s", target.URL())

	chaincodeQueryResponse, err := client.QueryInstalledChaincodes(target)
	if err != nil {
		t.Fatalf("QueryInstalledChaincodes return error: %v", err)
	}

	for _, chaincode := range chaincodeQueryResponse.Chaincodes {
		t.Logf("**InstalledCC: %s", chaincode)
	}

}
*/

/*
func testQueryByChaincode(t *testing.T, config apiconfig.Config, channel fab.Channel, targets []fab.ProposalProcessor) {

	request := fab.ChaincodeInvokeRequest{
		Targets:     targets,
		ChaincodeID: "lscc",
		Fcn:         "getinstalledchaincodes",
	}
	queryResponses, err := channel.QueryBySystemChaincode(request)
	if err != nil {
		t.Fatalf("QueryByChaincode failed %s", err)
	}

	// Number of responses should be the same as number of targets
	if len(queryResponses) != len(targets) {
		t.Fatalf("QueryByChaincode number of results mismatch. Expected: %d Got: %d", len(targets), len(queryResponses))
	}

	// Configured cert for cert pool
	certPath, err := config.CAClientCertPath(org1Name)

	if err != nil {
		t.Fatal(err)
	}

	certConfig := apiconfig.TLSConfig{Path: certPath}

	cert, err := certConfig.TLSCert()

	if err != nil {
		t.Fatal(err)
	}

	// Create invalid target
	firstInvalidTarget, err := peer.New(config, peer.WithURL("test:1111"), peer.WithTLSCert(cert))
	if err != nil {
		t.Fatalf("Create NewPeer error(%v)", err)
	}

	// Create second invalid target
	secondInvalidTarget, err := peer.New(config, peer.WithURL("test:2222"), peer.WithTLSCert(cert))
	if err != nil {
		t.Fatalf("Create NewPeer error(%v)", err)
	}

	// Add invalid targets to targets
	invalidTargets := append(targets, firstInvalidTarget)
	invalidTargets = append(invalidTargets, secondInvalidTarget)

	// Add invalid targets to channel otherwise validation will fail
	err = channel.AddPeer(firstInvalidTarget)
	if err != nil {
		t.Fatalf("Error adding peer: %v", err)
	}
	err = channel.AddPeer(secondInvalidTarget)
	if err != nil {
		t.Fatalf("Error adding peer: %v", err)
	}

	// Test valid + invalid targets
	request = fab.ChaincodeInvokeRequest{
		ChaincodeID: "lscc",
		Fcn:         "getinstalledchaincodes",
		Targets:     invalidTargets,
	}
	queryResponses, err = channel.QueryBySystemChaincode(request)
	if err == nil {
		t.Fatalf("QueryByChaincode failed to return error for non-existing target")
	}

	// Verify that valid targets returned response
	if len(queryResponses) != len(targets) {
		t.Fatalf("QueryByChaincode number of results mismatch. Expected: %d Got: %d (and error %v)", len(targets), len(queryResponses), err)
	}

	channel.RemovePeer(firstInvalidTarget)
	channel.RemovePeer(secondInvalidTarget)
}
*/
