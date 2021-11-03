package testsCommon

import "context"

// QuorumProviderStub -
type QuorumProviderStub struct {
	GetQuorumCalled func(ctx context.Context) (uint, error)

	GetQuorumError error
}

// GetQuorum -
func (s *QuorumProviderStub) GetQuorum(ctx context.Context) (uint, error) {
	if s.GetQuorumCalled != nil {
		return s.GetQuorumCalled(ctx)
	}
	return 0, s.GetQuorumError
}

// IsInterfaceNil returns true if there is no value under the interface
func (s *QuorumProviderStub) IsInterfaceNil() bool {
	return s == nil
}
