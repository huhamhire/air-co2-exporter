package monitor

// Config structure
type Config struct {
	randomKey []byte
	rType     uint8
	request   uint8
	val       uint16
	idx       uint16
}

var conf = Config{
	randomKey: []byte{
		0xc4, 0xc6, 0xc0, 0x92, 0x40, 0x23, 0xdc, 0x96,
	},
	rType:   uint8(0x21),
	request: uint8(0x09),
	val:     uint16(0x300),
	idx:     uint16(0x00),
}

const (
	dataTypeTemp byte = 0x42
	dataTypeCo2  byte = 0x50
	dataTypeHum  byte = 0x41
)
