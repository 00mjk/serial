package goembserial

// TODO: Add better Logging
// TODO: Make possible to run parallel tests for Serial configuration & transactions

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"os"
	"strconv"
)

// Port used for Testing
var actualPort = os.Getenv("TEST_PORT")
var actualBaud = os.Getenv("TEST_BAUD")
var rts_dsr_short = os.Getenv("TEST_RTS_DSR")
var rts_cts_short = os.Getenv("TEST_RTS_CTS")
var dtr_cts_short = os.Getenv("TEST_DTR_CTS")
var dtr_dsr_short = os.Getenv("TEST_DTR_DSR")
var actualLoopBack = os.Getenv("TEST_LOOPBACK")
var actualBreak = os.Getenv("TEST_TX_CTS")

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

func TestSerialConfig_N10(t *testing.T) {
	var c *SerialConfig

	if actualPort == "" {
		t.Skip("Port Name was not provided")
	}

	c = &SerialConfig{Name: actualPort}
	handle, err := OpenPort(c)
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
	assert.Equal(t,0,n)
	assert.Error(t,err)
}

func Test_serialPort_N02(t *testing.T) {
	handle := serialPort{}
	buf := make([]byte, 100)
	n, err := handle.Write(buf)
	assert.Equal(t,0,n)
	assert.Error(t,err)
}

func Test_serialPort_N03(t *testing.T) {
	handle := serialPort{}
	err := handle.Close()
	assert.Error(t,err)
}

/*
Positive Tests
 */

func TestSerialConfig_P01(t *testing.T) {
	var c *SerialConfig

	if actualPort == "" {
		t.Skip("Port Name was not provided")
	}

	baud, err := strconv.Atoi(actualBaud)
	if err != nil {
		t.Skip("Baud Rate was not provided")
	}

	c = &SerialConfig{Name: actualPort, Baud: baud}
	handle, err := OpenPort(c)
	assert.NoError(t, err)
	assert.NotNil(t, handle)
	err = handle.Close()
	assert.NoError(t, err)
}
