/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package logging

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/hyperledger/fabric-sdk-go/api/apilogging"
)

//Logger basic implementation of api.Logger interface
type Logger struct {
	defaultLogger *log.Logger
	customLogger  apilogging.Logger
	module        string
}

var moduleLevels = moduleLeveled{}

type customLogf func(string, ...interface{})
type customLog func(...interface{})

const (
	logLevelFormatter  = "UTC - %s -> %s "
	logPrefixFormatter = " [%s] "
)

// GetLogger creates and returns a Logger object based on the module name.
func GetLogger(module string) (*Logger, error) {
	newLogger := log.New(os.Stdout, fmt.Sprintf(logPrefixFormatter, module), log.Ldate|log.Ltime|log.LUTC)
	return &Logger{defaultLogger: newLogger, module: module, customLogger: &EmptyCustomLogger{}}, nil
}

// NewLogger is like GetLogger but panics if the logger can't be created.
func NewLogger(module string) *Logger {
	logger, err := GetLogger(module)
	if err != nil {
		panic("logger: " + module + ": " + err.Error())
	}
	return logger
}

// NewCustomLogger is like NewLogger where custom logger is plugged in, it panics if the logger can't be created.
func NewCustomLogger(module string, customLogger apilogging.Logger) *Logger {
	logger := NewLogger(module)
	logger.AddCustomLogger(customLogger)
	return logger
}

// SetLevel sets the logging level for the specified module. The module
// corresponds to the string specified in GetLogger.
func SetLevel(level Level, module string) {
	moduleLevels.SetLevel(level, module)
}

// GetLevel returns the logging level for the specified module.
func GetLevel(module string) Level {
	return moduleLevels.GetLevel(module)
}

// IsEnabledFor will return true if logging is enabled for the given module.
func IsEnabledFor(level Level, module string) bool {
	return moduleLevels.IsEnabledFor(level, module)
}

func (l *Logger) AddCustomLogger(customLogger apilogging.Logger) {
	if customLogger != nil {
		l.customLogger = customLogger
	}
}

// Fatal is CRITICAL log followed by a call to os.Exit(1).
func (l *Logger) Fatal(args ...interface{}) {
	if l.isCustomLoggerAvailable() {
		l.customLogger.Fatal(args...)
		return
	}
	l.log(l.customLogger.Fatal, CRITICAL, args...)
	l.defaultLogger.Fatal(args...)
}

// Fatalf is CRITICAL log formatted followed by a call to os.Exit(1).
func (l *Logger) Fatalf(format string, args ...interface{}) {
	if l.isCustomLoggerAvailable() {
		l.customLogger.Fatalf(format, args...)
		return
	}
	l.logf(l.customLogger.Fatalf, CRITICAL, format, args...)
	l.defaultLogger.Fatalf(format, args...)
}

// Fatalln is CRITICAL log ln followed by a call to os.Exit(1).
func (l *Logger) Fatalln(args ...interface{}) {
	if l.isCustomLoggerAvailable() {
		l.customLogger.Fatalln(args...)
		return
	}
	l.logln(l.customLogger.Fatalln, CRITICAL, args...)
	l.defaultLogger.Fatalln(args...)
}

// Panic is CRITICAL log followed by a call to panic()
func (l *Logger) Panic(args ...interface{}) {
	if l.isCustomLoggerAvailable() {
		l.customLogger.Panic(args...)
		return
	}
	l.log(l.customLogger.Panic, CRITICAL, args...)
	l.defaultLogger.Panic(args...)
}

// Panicf is CRITICAL log formatted followed by a call to panic()
func (l *Logger) Panicf(format string, args ...interface{}) {
	if l.isCustomLoggerAvailable() {
		l.customLogger.Panicf(format, args...)
		return
	}
	l.logf(l.customLogger.Panicf, CRITICAL, format, args...)
	l.defaultLogger.Panicf(format, args...)
}

