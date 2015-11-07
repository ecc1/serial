package usbserial

import (
	"fmt"
	s "syscall"
	"unsafe"
)

type Port int

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

func Write(dev Port, buf []byte) error {
	n, err := s.Write(int(dev), buf)
	if err != nil {
		return err
	}
	if n != len(buf) {
		return fmt.Errorf("wrote %d bytes instead of %d", n, len(buf))
	}
	return nil
}

func Read(dev Port, buf []byte) error {
	for off := 0; off < len(buf); {
		n, err := s.Read(int(dev), buf[off:])
		if err != nil {
			return err
		}
		off += n
	}
	return nil
}
