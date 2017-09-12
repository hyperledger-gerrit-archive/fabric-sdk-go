/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package apitxn

// QueryRequest contains the parameters for query
type QueryRequest struct {
	ChaincodeID string
	Fcn         string
	Args        []string
}

// QueryOpts allows the user to specify more advanced options
type QueryOpts struct {
	async chan string // async
}

// Channel Client
/*
 * A channel client instance provides a handler to interact with peers on specified channel.
 * An application that requires interaction with multiple channels should create a separate
 * instance of the channel client for each channel. Channel client supports non-admin functions only.
 *
 * Each Client instance maintains {@link Channel} instance representing channel and the associated
 * private ledgers.
 *
 */
type ChannelClient interface {
	// Query chaincode
	Query(request QueryRequest) (string, error)
	// QueryWithOpts allows the user to provide options for query (sync vs async, etc.)
	QueryWithOpts(request QueryRequest, opt QueryOpts) error
}
