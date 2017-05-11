package goembserial

// TODO: Add better Logging
// TODO: Make possible to run parallel tests for Serial configuration & transactions

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"os"
	"strconv"
	"testing"
	"time"
)

const (
	PARAM_PORT     = 1
	PARAM_BAUD     = PARAM_PORT + 1
	PARAM_LOOPBACK = PARAM_BAUD + 1
)

// Serial Port String
var sport string

// Baud rate value
var baudrate int

// Skip on Error
func skipTestOnError(t *testing.T, msg string) error {
	t.Skip(msg)
	return errors.New(msg)
}

// Function to check environment configuration
func verifySetup(t *testing.T, order int) error {

	sport = os.Getenv("TEST_PORT")
	if sport == "" {
		return skipTestOnError(t, "Port Name was not provided")
	}

	if order == PARAM_PORT {
		return nil
	}

	val, err := strconv.Atoi(os.Getenv("TEST_BAUD"))
	if err != nil || val == 0 {
		return skipTestOnError(t, "Baud Rate was not provided")
	}

	baudrate = val

	if order == PARAM_BAUD {
		return nil
	}

	if order == PARAM_LOOPBACK {
		if os.Getenv("TEST_LOOPBACK") != "YES" {
			return skipTestOnError(t,
				"Not Loop Back configuration :"+os.Getenv("TEST_LOOPBACK"))
		}
	}

	return nil
}

// Create the Serial Port
func createPort(t *testing.T, parity, stopbits, flow byte) (SerialInterface, error) {
	c := &SerialConfig{
		Name:     sport,
		Baud:     baudrate,
		Parity:   parity,
		StopBits: stopbits,
		Flow:     flow,
	}
	handle, err := OpenPort(c)
	assert.NoError(t, err)
	assert.NotNil(t, handle)
	return handle, err
}

// Close the Serial Port
func closePort(t *testing.T, handle SerialInterface) error {
	err := handle.Close()
	assert.NoError(t, err)
	return err
}

// Write Data to Port
func writePort(t *testing.T, handle SerialInterface, data []byte) (int, error) {
	n, err := handle.Write(data)
	assert.Equal(t, len(data), n)
	assert.NoError(t, err)
	return n, err
}

/*
Negative Unit Tests
*/

func TestSerialConfig_N01(t *testing.T) {
	c := &SerialConfig{}
	handle, err := OpenPort(c)
	assert.Nil(t, handle)
	assert.Error(t, err)
}

func TestSerialConfig_N02(t *testing.T) {
	var c *SerialConfig
	c = &SerialConfig{Name: "testport"}
	handle, err := OpenPort(c)
	assert.Nil(t, handle)
	assert.Error(t, err)
}

func TestSerialConfig_N03(t *testing.T) {
	s := stopBitStr(StopBits_2 + 1)
	assert.Contains(t, s, "Unknown")
}

func TestSerialConfig_N04(t *testing.T) {
	s := parityStr(ParitySpace + 1)
	assert.Contains(t, s, "Unknown")
}

func TestSerialConfig_N10(t *testing.T) {
	var c *SerialConfig

	verifySetup(t, PARAM_PORT)

	c = &SerialConfig{Name: sport}
	handle, err := OpenPort(c)
	assert.Error(t, err)
	assert.Nil(t, handle)
	if err == nil {
		handle.Close()
	}
}

func TestSerialIntegation_N01(t *testing.T) {

	verifySetup(t, PARAM_BAUD)

	c := &SerialConfig{
		Name:     sport,
		Baud:     baudrate,
		Parity:   ParityNone,
		StopBits: StopBits_1_5,
		Flow:     FlowNone,
	}

	handle, err := OpenPort(c)
	//t.Log(err)
	assert.Error(t, err)
	assert.Nil(t, handle)
	if err == nil {
		handle.Close()
	}
}

/*
Internal Serial Port instance test - Negative
*/

func Test_serialPort_N01(t *testing.T) {
	handle := serialPort{}
	buf := make([]byte, 100)
	n, err := handle.Read(buf)
	assert.Equal(t, 0, n)
	assert.Error(t, err)
}

func Test_serialPort_N02(t *testing.T) {
	handle := serialPort{}
	buf := make([]byte, 100)
	n, err := handle.Write(buf)
	assert.Equal(t, 0, n)
	assert.Error(t, err)
}

func Test_serialPort_N03(t *testing.T) {
	handle := serialPort{}
	err := handle.Close()
	assert.Error(t, err)
}

