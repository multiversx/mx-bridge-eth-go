package mock

import "context"

type QuorumProviderStub struct {
	GetQuorumCalled func(ctx context.Context) (uint, error)

	GetQuorumError error
}

func (s *QuorumProviderStub) GetQuorum(ctx context.Context) (uint, error) {
	if s.GetQuorumCalled != nil {
		return s.GetQuorumCalled(ctx)
	}
	return 0, s.GetQuorumError
}
