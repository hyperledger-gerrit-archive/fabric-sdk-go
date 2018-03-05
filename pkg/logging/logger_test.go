/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
package logging

import (
	"bytes"
	"sync"
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/testutils"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging/decorator"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging/modlog"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging/testdata"
)

var moduleName = "module-xyz"
var moduleName2 = "module-xyz-deftest"
var logPrefixFormatter = " [%s] "
var buf bytes.Buffer

func TestLoggingForCustomLogger(t *testing.T) {

	//Now add sample logger
	resetLoggerInstance()
	InitLogger(testdata.GetSampleLoggingProvider(&buf))
	//Create new logger
	logger := NewLogger(moduleName)

	//Test logger.print outputs
	modlog.VerifyBasicLogging(t, decorator.INFO, logger.Print, nil, &buf, true, moduleName)
	modlog.VerifyBasicLogging(t, decorator.INFO, logger.Println, nil, &buf, true, moduleName)
	modlog.VerifyBasicLogging(t, decorator.INFO, nil, logger.Printf, &buf, true, moduleName)

	//Test logger.info outputs
	modlog.VerifyBasicLogging(t, decorator.INFO, logger.Info, nil, &buf, true, moduleName)
	modlog.VerifyBasicLogging(t, decorator.INFO, logger.Infoln, nil, &buf, true, moduleName)
	modlog.VerifyBasicLogging(t, decorator.INFO, nil, logger.Infof, &buf, true, moduleName)

	//Test logger.warn outputs
	modlog.VerifyBasicLogging(t, decorator.WARNING, logger.Warn, nil, &buf, true, moduleName)
	modlog.VerifyBasicLogging(t, decorator.WARNING, logger.Warnln, nil, &buf, true, moduleName)
	modlog.VerifyBasicLogging(t, decorator.WARNING, nil, logger.Warnf, &buf, true, moduleName)

	//In middle of test, get new logger, it should still stick to custom logger
	logger = NewLogger(moduleName)

	//Test logger.error outputs
	modlog.VerifyBasicLogging(t, decorator.ERROR, logger.Error, nil, &buf, true, moduleName)
	modlog.VerifyBasicLogging(t, decorator.ERROR, logger.Errorln, nil, &buf, true, moduleName)
	modlog.VerifyBasicLogging(t, decorator.ERROR, nil, logger.Errorf, &buf, true, moduleName)

	//Test logger.debug outputs
	modlog.VerifyBasicLogging(t, decorator.DEBUG, logger.Debug, nil, &buf, true, moduleName)
	modlog.VerifyBasicLogging(t, decorator.DEBUG, logger.Debugln, nil, &buf, true, moduleName)
	modlog.VerifyBasicLogging(t, decorator.DEBUG, nil, logger.Debugf, &buf, true, moduleName)

	////Test logger.fatal outputs - this custom logger doesn't cause os exit code 1
	modlog.VerifyBasicLogging(t, decorator.CRITICAL, logger.Fatal, nil, &buf, true, moduleName)
	modlog.VerifyBasicLogging(t, decorator.CRITICAL, logger.Fatalln, nil, &buf, true, moduleName)
	modlog.VerifyBasicLogging(t, decorator.CRITICAL, nil, logger.Fatalf, &buf, true, moduleName)

	//Test logger.panic outputs - this custom logger doesn't cause panic
	modlog.VerifyBasicLogging(t, decorator.CRITICAL, logger.Panic, nil, &buf, true, moduleName)
	modlog.VerifyBasicLogging(t, decorator.CRITICAL, logger.Panicln, nil, &buf, true, moduleName)
	modlog.VerifyBasicLogging(t, decorator.CRITICAL, nil, logger.Panicf, &buf, true, moduleName)

}

func TestDefaultModulledLoggingBehavior(t *testing.T) {

	//Init logger with default logger
	resetLoggerInstance()
	InitLogger(modlog.LoggerProvider())
	//Get new logger
	dlogger := NewLogger(moduleName)
	// force initialization
	dlogger.logger()
	//Change output
	dlogger.instance.(*modlog.Log).ChangeOutput(&buf)

	//No level set for this module so log level should be info
	testutils.VerifyTrue(t, decorator.INFO == modlog.GetLevel(moduleName), " default log level is INFO")

	//Test logger.print outputs
	modlog.VerifyBasicLogging(t, -1, dlogger.Print, nil, &buf, false, moduleName)
	modlog.VerifyBasicLogging(t, -1, dlogger.Println, nil, &buf, false, moduleName)
	modlog.VerifyBasicLogging(t, -1, nil, dlogger.Printf, &buf, false, moduleName)

	//Test logger.info outputs
	modlog.VerifyBasicLogging(t, decorator.INFO, dlogger.Info, nil, &buf, false, moduleName)
	modlog.VerifyBasicLogging(t, decorator.INFO, dlogger.Infoln, nil, &buf, false, moduleName)
	modlog.VerifyBasicLogging(t, decorator.INFO, nil, dlogger.Infof, &buf, false, moduleName)

	//Test logger.warn outputs
	modlog.VerifyBasicLogging(t, decorator.WARNING, dlogger.Warn, nil, &buf, false, moduleName)
	modlog.VerifyBasicLogging(t, decorator.WARNING, dlogger.Warnln, nil, &buf, false, moduleName)
	modlog.VerifyBasicLogging(t, decorator.WARNING, nil, dlogger.Warnf, &buf, false, moduleName)

	//Test logger.error outputs
	modlog.VerifyBasicLogging(t, decorator.ERROR, dlogger.Error, nil, &buf, false, moduleName)
	modlog.VerifyBasicLogging(t, decorator.ERROR, dlogger.Errorln, nil, &buf, false, moduleName)
	modlog.VerifyBasicLogging(t, decorator.ERROR, nil, dlogger.Errorf, &buf, false, moduleName)

	/*
		SINCE DEBUG LOG IS NOT YET ENABLED, LOG OUTPUT SHOULD BE EMPTY
	*/
	//Test logger.debug outputs when DEBUG level is not enabled
	dlogger.Debug("brown fox jumps over the lazy dog")
	dlogger.Debugln("brown fox jumps over the lazy dog")
	dlogger.Debugf("brown %s jumps over the lazy %s", "fox", "dog")

	testutils.VerifyEmpty(t, buf.String(), "debug log isn't supposed to show up for info level")

	//Should be false
	testutils.VerifyFalse(t, modlog.IsEnabledFor(moduleName, decorator.DEBUG), "logging.IsEnabled for is not working as expected, expected false but got true")

	//Now change the log level to DEBUG
	modlog.SetLevel(moduleName, decorator.DEBUG)

	//Should be false
	testutils.VerifyTrue(t, modlog.IsEnabledFor(moduleName, decorator.DEBUG), "logging.IsEnabled for is not working as expected, expected true but got false")

	//Test logger.debug outputs
	modlog.VerifyBasicLogging(t, decorator.DEBUG, dlogger.Debug, nil, &buf, false, moduleName)
	modlog.VerifyBasicLogging(t, decorator.DEBUG, dlogger.Debugln, nil, &buf, false, moduleName)
	modlog.VerifyBasicLogging(t, decorator.DEBUG, nil, dlogger.Debugf, &buf, false, moduleName)

}

