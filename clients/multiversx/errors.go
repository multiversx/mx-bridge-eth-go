package multiversx

import "errors"

var (
	errNilLogger                = errors.New("nil logger")
	errNilProxy                 = errors.New("nil proxy")
	errNilAddressHandler        = errors.New("nil address handler")
	errNilRequest               = errors.New("nil request")
	errInvalidNumberOfArguments = errors.New("invalid number of arguments")
	errInvalidGasValue          = errors.New("invalid gas value")
	errNoStatusForBatchID       = errors.New("no status for batch ID")
	errBatchNotFinished         = errors.New("batch not finished")
	errMalformedBatchResponse   = errors.New("malformed batch response")
	errNilRoleProvider          = errors.New("nil role provider")
	errRelayerNotWhitelisted    = errors.New("relayer not whitelisted")
	errNilNodeStatusResponse    = errors.New("nil node status response")
	errInvalidBalance           = errors.New("invalid balance")
	errInsufficientESDTBalance  = errors.New("insufficient ESDT balance")
)
