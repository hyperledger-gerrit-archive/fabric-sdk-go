/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package config

import (
	"crypto/tls"
	"crypto/x509"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/status"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config/cryptoutil"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config/endpoint"
	"github.com/pkg/errors"

	"regexp"

	"sync"

	"io/ioutil"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	cs "github.com/hyperledger/fabric-sdk-go/pkg/core/cryptosuite"
	"github.com/hyperledger/fabric-sdk-go/pkg/util/pathvar"
)

const (
	cmdRoot                        = "FABRIC_SDK"
	defaultTimeout                 = time.Second * 5
	defaultConnIdleTimeout         = time.Second * 30
	defaultCacheSweepInterval      = time.Second * 15
	defaultEventServiceIdleTimeout = time.Minute * 2
	defaultResMgmtTimeout          = time.Second * 180
	defaultExecuteTimeout          = time.Second * 180
)

// EndpointConfig represents the endpoint configuration for the client
type EndpointConfig struct {
	backend             *Backend
	tlsCerts            []*x509.Certificate
	networkConfig       *fab.NetworkConfig
	networkConfigCached bool
	peerMatchers        map[int]*regexp.Regexp
	ordererMatchers     map[int]*regexp.Regexp
	caMatchers          map[int]*regexp.Regexp
	certPoolLock        sync.Mutex
}

// TimeoutOrDefault reads timeouts for the given timeout type, if not found, defaultTimeout is returned
func (c *EndpointConfig) TimeoutOrDefault(tType fab.TimeoutType) time.Duration {
	timeout := c.getTimeout(tType)
	if timeout == 0 {
		timeout = defaultTimeout
	}

	return timeout
}

// Timeout reads timeouts for the given timeout type, the default is 0 if type is not found in config
func (c *EndpointConfig) Timeout(tType fab.TimeoutType) time.Duration {
	return c.getTimeout(tType)
}

// MSPID returns the MSP ID for the requested organization
func (c *EndpointConfig) MSPID(org string) (string, error) {
	config, err := c.NetworkConfig()
	if err != nil {
		return "", err
	}
	// viper lowercases all key maps, org is lower case
	mspID := config.Organizations[strings.ToLower(org)].MSPID
	if mspID == "" {
		return "", errors.Errorf("MSP ID is empty for org: %s", org)
	}

	return mspID, nil
}

// PeerMSPID returns msp that peer belongs to
func (c *EndpointConfig) PeerMSPID(name string) (string, error) {
	netConfig, err := c.NetworkConfig()
	if err != nil {
		return "", err
	}

	var mspID string

	// Find organisation/msp that peer belongs to
	for _, org := range netConfig.Organizations {
		for i := 0; i < len(org.Peers); i++ {
			if strings.EqualFold(org.Peers[i], name) {
				// peer belongs to this org add org msp
				mspID = org.MSPID
				break
			} else {
				peer, err := c.findMatchingPeer(org.Peers[i])
				if err == nil && strings.EqualFold(peer, name) {
					mspID = org.MSPID
					break
				}
			}
		}
	}

	return mspID, nil

}

// OrderersConfig returns a list of defined orderers
func (c *EndpointConfig) OrderersConfig() ([]fab.OrdererConfig, error) {
	orderers := []fab.OrdererConfig{}
	config, err := c.NetworkConfig()
	if err != nil {
		return nil, err
	}

	for _, orderer := range config.Orderers {

		if orderer.TLSCACerts.Path != "" {
			orderer.TLSCACerts.Path = pathvar.Subst(orderer.TLSCACerts.Path)
		} else if len(orderer.TLSCACerts.Pem) == 0 && c.backend.getBool("client.tlsCerts.systemCertPool") == false {
			errors.Errorf("Orderer has no certs configured. Make sure TLSCACerts.Pem or TLSCACerts.Path is set for %s", orderer.URL)
		}

		orderers = append(orderers, orderer)
	}

	return orderers, nil
}

// RandomOrdererConfig returns a pseudo-random orderer from the network config
func (c *EndpointConfig) RandomOrdererConfig() (*fab.OrdererConfig, error) {
	orderers, err := c.OrderersConfig()
	if err != nil {
		return nil, err
	}

	return randomOrdererConfig(orderers)
}

