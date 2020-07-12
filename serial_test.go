package goembserial

import (
	"bytes"
	"errors"
	"fmt"
	"math/rand"
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

func TestSerialIntegration_P14(t *testing.T) {

	verifySetup(t, paramLOOPBACK)

	handle, err := createPort(t, ParityNone, StopBits1, FlowNone)
	defer closePort(t, handle)

	// Flip Polarity
	handle.SignalInvert(true)

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

func TestSerialIntegration_P15(t *testing.T) {

	verifySetup(t, paramLOOPBACK)

	handle, err := createPort(t, ParityNone, StopBits1, FlowNone)
	defer closePort(t, handle)

	// Flip Polarity
	handle.SignalInvert(true)

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

func TestSerialIntegration_P16(t *testing.T) {

	verifySetup(t, paramLOOPBACK)

	handle, err := createPort(t, ParityNone, StopBits1, FlowNone)
	defer closePort(t, handle)

	// Write Some Data
	buf := []byte("1 2 3 4 5 6 7 8 9 10")
	writePort(t, handle, buf)
	time.Sleep(500 * time.Millisecond)

	// Set the Break Signal
	err = handle.SendBreak(true)
	assert.NoError(t, err)

	time.Sleep(100 * time.Millisecond) // Minimum Wait Time

	// Clear Break Signal
	err = handle.SendBreak(false)
	assert.NoError(t, err)

	time.Sleep(100 * time.Millisecond) // Minimum Wait Time

	// Perform a Read
	rbuf := make([]byte, len(buf)+10)
	n, err := handle.Read(rbuf)
	assert.NoError(t, err)
	assert.Equal(t, len(buf)+1, n)
	assert.Equal(t, byte(0), rbuf[len(buf)])
	t.Logf("Break Signal Read - Size : %d , Data : %q", n, string(rbuf[:n]))
}

func TestBaudrates(t *testing.T) {

	// loadConfig(t)
	verifySetup(t, paramLOOPBACK)

	// Test Set
	ts := []struct {
		baud     int
		testSize int
		hasErr   bool
		timeout  time.Duration
	}{
		{0, 0, true, 0},
		{baudrate, 1024, false, 1200 * time.Millisecond},
		{300, 100, false, 5000 * time.Millisecond},
		{600, 100, false, 3000 * time.Millisecond},
		{1200, 100, false, 2000 * time.Millisecond},
		{1800, 100, false, 2000 * time.Millisecond},
		{2400, 100, false, 2000 * time.Millisecond},
		{4800, 100, false, 2000 * time.Millisecond},
		{19200, 1024, false, 1500 * time.Millisecond},
		{38400, 1024, false, 1200 * time.Millisecond},
		{57600, 1024, false, 1200 * time.Millisecond},
		{115200, 1024, false, 1200 * time.Millisecond},
		{230400, 1024, false, 1200 * time.Millisecond},
		{460800, 1024, false, 1200 * time.Millisecond},
		{500000, 1024, false, 1200 * time.Millisecond},
		{576000, 1024, false, 1200 * time.Millisecond},
		{921600, 1024, false, 1200 * time.Millisecond},
		{1000000, 1024, false, 1200 * time.Millisecond},
		{1152000, 1024, false, 1200 * time.Millisecond},
		{1500000, 1024, false, 1200 * time.Millisecond},
		{2000000, 1024, false, 1200 * time.Millisecond},
		{2500000, 1024, false, 1200 * time.Millisecond},
		{3000000, 1024, false, 1200 * time.Millisecond},
		{3500000, 1024, false, 1200 * time.Millisecond},
		{4000000, 1024, false, 1200 * time.Millisecond},
		{4500000, 1024, true, 0},
	}

	// Run Through the Test
	for _, tc := range ts {
		name := fmt.Sprintf("Baud Rate %d", tc.baud)

		t.Run(name, func(t *testing.T) {

			// Open Port with Respective BaudRate
			ref, err := OpenPort(&SerialConfig{
				Name:     sport,
				Baud:     tc.baud,
				Flow:     FlowNone,
				Parity:   ParityNone,
				StopBits: StopBits1,
			})

			// if Error was not Expected
			if err != nil && !tc.hasErr {
				t.Errorf("Error in Opening Port - %v", err)
				t.Fail()
			}

			// if Error was Expected
			if err == nil && tc.hasErr {
				t.Errorf("No Error's Returned even though it was expected")
				t.Fail()
			}

			// Log the Error for Information Purpose
			if err != nil {
				t.Logf("Info [Err]- %v", err) // Log the Error for Info
			}

			// Only Run if the test has not failed
			if !t.Failed() && !tc.hasErr {

				// Seed the Random number Generator
				rand.Seed(int64(time.Now().Nanosecond()))
				// Size of the Buffer
				var arrSize = tc.testSize
				// Buffer for Transaction
				buf := make([]byte, arrSize)
				_, err = rand.Read(buf)
				if err != nil {
					t.Errorf("Error create Random Buffer - %v", err)
					t.Fail()
				}

				if !t.Failed() {
					// Perform the Write
					n, err := ref.Write(buf)

					if err != nil {
						t.Errorf("Error Write Buffer - %v", err)
						t.Fail()
					}

					if n != arrSize {
						t.Errorf("Expected %d got %d", arrSize, n)
						t.Fail()
					}
				}

				if !t.Failed() {
					t.Logf("Engage Sleep of %v", tc.timeout)
					time.Sleep(tc.timeout)
				}

				if !t.Failed() {
					// Create the receive buffer
					rbuf := make([]byte, arrSize)

					//Perform the actual Read
					n, err := ref.Read(rbuf)
					if err != nil {
						t.Errorf("Error Read Buffer - %v", err)
						t.Fail()
					}

					if n != arrSize {
						t.Errorf("Expected %d got %d", arrSize, n)
						t.Fail()
					}

					if !bytes.Equal(buf, rbuf) {
						t.Errorf("Expected Buffers to be Equal")
						t.Fail()
					}
				}

			}

			// Do A close in case if 'ref' is not Nil
			if ref != nil {
				err = ref.Close()
				if err != nil {
					t.Errorf("Error Failed to Close Port - %v", err)
					t.Fail()
				}
			}
		})
	}
}
