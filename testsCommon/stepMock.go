package testsCommon

import (
	"context"

	"github.com/multiversx/mx-bridge-eth-go/core"
)

// StepMock -
type StepMock struct {
	ExecuteCalled    func(ctx context.Context) core.StepIdentifier
	IdentifierCalled func() core.StepIdentifier
}

// Execute -
func (sm *StepMock) Execute(ctx context.Context) core.StepIdentifier {
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
