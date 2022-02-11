package topology

import (
	"bytes"
	"time"

	"github.com/ElrondNetwork/elrond-eth-bridge/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	logger "github.com/ElrondNetwork/elrond-go-logger"
)

// ArgsTopologyHandler is the DTO used in the NewTopologyHandler constructor function
type ArgsTopologyHandler struct {
	PublicKeysProvider PublicKeysProvider
	Timer              core.Timer
	IntervalForLeader  time.Duration
	AddressBytes       []byte
	Log                logger.Logger
	AddressConverter   core.AddressConverter
}

// topologyHandler implements topologyProvider for a specific relay
type topologyHandler struct {
	publicKeysProvider PublicKeysProvider
	timer              core.Timer
	intervalForLeader  time.Duration
	addressBytes       []byte
	selector           *hashRandomSelector
	log                logger.Logger
	addressConverter   core.AddressConverter
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
		intervalForLeader:  args.IntervalForLeader,
		addressBytes:       args.AddressBytes,
		selector:           &hashRandomSelector{},
		log:                args.Log,
		addressConverter:   args.AddressConverter,
	}, nil
}

// MyTurnAsLeader returns true if the current relay is leader
func (t *topologyHandler) MyTurnAsLeader() bool {
	sortedPublicKeys := t.publicKeysProvider.SortedPublicKeys()

	if len(sortedPublicKeys) == 0 {
		t.log.Warn("topology handler: can not compute my turn as leader as the list is empty")
		return false
	} else {
		numberOfPeers := int64(len(sortedPublicKeys))

		seed := uint64(t.timer.NowUnix() / int64(t.intervalForLeader.Seconds()))
		index := t.selector.randomInt(seed, uint64(numberOfPeers))

		leaderAddress := sortedPublicKeys[index]
		isLeader := bytes.Equal(leaderAddress, t.addressBytes)
		msg := "topology handler"
		if isLeader {
			msg += " (my turn)"
		}

		t.log.Debug(msg,
			"leader", t.addressConverter.ToBech32String(leaderAddress),
			"index", index,
			"self address", t.addressConverter.ToBech32String(t.addressBytes))

		return isLeader
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
	if int64(args.IntervalForLeader.Seconds()) <= 0 {
		return errInvalidIntervalForLeader
	}
	if len(args.AddressBytes) == 0 {
		return errEmptyAddress
	}
	if check.IfNil(args.Log) {
		return errNilLogger
	}
	if check.IfNil(args.AddressConverter) {
		return errNilAddressConverter
	}

	return nil
}
