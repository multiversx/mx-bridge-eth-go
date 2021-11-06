package relay

import "errors"

// ErrInvalidDurationConfig signals that an invalid config duration was provided
var ErrInvalidDurationConfig = errors.New("invalid config duration")

// ErrMissingDurationConfig signals that a missing config duration was detected
var ErrMissingDurationConfig = errors.New("missing config duration")

// ErrMissingConfig signals that a missing config was detected
var ErrMissingConfig = errors.New("missing config")

// ErrMissingGeneralConfig signals that a missing general config was detected
var ErrMissingGeneralConfig = errors.New("missing general config")

// ErrMissingApiRoutesConfig signals that a missing api routes config was detected
var ErrMissingApiRoutesConfig = errors.New("missing api routes config")

// ErrMissingFlagsConfig signals that a missing flags config was detected
var ErrMissingFlagsConfig = errors.New("missing flags config")

// ErrNilElrondProxy signals that a nil elrond proxy was provided
var ErrNilElrondProxy = errors.New("nil elrond proxy")

// ErrNilEthClient signals that a nil eth client was provided
var ErrNilEthClient = errors.New("nil eth client")

// ErrNilEthInstance signals that a nil eth instance was provided
var ErrNilEthInstance = errors.New("nil eth instance")

// ErrNilMessenger signals that a nil messenger was provided
var ErrNilMessenger = errors.New("nil messenger")

// ErrNilErc20Contracts signals that a nil ERC20 contracts map was provided
var ErrNilErc20Contracts = errors.New("nil ERC20 contracts")

// ErrNilStatusHandler signals that a nil status handler was provided
var ErrNilStatusHandler = errors.New("nil status handler")

// ErrNilStatusStorer signals that a nil status storer was provided
var ErrNilStatusStorer = errors.New("nil status storer")
