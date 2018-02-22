/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package identitymgr

import (
	caapi "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric-ca/api"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/fab"
	"github.com/pkg/errors"
)

// Register a User with the Fabric CA
// request: Registration Request
// Returns Enrolment Secret
func (mgr *IdentityManager) Register(request *fab.RegistrationRequest) (string, error) {
	if err := mgr.initCAClient(); err != nil {
		return "", err
	}
	// Validate registration request
	if request == nil {
		return "", errors.New("registration request is required")
	}
	if request.Name == "" {
		return "", errors.New("request.Name is required")
	}
	registrar, err := mgr.getRegistrar()
	if err != nil {
		return "", errors.Wrapf(err, "failed to get registrar")
	}
	// Create request signing identity
	identity, err := mgr.createSigningIdentity(registrar)
	if err != nil {
		return "", errors.Wrap(err, "failed to create request for signing identity")
	}
	// Contruct request for Fabric CA client
	var attributes []caapi.Attribute
	for i := range request.Attributes {
		attributes = append(attributes, caapi.Attribute{Name: request.
			Attributes[i].Key, Value: request.Attributes[i].Value})
	}
	var req = caapi.RegistrationRequest{
		CAName:         request.CAName,
		Name:           request.Name,
		Type:           request.Type,
		MaxEnrollments: request.MaxEnrollments,
		Affiliation:    request.Affiliation,
		Secret:         request.Secret,
		Attributes:     attributes}
	// Make registration request
	response, err := identity.Register(&req)
	if err != nil {
		return "", errors.Wrap(err, "failed to register user")
	}

	return response.Secret, nil
}

// Revoke a User with the Fabric CA
// registrar: The User that is initiating the revocation
// request: Revocation Request
func (mgr *IdentityManager) Revoke(request *fab.RevocationRequest) (*fab.RevocationResponse, error) {
	if err := mgr.initCAClient(); err != nil {
		return nil, err
	}
	// Validate revocation request
	if request == nil {
		return nil, errors.New("revocation request is required")
	}
	registrar, err := mgr.getRegistrar()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to ret registrar")
	}
	// Create request signing identity
	identity, err := mgr.createSigningIdentity(registrar)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create request for signing identity")
	}
	// Create revocation request
	var req = caapi.RevocationRequest{
		CAName: request.CAName,
		Name:   request.Name,
		Serial: request.Serial,
		AKI:    request.AKI,
		Reason: request.Reason,
	}

	resp, err := identity.Revoke(&req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to revoke")
	}
	var revokedCerts []fab.RevokedCert
	for i := range resp.RevokedCerts {
		revokedCerts = append(
			revokedCerts,
			fab.RevokedCert{
				Serial: resp.RevokedCerts[i].Serial,
				AKI:    resp.RevokedCerts[i].AKI,
			})
	}

	// TODO complete the response mapping
	return &fab.RevocationResponse{
		RevokedCerts: revokedCerts,
		CRL:          resp.CRL,
	}, nil
}

func (mgr *IdentityManager) getRegistrar() (api.User, error) {
	user, err := mgr.userStore.Load(api.UserKey{MspID: mgr.mspID, Name: mgr.registrar.EnrollID})
	if err != nil {
		if err != api.ErrUserNotFound {
			return nil, err
		}
		if mgr.registrar.EnrollSecret == "" {
			return nil, errors.New("registrar not found and cannot be enrolled because enrollment secret is not present")
		}
		_, _, err = mgr.Enroll(mgr.registrar.EnrollID, mgr.registrar.EnrollSecret)
		if err != nil {
			return nil, err
		}
		user, err = mgr.userStore.Load(api.UserKey{MspID: mgr.mspID, Name: mgr.registrar.EnrollID})
	}
	return user, err
}
