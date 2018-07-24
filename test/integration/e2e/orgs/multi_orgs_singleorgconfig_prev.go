// +build prev

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package orgs

import (
	"testing"
)

//runMultiOrgTestWithSingleOrgConfig cannot be run in prev tests since fabric version greater than v1.1
// supports dynamic discovery
func runMultiOrgTestWithSingleOrgConfig(t *testing.T, examplecc string) { //nolint
	//test nothing
}
