// Copyright SecureKey Technologies Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0

module github.com/hyperledger/fabric-sdk-go/test/integration

replace github.com/hyperledger/fabric-sdk-go => ../../

replace github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric => ../../third_party/github.com/hyperledger/fabric

require (
	github.com/hyperledger/fabric-sdk-go v0.0.0-00010101000000-000000000000
	github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric v0.0.0-00010101000000-000000000000
	github.com/pkg/errors v0.8.0
	github.com/stretchr/testify v1.2.2
)
