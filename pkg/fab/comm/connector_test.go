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

	sdkContext "github.com/hyperledger/fabric-sdk-go/pkg/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
	"github.com/pkg/errors"
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
	time.Sleep(shortSleepTime * time.Millisecond)

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
	time.Sleep(shortSleepTime * time.Millisecond)
	assert.Equal(t, connectivity.Shutdown, conn1.GetState(), "connection should be shutdown")
	assert.NotEqual(t, connectivity.Shutdown, conn3.GetState(), "connection should not be shutdown")

	ctx, cancel = context.WithTimeout(context.Background(), normalTimeout)
	conn4, err := connector.DialContext(ctx, endorserAddr[0], grpc.WithInsecure())
	cancel()
	assert.Nil(t, err, "DialContext should have succeeded")

	assert.NotEqual(t, unsafe.Pointer(conn1), unsafe.Pointer(conn4), "connections should be different due to disconnect")
}

func TestConnectorConcurrent1(t *testing.T) {
	const goroutines = 500

	connector := NewCachingConnector(shortSweepTime, shortIdleTime)
	defer connector.Close()

	wg := sync.WaitGroup{}

	// Test immediate release
	wg.Add(goroutines)
	errChan := make(chan error, goroutines)
	for i := 0; i < goroutines; i++ {
		go testDial(t, &wg, errChan, connector, endorserAddr[i%len(endorserAddr)], 0, 1)
	}
	wg.Wait()
	select {
	case err := <-errChan:
		t.Fatalf("testDial failed %s", err)
	default:
	}

	// Test long intervals for releasing connection
	wg.Add(goroutines)
	errChan = make(chan error, goroutines)
	for i := 0; i < goroutines; i++ {
		go testDial(t, &wg, errChan, connector, endorserAddr[i%len(endorserAddr)], shortSleepTime*3, 1)
	}
	wg.Wait()
	select {
	case err := <-errChan:
		t.Fatalf("testDial failed %s", err)
	default:
	}
}

func TestConnectorConcurrent2(t *testing.T) {
	const goroutines = 500

	connector := NewCachingConnector(shortSweepTime, shortIdleTime)
	defer connector.Close()

	wg := sync.WaitGroup{}
	// Test mixed intervals for releasing connection
	wg.Add(goroutines)
	errChan := make(chan error, goroutines)
	for i := 0; i < goroutines/2; i++ {
		go testDial(t, &wg, errChan, connector, endorserAddr[0], 0, 1)
		go testDial(t, &wg, errChan, connector, endorserAddr[1], shortSleepTime*3, 1)
	}
	wg.Wait()
	select {
	case err := <-errChan:
		t.Fatalf("testDial failed %s", err)
	default:
	}

	// Test random intervals for releasing connection
	wg.Add(goroutines)
	errChan = make(chan error, goroutines)

	for i := 0; i < goroutines; i++ {
		go testDial(t, &wg, errChan, connector, endorserAddr[i%len(endorserAddr)], 0, shortSleepTime*3)
	}
	wg.Wait()
	select {
	case err := <-errChan:
		t.Fatalf("testDial failed %s", err)
	default:
	}
}

func TestConnectorConcurrentSweep(t *testing.T) {
	const goroutines = 500

	connector := NewCachingConnector(shortSweepTime, shortIdleTime)
	defer connector.Close()

	wg := sync.WaitGroup{}
	errChan := make(chan error, goroutines)

	for j := 0; j < len(endorserAddr); j++ {
		wg.Add(goroutines)
		for i := 0; i < goroutines; i++ {
			go testDial(t, &wg, errChan, connector, endorserAddr[0], 0, 0)
		}
		wg.Wait()
		select {
		case err := <-errChan:
			t.Fatalf("testDial failed %s", err)
		default:
		}

		//Sleeping to wait for sweep
		time.Sleep(shortIdleTime)
	}
}

