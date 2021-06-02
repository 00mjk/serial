// Copyright 2021 Abhijit Bose. All rights reserved.
// SPDX-License-Identifier: Apache-2.0
// Use of this source code is governed by a Apache 2.0 license that can be found
// in the LICENSE file.

// +build windows

package serial

import (
	"fmt"
	"log"
	"strconv"
	"syscall"
	"unsafe"
)

type structDCB struct {
	DCBlength, BaudRate                            uint32
	flags                                          [4]byte
	wReserved, XonLim, XoffLim                     uint16
	ByteSize, Parity, StopBits                     byte
	XonChar, XoffChar, ErrorChar, EofChar, EvtChar byte
	wReserved1                                     uint16
}

/*
type _DCB struct {
  DWORD DCBlength
  DWORD BaudRate

  DWORD fBinary  :1           // Flag[0]:0
  DWORD fParity  :1           // Flag[0]:1
  DWORD fOutxCtsFlow  :1      // Flag[0]:2
  DWORD fOutxDsrFlow  :1      // Flag[0]:3
  DWORD fDtrControl  :2       // Flag[0]:4-5
  DWORD fDsrSensitivity  :1   // Flag[0]:6
  DWORD fTXContinueOnXoff  :1 // Flag[1]:7

  DWORD fOutX  :1             // Flag[1]:0
  DWORD fInX  :1              // Flag[1]:1
  DWORD fErrorChar  :1        // Flag[1]:2
  DWORD fNull  :1             // Flag[1]:3
  DWORD fRtsControl  :2       // Flag[1]:4-5 // 13 and 14th bit, so [12:13]
  DWORD fAbortOnError  :1     // Flag[1]:6

  DWORD fDummy2  :17

  WORD  wReserved
  WORD  XonLim
  WORD  XoffLim
  BYTE  ByteSize
  BYTE  Parity
  BYTE  StopBits
  char  XonChar
  char  XoffChar
  char  ErrorChar
  char  EofChar
  char  EvtChar
  WORD  wReserved1
}
*/

type structTimeouts struct {
	ReadIntervalTimeout         uint32
	ReadTotalTimeoutMultiplier  uint32
	ReadTotalTimeoutConstant    uint32
	WriteTotalTimeoutMultiplier uint32
	WriteTotalTimeoutConstant   uint32
}

/**
EscapeCommFunction Constants
*/
const (
	ECF_ClrBreak uint32 = 9
	ECF_ClrDtr   uint32 = 6
	ECF_ClrRts   uint32 = 4
	ECF_SetBreak uint32 = 8
	ECF_SetDtr   uint32 = 5
	ECF_SetRts   uint32 = 3
	// ECF_SendXoff uint32 = 1
	// ECF_SendXon uint32 = 2
)

/**
Parity Conversion map for Windows
*/
var (
	parityMap = map[byte]byte{
		ParityNone:  0,
		ParityOdd:   1,
		ParityEven:  2,
		ParityMark:  3,
		ParitySpace: 4,
	}
)

/**
Stop Bit Conversion map for Windows
*/
var (
	stopbitMap = map[byte]byte{
		StopBits1: 0,
		//		StopBits15: 1,
		StopBits2: 2,
	}
)

// Receive Buffer size for Windows
const rxBufferSize = 64

// Transmit Buffer size for Windows
const txBufferSize = 64

/**
Modem Status Masks for Windows
*/
const (
	modemStatusMask_CTS_ON  = 0x0010
	modemStatusMask_DSR_ON  = 0x0020
	modemStatusMask_RING_ON = 0x0040
	// modemStatusMask_RLSD_ON = 0x0080
)

// DLL Functions
var (
	nSetCommState,
	nSetCommTimeouts,
	nSetCommMask,
	nSetupComm,
	//nClearCommError,
	nEscapeCommFunction,
	nGetCommModemStatus,
	nGetOverlappedResult,
	nCreateEvent,
	nResetEvent uintptr
)

// DLL Loader
func getProcAddr(lib syscall.Handle, name string) uintptr {
	addr, err := syscall.GetProcAddress(lib, name)
	if err != nil {
		panic(name + " " + err.Error())
	}
	return addr
}

/**
Initialization
*/

