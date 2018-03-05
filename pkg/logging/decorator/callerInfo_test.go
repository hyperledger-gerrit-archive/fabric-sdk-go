/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package decorator

import (
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/logging/testutils"
)

func TestCallerInfoSetting(t *testing.T) {

	sampleCallerInfoSetting := CallerInfo{}
	samppleModuleName := "sample-module-name"

	sampleCallerInfoSetting.ShowCallerInfo(samppleModuleName, DEBUG)
	testutils.VerifyTrue(t, sampleCallerInfoSetting.IsCallerInfoEnabled(samppleModuleName, DEBUG), "Callerinfo supposed to be enabled for this level")

	sampleCallerInfoSetting.HideCallerInfo(samppleModuleName, DEBUG)
	testutils.VerifyFalse(t, sampleCallerInfoSetting.IsCallerInfoEnabled(samppleModuleName, DEBUG), "Callerinfo supposed to be disabled for this level")

	//Reset existing caller info setting
	sampleCallerInfoSetting.showcaller = nil

	//By default caller info should be disabled if not set
	testutils.VerifyTrue(t, sampleCallerInfoSetting.IsCallerInfoEnabled(samppleModuleName, DEBUG), "Callerinfo supposed to be enabled for this level")
	testutils.VerifyTrue(t, sampleCallerInfoSetting.IsCallerInfoEnabled(samppleModuleName, INFO), "Callerinfo supposed to be disabled for this level")
	testutils.VerifyTrue(t, sampleCallerInfoSetting.IsCallerInfoEnabled(samppleModuleName, WARNING), "Callerinfo supposed to be disabled for this level")
	testutils.VerifyTrue(t, sampleCallerInfoSetting.IsCallerInfoEnabled(samppleModuleName, ERROR), "Callerinfo supposed to be disabled for this level")
	testutils.VerifyTrue(t, sampleCallerInfoSetting.IsCallerInfoEnabled(samppleModuleName, CRITICAL), "Callerinfo supposed to be disabled for this level")

	//By default caller info should be disabled if module name not found
	samppleInvalidModuleName := "sample-module-name-doesnt-exists"
	testutils.VerifyTrue(t, sampleCallerInfoSetting.IsCallerInfoEnabled(samppleInvalidModuleName, INFO), "Callerinfo supposed to be disabled for this level")
	testutils.VerifyTrue(t, sampleCallerInfoSetting.IsCallerInfoEnabled(samppleInvalidModuleName, WARNING), "Callerinfo supposed to be disabled for this level")
	testutils.VerifyTrue(t, sampleCallerInfoSetting.IsCallerInfoEnabled(samppleInvalidModuleName, ERROR), "Callerinfo supposed to be disabled for this level")
	testutils.VerifyTrue(t, sampleCallerInfoSetting.IsCallerInfoEnabled(samppleInvalidModuleName, CRITICAL), "Callerinfo supposed to be disabled for this level")
	testutils.VerifyTrue(t, sampleCallerInfoSetting.IsCallerInfoEnabled(samppleInvalidModuleName, DEBUG), "Callerinfo supposed to be disabled for this level")
}
