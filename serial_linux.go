// Copyright (C) 2020 Abhijit Bose
// SPDX-License-Identifier: GPL-2.0-only

// +build linux

package goembserial

import (
	"fmt"
	"os/exec"
	"sync"
	"unsafe"

	"golang.org/x/sys/unix"
)

// Linux Compatible Serial Port Structure
type serialPort struct {
	// Handle
	fd int
	// Lock for Handle - Make it Thread Safe by Default
	mx sync.Mutex
	// If Port is Open
	opened bool
	// Invert Modem signals
	sigInv bool
	// Configuration
	conf SerialConfig
}

// Platform Specific Open Port Function
func openPort(cfg *SerialConfig) (SerialInterface, error) {
	s := &serialPort{}

	// Interpret the Config for Potential Errors
	t, err := getTermiosFor(cfg)
	if err != nil {
		return nil, err
	}

	// Open Port
	err = s.Open(cfg.Name)
	if err != nil {
		return nil, err
	}

	// Set Terminos
	err = s.SetTermios(t)
	if err != nil {
		return nil, err
	}

	// Set the Configuration
	s.conf = *cfg
	s.SignalInvert(cfg.SignalInvert) // No Errors Expected here

	// Finally Success
	return s, err
}

func (s *serialPort) Open(name string) error {
	// Establish Lock
	s.mx.Lock()
	defer s.mx.Unlock()

	// Check If its Open
	if s.opened {
		// Release Log temporarily
		s.mx.Unlock()
		// Ignore Errors for Forced Close
		s.Close()
		// Re-Engage Lock
		s.mx.Lock()
	}

	// Check if Port is already open
	err := exec.Command("lsof", "-t", name).Run()
	// This is ODD but yes if there is no error then we know port is open
	if err == nil {
		return ErrAlreadyOpen
	} else if err.Error() != "exit status 1" {
		return ErrAccessDenied
	}

	// Try to Open
	fd, err := unix.Open(
		name,
		unix.O_RDWR|unix.O_NOCTTY|unix.O_NONBLOCK|unix.O_EXCL,
		0,
	)
	if err != nil {
		return err
	}
	// Assign fd
	s.fd = fd
	s.opened = true

	// Auto Close on Errors
	defer func(fd int, err error) {
		if fd != 0 && err != nil {
			unix.Close(fd)
			s.fd = 0 // Not Initialized state
			s.opened = false
		}
	}(fd, err)

	// Get Exclusive Access
	if _, _, e1 := unix.Syscall(
		unix.SYS_IOCTL,
		uintptr(fd),
		uintptr(unix.TIOCEXCL),
		0,
	); e1 != 0 {
		return fmt.Errorf("Failed to Get Exclusive Access - %v", e1)
	}

	return err
}

func (s *serialPort) Read(p []byte) (n int, err error) {
	// Establish Lock
	s.mx.Lock()
	defer s.mx.Unlock()

	// Check If its Open
	if !s.opened {
		return 0, ErrNotOpen
	}

	// Loop to Access the Port Data
	for {
		// Perform the Actual Read
		n, err = unix.Read(s.fd, p)
		// In case the Read was interrupted by a Signal
		if err == unix.EINTR {
			continue
		}
		// In Case of Negative values of n due to other errors
		if n < 0 {
			n = 0 // Don't let -1 pass on
		}
		return n, err
	}
}

func (s *serialPort) Write(p []byte) (n int, err error) {
	// Establish Lock
	s.mx.Lock()
	defer s.mx.Unlock()

	// Check If its Open
	if !s.opened {
		return 0, ErrNotOpen
	}

	n, err = unix.Write(s.fd, p)
	// In case -1 returned - don't pass it on
	if n < 0 {
		n = 0
	}
	return n, err
}

func (s *serialPort) Close() error {
	// Establish Lock
	s.mx.Lock()
	defer s.mx.Unlock()

	// Check If its Open
	if !s.opened {
		return ErrPortNotInitialized
		// return nil
	}

	// Auto Run at the End of the function
	defer func() {
		s.fd = 0
		s.opened = false
	}()

	// Release Exclusive Access
	if _, _, e1 := unix.Syscall(
		unix.SYS_IOCTL,
		uintptr(s.fd),
		uintptr(unix.TIOCNXCL),
		0,
	); e1 != 0 {
		return fmt.Errorf("Failed to Release Exclusive Access - %v", e1)
	}

	// Perform the Actual Close
	return unix.Close(s.fd)
}

func (s *serialPort) Rts(en bool) (err error) {
	// Signal Inversion
	if s.sigInv {
		en = !en
	}

	// Get the Signals
	status, err := s.GetModemSignals()
	if err != nil {
		return err
	}

	// Apply the Values to Arg
	status &^= unix.TIOCM_RTS
	if en {
		status |= unix.TIOCM_RTS
	}

	// Set the Signal
	return s.SetModemSignal(status)
}

func (s *serialPort) Cts() (en bool, err error) {
	// Signal
	en = false
	// Modem Side
	status, err := s.GetModemSignals()
	if err != nil {
		return false, err
	}

	// Get Status
	if (status & unix.TIOCM_CTS) != 0 {
		en = true
	}

	// Signal Inversion
	if s.sigInv {
		en = !en
	}
	return en, nil
}

