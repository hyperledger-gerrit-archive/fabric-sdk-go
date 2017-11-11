/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
package logging

import (
	"bytes"
	"fmt"
	"log"
	"testing"

	"github.com/hyperledger/fabric-sdk-go/api/apilogging"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging/deflogger"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging/utils"
)

var moduleName = "module-xyz"
var logPrefixFormatter = " [%s] "

type fn func(...interface{})
type fnf func(string, ...interface{})

func TestLevelledLoggingForCustomLogger(t *testing.T) {

	//prepare custom logger for which output is bytes buffer
	var buf bytes.Buffer
	customLogger := log.New(&buf, fmt.Sprintf(logPrefixFormatter, moduleName), log.Ldate|log.Ltime|log.LUTC)

	//Create new logger
	logger := NewLogger(moduleName)
	//Now add custom logger
	SetCustomLogger(&SampleCustomLogger{customLogger: customLogger, module: moduleName})

	//Test logger.print outputs
	deflogger.VerifyBasicLogging(t, apilogging.INFO, logger.Print, nil, &buf, true)
	deflogger.VerifyBasicLogging(t, apilogging.INFO, logger.Println, nil, &buf, true)
	deflogger.VerifyBasicLogging(t, apilogging.INFO, nil, logger.Printf, &buf, true)

	//Test logger.info outputs
	deflogger.VerifyBasicLogging(t, apilogging.INFO, logger.Info, nil, &buf, true)
	deflogger.VerifyBasicLogging(t, apilogging.INFO, logger.Infoln, nil, &buf, true)
	deflogger.VerifyBasicLogging(t, apilogging.INFO, nil, logger.Infof, &buf, true)

	//Test logger.warn outputs
	deflogger.VerifyBasicLogging(t, apilogging.WARNING, logger.Warn, nil, &buf, true)
	deflogger.VerifyBasicLogging(t, apilogging.WARNING, logger.Warnln, nil, &buf, true)
	deflogger.VerifyBasicLogging(t, apilogging.WARNING, nil, logger.Warnf, &buf, true)

	//In middle of test, get new logger, it should still stick to custom logger
	logger = NewLogger(moduleName)

	//Test logger.error outputs
	deflogger.VerifyBasicLogging(t, apilogging.ERROR, logger.Error, nil, &buf, true)
	deflogger.VerifyBasicLogging(t, apilogging.ERROR, logger.Errorln, nil, &buf, true)
	deflogger.VerifyBasicLogging(t, apilogging.ERROR, nil, logger.Errorf, &buf, true)

	//Test logger.debug outputs
	deflogger.VerifyBasicLogging(t, apilogging.DEBUG, logger.Debug, nil, &buf, true)
	deflogger.VerifyBasicLogging(t, apilogging.DEBUG, logger.Debugln, nil, &buf, true)
	deflogger.VerifyBasicLogging(t, apilogging.DEBUG, nil, logger.Debugf, &buf, true)

	////Test logger.fatal outputs - this custom logger doesn't cause os exit code 1
	deflogger.VerifyBasicLogging(t, apilogging.CRITICAL, logger.Fatal, nil, &buf, true)
	deflogger.VerifyBasicLogging(t, apilogging.CRITICAL, logger.Fatalln, nil, &buf, true)
	deflogger.VerifyBasicLogging(t, apilogging.CRITICAL, nil, logger.Fatalf, &buf, true)

	//Test logger.panic outputs - this custom logger doesn't cause panic
	deflogger.VerifyBasicLogging(t, apilogging.CRITICAL, logger.Panic, nil, &buf, true)
	deflogger.VerifyBasicLogging(t, apilogging.CRITICAL, logger.Panicln, nil, &buf, true)
	deflogger.VerifyBasicLogging(t, apilogging.CRITICAL, nil, logger.Panicf, &buf, true)

}

/*
	Test custom logger
*/

type SampleCustomLogger struct {
	customLogger *log.Logger
	module       string
}

