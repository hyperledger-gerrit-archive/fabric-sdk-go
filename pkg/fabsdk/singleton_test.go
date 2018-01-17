// +build testing

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabsdk

import (
	"testing"

	"github.com/hyperledger/fabric-sdk-go/api/apilogging"
	configImpl "github.com/hyperledger/fabric-sdk-go/pkg/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging/deflogger"
)

func TestDefLoggerFactory(t *testing.T) {
	// Cleanup logging singleton
	logging.UnsafeReset()

	config, err := configImpl.InitConfig("../../test/fixtures/config/config_test.yaml")

	if err != nil {
		t.Fatalf("Error loading config: %s", err)
	}

	_, err = New(config)

	if err != nil {
		t.Fatalf("Error initializing SDK: %s", err)
	}

	const moduleName = "mymodule"
	l1, err := logging.GetLogger(moduleName)
	if err != nil {
		t.Fatal("Unexpected error getting logger")
	}

	// output a log message to force initializatin
	l1.Info("message")

	// ensure that the logger cannot be overridden
	// (initializing a new logger should have no effect)
	lf := NewMockLoggerFactory()
	logging.InitLogger(lf)

	l2, err := logging.GetLogger(moduleName)
	if err != nil {
		t.Fatal("Unexpected error getting logger")
	}

	// output a log message to force initializatin
	l2.Info("message")

	if lf.ActiveModules[moduleName] {
		t.Fatal("Unexpected logger factory is set")
	}
}

func TestOptLoggerFactory(t *testing.T) {
	// Cleanup logging singleton
	logging.UnsafeReset()

	lf := NewMockLoggerFactory()

	config, err := configImpl.InitConfig("../../test/fixtures/config/config_test.yaml")

	if err != nil {
		t.Fatalf("Error loading config: %s", err)
	}

	_, err = New(config, WithLoggerProvider(lf))

	if err != nil {
		t.Fatalf("Error initializing SDK: %s", err)
	}

	const moduleName = "mymodule"
	l, err := logging.GetLogger(moduleName)
	if err != nil {
		t.Fatal("Unexpected error getting logger")
	}

	// output a log message to force initializatin
	l.Info("message")

	if !lf.ActiveModules[moduleName] {
		t.Fatal("Unexpected logger factory is set")
	}
}

// MockLoggerFactory records the modules that have loggers
type MockLoggerFactory struct {
	ActiveModules map[string]bool
	logger        apilogging.LoggerProvider
}

func NewMockLoggerFactory() *MockLoggerFactory {
	lf := MockLoggerFactory{}
	lf.ActiveModules = make(map[string]bool)
	lf.logger = deflogger.LoggerProvider()

	return &lf
}

func (lf *MockLoggerFactory) GetLogger(module string) apilogging.Logger {
	lf.ActiveModules[module] = true
	return lf.logger.GetLogger(module)
}
