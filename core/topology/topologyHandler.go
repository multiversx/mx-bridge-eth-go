package topology

import (
    "bytes"
    "time"

    "github.com/ElrondNetwork/elrond-eth-bridge/core"
)

// ArgsTopologyHandler is the DTO used in the NewTopologyHandler constructor function
type ArgsTopologyHandler struct {
    SortedPublicKeys [][]byte
    Timer            core.Timer
    StepDuration     time.Duration
    Address          []byte
}

// topologyHandler implements topologyProvider for a specific relay
type topologyHandler struct {
    sortedPublicKeys [][]byte
    timer            core.Timer
    stepDuration     time.Duration
    address          []byte
}

// NewTopologyHandler creates a new topologyHandler instance
func NewTopologyHandler(args ArgsTopologyHandler) *topologyHandler {
    return &topologyHandler{
        sortedPublicKeys: args.SortedPublicKeys,
        timer:            args.Timer,
        stepDuration:     args.StepDuration,
        address:          args.Address,
    }
}

// MyTurnAsLeader returns true if the current relay is leader
func (t *topologyHandler) MyTurnAsLeader() bool {
    if len(t.sortedPublicKeys) == 0 {
        return false
    } else {
        numberOfPeers := int64(len(t.sortedPublicKeys))
        index := (t.timer.NowUnix() / int64(t.stepDuration.Seconds())) % numberOfPeers

        return bytes.Equal(t.sortedPublicKeys[index], t.address)
    }
}

// IsInterfaceNil returns true if there is no value under the interface
func (t *topologyHandler) IsInterfaceNil() bool {
    return t == nil
}
