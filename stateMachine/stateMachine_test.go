package stateMachine

import (
	"context"
	"errors"
	"testing"

	"github.com/multiversx/mx-bridge-eth-go/core"
	"github.com/multiversx/mx-bridge-eth-go/testsCommon"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/stretchr/testify/assert"
)

func createMockArgs() ArgsStateMachine {
	return ArgsStateMachine{
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
		sm, err := NewStateMachine(args)

		assert.Nil(t, sm)
		assert.Equal(t, ErrNilStepsMap, err)
	})
	t.Run("nil step", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.Steps["mock"] = nil
		sm, err := NewStateMachine(args)

		assert.Nil(t, sm)
		assert.True(t, errors.Is(err, ErrNilStep))
	})
	t.Run("nil logger", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.Log = nil
		sm, err := NewStateMachine(args)

		assert.Nil(t, sm)
		assert.Equal(t, ErrNilLogger, err)
	})
	t.Run("invalid first step", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.StartStateIdentifier = "not found"
		sm, err := NewStateMachine(args)

		assert.Nil(t, sm)
		assert.True(t, errors.Is(err, ErrStepNotFound))
	})
	t.Run("nil status handler", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.StatusHandler = nil
		sm, err := NewStateMachine(args)

		assert.Nil(t, sm)
		assert.True(t, errors.Is(err, ErrNilStatusHandler))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		sm, err := NewStateMachine(args)

		assert.NotNil(t, sm)
		assert.Nil(t, err)
	})
}

func TestStateMachine_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var instance *stateMachine
	assert.True(t, instance.IsInterfaceNil())

	instance = &stateMachine{}
	assert.False(t, instance.IsInterfaceNil())
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
		sm, err := NewStateMachine(args)
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
