// Copyright (C) 2020 Abhijit Bose
// SPDX-License-Identifier: GPL-2.0-only

// +build linux

package goembserial

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"testing"
	"time"

	"golang.org/x/sys/unix"
)

// Name of the Configuration File
const configFile = "serial_linux_test.json"

// Configuration Type
type testConfig struct {
	PortName string `json:"port"`
	BaudRate int    `json:"baud"`
	LoopBack bool   `json:"loopBack"`
}

// Global for Configuration
var cfg testConfig

///
// Helper Functions
///

// Configuration Loader
func loadConfig(t *testing.T) {
	b, err := ioutil.ReadFile(configFile)
	if err != nil {
		t.Errorf("Unable to Load Configuration due to - %v", err)
		t.FailNow()
	}
	err = json.Unmarshal(b, &cfg)
	if err != nil {
		t.Errorf("Unable to Parse Configuration due to - %v", err)
		t.FailNow()
	}
}

///
// Test Bench
///

func TestExtOpenPortConfig(t *testing.T) {

	loadConfig(t)

	// Test Type
	tt := []struct {
		name   string
		args   *SerialConfig
		hasErr bool
		isNil  bool
	}{
		{
			name: "Error Baud Rate 0",
			args: &SerialConfig{
				Baud: 0,
			},
			hasErr: true, isNil: true,
		},
		{
			name: "Error Baud Rate of ESP8266 76800",
			args: &SerialConfig{
				Baud: 76800,
			},
			hasErr: true, isNil: true,
		},
		{
			name: "Error Baud Rate of Very High 4500000",
			args: &SerialConfig{
				Baud: 4500000,
			},
			hasErr: true, isNil: true,
		},
		{
			name: "Compatible Baud Rate of 115200",
			args: &SerialConfig{
				Name:     cfg.PortName,
				Baud:     115200,
				Flow:     FlowNone,
				Parity:   ParityNone,
				StopBits: StopBits1,
			},
			hasErr: false, isNil: false,
		},
		{
			name: "Error Invalid Parity",
			args: &SerialConfig{
				Name:   cfg.PortName,
				Baud:   cfg.BaudRate,
				Parity: ParitySpace + 1,
			},
			hasErr: true, isNil: true,
		},
		{
			name: "Error Invalid StopBit",
			args: &SerialConfig{
				Name:     cfg.PortName,
				Baud:     cfg.BaudRate,
				Parity:   ParityNone,
				StopBits: StopBits2 + 1,
			},
			hasErr: true, isNil: true,
		},
		{
			name: "Error Not Supported StopBit 1.5",
			args: &SerialConfig{
				Name:     cfg.PortName,
				Baud:     cfg.BaudRate,
				Parity:   ParityNone,
				StopBits: StopBits15,
			},
			hasErr: true, isNil: true,
		},
		{
			name: "Error Invalid Flow Control",
			args: &SerialConfig{
				Name:     cfg.PortName,
				Baud:     cfg.BaudRate,
				Parity:   ParityNone,
				StopBits: StopBits1,
				Flow:     FlowSoft + 1,
			},
			hasErr: true, isNil: true,
		},
		{
			name: "Error Invalid Flow Control",
			args: &SerialConfig{
				Name:     cfg.PortName,
				Baud:     cfg.BaudRate,
				Parity:   ParityNone,
				StopBits: StopBits1,
				Flow:     FlowSoft + 1,
			},
			hasErr: true, isNil: true,
		},
		{
			name: "Error Port Name that is not Connected",
			args: &SerialConfig{
				Name:     "/dev/ttyUSB100",
				Baud:     cfg.BaudRate,
				Parity:   ParityNone,
				StopBits: StopBits1,
				Flow:     FlowNone,
			},
			hasErr: true, isNil: true,
		},
		{
			name: "Error Port Name that set Correctly",
			args: &SerialConfig{
				Name:     "/dev/ttyS0",
				Baud:     cfg.BaudRate,
				Parity:   ParityNone,
				StopBits: StopBits1,
				Flow:     FlowNone,
			},
			hasErr: true, isNil: true,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {

			t.Logf("Running with Config - %v", tc.args)

			// Flag
			testFailed := false

			// Perform the Call
			ref, err := OpenPort(tc.args)

			// If Error is not Expected
			if err != nil && !tc.hasErr {
				t.Errorf("Error in Opening Port - %v", err)
				testFailed = true
			}

			// if Error was Expected
			if err == nil && tc.hasErr {
				t.Errorf("No Error's Returned even though it was expected")
				testFailed = true
			}

			// Log the Error for Information Purpose
			if err != nil {
				t.Logf("Info [Err]- %v", err) // Log the Error for Info
			}

			// If Nil Reference is expected
			if ref != nil && tc.isNil {
				t.Errorf("Expected NIL got %v", ref)
				testFailed = true
			}

			// If Nil was not Expected
			if ref == nil && !tc.isNil {
				t.Errorf("Got NIL even though value was expected")
				testFailed = true
			}

			// Log the reference for Information Purpose
			if ref != nil {
				t.Logf("Info [Ref]- %v", ref) // Log the Error for Info
			}

			// Do A close in case if 'ref' is not Nil
			if ref != nil {
				err = ref.Close()
				if err != nil {
					t.Errorf("Error Failed to Close Port - %v", err)
					testFailed = true
				}
			}

			// If the Test Has Failed
			if testFailed {
				t.Fail()
			}

		})
	}
}

