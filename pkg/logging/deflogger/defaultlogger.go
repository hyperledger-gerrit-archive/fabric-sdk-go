/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package deflogger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"sync"

	"io"

	"github.com/hyperledger/fabric-sdk-go/api/apilogging"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging/utils"
)

var rwmutex = &sync.RWMutex{}
var moduleLevels apilogging.Leveled = &moduleLeveled{}
var callerInfos = callerInfo{}

//GetDefaultLogger returns default logger implementation
func GetDefaultLogger(module string) apilogging.Logger {
	newLogger := log.New(os.Stdout, fmt.Sprintf(logPrefixFormatter, module), log.Ldate|log.Ltime|log.LUTC)
	return &DefaultLogger{defaultLogger: newLogger, module: module}
}

//DefaultLogger default underlying logger used by logging.Logger
type DefaultLogger struct {
	defaultLogger *log.Logger
	module        string
}

//LoggerOpts  for all logger customization options
type LoggerOpts struct {
	levelEnabled      bool
	callerInfoEnabled bool
}

const (
	logLevelFormatter   = "UTC %s-> %4.4s "
	logPrefixFormatter  = " [%s] "
	callerInfoFormatter = "- %s "
)

//SetLevel - setting log level for given module
func SetLevel(module string, level apilogging.Level) {
	rwmutex.Lock()
	defer rwmutex.Unlock()
	moduleLevels.SetLevel(module, level)
}

//GetLevel - getting log level for given module
func GetLevel(module string) apilogging.Level {
	rwmutex.RLock()
	defer rwmutex.RUnlock()
	return moduleLevels.GetLevel(module)
}

//IsEnabledFor - Check if given log level is enabled for given module
func IsEnabledFor(module string, level apilogging.Level) bool {
	rwmutex.RLock()
	defer rwmutex.RUnlock()
	return moduleLevels.IsEnabledFor(module, level)
}

//ShowCallerInfo - Show caller info in log lines for given log level
func ShowCallerInfo(module string, level apilogging.Level) {
	rwmutex.Lock()
	defer rwmutex.Unlock()
	callerInfos.ShowCallerInfo(module, level)
}

//HideCallerInfo - Do not show caller info in log lines for given log level
func HideCallerInfo(module string, level apilogging.Level) {
	rwmutex.Lock()
	defer rwmutex.Unlock()
	callerInfos.HideCallerInfo(module, level)
}

//getLoggerOpts - returns LoggerOpts which can be used for customization
func getLoggerOpts(module string, level apilogging.Level) *LoggerOpts {
	rwmutex.RLock()
	defer rwmutex.RUnlock()
	return &LoggerOpts{
		levelEnabled:      moduleLevels.IsEnabledFor(module, level),
		callerInfoEnabled: callerInfos.IsCallerInfoEnabled(module, level),
	}
}

// Fatal is CRITICAL log followed by a call to os.Exit(1).
func (l *DefaultLogger) Fatal(args ...interface{}) {

	l.log(apilogging.CRITICAL, args...)
	l.defaultLogger.Fatal(args...)
}

// Fatalf is CRITICAL log formatted followed by a call to os.Exit(1).
func (l *DefaultLogger) Fatalf(format string, args ...interface{}) {
	l.logf(apilogging.CRITICAL, format, args...)
	l.defaultLogger.Fatalf(format, args...)
}

// Fatalln is CRITICAL log ln followed by a call to os.Exit(1).
func (l *DefaultLogger) Fatalln(args ...interface{}) {
	l.logln(apilogging.CRITICAL, args...)
	l.defaultLogger.Fatalln(args...)
}

// Panic is CRITICAL log followed by a call to panic()
func (l *DefaultLogger) Panic(args ...interface{}) {
	l.log(apilogging.CRITICAL, args...)
	l.defaultLogger.Panic(args...)
}

// Panicf is CRITICAL log formatted followed by a call to panic()
func (l *DefaultLogger) Panicf(format string, args ...interface{}) {
	l.logf(apilogging.CRITICAL, format, args...)
	l.defaultLogger.Panicf(format, args...)
}

// Panicln is CRITICAL log ln followed by a call to panic()
func (l *DefaultLogger) Panicln(args ...interface{}) {
	l.logln(apilogging.CRITICAL, args...)
	l.defaultLogger.Panicln(args...)
}

// Print calls l.Output to print to the logger.
// Arguments are handled in the manner of fmt.Print.
func (l *DefaultLogger) Print(args ...interface{}) {
	l.defaultLogger.Print(args...)
}

// Printf calls l.Output to print to the logger.
// Arguments are handled in the manner of fmt.Printf.
func (l *DefaultLogger) Printf(format string, args ...interface{}) {
	l.defaultLogger.Printf(format, args...)
}

// Println calls l.Output to print to the logger.
// Arguments are handled in the manner of fmt.Println.
func (l *DefaultLogger) Println(args ...interface{}) {
	l.defaultLogger.Println(args...)
}

// Debug calls l.Output to print to the logger.
// Arguments are handled in the manner of fmt.Print.
func (l *DefaultLogger) Debug(args ...interface{}) {
	l.log(apilogging.DEBUG, args...)
}

// Debugf calls l.Output to print to the logger.
// Arguments are handled in the manner of fmt.Printf.
func (l *DefaultLogger) Debugf(format string, args ...interface{}) {
	l.logf(apilogging.DEBUG, format, args...)
}

// Debugln calls l.Output to print to the logger.
// Arguments are handled in the manner of fmt.Println.
func (l *DefaultLogger) Debugln(args ...interface{}) {
	l.logln(apilogging.DEBUG, args...)
}

