/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package identitymgr

import (
	"math/rand"
	"strconv"

	"github.com/hyperledger/fabric-sdk-go/pkg/context/api"
	"github.com/pkg/errors"
)

func checkUser(mgr api.IdentityManager, user string) error {
	u, err := mgr.GetUser(user)
	if err == api.ErrUserNotFound {
		return err
	}
	if err != nil {
		return errors.Wrapf(err, "Failed to retrieve signing identity: %s", err)
	}

	if u == nil {
		return errors.New("SigningIdentity is nil")
	}
	if u.EnrollmentCertificate() == nil {
		return errors.New("Enrollment cert is missing")
	}
	if u.MspID() == "" {
		return errors.New("MspID is missing")
	}
	if u.PrivateKey() == nil {
		return errors.New("private key is missing")
	}
	return nil
}

func createRandomName() string {
	return "user" + strconv.Itoa(rand.Intn(500000))
}
