/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
	idapi "github.com/hyperledger/fabric-sdk-go/pkg/context/api/identity"
	"github.com/pkg/errors"
)

// MockRegistrarService is a mock Registrar Service
type MockRegistrarService struct {
}

// NewMockRegistrarService Constructor for a registrar service.
func NewMockRegistrarService(orgName string, cryptoProvider core.CryptoSuite, config core.Config) (idapi.RegistrarService, error) {
	mcm := MockRegistrarService{}
	return &mcm, nil
}

// Register registers a user with a Fabric network
func (mgr *MockRegistrarService) Register(request *idapi.RegistrationRequest) (string, error) {
	return "", errors.New("not implemented")
}

// Revoke revokes a user
func (mgr *MockRegistrarService) Revoke(request *idapi.RevocationRequest) (*idapi.RevocationResponse, error) {
	return nil, errors.New("not implemented")
}

// CAName return the name of a CA associated with this identity manager
func (mgr *MockRegistrarService) CAName() string {
	return ""
}
