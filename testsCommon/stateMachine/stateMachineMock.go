package stateMachine

import (
	"context"
	"fmt"

	"github.com/ElrondNetwork/elrond-eth-bridge/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
)

// StateMachineMock -
type StateMachineMock struct {
	Steps         core.MachineStates
	ExecutedSteps []core.StepIdentifier
	InitialStep   core.StepIdentifier
	CurrentStep   core.Step
}

// NewStateMachineMock -
func NewStateMachineMock(steps core.MachineStates, initialStep core.StepIdentifier) *StateMachineMock {
	return &StateMachineMock{
		Steps:         steps,
		ExecutedSteps: make([]core.StepIdentifier, 0),
		InitialStep:   initialStep,
	}
}

// Initialize -
func (smm *StateMachineMock) Initialize() error {
	var err error
	smm.CurrentStep, err = smm.getNextStep(smm.InitialStep)

	return err
}

func (smm *StateMachineMock) getNextStep(identifier core.StepIdentifier) (core.Step, error) {
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

	fmt.Printf("executing step %s...\n", smm.CurrentStep.Identifier())
	nextStepIdentifier, err := smm.CurrentStep.Execute(context.Background())
	if err != nil {
		return err
	}

	smm.ExecutedSteps = append(smm.ExecutedSteps, smm.CurrentStep.Identifier())

	nextStep, err := smm.getNextStep(nextStepIdentifier)
	if err != nil {
		return err
	}

	smm.CurrentStep = nextStep

	return nil
}
