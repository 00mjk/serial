# Makefile for testing

PORT=/dev/ttyUSB0
BAUD=9600

# Get all Packages in current directory
PACKAGES = $(shell find ./ -type d -not -path '*/\.*')

## Tags



test:
	go test -v .

cover-count-start:
	echo "mode: count" > coverage-all.out

cover-count:
	go test -v -cover -coverprofile=coverage.out -covermode=count .
	tail -n +2 coverage.out >> coverage-all.out

cover:
	go test -v -cover -coverprofile=coverage.out .
	go tool cover -html=coverage.out

cover-all:
	go tool cover -html=coverage-all.out

hardwaresetup1:
	@echo .
	@echo . Hardware Test Section 1
	@echo .
	export TEST_PORT=$(PORT)
	export TEST_BAUD=$(BAUD)
	export TEST_LOOPBACK=YES
	@echo .
	@echo . SERIAL PORT SETUP
	@echo .
	@echo ". TX  <=> RX  Short"
	@echo ". CTS <=> RTS Short"
	@echo ". DTR <=> DSR Short"
	@echo .
	@echo . Configure this setup and Press Enter to continue
	@echo .
	read
	@echo .

hardwaretest1: hardwaresetup1 test
	@echo .

hardwaretest1cover: hardwaresetup1 cover
	@echo .


# Trick to Do a combined Coverage file
test-cover-html:
	echo "mode: count" > coverage-all.out
	$(foreach pkg,$(PACKAGES),\
		go test -coverprofile=coverage.out -covermode=count $(pkg);\
		tail -n +2 coverage.out >> coverage-all.out;)
	go tool cover -html=coverage-all.out


.PHONY: test cover-count-start cover cover-count cover-all hardwaretest1 test-cover-html
