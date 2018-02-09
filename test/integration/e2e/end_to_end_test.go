/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package e2e

import (
	"testing"
)

func TestE2E(t *testing.T) {
	runWithConfigFixture(t)

	//Using setup set by previous test run, run below test with new config
	runWithNoOrdererConfigFixture(t)
}
