package parsers

import "errors"

var (
	errBufferTooShortForMarker     = errors.New("buffer too short for protocol indicator")
	errUnexpectedMarker            = errors.New("unexpected protocol indicator")
	errBufferTooShortForLength     = errors.New("buffer too short while extracting the length")
	errBufferTooShortForString     = errors.New("buffer too short while extracting the string data")
	errBufferTooShortForUint64     = errors.New("buffer too short for uint64")
	errBufferTooShortForUint32     = errors.New("buffer too short for uint32 length")
	errBufferTooShortForEthAddress = errors.New("buffer too short for Ethereum address")
	errBufferTooShortForMvxAddress = errors.New("buffer too short for MultiversX address")
	errBufferTooShortForBigInt     = errors.New("buffer too short while extracting the big.Int value")
	errBufferLenMismatch           = errors.New("buffer length mismatch")
	errEmptyBuffer                 = errors.New("empty buffer")
)
