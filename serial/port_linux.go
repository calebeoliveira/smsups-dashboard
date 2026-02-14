//go:build linux

package serial

import "go.bug.st/serial"

// DiscoverPorts returns a list of available serial ports for Linux.
// Can be used as fallback when config PORTA is empty.
func DiscoverPorts() ([]string, error) {
	return serial.GetPortsList()
}
