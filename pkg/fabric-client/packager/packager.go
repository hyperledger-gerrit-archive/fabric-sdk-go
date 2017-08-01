/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package packager

import (
	"fmt"

	pb "github.com/hyperledger/fabric/protos/peer"
)

// PackageCC ...
/**
 * Utility function to package a chaincode. The contents will be returned as a byte array.
 *
 * @param {string} chaincodePath required - String of the path to location of
 *                the source code of the chaincode
 * @param {string} chaincodeType optional - String of the type of chaincode
 *                 ['golang', 'car', 'java'] (default 'golang')
 * @returns {[]byte} byte array
 */
func PackageCC(chaincodePath string, chaincodeType pb.ChaincodeSpec_Type) ([]byte, error) {
	logger.Debugf("packager: chaincodePath: %s, chaincodeType: %s", chaincodePath, chaincodeType)
	if chaincodePath == "" {
		return nil, fmt.Errorf("Missing 'chaincodePath' parameter")
	}

	logger.Debugf("packager: type %s ", chaincodeType)
	switch chaincodeType {
	case pb.ChaincodeSpec_GOLANG:
		return PackageGoLangCC(chaincodePath)
	case pb.ChaincodeSpec_BINARY:
		return PackageBinaryCC(chaincodePath)
	}
	return nil, fmt.Errorf("Undefined 'chaincodeType' value")
}
