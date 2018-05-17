/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package comm

import (
	"context"
	"math/rand"
	"sync"
	"testing"
	"time"
	"unsafe"

	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
)

const (
	normalTimeout = 5 * time.Second

	normalSweepTime = 5 * time.Second
	normalIdleTime  = 10 * time.Second
	shortSweepTime  = 10 * time.Nanosecond
	shortIdleTime   = 15 * time.Nanosecond
	shortSleepTime  = 1000
)

func TestConnectorHappyPath(t *testing.T) {
	connector := NewCachingConnector(normalSweepTime, normalIdleTime)
	defer connector.Close()

	ctx, cancel := context.WithTimeout(context.Background(), normalTimeout)
	conn1, err := connector.DialContext(ctx, endorserAddr[0], grpc.WithInsecure())
	cancel()
	assert.Nil(t, err, "DialContext should have succeeded")

	assert.NotEqual(t, connectivity.Connecting, conn1.GetState(), "connection should not be connecting")
	assert.NotEqual(t, connectivity.Shutdown, conn1.GetState(), "connection should not be shutdown")

	ctx, cancel = context.WithTimeout(context.Background(), normalTimeout)
	conn2, err := connector.DialContext(ctx, endorserAddr[0], grpc.WithInsecure())
	cancel()
	assert.Nil(t, err, "DialContext should have succeeded")
	assert.Equal(t, unsafe.Pointer(conn1), unsafe.Pointer(conn2), "connections should match")

	ctx, cancel = context.WithTimeout(context.Background(), normalTimeout)
	conn3, err := connector.DialContext(ctx, endorserAddr[1], grpc.WithInsecure())
	cancel()

	assert.NotEqual(t, connectivity.Connecting, conn3.GetState(), "connection should not be connecting")
	assert.NotEqual(t, connectivity.Shutdown, conn3.GetState(), "connection should not be shutdown")

	assert.Nil(t, err, "DialContext should have succeeded")
	assert.NotEqual(t, unsafe.Pointer(conn1), unsafe.Pointer(conn3), "connections should not match")
}

func TestConnectorDoubleClose(t *testing.T) {
	connector := NewCachingConnector(normalSweepTime, normalIdleTime)
	defer connector.Close()
	connector.Close()
}

func TestReleaseAfterClose(t *testing.T) {
	connector := NewCachingConnector(normalSweepTime, normalIdleTime)
	defer connector.Close()

	ctx, cancel := context.WithTimeout(context.Background(), normalTimeout)
	conn1, err := connector.DialContext(ctx, endorserAddr[0], grpc.WithInsecure())
	cancel()
	assert.Nil(t, err, "DialContext should have succeeded")
	connector.Close()
	assert.Equal(t, connectivity.Shutdown, conn1.GetState(), "connection should be shutdown")
	connector.ReleaseConn(conn1)
}

func TestDialAfterClose(t *testing.T) {
	connector := NewCachingConnector(normalSweepTime, normalIdleTime)

	ctx, cancel := context.WithTimeout(context.Background(), normalTimeout)
	conn1, err := connector.DialContext(ctx, endorserAddr[0], grpc.WithInsecure())
	cancel()
	assert.Nil(t, err, "DialContext should have succeeded")
	connector.Close()
	assert.Equal(t, connectivity.Shutdown, conn1.GetState(), "connection should be shutdown")
	_, err = connector.DialContext(ctx, endorserAddr[0], grpc.WithInsecure())
	assert.Error(t, err, "expecting error when dialing after connector is closed")
}

func TestConnectorHappyFlushNumber1(t *testing.T) {
	connector := NewCachingConnector(normalSweepTime, normalIdleTime)
	defer connector.Close()

	ctx, cancel := context.WithTimeout(context.Background(), normalTimeout)
	conn1, err := connector.DialContext(ctx, endorserAddr[0], grpc.WithInsecure())
	cancel()
	assert.Nil(t, err, "DialContext should have succeeded")

	connector.Close()
	assert.Equal(t, connectivity.Shutdown, conn1.GetState(), "connection should be shutdown")
}

func TestConnectorHappyFlushNumber2(t *testing.T) {
	connector := NewCachingConnector(normalSweepTime, normalIdleTime)
	defer connector.Close()

	ctx, cancel := context.WithTimeout(context.Background(), normalTimeout)
	conn1, err := connector.DialContext(ctx, endorserAddr[0], grpc.WithInsecure())
	cancel()
	assert.Nil(t, err, "DialContext should have succeeded")

	ctx, cancel = context.WithTimeout(context.Background(), normalTimeout)
	conn2, err := connector.DialContext(ctx, endorserAddr[0], grpc.WithInsecure())
	cancel()
	assert.Nil(t, err, "DialContext should have succeeded")

	ctx, cancel = context.WithTimeout(context.Background(), normalTimeout)
	conn3, err := connector.DialContext(ctx, endorserAddr[1], grpc.WithInsecure())
	cancel()
	assert.Nil(t, err, "DialContext should have succeeded")

	connector.Close()
	assert.Equal(t, connectivity.Shutdown, conn1.GetState(), "connection should be shutdown")
	assert.Equal(t, connectivity.Shutdown, conn2.GetState(), "connection should be shutdown")
	assert.Equal(t, connectivity.Shutdown, conn3.GetState(), "connection should be shutdown")
}

