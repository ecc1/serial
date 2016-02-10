/*
Package usbserial is a simple library for I/O over USB serial devices.
*/
package usbserial

import (
	"fmt"
	"unsafe"

	"golang.org/x/sys/unix"
)

// A Port represents an open USB serial device.
type Port struct {
	fd int
}

// Open returns a Port for the USB serial device with the given
// vendor and product identifiers.
// Its behavior is undefined if more than one such USB device is present.
func Open(vendor string, product string) (port Port, err error) {
	device, err := findUSB(vendor, product)
	if err != nil {
		return
	}
	const openFlags = unix.O_NONBLOCK | unix.O_NOCTTY | unix.O_RDWR
	port.fd, err = unix.Open(device, openFlags, 0)
	if err != nil {
		return
	}
	speed := uint32(unix.B115200)
	t := unix.Termios{
		Iflag: unix.IGNPAR,
		Cflag: unix.CLOCAL | unix.CREAD | unix.CS8 | speed,
	}
	_, _, errno := unix.Syscall6(
		unix.SYS_IOCTL, uintptr(port.fd), uintptr(unix.TCSETS),
		uintptr(unsafe.Pointer(&t)), 0, 0, 0,
	)
	if errno != 0 {
		err = error(errno)
	}
	return
}

// Close closes the given Port.
func (port Port) Close() error {
	return unix.Close(port.fd)
}

// Write writes len(buf) bytes from buf to port.
func (port Port) Write(buf []byte) error {
	n, err := unix.Write(port.fd, buf)
	if err != nil {
		return err
	}
	if n != len(buf) {
		return fmt.Errorf("wrote %d bytes instead of %d", n, len(buf))
	}
	return nil
}

// Read reads from port into buf, blocking if necessary
// until exactly len(buf) bytes have been read.
func (port Port) Read(buf []byte) error {
	for off := 0; off < len(buf); {
		n, err := unix.Read(port.fd, buf[off:])
		if err != nil {
			return err
		}
		off += n
	}
	return nil
}
