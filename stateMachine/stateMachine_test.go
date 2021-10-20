package stateMachine_test

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-eth-bridge/core"
	"github.com/ElrondNetwork/elrond-eth-bridge/stateMachine"
	"github.com/ElrondNetwork/elrond-eth-bridge/stateMachine/mock"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/stretchr/testify/assert"
)

func createMockArgs() stateMachine.ArgsStateMachine {
	return stateMachine.ArgsStateMachine{
		Steps: core.MachineStates{
			"mock": &mock.StepMock{
				ExecuteCalled: func(ctx context.Context) (core.StepIdentifier, error) {
					return "mock", nil
				},
			},
		},
		StartStateIdentifier: "mock",
		DurationBetweenSteps: time.Millisecond,
		Log:                  logger.GetOrCreate("test"),
	}
}

func TestNewStateMachine(t *testing.T) {
	t.Parallel()

	t.Run("nil steps map", func(t *testing.T) {
		args := createMockArgs()
		args.Steps = nil
		sm, err := stateMachine.NewStateMachine(args)

		assert.Nil(t, sm)
		assert.Equal(t, stateMachine.ErrNilStepsMap, err)
	})
	t.Run("nil step", func(t *testing.T) {
		args := createMockArgs()
		args.Steps["mock"] = nil
		sm, err := stateMachine.NewStateMachine(args)

		assert.Nil(t, sm)
		assert.True(t, errors.Is(err, stateMachine.ErrNilStep))
	})
	t.Run("nil logger", func(t *testing.T) {
		args := createMockArgs()
		args.Log = nil
		sm, err := stateMachine.NewStateMachine(args)

		assert.Nil(t, sm)
		assert.Equal(t, stateMachine.ErrNilLogger, err)
	})
	t.Run("invalid first step", func(t *testing.T) {
		args := createMockArgs()
		args.StartStateIdentifier = "not found"
		sm, err := stateMachine.NewStateMachine(args)

		assert.Nil(t, sm)
		assert.True(t, errors.Is(err, stateMachine.ErrStepNotFound))
	})
	t.Run("should work", func(t *testing.T) {
		args := createMockArgs()
		sm, err := stateMachine.NewStateMachine(args)

		assert.NotNil(t, sm)
		assert.Nil(t, err)

		time.Sleep(time.Millisecond * 100)
		assert.True(t, sm.LoopStatus())

		_ = sm.Close()
	})
}

func TestStateMachine_CloseDoesNotCallNewExecute(t *testing.T) {
	t.Parallel()
	args := createMockArgs()
	numCalled := uint32(0)
	wg := &sync.WaitGroup{}
	wg.Add(1)

	args.Steps["mock"] = &mock.StepMock{
		ExecuteCalled: func(ctx context.Context) (core.StepIdentifier, error) {
			atomic.AddUint32(&numCalled, 1)

			wg.Wait()

			return "mock", nil
		},
	}

	sm, _ := stateMachine.NewStateMachine(args)
	time.Sleep(time.Millisecond * 100)

	_ = sm.Close()
	wg.Done()

	time.Sleep(time.Millisecond * 100)

	assert.Equal(t, uint32(1), atomic.LoadUint32(&numCalled))
	assert.False(t, sm.LoopStatus())
}

func TestStateMachine_StateMachineErrors(t *testing.T) {
	t.Parallel()
	args := createMockArgs()
	args.Steps["mock"] = &mock.StepMock{
		ExecuteCalled: func(ctx context.Context) (core.StepIdentifier, error) {
			return "not found", nil
		},
	}

	sm, _ := stateMachine.NewStateMachine(args)
	time.Sleep(time.Millisecond * 100)

	assert.False(t, sm.LoopStatus())
}

func TestStateMachine_StepErrorsShouldStopTheStateMachine(t *testing.T) {
	t.Parallel()
	args := createMockArgs()
	expectedErr := errors.New("expected error")
	args.Steps["mock"] = &mock.StepMock{
		ExecuteCalled: func(ctx context.Context) (core.StepIdentifier, error) {
			return "mock", expectedErr
		},
	}

	sm, _ := stateMachine.NewStateMachine(args)
	time.Sleep(time.Millisecond * 100)

	assert.False(t, sm.LoopStatus())
}
