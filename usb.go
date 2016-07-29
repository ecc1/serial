package usbserial

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// hasProperty tests if the file path/property contains the given value
func hasProperty(path string, property string, value string) bool {
	v, err := ioutil.ReadFile(filepath.Join(path, property))
	if err != nil {
		return false
	}
	// compare without trailing '\n'
	return string(v[:len(v)-1]) == value
}

// findUSB returns the /dev/tty* path corresponding to the USB serial device
// with the given vendor and product identifiers
func findUSB(vendor string, product string) (path string, err error) {
	found := false
	device := ""
	check := func(path string, info os.FileInfo, err error) error {
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
	filepath.Walk("/sys/bus/usb/devices", check)
	if !found {
		return "", fmt.Errorf("no device with vendor %q and product %q", vendor, product)
	}
	// filepath.Walk won't follow symlink, so expand it first
	device, err = filepath.EvalSymlinks(device)
	if err != nil {
		return "", err
	}
	found = false
	tty := ""
	check = func(path string, info os.FileInfo, err error) error {
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
	filepath.Walk(device, check)
	if !found {
		return "", fmt.Errorf("no tty in %s: is the cdc-acm kernel module loaded?", device)
	}
	return filepath.Join("/dev", tty), nil
}
