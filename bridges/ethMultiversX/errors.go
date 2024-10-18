package ethmultiversx

import "errors"

// ErrNilBatch signals that a nil batch has been provided
var ErrNilBatch = errors.New("nil batch")

// ErrInvalidDepositNonce signals that an invalid deposit nonce has been provided
var ErrInvalidDepositNonce = errors.New("invalid deposit nonce")

// ErrNilLogger signals that a nil logger has been provided
var ErrNilLogger = errors.New("nil logger")

// ErrNilMultiversXClient signals that a nil MultiversX client has been provided
var ErrNilMultiversXClient = errors.New("nil MultiversX client")

// ErrNilEthereumClient signals that a nil Ethereum client has been provided
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

// ErrFinalBatchNotFound signals that a final batch was not found
var ErrFinalBatchNotFound = errors.New("final batch not found")

// ErrNilSignaturesHolder signals that a nil signatures holder was provided
var ErrNilSignaturesHolder = errors.New("nil signatures holder")

// ErrNilBalanceValidator signals that a nil balance validator was provided
var ErrNilBalanceValidator = errors.New("nil balance validator")
