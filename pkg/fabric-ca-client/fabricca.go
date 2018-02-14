/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabricca

import (
	"errors"

	config "github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	idapi "github.com/hyperledger/fabric-sdk-go/api/core/identity"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-txn/idmgmtclient"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging"

	"github.com/hyperledger/fabric-sdk-go/api/apicryptosuite"
)

var logger = logging.NewLogger("fabric_sdk_go")

// FabricCA represents a client to Fabric CA.
// Deprecated - use identity.Manager
type FabricCA struct {
	caname   string
	enroller idapi.EnrollmentService
	manager  idapi.Manager
}

// NewFabricCAClient creates a new fabric-ca client
//
// Deprecated - use identity.Enrollment Service and identity.Manager
func NewFabricCAClient(org string, config config.Config, cryptoSuite apicryptosuite.CryptoSuite) (*FabricCA, error) {
	if config == nil {
		return nil, errors.New("must provide config")
	}
	ctx := idmgmtclient.Context{
		MspID:       org,
		Config:      config,
		CryptoSuite: cryptoSuite,
	}
	caConfig, err := config.CAConfig(org)
	if err != nil {
		return nil, err
	}
	mgr, err := idmgmtclient.New(ctx)
	if err != nil {
		return nil, err
	}

	client := FabricCA{
		enroller: mgr,
		manager:  mgr,
		caname:   caConfig.CAName,
	}
	return &client, nil
}

// CAName returns the CA name.
//
// Deprecated
func (fabricCAServices *FabricCA) CAName() string {
	return fabricCAServices.caname
}

// Enroll a registered user in order to receive a signed X509 certificate.
// enrollmentID The registered ID to use for enrollment
// enrollmentSecret The secret associated with the enrollment ID
// Returns X509 certificate
//
// Deprecated - use identity.EnrollmentService
func (fabricCAServices *FabricCA) Enroll(enrollmentID string, enrollmentSecret string, attrs ...idapi.AttributeRequest) (apicryptosuite.Key, []byte, error) {
	return fabricCAServices.enroller.Enroll(enrollmentID, enrollmentSecret, attrs...)
}

// Reenroll an enrolled user in order to receive a signed X509 certificate
// Returns X509 certificate
//
// Deprecated - use identity.EnrollmentService
func (fabricCAServices *FabricCA) Reenroll(user idapi.User) (apicryptosuite.Key, []byte, error) {
	return fabricCAServices.enroller.Reenroll(user)
}

// Register a User with the Fabric CA
// registrar: The User that is initiating the registration
// request: Registration Request
// Returns Enrolment Secret
//
// Deprecated - use identity.Manager
func (fabricCAServices *FabricCA) Register(request *idapi.RegistrationRequest) (string, error) {
	return fabricCAServices.manager.Register(request)
}

// Revoke a User with the Fabric CA
// registrar: The User that is initiating the revocation
// request: Revocation Request
//
// Deprecated - use identity.Manager
func (fabricCAServices *FabricCA) Revoke(request *idapi.RevocationRequest) (*idapi.RevocationResponse, error) {
	return fabricCAServices.manager.Revoke(request)
}
