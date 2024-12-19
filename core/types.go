package core

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-sdk-go/core"
)

// StepIdentifier defines a step name
type StepIdentifier string

// MachineStates defines all available steps for a state machine to run
type MachineStates map[StepIdentifier]Step

// Step defines a state machine step
type Step interface {
	Execute(ctx context.Context) StepIdentifier
	Identifier() StepIdentifier
	IsInterfaceNil() bool
}

// EthGasPriceSelector defines the ethereum gas price selector
type EthGasPriceSelector string

// Timer defines operations related to time
type Timer interface {
	NowUnix() int64
	Start()
	Close() error
	IsInterfaceNil() bool
}

// AddressConverter can convert a provided address bytes to its string representation
type AddressConverter interface {
	ToHexString(addressBytes []byte) string
	ToHexStringWithPrefix(addressBytes []byte) string
	ToBech32String(addressBytes []byte) (string, error)
	ToBech32StringSilent(addressBytes []byte) string
	IsInterfaceNil() bool
}

// BroadcastClient defines a broadcast client that will get notified by the broadcaster
// when new messages arrive. It also should be able to respond with any stored messages it might
// have.
type BroadcastClient interface {
	ProcessNewMessage(msg *SignedMessage, ethMsg *EthereumSignature)
	AllStoredSignatures() []*SignedMessage
	IsInterfaceNil() bool
}

// StatusHandler is able to keep metrics
type StatusHandler interface {
	SetIntMetric(metric string, value int)
	AddIntMetric(metric string, delta int)
	SetStringMetric(metric string, val string)
	GetAllMetrics() GeneralMetrics
	Name() string
	IsInterfaceNil() bool
}

// MetricsHolder represents the component that can hold metrics
type MetricsHolder interface {
	AddStatusHandler(sh StatusHandler) error
	GetAvailableStatusHandlers() []string
	GetAllMetrics(name string) (GeneralMetrics, error)
	IsInterfaceNil() bool
}

// Storer defines a component able to store and load data
type Storer interface {
	Put(key, data []byte) error
	Get(key []byte) ([]byte, error)
	Close() error
	IsInterfaceNil() bool
}

// GeneralMetrics represents an objects metrics map
type GeneralMetrics map[string]interface{}

// StringMetrics represents string metrics map
type StringMetrics map[string]string

// IntMetrics represents string metrics map
type IntMetrics map[string]int

// ClientStatus represents the possible statuses of a client
type ClientStatus int

const (
	Available   ClientStatus = 0
	Unavailable ClientStatus = 1
)

// String will return status as string based on the int value
func (cs ClientStatus) String() string {
	switch cs {
	case Available:
		return "Available"
	case Unavailable:
		return "Unavailable"
	default:
		return fmt.Sprintf("Invalid status %d", cs)
	}
}

// CallData defines the struct holding SC call data parameters
type CallData struct {
	Type      byte
	Function  string
	GasLimit  uint64
	Arguments []string
}

// ProxySCCompleteCallData defines the struct holding Proxy SC complete call data
type ProxySCCompleteCallData struct {
	RawCallData []byte
	From        common.Address
	To          core.AddressHandler
	Token       string
	Amount      *big.Int
	Nonce       uint64
}

// String returns the human-readable string version of the call data
func (callData ProxySCCompleteCallData) String() string {
	toString := "<nil>"
	var err error
	if !check.IfNil(callData.To) {
		toString, err = callData.To.AddressAsBech32String()
		if err != nil {
			toString = "<err>"
		}
	}
	amountString := "<nil>"
	if callData.Amount != nil {
		amountString = callData.Amount.String()
	}

	return fmt.Sprintf("Eth address: %s, MvX address: %s, token: %s, amount: %s, nonce: %d, raw call data: %x",
		callData.From.String(),
		toString,
		callData.Token,
		amountString,
		callData.Nonce,
		callData.RawCallData,
	)
}
