## Copyright 2021 Abhijit Bose. All rights reserved.
## SPDX-License-Identifier: Apache-2.0
## Use of this source code is governed by a Apache 2.0 license that can be found
## in the LICENSE file.

# Makefile for testing
# Windows Specifics
ifeq ($(OS),Windows_NT)
# Set the Port where DUT is attached
PORT=COM3
else
# Linux Specifics

# Set the Port where DUT is attached
PORT=/dev/ttyUSB0
endif

# The default baud-rate that would be used to test out the DUT
BAUD=9600
# If we are ready with the require loop back setup to run the Tests
LOOPBACK=YES

# Command
TEST_CMD=TEST_PORT=$(PORT) TEST_BAUD=$(BAUD) TEST_LOOPBACK=$(LOOPBACK) go test -race

# Get all Packages in current directory
PACKAGES = $(shell find ./ -type d -not -path '*/\.*')

## Tags

test: hwsetup
	$(TEST_CMD) -v .
	go clean -testcache

cover-count-start:
	echo "mode: count" > coverage-all.out
	go clean -testcache

cover-count: hwsetup
	$(TEST_CMD) -v -cover -coverprofile=coverage.out -covermode=count .
	tail -n +2 coverage.out >> coverage-all.out
	go clean -testcache

cover: hwsetup
	$(TEST_CMD)	-v -cover -coverprofile=coverage.out .
	go tool cover -html=coverage.out
	go clean -testcache

cover-all:
	go tool cover -html=coverage-all.out
	go clean -testcache

hwsetup:
	@echo .
	@echo . Hardware Test Section 1
	@echo .
	export TEST_PORT=$(PORT)
	export TEST_BAUD=$(BAUD)
	export TEST_LOOPBACK=$(LOOPBACK)
	@echo .
	@echo . SERIAL PORT SETUP
	@echo .
	@echo ". TX  <===> RX  Short"
	@echo ". RTS <=+=> CTS Short"
	@echo ".       |"
	@echo ".       +=> RI  Short (Connected to RTS for Testing)"
	@echo ". DTR <=+=> DSR Short"
	@echo ".       |"
	@echo ".       +=> DCD Short (Connected to DTR for Testing)"
	@echo .
	@echo . Configure this setup and Press Enter to continue
	@echo .
	@read
	@echo .

gen-windows:
    rm zsyscall_windows.go
	go run golang.org/x/sys/windows/mkwinsyscall -output zsyscall_windows.go syscall_windows.go

# Trick to Do a combined Coverage file
test-cover-html:
	echo "mode: count" > coverage-all.out
	$(foreach pkg,$(PACKAGES),\
		go test -coverprofile=coverage.out -covermode=count $(pkg);\
		tail -n +2 coverage.out >> coverage-all.out;)
	go tool cover -html=coverage-all.out
	go clean -testcache


.PHONY: test cover-count-start cover cover-count cover-all test-cover-html gen-windows
