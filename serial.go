/*

GoEmbSerial is Embedded focused serial port package that allows you to read, write
and configure the serial port.

This project draws inspiration from the github.com/tarm/serial package
and github.com/johnlauer/goserial package

Initially this project aims to provide API and compatibility for windows.
As time progresses other architectures would be added.

This library is Context based and performs read write asynchronously.

By default this package uses 8 bits (byte) data format for exchange.

Note: Baud rates are defined as OS specifics

*/
package goembserial

import (
	"errors"
	"io"
	"time"
)

/*
Data size is always 8-bits
*/
const DataSize byte = 8

/*
Specific Stop bits type
*/
const (
	StopBits_1   byte = 1
	StopBits_1_5 byte = 15 // 1.5 Stop Bits
	StopBits_2   byte = 2
)

/*
Parity Types
*/
const (
	ParityNone  byte = iota
	ParityOdd   byte = iota
	ParityEven  byte = iota
	ParityMark  byte = iota
	ParitySpace byte = iota
)

/*
Synchronization Type
*/
const (
	FlowNone     byte = iota
	FlowHardware byte = iota
	FlowSoft     byte = iota // XON / XOFF based - Not Supported
)

/*
Serial Port configuration Storage Type
*/
type SerialConfig struct {
	Name         string
	Baud         int
	ReadTimeout  time.Duration // Blocks the Read operation for a specified time
	Parity       byte
	StopBits     byte
	Flow         byte
	SignalInvert bool // Option to invert the RTS/CTS/DTR/DSR Read outs
}

/*
Default Error Types returned
 */

var ErrNotImplemented error = errors.New("Not Implemented yet")

var ErrPortNotInitialized error = errors.New("Port not initialized or closed")

/*
Serial Port Interface Type for Multi platform implementation
*/
type SerialInterface interface {
	io.ReadWriteCloser
	Rts(en bool) (err error)
	Cts() (en bool, err error)
	Dtr(en bool) (err error)
	Dsr() (en bool, err error)
	SetBaud(baud int) (err error)
	SignalInvert(en bool) (err error)
	SendBreak(en bool) (err error)
}

/*
Function to Create the Serial Port and return an Interface type enclosing the configuration
*/
func OpenPort(cfg *SerialConfig) (SerialInterface, error) {
	return openPort(cfg)
}
