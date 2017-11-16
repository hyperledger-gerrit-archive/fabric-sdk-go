/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package integration

import (
	"strconv"
	"testing"
	"time"

	"fmt"

	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	"github.com/hyperledger/fabric-sdk-go/api/apicryptosuite"
	"github.com/hyperledger/fabric-sdk-go/api/apitxn"
	"github.com/hyperledger/fabric-sdk-go/def/fabapi"
	"github.com/hyperledger/fabric-sdk-go/def/fabapi/context/defprovider"
	"github.com/hyperledger/fabric-sdk-go/pkg/cryptosuite"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/bccsp"
	bccspFactory "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/bccsp/factory"
)

const samplekey = "sample-key"

func TestEndToEndForCustomCryptoSuite(t *testing.T) {

	testSetup := BaseSetupImpl{
		ConfigFile:      ConfigTestFile,
		ChannelID:       "mychannel",
		OrgID:           org1Name,
		ChannelConfig:   "../fixtures/channel/mychannel.tx",
		ConnectEventHub: true,
	}

	if err := testSetup.Initialize(t); err != nil {
		t.Fatalf(err.Error())
	}

	if err := testSetup.InstallAndInstantiateExampleCC(); err != nil {
		t.Fatalf("InstallAndInstantiateExampleCC return error: %v", err)
	}

	if err := testSetup.UpgradeExampleCC(); err != nil {
		t.Fatalf("UpgradeExampleCC return error: %v", err)
	}

	defaultConfig, err := testSetup.InitConfig()

	if err != nil {
		panic(fmt.Sprintf("Failed to get default config [%s]", err))
	}

	//Get Test BCCSP,
	// TODO Need to use external BCCSP here
	customBccspProvider := getTestBCCSP(defaultConfig)

	// Create SDK setup with custom cryptosuite provider factory
	sdkOptions := fabapi.Options{
		ConfigFile:      testSetup.ConfigFile,
		ProviderFactory: &CustomCryptoSuiteProviderFactory{bccspProvider: customBccspProvider},
	}

	sdk, err := fabapi.NewSDK(sdkOptions)
	if err != nil {
		t.Fatalf("Failed to create new SDK: %s", err)
	}

	chClient, err := sdk.NewChannelClient(testSetup.ChannelID, "User1")
	if err != nil {
		t.Fatalf("Failed to create new channel client: %s", err)
	}

	value, err := chClient.Query(apitxn.QueryRequest{ChaincodeID: testSetup.ChainCodeID, Fcn: "invoke", Args: queryArgs})
	if err != nil {
		t.Fatalf("Failed to query funds: %s", err)
	}

	t.Logf("*** QueryValue before invoke %s", value)

	// Check the Query value equals upgrade arguments (400)
	if string(value) != "400" {
		t.Fatalf("UpgradeExampleCC failed, query value doesn't match upgrade arguments")
	}

	eventID := "test([a-zA-Z]+)"

	// Register chaincode event (pass in channel which receives event details when the event is complete)
	notifier := make(chan *apitxn.CCEvent)
	rce := chClient.RegisterChaincodeEvent(notifier, testSetup.ChainCodeID, eventID)

	// Move funds
	_, err = chClient.ExecuteTx(apitxn.ExecuteTxRequest{ChaincodeID: testSetup.ChainCodeID, Fcn: "invoke", Args: txArgs})
	if err != nil {
		t.Fatalf("Failed to move funds: %s", err)
	}

	select {
	case ccEvent := <-notifier:
		t.Logf("Received CC event: %s\n", ccEvent)
	case <-time.After(time.Second * 20):
		t.Fatalf("Did NOT receive CC event for eventId(%s)\n", eventID)
	}

	// Unregister chain code event using registration handle
	err = chClient.UnregisterChaincodeEvent(rce)
	if err != nil {
		t.Fatalf("Unregister cc event failed: %s", err)
	}

	// Verify move funds transaction result
	valueAfterInvoke, err := chClient.Query(apitxn.QueryRequest{ChaincodeID: testSetup.ChainCodeID, Fcn: "invoke", Args: queryArgs})
	if err != nil {
		t.Fatalf("Failed to query funds after transaction: %s", err)
	}

	t.Logf("*** QueryValue after invoke %s", valueAfterInvoke)

	valueInt, _ := strconv.Atoi(string(value))
	valueAfterInvokeInt, _ := strconv.Atoi(string(valueAfterInvoke))
	if valueInt+1 != valueAfterInvokeInt {
		t.Fatalf("ExecuteTx failed. Before: %s, after: %s", value, valueAfterInvoke)
	}

	// Release all channel client resources
	chClient.Close()

}

