package mock

type TopologyProviderStub struct {
	AmITheLeaderCalled func() bool
	PeerCountCalled    func() int
	CleanCalled        func()
}

func (s *TopologyProviderStub) AmITheLeader() bool {
	if s.AmITheLeaderCalled != nil {
		return s.AmITheLeaderCalled()
	}
	return false
}

func (s *TopologyProviderStub) PeerCount() int {
	if s.PeerCountCalled != nil {
		return s.PeerCountCalled()
	}
	return 0
}

func (s *TopologyProviderStub) Clean() {
	if s.CleanCalled != nil {
		s.CleanCalled()
	}
}
