package clients

import "errors"

var (
	// ErrNilLogger signals that a nil logger was provided
	ErrNilLogger = errors.New("nil logger")

	// ErrNilDataGetter signals that a nil data getter was provided
	ErrNilDataGetter = errors.New("nil data getter")

	// ErrInvalidValue signals that an invalid value was provided
	ErrInvalidValue = errors.New("invalid value")

	// ErrNilPrivateKey signals that a nil private key was provided
	ErrNilPrivateKey = errors.New("nil private key")

	// ErrNilBatch signals that a nil batch was provided
	ErrNilBatch = errors.New("nil batch")

	// ErrNilTokensMapper signals that a nil tokens mapper was provided
	ErrNilTokensMapper = errors.New("nil tokens mapper")

	// ErrNilStatusHandler signals that a nil status handler was provided
	ErrNilStatusHandler = errors.New("nil status handler")

	// ErrNilAddressConverter signals that a nil address converter was provided
	ErrNilAddressConverter = errors.New("nil address converter")

	// ErrMultisigContractPaused signals that the multisig contract is paused
	ErrMultisigContractPaused = errors.New("multisig contract paused")

	// ErrNoBatchAvailable signals that the batch is not available
	ErrNoBatchAvailable = errors.New("no batch available")

	// ErrNoPendingBatchAvailable signals that no pending batch is available
	ErrNoPendingBatchAvailable = errors.New("no pending batch available")
)
