package mock

import (
	"github.com/ElrondNetwork/elrond-eth-bridge/relay"
)

// StepMock -
type StepMock struct {
	ExecuteCalled func() relay.StepIdentifier
}

// Execute -
func (sm *StepMock) Execute() relay.StepIdentifier {
	return sm.ExecuteCalled()
}

// IsInterfaceNil -
func (sm *StepMock) IsInterfaceNil() bool {
	return sm == nil
}