// OrdererConfig returns the requested orderer
func (c *EndpointConfig) OrdererConfig(name string) (*fab.OrdererConfig, error) {
	config, err := c.NetworkConfig()
	if err != nil {
		return nil, err
	}
	orderer, ok := config.Orderers[strings.ToLower(name)]
	if !ok {
		logger.Debugf("Could not find Orderer for [%s], trying with Entity Matchers", name)
		matchingOrdererConfig, matchErr := c.tryMatchingOrdererConfig(strings.ToLower(name))
		if matchErr != nil {
			return nil, errors.WithMessage(matchErr, "unable to find Orderer Config")
		}
		logger.Debugf("Found matching Orderer Config for [%s]", name)
		orderer = *matchingOrdererConfig
	}

	if orderer.TLSCACerts.Path != "" {
		orderer.TLSCACerts.Path = pathvar.Subst(orderer.TLSCACerts.Path)
	}

	return &orderer, nil
}

// PeersConfig Retrieves the fabric peers for the specified org from the
// config file provided
func (c *EndpointConfig) PeersConfig(org string) ([]fab.PeerConfig, error) {
	config, err := c.NetworkConfig()
	if err != nil {
		return nil, err
	}

	peersConfig := config.Organizations[strings.ToLower(org)].Peers
	peers := []fab.PeerConfig{}

	for _, peerName := range peersConfig {
		p := config.Peers[strings.ToLower(peerName)]
		if err = c.verifyPeerConfig(p, peerName, endpoint.IsTLSEnabled(p.URL)); err != nil {
			logger.Debugf("Could not verify Peer for [%s], trying with Entity Matchers", peerName)
			matchingPeerConfig, matchErr := c.tryMatchingPeerConfig(peerName)
			if matchErr != nil {
				return nil, errors.WithMessage(err, "unable to find Peer Config")
			}
			logger.Debugf("Found a matchingPeerConfig for [%s]", peerName)
			p = *matchingPeerConfig
		}
		if p.TLSCACerts.Path != "" {
			p.TLSCACerts.Path = pathvar.Subst(p.TLSCACerts.Path)
		}

		peers = append(peers, p)
	}
	return peers, nil
}

// PeerConfig Retrieves a specific peer from the configuration by org and name
func (c *EndpointConfig) PeerConfig(org string, name string) (*fab.PeerConfig, error) {
	config, err := c.NetworkConfig()
	if err != nil {
		return nil, err
	}

	peersConfig := config.Organizations[strings.ToLower(org)].Peers
	peerInOrg := false
	for _, p := range peersConfig {
		if p == name {
			peerInOrg = true
		}
	}
	if !peerInOrg {
		return nil, errors.Errorf("peer %s is not part of organization %s", name, org)
	}

	peerConfig, ok := config.Peers[strings.ToLower(name)]
	if !ok {
		logger.Debugf("Could not find Peer for [%s], trying with Entity Matchers", name)
		matchingPeerConfig, matchErr := c.tryMatchingPeerConfig(strings.ToLower(name))
		if matchErr != nil {
			return nil, errors.WithMessage(matchErr, "unable to find peer config")
		}
		logger.Debugf("Found MatchingPeerConfig for [%s]", name)
		peerConfig = *matchingPeerConfig
	}

	if peerConfig.TLSCACerts.Path != "" {
		peerConfig.TLSCACerts.Path = pathvar.Subst(peerConfig.TLSCACerts.Path)
	}
	return &peerConfig, nil
}

// PeerConfigByURL retrieves PeerConfig by URL
func (c *EndpointConfig) PeerConfigByURL(url string) (*fab.PeerConfig, error) {
	config, err := c.NetworkConfig()
	if err != nil {
		return nil, err
	}

	var matchPeerConfig *fab.PeerConfig
	staticPeers := config.Peers
	for _, staticPeerConfig := range staticPeers {
		if strings.EqualFold(staticPeerConfig.URL, url) {
			matchPeerConfig = &staticPeerConfig
			break
		}
	}

	if matchPeerConfig == nil {
		// try to match from entity matchers
		logger.Debugf("Could not find Peer for url [%s], trying with Entity Matchers", url)
		matchPeerConfig, err = c.tryMatchingPeerConfig(url)
		if err != nil {
			return nil, errors.WithMessage(err, "No Peer found with the url from config")
		}
		logger.Debugf("Found MatchingPeerConfig for url [%s]", url)
	}

	if matchPeerConfig != nil && matchPeerConfig.TLSCACerts.Path != "" {
		matchPeerConfig.TLSCACerts.Path = pathvar.Subst(matchPeerConfig.TLSCACerts.Path)
	}

	return matchPeerConfig, nil
}

// NetworkConfig returns the network configuration defined in the config file
func (c *EndpointConfig) NetworkConfig() (*fab.NetworkConfig, error) {
	if c.networkConfigCached {
		return c.networkConfig, nil
	}

	if err := c.cacheNetworkConfiguration(); err != nil {
		return nil, errors.WithMessage(err, "network configuration load failed")
	}
	return c.networkConfig, nil
}

