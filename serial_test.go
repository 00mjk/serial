package goembserial

import (
	"errors"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	paramPORT     = 1
	paramBAUD     = paramPORT + 1
	paramLOOPBACK = paramBAUD + 1
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

	if order == paramPORT {
		return nil
	}

	val, err := strconv.Atoi(os.Getenv("TEST_BAUD"))
	if err != nil || val == 0 {
		return skipTestOnError(t, "Baud Rate was not provided")
	}

	baudrate = val

	if order == paramBAUD {
		return nil
	}

	if order == paramLOOPBACK {
		if os.Getenv("TEST_LOOPBACK") != "YES" {
			return skipTestOnError(t, "Not in Loop Back configuration")
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
	t.Logf("Error - %v", err)
}

func TestSerialConfig_N02(t *testing.T) {
	var c *SerialConfig
	c = &SerialConfig{Name: "testport"}
	handle, err := OpenPort(c)
	assert.Nil(t, handle)
	assert.Error(t, err)
	t.Logf("Error - %v", err)
}

func TestSerialConfig_N03(t *testing.T) {
	s := stopBitStr(StopBits2 + 1)
	assert.Contains(t, s, "Unknown")
}

func TestSerialConfig_N04(t *testing.T) {
	s := parityStr(ParitySpace + 1)
	assert.Contains(t, s, "Unknown")
}

func TestSerialConfig_N05(t *testing.T) {
	var c *SerialConfig

	verifySetup(t, paramPORT)

	c = &SerialConfig{Name: sport}
	handle, err := OpenPort(c)
	assert.Error(t, err)
	assert.Nil(t, handle)
	t.Logf("Error - %v", err)
	if err == nil {
		handle.Close()
	}
}

func TestSerialIntegration_N01(t *testing.T) {

	verifySetup(t, paramBAUD)

	c := &SerialConfig{
		Name:     sport,
		Baud:     baudrate,
		Parity:   ParityNone,
		StopBits: StopBits15,
		Flow:     FlowNone,
	}

	handle, err := OpenPort(c)
	//t.Log(err)
	assert.Error(t, err)
	assert.Nil(t, handle)
	t.Logf("Error - %v", err)
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
	t.Logf("Error - %v", err)
}

func Test_serialPort_N02(t *testing.T) {
	handle := serialPort{}
	buf := make([]byte, 100)
	n, err := handle.Write(buf)
	assert.Equal(t, 0, n)
	assert.Error(t, err)
	t.Logf("Error - %v", err)
}

func Test_serialPort_N03(t *testing.T) {
	handle := serialPort{}
	err := handle.Close()
	assert.Error(t, err)
	t.Logf("Error - %v", err)
}

func Test_serialPort_N04(t *testing.T) {
	handle := serialPort{}
	err := handle.Rts(true)
	assert.Error(t, err)
	t.Logf("Error - %v", err)
}

func Test_serialPort_N05(t *testing.T) {
	handle := serialPort{}
	err := handle.Dtr(true)
	assert.Error(t, err)
	t.Logf("Error - %v", err)
}

func Test_serialPort_N06(t *testing.T) {
	handle := serialPort{}
	val, err := handle.Cts()
	assert.Equal(t, val, false)
	assert.Error(t, err)
	t.Logf("Error - %v", err)
}

func Test_serialPort_N07(t *testing.T) {
	handle := serialPort{}
	val, err := handle.Dsr()
	assert.Equal(t, val, false)
	assert.Error(t, err)
	t.Logf("Error - %v", err)
}

func Test_serialPort_N08(t *testing.T) {
	handle := serialPort{}
	val, err := handle.Ring()
	assert.Equal(t, val, false)
	assert.Error(t, err)
	t.Logf("Error - %v", err)
}

func Test_serialPort_N09(t *testing.T) {
	handle := serialPort{}
	err := handle.SignalInvert(true)
	assert.Error(t, err)
	t.Logf("Error - %v", err)
}

func Test_serialPort_N10(t *testing.T) {
	handle := serialPort{}
	err := handle.SetBaud(9600)
	assert.Error(t, err)
	t.Logf("Error - %v", err)
}

func Test_serialPort_N11(t *testing.T) {
	handle := serialPort{}
	err := handle.SendBreak(true)
	assert.Error(t, err)
	t.Logf("Error - %v", err)
}

/*
Positive Tests
*/

func TestSerialConfig_P01(t *testing.T) {

	verifySetup(t, paramBAUD)

	handle, _ := createPort(t, ParityNone, StopBits1, FlowNone)
	closePort(t, handle)
}

func TestSerialIntegration_P01(t *testing.T) {

	verifySetup(t, paramLOOPBACK)

	handle, err := createPort(t, ParityNone, StopBits1, FlowNone)
	defer closePort(t, handle)

	buf := []byte("1 2 3 4 5 6 7 8 9 10")
	writePort(t, handle, buf)
	time.Sleep(500 * time.Millisecond)
	rbuf := make([]byte, len(buf))
	n, err := handle.Read(rbuf)
	assert.Equal(t, len(buf), n)
	assert.NoError(t, err)
	assert.Equal(t, buf, rbuf)
}

func TestSerialIntegration_P02(t *testing.T) {

	verifySetup(t, paramLOOPBACK)

	handle, err := createPort(t, ParityNone, StopBits1, FlowNone)
	defer closePort(t, handle)

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
}

func TestSerialIntegration_P03(t *testing.T) {

	verifySetup(t, paramLOOPBACK)

	handle, err := createPort(t, ParityNone, StopBits1, FlowNone)
	defer closePort(t, handle)

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
}

func TestSerialIntegration_P04(t *testing.T) {

	verifySetup(t, paramLOOPBACK)

	handle, err := createPort(t, ParityOdd, StopBits1, FlowNone)
	defer closePort(t, handle)

	buf := []byte("1 2 3 4 5 6 7 8 9 10")
	writePort(t, handle, buf)
	time.Sleep(500 * time.Millisecond)
	rbuf := make([]byte, len(buf))
	n, err := handle.Read(rbuf)
	assert.Equal(t, len(buf), n)
	assert.NoError(t, err)
	assert.Equal(t, buf, rbuf)
}

func TestSerialIntegration_P05(t *testing.T) {

	verifySetup(t, paramLOOPBACK)

	handle, err := createPort(t, ParityEven, StopBits1, FlowNone)
	defer closePort(t, handle)

	buf := []byte("1 2 3 4 5 6 7 8 9 10")
	writePort(t, handle, buf)
	time.Sleep(500 * time.Millisecond)
	rbuf := make([]byte, len(buf))
	n, err := handle.Read(rbuf)
	assert.Equal(t, len(buf), n)
	assert.NoError(t, err)
	assert.Equal(t, buf, rbuf)
}

func TestSerialIntegration_P06(t *testing.T) {

	verifySetup(t, paramLOOPBACK)

	handle, err := createPort(t, ParitySpace, StopBits1, FlowNone)
	defer closePort(t, handle)

	buf := []byte("1 2 3 4 5 6 7 8 9 10")
	writePort(t, handle, buf)
	time.Sleep(500 * time.Millisecond)
	rbuf := make([]byte, len(buf))
	n, err := handle.Read(rbuf)
	assert.Equal(t, len(buf), n)
	assert.NoError(t, err)
	assert.Equal(t, buf, rbuf)
}

func TestSerialIntegration_P07(t *testing.T) {

	verifySetup(t, paramLOOPBACK)

	handle, err := createPort(t, ParityMark, StopBits1, FlowNone)
	defer closePort(t, handle)

	buf := []byte("1 2 3 4 5 6 7 8 9 10")
	writePort(t, handle, buf)
	time.Sleep(500 * time.Millisecond)
	rbuf := make([]byte, len(buf))
	n, err := handle.Read(rbuf)
	assert.Equal(t, len(buf), n)
	assert.NoError(t, err)
	assert.Equal(t, buf, rbuf)
}

func TestSerialIntegration_P08(t *testing.T) {

	verifySetup(t, paramLOOPBACK)

	handle, err := createPort(t, ParityNone, StopBits2, FlowNone)
	defer closePort(t, handle)

	buf := []byte("1 2 3 4 5 6 7 8 9 10")
	writePort(t, handle, buf)
	time.Sleep(500 * time.Millisecond)
	rbuf := make([]byte, len(buf))
	n, err := handle.Read(rbuf)
	assert.Equal(t, len(buf), n)
	assert.NoError(t, err)
	assert.Equal(t, buf, rbuf)
}

func TestSerialIntegration_P09(t *testing.T) {

	verifySetup(t, paramLOOPBACK)

	handle, err := createPort(t, ParityNone, StopBits1, FlowHardware)
	defer closePort(t, handle)

	buf := []byte("1 2 3 4 5 6 7 8 9 10")
	// handle.Rts(false) - Since Hardware Flow Control is Enabled
	writePort(t, handle, buf)
	time.Sleep(500 * time.Millisecond)
	rbuf := make([]byte, len(buf))
	n, err := handle.Read(rbuf)
	assert.Equal(t, len(buf), n)
	assert.NoError(t, err)
	assert.Equal(t, buf, rbuf)
}

func TestSerialIntegration_P10(t *testing.T) {

	verifySetup(t, paramLOOPBACK)

	handle, err := createPort(t, ParityNone, StopBits1, FlowSoft)
	defer closePort(t, handle)

	buf := []byte("1 2 3 4 5 6 7 8 9 10")
	handle.Rts(false)
	writePort(t, handle, buf)
	time.Sleep(500 * time.Millisecond)
	rbuf := make([]byte, len(buf))
	n, err := handle.Read(rbuf)
	assert.Equal(t, len(buf), n)
	assert.NoError(t, err)
	assert.Equal(t, buf, rbuf)
}

func TestSerialIntegration_P11(t *testing.T) {

	verifySetup(t, paramLOOPBACK)

	handle, err := createPort(t, ParityNone, StopBits1, FlowNone)
	defer closePort(t, handle)

	buf := []byte("1 2 3 4 5 6 7 8 9 10")
	writePort(t, handle, buf)
	time.Sleep(500 * time.Millisecond)
	rbuf := make([]byte, len(buf))
	n, err := handle.Read(rbuf)
	assert.Equal(t, len(buf), n)
	assert.NoError(t, err)
	assert.Equal(t, buf, rbuf)

	err = handle.SetBaud(115200)
	assert.NoError(t, err)
	writePort(t, handle, buf)
	time.Sleep(500 * time.Millisecond)
	n, err = handle.Read(rbuf)
	assert.Equal(t, len(buf), n)
	assert.NoError(t, err)
	assert.Equal(t, buf, rbuf)
}

func TestSerialIntegration_P12(t *testing.T) {

	verifySetup(t, paramLOOPBACK)

	handle, err := createPort(t, ParityNone, StopBits1, FlowNone)
	defer closePort(t, handle)

	err = handle.Rts(false)
	assert.NoError(t, err)

	// By Default its high
	val, err := handle.Ring()
	assert.NoError(t, err)
	assert.Equal(t, false, val)

	err = handle.Rts(true)
	assert.NoError(t, err)

	time.Sleep(100 * time.Millisecond) // Minimum Wait Time

	// Signal Low
	val, err = handle.Ring()
	assert.NoError(t, err)
	assert.Equal(t, true, val)

	err = handle.Rts(false)
	assert.NoError(t, err)

	time.Sleep(100 * time.Millisecond) // Minimum Wait Time

	val, err = handle.Ring()
	assert.NoError(t, err)
	assert.Equal(t, false, val)
}

func TestSerialIntegration_P13(t *testing.T) {

	verifySetup(t, paramLOOPBACK)

	handle, err := createPort(t, ParityNone, StopBits1, FlowNone)
	defer closePort(t, handle)

	err = handle.Dtr(false)
	assert.NoError(t, err)

	// By Default its high
	val, err := handle.Dsr()
	assert.NoError(t, err)
	assert.Equal(t, false, val)

	err = handle.Dtr(true)
	assert.NoError(t, err)

	time.Sleep(100 * time.Millisecond) // Minimum Wait Time

	// Signal Low
	val, err = handle.Dsr()
	assert.NoError(t, err)
	assert.Equal(t, true, val)

	err = handle.Dtr(false)
	assert.NoError(t, err)

	time.Sleep(100 * time.Millisecond) // Minimum Wait Time

	val, err = handle.Dsr()
	assert.NoError(t, err)
	assert.Equal(t, false, val)
}