func (l *SampleCustomLogger) Fatal(v ...interface{}) { l.customLogger.Print("CUSTOM LOG OUTPUT") }
func (l *SampleCustomLogger) Fatalf(format string, v ...interface{}) {
	l.customLogger.Print("CUSTOM LOG OUTPUT")
}
func (l *SampleCustomLogger) Fatalln(v ...interface{}) { l.customLogger.Print("CUSTOM LOG OUTPUT") }
func (l *SampleCustomLogger) Panic(v ...interface{})   { l.customLogger.Print("CUSTOM LOG OUTPUT") }
func (l *SampleCustomLogger) Panicf(format string, v ...interface{}) {
	l.customLogger.Print("CUSTOM LOG OUTPUT")
}
func (l *SampleCustomLogger) Panicln(v ...interface{}) { l.customLogger.Print("CUSTOM LOG OUTPUT") }
func (l *SampleCustomLogger) Print(v ...interface{})   { l.customLogger.Print("CUSTOM LOG OUTPUT") }
func (l *SampleCustomLogger) Printf(format string, v ...interface{}) {
	l.customLogger.Print("CUSTOM LOG OUTPUT")
}
func (l *SampleCustomLogger) Println(v ...interface{}) { l.customLogger.Print("CUSTOM LOG OUTPUT") }
func (l *SampleCustomLogger) Debug(args ...interface{}) {
	l.customLogger.Print("CUSTOM LOG OUTPUT")
}
func (l *SampleCustomLogger) Debugf(format string, args ...interface{}) {
	l.customLogger.Print("CUSTOM LOG OUTPUT")
}
func (l *SampleCustomLogger) Debugln(args ...interface{}) {
	l.customLogger.Print("CUSTOM LOG OUTPUT")
}
func (l *SampleCustomLogger) Info(args ...interface{}) {
	l.customLogger.Print("CUSTOM LOG OUTPUT")
}
func (l *SampleCustomLogger) Infof(format string, args ...interface{}) {
	l.customLogger.Print("CUSTOM LOG OUTPUT")
}
func (l *SampleCustomLogger) Infoln(args ...interface{}) {
	l.customLogger.Print("CUSTOM LOG OUTPUT")
}
func (l *SampleCustomLogger) Warn(args ...interface{}) {
	l.customLogger.Print("CUSTOM LOG OUTPUT")
}
func (l *SampleCustomLogger) Warnf(format string, args ...interface{}) {
	l.customLogger.Print("CUSTOM LOG OUTPUT")
}
func (l *SampleCustomLogger) Warnln(args ...interface{}) {
	l.customLogger.Print("CUSTOM LOG OUTPUT")
}
func (l *SampleCustomLogger) Error(args ...interface{}) { l.customLogger.Print("CUSTOM LOG OUTPUT") }
func (l *SampleCustomLogger) Errorf(format string, args ...interface{}) {
	l.customLogger.Print("CUSTOM LOG OUTPUT")
}
func (l *SampleCustomLogger) Errorln(args ...interface{}) { l.customLogger.Print("CUSTOM LOG OUTPUT") }

func TestDefaultBehavior(t *testing.T) {

	logger := NewLogger(moduleName)
	//Set custom logger to nil to force default logger
	SetCustomLogger(nil)

	//prepare custom logger for which output is bytes buffer
	var buf bytes.Buffer
	logger.logger.(*deflogger.DefaultLogger).ChangeOutput(&buf)

	//No level set for this module so log level should be info
	utils.VerifyTrue(t, apilogging.INFO == deflogger.GetLevel(moduleName), " default log level is INFO")

	//Test logger.print outputs
	deflogger.VerifyBasicLogging(t, -1, logger.Print, nil, &buf, false)
	deflogger.VerifyBasicLogging(t, -1, logger.Println, nil, &buf, false)
	deflogger.VerifyBasicLogging(t, -1, nil, logger.Printf, &buf, false)

	//Test logger.info outputs
	deflogger.VerifyBasicLogging(t, apilogging.INFO, logger.Info, nil, &buf, false)
	deflogger.VerifyBasicLogging(t, apilogging.INFO, logger.Infoln, nil, &buf, false)
	deflogger.VerifyBasicLogging(t, apilogging.INFO, nil, logger.Infof, &buf, false)

	//Test logger.warn outputs
	deflogger.VerifyBasicLogging(t, apilogging.WARNING, logger.Warn, nil, &buf, false)
	deflogger.VerifyBasicLogging(t, apilogging.WARNING, logger.Warnln, nil, &buf, false)
	deflogger.VerifyBasicLogging(t, apilogging.WARNING, nil, logger.Warnf, &buf, false)

	//Test logger.error outputs
	deflogger.VerifyBasicLogging(t, apilogging.ERROR, logger.Error, nil, &buf, false)
	deflogger.VerifyBasicLogging(t, apilogging.ERROR, logger.Errorln, nil, &buf, false)
	deflogger.VerifyBasicLogging(t, apilogging.ERROR, nil, logger.Errorf, &buf, false)

	/*
		SINCE DEBUG LOG IS NOT YET ENABLED, LOG OUTPUT SHOULD BE EMPTY
	*/
	//Test logger.debug outputs when DEBUG level is not enabled
	logger.Debug("brown fox jumps over the lazy dog")
	logger.Debugln("brown fox jumps over the lazy dog")
	logger.Debugf("brown %s jumps over the lazy %s", "fox", "dog")

	utils.VerifyEmpty(t, buf.String(), "debug log isn't supposed to show up for info level")

	//Should be false
	utils.VerifyFalse(t, deflogger.IsEnabledFor(moduleName, apilogging.DEBUG), "logging.IsEnabled for is not working as expected, expected false but got true")

	//Now change the log level to DEBUG
	deflogger.SetLevel(moduleName, apilogging.DEBUG)

	//Should be false
	utils.VerifyTrue(t, deflogger.IsEnabledFor(moduleName, apilogging.DEBUG), "logging.IsEnabled for is not working as expected, expected true but got false")

	//Test logger.debug outputs
	deflogger.VerifyBasicLogging(t, apilogging.DEBUG, logger.Debug, nil, &buf, false)
	deflogger.VerifyBasicLogging(t, apilogging.DEBUG, logger.Debugln, nil, &buf, false)
	deflogger.VerifyBasicLogging(t, apilogging.DEBUG, nil, logger.Debugf, &buf, false)

}
