// +build windows

package goembserial

import (
	"context"
	"log"
	"os"
	"sync"
	"syscall"
)

// TODO: Add Custom Logging for each instance
type serialPort struct {
	conf         SerialConfig
	fileInstance *os.File
	hWnd         syscall.Handle
	ctx          context.Context
	cancelfunc   context.CancelFunc
	//done    chan<- bool // Used only for Indication that the last operation was completed
	ro      *syscall.Overlapped
	wo      *syscall.Overlapped
	isBreak bool // Indicate if the Break is enabled

	rl sync.Mutex // Need to Eleminate these
	wl sync.Mutex
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
	defer func() {
		// On Error Closure
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
	sp.conf = *cfg
	sp.fileInstance = f
	sp.hWnd = h
	sp.ro = ro
	sp.wo = wo
	sp.ctx, sp.cancelfunc = context.WithCancel(context.Background())
	log.Println("SerialPort Instance Created for ", sp.conf.Name,
		sp.conf.Baud, parityStr(sp.conf.Parity), stopBitStr(sp.conf.StopBits))
	return sp, nil
}

/**
Interface Functions
*/

func (p *serialPort) Close() error {

	if p == nil || p.fileInstance == nil {
		return ErrPortNotInitialized
	}

	p.cancelfunc()
	return p.fileInstance.Close()
}

func (p *serialPort) Write(buf []byte) (int, error) {

	if p == nil || p.fileInstance == nil {
		return 0, ErrPortNotInitialized
	}

	p.wl.Lock()
	defer p.wl.Unlock()

	if err := wResetEvent(p.wo.HEvent); err != nil {
		return 0, err
	}
	var n uint32
	err := syscall.WriteFile(p.hWnd, buf, &n, p.wo)
	if err != nil && err != syscall.ERROR_IO_PENDING {
		return int(n), err
	}
	return wGetOverlappedResult(p.hWnd, p.wo)
}

func (p *serialPort) Read(buf []byte) (int, error) {

	if p == nil || p.fileInstance == nil {
		return 0, ErrPortNotInitialized
	}

	p.rl.Lock()
	defer p.rl.Unlock()

	if err := wResetEvent(p.ro.HEvent); err != nil {
		return 0, err
	}
	var done uint32
	err := syscall.ReadFile(p.hWnd, buf, &done, p.ro)
	if err != nil && err != syscall.ERROR_IO_PENDING {
		return int(done), err
	}
	return wGetOverlappedResult(p.hWnd, p.ro)
}

func (p *serialPort) Rts(en bool) error {

	if p == nil || p.fileInstance == nil {
		return ErrPortNotInitialized
	}

	if p.conf.SignalInvert {
		en = !en
	}

	val := ECF_SetRts
	if en {
		val = ECF_SetRts
	} else {
		val = ECF_ClrRts
	}

	return wEscapeCommFunction(p.hWnd, val)
}

func (p *serialPort) Cts() (bool, error) {

	if p == nil || p.fileInstance == nil {
		return false, ErrPortNotInitialized
	}

	status, err := wGetCommModemStatus(p.hWnd)
	if err != nil {
		return false, err
	}

	ret := ((status & modemStatusMask_CTS_ON) != 0)
	if p.conf.SignalInvert {
		ret = !ret
	}
	return ret, nil
}

func (p *serialPort) Dtr(en bool) error {

	if p == nil || p.fileInstance == nil {
		return ErrPortNotInitialized
	}

	if p.conf.SignalInvert {
		en = !en
	}

	val := ECF_SetDtr
	if en {
		val = ECF_SetDtr
	} else {
		val = ECF_ClrDtr
	}

	return wEscapeCommFunction(p.hWnd, val)
}

func (p *serialPort) Dsr() (bool, error) {

	if p == nil || p.fileInstance == nil {
		return false, ErrPortNotInitialized
	}

	status, err := wGetCommModemStatus(p.hWnd)
	if err != nil {
		return false, err
	}

	ret := ((status & modemStatusMask_DSR_ON) != 0)
	if p.conf.SignalInvert {
		ret = !ret
	}

	return ret, nil
}

func (p *serialPort) SetBaud(baud int) error {

	if p == nil || p.fileInstance == nil {
		return ErrPortNotInitialized
	}

	if err := wSetCommState(p.hWnd, baud, p.conf.StopBits, p.conf.Parity, p.conf.Flow); err != nil {
		return err
	}

	p.conf.Baud = baud
	return nil
}

func (p *serialPort) SignalInvert(en bool) error {

	if p == nil || p.fileInstance == nil {
		return ErrPortNotInitialized
	}

	p.conf.SignalInvert = true
	return nil
}

func (p *serialPort) SendBreak(en bool) error {

	if p == nil || p.fileInstance == nil {
		return ErrPortNotInitialized
	}

	val := ECF_SetBreak
	if !en {
		val = ECF_ClrBreak
	}

	return wEscapeCommFunction(p.hWnd, val)
}
