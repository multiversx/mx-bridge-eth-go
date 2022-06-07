package stateMachine

import "github.com/ElrondNetwork/elrond-eth-bridge/core"

// GetCurrentStep -
func (sm *stateMachine) GetCurrentStepIdentifier() core.StepIdentifier {
	return sm.currentStep.Identifier()
}
