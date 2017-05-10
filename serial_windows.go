// +build windows

package goembserial

import (
	"context"
	"log"
	"os"
	"strconv"
	"sync"
	"syscall"
	"unsafe"
)

type serialPort struct {
	c          SerialConfig
	f          *os.File
	h          syscall.Handle
	ctx        context.Context
	cancelfunc context.CancelFunc
	//done    chan<- bool // Used only for Indication that the last operation was completed
	ro      *syscall.Overlapped
	wo      *syscall.Overlapped
	isBreak bool // Indicate if the Break is enabled

	rl sync.Mutex // Need to Eleminate these
	wl sync.Mutex
}

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
	ECF_SendXoff uint32 = 1
	ECF_SendXon  uint32 = 2
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
		StopBits_1:   0,
		StopBits_1_5: 1,
		StopBits_2:   2,
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
	modemStatusMask_RLSD_ON = 0x0080
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
Library Access
*/

// Platform Specific Open Port Function
func openPort(cfg *SerialConfig) (SerialInterface, error) {

	if len(cfg.Name) > 0 && cfg.Name[0] != '\\' {
		cfg.Name = "\\\\.\\" + cfg.Name
	}

	// Create the Handle
	h, err := syscall.CreateFile(syscall.StringToUTF16Ptr(cfg.Name),
		syscall.GENERIC_READ|syscall.GENERIC_WRITE,
		0,
		nil,
		syscall.OPEN_EXISTING,
		syscall.FILE_ATTRIBUTE_NORMAL|syscall.FILE_FLAG_OVERLAPPED,
		0)
	if err != nil {
		return nil, err
	}

	// Actually Open Stream
	f := os.NewFile(uintptr(h), cfg.Name)
	defer func() { // On Error Closure
		if err != nil {
			f.Close()
		}
	}()

	if err = wSetCommState(h, cfg.Baud, cfg.StopBits, cfg.Parity, cfg.Flow); err != nil {
		return nil, err
	}

	if err = wSetupComm(h, rxBufferSize, txBufferSize); err != nil {
		return nil, err
	}
	if err = wSetCommTimeouts(h); err != nil {
		return nil, err
	}
	if err = wSetCommMask(h); err != nil {
		return nil, err
	}

	ro, err := wNewOverlapped()
	if err != nil {
		return nil, err
	}
	wo, err := wNewOverlapped()
	if err != nil {
		return nil, err
	}

	sp := new(serialPort)
	sp.c = *cfg
	sp.f = f
	sp.h = h
	sp.ro = ro
	sp.wo = wo
	sp.ctx, sp.cancelfunc = context.WithCancel(context.Background())

	return sp, nil
}

/**
Interface Functions
*/

func (p *serialPort) Close() error {

	if p == nil || p.f == nil {
		return ErrPortNotInitialized
	}

	p.cancelfunc()
	return p.f.Close()
}

func (p *serialPort) Write(buf []byte) (int, error) {

	if p == nil || p.f == nil {
		return 0, ErrPortNotInitialized
	}

	p.wl.Lock()
	defer p.wl.Unlock()

	if err := wResetEvent(p.wo.HEvent); err != nil {
		return 0, err
	}
	var n uint32
	err := syscall.WriteFile(p.h, buf, &n, p.wo)
	if err != nil && err != syscall.ERROR_IO_PENDING {
		return int(n), err
	}
	return wGetOverlappedResult(p.h, p.wo)
}

func (p *serialPort) Read(buf []byte) (int, error) {

	if p == nil || p.f == nil {
		return 0, ErrPortNotInitialized
	}

	p.rl.Lock()
	defer p.rl.Unlock()

	if err := wResetEvent(p.ro.HEvent); err != nil {
		return 0, err
	}
	var done uint32
	err := syscall.ReadFile(p.h, buf, &done, p.ro)
	if err != nil && err != syscall.ERROR_IO_PENDING {
		return int(done), err
	}
	return wGetOverlappedResult(p.h, p.ro)
}

func (p *serialPort) Rts(en bool) error {

	if p == nil || p.f == nil {
		return ErrPortNotInitialized
	}

	val := ECF_SetRts
	if !p.c.SignalInvert {
		if !en {
			val = ECF_ClrRts
		}
	} else {
		val = ECF_ClrRts
		if !en {
			val = ECF_SetRts
		}
	}

	return wEscapeCommFunction(p.h, val)
}

func (p *serialPort) Cts() (bool, error) {

	if p == nil || p.f == nil {
		return false, ErrPortNotInitialized
	}

	status, err := wGetCommModemStatus(p.h)
	if err != nil {
		return false, err
	}

	ret := ((status & modemStatusMask_CTS_ON) != 0)
	if p.c.SignalInvert {
		ret = !ret
	}
	return ret, nil
}

func (p *serialPort) Dtr(en bool) error {

	if p == nil || p.f == nil {
		return ErrPortNotInitialized
	}

	val := ECF_SetDtr
	if !p.c.SignalInvert {
		if !en {
			val = ECF_ClrDtr
		}
	} else {
		val = ECF_ClrDtr
		if !en {
			val = ECF_SetDtr
		}
	}

	return wEscapeCommFunction(p.h, val)
}

func (p *serialPort) Dsr() (bool, error) {

	if p == nil || p.f == nil {
		return false, ErrPortNotInitialized
	}

	status, err := wGetCommModemStatus(p.h)
	if err != nil {
		return false, err
	}

	ret := ((status & modemStatusMask_DSR_ON) != 0)
	if p.c.SignalInvert {
		ret = !ret
	}

	return ret, nil
}

func (p *serialPort) SetBaud(baud int) error {

	if p == nil || p.f == nil {
		return ErrPortNotInitialized
	}

	if err := wSetCommState(p.h, baud, p.c.StopBits, p.c.Parity, p.c.Flow); err != nil {
		return err
	}

	p.c.Baud = baud
	return nil
}

func (p *serialPort) SignalInvert(en bool) error {

	if p == nil || p.f == nil {
		return ErrPortNotInitialized
	}

	p.c.SignalInvert = true
	return nil
}

func (p *serialPort) SendBreak(en bool) error {

	if p == nil || p.f == nil {
		return ErrPortNotInitialized
	}

	val := ECF_SetBreak
	if !en {
		val = ECF_ClrBreak
	}

	return wEscapeCommFunction(p.h, val)
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
	params.flags[0] = 0x01  // fBinary  :1
	params.flags[0] |= 0x10 // fDtrControl  :2 DTR Flow Control Enabled and ON
	params.flags[1] = 0x10  // fRtsControl  :2 RTS Flow Control Enabled and ON

	if parity != ParityNone {
		params.flags[0] |= 0x02 //fParity  :1
	}

	if flow == FlowHardware {
		params.flags[0] |= 0x04 // fOutxCtsFlow  :1
		params.flags[1] = 0x30  // fRtsControl  :2 RTS Flow Control RTS_CONTROL_TOGGLE wrt buffer
	}

	// Currently Soft Flow not Supported
	if flow == FlowSoft {
		return ErrNotImplemented
	}

	log.Println("Byte val of commstat flags[0]:", strconv.FormatInt(int64(params.flags[0]), 2))
	log.Println("Byte val of commstat flags[1]:", strconv.FormatInt(int64(params.flags[1]), 2))

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
	var dwFunc uint32
	dwFunc = operation
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
