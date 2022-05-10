package elrond

import "errors"

var (
	errNilLogger                = errors.New("nil logger")
	errNilProxy                 = errors.New("nil ElrondProxy")
	errNilAddressHandler        = errors.New("nil address handler")
	errNilRequest               = errors.New("nil request")
	errInvalidNumberOfArguments = errors.New("invalid number of arguments")
	errNotUint64Bytes           = errors.New("provided bytes do not represent a valid uint64 number")
	errInvalidGasValue          = errors.New("invalid gas value")
	errNoStatusForBatchID       = errors.New("no status for batch ID")
	errBatchNotFinished         = errors.New("batch not finished")
	errMalformedBatchResponse   = errors.New("malformed batch response")
	errNilRoleProvider          = errors.New("nil role provider")
	errRelayerNotWhitelisted    = errors.New("relayer not whitelisted")
	errNilNodeStatusResponse    = errors.New("nil node status response")

	// ErrNoPendingBatchAvailable signals that no pending batch is available
	ErrNoPendingBatchAvailable = errors.New("no pending batch available")
)
