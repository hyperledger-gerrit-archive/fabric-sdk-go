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

// MockEnrollmentService is a mock EnrollmentService
type MockEnrollmentService struct {
}

// NewMockEnrollmentService Constructor for a enrollment service.
func NewMockEnrollmentService(orgName string, cryptoProvider core.CryptoSuite, config core.Config) (idapi.EnrollmentService, error) {
	mes := MockEnrollmentService{}
	return &mes, nil
}

// Enroll enrolls a user with a Fabric network
func (mes *MockEnrollmentService) Enroll(enrollmentID string, enrollmentSecret string) error {
	return errors.New("not implemented")
}

// Reenroll re-enrolls a user
func (mes *MockEnrollmentService) Reenroll(enrollmentID string) error {
	return errors.New("not implemented")
}

// CAName return the name of a CA associated with this identity manager
func (mes *MockEnrollmentService) CAName() string {
	return ""
}