// NetworkPeers returns the network peers configuration
func (c *EndpointConfig) NetworkPeers() ([]fab.NetworkPeer, error) {
	netConfig, err := c.NetworkConfig()
	if err != nil {
		return nil, err
	}

	netPeers := []fab.NetworkPeer{}

	for name, p := range netConfig.Peers {

		if err = c.verifyPeerConfig(p, name, endpoint.IsTLSEnabled(p.URL)); err != nil {
			return nil, err
		}

		if p.TLSCACerts.Path != "" {
			p.TLSCACerts.Path = pathvar.Subst(p.TLSCACerts.Path)
		}

		mspID, err := c.PeerMSPID(name)
		if err != nil {
			return nil, errors.Errorf("failed to retrieve msp id for peer %s", name)
		}

		netPeer := fab.NetworkPeer{PeerConfig: p, MSPID: mspID}
		netPeers = append(netPeers, netPeer)
	}

	return netPeers, nil
}

// ChannelConfig returns the channel configuration
func (c *EndpointConfig) ChannelConfig(name string) (*fab.ChannelNetworkConfig, error) {
	config, err := c.NetworkConfig()
	if err != nil {
		return nil, err
	}

	// viper lowercases all key maps
	ch, ok := config.Channels[strings.ToLower(name)]
	if !ok {
		return nil, nil
	}

	return &ch, nil
}

// ChannelPeers returns the channel peers configuration
func (c *EndpointConfig) ChannelPeers(name string) ([]fab.ChannelPeer, error) {
	netConfig, err := c.NetworkConfig()
	if err != nil {
		return nil, err
	}

	// viper lowercases all key maps
	chConfig, ok := netConfig.Channels[strings.ToLower(name)]
	if !ok {
		return nil, errors.Errorf("channel config not found for %s", name)
	}

	peers := []fab.ChannelPeer{}

	for peerName, chPeerConfig := range chConfig.Peers {

		// Get generic peer configuration
		p, ok := netConfig.Peers[strings.ToLower(peerName)]
		if !ok {
			logger.Debugf("Could not find Peer for [%s], trying with Entity Matchers", peerName)
			matchingPeerConfig, matchErr := c.tryMatchingPeerConfig(strings.ToLower(peerName))
			if matchErr != nil {
				return nil, errors.Errorf("peer config not found for %s", peerName)
			}
			logger.Debugf("Found matchingPeerConfig for [%s]", peerName)
			p = *matchingPeerConfig
		}

		if err = c.verifyPeerConfig(p, peerName, endpoint.IsTLSEnabled(p.URL)); err != nil {
			return nil, err
		}

		if p.TLSCACerts.Path != "" {
			p.TLSCACerts.Path = pathvar.Subst(p.TLSCACerts.Path)
		}

		mspID, err := c.PeerMSPID(peerName)
		if err != nil {
			return nil, errors.Errorf("failed to retrieve msp id for peer %s", peerName)
		}

		networkPeer := fab.NetworkPeer{PeerConfig: p, MSPID: mspID}

		peer := fab.ChannelPeer{PeerChannelConfig: chPeerConfig, NetworkPeer: networkPeer}

		peers = append(peers, peer)
	}

	return peers, nil

}

// ChannelOrderers returns a list of channel orderers
func (c *EndpointConfig) ChannelOrderers(name string) ([]fab.OrdererConfig, error) {
	orderers := []fab.OrdererConfig{}
	channel, err := c.ChannelConfig(name)
	if err != nil || channel == nil {
		return nil, errors.Errorf("Unable to retrieve channel config: %s", err)
	}

	for _, chOrderer := range channel.Orderers {
		orderer, err := c.OrdererConfig(chOrderer)
		if err != nil || orderer == nil {
			return nil, errors.Errorf("unable to retrieve orderer config: %s", err)
		}

		orderers = append(orderers, *orderer)
	}

	return orderers, nil
}

// TLSCACertPool returns the configured cert pool. If a certConfig
// is provided, the certficate is added to the pool
func (c *EndpointConfig) TLSCACertPool(certs ...*x509.Certificate) (*x509.CertPool, error) {

	c.certPoolLock.Lock()
	defer c.certPoolLock.Unlock()

	//add cert if it is not nil and doesn't exists already
	for _, newCert := range certs {
		if newCert != nil && !c.containsCert(newCert) {
			c.tlsCerts = append(c.tlsCerts, newCert)
		}
	}

	//get new cert pool
	tlsCertPool, err := c.getCertPool()
	if err != nil {
		return nil, errors.WithMessage(err, "failed to create cert pool")
	}

	//add all tls ca certs to cert pool
	for _, cert := range c.tlsCerts {
		tlsCertPool.AddCert(cert)
	}

	return tlsCertPool, nil
}

