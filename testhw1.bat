@set TEST_PORT=COM4
@set TEST_BAUD=9600
@set TEST_LOOPBACK=YES
@echo .
@echo . SERIAL PORT SETUP
@echo .
@echo ". TX  <=> RX  Short"
@echo ". CTS <=> RTS Short"
@echo ". DTR <=> DSR Short"
@echo .
@echo . Configure this setup and Press Enter to continue
@echo .
@pause
@go test -v -cover -coverprofile=coverage.out -covermode=count .
@tail -n +2 coverage.out >> coverage-all.out
@set TEST_PORT=
@set TEST_BAUD=
@set TEST_LOOPBACK=