func Test_serialPort_N04(t *testing.T) {
	handle := serialPort{}
	err := handle.Rts(true)
	assert.Error(t, err)
}

func Test_serialPort_N05(t *testing.T) {
	handle := serialPort{}
	err := handle.Dtr(true)
	assert.Error(t, err)
}

func Test_serialPort_N06(t *testing.T) {
	handle := serialPort{}
	val, err := handle.Cts()
	assert.Equal(t, val, false)
	assert.Error(t, err)
}

func Test_serialPort_N07(t *testing.T) {
	handle := serialPort{}
	val, err := handle.Dsr()
	assert.Equal(t, val, false)
	assert.Error(t, err)
}

func Test_serialPort_N08(t *testing.T) {
	handle := serialPort{}
	err := handle.SignalInvert(true)
	assert.Error(t, err)
}

func Test_serialPort_N09(t *testing.T) {
	handle := serialPort{}
	err := handle.SetBaud(9600)
	assert.Error(t, err)
}

func Test_serialPort_N10(t *testing.T) {
	handle := serialPort{}
	err := handle.SendBreak(true)
	assert.Error(t, err)
}

/*
Positive Tests
*/

func TestSerialConfig_P01(t *testing.T) {

	verifySetup(t, PARAM_BAUD)

	handle, _ := createPort(t, ParityNone, StopBits_1, FlowNone)
	closePort(t, handle)
}

func TestSerialIntegation_P01(t *testing.T) {

	verifySetup(t, PARAM_LOOPBACK)

	handle, err := createPort(t, ParityNone, StopBits_1, FlowNone)

	buf := []byte("1 2 3 4 5 6 7 8 9 10")
	writePort(t, handle, buf)
	time.Sleep(500 * time.Millisecond)
	rbuf := make([]byte, len(buf))
	n, err := handle.Read(rbuf)
	assert.Equal(t, len(buf), n)
	assert.NoError(t, err)
	assert.Equal(t, buf, rbuf)

	closePort(t, handle)
}

func TestSerialIntegation_P02(t *testing.T) {

	verifySetup(t, PARAM_LOOPBACK)

	handle, err := createPort(t, ParityNone, StopBits_1, FlowNone)

	// By default RTS is ON
	val, err := handle.Cts()
	assert.NoError(t, err)
	assert.Equal(t, true, val)

	err = handle.Rts(false)
	assert.NoError(t, err)

	time.Sleep(1 * time.Millisecond)

	val, err = handle.Cts()
	assert.NoError(t, err)
	assert.Equal(t, false, val)

	err = handle.SignalInvert(true)
	assert.NoError(t, err)

	time.Sleep(1 * time.Millisecond)

	val, err = handle.Cts()
	assert.NoError(t, err)
	assert.Equal(t, true, val)

	err = handle.Rts(true)
	assert.NoError(t, err)

	time.Sleep(1 * time.Millisecond)

	val, err = handle.Cts()
	assert.NoError(t, err)
	assert.Equal(t, true, val)

	err = handle.Rts(false)
	assert.NoError(t, err)

	time.Sleep(1 * time.Millisecond)

	val, err = handle.Cts()
	assert.NoError(t, err)
	assert.Equal(t, false, val)

	closePort(t, handle)
}

func TestSerialIntegation_P03(t *testing.T) {

	verifySetup(t, PARAM_LOOPBACK)

	handle, err := createPort(t, ParityNone, StopBits_1, FlowNone)

	// By default DTR is ON
	val, err := handle.Dsr()
	assert.NoError(t, err)
	assert.Equal(t, true, val)

	err = handle.Dtr(false)
	assert.NoError(t, err)

	time.Sleep(1 * time.Millisecond)

	val, err = handle.Dsr()
	assert.NoError(t, err)
	assert.Equal(t, false, val)

	err = handle.SignalInvert(true)
	assert.NoError(t, err)

	time.Sleep(1 * time.Millisecond)

	val, err = handle.Dsr()
	assert.NoError(t, err)
	assert.Equal(t, true, val)

	err = handle.Dtr(true)
	assert.NoError(t, err)

	time.Sleep(1 * time.Millisecond)

	val, err = handle.Dsr()
	assert.NoError(t, err)
	assert.Equal(t, true, val)

	err = handle.Dtr(false)
	assert.NoError(t, err)

	time.Sleep(1 * time.Millisecond)

	val, err = handle.Dsr()
	assert.NoError(t, err)
	assert.Equal(t, false, val)

	closePort(t, handle)
}

