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

// Function to check environment configuration
func verifySetup(order int) error {

	sport = os.Getenv("TEST_PORT")
	if sport == "" {
		return errors.New("Port Name was not provided")
	}

	if order == PARAM_PORT {
		return nil
	}

	val, err := strconv.Atoi(os.Getenv("TEST_BAUD"))
	if err != nil || val == 0 {
		return errors.New("Baud Rate was not provided")
	}

	baudrate = val

	if order == PARAM_BAUD {
		return nil
	}

	if order == PARAM_LOOPBACK {
		notLoopback := errors.New("Not Loop Back configuration :" + os.Getenv("TEST_LOOPBACK"))
		if os.Getenv("TEST_LOOPBACK") != "YES" {
			return notLoopback
		}
	}

	return nil
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

func TestSerialConfig_N10(t *testing.T) {
	var c *SerialConfig

	err := verifySetup(PARAM_PORT)
	if err != nil {
		t.Skip(err)
	}

	c = &SerialConfig{Name: sport}
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

/*
Positive Tests
*/

func TestSerialConfig_P01(t *testing.T) {
	var c *SerialConfig

	err := verifySetup(PARAM_BAUD)
	if err != nil {
		t.Skip(err)
	}

	c = &SerialConfig{Name: sport, Baud: baudrate}
	handle, err := OpenPort(c)
	assert.NoError(t, err)
	assert.NotNil(t, handle)
	err = handle.Close()
	assert.NoError(t, err)
}

func TestSerialConfig_P02(t *testing.T) {
	var c *SerialConfig

	err := verifySetup(PARAM_LOOPBACK)
	if err != nil {
		t.Skip(err)
	}

	c = &SerialConfig{Name: sport, Baud: baudrate}
	handle, err := OpenPort(c)
	assert.NoError(t, err)
	assert.NotNil(t, handle)

	buf := []byte("1 2 3 4 5 6 7 8 9 10")
	n, err := handle.Write(buf)
	assert.Equal(t, len(buf), n)
	assert.NoError(t, err)
	time.Sleep(500 * time.Millisecond)
	rbuf := make([]byte, len(buf))
	n, err = handle.Read(rbuf)
	assert.Equal(t, len(buf), n)
	assert.NoError(t, err)
	assert.Equal(t, buf, rbuf)

	err = handle.Close()
	assert.NoError(t, err)
}
