/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package identitymgr

import (
	"github.com/pkg/errors"

	api "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric-ca/api"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/identity"

	contextApi "github.com/hyperledger/fabric-sdk-go/pkg/context/api"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
)

// Enroll a registered user in order to receive a signed X509 certificate.
// enrollmentID The registered ID to use for enrollment
// enrollmentSecret The secret associated with the enrollment ID
// Returns X509 certificate
func (mgr *IdentityManager) Enroll(enrollmentID string, enrollmentSecret string) (core.Key, []byte, error) {
	if err := mgr.initCAClient(); err != nil {
		return nil, nil, err
	}
	if enrollmentID == "" {
		return nil, nil, errors.New("enrollmentID is required")
	}
	if enrollmentSecret == "" {
		return nil, nil, errors.New("enrollmentSecret is required")
	}
	// TODO add attributes
	careq := &api.EnrollmentRequest{
		CAName: mgr.caClient.Config.CAName,
		Name:   enrollmentID,
		Secret: enrollmentSecret,
	}
	caresp, err := mgr.caClient.Enroll(careq)
	if err != nil {
		return nil, nil, errors.Wrap(err, "enroll failed")
	}
	user := identity.NewUser(mgr.mspID, enrollmentID)
	user.SetEnrollmentCertificate(caresp.Identity.GetECert().Cert())
	user.SetPrivateKey(caresp.Identity.GetECert().Key())
	err = mgr.userStore.Store(user)
	if err != nil {
		return nil, nil, errors.Wrap(err, "enroll failed")
	}
	return caresp.Identity.GetECert().Key(), caresp.Identity.GetECert().Cert(), nil
}

// Reenroll an enrolled user in order to receive a signed X509 certificate
// Returns X509 certificate
func (mgr *IdentityManager) Reenroll(user contextApi.User) (core.Key, []byte, error) {
	if err := mgr.initCAClient(); err != nil {
		return nil, nil, err
	}
	if user == nil {
		return nil, nil, errors.New("user required")
	}
	if user.Name() == "" {
		logger.Infof("Invalid re-enroll request, missing argument user")
		return nil, nil, errors.New("user name missing")
	}
	req := &api.ReenrollmentRequest{
		CAName: mgr.caClient.Config.CAName,
	}
	// Create signing identity
	identity, err := mgr.createSigningIdentity(user)
	if err != nil {
		logger.Debugf("Invalid re-enroll request, %s is not a valid user  %s\n", user.Name(), err)
		return nil, nil, errors.Wrap(err, "createSigningIdentity failed")
	}

	reenrollmentResponse, err := identity.Reenroll(req)
	if err != nil {
		return nil, nil, errors.Wrap(err, "reenroll failed")
	}
	return reenrollmentResponse.Identity.GetECert().Key(), reenrollmentResponse.Identity.GetECert().Cert(), nil
}
