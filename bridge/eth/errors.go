package eth

import "errors"

// ErrNilConfig signals that a nil config was provided
var ErrNilConfig = errors.New("nil config")

// ErrNilBroadcaster signals that a nil broadcaster was provided
var ErrNilBroadcaster = errors.New("nil broadcaster")

// ErrNilMapper signals that a nil mapper was provided
var ErrNilMapper = errors.New("nil mapper")

// ErrNilGasHandler signals that a nil gas handler was provided
var ErrNilGasHandler = errors.New("nil gas handler")

// ErrNilBlockchainClient signals that a nil blockchain client was provided
var ErrNilBlockchainClient = errors.New("nil blockchain client")

// ErrNilBrdgeContract signals that a nil blockchain client was provided
var ErrNilBrdgeContract = errors.New("nil bridge contract")

// ErrNilSignaturesHolder signals that a nil signatures holder was provided
var ErrNilSignaturesHolder = errors.New("nil signatures holder")
