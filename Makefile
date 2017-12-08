#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

# Supported Targets:
# all : runs unit and integration tests
# depend: checks that test dependencies are installed
# depend-install: installs test dependencies
# unit-test: runs all the unit tests
# integration-test: runs all the integration tests
# checks: runs all check conditions (license, spelling, linting)
# clean: stops docker conatainers used for integration testing
# mock-gen: generate mocks needed for testing (using mockgen)
# channel-config-[codelevel]-gen: generates the channel configuration transactions and blocks used by tests
# populate: populates generated files (not included in git) - currently only vendor
# populate-vendor: populate the vendor directory based on the lock
# populate-clean: cleans up populated files (might become part of clean eventually) 
# thirdparty-pin: pulls (and patches) pinned dependencies into the project under internal
#

# Tool commands (overridable)
GO_CMD             ?= go
GO_DEP_CMD         ?= dep
DOCKER_CMD         ?= docker
DOCKER_COMPOSE_CMD ?= docker-compose

# Fabric versions used in the Makefile
FABRIC_STABLE_VERSION           := 1.0.5
FABRIC_STABLE_VERSION_MAJOR     := 1.0
FABRIC_BASEIMAGE_STABLE_VERSION := 0.4.2

FABRIC_PRERELEASE_VERSION       := 1.1.0-preview
FABRIC_PREV_VERSION             := 1.0.0
FABRIC_DEVSTABLE_VERSION_MAJOR  := 1.1

# Build flags (overridable)
GO_LDFLAGS                 ?= -ldflags=-s
GO_TESTFLAGS               ?=
FABRIC_SDK_EXPERIMENTAL    ?= true
FABRIC_SDK_EXTRA_GO_TAGS   ?=
FABRIC_SDK_POPULATE_VENDOR ?= true

# Fabric tool versions (overridable)
FABRIC_TOOLS_VERSION ?= $(FABRIC_STABLE_VERSION)
FABRIC_BASE_VERSION  ?= $(FABRIC_BASEIMAGE_STABLE_VERSION)

# Fabric base docker image (overridable)
FABRIC_BASE_IMAGE   ?= hyperledger/fabric-baseimage
FABRIC_BASE_TAG     ?= $(ARCH)-$(FABRIC_BASE_VERSION)

# Fabric tools docker image (overridable)
FABRIC_TOOLS_IMAGE ?= hyperledger/fabric-tools
FABRIC_TOOLS_TAG   ?= $(ARCH)-$(FABRIC_TOOLS_VERSION)

# Fabric docker registries (overridable)
FABRIC_RELEASE_REGISTRY     ?= registry.hub.docker.com
FABRIC_DEV_REGISTRY         ?= nexus3.hyperledger.org:10001
FABRIC_DEV_REGISTRY_PRE_CMD ?= docker login -u docker -p docker nexus3.hyperledger.org:10001

# Upstream fabric patching (overridable)
THIRDPARTY_FABRIC_CA_BRANCH ?= master
THIRDPARTY_FABRIC_CA_COMMIT ?= v1.1.0-preview
THIRDPARTY_FABRIC_BRANCH    ?= master
THIRDPARTY_FABRIC_COMMIT    ?= 750c1393168aae3f910c8d7831860dbd6f259078

# Force removal of images in cleanup (overridable)
FIXTURE_DOCKER_REMOVE_FORCE ?= false

# Code levels to exercise integration tests against (overridable)
FABRIC_STABLE_CODELEVEL_TEST        ?= true
FABRIC_STABLE_CODELEVEL_PKCS11_TEST ?= false
FABRIC_PREV_CODELEVEL_TEST          ?= false
FABRIC_PRERELEASE_CODELEVEL_TEST    ?= false
FABRIC_DEVSTABLE_CODELEVEL_TEST     ?= false

# Code levels
FABRIC_STABLE_CODELEVEL     := stable
FABRIC_PREV_CODELEVEL       := prev
FABRIC_PRERELEASE_CODELEVEL := prerelease
FABRIC_DEVSTABLE_CODELEVEL  := devstable

