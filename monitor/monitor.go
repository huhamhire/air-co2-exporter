package monitor

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"sync"

	"github.com/google/gousb"
)

// DeviceMonitorMetrics holds the latest readings from the sensor.
type DeviceMonitorMetrics struct {
	Temperature float64
	PpmCo2      uint16
	Humidity    float64
}

// DeviceMonitor reads a single USB sensor. Its methods require serial access.
type DeviceMonitor struct {
	device    Device
	metrics   DeviceMonitorMetrics
	connected bool
	endPoint  *gousb.InEndpoint
	logger    *slog.Logger
}

// NewDeviceMonitor creates a monitor for dev.
func NewDeviceMonitor(dev Device) *DeviceMonitor {
	return &DeviceMonitor{device: dev}
}

// SetLogger sets the optional logger for the monitor.
func (m *DeviceMonitor) SetLogger(logger *slog.Logger) {
	m.logger = logger
}

// Connect opens the USB device and returns an idempotent cleanup function.
// Stop pending reads before calling the cleanup function.
func (m *DeviceMonitor) Connect() (func(), error) {
	if m.connected {
		return nil, errors.New("device already connected")
	}
	ctx := gousb.NewContext()
	dev, err := ctx.OpenDeviceWithVIDPID(m.device.vid, m.device.pid)
	if err != nil {
		if dev != nil {
			_ = dev.Close()
		}
		_ = ctx.Close()
		return nil, fmt.Errorf("open USB device: %w", err)
	}
	if dev == nil {
		_ = ctx.Close()
		return nil, errors.New("device not found")
	}
	var releaseInterface func()
	done := sync.OnceFunc(func() {
		m.connected = false
		m.endPoint = nil
		if releaseInterface != nil {
			releaseInterface()
		}
		_ = dev.Close()
		_ = ctx.Close()
	})
	if err := dev.SetAutoDetach(true); err != nil {
		done()
		return nil, fmt.Errorf("configure USB auto-detach: %w", err)
	}
	inf, release, err := dev.DefaultInterface()
	if err != nil {
		done()
		return nil, fmt.Errorf("claim USB interface: %w", err)
	}
	releaseInterface = release

	_, err = dev.Control(conf.rType, conf.request, conf.val, conf.idx, conf.randomKey)
	if err != nil {
		done()
		return nil, fmt.Errorf("initialize sensor: %w", err)
	}

	m.endPoint, err = inf.InEndpoint(1)
	if err != nil {
		done()
		return nil, fmt.Errorf("open USB input endpoint: %w", err)
	}
	m.connected = true

	return done, nil
}

// ReadData reads and decodes one packet without a cancellation deadline.
func (m *DeviceMonitor) ReadData() error {
	return m.ReadDataContext(context.Background())
}

// ReadDataContext reads one packet, allowing cancellation of a blocked USB transfer.
func (m *DeviceMonitor) ReadDataContext(ctx context.Context) error {
	if !m.connected || m.endPoint == nil {
		return errors.New("device not connected")
	}
	rawData := make([]byte, m.endPoint.Desc.MaxPacketSize)
	n, err := m.endPoint.ReadContext(ctx, rawData)

	if err != nil {
		return err
	}
	return m.decodePacket(rawData[:n])
}

func (m *DeviceMonitor) decodePacket(rawData []byte) error {
	if len(rawData) < 8 {
		return fmt.Errorf("short sensor packet: got %d bytes, want at least 8", len(rawData))
	}

	if m.logger != nil {
		m.logger.Debug("Raw data", "data", rawData)
	}

	var data [8]byte
	if rawData[4] != 0x0d {
		data = DecryptRawData(rawData)
	} else {
		copy(data[:], rawData)
	}

	return m.decodeSensorData(data)
}

func toFixed(num float64, precision int) float64 {
	p := math.Pow10(precision)
	return float64(int(num*p)) / p
}

// decodeSensorData updates the reading represented by a valid packet.
func (m *DeviceMonitor) decodeSensorData(data [8]byte) error {
	if !m.connected {
		return errors.New("device not connected")
	}
	if err := ValidateChecksum(data); err != nil {
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
			m.logger.Debug("temperature metric", "value", temp)
		}
	case dataTypeCo2:
		ppmCo2 := value
		m.metrics.PpmCo2 = ppmCo2
		if m.logger != nil {
			m.logger.Debug("co2 metric", "value", ppmCo2)
		}
	case dataTypeHum:
		humidity := float64(value) / 100
		m.metrics.Humidity = humidity
		if m.logger != nil {
			m.logger.Debug("humidity metric", "value", humidity)
		}
	}
	return nil
}

// GetTemp returns the latest temperature in degrees Celsius.
func (m *DeviceMonitor) GetTemp() float64 {
	return m.metrics.Temperature
}

// GetCo2 returns the latest CO2 concentration in ppm.
func (m *DeviceMonitor) GetCo2() uint16 {
	return m.metrics.PpmCo2
}

// GetHum returns the latest relative humidity in percent.
func (m *DeviceMonitor) GetHum() float64 {
	return m.metrics.Humidity
}
