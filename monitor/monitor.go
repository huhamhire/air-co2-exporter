package monitor

import (
	"errors"
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/google/gousb"
	"math"
)

type DeviceMonitorMetrics struct {
	Temperature float64
	PpmCo2      uint16
}

type DeviceMonitor struct {
	device    Device
	metrics   DeviceMonitorMetrics
	connected bool
	endPoint  *gousb.InEndpoint
	logger    *log.Logger
}

func NewDeviceMonitor(dev Device) *DeviceMonitor {
	return &DeviceMonitor{
		device: dev,
		metrics: DeviceMonitorMetrics{
			0,
			0,
		},
		connected: false,
		endPoint:  nil,
		logger:    nil,
	}
}

func (m *DeviceMonitor) SetLogger(logger *log.Logger) {
	m.logger = logger
}

func (m *DeviceMonitor) Connect() (done func(), err error) {
	ctx := gousb.NewContext()
	dev, err := ctx.OpenDeviceWithVIDPID(m.device.vid, m.device.pid)
	if err != nil {
		return nil, err
	}
	if dev == nil {
		return nil, errors.New("device not found")
	}
	_ = dev.SetAutoDetach(true)
	inf, done, _ := dev.DefaultInterface()

	_, err = dev.Control(conf.rType, conf.request, conf.val, conf.idx, conf.randomKey)
	if err != nil {
		return done, err
	}

	m.endPoint, err = inf.InEndpoint(1)
	if err != nil {
		m.connected = false
		return done, err
	}
	m.connected = true

	return done, nil
}

func (m *DeviceMonitor) ReadData() error {
	rawData := make([]byte, m.endPoint.Desc.MaxPacketSize)
	_, err := m.endPoint.Read(rawData[:])

	if err != nil {
		return err
	}

	level.Debug(*m.logger).Log("msg", fmt.Sprintf("Raw data: %v", rawData))

	var data [8]byte
	if rawData[4] != 0x0d {
		data = DecryptRawData(rawData[:])
	} else {
		copy(data[:], rawData)
	}

	err = m.decodeSensorData(data)
	if err != nil {
		return err
	}
	return nil
}

func toFixed(num float64, precision int) float64 {
	p := math.Pow10(precision)
	return float64(int(num*p)) / p
}

// Decode raw data from sensor device
func (m *DeviceMonitor) decodeSensorData(data [8]byte) error {
	if m.connected != true {
		return errors.New("device not connected")
	}
	err := ValidateChecksum(data)
	if err != nil {
		// skip incorrect data
		return nil
	}
	dataType := data[0]
	value := uint16(data[1])<<8 | uint16(data[2])
	switch dataType {
	case dataTypeTemp:
		temp := toFixed(float64(value)/16-273.15, 2)
		m.metrics.Temperature = temp
		if m.logger != nil {
			_ = level.Debug(*m.logger).Log("msg", "temperature metric", "value", temp)
		}
	case dataTypeCo2:
		ppmCo2 := value
		m.metrics.PpmCo2 = ppmCo2
		if m.logger != nil {
			_ = level.Debug(*m.logger).Log("msg", "co2 metric", "value", ppmCo2)
		}
	}
	return nil
}

func (m *DeviceMonitor) GetTemp() float64 {
	return m.metrics.Temperature
}

func (m *DeviceMonitor) GetCo2() uint16 {
	return m.metrics.PpmCo2
}