# Code level version targets
FABRIC_STABLE_CODELEVEL_VER     := v$(FABRIC_STABLE_VERSION_MAJOR)
FABRIC_PREV_CODELEVEL_VER       := v$(FABRIC_PREV_VERSION)
FABRIC_PRERELEASE_CODELEVEL_VER := v$(FABRIC_PRERELEASE_VERSION)
FABRIC_DEVSTABLE_CODELEVEL_VER  := v$(FABRIC_DEVSTABLE_VERSION_MAJOR)
FABRIC_CODELEVEL_VER            ?= v$(FABRIC_STABLE_CODELEVEL_VER)

# Local variables used by makefile
PACKAGE_NAME         := github.com/hyperledger/fabric-sdk-go
ARCH                 := $(shell uname -m)
FIXTURE_PROJECT_NAME := fabsdkgo
MAKEFILE_THIS        := $(lastword $(MAKEFILE_LIST))

# Fabric tool docker tags at code levels
FABRIC_TOOLS_STABLE_TAG     := $(ARCH)-$(FABRIC_STABLE_VERSION)
FABRIC_TOOLS_PREV_TAG       := $(ARCH)-$(FABRIC_PREV_VERSION)
FABRIC_TOOLS_PRERELEASE_TAG := $(ARCH)-$(FABRIC_PRERELEASE_VERSION)
FABRIC_TOOLS_DEVSTABLE_TAG  := DEV_STABLE

# The version of dep that will be installed by depend-install (or in the CI)
GO_DEP_COMMIT := v0.3.1

# The version of mockgen that will be installed by depend-install
GO_MOCKGEN_COMMIT := v1.0.0

# Setup Go Tags
GO_TAGS := $(FABRIC_SDK_EXTRA_GO_TAGS)
ifeq ($(FABRIC_SDK_EXPERIMENTAL),true)
GO_TAGS += experimental
endif

# Detect CI
# TODO introduce nightly and adjust verify
ifdef JENKINS_URL
export FABRIC_SDKGO_DEPEND_INSTALL  := true

FABRIC_STABLE_CODELEVEL_TEST        := true
FABRIC_STABLE_CODELEVEL_PKCS11_TEST := true
FABRIC_PREV_CODELEVEL_TEST          := true
FABRIC_PRERELEASE_CODELEVEL_TEST    := true
FABRIC_DEVSTABLE_CODELEVEL_TEST     := true
endif

# DEVSTABLE images are currently only x86_64 
ifneq ($(ARCH),x86_64)
FABRIC_DEVSTABLE_CODELEVEL_TEST := false
endif

# Global environment exported for scripts
export GO_CMD
export GO_DEP_CMD
export ARCH
export BASE_ARCH=$(ARCH)
export GO_LDFLAGS
export GO_DEP_COMMIT
export GO_MOCKGEN_COMMIT
export GO_TAGS
export GO_TESTFLAGS
export DOCKER_CMD
export DOCKER_COMPOSE_CMD

.PHONY: all
all: checks unit-test integration-test

.PHONY: depend
depend:
	@test/scripts/dependencies.sh

.PHONY: depend-install
depend-install:
	@FABRIC_SDKGO_DEPEND_INSTALL="true" test/scripts/dependencies.sh

.PHONY: checks
checks: depend license lint spelling

.PHONY: license 
license:
	@test/scripts/check_license.sh

.PHONY: lint
lint: populate
	@test/scripts/check_lint.sh

.PHONY: spelling
spelling:
	@test/scripts/check_spelling.sh

.PHONY: build-softhsm2-image
build-softhsm2-image:
	 @$(DOCKER_CMD) build --no-cache -q -t "softhsm2-image" \
		--build-arg FABRIC_BASE_IMAGE=$(FABRIC_BASE_IMAGE) \
		--build-arg FABRIC_BASE_TAG=$(FABRIC_BASE_TAG) \
		./test/fixtures/softhsm2

.PHONY: unit-test
unit-test: checks depend populate
	@test/scripts/unit.sh

.PHONY: unit-tests
unit-tests: unit-test

