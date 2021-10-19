package mock

import (
	"github.com/ElrondNetwork/elrond-eth-bridge/relay"
)

// StepMock -
type StepMock struct {
	ExecuteCalled    func() relay.StepIdentifier
	IdentifierCalled func() relay.StepIdentifier
}

// Execute -
func (sm *StepMock) Execute() relay.StepIdentifier {
	return sm.ExecuteCalled()
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
