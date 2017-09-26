/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
package logging

import (
	"bytes"
	"fmt"
	"regexp"
	"testing"
)

const (
	basicLevelOutputExpectedRegex = "\\[%s\\] .* UTC - logging.* -> %s brown fox jumps over the lazy dog"
	printLevelOutputExpectedRegex = "\\[%s\\] .* brown fox jumps over the lazy dog"
	moduleName                    = "module-xyz"
)

type fn func(...interface{})
type fnf func(string, ...interface{})

func TestLevelledLogging(t *testing.T) {

	logger := NewLogger(moduleName)

	//change the output to buffer
	var buf bytes.Buffer
	logger.logger.SetOutput(&buf)

	//No level set for this module so log level should be info
	verifyTrue(t, INFO == GetLevel(moduleName), " default log level is INFO")

	//Test logger.print outputs
	verifyBasicLogging(t, -1, logger.Print, nil, &buf)
	verifyBasicLogging(t, -1, logger.Println, nil, &buf)
	verifyBasicLogging(t, -1, nil, logger.Printf, &buf)

	//Test logger.info outputs
	verifyBasicLogging(t, INFO, logger.Info, nil, &buf)
	verifyBasicLogging(t, INFO, logger.Infoln, nil, &buf)
	verifyBasicLogging(t, INFO, nil, logger.Infof, &buf)

	//Test logger.warn outputs
	verifyBasicLogging(t, WARNING, logger.Warn, nil, &buf)
	verifyBasicLogging(t, WARNING, logger.Warnln, nil, &buf)
	verifyBasicLogging(t, WARNING, nil, logger.Warnf, &buf)

	//Test logger.error outputs
	verifyBasicLogging(t, ERROR, logger.Error, nil, &buf)
	verifyBasicLogging(t, ERROR, logger.Errorln, nil, &buf)
	verifyBasicLogging(t, ERROR, nil, logger.Errorf, &buf)

	/*
		SINCE DEBUG LOG IS NOT YET ENABLED, LOG OUTPUT SHOULD BE EMPTY
	*/
	//Test logger.debug outputs when DEBUG level is not enabled
	logger.Debug("brown fox jumps over the lazy dog")
	logger.Debugln("brown fox jumps over the lazy dog")
	logger.Debugf("brown %s jumps over the lazy %s", "fox", "dog")

	verifyEmpty(t, buf.String(), "debug log isn't supposed to show up for info level")

	//Now change the log level to DEBUG
	SetLevel(DEBUG, moduleName)

	//Test logger.debug outputs
	verifyBasicLogging(t, DEBUG, logger.Debug, nil, &buf)
	verifyBasicLogging(t, DEBUG, logger.Debugln, nil, &buf)
	verifyBasicLogging(t, DEBUG, nil, logger.Debugf, &buf)
}

func TestPanic(t *testing.T) {

	logger := NewLogger(moduleName)

	verifyCriticalLoggings(t, CRITICAL, logger.Panic, nil, logger)
	verifyCriticalLoggings(t, CRITICAL, logger.Panicln, nil, logger)
	verifyCriticalLoggings(t, CRITICAL, nil, logger.Panicf, logger)

}

//verifyCriticalLoggings utility func which does job calling and verifying CRITICAL log level functions - PANIC
func verifyCriticalLoggings(t *testing.T, level Level, loggerFunc fn, loggerFuncf fnf, logger *Logger) {

	//change the output to buffer
	var buf bytes.Buffer
	logger.logger.SetOutput(&buf)

	//Handling panic as well as checking log output
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("%v was supposed to panic", loggerFunc)
		}
		regex := fmt.Sprintf(basicLevelOutputExpectedRegex, moduleName, levelNames[level])
		match, err := regexp.MatchString(regex, buf.String())
		verifyEmpty(t, err, "error while matching regex with logoutput wasnt expected")
		verifyTrue(t, match, "CRITICAL logger isn't producing output as expected, \n logoutput:%s\n regex: %s", buf.String(), regex)

	}()

	//Call logger func
	if loggerFunc != nil {
		loggerFunc("brown fox jumps over the lazy dog")
	} else if loggerFuncf != nil {
		loggerFuncf("brown %s jumps over the lazy %s", "fox", "dog")
	}
}

//verifyBasicLogging utility func which does job calling and verifying basic log level functions - DEBUG, INFO, ERROR, WARNING
func verifyBasicLogging(t *testing.T, level Level, loggerFunc fn, loggerFuncf fnf, buf *bytes.Buffer) {

	//Call logger func
	if loggerFunc != nil {
		loggerFunc("brown fox jumps over the lazy dog")
	} else if loggerFuncf != nil {
		loggerFuncf("brown %s jumps over the lazy %s", "fox", "dog")
	}

	//check output
	regex := ""
	levelName := "print"

	if level > 0 {
		levelName = levelNames[level]
		regex = fmt.Sprintf(basicLevelOutputExpectedRegex, moduleName, levelName)
	} else {
		regex = fmt.Sprintf(printLevelOutputExpectedRegex, moduleName)
	}

	match, err := regexp.MatchString(regex, buf.String())

	verifyEmpty(t, err, "error while matching regex with logoutput wasnt expected")
	verifyTrue(t, match, "%s logger isn't producing output as expected, \n logoutput:%s\n regex: %s", levelName, buf.String(), regex)

	//Reset output buffer, for next use
	buf.Reset()
}

func verifyTrue(t *testing.T, input bool, msgAndArgs ...interface{}) {
	if !input {
		failTest(t, msgAndArgs)
	}
}

func verifyEmpty(t *testing.T, in interface{}, msgAndArgs ...interface{}) {
	if in == nil {
		return
	} else if in == "" {
		return
	}
	failTest(t, msgAndArgs...)
}

func failTest(t *testing.T, msgAndArgs ...interface{}) {
	if len(msgAndArgs) == 1 {
		t.Fatal(msgAndArgs[0])
	}
	if len(msgAndArgs) > 1 {
		t.Fatalf(msgAndArgs[0].(string), msgAndArgs[1:]...)
	}
}
