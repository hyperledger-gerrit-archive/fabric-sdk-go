/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package logging

import (
	"sync"

	"fmt"

	"github.com/hyperledger/fabric-sdk-go/api/apilogging"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging/deflogger"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging/utils"
)

var mutex = &sync.Mutex{}

//Logger basic implementation of api.Logger interface
type Logger struct {
	logger apilogging.Logger
	module string
}

var loggingProvider apilogging.LoggingProvider

const (
	//loggerNotInitializedMsg is used when a logger is not initialized before logging
	loggerNotInitializedMsg = "logger not initialized"
)

// GetLogger creates and returns a Logger object based on the module name.
func GetLogger(module string) (*Logger, error) {
	var logger apilogging.Logger
	if loggingProvider != nil {
		logger = loggingProvider.GetLogger(module)
	}
	return &Logger{logger: logger, module: module}, nil
}

// NewLogger is like GetLogger but panics if the logger can't be created.
func NewLogger(module string) *Logger {
	logger, err := GetLogger(module)
	if err != nil {
		panic("logger: " + module + ": " + err.Error())
	}
	return logger
}

//InitLogger sets new logger which takes over logging operations.
//It is recommended to call this function before making any loggings.
func InitLogger(newLoggingProvider apilogging.LoggingProvider) {
	mutex.Lock()
	defer mutex.Unlock()
	loggingProvider = newLoggingProvider
}

//IsLoggerInitialized returns true logging provider is set already
func IsLoggerInitialized() bool {
	mutex.Lock()
	defer mutex.Unlock()
	return loggingProvider != nil
}

//SetLevel - setting log level for given module
func SetLevel(module string, level apilogging.Level) {
	deflogger.SetLevel(module, level)
}

//GetLevel - getting log level for given module
func GetLevel(module string) apilogging.Level {
	return deflogger.GetLevel(module)
}

//IsEnabledFor - Check if given log level is enabled for given module
func IsEnabledFor(module string, level apilogging.Level) bool {
	return deflogger.IsEnabledFor(module, level)
}

// LogLevel returns the log level from a string representation.
func LogLevel(level string) (apilogging.Level, error) {
	return utils.LogLevel(level)
}

//Fatal calls Fatal function of underlying logger
func (l *Logger) Fatal(args ...interface{}) {
	if l.checkLogger() {
		l.logger.Fatal(args...)
	}
}

//Fatalf calls Fatalf function of underlying logger
func (l *Logger) Fatalf(format string, args ...interface{}) {
	if l.checkLogger() {
		l.logger.Fatalf(format, args...)
	}
}

//Fatalln calls Fatalln function of underlying logger
func (l *Logger) Fatalln(args ...interface{}) {
	if l.checkLogger() {
		l.logger.Fatalln(args...)
	}
}

//Panic calls Panic function of underlying logger
func (l *Logger) Panic(args ...interface{}) {
	if l.checkLogger() {
		l.logger.Panic(args...)
	}
}

//Panicf calls Panicf function of underlying logger
func (l *Logger) Panicf(format string, args ...interface{}) {
	if l.checkLogger() {
		l.logger.Panicf(format, args...)
	}
}

//Panicln calls Panicln function of underlying logger
func (l *Logger) Panicln(args ...interface{}) {
	if l.checkLogger() {
		l.logger.Panicln(args...)
	}
}

//Print calls Print function of underlying logger
func (l *Logger) Print(args ...interface{}) {
	if l.checkLogger() {
		l.logger.Print(args...)
	}
}

//Printf calls Printf function of underlying logger
func (l *Logger) Printf(format string, args ...interface{}) {
	if l.checkLogger() {
		l.logger.Printf(format, args...)
	}
}

//Println calls Println function of underlying logger
func (l *Logger) Println(args ...interface{}) {
	if l.checkLogger() {
		l.logger.Println(args...)
	}
}

//Debug calls Debug function of underlying logger
func (l *Logger) Debug(args ...interface{}) {
	if l.checkLogger() {
		l.logger.Debug(args...)
	}
}

//Debugf calls Debugf function of underlying logger
func (l *Logger) Debugf(format string, args ...interface{}) {
	if l.checkLogger() {
		l.logger.Debugf(format, args...)
	}
}

//Debugln calls Debugln function of underlying logger
func (l *Logger) Debugln(args ...interface{}) {
	if l.checkLogger() {
		l.logger.Debugln(args...)
	}
}

//Info calls Info function of underlying logger
func (l *Logger) Info(args ...interface{}) {
	if l.checkLogger() {
		l.logger.Info(args...)
	}
}

//Infof calls Infof function of underlying logger
func (l *Logger) Infof(format string, args ...interface{}) {
	if l.checkLogger() {
		l.logger.Infof(format, args...)
	}
}

//Infoln calls Infoln function of underlying logger
func (l *Logger) Infoln(args ...interface{}) {
	if l.checkLogger() {
		l.logger.Infoln(args...)
	}
}

//Warn calls Warn function of underlying logger
func (l *Logger) Warn(args ...interface{}) {
	if l.checkLogger() {
		l.logger.Warn(args...)
	}
}

//Warnf calls Warnf function of underlying logger
func (l *Logger) Warnf(format string, args ...interface{}) {
	if l.checkLogger() {
		l.logger.Warnf(format, args...)
	}
}

//Warnln calls Warnln function of underlying logger
func (l *Logger) Warnln(args ...interface{}) {
	if l.checkLogger() {
		l.logger.Warnln(args...)
	}
}

//Error calls Error function of underlying logger
func (l *Logger) Error(args ...interface{}) {
	if l.checkLogger() {
		l.logger.Error(args...)
	}
}

//Errorf calls Errorf function of underlying logger
func (l *Logger) Errorf(format string, args ...interface{}) {
	if l.checkLogger() {
		l.logger.Errorf(format, args...)
	}
}

//Errorln calls Errorln function of underlying logger
func (l *Logger) Errorln(args ...interface{}) {
	if l.checkLogger() {
		l.logger.Errorln(args...)
	}
}

func (l *Logger) checkLogger() bool {

	if l.logger != nil {
		return true
	}

	if loggingProvider == nil {
		fmt.Println(loggerNotInitializedMsg)
		return false
	}

	//TODO: atomic.SwapPointer needs to be used instead of loadLoggerFromFactory() call
	// Race condition problems found in concurrency testing
	//atomic.SwapPointer((*unsafe.Pointer)((unsafe.Pointer)(&l.logger)), ((*unsafe.Pointer)(&newLogger)))
	l.loadLoggerFromFactory()
	return true
}

func (l *Logger) loadLoggerFromFactory() {
	mutex.Lock()
	defer mutex.Unlock()
	l.logger = loggingProvider.GetLogger(l.module)
}
