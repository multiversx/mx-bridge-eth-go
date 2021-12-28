package topology

import (
	"bytes"
	"time"

	"github.com/ElrondNetwork/elrond-eth-bridge/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
)

// ArgsTopologyHandler is the DTO used in the NewTopologyHandler constructor function
type ArgsTopologyHandler struct {
	PublicKeysProvider PublicKeysProvider
	Timer              core.Timer
	StepDuration       time.Duration
	AddressBytes       []byte
}

// topologyHandler implements topologyProvider for a specific relay
type topologyHandler struct {
	publicKeysProvider PublicKeysProvider
	timer              core.Timer
	stepDuration       time.Duration
	addressBytes       []byte
}

// NewTopologyHandler creates a new topologyHandler instance
func NewTopologyHandler(args ArgsTopologyHandler) (*topologyHandler, error) {
	err := checkArgs(args)
	if err != nil {
		return nil, err
	}

	return &topologyHandler{
		publicKeysProvider: args.PublicKeysProvider,
		timer:              args.Timer,
		stepDuration:       args.StepDuration,
		addressBytes:       args.AddressBytes,
	}, nil
}

// MyTurnAsLeader returns true if the current relay is leader
func (t *topologyHandler) MyTurnAsLeader() bool {
	sortedPublicKeys := t.publicKeysProvider.SortedPublicKeys()

	if len(sortedPublicKeys) == 0 {
		return false
	} else {
		numberOfPeers := int64(len(sortedPublicKeys))
		index := (t.timer.NowUnix() / int64(t.stepDuration.Seconds())) % numberOfPeers

		return bytes.Equal(sortedPublicKeys[index], t.addressBytes)
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (t *topologyHandler) IsInterfaceNil() bool {
	return t == nil
}

func checkArgs(args ArgsTopologyHandler) error {
	if args.PublicKeysProvider == nil {
		return errNilPublicKeysProvider
	}
	if check.IfNil(args.Timer) {
		return errNilTimer
	}
	if int64(args.StepDuration.Seconds()) <= 0 {
		return errInvalidStepDuration
	}
	if len(args.AddressBytes) == 0 {
		return errEmptyAddress
	}

	return nil
}
