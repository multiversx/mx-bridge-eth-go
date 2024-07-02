package parsers

import "errors"

var (
	errBufferTooShortForMarker   = errors.New("buffer too short for protocol indicator")
	errUnexpectedMarker          = errors.New("unexpected protocol indicator")
	errBufferTooShortForLength   = errors.New("buffer too short while extracting the length")
	errBufferTooShortForString   = errors.New("buffer too short while extracting the string data")
	errBufferTooShortForGasLimit = errors.New("buffer too short for gas limit")
	errBufferTooShortForNumArgs  = errors.New("buffer too short for numArguments length")
)
