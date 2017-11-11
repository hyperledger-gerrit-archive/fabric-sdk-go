/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package logging

import (
	"errors"
	"sync"

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

var customLogger apilogging.Logger
var customLeveled apilogging.Leveled

// ErrInvalidLogLevel is used when an invalid log level has been used.
var ErrInvalidLogLevel = errors.New("logger: invalid log level")

// GetLogger creates and returns a Logger object based on the module name.
func GetLogger(module string) (*Logger, error) {
	return &Logger{logger: deflogger.GetDefaultLogger(module), module: module}, nil
}

// NewLogger is like GetLogger but panics if the logger can't be created.
func NewLogger(module string) *Logger {
	logger, err := GetLogger(module)
	if err != nil {
		panic("logger: " + module + ": " + err.Error())
	}
	return logger
}

//SetCustomLogger sets new custom logger which takes over logging operations already created and
//new logger which are going to be created. Care should be taken while using this method.
//It is recommended to add Custom loggers before making any loggings.
func SetCustomLogger(newCustomLogger apilogging.Logger) {
	mutex.Lock()
	defer mutex.Unlock()
	customLogger = newCustomLogger
}

//SetCustomLevelled sets new custom log levels with modules which takes over
// levelled modules. Care should be taken while using this function.
//It is recommended to call SetCustomLevelled before making any loggings.
func SetCustomLevelled(customLeveledModules apilogging.Leveled) {
	mutex.Lock()
	defer mutex.Unlock()
	customLeveled = customLeveledModules
}

//SetLevel - setting log level for given module
func SetLevel(module string, level apilogging.Level) {
	if customLeveled != nil {
		customLeveled.SetLevel(module, level)
	}
	deflogger.SetLevel(module, level)
}

//GetLevel - getting log level for given module
func GetLevel(module string) apilogging.Level {
	if customLeveled != nil {
		return customLeveled.GetLevel(module)
	}
	return deflogger.GetLevel(module)
}

//IsEnabledFor - Check if given log level is enabled for given module
func IsEnabledFor(module string, level apilogging.Level) bool {
	if customLeveled != nil {
		return customLeveled.IsEnabledFor(module, level)
	}
	return deflogger.IsEnabledFor(module, level)
}

// LogLevel returns the log level from a string representation.
func LogLevel(level string) (apilogging.Level, error) {
	return utils.LogLevel(level)
}

//Fatal calls Fatal function of underlying logger
func (l *Logger) Fatal(args ...interface{}) {
	l.getCurrentLogger().Fatal(args...)
}

//Fatalf calls Fatalf function of underlying logger
func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.getCurrentLogger().Fatalf(format, args...)
}

//Fatalln calls Fatalln function of underlying logger
func (l *Logger) Fatalln(args ...interface{}) {
	l.getCurrentLogger().Fatalln(args...)
}

//Panic calls Panic function of underlying logger
func (l *Logger) Panic(args ...interface{}) {
	l.getCurrentLogger().Panic(args...)
}

//Panicf calls Panicf function of underlying logger
func (l *Logger) Panicf(format string, args ...interface{}) {
	l.getCurrentLogger().Panicf(format, args...)
}

//Panicln calls Panicln function of underlying logger
func (l *Logger) Panicln(args ...interface{}) {
	l.getCurrentLogger().Panicln(args...)
}

//Print calls Print function of underlying logger
func (l *Logger) Print(args ...interface{}) {
	l.getCurrentLogger().Print(args...)
}

//Printf calls Printf function of underlying logger
func (l *Logger) Printf(format string, args ...interface{}) {
	l.getCurrentLogger().Printf(format, args...)
}

//Println calls Println function of underlying logger
func (l *Logger) Println(args ...interface{}) {
	l.getCurrentLogger().Println(args...)
}

//Debug calls Debug function of underlying logger
func (l *Logger) Debug(args ...interface{}) {
	l.getCurrentLogger().Debug(args...)
}

//Debugf calls Debugf function of underlying logger
func (l *Logger) Debugf(format string, args ...interface{}) {
	l.getCurrentLogger().Debugf(format, args...)
}

//Debugln calls Debugln function of underlying logger
func (l *Logger) Debugln(args ...interface{}) {
	l.getCurrentLogger().Debugln(args...)
}

//Info calls Info function of underlying logger
func (l *Logger) Info(args ...interface{}) {
	l.getCurrentLogger().Info(args...)
}

//Infof calls Infof function of underlying logger
func (l *Logger) Infof(format string, args ...interface{}) {
	l.getCurrentLogger().Infof(format, args...)
}

//Infoln calls Infoln function of underlying logger
func (l *Logger) Infoln(args ...interface{}) {
	l.getCurrentLogger().Infoln(args...)
}

//Warn calls Warn function of underlying logger
func (l *Logger) Warn(args ...interface{}) {
	l.getCurrentLogger().Warn(args...)
}

//Warnf calls Warnf function of underlying logger
func (l *Logger) Warnf(format string, args ...interface{}) {
	l.getCurrentLogger().Warnf(format, args...)
}

//Warnln calls Warnln function of underlying logger
func (l *Logger) Warnln(args ...interface{}) {
	l.getCurrentLogger().Warnln(args...)
}

//Error calls Error function of underlying logger
func (l *Logger) Error(args ...interface{}) {
	l.getCurrentLogger().Error(args...)
}

//Errorf calls Errorf function of underlying logger
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.getCurrentLogger().Errorf(format, args...)
}

//Errorln calls Errorln function of underlying logger
func (l *Logger) Errorln(args ...interface{}) {
	l.getCurrentLogger().Errorln(args...)
}

//getCurrentLogger - returns customlogger is set, or default logger
func (l *Logger) getCurrentLogger() apilogging.Logger {
	if customLogger != nil {
		return customLogger
	}
	return l.logger
}