// Init for Loading System DLL
func init() {
	k32, err := syscall.LoadLibrary("kernel32.dll")
	if err != nil {
		panic("LoadLibrary " + err.Error())
	}
	defer syscall.FreeLibrary(k32)

	nSetCommState = getProcAddr(k32, "SetCommState")
	nSetCommTimeouts = getProcAddr(k32, "SetCommTimeouts")
	nSetCommMask = getProcAddr(k32, "SetCommMask")
	nSetupComm = getProcAddr(k32, "SetupComm")
	//nClearCommError = getProcAddr(k32, "ClearCommError")
	nEscapeCommFunction = getProcAddr(k32, "EscapeCommFunction")
	nGetCommModemStatus = getProcAddr(k32, "GetCommModemStatus")
	nGetOverlappedResult = getProcAddr(k32, "GetOverlappedResult")
	nCreateEvent = getProcAddr(k32, "CreateEventW")
	nResetEvent = getProcAddr(k32, "ResetEvent")
}

/**
Windows Internal Function
*/

func wSetCommState(h syscall.Handle, baud int, stopbits byte, parity byte, flow byte) error {
	var params structDCB
	params.DCBlength = uint32(unsafe.Sizeof(params))

	/*
			DWORD fBinary  :1           // Flag[0]:0
		  DWORD fParity  :1           // Flag[0]:1
		  DWORD fOutxCtsFlow  :1      // Flag[0]:2
		  DWORD fOutxDsrFlow  :1      // Flag[0]:3
		  DWORD fDtrControl  :2       // Flag[0]:4-5
		  DWORD fDsrSensitivity  :1   // Flag[0]:6
		  DWORD fTXContinueOnXoff  :1 // Flag[1]:7

		  DWORD fOutX  :1             // Flag[1]:0
		  DWORD fInX  :1              // Flag[1]:1
		  DWORD fErrorChar  :1        // Flag[1]:2
		  DWORD fNull  :1             // Flag[1]:3
		  DWORD fRtsControl  :2       // Flag[1]:4-5 // 13 and 14th bit, so [12:13]
		  DWORD fAbortOnError  :1     // Flag[1]:6
	*/
	params.flags[0] = (1 << 0)    // fBinary  :1
	params.flags[0] |= (0x1 << 4) // fDtrControl  :2 DTR_CONTROL_ENABLE and ON
	params.flags[1] = (0x1 << 4)  // fRtsControl  :2 RTS_CONTROL_ENABLE and ON

	if parity != ParityNone {
		params.flags[0] |= (1 << 1) //fParity  :1
	}

	/*
			// Flow Control
			if ( flowControl == "hardware" ){
			      dcb.fOutX = false;
		        dcb.fInX = false;
		        dcb.fOutxCtsFlow = true;
		        dcb.fOutxDsrFlow = true;
		        dcb.fDsrSensitivity = true;
		        dcb.fRtsControl = RTS_CONTROL_HANDSHAKE;
		        dcb.fDtrControl = DTR_CONTROL_HANDSHAKE;
			}else if ( flowControl == "xon/xoff" ){
		        dcb.fOutX = true;
		        dcb.fInX = true;
		        dcb.fOutxCtsFlow = false;
		        dcb.fOutxDsrFlow = false;
		        dcb.fDsrSensitivity = false;
		        dcb.fRtsControl = RTS_CONTROL_DISABLE;
		        dcb.fDtrControl = DTR_CONTROL_DISABLE;
			}
	*/

	if flow == FlowHardware {

		// fOutxCtsFlow  :1 Enabled
		params.flags[0] = (params.flags[0] & (^byte(1 << 2))) | (1 << 2)

		// fOutxDsrFlow  :1 Enabled
		// params.flags[0] = (params.flags[0] & (^byte(1<<3))) | (1<<3)

		// fDtrControl  :2 DTR Flow Control DTR_CONTROL_HANDSHAKE
		// params.flags[0] = (params.flags[0] & (^byte(0x3<<4))) | (0x2<<4)

		// fRtsControl  :2 RTS Flow Control RTS_CONTROL_HANDSHAKE
		params.flags[1] = (0x2 << 4)
	} else if flow == FlowSoft {
		//fOutX  :1   Enabled // Flag[1]:0
		params.flags[1] |= (1 << 0)
		//fInX  :1    Enabled // Flag[1]:1
		params.flags[1] |= (1 << 1)
	}

	// Currently Soft Flow not Supported
	/*if flow == FlowSoft {
		return ErrNotImplemented
	}*/

	log.Println("Byte val of commstat flags[0]:", strconv.FormatInt(int64(params.flags[0]), 2))
	log.Println("Byte val of commstat flags[1]:", strconv.FormatInt(int64(params.flags[1]), 2))

	if baud == 0 || baud < 0 {
		return fmt.Errorf("error in baudrate %d", baud)
	}

	if stopbits == StopBits15 {
		return fmt.Errorf("stopbits %s not supported by many serial ports", stopBitStr(stopbits))
	}

	params.BaudRate = uint32(baud)
	params.Parity = parityMap[parity]
	params.StopBits = stopbitMap[stopbits]
	params.ByteSize = DataSize

	r, _, err := syscall.Syscall(nSetCommState, 2, uintptr(h), uintptr(unsafe.Pointer(&params)), 0)
	if r == 0 {
		return err
	}
	return nil
}

