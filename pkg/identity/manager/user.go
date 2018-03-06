/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package manager

import (
	"github.com/golang/protobuf/proto"

	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
	pb_msp "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/msp"
	"github.com/pkg/errors"
)

// User is a representation of a Fabric user
type User struct {
	MspID_                 string
	Name_                  string
	EnrollmentCertificate_ []byte
	PrivateKey_            core.Key
}

func userIdentifier(userData UserData) UserIdentifier {
	return UserIdentifier{MspID: userData.MspID, Name: userData.Name}
}

// Name Get the user name.
// @returns {string} The user name.
func (u *User) Name() string {
	return u.Name_
}

// EnrollmentCertificate Returns the underlying ECert representing this user’s identity.
func (u *User) EnrollmentCertificate() []byte {
	return u.EnrollmentCertificate_
}

// PrivateKey returns the crypto suite representation of the private key
func (u *User) PrivateKey() core.Key {
	return u.PrivateKey_
}

// MspID returns the MSP for this user
func (u *User) MspID() string {
	return u.MspID_
}

// SerializedIdentity returns client's serialized identity
func (u *User) SerializedIdentity() ([]byte, error) {
	serializedIdentity := &pb_msp.SerializedIdentity{Mspid: u.MspID_,
		IdBytes: u.EnrollmentCertificate_}
	identity, err := proto.Marshal(serializedIdentity)
	if err != nil {
		return nil, errors.Wrap(err, "marshal serializedIdentity failed")
	}
	return identity, nil
}

// UserStore is responsible for UserData persistence
type UserStore interface {
	Store(UserData) error
	Load(UserIdentifier) (UserData, error)
}