func TestIntOpenPortReOpen(t *testing.T) {

	loadConfig(t)

	// Check If an Open Port AutoCloses
	extRef, err := OpenPort(&SerialConfig{
		Name:     cfg.PortName,
		Baud:     cfg.BaudRate,
		Flow:     FlowNone,
		Parity:   ParityNone,
		StopBits: StopBits1,
	})
	if err != nil {
		t.Errorf("Error in Opening Port - %v", err)
		t.FailNow()
	}

	// Get Internal Reference
	intRef, ok := extRef.(*serialPort)
	if !ok {
		t.Errorf("Error Getting Internal Reference")
		err = extRef.Close()
		if err != nil {
			t.Errorf("Error in Closing Port - %v", err)
		}
		t.FailNow()
	}

	// Try Re-Opening the Same Port
	err = intRef.Open(cfg.PortName)
	if err != nil {
		t.Errorf("Error in Re-Opening Port - %v", err)
		err = extRef.Close()
		if err != nil {
			t.Errorf("Error in Closing Port - %v", err)
		}
		t.FailNow()
	}

	// Make Sure to Close
	err = extRef.Close()
	if err != nil {
		t.Errorf("Error in Closing Port - %v", err)
		t.FailNow()
	}
}

func TestOpenPortBlocked(t *testing.T) {

	loadConfig(t)

	// Externally Open the Port
	ext, err := unix.Open(cfg.PortName, unix.O_RDWR|unix.O_NOCTTY|unix.O_NONBLOCK, 0)
	if err != nil {
		t.Errorf("Error in Opening Port - %v", err)
		t.FailNow()
	}

	// Make Sure to Close Port After wards
	defer unix.Close(ext)

	// Attempt to Open Port Again
	ref, err := OpenPort(&SerialConfig{
		Name:     cfg.PortName,
		Baud:     cfg.BaudRate,
		Flow:     FlowNone,
		Parity:   ParityNone,
		StopBits: StopBits1,
	})
	if err == nil {
		t.Errorf("Error was Expected But did not get it")
		t.Fail()
	} else if ref != nil {
		t.Errorf("Reference should be NIL but got %v", ref)
		t.Fail()
	}
}

func TestSetBaudErrors(t *testing.T) {

	// Open the Stty Port
	ext, err := unix.Open("/dev/ttyS1", unix.O_RDWR|unix.O_NOCTTY|unix.O_NONBLOCK, 0)
	if err != nil {
		t.Errorf("Error in Opening Port - %v", err)
		t.FailNow()
	}

	// Create Dummy Instance with required Values
	s := &serialPort{}
	s.fd = ext
	s.opened = true

	// Try to Set the Baud Rate
	err = s.SetBaud(115200)
	t.Logf("Info - %v", err)
	if err == nil {
		t.Errorf("Expected Error but got NIL")
		t.Fail()
	}

	// Try to Set Wrong Baud Rate
	err = s.SetBaud(0)
	t.Logf("Info - %v", err)
	if err == nil {
		t.Errorf("Expected Error but got NIL")
		t.Fail()
	}

	err = unix.Close(ext)
	if err != nil {
		t.Errorf("Error in Closing Port - %v", err)
		t.FailNow()
	}
}

