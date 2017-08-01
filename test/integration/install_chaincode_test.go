/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package integration

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	packager "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/packager"
	pb "github.com/hyperledger/fabric/protos/peer"
)

const (
	chainCodeIDGOLANG   = "golangcc"
	chainCodePathGOLANG = "github.com/example_cc"
	chainCodeIDBINARY   = "binarycc"
	chainCodePathBINARY = "../fixtures/src/github.com/example_cc_binary/example_cc"
)

var origGoPath = os.Getenv("GOPATH")

func TestChaincodeInstal(t *testing.T) {

	testSetup := &BaseSetupImpl{
		ConfigFile:      "../fixtures/config/config_test.yaml",
		ChannelID:       "mychannel",
		OrgID:           "peerorg1",
		ChannelConfig:   "../fixtures/channel/mychannel.tx",
		ConnectEventHub: true,
	}

	if err := testSetup.Initialize(); err != nil {
		t.Fatalf(err.Error())
	}

	//Test Install Chaincode scenario for GOLANG chaincodes
	fmt.Println("Testing install chaincode for GOLANG chaincode type")
	testChaincodeInstallUsingChaincodePath(t, testSetup, pb.ChaincodeSpec_GOLANG)
	testChaincodeInstallUsingChaincodePackage(t, testSetup, pb.ChaincodeSpec_GOLANG)
	fmt.Println("Testing install chaincode for GOLANG chaincode type is completed")

	//Test Install Chaincode scenario for BINARY chaincodes
	fmt.Println("Testing install chaincode for BINARY chaincode type")
	testChaincodeInstallUsingChaincodePath(t, testSetup, pb.ChaincodeSpec_BINARY)
	testChaincodeInstallUsingChaincodePackage(t, testSetup, pb.ChaincodeSpec_BINARY)
	fmt.Println("Testing install chaincode for BINARY chaincode type is completed")
}

// Test chaincode install using chaincodePath to create chaincodePackage
func testChaincodeInstallUsingChaincodePath(t *testing.T, testSetup *BaseSetupImpl, ccType pb.ChaincodeSpec_Type) {
	chainCodeVersion := getRandomCCVersion()

	// Install and Instantiate Events CC
	// Retrieve installed chaincodes
	client := testSetup.Client
	chainCodeID, chainCodePath := getChainCodeDetails(ccType)

	if err := testSetup.InstallCC(chainCodeID, chainCodePath, chainCodeVersion, nil, ccType); err != nil {
		t.Fatalf("installCC return error: %v", err)
	}

	// set Client User Context to Admin
	testSetup.Client.SetUserContext(testSetup.AdminUser)
	defer testSetup.Client.SetUserContext(testSetup.NormalUser)
	chaincodeQueryResponse, err := client.QueryInstalledChaincodes(testSetup.Channel.PrimaryPeer())
	if err != nil {
		t.Fatalf("QueryInstalledChaincodes return error: %v", err)
	}
	ccFound := false
	for _, chaincode := range chaincodeQueryResponse.Chaincodes {
		if chaincode.Name == chainCodeID && chaincode.Path == chainCodePath && chaincode.Version == chainCodeVersion {
			fmt.Printf("Found chaincode: %s\n", chaincode)
			ccFound = true
		}
	}

	if !ccFound {
		t.Fatalf("Failed to retrieve installed chaincode.")
	}
	//Install same chaincode again, should fail
	err = testSetup.InstallCC(chainCodeID, chainCodePath, chainCodeVersion, nil, ccType)
	if err == nil {
		t.Fatalf("install same chaincode didn't return error")
	}
	if strings.Contains(err.Error(), "chaincodes/install.v"+chainCodeVersion+" exists") {
		t.Fatalf("install same chaincode didn't return the correct error")
	}
}

// Test chaincode install using chaincodePackage[byte]
func testChaincodeInstallUsingChaincodePackage(t *testing.T, testSetup *BaseSetupImpl, ccType pb.ChaincodeSpec_Type) {

	chainCodeVersion := getRandomCCVersion()
	changeGOPATHToDeploy(testSetup.GetDeployPath())
	_, chainCodePath := getChainCodeDetails(ccType)
	chaincodePackage, err := packager.PackageCC(chainCodePath, ccType)
	resetGOPATH()
	if err != nil {
		t.Fatalf("PackageCC return error: %s", err)
	}

	err = testSetup.InstallCC("install", "github.com/example_cc_pkg", chainCodeVersion, chaincodePackage, ccType)
	if err != nil {
		t.Fatalf("installCC return error: %v", err)
	}
	//Install same chaincode again, should fail
	err = testSetup.InstallCC("install", chainCodePath, chainCodeVersion, chaincodePackage, ccType)
	if err == nil {
		t.Fatalf("install same chaincode didn't return error")
	}
	if strings.Contains(err.Error(), "chaincodes/install.v"+chainCodeVersion+" exists") {
		t.Fatalf("install same chaincode didn't return the correct error")
	}
}

func getRandomCCVersion() string {
	rand.Seed(time.Now().UnixNano())
	return "v0" + strconv.Itoa(rand.Intn(10000000))
}

// ChangeGOPATHToDeploy changes go path to fixtures folder
func changeGOPATHToDeploy(deployPath string) {
	os.Setenv("GOPATH", deployPath)
}

// ResetGOPATH resets go path to original
func resetGOPATH() {
	os.Setenv("GOPATH", origGoPath)
}

func getChainCodeDetails(ccType pb.ChaincodeSpec_Type) (string, string) {
	switch ccType {
	case pb.ChaincodeSpec_GOLANG:
		return chainCodeIDGOLANG, chainCodePathGOLANG
	case pb.ChaincodeSpec_BINARY:
		return chainCodeIDBINARY, chainCodePathBINARY
	default:
		return "", ""
	}
}
