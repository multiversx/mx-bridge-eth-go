package topology

import (
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-eth-bridge/testsCommon"
	"github.com/stretchr/testify/assert"
)

var duration = time.Second

func TestNewTopologyHandler(t *testing.T) {
	t.Parallel()

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsTopologyHandler()
		tph, err := NewTopologyHandler(args)

		assert.NotNil(t, tph) // pointer check
		assert.Nil(t, err)
		assert.False(t, tph.IsInterfaceNil()) // IsInterfaceNIl

		assert.True(t, args.PublicKeysProvider == tph.publicKeysProvider) // pointer testing
		assert.Equal(t, args.Timer, tph.timer)
		assert.Equal(t, args.StepDuration, tph.stepDuration)
		assert.Equal(t, args.AddressBytes, tph.addressBytes)
	})

	t.Run("nil PublicKeysProvider", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsTopologyHandler()
		args.PublicKeysProvider = nil
		tph, err := NewTopologyHandler(args)

		assert.Nil(t, tph)
		assert.Equal(t, errNilPublicKeysProvider, err)
	})

	t.Run("nil timer", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsTopologyHandler()
		args.Timer = nil
		tph, err := NewTopologyHandler(args)

		assert.Nil(t, tph)
		assert.Equal(t, errNilTimer, err)
	})

	t.Run("invalid step duration", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsTopologyHandler()
		args.StepDuration = time.Duration(12345)
		tph, err := NewTopologyHandler(args)

		assert.Nil(t, tph)
		assert.Equal(t, errInvalidStepDuration, err)
	})

	t.Run("nil address", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsTopologyHandler()
		args.AddressBytes = nil
		tph, err := NewTopologyHandler(args)

		assert.Nil(t, tph)
		assert.Equal(t, errNilAddress, err)
	})
}

func TestMyTurnAsLeader(t *testing.T) {
	t.Parallel()

	t.Run("not leader - SortedPublicKeys empty", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsTopologyHandler()
		args.PublicKeysProvider = &testsCommon.BroadcasterStub{
			SortedPublicKeysCalled: func() [][]byte {
				return make([][]byte, 0)
			},
		}
		tph, _ := NewTopologyHandler(args)

		assert.False(t, tph.MyTurnAsLeader())
	})

	t.Run("not leader", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsTopologyHandler()
		args.AddressBytes = []byte("abc")
		tph, _ := NewTopologyHandler(args)

		// 0/1%2=0 -> providedSortedPublicKeys[0] != providedAddress -> not leader
		assert.False(t, tph.MyTurnAsLeader())
	})

	t.Run("leader", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsTopologyHandler()
		tph, _ := NewTopologyHandler(args)

		// index=0/1%2=0 -> providedSortedPublicKeys[0] == providedAddress -> leader
		assert.True(t, tph.MyTurnAsLeader())
	})
}

func createTimerStubWithUnixValue(value int64) *testsCommon.TimerStub {
	stub := testsCommon.NewTimerStub()
	stub.NowUnixCalled = func() int64 {
		return value
	}
	return stub
}

func createMockArgsTopologyHandler() ArgsTopologyHandler {
	return ArgsTopologyHandler{
		PublicKeysProvider: &testsCommon.BroadcasterStub{
			SortedPublicKeysCalled: func() [][]byte {
				return [][]byte{[]byte("aaa"), []byte("bbb")}
			},
		},
		Timer:        createTimerStubWithUnixValue(0),
		StepDuration: duration,
		AddressBytes: []byte("aaa"),
	}
}