// Panicln is CRITICAL log ln followed by a call to panic()
func (l *Logger) Panicln(args ...interface{}) {
	if l.isCustomLoggerAvailable() {
		l.customLogger.Panicln(args...)
		return
	}
	l.logln(l.customLogger.Panicln, CRITICAL, args...)
	l.defaultLogger.Panicln(args...)
}

// Print calls l.Output to print to the logger.
// Arguments are handled in the manner of fmt.Print.
func (l *Logger) Print(args ...interface{}) {
	if l.isCustomLoggerAvailable() {
		l.customLogger.Print(args...)
		return
	}
	l.defaultLogger.Print(args...)
}

// Printf calls l.Output to print to the logger.
// Arguments are handled in the manner of fmt.Printf.
func (l *Logger) Printf(format string, args ...interface{}) {
	if l.isCustomLoggerAvailable() {
		l.customLogger.Printf(format, args...)
		return
	}
	l.defaultLogger.Printf(format, args...)
}

// Println calls l.Output to print to the logger.
// Arguments are handled in the manner of fmt.Println.
func (l *Logger) Println(args ...interface{}) {
	if l.isCustomLoggerAvailable() {
		l.customLogger.Println(args...)
		return
	}
	l.defaultLogger.Println(args...)
}

// Debug calls l.Output to print to the logger.
// Arguments are handled in the manner of fmt.Print.
func (l *Logger) Debug(args ...interface{}) {
	l.log(l.customLogger.Debug, DEBUG, args...)
}

// Debugf calls l.Output to print to the logger.
// Arguments are handled in the manner of fmt.Printf.
func (l *Logger) Debugf(format string, args ...interface{}) {
	l.logf(l.customLogger.Debugf, DEBUG, format, args...)
}

// Debugln calls l.Output to print to the logger.
// Arguments are handled in the manner of fmt.Println.
func (l *Logger) Debugln(args ...interface{}) {
	l.logln(l.customLogger.Debugln, DEBUG, args...)
}

// Info calls l.Output to print to the logger.
// Arguments are handled in the manner of fmt.Print.
func (l *Logger) Info(args ...interface{}) {
	l.log(l.customLogger.Info, INFO, args...)
}

// Infof calls l.Output to print to the logger.
// Arguments are handled in the manner of fmt.Printf.
func (l *Logger) Infof(format string, args ...interface{}) {
	l.logf(l.customLogger.Infof, INFO, format, args...)
}

// Infoln calls l.Output to print to the logger.
// Arguments are handled in the manner of fmt.Println.
func (l *Logger) Infoln(args ...interface{}) {
	l.logln(l.customLogger.Infoln, INFO, args...)
}

// Warn calls l.Output to print to the logger.
// Arguments are handled in the manner of fmt.Print.
func (l *Logger) Warn(args ...interface{}) {
	l.log(l.customLogger.Warn, WARNING, args...)
}

// Warnf calls l.Output to print to the logger.
// Arguments are handled in the manner of fmt.Printf.
func (l *Logger) Warnf(format string, args ...interface{}) {
	l.logf(l.customLogger.Warnf, WARNING, format, args...)
}

// Warnln calls l.Output to print to the logger.
// Arguments are handled in the manner of fmt.Println.
func (l *Logger) Warnln(args ...interface{}) {
	l.logln(l.customLogger.Warnln, WARNING, args...)
}

// Error calls l.Output to print to the logger.
// Arguments are handled in the manner of fmt.Print.
func (l *Logger) Error(args ...interface{}) {
	l.log(l.customLogger.Error, ERROR, args...)
}

// Errorf calls l.Output to print to the logger.
// Arguments are handled in the manner of fmt.Printf.
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.logf(l.customLogger.Errorf, ERROR, format, args...)
}

// Errorln calls l.Output to print to the logger.
// Arguments are handled in the manner of fmt.Println.
func (l *Logger) Errorln(args ...interface{}) {
	l.logln(l.customLogger.Errorln, ERROR, args...)
}

