/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
/*
Notice: This file has been modified for Hyperledger Fabric SDK Go usage.
Please review third_party pinning scripts and patches for more details.
*/

package persistence

// PackageID encapsulates chaincode ID
type PackageID string

// String returns a string version of the package ID
func (p PackageID) String() string {
	return string(p)
}
