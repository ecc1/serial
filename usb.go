package serial

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// DeviceNotFoundError indicates that no USB device was found
// with the given vendor and product identifiers.
type DeviceNotFoundError struct {
	Vendor  int
	Product int
}

func (e DeviceNotFoundError) Error() string {
	return fmt.Sprintf("USB device %04x:%04x not found", e.Vendor, e.Product)
}

// TTYNotFoundError indicates that no tty device was associated with the given USB device,
// probably because the required cdc_acm kernel module has not been loaeded.
type TTYNotFoundError string

func (e TTYNotFoundError) Error() string {
	return fmt.Sprintf("no tty in %s: is the cdc-acm kernel module loaded?", string(e))
}

// NotCharDeviceError indicates that the given device pathname
// is not a character device.
type NotCharDeviceError string

func (e NotCharDeviceError) Error() string {
	return fmt.Sprintf("%s is not a character device", string(e))
}

// FindUSB returns the /dev/tty* path corresponding to the USB serial device
// with the given vendor and product identifiers.
func FindUSB(vendor, product int) (string, error) {
	device, err := findUSBDevice(vendor, product)
	if err != nil {
		return device, err
	}
	tty := ""
	err = filepath.Walk(device, func(path string, info os.FileInfo, err error) error {
		if tty != "" {
			return filepath.SkipDir
		}
		if err != nil {
			return err
		}
		dir, file := filepath.Split(path)
		if filepath.Base(dir) == "tty" && strings.HasPrefix(file, "tty") {
			tty = file
			return filepath.SkipDir
		}
		return nil
	})
	if err != nil && err != filepath.SkipDir {
		return "", err
	}
	if tty == "" {
		return "", TTYNotFoundError(device)
	}
	return charDevice(tty)
}

// charDevice returns the pathname for the given tty,
// after verifying it exists and is a character device.
func charDevice(tty string) (string, error) {
	device := filepath.Join("/dev", tty)
	s, err := os.Stat(device)
	if err != nil {
		return device, err
	}
	m := os.ModeDevice | os.ModeCharDevice
	if s.Mode()&m != m {
		return device, NotCharDeviceError(device)
	}
	return device, nil
}

// findUSBDevice returns the /sys/bus/usb/devices/... path for
// the USB device with the given vendor and product identifiers.
func findUSBDevice(vendor, product int) (string, error) {
	device := ""
	err := filepath.Walk("/sys/bus/usb/devices", func(path string, info os.FileInfo, err error) error {
		if device != "" {
			return filepath.SkipDir
		}
		if err != nil {
			return err
		}
		if hasID(path, "idVendor", vendor) && hasID(path, "idProduct", product) {
			device = path
			return filepath.SkipDir
		}
		return nil
	})
	if err != nil && err != filepath.SkipDir {
		return device, err
	}
	if device == "" {
		return device, DeviceNotFoundError{Vendor: vendor, Product: product}
	}
	// filepath.Walk won't follow symlinks, so expand it first.
	return filepath.EvalSymlinks(device)
}

// hasID tests if the file path/id contains the given value.
func hasID(path string, id string, value int) bool {
	v, err := ioutil.ReadFile(filepath.Join(path, id))
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(v)) == fmt.Sprintf("%04x", value)
}
