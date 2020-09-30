package monitor

import "github.com/google/gousb"

type Device struct {
	vid gousb.ID
	pid gousb.ID
}

func DefaultDevice() Device {
	return NewDevice(0x04D9, 0xA052)
}

func NewDevice(vid gousb.ID, pid gousb.ID) Device {
	return Device{vid, pid}
}
