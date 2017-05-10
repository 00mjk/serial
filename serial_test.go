package goembserial

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"runtime"
)

func TestSerialConfig_N01(t *testing.T) {
	c := &SerialConfig{}
	handle, err := OpenPort(c)
	assert.Nil(t, handle)
	assert.Error(t, err)
}

func TestSerialConfig_N02(t *testing.T) {
	var c *SerialConfig

	if runtime.GOOS == "windows" {
		c = &SerialConfig{Name:"COM100"}
	} else {
		t.Error("Can't use for ", runtime.GOOS)
	}
	handle, err := OpenPort(c)
	assert.Nil(t, handle)
	assert.Error(t, err)
}

func TestSerialConfig_N03(t *testing.T) {
	var c *SerialConfig

	if runtime.GOOS == "windows" {
		c = &SerialConfig{Name:"COM3"}
	} else {
		t.Error("Can't use for ", runtime.GOOS)
	}
	handle, err := OpenPort(c)
	assert.Error(t, err)
	assert.Nil(t, handle)
	if err == nil {
		handle.Close()
	}
}

func TestSerialConfig_P01(t *testing.T) {
	var c *SerialConfig

	if runtime.GOOS == "windows" {
		c = &SerialConfig{Name:"COM3" , Baud:1000}
	} else {
		t.Error("Can't use for ", runtime.GOOS)
	}
	handle, err := OpenPort(c)
	assert.NoError(t, err)
	assert.NotNil(t, handle)
	if err == nil {
		handle.Close()
	}
}