// EventServiceType returns the type of event service client to use
func (c *EndpointConfig) EventServiceType() fab.EventServiceType {
	etype := c.backend.getString("client.eventService.type")
	switch etype {
	case "eventhub":
		return fab.EventHubEventServiceType
	default:
		return fab.DeliverEventServiceType
	}
}

// TLSClientCerts loads the client's certs for mutual TLS
// It checks the config for embedded pem files before looking for cert files
func (c *EndpointConfig) TLSClientCerts() ([]tls.Certificate, error) {
	clientConfig, err := c.client()
	if err != nil {
		return nil, err
	}
	var clientCerts tls.Certificate
	var cb, kb []byte
	cb, err = clientConfig.TLSCerts.Client.Cert.Bytes()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to load tls client cert")
	}

	if len(cb) == 0 {
		// if no cert found in the config, return empty cert chain
		return []tls.Certificate{clientCerts}, nil
	}

	// Load private key from cert using default crypto suite
	cs := cs.GetDefault()
	pk, err := cryptoutil.GetPrivateKeyFromCert(cb, cs)

	// If CryptoSuite fails to load private key from cert then load private key from config
	if err != nil || pk == nil {
		logger.Debugf("Reading pk from config, unable to retrieve from cert: %s", err)
		if clientConfig.TLSCerts.Client.Key.Pem != "" {
			kb = []byte(clientConfig.TLSCerts.Client.Key.Pem)
		} else if clientConfig.TLSCerts.Client.Key.Path != "" {
			kb, err = loadByteKeyOrCertFromFile(clientConfig, true)
			if err != nil {
				return nil, errors.Wrapf(err, "Failed to load key from file path '%s'", clientConfig.TLSCerts.Client.Key.Path)
			}
		}

		// load the key/cert pair from []byte
		clientCerts, err = tls.X509KeyPair(cb, kb)
		if err != nil {
			return nil, errors.Errorf("Error loading cert/key pair as TLS client credentials: %v", err)
		}

		logger.Debug("pk read from config successfully")

		return []tls.Certificate{clientCerts}, nil

	}

	// private key was retrieved from cert
	clientCerts, err = cryptoutil.X509KeyPair(cb, pk, cs)
	if err != nil {
		return nil, err
	}

	return []tls.Certificate{clientCerts}, nil
}

// CryptoConfigPath ...
func (c *EndpointConfig) CryptoConfigPath() string {
	return pathvar.Subst(c.backend.getString("client.cryptoconfig.path"))
}

func (c *EndpointConfig) getTimeout(tType fab.TimeoutType) time.Duration {
	var timeout time.Duration
	switch tType {
	case fab.EndorserConnection:
		timeout = c.backend.getDuration("client.peer.timeout.connection")
	case fab.Query:
		timeout = c.backend.getDuration("client.global.timeout.query")
	case fab.Execute:
		timeout = c.backend.getDuration("client.global.timeout.execute")
		if timeout == 0 {
			timeout = defaultExecuteTimeout
		}
	case fab.DiscoveryGreylistExpiry:
		timeout = c.backend.getDuration("client.peer.timeout.discovery.greylistExpiry")
	case fab.PeerResponse:
		timeout = c.backend.getDuration("client.peer.timeout.response")
	case fab.EventHubConnection:
		timeout = c.backend.getDuration("client.eventService.timeout.connection")
	case fab.EventReg:
		timeout = c.backend.getDuration("client.eventService.timeout.registrationResponse")
	case fab.OrdererConnection:
		timeout = c.backend.getDuration("client.orderer.timeout.connection")
	case fab.OrdererResponse:
		timeout = c.backend.getDuration("client.orderer.timeout.response")
	case fab.ChannelConfigRefresh:
		timeout = c.backend.getDuration("client.global.cache.channelConfig")
	case fab.ChannelMembershipRefresh:
		timeout = c.backend.getDuration("client.global.cache.channelMembership")
	case fab.CacheSweepInterval: // EXPERIMENTAL - do we need this to be configurable?
		timeout = c.backend.getDuration("client.cache.interval.sweep")
		if timeout == 0 {
			timeout = defaultCacheSweepInterval
		}
	case fab.ConnectionIdle:
		timeout = c.backend.getDuration("client.global.cache.connectionIdle")
		if timeout == 0 {
			timeout = defaultConnIdleTimeout
		}
	case fab.EventServiceIdle:
		timeout = c.backend.getDuration("client.global.cache.eventServiceIdle")
		if timeout == 0 {
			timeout = defaultEventServiceIdleTimeout
		}
	case fab.ResMgmt:
		timeout = c.backend.getDuration("client.global.timeout.resmgmt")
		if timeout == 0 {
			timeout = defaultResMgmtTimeout
		}
	}

	return timeout
}

