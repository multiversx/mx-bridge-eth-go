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

// ErrNilClientWrapper signals that a nil client wrapper was provided
var ErrNilClientWrapper = errors.New("nil client wrapper")

// ErrNilSignaturesHolder signals that a nil signatures holder was provided
var ErrNilSignaturesHolder = errors.New("nil signatures holder")

// ErrNilErc20Contracts signals that a nil ERC20 contracts map was provided
var ErrNilErc20Contracts = errors.New("nil ERC20 contracts")

// ErrNilErc20ContractInstance signals that a nil ERC20 contract instance was provided
var ErrNilErc20ContractInstance = errors.New("nil ERC20 contract instance")

// ErrInsufficientErc20Balance signals that the ERC20 balance is insufficient
var ErrInsufficientErc20Balance = errors.New("insufficient ERC20 balance")

// ErrMissingErc20ContractDefinition signals that the ERC20 contract was not defined
var ErrMissingErc20ContractDefinition = errors.New("missing ERC20 contract definition")

// ErrEmptyBridgeAddress signals that an empty bridge address was provided
var ErrEmptyBridgeAddress = errors.New("empty bridge address")
