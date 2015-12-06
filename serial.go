/*
Package usbserial is a simple library for I/O over USB serial devices.
*/
package usbserial

import (
	"fmt"
	s "syscall"
	"unsafe"
)

// A Port represents an open USB serial device.
type Port int

// Open returns a Port for the USB serial device with the given
// vendor and product identifiers.
// Its behavior is undefined if more than one such USB device is present.
func Open(vendor string, product string) (Port, error) {
	device, err := findUSB(vendor, product)
	if err != nil {
		return -1, err
	}
	const openFlags = s.O_NONBLOCK | s.O_NOCTTY | s.O_RDWR
	fd, err := s.Open(device, openFlags, 0)
	if err != nil {
		return -1, err
	}
	speed := uint32(s.B115200)
	t := s.Termios{
		Iflag: s.IGNPAR,
		Cflag: s.CLOCAL | s.CREAD | s.CS8 | speed,
	}
	_, _, errno := s.Syscall6(s.SYS_IOCTL, uintptr(fd), uintptr(s.TCSETS), uintptr(unsafe.Pointer(&t)), 0, 0, 0)
	if errno != 0 {
		return -1, errno
	}
	return Port(fd), nil
}

// Close closes the given Port.
func (fd Port) Close() error {
	return s.Close(int(fd))
}

// Write writes len(buf) bytes from buf to Port.
func (fd Port) Write(buf []byte) error {
	n, err := s.Write(int(fd), buf)
	if err != nil {
		return err
	}
	if n != len(buf) {
		return fmt.Errorf("wrote %d bytes instead of %d", n, len(buf))
	}
	return nil
}

// Read reads from Port into buf, blocking if necessary
// until exactly len(buf) bytes have been read.
func (fd Port) Read(buf []byte) error {
	for off := 0; off < len(buf); {
		n, err := s.Read(int(fd), buf[off:])
		if err != nil {
			return err
		}
		off += n
	}
	return nil
}