func (c *EndpointConfig) cacheNetworkConfiguration() error {
	networkConfig := fab.NetworkConfig{}
	networkConfig.Name = c.backend.getString("name")
	networkConfig.Description = c.backend.getString("description")
	networkConfig.Version = c.backend.getString("version")

	ok := c.backend.unmarshalKey("client", &networkConfig.Client)
	logger.Debugf("Client is: %+v", networkConfig.Client)
	if !ok {
		return errors.New("failed to parse 'client' config item to networkConfig.Client type")
	}

	ok = c.backend.unmarshalKey("channels", &networkConfig.Channels)
	logger.Debugf("channels are: %+v", networkConfig.Channels)
	if !ok {
		return errors.New("failed to parse 'channels' config item to networkConfig.Channels type")
	}

	ok = c.backend.unmarshalKey("organizations", &networkConfig.Organizations)
	logger.Debugf("organizations are: %+v", networkConfig.Organizations)
	if !ok {
		return errors.New("failed to parse 'organizations' config item to networkConfig.Organizations type")
	}

	ok = c.backend.unmarshalKey("orderers", &networkConfig.Orderers)
	logger.Debugf("orderers are: %+v", networkConfig.Orderers)
	if !ok {
		return errors.New("failed to parse 'orderers' config item to networkConfig.Orderers type")
	}

	ok = c.backend.unmarshalKey("peers", &networkConfig.Peers)
	logger.Debugf("peers are: %+v", networkConfig.Peers)
	if !ok {
		return errors.New("failed to parse 'peers' config item to networkConfig.Peers type")
	}

	ok = c.backend.unmarshalKey("certificateAuthorities", &networkConfig.CertificateAuthorities)
	logger.Debugf("certificateAuthorities are: %+v", networkConfig.CertificateAuthorities)
	if !ok {
		return errors.New("failed to parse 'certificateAuthorities' config item to networkConfig.CertificateAuthorities type")
	}

	ok = c.backend.unmarshalKey("entityMatchers", &networkConfig.EntityMatchers)
	logger.Debugf("Matchers are: %+v", networkConfig.EntityMatchers)
	if !ok {
		return errors.New("failed to parse 'entityMatchers' config item to networkConfig.EntityMatchers type")
	}

	c.networkConfig = &networkConfig
	c.networkConfigCached = true
	return nil
}

// randomOrdererConfig returns a pseudo-random orderer from the list of orderers
func randomOrdererConfig(orderers []fab.OrdererConfig) (*fab.OrdererConfig, error) {

	rs := rand.NewSource(time.Now().Unix())
	r := rand.New(rs)
	randomNumber := r.Intn(len(orderers))

	return &orderers[randomNumber], nil
}

func (c *EndpointConfig) getPortIfPresent(url string) (int, bool) {
	s := strings.Split(url, ":")
	if len(s) > 1 {
		if port, err := strconv.Atoi(s[len(s)-1]); err == nil {
			return port, true
		}
	}
	return 0, false
}

