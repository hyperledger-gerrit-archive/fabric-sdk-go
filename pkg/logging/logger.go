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
)

//Logger basic implementation of api.Logger interface
type Logger struct {
	logger *log.Logger
	module string
}

var moduleLevels = moduleLeveled{}

const (
	logLevelFormatter  = "UTC - %s -> %s "
	logPrefixFormatter = " [%s] "
)

// GetLogger creates and returns a Logger object based on the module name.
func GetLogger(module string) (*Logger, error) {
	//var buf bytes.Buffer
	newLogger := log.New(os.Stdout, fmt.Sprintf(logPrefixFormatter, module), log.Ldate|log.Ltime|log.LUTC)
	return &Logger{logger: newLogger, module: module}, nil
}

// NewLogger is like GetLogger but panics if the logger can't be created.
func NewLogger(module string) *Logger {
	logger, err := GetLogger(module)
	if err != nil {
		panic("logger: " + module + ": " + err.Error())
	}
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

// Fatal is CRITICAL log followed by a call to os.Exit(1).
func (l *Logger) Fatal(args ...interface{}) {
	l.log(CRITICAL, args...)
	l.logger.Fatal(args...)
}

// Fatalf is CRITICAL log formatted followed by a call to os.Exit(1).
func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.logf(CRITICAL, format, args...)
	l.logger.Fatalf(format, args...)
}

// Fatalln is CRITICAL log ln followed by a call to os.Exit(1).
func (l *Logger) Fatalln(args ...interface{}) {
	l.logln(CRITICAL, args...)
	l.logger.Fatalln(args...)
}

// Panic is CRITICAL log followed by a call to panic()
func (l *Logger) Panic(args ...interface{}) {
	l.log(CRITICAL, args...)
	l.logger.Panic(args...)
}

// Panicf is CRITICAL log formatted followed by a call to panic()
func (l *Logger) Panicf(format string, args ...interface{}) {
	l.logf(CRITICAL, format, args...)
	l.logger.Panicf(format, args...)
}

// Panicln is CRITICAL log ln followed by a call to panic()
func (l *Logger) Panicln(args ...interface{}) {
	l.logln(CRITICAL, args...)
	l.logger.Panicln(args...)
}

// Print calls l.Output to print to the logger.
// Arguments are handled in the manner of fmt.Print.
func (l *Logger) Print(args ...interface{}) {
	l.logger.Print(args...)
}

// Printf calls l.Output to print to the logger.
// Arguments are handled in the manner of fmt.Printf.
func (l *Logger) Printf(format string, args ...interface{}) {
	l.logger.Printf(format, args...)
}

// Println calls l.Output to print to the logger.
// Arguments are handled in the manner of fmt.Println.
func (l *Logger) Println(args ...interface{}) {
	l.logger.Println(args...)
}

// Debug calls l.Output to print to the logger.
// Arguments are handled in the manner of fmt.Print.
func (l *Logger) Debug(args ...interface{}) {
	l.log(DEBUG, args...)
}

// Debugf calls l.Output to print to the logger.
// Arguments are handled in the manner of fmt.Printf.
func (l *Logger) Debugf(format string, args ...interface{}) {
	l.logf(DEBUG, format, args...)
}

// Debugln calls l.Output to print to the logger.
// Arguments are handled in the manner of fmt.Println.
func (l *Logger) Debugln(args ...interface{}) {
	l.logln(DEBUG, args...)
}

// Info calls l.Output to print to the logger.
// Arguments are handled in the manner of fmt.Print.
func (l *Logger) Info(args ...interface{}) {
	l.log(INFO, args...)
}

// Infof calls l.Output to print to the logger.
// Arguments are handled in the manner of fmt.Printf.
func (l *Logger) Infof(format string, args ...interface{}) {
	l.logf(INFO, format, args...)
}

// Infoln calls l.Output to print to the logger.
// Arguments are handled in the manner of fmt.Println.
func (l *Logger) Infoln(args ...interface{}) {
	l.logln(INFO, args...)
}

// Warn calls l.Output to print to the logger.
// Arguments are handled in the manner of fmt.Print.
func (l *Logger) Warn(args ...interface{}) {
	l.log(WARNING, args...)
}

// Warnf calls l.Output to print to the logger.
// Arguments are handled in the manner of fmt.Printf.
func (l *Logger) Warnf(format string, args ...interface{}) {
	l.logf(WARNING, format, args...)
}

// Warnln calls l.Output to print to the logger.
// Arguments are handled in the manner of fmt.Println.
func (l *Logger) Warnln(args ...interface{}) {
	l.logln(WARNING, args...)
}

// Error calls l.Output to print to the logger.
// Arguments are handled in the manner of fmt.Print.
func (l *Logger) Error(args ...interface{}) {
	l.log(ERROR, args...)
}

// Errorf calls l.Output to print to the logger.
// Arguments are handled in the manner of fmt.Printf.
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.logf(ERROR, format, args...)
}

// Errorln calls l.Output to print to the logger.
// Arguments are handled in the manner of fmt.Println.
func (l *Logger) Errorln(args ...interface{}) {
	l.logln(ERROR, args...)
}

func (l *Logger) logf(level Level, format string, args ...interface{}) {
	if IsEnabledFor(level, l.module) {
		//Format prefix to show function name and log level and to indicate that timezone used is UTC
		customPrefix := fmt.Sprintf(logLevelFormatter, l.getCaller(), level)
		l.logger.Output(2, customPrefix+fmt.Sprintf(format, args...))
	}
}

func (l *Logger) log(level Level, args ...interface{}) {
	if IsEnabledFor(level, l.module) {
		//Format prefix to show function name and log level and to indicate that timezone used is UTC
		customPrefix := fmt.Sprintf(logLevelFormatter, l.getCaller(), level)
		l.logger.Output(2, customPrefix+fmt.Sprint(args...))
	}
}

func (l *Logger) logln(level Level, args ...interface{}) {
	if IsEnabledFor(level, l.module) {
		//Format prefix to show function name and log level and to indicate that timezone used is UTC
		customPrefix := fmt.Sprintf(logLevelFormatter, l.getCaller(), level)
		l.logger.Output(2, customPrefix+fmt.Sprintln(args...))
	}
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