// Info calls l.Output to print to the logger.
// Arguments are handled in the manner of fmt.Print.
func (l *DefaultLogger) Info(args ...interface{}) {
	l.log(apilogging.INFO, args...)
}

// Infof calls l.Output to print to the logger.
// Arguments are handled in the manner of fmt.Printf.
func (l *DefaultLogger) Infof(format string, args ...interface{}) {
	l.logf(apilogging.INFO, format, args...)
}

// Infoln calls l.Output to print to the logger.
// Arguments are handled in the manner of fmt.Println.
func (l *DefaultLogger) Infoln(args ...interface{}) {
	l.logln(apilogging.INFO, args...)
}

// Warn calls l.Output to print to the logger.
// Arguments are handled in the manner of fmt.Print.
func (l *DefaultLogger) Warn(args ...interface{}) {
	l.log(apilogging.WARNING, args...)
}

// Warnf calls l.Output to print to the logger.
// Arguments are handled in the manner of fmt.Printf.
func (l *DefaultLogger) Warnf(format string, args ...interface{}) {
	l.logf(apilogging.WARNING, format, args...)
}

// Warnln calls l.Output to print to the logger.
// Arguments are handled in the manner of fmt.Println.
func (l *DefaultLogger) Warnln(args ...interface{}) {
	l.logln(apilogging.WARNING, args...)
}

// Error calls l.Output to print to the logger.
// Arguments are handled in the manner of fmt.Print.
func (l *DefaultLogger) Error(args ...interface{}) {
	l.log(apilogging.ERROR, args...)
}

// Errorf calls l.Output to print to the logger.
// Arguments are handled in the manner of fmt.Printf.
func (l *DefaultLogger) Errorf(format string, args ...interface{}) {
	l.logf(apilogging.ERROR, format, args...)
}

// Errorln calls l.Output to print to the logger.
// Arguments are handled in the manner of fmt.Println.
func (l *DefaultLogger) Errorln(args ...interface{}) {
	l.logln(apilogging.ERROR, args...)
}

//ChangeOutput for changing output destination for the logger.
func (l *DefaultLogger) ChangeOutput(output io.Writer) {
	l.defaultLogger.SetOutput(output)
}

func (l *DefaultLogger) logf(level apilogging.Level, format string, args ...interface{}) {
	opts := getLoggerOpts(l.module, level)
	if !opts.levelEnabled {
		//Current logging level is disabled
		return
	}
	//Format prefix to show function name and log level and to indicate that timezone used is UTC
	customPrefix := fmt.Sprintf(logLevelFormatter, l.getCallerInfo(opts), utils.LogLevelString(level))
	l.defaultLogger.Output(2, customPrefix+fmt.Sprintf(format, args...))
}

func (l *DefaultLogger) log(level apilogging.Level, args ...interface{}) {
	opts := getLoggerOpts(l.module, level)
	if !opts.levelEnabled {
		//Current logging level is disabled
		return
	}
	//Format prefix to show function name and log level and to indicate that timezone used is UTC
	customPrefix := fmt.Sprintf(logLevelFormatter, l.getCallerInfo(opts), utils.LogLevelString(level))
	l.defaultLogger.Output(2, customPrefix+fmt.Sprint(args...))
}

func (l *DefaultLogger) logln(level apilogging.Level, args ...interface{}) {
	opts := getLoggerOpts(l.module, level)
	if !opts.levelEnabled {
		//Current logging level is disabled
		return
	}
	//Format prefix to show function name and log level and to indicate that timezone used is UTC
	customPrefix := fmt.Sprintf(logLevelFormatter, l.getCallerInfo(opts), utils.LogLevelString(level))
	l.defaultLogger.Output(2, customPrefix+fmt.Sprintln(args...))
}

func (l *DefaultLogger) getCallerInfo(opts *LoggerOpts) string {

	if !opts.callerInfoEnabled {
		return ""
	}

	const MAXCALLERS = 5                           // search MAXCALLERS frames for the real caller
	const SKIPCALLERS = 4                          // skip SKIPCALLERS frames when determining the real caller
	const DEFAULTLOGPREFIX = "apilogging.(Logger)" // LOGPREFIX indicates the upcoming frame contains the real caller and skip the frame
	const LOGPREFIX = "logging.(*Logger)"          // LOGPREFIX indicates the upcoming frame contains the real caller and skip the frame
	const LOGBRIDGEPREFIX = "logbridge."           // LOGBRIDGEPREFIX indicates to skip the frame due to being a logbridge
	const NOTFOUND = "n/a"

	fpcs := make([]uintptr, MAXCALLERS)

	n := runtime.Callers(SKIPCALLERS, fpcs)
	if n == 0 {
		return fmt.Sprintf(callerInfoFormatter, NOTFOUND)
	}

	frames := runtime.CallersFrames(fpcs[:n])
	funcIsNext := false
	for f, more := frames.Next(); more; f, more = frames.Next() {
		_, funName := filepath.Split(f.Function)
		if f.Func == nil || f.Function == "" {
			funName = NOTFOUND // not a function or unknown
		}

		if strings.HasPrefix(funName, LOGPREFIX) || strings.HasPrefix(funName, LOGBRIDGEPREFIX) || strings.HasPrefix(funName, DEFAULTLOGPREFIX) {
			funcIsNext = true
		} else if funcIsNext {
			return fmt.Sprintf(callerInfoFormatter, funName)
		}
	}

	return fmt.Sprintf(callerInfoFormatter, NOTFOUND)
}
