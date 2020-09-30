package monitor

import (
	"github.com/google/gousb"
	"log"
)

type DeviceMonitorMetrics struct {
	Temperature float64
	PpmCo2      uint16
}

type DeviceMonitor struct {
	device  Device
	metrics DeviceMonitorMetrics
}

func NewDeviceMonitor(dev Device) DeviceMonitor {
	return DeviceMonitor{
		device: dev,
		metrics: DeviceMonitorMetrics{
			0,
			0,
		},
	}
}

func (m *DeviceMonitor) Connect() error {
	ctx := gousb.NewContext()
	dev, err := ctx.OpenDeviceWithVIDPID(m.device.vid, m.device.pid)
	if err != nil {
		log.Fatal(err)
		return err
	}
	_ = dev.SetAutoDetach(true)
	inf, done, _ := dev.DefaultInterface()
	_, _ = dev.Control(conf.rType, conf.request, conf.val, conf.idx, conf.randomKey)

	ep, _ := inf.InEndpoint(0)
	var rawData [64]byte
	_, _ = ep.Read(rawData[:])
	data := DecryptRawData(rawData[:])
	err = m.ReadData(data)
	if err != nil {
		log.Fatal(err)
	}
	done()
	return nil
}

func (m *DeviceMonitor) ReadData(data [8]byte) error {
	err := ValidateChecksum(data)
	if err != nil {
		return err
	}
	dataType := data[0]
	value := uint16(data[1])<<8 | uint16(data[2])
	switch dataType {
	case dataTypeTemp:
		temp := float64(value)/16 - 273.15
		m.metrics.Temperature = temp
		log.Printf("temperature %.2f", temp)
	case dataTypeCo2:
		ppmCo2 := value
		m.metrics.PpmCo2 = ppmCo2
		log.Printf("ppmCo2: %d", ppmCo2)
	}
	return nil
}
