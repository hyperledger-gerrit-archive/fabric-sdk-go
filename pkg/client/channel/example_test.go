/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
package channel

import (
	"fmt"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel/invoke"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	"github.com/spf13/viper"
)

func setSystemSettings() {
	viper.Set("operations.listenAddress", "127.0.0.1:9443")
	viper.Set("metrics.provider", "disabled")
	viper.Set("metrics.statsd.network", "udp")
	viper.Set("metrics.statsd.address", "127.0.0.1:8125")
	viper.Set("metrics.statsd.writeInterval", "10s")
	viper.Set("metrics.statsd.prefix", "")

	viper.Set("operations.tls.enabled", false)
	viper.Set("operations.tls.cert.file", "")
	viper.Set("operations.tls.key.file", "")
	viper.Set("operations.tls.clientAuthRequired", false)
	viper.Set("operations.tls.clientRootCAs.files", "")
}

func Example() {
	setSystemSettings()
	c, err := New(mockChannelProvider("mychannel"))
	if err != nil {
		fmt.Println("failed to create client")
	}

	response, err := c.Query(Request{ChaincodeID: "testCC", Fcn: "invoke", Args: [][]byte{[]byte("query"), []byte("data")}})
	if err != nil {
		fmt.Printf("failed to query chaincode: %s\n", err)
	}

	fmt.Println(string(response.Payload))

	// Output: abc
}

func ExampleNew() {
	setSystemSettings()
	ctx := mockChannelProvider("mychannel")

	c, err := New(ctx)
	if err != nil {
		fmt.Println(err)
	}

	if c != nil {
		fmt.Println("channel client created")
	} else {
		fmt.Println("channel client is nil")
	}

	// Output: channel client created

}

func ExampleClient_Query() {
	setSystemSettings()
	c, err := New(mockChannelProvider("mychannel"))
	if err != nil {
		fmt.Println("failed to create client")
	}

	response, err := c.Query(Request{ChaincodeID: "testCC", Fcn: "invoke", Args: [][]byte{[]byte("query"), []byte("b")}})
	if err != nil {
		fmt.Printf("failed to query chaincode: %s\n", err)
	}

	if len(response.Payload) > 0 {
		fmt.Println("chaincode query success")
	}

	// Output: chaincode query success
}

func ExampleClient_Execute() {
	setSystemSettings()
	c, err := New(mockChannelProvider("mychannel"))
	if err != nil {
		fmt.Println("failed to create client")
	}

	_, err = c.Execute(Request{ChaincodeID: "testCC", Fcn: "invoke", Args: [][]byte{[]byte("move"), []byte("a"), []byte("b"), []byte("1")}})
	if err != nil {
		fmt.Println(err.Error())
	}

	fmt.Println("Chaincode transaction completed")

	// Output: Chaincode transaction completed
}

func ExampleClient_RegisterChaincodeEvent() {
	setSystemSettings()
	c, err := New(mockChannelProvider("mychannel"))
	if err != nil {
		fmt.Println("failed to create client")
	}

	registration, _, err := c.RegisterChaincodeEvent("examplecc", "event123")
	if err != nil {
		fmt.Println("failed to register chaincode event")
	}
	defer c.UnregisterChaincodeEvent(registration)

	fmt.Println("chaincode event registered successfully")

	// Output: chaincode event registered successfully

}

func ExampleClient_InvokeHandler() {
	setSystemSettings()
	c, err := New(mockChannelProvider("mychannel"))
	if err != nil {
		fmt.Println("failed to create client")
	}

	response, err := c.InvokeHandler(&exampleHandler{}, Request{ChaincodeID: "testCC", Fcn: "invoke", Args: [][]byte{[]byte("query"), []byte("data")}})
	if err != nil {
		fmt.Printf("failed to query chaincode: %s\n", err)
	}

	fmt.Println(string(response.Payload))

	// Output: custom
}

type exampleHandler struct {
}

func (c *exampleHandler) Handle(requestContext *invoke.RequestContext, clientContext *invoke.ClientContext) {
	requestContext.Response.Payload = []byte("custom")
}

func mockChannelProvider(channelID string) context.ChannelProvider {

	channelProvider := func() (context.Channel, error) {
		return mocks.NewMockChannel(channelID)
	}

	return channelProvider
}
