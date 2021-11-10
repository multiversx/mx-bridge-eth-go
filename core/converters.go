package core

// TODO - move this as a method in AddressHandler

// ConvertFromByteSliceToArray will convert the provided buffer to its [32]byte representation
func ConvertFromByteSliceToArray(buff []byte) [32]byte {
	var result [32]byte
	copy(result[:], buff)

	return result
}
