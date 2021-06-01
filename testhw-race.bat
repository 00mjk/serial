@rem Copyright 2021 Abhijit Bose. All rights reserved.
@rem SPDX-License-Identifier: Apache-2.0
@rem Use of this source code is governed by a Apache 2.0 license that can be found
@rem in the LICENSE file.
@set TEST_PORT=COM3
@set TEST_BAUD=9600
@set TEST_LOOPBACK=YES
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
@pause
@echo mode: count > coverage-all.out
@go test -race -v -cover -coverprofile=coverage.out -covermode=atomic .
@tail -n +2 coverage.out >> coverage-all.out
@set TEST_PORT=
@set TEST_BAUD=
@set TEST_LOOPBACK=