func TestLoggerSetting(t *testing.T) {
	resetLoggerInstance()
	logger := NewLogger(moduleName)
	testutils.VerifyTrue(t, loggerProviderInstance == nil, "Logger is not supposed to be initialized now")
	logger.Info("brown fox jumps over the lazy dog")
	testutils.VerifyTrue(t, loggerProviderInstance != nil, "Logger is supposed to be initialized now")
	resetLoggerInstance()
	InitLogger(modlog.LoggerProvider())
	testutils.VerifyTrue(t, loggerProviderInstance != nil, "Logger is supposed to be initialized now")
}

func resetLoggerInstance() {
	loggerProviderInstance = nil
	loggerProviderOnce = sync.Once{}
}

func TestDefaultCustomModuledLoggingBehavior(t *testing.T) {

	//Init logger with default logger
	resetLoggerInstance()
	//Set custom logger in place of default logger
	modlog.InitLogger(testdata.GetSampleLoggingProvider(&buf))
	//Get new logger
	dlogger := NewLogger(moduleName2)
	// force initialization
	dlogger.logger()

	//No level set for this module so log level should be info
	testutils.VerifyTrue(t, decorator.INFO == modlog.GetLevel(moduleName2), " default log level is INFO")

	//Test logger.print outputs
	modlog.VerifyBasicLogging(t, decorator.INFO, dlogger.Print, nil, &buf, true, moduleName2)
	modlog.VerifyBasicLogging(t, decorator.INFO, dlogger.Println, nil, &buf, true, moduleName2)
	modlog.VerifyBasicLogging(t, decorator.INFO, nil, dlogger.Printf, &buf, true, moduleName2)

	//Test logger.info outputs
	modlog.VerifyBasicLogging(t, decorator.INFO, dlogger.Info, nil, &buf, true, moduleName2)
	modlog.VerifyBasicLogging(t, decorator.INFO, dlogger.Infoln, nil, &buf, true, moduleName2)
	modlog.VerifyBasicLogging(t, decorator.INFO, nil, dlogger.Infof, &buf, true, moduleName2)

	//Test logger.warn outputs
	modlog.VerifyBasicLogging(t, decorator.WARNING, dlogger.Warn, nil, &buf, true, moduleName2)
	modlog.VerifyBasicLogging(t, decorator.WARNING, dlogger.Warnln, nil, &buf, true, moduleName2)
	modlog.VerifyBasicLogging(t, decorator.WARNING, nil, dlogger.Warnf, &buf, true, moduleName2)

	//Test logger.error outputs
	modlog.VerifyBasicLogging(t, decorator.ERROR, dlogger.Error, nil, &buf, true, moduleName2)
	modlog.VerifyBasicLogging(t, decorator.ERROR, dlogger.Errorln, nil, &buf, true, moduleName2)
	modlog.VerifyBasicLogging(t, decorator.ERROR, nil, dlogger.Errorf, &buf, true, moduleName2)

	//Should be false
	testutils.VerifyFalse(t, modlog.IsEnabledFor(moduleName2, decorator.DEBUG), "logging.IsEnabled for is not working as expected, expected false but got true")

	//Now change the log level to DEBUG
	modlog.SetLevel(moduleName2, decorator.DEBUG)

	//Should be false
	testutils.VerifyTrue(t, modlog.IsEnabledFor(moduleName2, decorator.DEBUG), "logging.IsEnabled for is not working as expected, expected true but got false")

	//Test logger.debug outputs
	modlog.VerifyBasicLogging(t, decorator.DEBUG, dlogger.Debug, nil, &buf, true, moduleName2)
	modlog.VerifyBasicLogging(t, decorator.DEBUG, dlogger.Debugln, nil, &buf, true, moduleName2)
	modlog.VerifyBasicLogging(t, decorator.DEBUG, nil, dlogger.Debugf, &buf, true, moduleName2)

}
