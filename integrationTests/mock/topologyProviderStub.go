package mock

// TopologyProviderStub -
type TopologyProviderStub struct {
	AmITheLeaderCalled func() bool
	CleanCalled        func()
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