func TestConnectorShouldJanitorRestart(t *testing.T) {
	connector := NewCachingConnector(shortSweepTime, shortIdleTime)
	defer connector.Close()

	ctx, cancel := context.WithTimeout(context.Background(), normalTimeout)
	conn1, err := connector.DialContext(ctx, endorserAddr[0], grpc.WithInsecure())
	cancel()
	assert.Nil(t, err, "DialContext should have succeeded")

	connector.ReleaseConn(conn1)
	time.Sleep(shortIdleTime * 50)

	ctx, cancel = context.WithTimeout(context.Background(), normalTimeout)
	conn2, err := connector.DialContext(ctx, endorserAddr[0], grpc.WithInsecure())
	cancel()
	assert.Nil(t, err, "DialContext should have succeeded")

	assert.NotEqual(t, unsafe.Pointer(conn1), unsafe.Pointer(conn2), "connections should be different due to disconnect")
}

func TestConnectorShouldSweep(t *testing.T) {
	connector := NewCachingConnector(shortSweepTime, shortIdleTime)
	defer connector.Close()

	ctx, cancel := context.WithTimeout(context.Background(), normalTimeout)
	conn1, err := connector.DialContext(ctx, endorserAddr[0], grpc.WithInsecure())
	cancel()
	assert.Nil(t, err, "DialContext should have succeeded")

	ctx, cancel = context.WithTimeout(context.Background(), normalTimeout)
	conn3, err := connector.DialContext(ctx, endorserAddr[1], grpc.WithInsecure())
	cancel()
	assert.Nil(t, err, "DialContext should have succeeded")

	connector.ReleaseConn(conn1)
	time.Sleep(normalTimeout)
	assert.Equal(t, connectivity.Shutdown, conn1.GetState(), "connection should be shutdown")
	assert.NotEqual(t, connectivity.Shutdown, conn3.GetState(), "connection should not be shutdown")

	ctx, cancel = context.WithTimeout(context.Background(), normalTimeout)
	conn4, err := connector.DialContext(ctx, endorserAddr[0], grpc.WithInsecure())
	cancel()
	assert.Nil(t, err, "DialContext should have succeeded")

	assert.NotEqual(t, unsafe.Pointer(conn1), unsafe.Pointer(conn4), "connections should be different due to disconnect")
}

func TestConnectorConcurrent(t *testing.T) {
	const goroutines = 5000

	connector := NewCachingConnector(shortSweepTime, shortIdleTime)
	defer connector.Close()

	wg := sync.WaitGroup{}

	// Test immediate release
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go testDial(t, &wg, connector, endorserAddr[i%len(endorserAddr)], 0, 1)
	}
	wg.Wait()

	// Test long intervals for releasing connection
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go testDial(t, &wg, connector, endorserAddr[i%len(endorserAddr)], shortSleepTime*3, 1)
	}
	wg.Wait()

	// Test mixed intervals for releasing connection
	wg.Add(goroutines)
	for i := 0; i < goroutines/2; i++ {
		go testDial(t, &wg, connector, endorserAddr[0], 0, 1)
		go testDial(t, &wg, connector, endorserAddr[1], shortSleepTime*3, 1)
	}
	wg.Wait()

	// Test random intervals for releasing connection
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go testDial(t, &wg, connector, endorserAddr[i%len(endorserAddr)], 0, shortSleepTime*3)
	}
	wg.Wait()
}

func TestConnectorConcurrentSweep(t *testing.T) {
	const goroutines = 50

	connector := NewCachingConnector(shortSweepTime, shortIdleTime)
	defer connector.Close()

	wg := sync.WaitGroup{}

	for j := 0; j < len(endorserAddr); j++ {
		wg.Add(goroutines)
		for i := 0; i < goroutines; i++ {
			go testDial(t, &wg, connector, endorserAddr[0], 0, 1)
		}
		wg.Wait()

		//Sleeping to wait for sweep
		time.Sleep(shortIdleTime)
	}
}

func testDial(t *testing.T, wg *sync.WaitGroup, connector *CachingConnector, addr string, minSleepBeforeRelease int, maxSleepBeforeRelease int) {
	defer wg.Done()

	ctx, cancel := context.WithTimeout(context.Background(), normalTimeout)
	conn, err := connector.DialContext(ctx, addr, grpc.WithInsecure())
	cancel()
	assert.Nil(t, err, "DialContext should have succeeded")
	defer connector.ReleaseConn(conn)

	endorserClient := pb.NewEndorserClient(conn)
	proposal := pb.SignedProposal{}
	resp, err := endorserClient.ProcessProposal(context.Background(), &proposal)

	require.Nil(t, err, "peer process proposal should not have error")
	require.NotNil(t, resp)
	require.Equal(t, int32(200), resp.GetResponse().Status)

	var randomSleep int
	if maxSleepBeforeRelease == 0 {
		randomSleep = 0
	} else {
		randomSleep = rand.Intn(maxSleepBeforeRelease)
	}
	time.Sleep(time.Duration(minSleepBeforeRelease)*time.Millisecond + time.Duration(randomSleep)*time.Millisecond)
}
