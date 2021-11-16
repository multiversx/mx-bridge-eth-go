package wrappers

import "errors"

// ErrNilBlockchainClient signals that a nil blockchain client was provided
var ErrNilBlockchainClient = errors.New("nil blockchain client")

// ErrNilBrdgeContract signals that a nil blockchain client was provided
var ErrNilBrdgeContract = errors.New("nil bridge contract")

// ErrNilStatusHandler signals that a nil status handler was provided
var ErrNilStatusHandler = errors.New("nil status handler")

// ErrNilErc20Contract signals that a nil ERC20 contract was provided
var ErrNilErc20Contract = errors.New("nil ERC20 contract")

// ErrInvalidQuorumValue signals that an invalid quorum value was received
var ErrInvalidQuorumValue = errors.New("invalid quorum value")
