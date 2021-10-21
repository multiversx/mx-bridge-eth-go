package mock

import (
	"context"
	"errors"
)

// QuorumProviderStub -
type QuorumProviderStub struct {
	GetQuorumCalled func(ctx context.Context) (uint, error)
}

// GetQuorum -
func (qps *QuorumProviderStub) GetQuorum(ctx context.Context) (uint, error) {
	if qps.GetQuorumCalled != nil {
		return qps.GetQuorumCalled(ctx)
	}

	return 0, errors.New("not implemented")
}
