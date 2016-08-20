package usbserial

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// hasProperty tests if the file path/property contains the given value.
func hasProperty(path string, property string, value string) bool {
	v, err := ioutil.ReadFile(filepath.Join(path, property))
	if err != nil {
		return false
	}
	if len(v) != 0 && v[len(v)-1] == '\n' {
		// Compare without trailing '\n'
		return string(v[:len(v)-1]) == value
	}
	return string(v) == value
}

type DeviceNotFoundError struct {
	Vendor  string
	Product string
}

func (e DeviceNotFoundError) Error() string {
	return fmt.Sprintf("USB device %s:%s not found", e.Vendor, e.Product)
}

type TtyNotFoundError string

func (e TtyNotFoundError) Error() string {
	return fmt.Sprintf("no tty in %s: is the cdc-acm kernel module loaded?", string(e))
}

// findUSB returns the /dev/tty* path corresponding to the USB serial device
// with the given vendor and product identifiers.
func findUSB(vendor string, product string) (path string, err error) {
	found := false
	device := ""
	checkId := func(path string, info os.FileInfo, err error) error {
		if found {
			return filepath.SkipDir
		}
		if err != nil {
			return err
		}
		if hasProperty(path, "idVendor", vendor) && hasProperty(path, "idProduct", product) {
			found = true
			device = path
			return filepath.SkipDir
		}
		return nil
	}
	filepath.Walk("/sys/bus/usb/devices", checkId)
	if !found {
		return "", DeviceNotFoundError{Vendor: vendor, Product: product}
	}
	// filepath.Walk won't follow symlinks, so expand it first.
	device, err = filepath.EvalSymlinks(device)
	if err != nil {
		return "", err
	}
	found = false
	tty := ""
	checkTty := func(path string, info os.FileInfo, err error) error {
		if found {
			return filepath.SkipDir
		}
		if err != nil {
			return err
		}
		dir, file := filepath.Split(path)
		if filepath.Base(dir) == "tty" && strings.HasPrefix(file, "tty") {
			found = true
			tty = file
			return filepath.SkipDir
		}
		return nil
	}
	filepath.Walk(device, checkTty)
	if !found {
		return "", TtyNotFoundError(device)
	}
	return filepath.Join("/dev", tty), nil
}