func (c *EndpointConfig) tryMatchingPeerConfig(peerName string) (*fab.PeerConfig, error) {
	networkConfig, err := c.NetworkConfig()
	if err != nil {
		return nil, err
	}
	//Return if no peerMatchers are configured
	if len(c.peerMatchers) == 0 {
		return nil, errors.New("no Peer entityMatchers are found")
	}

	//sort the keys
	var keys []int
	for k := range c.peerMatchers {
		keys = append(keys, k)
	}
	sort.Ints(keys)

	//loop over peerentityMatchers to find the matching peer
	for _, k := range keys {
		v := c.peerMatchers[k]
		if v.MatchString(peerName) {
			// get the matching matchConfig from the index number
			peerMatchConfig := networkConfig.EntityMatchers["peer"][k]
			//Get the peerConfig from mapped host
			peerConfig, ok := networkConfig.Peers[strings.ToLower(peerMatchConfig.MappedHost)]
			if !ok {
				return nil, errors.New("failed to load config from matched Peer")
			}

			// Make a copy of GRPC options (as it is manipulated below)
			peerConfig.GRPCOptions = copyPropertiesMap(peerConfig.GRPCOptions)

			_, isPortPresentInPeerName := c.getPortIfPresent(peerName)
			//if substitution url is empty, use the same network peer url
			if peerMatchConfig.URLSubstitutionExp == "" {
				port, isPortPresent := c.getPortIfPresent(peerConfig.URL)
				peerConfig.URL = peerName
				//append port of matched config
				if isPortPresent && !isPortPresentInPeerName {
					peerConfig.URL += ":" + strconv.Itoa(port)
				}
			} else {
				//else, replace url with urlSubstitutionExp if it doesnt have any variable declarations like $
				if strings.Index(peerMatchConfig.URLSubstitutionExp, "$") < 0 {
					peerConfig.URL = peerMatchConfig.URLSubstitutionExp
				} else {
					//if the urlSubstitutionExp has $ variable declarations, use regex replaceallstring to replace networkhostname with substituionexp pattern
					peerConfig.URL = v.ReplaceAllString(peerName, peerMatchConfig.URLSubstitutionExp)
				}

			}

			//if eventSubstitution url is empty, use the same network peer url
			if peerMatchConfig.EventURLSubstitutionExp == "" {
				port, isPortPresent := c.getPortIfPresent(peerConfig.EventURL)
				peerConfig.EventURL = peerName
				//append port of matched config
				if isPortPresent && !isPortPresentInPeerName {
					peerConfig.EventURL += ":" + strconv.Itoa(port)
				}
			} else {
				//else, replace url with eventUrlSubstitutionExp if it doesnt have any variable declarations like $
				if strings.Index(peerMatchConfig.EventURLSubstitutionExp, "$") < 0 {
					peerConfig.EventURL = peerMatchConfig.EventURLSubstitutionExp
				} else {
					//if the eventUrlSubstitutionExp has $ variable declarations, use regex replaceallstring to replace networkhostname with eventsubstituionexp pattern
					peerConfig.EventURL = v.ReplaceAllString(peerName, peerMatchConfig.EventURLSubstitutionExp)
				}

			}

			//if sslTargetOverrideUrlSubstitutionExp is empty, use the same network peer host
			if peerMatchConfig.SSLTargetOverrideURLSubstitutionExp == "" {
				if strings.Index(peerName, ":") < 0 {
					peerConfig.GRPCOptions["ssl-target-name-override"] = peerName
				} else {
					//Remove port and protocol of the peerName
					s := strings.Split(peerName, ":")
					if isPortPresentInPeerName {
						peerConfig.GRPCOptions["ssl-target-name-override"] = s[len(s)-2]
					} else {
						peerConfig.GRPCOptions["ssl-target-name-override"] = s[len(s)-1]
					}
				}

			} else {
				//else, replace url with sslTargetOverrideUrlSubstitutionExp if it doesnt have any variable declarations like $
				if strings.Index(peerMatchConfig.SSLTargetOverrideURLSubstitutionExp, "$") < 0 {
					peerConfig.GRPCOptions["ssl-target-name-override"] = peerMatchConfig.SSLTargetOverrideURLSubstitutionExp
				} else {
					//if the sslTargetOverrideUrlSubstitutionExp has $ variable declarations, use regex replaceallstring to replace networkhostname with eventsubstituionexp pattern
					peerConfig.GRPCOptions["ssl-target-name-override"] = v.ReplaceAllString(peerName, peerMatchConfig.SSLTargetOverrideURLSubstitutionExp)
				}

			}
			return &peerConfig, nil
		}
	}

	return nil, errors.WithStack(status.New(status.ClientStatus, status.NoMatchingPeerEntity.ToInt32(), "no matching peer config found", nil))
}

