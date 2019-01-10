// +build pprof

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package channel

import (
	"fmt"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel/invoke"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/common/discovery/greylist"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/status"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/metrics"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
)

func newClient(channelContext context.Channel, membership fab.ChannelMembership, eventService fab.EventService, greylistProvider *greylist.Filter) Client {
	channelClient := Client{
		membership:   membership,
		eventService: eventService,
		greylist:     greylistProvider,
		context:      channelContext,
	}
	return channelClient
}

func callQuery(cc *Client, request Request, options ...RequestOption) (Response, error) {
	meterLabels := []string{
		"chaincode", request.ChaincodeID,
		"Fcn", request.Fcn,
	}
	metrics.SdkMetrics.QueriesReceived.With(meterLabels...).Add(1)
	startTime := time.Now()
	r, err := cc.InvokeHandler(invoke.NewQueryHandler(), request, options...)
	if err != nil {
		if s, ok := err.(*status.Status); ok {
			if s.Code == status.Timeout.ToInt32() {
				meterLabels = append(meterLabels, "fail", "timeout")
				metrics.SdkMetrics.QueryTimeouts.With(meterLabels...).Add(1)
				return r, err
			}
			meterLabels = append(meterLabels, "fail", fmt.Sprintf("Error - Group:%s - Code:%d", s.Group.String(), s.Code))
			metrics.SdkMetrics.QueriesFailed.With(meterLabels...).Add(1)
			return r, err
		}
		meterLabels = append(meterLabels, "fail", fmt.Sprintf("Error - Generic: %s", err))
		metrics.SdkMetrics.QueriesFailed.With(meterLabels...).Add(1)
		return r, err
	}

	metrics.SdkMetrics.QueryDuration.With(meterLabels...).Observe(time.Since(startTime).Seconds())
	return r, err
}

func callExecute(cc *Client, request Request, options ...RequestOption) (Response, error) {
	meterLabels := []string{
		"chaincode", request.ChaincodeID,
		"Fcn", request.Fcn,
	}
	metrics.SdkMetrics.ExecutionsReceived.With(meterLabels...).Add(1)
	startTime := time.Now()
	r, err := cc.InvokeHandler(invoke.NewExecuteHandler(), request, options...)
	if err != nil {
		if s, ok := err.(*status.Status); ok {
			if s.Code == status.Timeout.ToInt32() {
				meterLabels = append(meterLabels, "fail", "timeout")
				metrics.SdkMetrics.ExecutionTimeouts.With(meterLabels...).Add(1)
				return r, err
			}
			meterLabels = append(meterLabels, "fail", fmt.Sprintf("Error - Group:%s - Code:%d", s.Group.String(), s.Code))
			metrics.SdkMetrics.ExecutionsFailed.With(meterLabels...).Add(1)
			return r, err
		}
		meterLabels = append(meterLabels, "fail", fmt.Sprintf("Error - Generic: %s", err))
		metrics.SdkMetrics.ExecutionsFailed.With(meterLabels...).Add(1)
		return r, err
	}

	metrics.SdkMetrics.ExecutionDuration.With(meterLabels...).Observe(time.Since(startTime).Seconds())
	return r, err
}
