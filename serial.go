// Copyright 2021 Abhijit Bose. All rights reserved.
// SPDX-License-Identifier: Apache-2.0
// Use of this source code is governed by a Apache 2.0 license that can be found
// in the LICENSE file.

// Package serial is Embedded focused serial port package that allows you to read, write
// and configure the serial port.
//
// This project draws inspiration from the github.com/tarm/serial package,
// github.com/johnlauer/goserial package and go.bug.st/serial package
//
// Initially this project aims to provide API and compatibility for Linux.
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
package serial

//go:generate go run golang.org/x/sys/windows/mkwinsyscall -output zsyscall_windows.go syscall_windows.go

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

// Config stores the complete configuration of a Serial Port
type Config struct {
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
	ErrNotImplemented = fmt.Errorf("not implemented yet")
	// ErrPortNotInitialized -
	ErrPortNotInitialized = fmt.Errorf("port not initialized or closed")
	// ErrNotOpen -
	ErrNotOpen = fmt.Errorf("port not open")
	// ErrAlreadyOpen -
	ErrAlreadyOpen = fmt.Errorf("port is already open")
	// ErrAccessDenied -
	ErrAccessDenied = fmt.Errorf("access denied")
)

// Port Type for Multi platform implementation of Serial port functionality
type Port interface {
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
func OpenPort(cfg *Config) (Port, error) {
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

// Internal function for Loggin Flow Control bits
func flowStr(f byte) string {
	if f == FlowNone {
		return "None"
	} else if f == FlowHardware {
		return "CTS/RTS"
	} else if f == FlowSoft {
		return "XON/XOFF"
	}
	return "Unknown " + strconv.Itoa(int(f))
}

// String is the implementation of the Stringer interface
func (c *Config) String() string {
	return fmt.Sprintf(
		"Port : %q Baud: %d Parity: %s StopBits: %s bits FlowControl: %s SignalInversion: %t",
		c.Name, c.Baud, parityStr(c.Parity), stopBitStr(c.StopBits),
		flowStr(c.Flow), c.SignalInvert,
	)
}
