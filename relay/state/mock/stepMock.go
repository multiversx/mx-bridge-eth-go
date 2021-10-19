package mock

import (
	"context"

	"github.com/ElrondNetwork/elrond-eth-bridge/relay"
)

// StepMock -
type StepMock struct {
	ExecuteCalled    func(ctx context.Context) relay.StepIdentifier
	IdentifierCalled func() relay.StepIdentifier
}

// Execute -
func (sm *StepMock) Execute(ctx context.Context) relay.StepIdentifier {
	return sm.ExecuteCalled(ctx)
}

// Identifier -
func (sm *StepMock) Identifier() relay.StepIdentifier {
	if sm.IdentifierCalled != nil {
		return sm.IdentifierCalled()
	}

	return ""
}

// IsInterfaceNil -
func (sm *StepMock) IsInterfaceNil() bool {
	return sm == nil
}