func (c *EndpointConfig) tryMatchingOrdererConfig(ordererName string) (*fab.OrdererConfig, error) {
	networkConfig, err := c.NetworkConfig()
	if err != nil {
		return nil, err
	}
	//Return if no ordererMatchers are configured
	if len(c.ordererMatchers) == 0 {
		return nil, errors.New("no Orderer entityMatchers are found")
	}

	//sort the keys
	var keys []int
	for k := range c.ordererMatchers {
		keys = append(keys, k)
	}
	sort.Ints(keys)

	//loop over ordererentityMatchers to find the matching orderer
	for _, k := range keys {
		v := c.ordererMatchers[k]
		if v.MatchString(ordererName) {
			// get the matching matchConfig from the index number
			ordererMatchConfig := networkConfig.EntityMatchers["orderer"][k]
			//Get the ordererConfig from mapped host
			ordererConfig, ok := networkConfig.Orderers[strings.ToLower(ordererMatchConfig.MappedHost)]
			if !ok {
				return nil, errors.New("failed to load config from matched Orderer")
			}

			// Make a copy of GRPC options (as it is manipulated below)
			ordererConfig.GRPCOptions = copyPropertiesMap(ordererConfig.GRPCOptions)

			_, isPortPresentInOrdererName := c.getPortIfPresent(ordererName)
			//if substitution url is empty, use the same network orderer url
			if ordererMatchConfig.URLSubstitutionExp == "" {
				port, isPortPresent := c.getPortIfPresent(ordererConfig.URL)
				ordererConfig.URL = ordererName

				//append port of matched config
				if isPortPresent && !isPortPresentInOrdererName {
					ordererConfig.URL += ":" + strconv.Itoa(port)
				}
			} else {
				//else, replace url with urlSubstitutionExp if it doesnt have any variable declarations like $
				if strings.Index(ordererMatchConfig.URLSubstitutionExp, "$") < 0 {
					ordererConfig.URL = ordererMatchConfig.URLSubstitutionExp
				} else {
					//if the urlSubstitutionExp has $ variable declarations, use regex replaceallstring to replace networkhostname with substituionexp pattern
					ordererConfig.URL = v.ReplaceAllString(ordererName, ordererMatchConfig.URLSubstitutionExp)
				}
			}

			//if sslTargetOverrideUrlSubstitutionExp is empty, use the same network peer host
			if ordererMatchConfig.SSLTargetOverrideURLSubstitutionExp == "" {
				if strings.Index(ordererName, ":") < 0 {
					ordererConfig.GRPCOptions["ssl-target-name-override"] = ordererName
				} else {
					//Remove port and protocol of the ordererName
					s := strings.Split(ordererName, ":")
					if isPortPresentInOrdererName {
						ordererConfig.GRPCOptions["ssl-target-name-override"] = s[len(s)-2]
					} else {
						ordererConfig.GRPCOptions["ssl-target-name-override"] = s[len(s)-1]
					}
				}

			} else {
				//else, replace url with sslTargetOverrideUrlSubstitutionExp if it doesnt have any variable declarations like $
				if strings.Index(ordererMatchConfig.SSLTargetOverrideURLSubstitutionExp, "$") < 0 {
					ordererConfig.GRPCOptions["ssl-target-name-override"] = ordererMatchConfig.SSLTargetOverrideURLSubstitutionExp
				} else {
					//if the sslTargetOverrideUrlSubstitutionExp has $ variable declarations, use regex replaceallstring to replace networkhostname with eventsubstituionexp pattern
					ordererConfig.GRPCOptions["ssl-target-name-override"] = v.ReplaceAllString(ordererName, ordererMatchConfig.SSLTargetOverrideURLSubstitutionExp)
				}

			}
			return &ordererConfig, nil
		}
	}

	return nil, errors.WithStack(status.New(status.ClientStatus, status.NoMatchingOrdererEntity.ToInt32(), "no matching orderer config found", nil))
}

func copyPropertiesMap(origMap map[string]interface{}) map[string]interface{} {
	newMap := make(map[string]interface{}, len(origMap))
	for k, v := range origMap {
		newMap[k] = v
	}
	return newMap
}

func (c *EndpointConfig) findMatchingPeer(peerName string) (string, error) {
	networkConfig, err := c.NetworkConfig()
	if err != nil {
		return "", err
	}
	//Return if no peerMatchers are configured
	if len(c.peerMatchers) == 0 {
		return "", errors.New("no Peer entityMatchers are found")
	}

	//sort the keys
	var keys []int
	for k := range c.peerMatchers {
		keys = append(keys, k)
	}
	sort.Ints(keys)

	//loop over peerentityMatchers to find the matching peer
	for _, k := range keys {
		v := c.peerMatchers[k]
		if v.MatchString(peerName) {
			// get the matching matchConfig from the index number
			peerMatchConfig := networkConfig.EntityMatchers["peer"][k]
			return peerMatchConfig.MappedHost, nil
		}
	}

	return "", errors.WithStack(status.New(status.ClientStatus, status.NoMatchingPeerEntity.ToInt32(), "no matching peer config found", nil))
}

