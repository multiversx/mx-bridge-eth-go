package ethElrond

import "errors"

// ErrNilBatch signals that a nil batch has been provided
var ErrNilBatch = errors.New("nil batch")

// ErrInvalidDepositNonce signals that an invalid deposit nonce has been provided
var ErrInvalidDepositNonce = errors.New("invalid deposit nonce")

// ErrNilLogger signals that a nil logger has been provided
var ErrNilLogger = errors.New("nil logger")

// ErrNilElrondClient signals that a nil elrond client has been provided
var ErrNilElrondClient = errors.New("nil Elrond client")

// ErrNilEthereumClient signals that a nil ethereum client has been provided
var ErrNilEthereumClient = errors.New("nil Ethereum client")

// ErrNilTopologyProvider signals that a nil topology provider has been provided
var ErrNilTopologyProvider = errors.New("nil topology provider")

// ErrInvalidDuration signals that an invalid duration has been provided
var ErrInvalidDuration = errors.New("invalid duration")

// ErrNilExecutor signals that a nil bridge executor has been provided
var ErrNilExecutor = errors.New("nil bridge executor")

// ErrDuplicatedStepIdentifier signals that the same step identifier was used in 2 or more steps
var ErrDuplicatedStepIdentifier = errors.New("duplicated step identifier used in multiple steps")

// ErrNilStatusHandler signals that a nil status handler was provided
var ErrNilStatusHandler = errors.New("nil status handler")

// ErrBatchNotFound signals that the batch was not found
var ErrBatchNotFound = errors.New("batch not found")