func TestConnectorConcurrentSweepWithParentContextAndServerDown(t *testing.T) {
	const goroutines = 500

	connector := NewCachingConnector(shortSweepTime, shortIdleTime)
	defer connector.Close()

	failingServerAddr := "127.0.0.1:8888"
	failingServer := &mocks.MockEndorserServer{}
	fAddr := failingServer.Start(failingServerAddr)
	// simulate server down by stopping it
	failingServer.Stop()

	parentCtx, parentCancel := context.WithTimeout(context.Background(), normalTimeout*10)
	defer parentCancel()
	failingChildCtx, cancel := context.WithTimeout(parentCtx, normalTimeout)
	shortDialTimeout := 500 * time.Millisecond // short timeout to avoid unit test waiting longer than 0.5 seconds
	failingChildCtx = context.WithValue(parentCtx, sdkContext.ChildTimeoutContextKey, shortDialTimeout)
	// start 1 DialContext with failing server
	conn, err := connector.DialContext(failingChildCtx, fAddr, grpc.WithInsecure(), grpc.WithBlock())
	if err == nil {
		connector.ReleaseConn(conn)
		t.Fatal("testDial with shutdown server did not fail")
	}
	cancel()
	errChan := make(chan error, goroutines)
	wg := sync.WaitGroup{}

	// execute with running server using same context
	for j := 0; j < len(endorserAddr); j++ {
		wg.Add(goroutines)
		for i := 0; i < goroutines; i++ {
			go testDialWithContext(parentCtx, t, &wg, errChan, connector, endorserAddr[0], 0, 0)
		}
		wg.Wait()
		select {
		case err := <-errChan:
			t.Fatalf("testDial failed %s", err)
		default:
		}

		//Sleeping to wait for sweep
		time.Sleep(shortIdleTime)
	}

	// execute with running server using new context
	for j := 0; j < len(endorserAddr); j++ {
		wg.Add(goroutines)
		for i := 0; i < goroutines; i++ {
			go testDialWithContext(nil, t, &wg, errChan, connector, endorserAddr[0], 0, 0)
		}
		wg.Wait()
		select {
		case err := <-errChan:
			t.Fatalf("testDial failed %s", err)
		default:
		}

		//Sleeping to wait for sweep
		time.Sleep(shortIdleTime)
	}
}

func testDial(t *testing.T, wg *sync.WaitGroup, errChan chan error, connector *CachingConnector, addr string, minSleepBeforeRelease int, maxSleepBeforeRelease int) {
	testDialWithContext(nil, t, wg, errChan, connector, addr, minSleepBeforeRelease, maxSleepBeforeRelease)
}

func testDialWithContext(parentCtx context.Context, t *testing.T, wg *sync.WaitGroup, errChan chan error, connector *CachingConnector, addr string, minSleepBeforeRelease int, maxSleepBeforeRelease int) {
	defer wg.Done()
	var ctx context.Context
	var cancel context.CancelFunc
	// if parentCtx is not empty, then this DialWithContext is for a child context,
	// in this case, provide only a fraction of the parent timeout
	if parentCtx != nil {
		ctx, cancel = context.WithTimeout(parentCtx, normalTimeout)
		ctx = context.WithValue(ctx, sdkContext.ChildTimeoutContextKey, normalTimeout)
	} else {
		ctx, cancel = context.WithTimeout(context.Background(), normalTimeout*10)
	}

	conn, err := connector.DialContext(ctx, addr, grpc.WithInsecure())
	cancel()

	//t.Logf("connector.DialContext called with ctx %+v, error: %s", ctx, err)
	if err != nil {
		errChan <- errors.WithMessage(err, "Connect failed for target "+addr)
		return
	}
	defer connector.ReleaseConn(conn)

	endorserClient := pb.NewEndorserClient(conn)
	proposal := pb.SignedProposal{}
	resp, err := endorserClient.ProcessProposal(context.Background(), &proposal, grpc.FailFast(false))
	if err != nil {
		errChan <- errors.WithMessage(err, "RPC failed for target "+addr)
		return
	}
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
