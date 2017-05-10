/*

GoEmbSerial is Embedded focused serial port package that allows you to read, write
and configure the serial port.

This project draws inspiration from the github.com/tarm/serial package
and github.com/johnlauer/goserial package

Initially this project aims to provide API and compatibility for windows.
As time progresses other architectures would be added.

A special thing about this library is that its inherently Context based and
performs read write asynchronously as expected.

By default this package uses 8 bits (byte) data format for exchange.

Note: Baud rates are defined as OS specifics

*/
package goembserial

import (
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
	FlowSoft     byte = iota // XON / XOFF based
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
	SendBreak(t time.Duration) (err error)
	GetEvents(ev byte) (err error)
}
