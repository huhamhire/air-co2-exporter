package monitor

import "github.com/google/gousb"

// Device identifies a USB sensor by vendor and product ID.
type Device struct {
	vid gousb.ID
	pid gousb.ID
}

// DefaultDevice returns the supported sensor's default USB IDs.
func DefaultDevice() Device {
	return NewDevice(0x04D9, 0xA052)
}

// NewDevice creates a device configuration with the given USB IDs.
func NewDevice(vid gousb.ID, pid gousb.ID) Device {
	return Device{vid, pid}
}
