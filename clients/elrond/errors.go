package elrond

import "errors"

var (
	errNilProxy                 = errors.New("nil ElrondProxy")
	errNilAddressHandler        = errors.New("nil address handler")
	errNilRequest               = errors.New("nil request")
	errInvalidNumberOfArguments = errors.New("invalid number of arguments")
	errNotUint64Bytes           = errors.New("provided bytes do not represent a valid uint64 number")
	errInvalidGasValue          = errors.New("invalid gas value")
	errNilLogger                = errors.New("nil logger")
	errNilPrivateKey            = errors.New("nil private key")

	// ErrNoPendingBatchAvailable signals that no pending batch is available
	ErrNoPendingBatchAvailable = errors.New("no pending batch available")
)
