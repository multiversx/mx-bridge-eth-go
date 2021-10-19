package state

import (
	"context"
	"fmt"
	"time"

	"github.com/ElrondNetwork/elrond-eth-bridge/relay"
	"github.com/ElrondNetwork/elrond-go-core/core/atomic"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	logger "github.com/ElrondNetwork/elrond-go-logger"
)

// ArgsStateMachine represents the state machine arguments
type ArgsStateMachine struct {
	Steps                relay.MachineStates
	StartStateIdentifier relay.StepIdentifier
	DurationBetweenSteps time.Duration
	Log                  logger.Logger
}

type stateMachine struct {
	steps                relay.MachineStates
	currentStep          relay.Step
	durationBetweenSteps time.Duration
	log                  logger.Logger
	cancel               func()
	loopStatus           *atomic.Flag
}

// NewStateMachine creates a state machine able to execute all provided steps
func NewStateMachine(args ArgsStateMachine) (*stateMachine, error) {
	err := checkArgs(args)
	if err != nil {
		return nil, err
	}

	sm := &stateMachine{
		steps:                args.Steps,
		durationBetweenSteps: args.DurationBetweenSteps,
		log:                  args.Log,
		loopStatus:           &atomic.Flag{},
	}
	sm.currentStep, err = sm.getNextStep(args.StartStateIdentifier)
	if err != nil {
		return nil, err
	}

	var ctx context.Context
	ctx, sm.cancel = context.WithCancel(context.Background())
	go sm.executeLoop(ctx)

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

	return nil
}

func (sm *stateMachine) executeLoop(ctx context.Context) {
	sm.loopStatus.Set()
	defer sm.loopStatus.Unset()

	for {
		select {
		case <-ctx.Done():
			sm.log.Debug("state machine main execute loop is closing...")
			return
		case <-time.After(sm.durationBetweenSteps):
			err := sm.executeStep()
			if err != nil {
				sm.log.Error("state machine error", "status", "state machine stopped", "error", err)
				return
			}
		}
	}
}

func (sm *stateMachine) executeStep() error {
	nextStepIdentifier := sm.currentStep.Execute()

	var err error
	sm.currentStep, err = sm.getNextStep(nextStepIdentifier)

	return err
}

func (sm *stateMachine) getNextStep(identifier relay.StepIdentifier) (relay.Step, error) {
	nextStep, ok := sm.steps[identifier]
	if !ok {
		return nil, fmt.Errorf("%w for identifier '%s'", ErrStepNotFound, identifier)
	}

	return nextStep, nil
}

// Close will close the state machine's main loop
func (sm *stateMachine) Close() error {
	sm.cancel()

	return nil
}
