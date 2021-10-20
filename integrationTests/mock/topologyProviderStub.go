package mock

// TopologyProviderStub -
type TopologyProviderStub struct {
	PeerCountCalled    func() int
	AmITheLeaderCalled func() bool
	CleanCalled        func()
}

// PeerCount -
func (tps *TopologyProviderStub) PeerCount() int {
	if tps.PeerCountCalled != nil {
		return tps.PeerCountCalled()
	}

	return 0
}

// AmITheLeader -
func (tps *TopologyProviderStub) AmITheLeader() bool {
	if tps.AmITheLeaderCalled != nil {
		return tps.AmITheLeaderCalled()
	}

	return false
}

// Clean -
func (tps *TopologyProviderStub) Clean() {
	if tps.CleanCalled != nil {
		tps.CleanCalled()
	}
}
