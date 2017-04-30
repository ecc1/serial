package usbserial

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// hasID tests if the file path/id contains the given value.
func hasID(path string, id string, value int) bool {
	v, err := ioutil.ReadFile(filepath.Join(path, id))
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(v)) == fmt.Sprintf("%04x", value)
}

type DeviceNotFoundError struct {
	Vendor  int
	Product int
}

func (e DeviceNotFoundError) Error() string {
	return fmt.Sprintf("USB device %s:%s not found", e.Vendor, e.Product)
}

type TTYNotFoundError string

func (e TTYNotFoundError) Error() string {
	return fmt.Sprintf("no tty in %s: is the cdc-acm kernel module loaded?", string(e))
}

// findUSB returns the /dev/tty* path corresponding to the USB serial device
// with the given vendor and product identifiers.
func findUSB(vendor, product int) (string, error) {
	found := false
	device := ""
	checkId := func(path string, info os.FileInfo, err error) error {
		if found {
			return filepath.SkipDir
		}
		if err != nil {
			return err
		}
		if hasID(path, "idVendor", vendor) && hasID(path, "idProduct", product) {
			found = true
			device = path
			return filepath.SkipDir
		}
		return nil
	}
	err := filepath.Walk("/sys/bus/usb/devices", checkId)
	if err != nil && err != filepath.SkipDir {
		return "", err
	}
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
	checkTTY := func(path string, info os.FileInfo, err error) error {
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
	filepath.Walk(device, checkTTY)
	if err != nil && err != filepath.SkipDir {
		return "", err
	}
	if !found {
		return "", TTYNotFoundError(device)
	}
	return filepath.Join("/dev", tty), nil
}
