/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
package logging

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogLevels(t *testing.T) {

	mlevel := moduleLeveled{}

	mlevel.SetLevel(INFO, "module-xyz-info")
	mlevel.SetLevel(DEBUG, "module-xyz-debug")
	mlevel.SetLevel(ERROR, "module-xyz-error")
	mlevel.SetLevel(WARNING, "module-xyz-warning")

	//Run info level checks
	assert.True(t, mlevel.IsEnabledFor(INFO, "module-xyz-info"))
	assert.False(t, mlevel.IsEnabledFor(DEBUG, "module-xyz-info"))
	assert.True(t, mlevel.IsEnabledFor(ERROR, "module-xyz-info"))
	assert.True(t, mlevel.IsEnabledFor(WARNING, "module-xyz-info"))

	//Run debug level checks
	assert.True(t, mlevel.IsEnabledFor(INFO, "module-xyz-debug"))
	assert.True(t, mlevel.IsEnabledFor(DEBUG, "module-xyz-debug"))
	assert.True(t, mlevel.IsEnabledFor(ERROR, "module-xyz-debug"))
	assert.True(t, mlevel.IsEnabledFor(WARNING, "module-xyz-debug"))

	//Run info level checks
	assert.False(t, mlevel.IsEnabledFor(INFO, "module-xyz-error"))
	assert.False(t, mlevel.IsEnabledFor(DEBUG, "module-xyz-error"))
	assert.True(t, mlevel.IsEnabledFor(ERROR, "module-xyz-error"))
	assert.False(t, mlevel.IsEnabledFor(WARNING, "module-xyz-error"))

	//Run info level checks
	assert.False(t, mlevel.IsEnabledFor(INFO, "module-xyz-warning"))
	assert.False(t, mlevel.IsEnabledFor(DEBUG, "module-xyz-warning"))
	assert.True(t, mlevel.IsEnabledFor(ERROR, "module-xyz-warning"))
	assert.True(t, mlevel.IsEnabledFor(WARNING, "module-xyz-warning"))

	//Run default log level check --> which is info currently
	assert.True(t, mlevel.IsEnabledFor(INFO, "module-xyz-random-module"))
	assert.False(t, mlevel.IsEnabledFor(DEBUG, "module-xyz-random-module"))
	assert.True(t, mlevel.IsEnabledFor(ERROR, "module-xyz-random-module"))
	assert.True(t, mlevel.IsEnabledFor(WARNING, "module-xyz-random-module"))

}

func TestGetLogLevels(t *testing.T) {

	level, err := LogLevel("info")
	verifyLogLevel(t, INFO, level, err, true)

	level, err = LogLevel("iNfO")
	verifyLogLevel(t, INFO, level, err, true)

	level, err = LogLevel("debug")
	verifyLogLevel(t, DEBUG, level, err, true)

	level, err = LogLevel("DeBuG")
	verifyLogLevel(t, DEBUG, level, err, true)

	level, err = LogLevel("warning")
	verifyLogLevel(t, WARNING, level, err, true)

	level, err = LogLevel("WarNIng")
	verifyLogLevel(t, WARNING, level, err, true)

	level, err = LogLevel("error")
	verifyLogLevel(t, ERROR, level, err, true)

	level, err = LogLevel("eRRoR")
	verifyLogLevel(t, ERROR, level, err, true)

	level, err = LogLevel("outofthebox")
	verifyLogLevel(t, -1, level, err, false)

	level, err = LogLevel("")
	verifyLogLevel(t, -1, level, err, false)
}

func verifyLogLevel(t *testing.T, expectedLevel Level, currentlevel Level, err error, success bool) {
	if success {
		assert.Nil(t, err, "not supposed to get error for this scenario")
	} else {
		assert.NotNil(t, err, "supposed to get error for this scenario, but got error : %v", err)
		return
	}

	assert.True(t, currentlevel == expectedLevel, "unexpected log level : expected '%s', but got '%s'", expectedLevel, currentlevel)
}