func wSetCommTimeouts(h syscall.Handle) error {
	var timeouts structTimeouts
	const MAXDWORD = 1<<32 - 1
	timeouts.ReadIntervalTimeout = MAXDWORD
	timeouts.ReadTotalTimeoutMultiplier = MAXDWORD
	timeouts.ReadTotalTimeoutConstant = MAXDWORD - 1

	/* From http://msdn.microsoft.com/en-us/library/aa363190(v=VS.85).aspx
		 For blocking I/O see below:
		 Remarks:
		 If an application sets ReadIntervalTimeout and
		 ReadTotalTimeoutMultiplier to MAXDWORD and sets
		 ReadTotalTimeoutConstant to a value greater than zero and
		 less than MAXDWORD, one of the following occurs when the
		 ReadFile function is called:
		 If there are any bytes in the input buffer, ReadFile returns
		       immediately with the bytes in the buffer.
		 If there are no bytes in the input buffer, ReadFile waits
	               until a byte arrives and then returns immediately.
		 If no bytes arrive within the time specified by
		       ReadTotalTimeoutConstant, ReadFile times out.
	*/

	r, _, err := syscall.Syscall(nSetCommTimeouts, 2, uintptr(h), uintptr(unsafe.Pointer(&timeouts)), 0)
	if r == 0 {
		return err
	}
	return nil
}

func wSetCommMask(h syscall.Handle) error {
	const EV_RXCHAR = 0x0001
	// Set for Overlapped Interrupt on Received Data
	r, _, err := syscall.Syscall(nSetCommMask, 2, uintptr(h), EV_RXCHAR, 0)
	if r == 0 {
		return err
	}
	return nil
}

func wSetupComm(h syscall.Handle, in, out int) error {
	r, _, err := syscall.Syscall(nSetupComm, 3, uintptr(h), uintptr(in), uintptr(out))
	if r == 0 {
		return err
	}
	return nil
}

func wEscapeCommFunction(h syscall.Handle, operation uint32) error {
	dwFunc := operation
	r, _, err := syscall.Syscall(nEscapeCommFunction, 2, uintptr(h), uintptr(dwFunc), 0)
	if r == 0 {
		return err
	}
	return nil
}

func wGetCommModemStatus(h syscall.Handle) (int, error) {
	var n int
	r, _, err := syscall.Syscall(nGetCommModemStatus, 2, uintptr(h), uintptr(unsafe.Pointer(&n)), 0)
	if r == 0 {
		return n, err
	}
	return n, nil
}

func wGetOverlappedResult(h syscall.Handle, overlapped *syscall.Overlapped) (int, error) {
	var n int
	r, _, err := syscall.Syscall6(nGetOverlappedResult, 4,
		uintptr(h),
		uintptr(unsafe.Pointer(overlapped)),
		uintptr(unsafe.Pointer(&n)), 1, 0, 0)
	if r == 0 {
		return n, err
	}

	return n, nil
}

func wNewOverlapped() (*syscall.Overlapped, error) {
	var overlapped syscall.Overlapped
	r, _, err := syscall.Syscall6(nCreateEvent, 4, 0, 1, 0, 0, 0, 0)
	if r == 0 {
		return nil, err
	}
	overlapped.HEvent = syscall.Handle(r)
	return &overlapped, nil
}

func wResetEvent(h syscall.Handle) error {
	r, _, err := syscall.Syscall(nResetEvent, 1, uintptr(h), 0, 0)
	if r == 0 {
		return err
	}
	return nil
}
