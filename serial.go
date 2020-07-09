// Copyright (C) 2020 Abhijit Bose
// SPDX-License-Identifier: GPL-2.0-only

// Package goembserial or GoEmbSerial is Embedded focused serial port package that allows you to read, write
// and configure the serial port.
//
// This project draws inspiration from the github.com/tarm/serial package
// and github.com/johnlauer/goserial package
//
// Initially this project aims to provide API and compatibility for windows.
// As time progresses other architectures would be added.
//
// This library is Context based and performs read write asynchronously.
//
// By default this package uses 8 bits (byte) data format for exchange.
//
// Note: Baud rates are defined as OS specifics
//
// Currently Following Features are supported:
//
//  1. All types of BAUD rates
//  2. Flow Control - Hardware, Software (XON/XOFF)
//  3. RTS , DTR control
//  4. CTS , DSR, RING read back
//  5. Parity Control - Odd, Even, Mark, Space
//  6. Stop Bit Control - 1 bit and 2 bits
//  7. Hardware to Software Signal Inversion for all Signals RTS, CTS, DTR, DSR
//  8. Sending Break from TX line
//  X. ... More on the way ...
//
package goembserial

import (
	"fmt"
	"io"
	"strconv"
	"time"
)

// DataSize defines the unit data size in bits used for Serial communication
const DataSize byte = 8

// Specific Stop bits type
const (
	// StopBits1 defines a single Stop bit sent after every data unit block
	StopBits1 byte = iota
	// StopBits15 defines a 1 and 1/2 Stop bits sent after every data unit block
	StopBits15 byte = iota // 1.5 Stop Bits
	// StopBits2 defines a 2 Stop bits sent after every data unit block
	StopBits2 byte = iota
)

//
// Parity Constants
//
const (
	ParityNone  byte = iota
	ParityOdd   byte = iota
	ParityEven  byte = iota
	ParityMark  byte = iota
	ParitySpace byte = iota
)

// Synchronization Constants

const (
	// FlowNone for no flow control to be used for Serial port
	FlowNone byte = iota
	// FlowHardware for CTS / RTS base Hardware flow control to be used for Serial port
	FlowHardware byte = iota
	// FlowSoft for Software flow control to be used for Serial port
	FlowSoft byte = iota // XON / XOFF based - Not Supported
)

// SerialConfig store's the given Serial Port configuration
type SerialConfig struct {
	Name         string
	Baud         int
	ReadTimeout  time.Duration // Blocks the Read operation for a specified time
	Parity       byte
	StopBits     byte
	Flow         byte
	SignalInvert bool // Option to invert the RTS/CTS/DTR/DSR Read outs
}

// Default Errors

var (
	// ErrNotImplemented -
	ErrNotImplemented = fmt.Errorf("Not Implemented yet")
	// ErrPortNotInitialized -
	ErrPortNotInitialized = fmt.Errorf("Port not initialized or closed")
	// ErrNotOpen -
	ErrNotOpen = fmt.Errorf("Error Port Not Open")
	// ErrAlreadyOpen -
	ErrAlreadyOpen = fmt.Errorf("Error Port is Already Open")
	// ErrAccessDenied -
	ErrAccessDenied = fmt.Errorf("Access Denied")
)

// SerialInterface Type for Multi platform implementation of Serial port functionality
type SerialInterface interface {
	io.ReadWriteCloser
	Rts(en bool) (err error)
	Cts() (en bool, err error)
	Dtr(en bool) (err error)
	Dsr() (en bool, err error)
	Ring() (en bool, err error)
	SetBaud(baud int) (err error)
	SignalInvert(en bool) (err error)
	SendBreak(en bool) (err error)
}

// OpenPort is a Function to Create the Serial Port and return an Interface type enclosing the configuration
func OpenPort(cfg *SerialConfig) (SerialInterface, error) {
	return openPort(cfg)
}

// Internal function for Logging of Stop bits
func stopBitStr(s byte) string {
	if s == StopBits1 {
		return "1"
	} else if s == StopBits15 {
		return "1.5"
	} else if s == StopBits2 {
		return "2"
	} else {
		return "Unknown " + strconv.Itoa(int(s))
	}
}

// Internal function for Logging Parity bits
func parityStr(p byte) string {
	if p == ParityNone {
		return "N"
	} else if p == ParityEven {
		return "E"
	} else if p == ParityOdd {
		return "O"
	} else if p == ParityMark {
		return "MARK"
	} else if p == ParitySpace {
		return "SPACE"
	} else {
		return "Unknown " + strconv.Itoa(int(p))
	}
}
