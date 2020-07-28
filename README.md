# Golang Embedded Serial port

Package `goembserial` or ***GoEmbSerial*** is **Embedded** focused serial port package.
It allows you to read, write and configure the serial port.

This project draws inspiration from the `github.com/tarm/serial` package
and `github.com/johnlauer/goserial` package

This project aims to provide API and compatibility for windows and Linux.
As time progresses other architectures would be added.

This library perform read write in Non-Blocking Manner.

By default this package uses 8 bits (byte) data format for exchange.
This is typical for **Embedded Applications** such as `UART` of an MCU.

Note: Baud rates are defined as OS specifics

Currently Following Features are supported:

 1. All types of BAUD rates
 2. Flow Control - Hardware, Software (XON/XOFF)
 3. RTS , DTR control
 4. CTS , DSR, RING read back
 5. Parity Control - Odd, Even, Mark, Space
 6. Stop Bit Control - 1 bit and 2 bits
 7. Hardware to Software Signal Inversion for all Signals RTS, CTS, DTR, DSR, RI
 8. Sending Break from TX line
 X. ... More on the way ...

## License

Copyright (c) 2020 Abhijit Bose < https://boseji.com >

All the files in this repository conform to 
[**GNU General Public License v2.0**](https://github.com/boseji/dotfiles/blob/master/LICENSE) unless otherwise specified.

SPDX-License-Identifier: GPL-2.0-only