func TestSerialIntegation_P04(t *testing.T) {

	verifySetup(t, PARAM_LOOPBACK)

	handle, err := createPort(t, ParityOdd, StopBits_1, FlowNone)

	buf := []byte("1 2 3 4 5 6 7 8 9 10")
	writePort(t, handle, buf)
	time.Sleep(500 * time.Millisecond)
	rbuf := make([]byte, len(buf))
	n, err := handle.Read(rbuf)
	assert.Equal(t, len(buf), n)
	assert.NoError(t, err)
	assert.Equal(t, buf, rbuf)

	closePort(t, handle)
}

func TestSerialIntegation_P05(t *testing.T) {

	verifySetup(t, PARAM_LOOPBACK)

	handle, err := createPort(t, ParityEven, StopBits_1, FlowNone)

	buf := []byte("1 2 3 4 5 6 7 8 9 10")
	writePort(t, handle, buf)
	time.Sleep(500 * time.Millisecond)
	rbuf := make([]byte, len(buf))
	n, err := handle.Read(rbuf)
	assert.Equal(t, len(buf), n)
	assert.NoError(t, err)
	assert.Equal(t, buf, rbuf)

	closePort(t, handle)
}

func TestSerialIntegation_P06(t *testing.T) {

	verifySetup(t, PARAM_LOOPBACK)

	handle, err := createPort(t, ParitySpace, StopBits_1, FlowNone)

	buf := []byte("1 2 3 4 5 6 7 8 9 10")
	writePort(t, handle, buf)
	time.Sleep(500 * time.Millisecond)
	rbuf := make([]byte, len(buf))
	n, err := handle.Read(rbuf)
	assert.Equal(t, len(buf), n)
	assert.NoError(t, err)
	assert.Equal(t, buf, rbuf)

	closePort(t, handle)
}

func TestSerialIntegation_P07(t *testing.T) {

	verifySetup(t, PARAM_LOOPBACK)

	handle, err := createPort(t, ParityMark, StopBits_1, FlowNone)

	buf := []byte("1 2 3 4 5 6 7 8 9 10")
	writePort(t, handle, buf)
	time.Sleep(500 * time.Millisecond)
	rbuf := make([]byte, len(buf))
	n, err := handle.Read(rbuf)
	assert.Equal(t, len(buf), n)
	assert.NoError(t, err)
	assert.Equal(t, buf, rbuf)

	closePort(t, handle)
}

func TestSerialIntegation_P08(t *testing.T) {

	verifySetup(t, PARAM_LOOPBACK)

	handle, err := createPort(t, ParityNone, StopBits_2, FlowNone)

	buf := []byte("1 2 3 4 5 6 7 8 9 10")
	writePort(t, handle, buf)
	time.Sleep(500 * time.Millisecond)
	rbuf := make([]byte, len(buf))
	n, err := handle.Read(rbuf)
	assert.Equal(t, len(buf), n)
	assert.NoError(t, err)
	assert.Equal(t, buf, rbuf)

	closePort(t, handle)
}

func TestSerialIntegation_P09(t *testing.T) {

	verifySetup(t, PARAM_LOOPBACK)

	handle, err := createPort(t, ParityNone, StopBits_1, FlowHardware)

	buf := []byte("1 2 3 4 5 6 7 8 9 10")
	handle.Rts(false)
	writePort(t, handle, buf)
	time.Sleep(500 * time.Millisecond)
	rbuf := make([]byte, len(buf))
	n, err := handle.Read(rbuf)
	assert.Equal(t, len(buf), n)
	assert.NoError(t, err)
	assert.Equal(t, buf, rbuf)

	closePort(t, handle)
}

func TestSerialIntegation_P10(t *testing.T) {

	verifySetup(t, PARAM_LOOPBACK)

	handle, err := createPort(t, ParityNone, StopBits_1, FlowSoft)

	buf := []byte("1 2 3 4 5 6 7 8 9 10")
	handle.Rts(false)
	writePort(t, handle, buf)
	time.Sleep(500 * time.Millisecond)
	rbuf := make([]byte, len(buf))
	n, err := handle.Read(rbuf)
	assert.Equal(t, len(buf), n)
	assert.NoError(t, err)
	assert.Equal(t, buf, rbuf)

	closePort(t, handle)
}