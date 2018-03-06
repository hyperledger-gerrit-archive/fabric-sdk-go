/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	idapi "github.com/hyperledger/fabric-sdk-go/pkg/context/api/ca"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
	"github.com/pkg/errors"
)

// MockCAService is a mock Client
type MockCAService struct {
}

// NewMockCAService Constructor for a enrollment service.
func NewMockCAService(orgName string, cryptoProvider core.CryptoSuite, config core.Config) (idapi.Client, error) {
	mes := MockCAService{}
	return &mes, nil
}

// CAName return the name of a CA associated with this identity manager
func (mes *MockCAService) CAName() string {
	return ""
}

// Enroll enrolls a user with a Fabric network
func (mes *MockCAService) Enroll(enrollmentID string, enrollmentSecret string) error {
	return errors.New("not implemented")
}

// Reenroll re-enrolls a user
func (mes *MockCAService) Reenroll(enrollmentID string) error {
	return errors.New("not implemented")
}

// Register registers a user with a Fabric network
func (mes *MockCAService) Register(request *idapi.RegistrationRequest) (string, error) {
	return "", errors.New("not implemented")
}

// Revoke revokes a user
func (mes *MockCAService) Revoke(request *idapi.RevocationRequest) (*idapi.RevocationResponse, error) {
	return nil, errors.New("not implemented")
}
