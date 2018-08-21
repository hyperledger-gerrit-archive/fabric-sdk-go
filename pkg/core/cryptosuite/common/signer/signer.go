package signer

import (
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric-ca/util"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/pkg/errors"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	mspImpl "github.com/hyperledger/fabric-sdk-go/pkg/msp"
)

// New creates a new user
func New(userData *msp.UserData, key []byte, cryptoSuite core.CryptoSuite, tmp bool) (*mspImpl.User, error) {
	privateKey, err := util.ImportBCCSPKeyFromPEMBytes(key, cryptoSuite, tmp)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to import key")
	}
	return mspImpl.NewUser(userData, privateKey), nil
}