func (c *EndpointConfig) compileMatchers() error {
	networkConfig, err := c.NetworkConfig()
	if err != nil {
		return err
	}
	//return no error if entityMatchers is not configured
	if networkConfig.EntityMatchers == nil {
		return nil
	}

	if networkConfig.EntityMatchers["peer"] != nil {
		peerMatchersConfig := networkConfig.EntityMatchers["peer"]
		for i := 0; i < len(peerMatchersConfig); i++ {
			if peerMatchersConfig[i].Pattern != "" {
				c.peerMatchers[i], err = regexp.Compile(peerMatchersConfig[i].Pattern)
				if err != nil {
					return err
				}
			}
		}
	}
	if networkConfig.EntityMatchers["orderer"] != nil {
		ordererMatchersConfig := networkConfig.EntityMatchers["orderer"]
		for i := 0; i < len(ordererMatchersConfig); i++ {
			if ordererMatchersConfig[i].Pattern != "" {
				c.ordererMatchers[i], err = regexp.Compile(ordererMatchersConfig[i].Pattern)
				if err != nil {
					return err
				}
			}
		}
	}
	if networkConfig.EntityMatchers["certificateauthorities"] != nil {
		certMatchersConfig := networkConfig.EntityMatchers["certificateauthorities"]
		for i := 0; i < len(certMatchersConfig); i++ {
			if certMatchersConfig[i].Pattern != "" {
				c.caMatchers[i], err = regexp.Compile(certMatchersConfig[i].Pattern)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// PeerConfig Retrieves a specific peer by name
func (c *EndpointConfig) peerConfig(name string) (*fab.PeerConfig, error) {
	config, err := c.NetworkConfig()
	if err != nil {
		return nil, err
	}
	peerConfig, ok := config.Peers[strings.ToLower(name)]
	if !ok {
		logger.Debugf("Could not find PeerConfig for [%s], trying with Entity Matchers", name)
		matchingPeerConfig, matchErr := c.tryMatchingPeerConfig(strings.ToLower(name))
		if matchErr != nil {
			return nil, errors.WithMessage(matchErr, "unable to find peer config")
		}
		logger.Debugf("Found MatchingPeerConfig for [%s]", name)
		peerConfig = *matchingPeerConfig
	}

	if peerConfig.TLSCACerts.Path != "" {
		peerConfig.TLSCACerts.Path = pathvar.Subst(peerConfig.TLSCACerts.Path)
	}
	return &peerConfig, nil
}

func (c *EndpointConfig) verifyPeerConfig(p fab.PeerConfig, peerName string, tlsEnabled bool) error {
	if p.URL == "" {
		return errors.Errorf("URL does not exist or empty for peer %s", peerName)
	}
	if tlsEnabled && len(p.TLSCACerts.Pem) == 0 && p.TLSCACerts.Path == "" && c.backend.getBool("client.tlsCerts.systemCertPool") == false {
		return errors.Errorf("tls.certificate does not exist or empty for peer %s", peerName)
	}
	return nil
}

func (c *EndpointConfig) containsCert(newCert *x509.Certificate) bool {
	//TODO may need to maintain separate map of {cert.RawSubject, cert} to improve performance on search
	for _, cert := range c.tlsCerts {
		if cert.Equal(newCert) {
			return true
		}
	}
	return false
}

func (c *EndpointConfig) getCertPool() (*x509.CertPool, error) {
	tlsCertPool := x509.NewCertPool()
	if c.backend.getBool("client.tlsCerts.systemCertPool") == true {
		var err error
		if tlsCertPool, err = x509.SystemCertPool(); err != nil {
			return nil, err
		}
		logger.Debugf("Loaded system cert pool of size: %d", len(tlsCertPool.Subjects()))
	}
	return tlsCertPool, nil
}

// Client returns the Client config
func (c *EndpointConfig) client() (*msp.ClientConfig, error) {
	config, err := c.NetworkConfig()
	if err != nil {
		return nil, err
	}
	client := config.Client

	client.Organization = strings.ToLower(client.Organization)
	client.TLSCerts.Path = pathvar.Subst(client.TLSCerts.Path)
	client.TLSCerts.Client.Key.Path = pathvar.Subst(client.TLSCerts.Client.Key.Path)
	client.TLSCerts.Client.Cert.Path = pathvar.Subst(client.TLSCerts.Client.Cert.Path)

	return &client, nil
}

func loadByteKeyOrCertFromFile(c *msp.ClientConfig, isKey bool) ([]byte, error) {
	var path string
	a := "key"
	if isKey {
		path = pathvar.Subst(c.TLSCerts.Client.Key.Path)
		c.TLSCerts.Client.Key.Path = path
	} else {
		a = "cert"
		path = pathvar.Subst(c.TLSCerts.Client.Cert.Path)
		c.TLSCerts.Client.Cert.Path = path
	}
	bts, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.Errorf("Error loading %s file from '%s' err: %v", a, path, err)
	}
	return bts, nil
}
