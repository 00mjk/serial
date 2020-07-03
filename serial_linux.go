// +build linux

package goembserial

import "fmt"

type serialPort struct{}

// Platform Specific Open Port Function
func openPort(cfg *SerialConfig) (SerialInterface, error) {
	return nil, fmt.Errorf("not implemented")
}

func (serialPort) Read(p []byte) (n int, err error) {
	return 0, fmt.Errorf("not implemented")
}

func (serialPort) Write(p []byte) (n int, err error) {
	return 0, fmt.Errorf("not implemented")
}

func (serialPort) Close() error {
	return fmt.Errorf("not implemented")
}

func (serialPort) Rts(en bool) (err error) {
	return fmt.Errorf("not implemented")
}

func (serialPort) Cts() (en bool, err error) {
	return false, fmt.Errorf("not implemented")
}

func (serialPort) Dtr(en bool) (err error) {
	return fmt.Errorf("not implemented")
}

func (serialPort) Dsr() (en bool, err error) {
	return false, fmt.Errorf("not implemented")
}

func (serialPort) Ring() (en bool, err error) {
	return false, fmt.Errorf("not implemented")
}

func (serialPort) SetBaud(baud int) (err error) {
	return fmt.Errorf("not implemented")
}

func (serialPort) SignalInvert(en bool) (err error) {
	return fmt.Errorf("not implemented")
}

func (serialPort) SendBreak(en bool) (err error) {
	return fmt.Errorf("not implemented")
}