.PHONY: integration-tests-stable
integration-tests-stable: clean depend populate
	@cd ./test/fixtures && \
		FABRIC_SDKGO_CODELEVEL_VER=$(FABRIC_STABLE_CODELEVEL_VER) FABRIC_SDKGO_CODELEVEL=$(FABRIC_STABLE_CODELEVEL) FABRIC_DOCKER_REGISTRY=$(FABRIC_RELEASE_REGISTRY)/ $(DOCKER_COMPOSE_CMD) -f docker-compose.yaml -f docker-compose-nopkcs11-test.yaml up --force-recreate --abort-on-container-exit
	@cd test/fixtures && FABRIC_DOCKER_REGISTRY=$(FABRIC_RELEASE_REGISTRY)/ ../scripts/check_status.sh "-f ./docker-compose.yaml -f ./docker-compose-nopkcs11-test.yaml"

.PHONY: integration-tests-prev
integration-tests-prev: clean depend populate
	@. ./test/fixtures/prev-env.sh && \
		cd ./test/fixtures && \
		FABRIC_SDKGO_CODELEVEL_VER=$(FABRIC_PREV_CODELEVEL_VER) FABRIC_SDKGO_CODELEVEL=$(FABRIC_PREV_CODELEVEL) FABRIC_DOCKER_REGISTRY=$(FABRIC_RELEASE_REGISTRY)/ $(DOCKER_COMPOSE_CMD) -f docker-compose.yaml -f docker-compose-nopkcs11-test.yaml up --force-recreate --abort-on-container-exit
	@cd test/fixtures && FABRIC_DOCKER_REGISTRY=$(FABRIC_RELEASE_REGISTRY)/ ../scripts/check_status.sh "-f ./docker-compose.yaml -f ./docker-compose-nopkcs11-test.yaml"

.PHONY: integration-tests-prerelease
integration-tests-prerelease: clean depend populate
	@. ./test/fixtures/prerelease-env.sh && \
		cd ./test/fixtures && \
		FABRIC_SDKGO_CODELEVEL_VER=$(FABRIC_PRERELEASE_CODELEVEL_VER) FABRIC_SDKGO_CODELEVEL=$(FABRIC_PRERELEASE_CODELEVEL) FABRIC_DOCKER_REGISTRY=$(FABRIC_RELEASE_REGISTRY)/ $(DOCKER_COMPOSE_CMD) -f docker-compose.yaml -f docker-compose-nopkcs11-test.yaml up --force-recreate --abort-on-container-exit
	@cd test/fixtures && FABRIC_DOCKER_REGISTRY=$(FABRIC_RELEASE_REGISTRY)/ ../scripts/check_status.sh "-f ./docker-compose.yaml -f ./docker-compose-nopkcs11-test.yaml"

.PHONY: integration-tests-devstable
integration-tests-devstable: clean depend populate
	@. ./test/fixtures/devstable-env.sh && \
		$(FABRIC_DEV_REGISTRY_PRE_CMD) && \
		cd ./test/fixtures && \
		FABRIC_SDKGO_CODELEVEL_VER=$(FABRIC_DEVSTABLE_CODELEVEL_VER) FABRIC_SDKGO_CODELEVEL=$(FABRIC_DEVSTABLE_CODELEVEL) FABRIC_DOCKER_REGISTRY=$(FABRIC_DEV_REGISTRY)/ $(DOCKER_COMPOSE_CMD) -f docker-compose.yaml -f docker-compose-nopkcs11-test.yaml up --force-recreate --abort-on-container-exit
	@cd test/fixtures && FABRIC_DOCKER_REGISTRY=$(FABRIC_DEV_REGISTRY)/ ../scripts/check_status.sh "-f ./docker-compose.yaml -f ./docker-compose-nopkcs11-test.yaml"

