package mock

type TopologyProviderStub struct {
	amITheLeader bool
	peerCount    int
	cleaned      bool
}

func (s *TopologyProviderStub) AmITheLeader() bool {
	return s.amITheLeader
}

func (s *TopologyProviderStub) PeerCount() int {
	return s.peerCount
}

func (s *TopologyProviderStub) Clean() {
	s.cleaned = true
}
