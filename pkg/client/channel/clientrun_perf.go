// +build pprof

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package channel

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel/invoke"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/common/discovery/greylist"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/test/performance/metrics"
	"github.com/uber-go/tally"
)

type clientTally struct {
	queryCount     tally.Counter
	queryFailCount tally.Counter
	queryTimer     tally.Timer

	executeCount     tally.Counter
	executeFailCount tally.Counter
	executeTimer     tally.Timer
}

// Client enables access to a channel on a Fabric network.
//
// A channel client instance provides a handler to interact with peers on specified channel.
// An application that requires interaction with multiple channels should create a separate
// instance of the channel client for each channel. Channel client supports non-admin functions only.
type Client struct {
	context      context.Channel
	membership   fab.ChannelMembership
	eventService fab.EventService
	greylist     *greylist.Filter
	tally        *clientTally
}

func newClientTally(channelContext context.Channel) *clientTally {
	return &clientTally{
		queryCount:     metrics.RootScope.SubScope(channelContext.ChannelID()).Counter("ch_client_query_calls"),
		queryFailCount: metrics.RootScope.SubScope(channelContext.ChannelID()).Counter("ch_client_query_errors"),
		queryTimer:     metrics.RootScope.SubScope(channelContext.ChannelID()).Timer("ch_client_query_processing_time_seconds"),

		executeCount:     metrics.RootScope.SubScope(channelContext.ChannelID()).Counter("ch_client_execute_calls"),
		executeFailCount: metrics.RootScope.SubScope(channelContext.ChannelID()).Counter("ch_client_execute_errors"),
		executeTimer:     metrics.RootScope.SubScope(channelContext.ChannelID()).Timer("ch_client_execute_processing_time_seconds"),
	}
}

func newClient(channelContext context.Channel, membership fab.ChannelMembership, eventService fab.EventService, greylistProvider *greylist.Filter) Client {
	ct := newClientTally(channelContext)

	channelClient := Client{
		membership:   membership,
		eventService: eventService,
		greylist:     greylistProvider,
		context:      channelContext,
		tally:        ct,
	}
	return channelClient
}

func callQuery(cc *Client, request Request, options ...RequestOption) (Response, error) {
	cc.tally.executeCount.Inc(1)
	stopWatch := cc.tally.queryTimer.Start()
	defer stopWatch.Stop()

	r, err := cc.InvokeHandler(invoke.NewQueryHandler(), request, options...)
	if err != nil {
		cc.tally.queryFailCount.Inc(1)
	}
	return r, err
}

func callExecute(cc *Client, request Request, options ...RequestOption) (Response, error) {
	cc.tally.executeCount.Inc(1)
	stopWatch := cc.tally.executeTimer.Start()
	defer stopWatch.Stop()

	r, err := cc.InvokeHandler(invoke.NewExecuteHandler(), request, options...)
	if err != nil {
		cc.tally.executeFailCount.Inc(1)
	}
	return r, err
}
