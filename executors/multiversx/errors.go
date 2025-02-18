package multiversx

import "errors"

var (
	errInvalidNumberOfResponseLines      = errors.New("invalid number of responses")
	errNilProxy                          = errors.New("nil proxy")
	errNilCodec                          = errors.New("nil codec")
	errNilFilter                         = errors.New("nil filter")
	errNilLogger                         = errors.New("nil logger")
	errNilNonceTxHandler                 = errors.New("nil nonce transaction handler")
	errNilPrivateKey                     = errors.New("nil private key")
	errNilSingleSigner                   = errors.New("nil single signer")
	errInvalidValue                      = errors.New("invalid value")
	errTransactionFailed                 = errors.New("transaction failed")
	errGasLimitIsLessThanAbsoluteMinimum = errors.New("provided gas limit is less than absolute minimum required")
	errEmptyListOfBridgeSCProxy          = errors.New("the bridge SC proxy addresses list is empty")
	errNilTransactionExecutor            = errors.New("nil transaction executor")
)
