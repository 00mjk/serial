// Copyright 2021 Abhijit Bose. All rights reserved.
// SPDX-License-Identifier: Apache-2.0
// Use of this source code is governed by a Apache 2.0 license that can be found
// in the LICENSE file.

package RS485

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/boseji/serial"
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
func createPort(t *testing.T, parity, stopbits, flow byte) (serial.Port, error) {
	c := &serial.Config{
		Name:     sport,
		Baud:     baudrate,
		Parity:   parity,
		StopBits: stopbits,
		Flow:     flow,
	}
	handle, err := serial.OpenPort(c)
	assert.NoError(t, err)
	assert.NotNil(t, handle)
	return handle, err
}

func mockBadSignal(_ bool) error {
	return fmt.Errorf("this is mocked error for signalling of RS485")
}

func mockGoodSignal(_ bool) error {
	return nil
}

func TestNew(t *testing.T) {
	type args struct {
		port        serial.Port
		delayBefore time.Duration
		delayAfter  time.Duration
		sig         Control
	}
	tests := []struct {
		name    string
		args    args
		want    *Port
		wantErr bool
	}{
		{
			name: "Nil Port",
			args: args{
				port:        nil,
				delayBefore: 10 * time.Millisecond,
				delayAfter:  10 * time.Millisecond,
				sig:         mockGoodSignal,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Nil Signal function",
			args: args{
				port: func() serial.Port {
					verifySetup(t, paramBAUD)
					p, err := createPort(t, serial.ParityNone, serial.StopBits1, serial.FlowNone)
					assert.NoError(t, err)
					defer p.Close()
					return p
				}(),
				delayBefore: 10 * time.Millisecond,
				delayAfter:  10 * time.Millisecond,
				sig:         nil,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Bad Signal function",
			args: args{
				port: func() serial.Port {
					verifySetup(t, paramBAUD)
					p, err := createPort(t, serial.ParityNone, serial.StopBits1, serial.FlowNone)
					assert.NoError(t, err)
					defer p.Close()
					return p
				}(),
				delayBefore: 10 * time.Millisecond,
				delayAfter:  10 * time.Millisecond,
				sig:         mockBadSignal,
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.args.port, tt.args.delayBefore, tt.args.delayAfter, tt.args.sig)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}

	t.Run("Working", func(t *testing.T) {
		verifySetup(t, paramBAUD)
		p, err := createPort(t, serial.ParityNone, serial.StopBits1, serial.FlowNone)
		assert.NoError(t, err)
		defer p.Close()

		rs485, err := New(p, 10*time.Millisecond, 10*time.Millisecond, p.Rts)
		assert.NoError(t, err)
		defer rs485.Close()
	})
}

func TestPort_Write(t *testing.T) {
	t.Run("Positive Test 1", func(t *testing.T) {
		err := verifySetup(t, paramBAUD)
		assert.NoError(t, err)

		p, err := createPort(t, serial.ParityNone, serial.StopBits1, serial.FlowNone)
		if err != nil {
			t.Error("failed to open Port")
			return
		}
		defer p.Close()

		rs485, err := New(p, 10*time.Millisecond, 10*time.Millisecond, p.Rts)
		assert.NoError(t, err)
		defer rs485.Close()

		// Check the Setting of RTS Signal
		cts, err := p.Cts()
		assert.NoError(t, err)
		assert.EqualValues(t, cts, false)

		// Write to Port in Parallel
		var wg sync.WaitGroup
		wg.Add(1)
		go func(r *Port, t *testing.T) {
			defer wg.Done()
			_, err := rs485.Write([]byte("Hari Aum"))
			assert.NoError(t, err)
			t.Log("done!")
		}(rs485, t)

		// Wait for Small time
		time.Sleep(5 * time.Millisecond)

		// Check the Setting of RTS Signal
		cts, err = p.Cts()
		assert.NoError(t, err)
		assert.EqualValues(t, cts, true)

		// Wait for the Write to Complete
		wg.Wait()

		// Check if the RTS Signal is finally Low
		cts, err = p.Cts()
		assert.NoError(t, err)
		assert.EqualValues(t, cts, false)
	})

	t.Run("Bad Port1", func(t *testing.T) {
		var p *Port
		_, err := p.Write([]byte("Hari Aum"))
		assert.Error(t, err)
	})

	t.Run("Empty Data", func(t *testing.T) {
		err := verifySetup(t, paramBAUD)
		assert.NoError(t, err)

		p, err := createPort(t, serial.ParityNone, serial.StopBits1, serial.FlowNone)
		if err != nil {
			t.Error("failed to open Port")
			return
		}
		defer p.Close()

		rs485, err := New(p, 10*time.Millisecond, 10*time.Millisecond, p.Rts)
		assert.NoError(t, err)
		defer rs485.Close()

		_, err = rs485.Write([]byte{})
		assert.Error(t, err)
	})

	t.Run("Bad Signal", func(t *testing.T) {
		err := verifySetup(t, paramBAUD)
		assert.NoError(t, err)

		p, err := createPort(t, serial.ParityNone, serial.StopBits1, serial.FlowNone)
		if err != nil {
			t.Error("failed to open Port")
			return
		}
		defer p.Close()

		rs485, err := New(p, 10*time.Millisecond, 10*time.Millisecond, p.Rts)
		assert.NoError(t, err)
		defer rs485.Close()

		// Mock for Bad
		rs485.sig = mockBadSignal

		_, err = rs485.Write([]byte("Hari Aum"))
		assert.Error(t, err)
	})
}

func TestPort_Close(t *testing.T) {
	var p *Port
	err := p.Close()
	assert.Error(t, err)
}

func TestPort_Read(t *testing.T) {
	t.Run("Nil Port", func(t *testing.T) {
		var p *Port
		_, err := p.Read(make([]byte, 5))
		assert.Error(t, err)
	})
	t.Run("No Data", func(t *testing.T) {
		err := verifySetup(t, paramBAUD)
		assert.NoError(t, err)

		p, err := createPort(t, serial.ParityNone, serial.StopBits1, serial.FlowNone)
		if err != nil {
			t.Error("failed to open Port")
			return
		}
		defer p.Close()

		rs485, err := New(p, 10*time.Millisecond, 10*time.Millisecond, p.Rts)
		assert.NoError(t, err)
		defer rs485.Close()

		_, err = rs485.Read(nil)
		assert.Error(t, err)
	})
	t.Run("Bad Signal", func(t *testing.T) {
		err := verifySetup(t, paramBAUD)
		assert.NoError(t, err)

		p, err := createPort(t, serial.ParityNone, serial.StopBits1, serial.FlowNone)
		if err != nil {
			t.Error("failed to open Port")
			return
		}
		defer p.Close()

		rs485, err := New(p, 10*time.Millisecond, 10*time.Millisecond, p.Rts)
		assert.NoError(t, err)
		defer rs485.Close()

		rs485.sig = mockBadSignal

		buf := make([]byte, 2)
		_, err = rs485.Read(buf)
		assert.Error(t, err)
	})
	t.Run("Positive", func(t *testing.T) {
		err := verifySetup(t, paramBAUD)
		assert.NoError(t, err)

		p, err := createPort(t, serial.ParityNone, serial.StopBits1, serial.FlowNone)
		if err != nil {
			t.Error("failed to open Port")
			return
		}
		defer p.Close()

		rs485, err := New(p, 10*time.Millisecond, 10*time.Millisecond, p.Rts)
		assert.NoError(t, err)
		defer rs485.Close()

		message := []byte("Hari Aum")
		_, err = rs485.Write(message)
		assert.NoError(t, err)

		time.Sleep(5 * time.Millisecond)

		// Check if the RTS Signal is Low
		cts, err := p.Cts()
		assert.NoError(t, err)
		assert.EqualValues(t, cts, false)

		buf := make([]byte, len(message))
		n, err := rs485.Read(buf)
		assert.NoError(t, err)
		assert.Equal(t, len(message), n)
	})
}
