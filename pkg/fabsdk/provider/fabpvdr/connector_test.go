/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabpvdr

import (
	"context"
	"fmt"
	"net"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"

	"github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging/loglevel"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
)

const (
	testAddress   = "127.0.0.1:0"
	normalTimeout = 5 * time.Second

	normalSweepTime = 5 * time.Second
	normalIdleTime  = 10 * time.Second
	shortSweepTime  = 100 * time.Millisecond
	shortIdleTime   = 200 * time.Millisecond
)

var addr1 string
var addr2 string

func TestMain(m *testing.M) {
	var ok bool

	logging.SetLevel("fabric_sdk_go", loglevel.DEBUG)

	grpcServer1 := grpc.NewServer()
	defer grpcServer1.Stop()
	_, addr1, ok = startEndorserServer(grpcServer1)
	if !ok {
		return
	}

	grpcServer2 := grpc.NewServer()
	defer grpcServer2.Stop()
	_, addr2, ok = startEndorserServer(grpcServer2)
	if !ok {
		return
	}

	os.Exit(m.Run())
}

func TestHappyPath(t *testing.T) {
	connector := newCachingConnector(normalSweepTime, normalIdleTime)
	defer connector.Close()

	ctx, cancel := context.WithTimeout(context.Background(), normalTimeout)
	conn1, err := connector.DialContext(ctx, addr1, grpc.WithInsecure())
	cancel()
	assert.Nil(t, err, "DialContext should have succeeded")

	assert.NotEqual(t, connectivity.Connecting, conn1.GetState(), "connection should not be connecting")
	assert.NotEqual(t, connectivity.Shutdown, conn1.GetState(), "connection should not be shutdown")

	ctx, cancel = context.WithTimeout(context.Background(), normalTimeout)
	conn2, err := connector.DialContext(ctx, addr1, grpc.WithInsecure())
	cancel()
	assert.Nil(t, err, "DialContext should have succeeded")
	assert.Equal(t, conn1, conn2, "connections should match")

	ctx, cancel = context.WithTimeout(context.Background(), normalTimeout)
	conn3, err := connector.DialContext(ctx, addr2, grpc.WithInsecure())
	cancel()

	assert.NotEqual(t, connectivity.Connecting, conn3.GetState(), "connection should not be connecting")
	assert.NotEqual(t, connectivity.Shutdown, conn3.GetState(), "connection should not be shutdown")

	assert.Nil(t, err, "DialContext should have succeeded")
	assert.NotEqual(t, conn1, conn3, "connections should not match")
}

func TestDoubleClose(t *testing.T) {
	connector := newCachingConnector(normalSweepTime, normalIdleTime)
	defer connector.Close()
	connector.Close()
}

func TestHappyFlushNumber1(t *testing.T) {
	connector := newCachingConnector(normalSweepTime, normalIdleTime)
	defer connector.Close()

	ctx, cancel := context.WithTimeout(context.Background(), normalTimeout)
	conn1, err := connector.DialContext(ctx, addr1, grpc.WithInsecure())
	cancel()
	assert.Nil(t, err, "DialContext should have succeeded")

	connector.Close()
	assert.Equal(t, connectivity.Shutdown, conn1.GetState(), "connection should be shutdown")
}

func TestHappyFlushNumber2(t *testing.T) {
	connector := newCachingConnector(normalSweepTime, normalIdleTime)
	defer connector.Close()

	ctx, cancel := context.WithTimeout(context.Background(), normalTimeout)
	conn1, err := connector.DialContext(ctx, addr1, grpc.WithInsecure())
	cancel()
	assert.Nil(t, err, "DialContext should have succeeded")

	ctx, cancel = context.WithTimeout(context.Background(), normalTimeout)
	conn2, err := connector.DialContext(ctx, addr1, grpc.WithInsecure())
	cancel()
	assert.Nil(t, err, "DialContext should have succeeded")

	ctx, cancel = context.WithTimeout(context.Background(), normalTimeout)
	conn3, err := connector.DialContext(ctx, addr2, grpc.WithInsecure())
	cancel()
	assert.Nil(t, err, "DialContext should have succeeded")

	connector.Close()
	assert.Equal(t, connectivity.Shutdown, conn1.GetState(), "connection should be shutdown")
	assert.Equal(t, connectivity.Shutdown, conn2.GetState(), "connection should be shutdown")
	assert.Equal(t, connectivity.Shutdown, conn3.GetState(), "connection should be shutdown")
}

func TestShouldSweep(t *testing.T) {
	connector := newCachingConnector(shortSweepTime, shortIdleTime)
	defer connector.Close()

	ctx, cancel := context.WithTimeout(context.Background(), normalTimeout)
	conn1, err := connector.DialContext(ctx, addr1, grpc.WithInsecure())
	cancel()
	assert.Nil(t, err, "DialContext should have succeeded")

	ctx, cancel = context.WithTimeout(context.Background(), normalTimeout)
	conn3, err := connector.DialContext(ctx, addr2, grpc.WithInsecure())
	cancel()
	assert.Nil(t, err, "DialContext should have succeeded")

	connector.ReleaseConn(conn1)
	time.Sleep(shortIdleTime * 3)
	assert.Equal(t, connectivity.Shutdown, conn1.GetState(), "connection should be shutdown")
	assert.NotEqual(t, connectivity.Shutdown, conn3.GetState(), "connection should not be shutdown")

	ctx, cancel = context.WithTimeout(context.Background(), normalTimeout)
	conn4, err := connector.DialContext(ctx, addr1, grpc.WithInsecure())
	cancel()
	assert.Nil(t, err, "DialContext should have succeeded")

	assert.NotEqual(t, conn1, conn4, "connections should be different due to disconnect")
}

func startEndorserServer(grpcServer *grpc.Server) (*mocks.MockEndorserServer, string, bool) {
	lis, err := net.Listen("tcp", testAddress)
	if err != nil {
		fmt.Printf("Error starting test server %s", err)
		return nil, "", false
	}
	addr := lis.Addr().String()

	endorserServer := &mocks.MockEndorserServer{}
	pb.RegisterEndorserServer(grpcServer, endorserServer)
	fmt.Printf("Starting test server on %s", addr)
	go grpcServer.Serve(lis)
	return endorserServer, addr, true
}
