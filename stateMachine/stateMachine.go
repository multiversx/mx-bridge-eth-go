package stateMachine

import (
	"context"
	"fmt"
	"time"

	"github.com/ElrondNetwork/elrond-eth-bridge/core"
	"github.com/ElrondNetwork/elrond-go-core/core/atomic"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	logger "github.com/ElrondNetwork/elrond-go-logger"
)

const stopped = "stopped"

// ArgsStateMachine represents the state machine arguments
type ArgsStateMachine struct {
	StateMachineName     string
	Steps                core.MachineStates
	StartStateIdentifier core.StepIdentifier
	DurationBetweenSteps time.Duration
	Log                  logger.Logger
	Timer                core.Timer
	StatusHandler        core.StatusHandler
}

type stateMachine struct {
	stateMachineName     string
	steps                core.MachineStates
	currentStep          core.Step
	durationBetweenSteps time.Duration
	log                  logger.Logger
	cancel               func()
	loopStatus           *atomic.Flag
	timer                core.Timer
	statusHandler        core.StatusHandler
}

// NewStateMachine creates a state machine able to execute all provided steps
func NewStateMachine(args ArgsStateMachine) (*stateMachine, error) {
	err := checkArgs(args)
	if err != nil {
		return nil, err
	}

	sm := &stateMachine{
		stateMachineName:     args.StateMachineName,
		steps:                args.Steps,
		durationBetweenSteps: args.DurationBetweenSteps,
		log:                  args.Log,
		loopStatus:           &atomic.Flag{},
		timer:                args.Timer,
		statusHandler:        args.StatusHandler,
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
	if check.IfNil(args.Timer) {
		return ErrNilTimer
	}
	if check.IfNil(args.StatusHandler) {
		return ErrNilStatusHandler
	}

	return nil
}

func (sm *stateMachine) executeLoop(ctx context.Context) {
	sm.loopStatus.Set()
	defer sm.loopStatus.Unset()
	defer sm.statusHandler.SetStringMetric(core.MetricCurrentStateMachineStep, stopped)

	for {
		select {
		case <-ctx.Done():
			sm.log.Debug(fmt.Sprintf("%s: state machine main execute loop is closing...", sm.stateMachineName))
			return
		case <-sm.timer.After(sm.durationBetweenSteps):
			err := sm.executeStep(ctx)
			if err != nil {
				sm.log.Error(fmt.Sprintf("%s: state machine stopped", sm.stateMachineName),
					"error", err)
				return
			}
		}
	}
}

func (sm *stateMachine) executeStep(ctx context.Context) error {
	sm.log.Trace(fmt.Sprintf("%s: executing step", sm.stateMachineName),
		"step", sm.currentStep.Identifier())
	sm.statusHandler.SetStringMetric(core.MetricCurrentStateMachineStep, string(sm.currentStep.Identifier()))
	nextStepIdentifier, err := sm.currentStep.Execute(ctx)
	if err != nil {
		return err
	}

	sm.currentStep, err = sm.getNextStep(nextStepIdentifier)

	return err
}

func (sm *stateMachine) getNextStep(identifier core.StepIdentifier) (core.Step, error) {
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
