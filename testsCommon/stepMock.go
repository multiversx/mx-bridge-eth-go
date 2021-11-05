package testsCommon

import (
	"context"

	"github.com/ElrondNetwork/elrond-eth-bridge/core"
)

// StepMock -
type StepMock struct {
	ExecuteCalled    func(ctx context.Context) (core.StepIdentifier, error)
	IdentifierCalled func() core.StepIdentifier
}

// Execute -
func (sm *StepMock) Execute(ctx context.Context) (core.StepIdentifier, error) {
	return sm.ExecuteCalled(ctx)
}

// Identifier -
func (sm *StepMock) Identifier() core.StepIdentifier {
	if sm.IdentifierCalled != nil {
		return sm.IdentifierCalled()
	}

	return ""
}

// IsInterfaceNil -
func (sm *StepMock) IsInterfaceNil() bool {
	return sm == nil
}
