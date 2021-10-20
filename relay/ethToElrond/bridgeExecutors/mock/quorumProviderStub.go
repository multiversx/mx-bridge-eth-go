package mock

import "context"

type QuorumProviderStub struct {
	quorum uint
}

func (s *QuorumProviderStub) GetQuorum(_ context.Context) (uint, error) {
	return s.quorum, nil
}
