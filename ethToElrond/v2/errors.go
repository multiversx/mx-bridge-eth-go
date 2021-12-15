package v2

import "errors"

// TODO add comments
var ErrNilBatch = errors.New("nil batch")
var ErrInvalidDepositNonce = errors.New("invalid deposit nonce")
var ErrNilLogger = errors.New("nil logger")
var ErrNilElrondClient = errors.New("nil Elrond client")
var ErrNilEthereumClient = errors.New("nil Ethereum client")
var ErrNilTopologyProvider = errors.New("nil topology provider")
var ErrNilElrondTopologyProvider = errors.New("nil elrond topology provider")
var ErrNilEthereumTopologyProvider = errors.New("nil ethereum topology provider")
var ErrInvalidDuration = errors.New("invalid duration")

// ErrNilExecutor signals that a nil bridge executor has been provided
var ErrNilExecutor = errors.New("nil bridge executor")

// ErrDuplicatedStepIdentifier signals that the same step identifier was used in 2 or more steps
var ErrDuplicatedStepIdentifier = errors.New("duplicated step identifier used in multiple steps")
