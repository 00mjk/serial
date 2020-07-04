@set TEST_PORT=COM3
@set TEST_BAUD=9600
@set TEST_LOOPBACK=YES
@echo .
@echo . SERIAL PORT SETUP
@echo .
@echo ". TX  <=> RX  Short"
@echo ". CTS <=> RTS Short"
@echo ". DTR <=> DSR Short"
@echo ". TX  <=> RI  Short for break test"
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
