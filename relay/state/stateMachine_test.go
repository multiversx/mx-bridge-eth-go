package state_test

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-eth-bridge/relay"
	"github.com/ElrondNetwork/elrond-eth-bridge/relay/state"
	"github.com/ElrondNetwork/elrond-eth-bridge/relay/state/mock"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/stretchr/testify/assert"
)

func createMockArgs() state.ArgsStateMachine {
	return state.ArgsStateMachine{
		Steps: relay.MachineStates{
			"mock": &mock.StepMock{
				ExecuteCalled: func() relay.StepIdentifier {
					return "mock"
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
		sm, err := state.NewStateMachine(args)

		assert.Nil(t, sm)
		assert.Equal(t, state.ErrNilStepsMap, err)
	})
	t.Run("nil step", func(t *testing.T) {
		args := createMockArgs()
		args.Steps["mock"] = nil
		sm, err := state.NewStateMachine(args)

		assert.Nil(t, sm)
		assert.True(t, errors.Is(err, state.ErrNilStep))
	})
	t.Run("nil logger", func(t *testing.T) {
		args := createMockArgs()
		args.Log = nil
		sm, err := state.NewStateMachine(args)

		assert.Nil(t, sm)
		assert.Equal(t, state.ErrNilLogger, err)
	})
	t.Run("invalid first step", func(t *testing.T) {
		args := createMockArgs()
		args.StartStateIdentifier = "not found"
		sm, err := state.NewStateMachine(args)

		assert.Nil(t, sm)
		assert.True(t, errors.Is(err, state.ErrStepNotFound))
	})
	t.Run("should work", func(t *testing.T) {
		args := createMockArgs()
		sm, err := state.NewStateMachine(args)

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
		ExecuteCalled: func() relay.StepIdentifier {
			atomic.AddUint32(&numCalled, 1)

			wg.Wait()

			return "mock"
		},
	}

	sm, _ := state.NewStateMachine(args)
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
		ExecuteCalled: func() relay.StepIdentifier {
			return "not found"
		},
	}

	sm, _ := state.NewStateMachine(args)
	time.Sleep(time.Millisecond * 100)

	assert.False(t, sm.LoopStatus())
}
