package integration

import (
	"fmt"
	"os"
	"testing"

	config "github.com/hyperledger/fabric-sdk-go/api/apiconfig"
)

var testFabricConfig config.Config
var testDefaultCount = 0

func TestMain(m *testing.M) {
	setup()
	r := m.Run()
	teardown()
	os.Exit(r)
}

func setup() {
	// do any test setup for all tests here...
	var err error

	testSetup := BaseSetupImpl{
		ConfigFile: "../fixtures/config/config_test.yaml",
	}

	testFabricConfig, err = testSetup.InitConfig()
	if err != nil {
		fmt.Printf("Failed InitConfig [%s]\n", err)
		os.Exit(1)
	}
}

func teardown() {
	// do any teadown activities here ..
	testFabricConfig = nil
}

func TestDefaultConfig(t *testing.T) {
	testSetup := &BaseSetupImpl{
		ConfigFile:      "../../pkg/config/config.yaml", // explicitly set default config.yaml as setup() sets config_test.yaml for all tests
		ChannelID:       "mychannel",
		OrgID:           org1Name,
		ChannelConfig:   "../fixtures/channel/mychannel.tx",
		ConnectEventHub: true,
	}

	c, err := testSetup.InitConfig()
	if err != nil {
		t.Fatalf("Failed to load default config: %v", err)
	}
	n, err := c.NetworkConfig()
	if err != nil {
		t.Fatalf("Failed to load default network config: %v", err)
	}

	if n.Name != "default-network" {
		t.Fatalf("Default network was not loaded. Network name loaded is: %s", n.Name)
	}

	testDefaultCount = testDefaultCount + 1
}

func TestDefaultConfigFromEnvVariable(t *testing.T) {
	testSetup := &BaseSetupImpl{
		ConfigFile:      "../../pkg/config/config.yaml", // explicitly set default config.yaml as Setup test sets config_test.yaml for all tests
		ChannelID:       "mychannel",
		OrgID:           org1Name,
		ChannelConfig:   "../fixtures/channel/mychannel.tx",
		ConnectEventHub: true,
	}
	// set env variable
	os.Setenv("DEFAULT_SDK_CONFIG_PATH", "$GOPATH/src/github.com/hyperledger/fabric-sdk-go/pkg/config")
	c, err := testSetup.InitConfig()
	if err != nil {
		t.Fatalf("Failed to load default config: %v", err)
	}
	n, err := c.NetworkConfig()
	if err != nil {
		t.Fatalf("Failed to load default network config: %v", err)
	}

	if n.Name != "default-network" {
		t.Fatalf("Default network was not loaded. Network name loaded is: %s", n.Name)
	}
	testDefaultCount = testDefaultCount + 1
}
