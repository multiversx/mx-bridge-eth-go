package multiversx

import "errors"

var (
	errInvalidNumberOfResponseLines = errors.New("invalid number of responses")
	errNilProxy                     = errors.New("nil proxy")
	errNilCodec                     = errors.New("nil codec")
	errNilFilter                    = errors.New("nil filter")
	errNilLogger                    = errors.New("nil logger")
	errNilNonceTxHandler            = errors.New("nil nonce transaction handler")
	errNilPrivateKey                = errors.New("nil private key")
	errNilSingleSigner              = errors.New("nil single signer")
)
