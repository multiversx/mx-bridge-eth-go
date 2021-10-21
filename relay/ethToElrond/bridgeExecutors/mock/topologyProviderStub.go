package mock

import (
	"fmt"
	"runtime"
	"sync"
)

var fullPathTopologyProviderStub = "github.com/ElrondNetwork/elrond-eth-bridge/relay/ethToElrond/bridgeExecutors/mock.(*TopologyProviderStub)."

type TopologyProviderStub struct {
	functionCalledCounter map[string]int
	mutTopology           sync.RWMutex

	AmITheLeaderCalled func() bool
	PeerCountCalled    func() int
	CleanCalled        func()
}

// NewTopologyProviderStub creates a new TopologyProviderStub instance
func NewTopologyProviderStub() *TopologyProviderStub {
	return &TopologyProviderStub{
		functionCalledCounter: make(map[string]int),
	}
}

func (s *TopologyProviderStub) AmITheLeader() bool {
	s.incrementFunctionCounter()
	if s.AmITheLeaderCalled != nil {
		return s.AmITheLeaderCalled()
	}
	return false
}

func (s *TopologyProviderStub) PeerCount() int {
	s.incrementFunctionCounter()
	if s.PeerCountCalled != nil {
		return s.PeerCountCalled()
	}
	return 0
}

func (s *TopologyProviderStub) Clean() {
	s.incrementFunctionCounter()
	if s.CleanCalled != nil {
		s.CleanCalled()
	}
}

// -------- helper functions

// incrementFunctionCounter increments the counter for the function that called it
func (s *TopologyProviderStub) incrementFunctionCounter() {
	s.mutTopology.Lock()
	defer s.mutTopology.Unlock()

	pc, _, _, _ := runtime.Caller(1)
	fmt.Printf("BridgeExecutorMock: called %s\n", runtime.FuncForPC(pc).Name())
	s.functionCalledCounter[runtime.FuncForPC(pc).Name()]++
}

// GetFunctionCounter returns the called counter of a given function
func (s *TopologyProviderStub) GetFunctionCounter(function string) int {
	s.mutTopology.Lock()
	defer s.mutTopology.Unlock()

	return s.functionCalledCounter[fullPathTopologyProviderStub+function]
}
