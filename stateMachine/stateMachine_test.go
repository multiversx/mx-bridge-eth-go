package stateMachine_test

import (
	"context"
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-eth-bridge/core"
	"github.com/ElrondNetwork/elrond-eth-bridge/stateMachine"
	"github.com/ElrondNetwork/elrond-eth-bridge/testsCommon"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/stretchr/testify/assert"
)

func createMockArgs() stateMachine.ArgsStateMachine {
	return stateMachine.ArgsStateMachine{
		Steps: core.MachineStates{
			"mock": &testsCommon.StepMock{
				ExecuteCalled: func(ctx context.Context) core.StepIdentifier {
					return "mock"
				},
			},
		},
		StartStateIdentifier: "mock",
		Log:                  logger.GetOrCreate("test"),
		StatusHandler:        testsCommon.NewStatusHandlerMock("mock"),
	}
}

func TestNewStateMachine(t *testing.T) {
	t.Parallel()

	t.Run("nil steps map", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.Steps = nil
		sm, err := stateMachine.NewStateMachine(args)

		assert.Nil(t, sm)
		assert.Equal(t, stateMachine.ErrNilStepsMap, err)
	})
	t.Run("nil step", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.Steps["mock"] = nil
		sm, err := stateMachine.NewStateMachine(args)

		assert.Nil(t, sm)
		assert.True(t, errors.Is(err, stateMachine.ErrNilStep))
	})
	t.Run("nil logger", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.Log = nil
		sm, err := stateMachine.NewStateMachine(args)

		assert.Nil(t, sm)
		assert.Equal(t, stateMachine.ErrNilLogger, err)
	})
	t.Run("invalid first step", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.StartStateIdentifier = "not found"
		sm, err := stateMachine.NewStateMachine(args)

		assert.Nil(t, sm)
		assert.True(t, errors.Is(err, stateMachine.ErrStepNotFound))
	})
	t.Run("nil status handler", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.StatusHandler = nil
		sm, err := stateMachine.NewStateMachine(args)

		assert.Nil(t, sm)
		assert.True(t, errors.Is(err, stateMachine.ErrNilStatusHandler))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		sm, err := stateMachine.NewStateMachine(args)

		assert.NotNil(t, sm)
		assert.Nil(t, err)
	})
}

func TestExecute(t *testing.T) {
	t.Parallel()

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		providedIdentifier0 := core.StepIdentifier("step0")
		providedIdentifier1 := core.StepIdentifier("step1")
		providedIdentifier2 := core.StepIdentifier("step2")
		args.Steps = map[core.StepIdentifier]core.Step{
			providedIdentifier0: &testsCommon.StepMock{
				ExecuteCalled: func(ctx context.Context) core.StepIdentifier {
					return providedIdentifier1
				},
				IdentifierCalled: func() core.StepIdentifier {
					return providedIdentifier0
				},
			},
			providedIdentifier1: &testsCommon.StepMock{
				ExecuteCalled: func(ctx context.Context) core.StepIdentifier {
					return providedIdentifier2
				},
				IdentifierCalled: func() core.StepIdentifier {
					return providedIdentifier1
				},
			},
			providedIdentifier2: &testsCommon.StepMock{
				ExecuteCalled: func(ctx context.Context) core.StepIdentifier {
					return providedIdentifier2
				},
				IdentifierCalled: func() core.StepIdentifier {
					return providedIdentifier2
				},
			},
		}
		args.StartStateIdentifier = providedIdentifier0
		sm, err := stateMachine.NewStateMachine(args)
		assert.NotNil(t, sm)
		assert.Nil(t, err)

		err = sm.Execute(context.Background())
		assert.Nil(t, err)
		assert.Equal(t, providedIdentifier1, sm.GetCurrentStepIdentifier())

		err = sm.Execute(context.Background())
		assert.Nil(t, err)
		assert.Equal(t, providedIdentifier2, sm.GetCurrentStepIdentifier())

		err = sm.Execute(context.Background())
		assert.Nil(t, err)
		assert.Equal(t, providedIdentifier2, sm.GetCurrentStepIdentifier())
	})
}
