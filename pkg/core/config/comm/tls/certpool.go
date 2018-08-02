/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package tls

import (
	"crypto/x509"
	"sync"

	"sync/atomic"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/logging"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
)

var logger = logging.NewLogger("fabsdk/core")

// certPool is a thread safe wrapper around the x509 standard library
// cert pool implementation.
// It optionally allows loading the system trust store.
type certPool struct {
	certPool    *x509.CertPool
	certs       []*x509.Certificate
	certsByName map[string][]int
	lock        sync.Mutex
	dirty       int32
	certQueue   []*x509.Certificate
	orgsInPool  map[string]bool
	rwlock      sync.RWMutex
}

// NewCertPool new CertPool implementation
func NewCertPool(useSystemCertPool bool) (fab.CertPool, error) {

	c, err := loadSystemCertPool(useSystemCertPool)
	if err != nil {
		return nil, err
	}

	newCertPool := &certPool{
		certsByName: make(map[string][]int),
		certPool:    c,
		orgsInPool:  make(map[string]bool),
	}

	return newCertPool, nil
}

//Get returns certpool
//if there are any certs in cert queue added by any previous Add() call, it adds those certs to certpool before returning
func (c *certPool) Get() (*x509.CertPool, error) {

	//if dirty then add certs from queue to cert pool
	if atomic.CompareAndSwapInt32(&c.dirty, 1, 0) {

		c.lock.Lock()
		defer c.lock.Unlock()

		//add all new certs in queue to cert pool
		for _, cert := range c.certQueue {
			c.certPool.AddCert(cert)
		}
		c.certQueue = []*x509.Certificate{}
	}

	return c.certPool, nil
}

//Add adds given certs to cert pool queue, those certs will be added to certpool during subsequent Get() call
func (c *certPool) Add(certs ...*x509.Certificate) {
	if len(certs) == 0 {
		return
	}

	c.lock.Lock()
	defer c.lock.Unlock()

	//filter certs to be added, check if they already exist or duplicate
	certsToBeAdded := c.filterCerts(certs...)

	if len(certsToBeAdded) > 0 {

		for _, newCert := range certsToBeAdded {
			c.certQueue = append(c.certQueue, newCert)
			// Store cert name index
			name := string(newCert.RawSubject)
			c.certsByName[name] = append(c.certsByName[name], len(c.certs))
			// Store cert
			c.certs = append(c.certs, newCert)
		}

		atomic.CompareAndSwapInt32(&c.dirty, 0, 1)
	}
}

func (c *certPool) IsOrgAdded(mspID string) bool {
	c.rwlock.RLock()
	defer c.rwlock.RUnlock()
	return c.orgsInPool[mspID]
}

func (c *certPool) AddOrg(mspID string) {
	c.rwlock.Lock()
	defer c.rwlock.Unlock()
	c.orgsInPool[mspID] = true
}

//filterCerts remove certs from list if they already exist in pool or duplicate
func (c *certPool) filterCerts(certs ...*x509.Certificate) []*x509.Certificate {
	filtered := []*x509.Certificate{}

CertLoop:
	for _, cert := range certs {
		if cert == nil {
			continue
		}
		possibilities := c.certsByName[string(cert.RawSubject)]
		for _, p := range possibilities {
			if c.certs[p].Equal(cert) {
				continue CertLoop
			}
		}
		filtered = append(filtered, cert)
	}

	//remove duplicate from list of certs being passed
	return removeDuplicates(filtered...)
}

func removeDuplicates(certs ...*x509.Certificate) []*x509.Certificate {
	encountered := map[*x509.Certificate]bool{}
	result := []*x509.Certificate{}

	for v := range certs {
		if !encountered[certs[v]] {
			encountered[certs[v]] = true
			result = append(result, certs[v])
		}
	}
	return result
}

func loadSystemCertPool(useSystemCertPool bool) (*x509.CertPool, error) {
	if !useSystemCertPool {
		return x509.NewCertPool(), nil
	}
	systemCertPool, err := x509.SystemCertPool()
	if err != nil {
		return nil, err
	}
	logger.Debugf("Loaded system cert pool of size: %d", len(systemCertPool.Subjects()))

	return systemCertPool, nil
}
