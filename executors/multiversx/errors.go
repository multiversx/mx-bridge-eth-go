package multiversx

import "errors"

var (
	errInvalidNumberOfResponseLines  = errors.New("invalid number of responses")
	errNilProxy                      = errors.New("nil proxy")
	errNilCodec                      = errors.New("nil codec")
	errNilFilter                     = errors.New("nil filter")
	errNilLogger                     = errors.New("nil logger")
	errNilNonceTxHandler             = errors.New("nil nonce transaction handler")
	errNilPrivateKey                 = errors.New("nil private key")
	errNilSingleSigner               = errors.New("nil single signer")
	errInvalidValue                  = errors.New("invalid value")
	errNilCloseAppChannel            = errors.New("nil close application channel")
	errTransactionFailed             = errors.New("transaction failed")
	errMaxGasLimitIsLessThanRequired = errors.New("max gas limit to execute a SC call is less than the minimum required")
)
