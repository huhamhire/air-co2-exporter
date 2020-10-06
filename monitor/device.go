package monitor

import "github.com/google/gousb"

// Device config structure
type Device struct {
	vid gousb.ID
	pid gousb.ID
}

// DefaultDevice - get default device config
func DefaultDevice() Device {
	return NewDevice(0x04D9, 0xA052)
}

// NewDevice - create a new Device config instance
func NewDevice(vid gousb.ID, pid gousb.ID) Device {
	return Device{vid, pid}
}
