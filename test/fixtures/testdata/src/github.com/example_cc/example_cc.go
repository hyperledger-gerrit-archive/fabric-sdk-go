/*
Copyright SecureKey Technologies 2018 All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

		 http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"fmt"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

// Init ...
func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {

	// Write some test data to the ledger
	err := stub.PutState("testRead", []byte("test"))
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)

}

// Invoke ...
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()

	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting at least 1")
	}

	if function == "query" {
		return t.query(stub, args)
	}

	if function == "queryPrivateState" {
		return t.queryPrivateState(stub, args)
	}

	if function == "setState" {
		return t.setState(stub, args)
	}

	if function == "setStateWithRead" {
		return t.setStateWithRead(stub, args)
	}

	if function == "setPrivateState" {
		return t.setPrivateState(stub, args)
	}

	if function == "setPrivateStateWithRead" {
		return t.setPrivateStateWithRead(stub, args)
	}

	return shim.Error("Unknown action")
}

func (t *SimpleChaincode) query(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if args[0] == "" {
		return shim.Error("query key is a required argument")
	}
	resp, err := stub.GetState(args[0])
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(resp)
}

func (t *SimpleChaincode) queryPrivateState(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if args[0] == "" {
		return shim.Error("query key is a required argument")
	}
	resp, err := stub.GetPrivateData("billing", args[0])
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(resp)
}

func (t *SimpleChaincode) setState(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 2 || args[0] == "" || args[1] == "" {
		return shim.Error("two args are required to setState, key and value")
	}
	err := stub.PutState(args[0], []byte(args[1]))
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

func (t *SimpleChaincode) setStateWithRead(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 2 || args[0] == "" || args[1] == "" {
		return shim.Error("two args are required to setState, key and value")
	}

	_, err := stub.GetState("testRead")
	if err != nil {
		return shim.Error(err.Error())
	}

	err = stub.PutState(args[0], []byte(args[1]))
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

func (t *SimpleChaincode) setPrivateState(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 2 || args[0] == "" || args[1] == "" {
		return shim.Error("two args are required to setState, key and value")
	}

	err := stub.PutPrivateData("billing", args[0], []byte(args[1]))
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

func (t *SimpleChaincode) setPrivateStateWithRead(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 2 || args[0] == "" || args[1] == "" {
		return shim.Error("two args are required to setState, key and value")
	}

	_, err := stub.GetPrivateData("billing", "testRead")
	if err != nil {
		return shim.Error(err.Error())
	}

	err = stub.PutPrivateData("billing", args[0], []byte(args[1]))
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}
