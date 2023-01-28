package stateMachine

import (
	"context"
	"fmt"

	"github.com/multiversx/mx-bridge-eth-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	logger "github.com/multiversx/mx-chain-logger-go"
)

// ArgsStateMachine represents the state machine arguments
type ArgsStateMachine struct {
	StateMachineName     string
	Steps                core.MachineStates
	StartStateIdentifier core.StepIdentifier
	Log                  logger.Logger
	StatusHandler        core.StatusHandler
}

type stateMachine struct {
	stateMachineName string
	steps            core.MachineStates
	currentStep      core.Step
	log              logger.Logger
	statusHandler    core.StatusHandler
}

// NewStateMachine creates a state machine able to execute all provided steps
func NewStateMachine(args ArgsStateMachine) (*stateMachine, error) {
	err := checkArgs(args)
	if err != nil {
		return nil, err
	}

	sm := &stateMachine{
		stateMachineName: args.StateMachineName,
		steps:            args.Steps,
		log:              args.Log,
		statusHandler:    args.StatusHandler,
	}
	sm.currentStep, err = sm.getNextStep(args.StartStateIdentifier)
	if err != nil {
		return nil, err
	}

	return sm, nil
}

func checkArgs(args ArgsStateMachine) error {
	if args.Steps == nil {
		return ErrNilStepsMap
	}
	for identifier, step := range args.Steps {
		if check.IfNil(step) {
			return fmt.Errorf("%w for identifier %s", ErrNilStep, identifier)
		}
	}
	if check.IfNil(args.Log) {
		return ErrNilLogger
	}
	if check.IfNil(args.StatusHandler) {
		return ErrNilStatusHandler
	}

	return nil
}

// Execute will execute one step
func (sm *stateMachine) Execute(ctx context.Context) error {
	return sm.executeStep(ctx)
}

func (sm *stateMachine) executeStep(ctx context.Context) error {
	sm.log.Debug(fmt.Sprintf("%s: executing step", sm.stateMachineName),
		"step", sm.currentStep.Identifier())
	sm.statusHandler.SetStringMetric(core.MetricCurrentStateMachineStep, string(sm.currentStep.Identifier()))
	nextStepIdentifier := sm.currentStep.Execute(ctx)

	currentStep, err := sm.getNextStep(nextStepIdentifier)
	sm.currentStep = currentStep

	return err
}

func (sm *stateMachine) getNextStep(identifier core.StepIdentifier) (core.Step, error) {
	nextStep, ok := sm.steps[identifier]
	if !ok {
		return nil, fmt.Errorf("%w for identifier '%s'", ErrStepNotFound, identifier)
	}

	return nextStep, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (sm *stateMachine) IsInterfaceNil() bool {
	return sm == nil
}
