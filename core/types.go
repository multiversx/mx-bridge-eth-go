package core

import (
	"context"
	"time"
)

// StepIdentifier defines a step name
type StepIdentifier string

// MachineStates defines all available steps for a state machine to run
type MachineStates map[StepIdentifier]Step

// Step defines a state machine step
type Step interface {
	Execute(ctx context.Context) (StepIdentifier, error)
	Identifier() StepIdentifier
	IsInterfaceNil() bool
}

// EthGasPriceSelector defines the ethereum gas price selector
type EthGasPriceSelector string

// Timer defines operations related to time
type Timer interface {
	After(d time.Duration) <-chan time.Time
	NowUnix() int64
	Start()
	Close() error
	IsInterfaceNil() bool
}

// BroadcastClient defines a broadcast client that will get notified by te broadcaster
// when new messages arrive. It also should be able to respond with any stored messages it might
// have.
type BroadcastClient interface {
	ProcessNewMessage(msg *SignedMessage, ethMsg *EthereumSignature)
	AllStoredSignatures() []*SignedMessage
	IsInterfaceNil() bool
}