func TestTermiosErrors(t *testing.T) {

	// Create a Fake Structure
	s := &serialPort{}

	// Check Error on Port not opened
	err := s.SetTermios(unix.Termios{})
	t.Logf("Info Expected Error port Not open - %v", err)
	if err == nil {
		t.Errorf("Expected Error but got NIL")
		t.Fail()
	}

	_, err = s.GetTermios()
	t.Logf("Info Expected Error port Not open - %v", err)
	if err == nil {
		t.Errorf("Expected Error but got NIL")
		t.Fail()
	}

	err = s.SetModemSignal(0)
	t.Logf("Info Expected Error port Not open - %v", err)
	if err == nil {
		t.Errorf("Expected Error but got NIL")
		t.Fail()
	}

	_, err = s.GetModemSignals()
	t.Logf("Info Expected Error port Not open - %v", err)
	if err == nil {
		t.Errorf("Expected Error but got NIL")
		t.Fail()
	}
}

func TestTermiosErrorsWithOpenPort(t *testing.T) {

	// Open the Stty Port
	ext, err := unix.Open("/dev/ttyS1", unix.O_RDWR|unix.O_NOCTTY|unix.O_NONBLOCK, 0)
	if err != nil {
		t.Errorf("Error in Opening Port - %v", err)
		t.FailNow()
	}

	// Create Dummy Instance with required Values
	s := &serialPort{}
	s.fd = ext
	s.opened = true

	// Check Error
	err = s.SetTermios(unix.Termios{})
	t.Logf("Info Expected Error I/O - %v", err)
	if err == nil {
		t.Errorf("Expected Error but got NIL")
		t.Fail()
	}

	_, err = s.GetTermios()
	t.Logf("Info Expected Error I/O - %v", err)
	if err == nil {
		t.Errorf("Expected Error but got NIL")
		t.Fail()
	}

	err = s.SetModemSignal(0)
	t.Logf("Info Expected Error I/O - %v", err)
	if err == nil {
		t.Errorf("Expected Error but got NIL")
		t.Fail()
	}

	_, err = s.GetModemSignals()
	t.Logf("Info Expected Error I/O - %v", err)
	if err == nil {
		t.Errorf("Expected Error but got NIL")
		t.Fail()
	}

	err = unix.Close(ext)
	if err != nil {
		t.Errorf("Error in Closing Port - %v", err)
		t.FailNow()
	}
}

func TestReadWithoutWrite(t *testing.T) {

	loadConfig(t)

	// Open the Standard Port Config
	ref, err := OpenPort(&SerialConfig{
		Name:     cfg.PortName,
		Baud:     cfg.BaudRate,
		Flow:     FlowNone,
		Parity:   ParityNone,
		StopBits: StopBits1,
	})
	if err != nil {
		t.Errorf("Error in Opening Port - %v", err)
		t.FailNow()
	}

	// Attempt Read
	buf := make([]byte, 100)
	n, err := ref.Read(buf)
	t.Logf("Info Expected Error in Empty Queue - %v", err)
	if err == nil {
		t.Errorf("Expected Error but got NIL instead")
		t.Fail()
	}
	if n != 0 {
		t.Errorf("Expected Number of Bytes 0 but Got %v", n)
		t.Fail()
	}

	err = ref.Close()
	if err != nil {
		t.Errorf("Error in Closing Port - %v", err)
		t.FailNow()
	}

}

func TestBaudrates(t *testing.T) {

	loadConfig(t)

	// Test Set
	ts := []struct {
		baud     int
		testSize int
		hasErr   bool
		timeout  time.Duration
	}{
		{0, 0, true, 0},
		{cfg.BaudRate, 1024, false, 1200 * time.Millisecond},
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
				Name:     cfg.PortName,
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