.PHONY: integration-tests-stable-pkcs11
integration-tests-stable-pkcs11: clean depend populate build-softhsm2-image
	@cd ./test/fixtures && \
		FABRIC_SDKGO_CODELEVEL_VER=$(FABRIC_STABLE_CODELEVEL_VER) FABRIC_SDKGO_CODELEVEL=$(FABRIC_STABLE_CODELEVEL) FABRIC_DOCKER_REGISTRY=$(FABRIC_RELEASE_REGISTRY)/ $(DOCKER_COMPOSE_CMD) -f docker-compose.yaml -f docker-compose-pkcs11-test.yaml up --force-recreate --abort-on-container-exit
	@cd test/fixtures && FABRIC_DOCKER_REGISTRY=$(FABRIC_RELEASE_REGISTRY)/ ../scripts/check_status.sh "-f ./docker-compose.yaml -f ./docker-compose-pkcs11-test.yaml"

.PHONY: integration-tests
integration-tests: integration-test

.PHONY: integration-test
integration-test:
ifeq ($(FABRIC_STABLE_CODELEVEL_TEST),true)
	@$(MAKE) -f $(MAKEFILE_THIS) clean
	@$(MAKE) -f $(MAKEFILE_THIS) integration-tests-stable
endif
ifeq ($(FABRIC_STABLE_CODELEVEL_PKCS11_TEST),true)
	@$(MAKE) -f $(MAKEFILE_THIS) clean
	@$(MAKE) -f $(MAKEFILE_THIS) integration-tests-stable-pkcs11
endif
ifeq ($(FABRIC_PRERELEASE_CODELEVEL_TEST),true)
	@$(MAKE) -f $(MAKEFILE_THIS) clean
	@$(MAKE) -f $(MAKEFILE_THIS) integration-tests-prerelease
endif
ifeq ($(FABRIC_DEVSTABLE_CODELEVEL_TEST),true)
	@$(MAKE) -f $(MAKEFILE_THIS) clean
	@$(MAKE) -f $(MAKEFILE_THIS) integration-tests-devstable
endif
ifeq ($(FABRIC_PREV_CODELEVEL_TEST),true)
	@$(MAKE) -f $(MAKEFILE_THIS) clean
	@$(MAKE) -f $(MAKEFILE_THIS) integration-tests-prev
endif
	@$(MAKE) -f $(MAKEFILE_THIS) clean

.PHONY: mock-gen
mock-gen:
	mockgen -build_flags '$(GO_LDFLAGS)' github.com/hyperledger/fabric-sdk-go/api/apitxn ProposalProcessor | sed "s/github.com\/hyperledger\/fabric-sdk-go\/vendor\///g" | goimports > api/apitxn/mocks/mockapitxn.gen.go
	mockgen -build_flags '$(GO_LDFLAGS)' github.com/hyperledger/fabric-sdk-go/api/apiconfig Config | sed "s/github.com\/hyperledger\/fabric-sdk-go\/vendor\///g" | goimports > api/apiconfig/mocks/mockconfig.gen.go
	mockgen -build_flags '$(GO_LDFLAGS)' github.com/hyperledger/fabric-sdk-go/api/apifabca FabricCAClient | sed "s/github.com\/hyperledger\/fabric-sdk-go\/vendor\///g" | goimports > api/apifabca/mocks/mockfabriccaclient.gen.go

# TODO - Add cryptogen
.PHONY: channel-config-gen
channel-config-gen:
	@echo "Generating test channel configuration transactions and blocks ..."
	@$(DOCKER_CMD) run -i \
		-v $(abspath .):/opt/gopath/src/$(PACKAGE_NAME) \
		$(FABRIC_TOOLS_IMAGE):$(FABRIC_TOOLS_TAG) \
		/bin/bash -c "FABRIC_VERSION_DIR=fabric-$(FABRIC_CODELEVEL_VER)/ /opt/gopath/src/${PACKAGE_NAME}/test/scripts/generate_channeltx.sh"

.PHONY: channel-config-all-gen
channel-config-all-gen: channel-config-stable-gen channel-config-prev-gen channel-config-prerelease-gen channel-config-devstable-gen

