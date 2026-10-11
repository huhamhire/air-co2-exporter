package monitor

import (
	"context"
	"testing"
)

func TestDecodeSensorData(t *testing.T) {
	tests := []struct {
		name     string
		data     [8]byte
		expected DeviceMonitorMetrics
	}{
		{"temperature", [8]byte{0x42, 0x12, 0xc0, 0x14, 0x0d}, DeviceMonitorMetrics{Temperature: 26.85}},
		{"co2", [8]byte{0x50, 0x03, 0x20, 0x73, 0x0d}, DeviceMonitorMetrics{PpmCo2: 800}},
		{"humidity", [8]byte{0x41, 0x0d, 0x80, 0xce, 0x0d}, DeviceMonitorMetrics{Humidity: 34.56}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &DeviceMonitor{connected: true}
			if err := m.decodeSensorData(tt.data); err != nil {
				t.Fatal(err)
			}
			if m.metrics != tt.expected {
				t.Fatalf("readings = %+v, want %+v", m.metrics, tt.expected)
			}
		})
	}
}

func TestDecodePacket(t *testing.T) {
	// Both fixtures represent a CO2 reading of 800 ppm.
	for _, packet := range [][]byte{
		{0x50, 0x03, 0x20, 0x73, 0x0d, 0, 0, 0},
		{0x72, 0xe4, 0x66, 0x20, 0x95, 0x46, 0xbf, 0xda},
	} {
		m := &DeviceMonitor{connected: true}
		if err := m.decodePacket(packet); err != nil {
			t.Fatal(err)
		}
		if got := m.GetCo2(); got != 800 {
			t.Fatalf("CO2 = %d, want 800", got)
		}
	}
}

func TestInvalidPacketsPreserveReadings(t *testing.T) {
	initial := DeviceMonitorMetrics{Temperature: 20, PpmCo2: 600, Humidity: 45}
	m := &DeviceMonitor{connected: true, metrics: initial}
	for n := range 8 {
		if err := m.decodePacket(make([]byte, n)); err == nil {
			t.Fatalf("packet of length %d should be rejected", n)
		}
	}
	for _, packet := range [][8]byte{
		{0x50, 0x03, 0x20, 0x74, 0x0d}, // Incorrect checksum.
		{0x50, 0x03, 0x20, 0x73, 0x0e}, // Incorrect terminator.
		{0x60, 0x03, 0x20, 0x83, 0x0d}, // Unknown measurement type.
	} {
		if err := m.decodeSensorData(packet); err != nil {
			t.Fatal(err)
		}
	}
	if m.metrics != initial {
		t.Fatalf("invalid packets changed readings: %+v", m.metrics)
	}
}

func TestReadDisconnected(t *testing.T) {
	m := NewDeviceMonitor(DefaultDevice())
	if err := m.ReadDataContext(context.Background()); err == nil {
		t.Fatal("reading a disconnected device should return an error")
	}
}

func TestValidateChecksum(t *testing.T) {
	for _, packet := range [][8]byte{
		{0x50, 0x03, 0x20, 0x73, 0x0d},
		{0x50, 0xff, 0xff, 0x4e, 0x0d}, // Checksum wraps at 256.
	} {
		if err := ValidateChecksum(packet); err != nil {
			t.Fatal(err)
		}
		packet[3]++
		if err := ValidateChecksum(packet); err == nil {
			t.Fatal("invalid checksum should be rejected")
		}
	}
}
