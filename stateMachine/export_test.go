package stateMachine

import "github.com/multiversx/mx-bridge-eth-go/core"

// GetCurrentStep -
func (sm *stateMachine) GetCurrentStepIdentifier() core.StepIdentifier {
	return sm.currentStep.Identifier()
}