.PHONY: channel-config-stable-gen
channel-config-stable-gen:
	@echo "Generating test channel configuration transactions and blocks (code level stable) ..."
	@$(DOCKER_CMD) run -i \
		-v $(abspath .):/opt/gopath/src/$(PACKAGE_NAME) \
		$(FABRIC_TOOLS_IMAGE):$(FABRIC_TOOLS_STABLE_TAG) \
		/bin/bash -c "FABRIC_VERSION_DIR=fabric-$(FABRIC_STABLE_CODELEVEL_VER)/ /opt/gopath/src/${PACKAGE_NAME}/test/scripts/generate_channeltx.sh"

.PHONY: channel-config-prev-gen
channel-config-prev-gen:
	@echo "Generating test channel configuration transactions and blocks (code level prev) ..."
	$(DOCKER_CMD) run -i \
		-v $(abspath .):/opt/gopath/src/$(PACKAGE_NAME) \
		$(FABRIC_TOOLS_IMAGE):$(FABRIC_TOOLS_PREV_TAG) \
		/bin/bash -c "FABRIC_VERSION_DIR=fabric-$(FABRIC_PREV_CODELEVEL_VER)/ /opt/gopath/src/${PACKAGE_NAME}/test/scripts/generate_channeltx.sh"

.PHONY: channel-config-prerelease-gen
channel-config-prerelease-gen:
	@echo "Generating test channel configuration transactions and blocks (code level prerelease) ..."
	$(DOCKER_CMD) run -i \
		-v $(abspath .):/opt/gopath/src/$(PACKAGE_NAME) \
		$(FABRIC_TOOLS_IMAGE):$(FABRIC_TOOLS_PRERELEASE_TAG) \
		/bin/bash -c "FABRIC_VERSION_DIR=fabric-$(FABRIC_PRERELEASE_CODELEVEL_VER)/ /opt/gopath/src/${PACKAGE_NAME}/test/scripts/generate_channeltx.sh"

.PHONY: channel-config-devstable-gen
channel-config-devstable-gen:
	@echo "Generating test channel configuration transactions and blocks (code level devstable) ..."
	@$(FABRIC_DEV_REGISTRY_PRE_CMD) && \
		$(DOCKER_CMD) run -i \
			-v $(abspath .):/opt/gopath/src/$(PACKAGE_NAME) \
			$(FABRIC_DEV_REGISTRY)/$(FABRIC_TOOLS_IMAGE):$(FABRIC_TOOLS_DEVSTABLE_TAG) \
			/bin/bash -c "FABRIC_VERSION_DIR=fabric-$(FABRIC_DEVSTABLE_CODELEVEL_VER)/ /opt/gopath/src/${PACKAGE_NAME}/test/scripts/generate_channeltx.sh"

.PHONY: thirdparty-pin
thirdparty-pin:
	@echo "Pinning third party packages ..."
	@UPSTREAM_COMMIT=$(THIRDPARTY_FABRIC_COMMIT) UPSTREAM_BRANCH=$(THIRDPARTY_FABRIC_BRANCH) scripts/third_party_pins/fabric/apply_upstream.sh
	@UPSTREAM_COMMIT=$(THIRDPARTY_FABRIC_CA_COMMIT) UPSTREAM_BRANCH=$(THIRDPARTY_FABRIC_CA_BRANCH) scripts/third_party_pins/fabric-ca/apply_upstream.sh

.PHONY: populate
populate: populate-vendor

.PHONY: populate-vendor
populate-vendor:
ifeq ($(FABRIC_SDK_POPULATE_VENDOR),true)
	@echo "Populating vendor ..."
	@$(GO_DEP_CMD) ensure -vendor-only
endif

.PHONY: populate-clean
populate-clean:
	rm -Rf vendor

.PHONY: clean
clean:
	-$(GO_CMD) clean
	-rm -Rf /tmp/enroll_user /tmp/msp /tmp/keyvaluestore /tmp/hfc-kvs /tmp/state
	-rm -f integration-report.xml report.xml
	-FIXTURE_PROJECT_NAME=$(FIXTURE_PROJECT_NAME) DOCKER_REMOVE_FORCE=$(FIXTURE_DOCKER_REMOVE_FORCE) test/scripts/clean_integration.sh