func (l *Logger) logf(customFunc customLogf, level Level, format string, args ...interface{}) {
	if IsEnabledFor(level, l.module) {
		if l.isCustomLoggerAvailable() {
			customFunc(format, args...)
			return
		}
		//Format prefix to show function name and log level and to indicate that timezone used is UTC
		customPrefix := fmt.Sprintf(logLevelFormatter, l.getCaller(), level)
		l.defaultLogger.Output(2, customPrefix+fmt.Sprintf(format, args...))
	}
}

func (l *Logger) log(customFunc customLog, level Level, args ...interface{}) {
	if IsEnabledFor(level, l.module) {
		if l.isCustomLoggerAvailable() {
			customFunc(args...)
			return
		}
		//Format prefix to show function name and log level and to indicate that timezone used is UTC
		customPrefix := fmt.Sprintf(logLevelFormatter, l.getCaller(), level)
		l.defaultLogger.Output(2, customPrefix+fmt.Sprint(args...))
	}
}

func (l *Logger) logln(customFunc customLog, level Level, args ...interface{}) {
	if IsEnabledFor(level, l.module) {
		if l.isCustomLoggerAvailable() {
			customFunc(args...)
			return
		}
		//Format prefix to show function name and log level and to indicate that timezone used is UTC
		customPrefix := fmt.Sprintf(logLevelFormatter, l.getCaller(), level)
		l.defaultLogger.Output(2, customPrefix+fmt.Sprintln(args...))
	}
}

func (l *Logger) isCustomLoggerAvailable() bool {
	_, ok := l.customLogger.(*EmptyCustomLogger)
	return !ok
}

// getCaller utility to find caller function used to mention in log lines
func (l *Logger) getCaller() string {
	fpcs := make([]uintptr, 1)
	// skip 3 levels to get to the caller of whoever called getCaller()
	n := runtime.Callers(4, fpcs)
	if n == 0 {
		return "n/a"
	}

	fun := runtime.FuncForPC(fpcs[0] - 1)
	if fun == nil {
		return "n/a"
	}
	_, funName := filepath.Split(fun.Name())
	return funName
}

type EmptyCustomLogger struct {
	customLogger *log.Logger
}

func (l *EmptyCustomLogger) AddCustomLogger(customLogger apilogging.Logger) {}
func (l *EmptyCustomLogger) Fatal(v ...interface{})                         {}
func (l *EmptyCustomLogger) Fatalf(format string, v ...interface{})         {}
func (l *EmptyCustomLogger) Fatalln(v ...interface{})                       {}
func (l *EmptyCustomLogger) Panic(v ...interface{})                         {}
func (l *EmptyCustomLogger) Panicf(format string, v ...interface{})         {}
func (l *EmptyCustomLogger) Panicln(v ...interface{})                       {}
func (l *EmptyCustomLogger) Print(v ...interface{})                         {}
func (l *EmptyCustomLogger) Printf(format string, v ...interface{})         {}
func (l *EmptyCustomLogger) Println(v ...interface{})                       {}
func (l *EmptyCustomLogger) Debug(args ...interface{})                      {}
func (l *EmptyCustomLogger) Debugf(format string, args ...interface{})      {}
func (l *EmptyCustomLogger) Debugln(args ...interface{})                    {}
func (l *EmptyCustomLogger) Info(args ...interface{})                       {}
func (l *EmptyCustomLogger) Infof(format string, args ...interface{})       {}
func (l *EmptyCustomLogger) Infoln(args ...interface{})                     {}
func (l *EmptyCustomLogger) Warn(args ...interface{})                       {}
func (l *EmptyCustomLogger) Warnf(format string, args ...interface{})       {}
func (l *EmptyCustomLogger) Warnln(args ...interface{})                     {}
func (l *EmptyCustomLogger) Error(args ...interface{})                      {}
func (l *EmptyCustomLogger) Errorf(format string, args ...interface{})      {}
func (l *EmptyCustomLogger) Errorln(args ...interface{})                    {}