func (s *serialPort) Dtr(en bool) (err error) {
	// Signal Inversion
	if s.sigInv {
		en = !en
	}

	// Get the Signals
	status, err := s.GetModemSignals()
	if err != nil {
		return err
	}

	// Apply the Values to Arg
	status &^= unix.TIOCM_DTR
	if en {
		status |= unix.TIOCM_DTR
	}

	// Set the Signal
	return s.SetModemSignal(status)
}

func (s *serialPort) Dsr() (en bool, err error) {
	// Signal
	en = false
	// Modem Side
	status, err := s.GetModemSignals()
	if err != nil {
		return false, err
	}

	// Get Status
	if (status & unix.TIOCM_DSR) != 0 {
		en = true
	}

	// Signal Inversion
	if s.sigInv {
		en = !en
	}
	return en, nil
}

func (s *serialPort) Ring() (en bool, err error) {
	// Signal
	en = false
	// Modem Side
	status, err := s.GetModemSignals()
	if err != nil {
		return false, err
	}

	// Get Status
	if (status & unix.TIOCM_RI) != 0 {
		en = true
	}

	// Signal Inversion
	if s.sigInv {
		en = !en
	}
	return en, nil
}

func (s *serialPort) SetBaud(baud int) (err error) {
	// Already done in the GetTermios and SetTermios

	// Establish Lock
	// s.mx.Lock()
	// defer s.mx.Unlock()

	// Check If its Open
	// if !s.opened {
	// 	return ErrNotOpen
	// }

	// Default Baud rate
	baudSet, err := linuxFindBaud(baud)
	if err != nil {
		return err
	}

	// Process the Given Baud Rates

	// Termios
	var t unix.Termios
	// Get Values
	t, err = s.GetTermios()
	if err != nil {
		return err
	}

	// Set Baud rate
	t.Cflag &^= unix.CBAUD | unix.CBAUDEX
	t.Cflag |= uint32(baudSet)
	t.Ospeed = uint32(baudSet)
	t.Ispeed = uint32(baudSet)

	// Set Values
	err = s.SetTermios(t)
	if err != nil {
		return err
	}
	// Store the Baud
	s.conf.Baud = baud
	return nil
}

func (s *serialPort) SignalInvert(en bool) (err error) {
	// Check If its Open
	if !s.opened {
		return ErrNotOpen
	}
	s.sigInv = en
	return nil
}

func (s *serialPort) SendBreak(en bool) (err error) {
	// Establish Lock
	s.mx.Lock()
	defer s.mx.Unlock()

	// Check If its Open
	if !s.opened {
		return ErrNotOpen
	}

	// Argument for the Break Condition
	arg := unix.TIOCCBRK
	if en {
		arg = unix.TIOCSBRK
	}
	err = nil
	if _, _, e1 := unix.Syscall(
		unix.SYS_IOCTL,
		uintptr(s.fd),
		uintptr(arg),
		0,
	); e1 != 0 {
		err = e1
	}
	return err
}

func (s *serialPort) SetTermios(t unix.Termios) error {
	// Establish Lock
	s.mx.Lock()
	defer s.mx.Unlock()

	// Check If its Open
	if !s.opened {
		return ErrNotOpen
	}

	// Set Value
	if _, _, e1 := unix.Syscall6(
		unix.SYS_IOCTL,
		uintptr(s.fd),
		uintptr(unix.TCSETS),
		uintptr(unsafe.Pointer(&t)),
		0,
		0,
		0,
	); e1 != 0 {
		return error(e1)
	}
	return nil
}

func (s *serialPort) GetTermios() (t unix.Termios, err error) {
	// Establish Lock
	s.mx.Lock()
	defer s.mx.Unlock()

	// Check If its Open
	if !s.opened {
		return t, ErrNotOpen
	}

	// Set Value
	if _, _, e1 := unix.Syscall6(
		unix.SYS_IOCTL,
		uintptr(s.fd),
		uintptr(unix.TCGETS),
		uintptr(unsafe.Pointer(&t)),
		0,
		0,
		0,
	); e1 != 0 {
		return unix.Termios{}, error(e1)
	}
	return t, nil
}

func (s *serialPort) GetModemSignals() (status int, err error) {
	// Establish Lock
	s.mx.Lock()
	defer s.mx.Unlock()

	// Check If its Open
	if !s.opened {
		return 0, ErrNotOpen
	}

	err = nil
	// Read Signal
	if _, _, e1 := unix.Syscall(
		unix.SYS_IOCTL,
		uintptr(s.fd),
		uintptr(unix.TIOCMGET),
		uintptr(unsafe.Pointer(&status)),
	); e1 != 0 {
		err = error(e1)
	}
	return status, err
}

func (s *serialPort) SetModemSignal(status int) (err error) {
	// Establish Lock
	s.mx.Lock()
	defer s.mx.Unlock()

	// Check If its Open
	if !s.opened {
		return ErrNotOpen
	}

	// Set the Value
	if _, _, e1 := unix.Syscall(
		unix.SYS_IOCTL,
		uintptr(s.fd),
		uintptr(unix.TIOCMSET),
		uintptr(unsafe.Pointer(&status)),
	); e1 != 0 {
		return error(e1)
	}
	return nil
}

