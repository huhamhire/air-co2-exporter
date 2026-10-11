package monitor

import "errors"

var (
	cState = []byte{
		0x48, 0x74, 0x65, 0x6D, 0x70, 0x39, 0x39, 0x65,
	}
	shuffle = []uint8{2, 4, 0, 7, 1, 6, 5, 3}
)

// DecryptRawData decrypts a sensor packet. rawData must contain at least eight bytes.
func DecryptRawData(rawData []byte) [8]byte {
	var dataXor [8]byte
	for i := range cState {
		idx := shuffle[i]
		dataXor[idx] = rawData[i] ^ conf.randomKey[idx]
	}
	var dataTemp [8]byte
	for i := range cState {
		dataTemp[i] = ((dataXor[i] >> 3) | (dataXor[(i-1+8)%8] << 5)) & 0xff
	}
	var results [8]byte
	for i, state := range cState {
		cTemp := ((state >> 4) | (state << 4)) & 0xff
		results[i] = uint8((0x100 + uint16(dataTemp[i]) - uint16(cTemp)) & uint16(0xff))
	}
	return results
}

// ValidateChecksum checks the checksum and terminator of a decrypted packet.
func ValidateChecksum(data [8]byte) error {
	checksum := data[3]
	sum := (data[0] + data[1] + data[2]) & 0xff
	if data[4] != 0x0d || checksum != sum {
		return errors.New("checksum not matched, please check the device")
	}
	return nil
}
