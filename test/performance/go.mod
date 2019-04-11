// Copyright SecureKey Technologies Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0

module github.com/hyperledger/fabric-sdk-go/test/performance

replace github.com/hyperledger/fabric-sdk-go => ../../

replace github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric => ../../third_party/github.com/hyperledger/fabric

require (
	github.com/fsouza/go-dockerclient v1.3.6 // indirect
	github.com/golang/protobuf v1.2.0
	github.com/hyperledger/fabric v2.0.0-alpha+incompatible
	github.com/hyperledger/fabric-amcl v0.0.0-20181230093703-5ccba6eab8d6 // indirect
	github.com/hyperledger/fabric-sdk-go v0.0.0-00010101000000-000000000000
	github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric v0.0.0-20190405202032-8d90000c76e2
	github.com/op/go-logging v0.0.0-20160315200505-970db520ece7 // indirect
	github.com/pkg/errors v0.8.1
	github.com/sykesm/zap-logfmt v0.0.2 // indirect
	golang.org/x/net v0.0.0-20190213061140-3a22650c66bd
	google.golang.org/grpc v1.19.0
)
