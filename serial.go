/*
Package serial is a library for I/O over serial devices.
*/
package serial

import (
	"fmt"

	"golang.org/x/sys/unix"
)

// A Port represents an open serial device.
type Port struct {
	fd int
}

// Open opens the specified serial device.
func Open(device string, speed int) (*Port, error) {
	const openFlags = unix.O_NONBLOCK | unix.O_NOCTTY | unix.O_RDWR
	fd, err := unix.Open(device, openFlags, 0)
	if err != nil {
		return nil, err
	}
	baudrate, err := getBaudRate(speed)
	if err != nil {
		return nil, err
	}
	err = unix.IoctlSetTermios(fd, unix.TCSETS, &unix.Termios{
		Iflag: unix.IGNPAR,
		Cflag: unix.CLOCAL | unix.CREAD | unix.CS8 | baudrate,
	})
	if err != nil {
		return nil, err
	}
	err = unix.SetNonblock(fd, false)
	if err != nil {
		return nil, err
	}
	return &Port{fd}, nil
}

// Close closes the given Port.
func (port *Port) Close() error {
	return unix.Close(port.fd)
}

// Write writes len(buf) bytes from buf to port.
func (port *Port) Write(buf []byte) error {
	n, err := unix.Write(port.fd, buf)
	if err != nil {
		return err
	}
	if n != len(buf) {
		return fmt.Errorf("wrote %d bytes instead of %d", n, len(buf))
	}
	return nil
}

// ReadAvailable reads available data from port into buf
// and returns the number of bytes read.
func (port *Port) ReadAvailable(buf []byte) (int, error) {
	return unix.Read(port.fd, buf)
}

// Read reads from port into buf, blocking if necessary
// until exactly len(buf) bytes have been read.
func (port *Port) Read(buf []byte) error {
	for off := 0; off < len(buf); {
		n, err := port.ReadAvailable(buf[off:])
		if err != nil {
			return err
		}
		off += n
	}
	return nil
}
