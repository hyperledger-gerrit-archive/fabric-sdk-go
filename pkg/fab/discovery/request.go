/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package discovery

import (
	discclient "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/discovery/client"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/discovery"
)

type Request struct {
	*discclient.Request
}

func NewRequest() *Request{
	return &Request{discclient.NewRequest()}
}

func (req *Request) AddConfigQuery() *Request {
	req.Request.AddConfigQuery()
	return req
}

func (req *Request) AddLocalPeersQuery() *Request {
	req.Request.AddLocalPeersQuery()
	return req
}

func (req *Request) OfChannel(ch string) *Request {
	req.Request.OfChannel(ch)
	return req
}

func (req *Request) AddEndorsersQuery(interests ...*discovery.ChaincodeInterest) (*Request, error) {
	_, err := req.Request.AddEndorsersQuery(interests...)
	return req, err
}

func (req *Request) AddPeersQuery(invocationChain ...*discovery.ChaincodeCall) *Request {
	req.Request.AddPeersQuery(invocationChain...)
	return req
}

func CcCalls(ccNames ...string) []*discovery.ChaincodeCall {
	var call []*discovery.ChaincodeCall

	for _, ccName := range ccNames {
		call = append(call, &discovery.ChaincodeCall{
			Name: ccName,
		})
	}

	return call
}

func CcInterests(invocationsChains ...[]*discovery.ChaincodeCall) []*discovery.ChaincodeInterest {
	var interests []*discovery.ChaincodeInterest

	for _, invocationChain := range invocationsChains {
		interests = append(interests, &discovery.ChaincodeInterest{
			Chaincodes: invocationChain,
		})
	}

	return interests
}




