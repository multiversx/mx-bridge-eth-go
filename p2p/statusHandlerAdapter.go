package p2p

import (
	"context"
	"strings"

	"github.com/ElrondNetwork/elrond-eth-bridge/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
)

// ArgsStatusHandlerAdapter is the DTO used to construct a new instance of type statusHandlerAdapter
type ArgsStatusHandlerAdapter struct {
	StatusHandler core.StatusHandler
	Messenger     NetMessenger
}

type statusHandlerAdapter struct {
	core.StatusHandler
	messenger NetMessenger
}

// NewStatusHandlerAdapter creates a new instance of statusHandlerAdapter able to gather p2p status metrics
func NewStatusHandlerAdapter(args ArgsStatusHandlerAdapter) (*statusHandlerAdapter, error) {
	if check.IfNil(args.StatusHandler) {
		return nil, ErrNilStatusHandler
	}
	if check.IfNil(args.Messenger) {
		return nil, ErrNilMessenger
	}

	return &statusHandlerAdapter{
		StatusHandler: args.StatusHandler,
		messenger:     args.Messenger,
	}, nil
}

// Execute will update the metrics according to the network messenger's current state
func (adapter *statusHandlerAdapter) Execute(_ context.Context) error {
	hostAddresses := adapter.messenger.Addresses()
	adapter.SetStringMetric(core.MetricRelayerP2PAddresses, strings.Join(hostAddresses, " "))

	connectedAddresses := adapter.messenger.ConnectedAddresses()
	adapter.SetStringMetric(core.MetricConnectedP2PAddresses, strings.Join(connectedAddresses, " "))

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (adapter *statusHandlerAdapter) IsInterfaceNil() bool {
	return adapter == nil
}