// CustomCryptoSuiteProviderFactory is will provide custom cryptosuite (bccsp.BCCSP)
type CustomCryptoSuiteProviderFactory struct {
	defprovider.DefaultProviderFactory
	bccspProvider bccsp.BCCSP
}

// NewCryptoSuiteProvider returns a new default implementation of BCCSP
func (f *CustomCryptoSuiteProviderFactory) NewCryptoSuiteProvider(config apiconfig.Config) (apicryptosuite.CryptoSuite, error) {
	return cryptosuite.GetSuite(f.bccspProvider), nil
}

func getTestBCCSP(config apiconfig.Config) bccsp.BCCSP {

	// Initialize bccsp factories before calling get client
	err := bccspFactory.InitFactories(&bccspFactory.FactoryOpts{
		ProviderName: config.SecurityProvider(),
		SwOpts: &bccspFactory.SwOpts{
			HashFamily: config.SecurityAlgorithm(),
			SecLevel:   config.SecurityLevel(),
			FileKeystore: &bccspFactory.FileKeystoreOpts{
				KeyStorePath: config.KeyStorePath(),
			},
			Ephemeral: false,
		},
	})
	if err != nil {
		panic(fmt.Sprintf("Failed getting ephemeral software-based BCCSP [%s]", err))
	}

	return bccspFactory.GetDefault()
}

func TestCustomCryptoSuite(t *testing.T) {
	testSetup := BaseSetupImpl{
		ConfigFile:      ConfigTestFile,
		ChannelID:       "mychannel",
		OrgID:           org1Name,
		ChannelConfig:   "../fixtures/channel/mychannel.tx",
		ConnectEventHub: true,
	}

	if err := testSetup.Initialize(t); err != nil {
		t.Fatalf(err.Error())
	}

	if err := testSetup.InstallAndInstantiateExampleCC(); err != nil {
		t.Fatalf("InstallAndInstantiateExampleCC return error: %v", err)
	}

	if err := testSetup.UpgradeExampleCC(); err != nil {
		t.Fatalf("UpgradeExampleCC return error: %v", err)
	}

	defaultConfig, err := testSetup.InitConfig()

	if err != nil {
		panic(fmt.Sprintf("Failed to get default config [%s]", err))
	}

	//Get Test BCCSP,
	// TODO Need to use external BCCSP here
	customBccspProvider := getTestBCCSP(defaultConfig)
	customBccspWrapper := getBCCSPWrapper(customBccspProvider)

	// Create SDK setup with custom cryptosuite provider factory
	sdkOptions := fabapi.Options{
		ConfigFile:      testSetup.ConfigFile,
		ProviderFactory: &CustomCryptoSuiteProviderFactory{bccspProvider: customBccspWrapper},
	}

	sdk, err := fabapi.NewSDK(sdkOptions)
	if err != nil {
		t.Fatalf("Failed to create new SDK: %s", err)
	}

	key, err := sdk.CryptoSuiteProvider().KeyGen(nil)
	if err != nil {
		t.Fatalf("Failed to get key from  sdk.CryptoSuiteProvider().KeyGen(): %s", err)
	}

	bytes, err := key.Bytes()
	if err != nil {
		t.Fatalf("Failed to get key bytes from  sdk.CryptoSuiteProvider().KeyGen(): %s", err)
	}

	if string(bytes) != samplekey {
		t.Fatalf("Upexpected sdk.CryptoSuiteProvider(), expected to find BCCSPWrapper features : %s", err)
	}
}

/*
	BCCSP Wrapper for test
*/

func getBCCSPWrapper(bccsp bccsp.BCCSP) bccsp.BCCSP {
	return &bccspWrapper{bccsp}
}

func getBCCSPKeyWrapper(key bccsp.Key) bccsp.Key {
	return &bccspKeyWrapper{key}
}

type bccspWrapper struct {
	bccsp.BCCSP
}

func (mock *bccspWrapper) KeyGen(opts bccsp.KeyGenOpts) (k bccsp.Key, err error) {
	return getBCCSPKeyWrapper(nil), nil
}

type bccspKeyWrapper struct {
	bccsp.Key
}

func (mock *bccspKeyWrapper) Bytes() ([]byte, error) {
	return []byte("sample-key"), nil
}