func linuxFindBaud(baud int) (int, error) {
	// if baud < 0 || baud == 0 {
	// 	return unix.B9600, nil
	// }
	baudRate := unix.B9600
	// Check Baud Rate
	switch baud {
	// case 0: // Default Baud rate 9600
	case 300:
		baudRate = unix.B300
	case 600:
		baudRate = unix.B600
	case 1200:
		baudRate = unix.B1200
	case 1800:
		baudRate = unix.B1800
	case 2400:
		baudRate = unix.B2400
	case 4800:
		baudRate = unix.B4800
	case 9600: // Default Baud rate 9600
	case 19200:
		baudRate = unix.B19200
	case 38400:
		baudRate = unix.B38400
	case 57600:
		baudRate = unix.B57600
	case 115200:
		baudRate = unix.B115200
	case 230400:
		baudRate = unix.B230400
	case 460800:
		baudRate = unix.B460800
	case 500000:
		baudRate = unix.B500000
	case 576000:
		baudRate = unix.B576000
	case 921600:
		baudRate = unix.B921600
	case 1000000:
		baudRate = unix.B1000000
	case 1152000:
		baudRate = unix.B1152000
	case 1500000:
		baudRate = unix.B1500000
	case 2000000:
		baudRate = unix.B2000000
	case 2500000:
		baudRate = unix.B2500000
	case 3000000:
		baudRate = unix.B3000000
	case 3500000:
		baudRate = unix.B3500000
	case 4000000:
		baudRate = unix.B4000000
	default:
		baudRate = 0 // TODO: Indicate we might Have Custom Baud Rate
		return 0, fmt.Errorf("Error Incorrect Baudrate or not supported")
	}
	return baudRate, nil
}

func getTermiosFor(cfg *SerialConfig) (unix.Termios, error) {
	var t unix.Termios
	// Set the Base RAW Mode - default 8 Bits
	t.Cflag = unix.CREAD | unix.CLOCAL | unix.CS8
	t.Iflag = unix.IGNPAR
	t.Cc[unix.VMIN] = 0
	t.Cc[unix.VTIME] = 0
	// Set Baud Rate
	baud, err := linuxFindBaud(cfg.Baud)
	if err != nil {
		return unix.Termios{}, err
	}
	t.Cflag |= uint32(baud)
	t.Ispeed = uint32(baud)
	t.Ospeed = uint32(baud)
	// Set Parity
	t.Cflag &^= unix.PARENB | unix.PARODD | unix.CMSPAR
	switch cfg.Parity {
	case ParityNone:
	case ParityEven:
		t.Cflag |= unix.PARENB
	case ParityOdd:
		t.Cflag |= unix.PARENB | unix.PARODD
	case ParitySpace:
		t.Cflag |= unix.PARENB | unix.CMSPAR
	case ParityMark:
		t.Cflag |= unix.PARENB | unix.PARODD | unix.CMSPAR
	default:
		return unix.Termios{}, fmt.Errorf("Invalid or not supported Parity")
	}
	// Set Stop Bits
	t.Cflag &^= unix.CSTOPB
	switch cfg.StopBits {
	case StopBits1:
	// case StopBits_1_5:
	// 	t.Cflag |= unix.CSTOPB // Do a 2 bit even for 1.5bit
	case StopBits2:
		t.Cflag |= unix.CSTOPB
	default:
		return unix.Termios{}, fmt.Errorf("Invalid or not supported Stop Bits")
	}
	// Set Flow Control
	t.Cflag &^= unix.CRTSCTS
	t.Iflag &^= unix.IXON | unix.IXOFF
	switch cfg.Flow {
	case FlowNone:
	case FlowSoft:
		t.Iflag |= unix.IXON | unix.IXOFF
	case FlowHardware:
		t.Cflag |= unix.CRTSCTS
	default:
		return unix.Termios{}, fmt.Errorf("Invalid or not supported Flow Control")
	}
	// Timeout Settings
	// Convert Time Out to Deci Seconds (1/10 of a Seconds)
	var deciSecTimeout int64 = 0
	// Minmum Number of Bytes
	var minBytes uint8 = 1
	// We have been supplied some timeout
	if cfg.ReadTimeout > 0 {
		// Get For Blocking on Timeout
		deciSecTimeout = cfg.ReadTimeout.Nanoseconds() / 1e8
		// No Need for Byte Blocking - hence EOF on Zero Read
		minBytes = 0
		if deciSecTimeout < 1 { // For Less than 100 mS
			// min possible timeout 1 Deciseconds (0.1s)
			deciSecTimeout = 1
		} else if deciSecTimeout > 255 {
			// max possible timeout is 255 deciseconds (25.5s)
			deciSecTimeout = 255
		}
	}
	// Set the Values
	t.Cc[unix.VMIN] = uint8(minBytes)
	t.Cc[unix.VTIME] = uint8(deciSecTimeout)
	// We are done
	return t, nil
}
