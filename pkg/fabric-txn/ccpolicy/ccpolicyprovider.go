/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package ccpolicy

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/api/apitxn"
	"github.com/hyperledger/fabric-sdk-go/pkg/errors"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"

	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/core/common/ccprovider"

	peerImpl "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/peer"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging"
)

var logger = logging.NewLogger("fabric_sdk_go")

const (
	ccDataProviderSCC      = "lscc"
	ccDataProviderfunction = "getccdata"
)

// PolicyProvider implements chaincode policy provider
type PolicyProvider struct {
	config apiconfig.Config
}

// NewCCPolicyProvider returns chaincode policy provider
func NewCCPolicyProvider(config apiconfig.Config) (*PolicyProvider, error) {
	return &PolicyProvider{config: config}, nil
}

// NewCCPolicyService creates new chaincode policy service
func (dp *PolicyProvider) NewCCPolicyService(client fab.FabricClient) (fab.CCPolicyService, error) {
	return &ccPolicyService{client: client, ccDataMap: make(map[string]*ccprovider.ChaincodeData)}, nil
}

type ccPolicyService struct {
	client    fab.FabricClient
	ccDataMap map[string]*ccprovider.ChaincodeData
	mutex     sync.RWMutex
}

func (dp *ccPolicyService) GetChaincodePolicy(channelID string, chaincodeID string) (*common.SignaturePolicyEnvelope, error) {
	key := newResolverKey(channelID, chaincodeID)
	var ccData *ccprovider.ChaincodeData

	dp.mutex.RLock()
	ccData = dp.ccDataMap[key.String()]
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

func (dp *ccPolicyService) clearCache() {
	dp.mutex.Lock()
	defer dp.mutex.Unlock()
	dp.ccDataMap = make(map[string]*ccprovider.ChaincodeData)
}

type resolverKey struct {
	channelID    string
	chaincodeIDs []string
	key          string
}

func (k *resolverKey) String() string {
	return k.key
}

func newResolverKey(channelID string, chaincodeIDs ...string) *resolverKey {
	arr := chaincodeIDs[:]
	sort.Strings(arr)

	key := channelID + "-"
	for i, s := range arr {
		key += s
		if i+1 < len(arr) {
			key += ":"
		}
	}
	return &resolverKey{channelID: channelID, chaincodeIDs: arr, key: key}
}

func (dp *ccPolicyService) queryChaincode(channelID string, ccID string, ccFcn string, ccArgs [][]byte) (*apitxn.TransactionProposalResponse, error) {
	logger.Debugf("queryChaincode channelID:%s", channelID)

	channel := dp.client.Channel(channelID)

	if channel == nil {
		var err error
		channel, err = dp.client.NewChannel(channelID)
		if err != nil {
			return nil, err
		}
	}

	chPeers, err := dp.client.Config().ChannelPeers(channel.Name())
	if err != nil {
		return nil, errors.WithMessage(err, "unable to read configuration for channel peers")
	}

	var queryErrors []string
	var response *apitxn.TransactionProposalResponse
	for _, p := range chPeers {

		serverHostOverride := ""
		if str, ok := p.GRPCOptions["ssl-target-name-override"].(string); ok {
			serverHostOverride = str
		}

		peer, err := peerImpl.NewPeerTLSFromCert(p.URL, p.TLSCACerts.Path, serverHostOverride, dp.client.Config())
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

		responses, _, err := channel.SendTransactionProposal(request)
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
	if len(queryErrors) == len(chPeers) {
		errMsg := fmt.Sprintf("Error querying peers for channel %s: %s", channelID, strings.Join(queryErrors, "\n"))
		return nil, errors.New(errMsg)
	}

	return response, nil
}
