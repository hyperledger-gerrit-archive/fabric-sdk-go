/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package dynamicselection

import (
	"fmt"
	"strings"
	"sync"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/api/apitxn"
	"github.com/hyperledger/fabric-sdk-go/def/fabapi"
	"github.com/hyperledger/fabric-sdk-go/pkg/errors"
	clientImpl "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"

	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/core/common/ccprovider"

	peerImpl "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/peer"
)

const (
	ccDataProviderSCC      = "lscc"
	ccDataProviderfunction = "getccdata"
)

// CCPolicyProvider retrieves policy for the given chaincode ID
type CCPolicyProvider interface {
	GetChaincodePolicy(chaincodeID string) (*common.SignaturePolicyEnvelope, error)
}

// NewCCPolicyProvider creates new chaincode policy data provider
func newCCPolicyProvider(sdk *fabapi.FabricSDK, channelID string, userName string, orgName string) (CCPolicyProvider, error) {

	if channelID == "" || userName == "" || orgName == "" {
		return nil, errors.New("Must provide channel ID, user name and organisation for cc policy provider")
	}

	if sdk == nil {
		return nil, errors.New("Must provide sdk")
	}

	session, err := sdk.NewPreEnrolledUserSession(orgName, userName)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to get pre-enrolled user session")
	}

	client := clientImpl.NewClient(sdk.ConfigProvider())
	client.SetCryptoSuite(sdk.CryptoSuiteProvider())
	client.SetStateStore(sdk.StateStoreProvider())
	client.SetUserContext(session.Identity())
	client.SetSigningManager(sdk.SigningManager())

	targetPeers, err := sdk.ConfigProvider().ChannelPeers(channelID)
	if err != nil {
		return nil, errors.WithMessage(err, "unable to read configuration for channel peers")
	}

	channel, err := client.NewChannel(channelID)
	if err != nil {
		return nil, errors.WithMessage(err, "NewChannel failed")
	}

	return &ccPolicyProvider{config: sdk.ConfigProvider(), channel: channel, targetPeers: targetPeers, ccDataMap: make(map[string]*ccprovider.ChaincodeData)}, nil
}

type ccPolicyProvider struct {
	config      apiconfig.Config
	channel     fab.Channel
	targetPeers []apiconfig.ChannelPeer
	ccDataMap   map[string]*ccprovider.ChaincodeData
	mutex       sync.RWMutex
}

func (dp *ccPolicyProvider) GetChaincodePolicy(chaincodeID string) (*common.SignaturePolicyEnvelope, error) {

	if chaincodeID == "" {
		return nil, errors.New("Must provide chaincode ID")
	}

	channelID := dp.channel.Name()
	key := newResolverKey(channelID, chaincodeID)
	var ccData *ccprovider.ChaincodeData

	dp.mutex.RLock()
	ccData = dp.ccDataMap[chaincodeID]
	dp.mutex.RUnlock()
	if ccData != nil {
		return unmarshalPolicy(ccData.Policy)
	}

	dp.mutex.Lock()
	defer dp.mutex.Unlock()

	response, err := dp.queryChaincode(channelID, ccDataProviderSCC, ccDataProviderfunction, [][]byte{[]byte(channelID), []byte(chaincodeID)})
	if err != nil {
		return nil, errors.WithMessage(err, fmt.Sprintf("error querying chaincode data for chaincode [%s] on channel [%s]", chaincodeID, channelID))
	}

	ccData = &ccprovider.ChaincodeData{}
	err = proto.Unmarshal(response.ProposalResponse.Response.Payload, ccData)
	if err != nil {
		return nil, errors.WithMessage(err, "Error unmarshalling chaincode data")
	}

	dp.ccDataMap[key.String()] = ccData

	return unmarshalPolicy(ccData.Policy)
}

func unmarshalPolicy(policy []byte) (*common.SignaturePolicyEnvelope, error) {

	sigPolicyEnv := &common.SignaturePolicyEnvelope{}
	if err := proto.Unmarshal(policy, sigPolicyEnv); err != nil {
		return nil, errors.WithMessage(err, "error unmarshalling SignaturePolicyEnvelope")
	}

	return sigPolicyEnv, nil
}

func (dp *ccPolicyProvider) clearCache() {
	dp.mutex.Lock()
	defer dp.mutex.Unlock()
	dp.ccDataMap = make(map[string]*ccprovider.ChaincodeData)
}

func (dp *ccPolicyProvider) queryChaincode(channelID string, ccID string, ccFcn string, ccArgs [][]byte) (*apitxn.TransactionProposalResponse, error) {
	logger.Debugf("queryChaincode channelID:%s", channelID)

	var queryErrors []string
	var response *apitxn.TransactionProposalResponse
	for _, p := range dp.targetPeers {

		serverHostOverride := ""
		if str, ok := p.GRPCOptions["ssl-target-name-override"].(string); ok {
			serverHostOverride = str
		}

		peer, err := peerImpl.NewPeerTLSFromCert(p.URL, p.TLSCACerts.Path, serverHostOverride, dp.config)
		if err != nil {
			queryErrors = append(queryErrors, err.Error())
			continue
		}

		// Send query to channel peer
		request := apitxn.ChaincodeInvokeRequest{
			Targets:      []apitxn.ProposalProcessor{peer},
			Fcn:          ccFcn,
			Args:         ccArgs,
			TransientMap: nil,
			ChaincodeID:  ccID,
		}

		responses, _, err := dp.channel.SendTransactionProposal(request)
		if err != nil {
			queryErrors = append(queryErrors, err.Error())
			continue
		} else if responses[0].Err != nil {
			queryErrors = append(queryErrors, responses[0].Err.Error())
			continue
		} else {
			// Valid response obtained, stop querying
			response = responses[0]
			break
		}
	}
	logger.Debugf("queryErrors: %v", queryErrors)

	// If all queries failed, return error
	if len(queryErrors) == len(dp.targetPeers) {
		errMsg := fmt.Sprintf("Error querying peers for channel %s: %s", channelID, strings.Join(queryErrors, "\n"))
		return nil, errors.New(errMsg)
	}

	return response, nil
}
