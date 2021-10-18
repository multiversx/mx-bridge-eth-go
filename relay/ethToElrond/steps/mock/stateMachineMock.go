package mock

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-eth-bridge/relay"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
)

// StateMachineMock -
type StateMachineMock struct {
	Steps                 relay.MachineStates
	ExecutedSteps         []relay.StepIdentifier
	InitialStep           relay.StepIdentifier
	CurrentStep           relay.Step
	CurrentStepIdentifier relay.StepIdentifier
}

// NewStateMachineMock -
func NewStateMachineMock(steps relay.MachineStates, initialStep relay.StepIdentifier) *StateMachineMock {
	return &StateMachineMock{
		Steps:         steps,
		ExecutedSteps: make([]relay.StepIdentifier, 0),
		InitialStep:   initialStep,
	}
}

// Initialize -
func (smm *StateMachineMock) Initialize() error {
	var err error
	smm.CurrentStep, err = smm.getNextStep(smm.InitialStep)
	smm.CurrentStepIdentifier = smm.InitialStep

	return err
}

func (smm *StateMachineMock) getNextStep(identifier relay.StepIdentifier) (relay.Step, error) {
	nextStep, ok := smm.Steps[identifier]
	if !ok {
		return nil, fmt.Errorf("step not found for identifier '%s'", identifier)
	}

	return nextStep, nil
}

// ExecuteOneStep -
func (smm *StateMachineMock) ExecuteOneStep() error {
	if check.IfNil(smm.CurrentStep) {
		return fmt.Errorf("current step is nil. Call Initialize() first")
	}

	nextStepIdentifier := smm.CurrentStep.Execute()
	smm.ExecutedSteps = append(smm.ExecutedSteps, smm.CurrentStepIdentifier)

	nextStep, err := smm.getNextStep(nextStepIdentifier)
	if err != nil {
		return err
	}

	smm.CurrentStepIdentifier = nextStepIdentifier
	smm.CurrentStep = nextStep

	return nil
}
