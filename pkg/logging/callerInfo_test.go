/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package logging

import "testing"

func TestCallerInfoSetting(t *testing.T) {

	sampleCallerInfoSetting := callerInfo{}

	sampleCallerInfoSetting.HideCallerInfo(DEBUG)
	verifyFalse(t, sampleCallerInfoSetting.IsCallerInfoEnabled(DEBUG), "Callerinfo supposed to be disabled for this level")

	sampleCallerInfoSetting.ShowCallerInfo(DEBUG)
	verifyTrue(t, sampleCallerInfoSetting.IsCallerInfoEnabled(DEBUG), "Callerinfo supposed to be enabled for this module")

	//Reset existing caller info setting
	sampleCallerInfoSetting.showcaller = nil

	sampleCallerInfoSetting.ShowCallerInfo(DEBUG)
	verifyTrue(t, sampleCallerInfoSetting.IsCallerInfoEnabled(DEBUG), "Callerinfo supposed to be enabled for this module")

	//By default caller info should be disabled
	verifyFalse(t, sampleCallerInfoSetting.IsCallerInfoEnabled(INFO), "Callerinfo supposed to be disabled for this level")
	verifyFalse(t, sampleCallerInfoSetting.IsCallerInfoEnabled(WARNING), "Callerinfo supposed to be disabled for this level")
	verifyFalse(t, sampleCallerInfoSetting.IsCallerInfoEnabled(ERROR), "Callerinfo supposed to be disabled for this level")
	verifyFalse(t, sampleCallerInfoSetting.IsCallerInfoEnabled(CRITICAL), "Callerinfo supposed to be disabled for this level")

}
