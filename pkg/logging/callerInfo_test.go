/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package logging

import (
	"testing"
)

func TestCallerInfoSetting(t *testing.T) {

	sampleCallerInfoSetting := callerInfo{}
	samppleModuleName := "sample-module-name"

	sampleCallerInfoSetting.HideCallerInfo(samppleModuleName, DEBUG)
	verifyFalse(t, sampleCallerInfoSetting.IsCallerInfoEnabled(samppleModuleName, DEBUG), "Callerinfo supposed to be disabled for this level")

	sampleCallerInfoSetting.ShowCallerInfo(samppleModuleName, DEBUG)
	verifyTrue(t, sampleCallerInfoSetting.IsCallerInfoEnabled(samppleModuleName, DEBUG), "Callerinfo supposed to be enabled for this level")

	//Reset existing caller info setting
	sampleCallerInfoSetting.showcaller = nil

	sampleCallerInfoSetting.ShowCallerInfo(samppleModuleName, DEBUG)
	verifyTrue(t, sampleCallerInfoSetting.IsCallerInfoEnabled(samppleModuleName, DEBUG), "Callerinfo supposed to be enabled for this level")

	//By default caller info should be disabled if not set
	verifyFalse(t, sampleCallerInfoSetting.IsCallerInfoEnabled(samppleModuleName, INFO), "Callerinfo supposed to be disabled for this level")
	verifyFalse(t, sampleCallerInfoSetting.IsCallerInfoEnabled(samppleModuleName, WARNING), "Callerinfo supposed to be disabled for this level")
	verifyFalse(t, sampleCallerInfoSetting.IsCallerInfoEnabled(samppleModuleName, ERROR), "Callerinfo supposed to be disabled for this level")
	verifyFalse(t, sampleCallerInfoSetting.IsCallerInfoEnabled(samppleModuleName, CRITICAL), "Callerinfo supposed to be disabled for this level")

	//By default caller info should be disabled if module name not found
	samppleInvalidModuleName := "sample-module-name-doesnt-exists"
	verifyFalse(t, sampleCallerInfoSetting.IsCallerInfoEnabled(samppleInvalidModuleName, INFO), "Callerinfo supposed to be disabled for this level")
	verifyFalse(t, sampleCallerInfoSetting.IsCallerInfoEnabled(samppleInvalidModuleName, WARNING), "Callerinfo supposed to be disabled for this level")
	verifyFalse(t, sampleCallerInfoSetting.IsCallerInfoEnabled(samppleInvalidModuleName, ERROR), "Callerinfo supposed to be disabled for this level")
	verifyFalse(t, sampleCallerInfoSetting.IsCallerInfoEnabled(samppleInvalidModuleName, CRITICAL), "Callerinfo supposed to be disabled for this level")
	verifyFalse(t, sampleCallerInfoSetting.IsCallerInfoEnabled(samppleInvalidModuleName, DEBUG), "Callerinfo supposed to be disabled for this level")
}
