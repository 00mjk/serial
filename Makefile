# Makefile for testing

# For Export of Environmental variables
ifeq ($(OS),Windows_NT)
	EXP=set
	BRK=@pause
else
	EXP=export
	BRK=@read
endif

PORT=COM3
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
	$(EXP) TEST_PORT=$(PORT)
	$(EXP) TEST_BAUD=$(BAUD)
	$(EXP) TEST_RTS_CTS=YES
	$(EXP) TEST_DTR_DSR=YES
	$(EXP) TEST_LOOPBACK=YES
	@echo .
	@echo . Configure this setup and Press Enter to continue
	$(BRK)
	@echo .

hardwaretest1: hardwaresetup1
	$(MAKE) test
	@echo .

hardwaretest1cover: hardwaresetup1
	$(MAKE) cover
	@echo .


# Trick to Do a combined Coverage file
test-cover-html:
	echo "mode: count" > coverage-all.out
	$(foreach pkg,$(PACKAGES),\
		go test -coverprofile=coverage.out -covermode=count $(pkg);\
		tail -n +2 coverage.out >> coverage-all.out;)
	go tool cover -html=coverage-all.out


.PHONY: test cover-count-start cover cover-count cover-all hardwaretest1 test-cover-html
