package sui

import (
	"encoding/binary"
	"encoding/hex"
)

func suiAddressFromBytes(bytes []byte) string {
	hexAddress := hex.EncodeToString(bytes[:])
	return "0x" + hexAddress
}

func AppendLengthToData(data []byte) []byte {
	if len(data) == 0 {
		return data
	}

	lenBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(lenBytes, uint32(len(data)))
	encoded := append(lenBytes, data...)

	return encoded
}
