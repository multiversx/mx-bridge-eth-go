package topology

import (
	"bytes"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-eth-bridge/core/converters"
	"github.com/ElrondNetwork/elrond-eth-bridge/testsCommon"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/stretchr/testify/assert"
)

var duration = time.Second

func TestNewTopologyHandler(t *testing.T) {
	t.Parallel()

	t.Run("nil PublicKeysProvider", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsTopologyHandler()
		args.PublicKeysProvider = nil
		tph, err := NewTopologyHandler(args)

		assert.True(t, check.IfNil(tph))
		assert.Equal(t, errNilPublicKeysProvider, err)
	})
	t.Run("nil timer", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsTopologyHandler()
		args.Timer = nil
		tph, err := NewTopologyHandler(args)

		assert.True(t, check.IfNil(tph))
		assert.Equal(t, errNilTimer, err)
	})
	t.Run("invalid step duration", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsTopologyHandler()
		args.IntervalForLeader = time.Duration(12345)
		tph, err := NewTopologyHandler(args)

		assert.True(t, check.IfNil(tph))
		assert.Equal(t, errInvalidIntervalForLeader, err)
	})
	t.Run("empty address", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsTopologyHandler()
		args.AddressBytes = nil
		tph, err := NewTopologyHandler(args)

		assert.True(t, check.IfNil(tph))
		assert.Equal(t, errEmptyAddress, err)

		args.AddressBytes = make([]byte, 0)
		tph, err = NewTopologyHandler(args)

		assert.True(t, check.IfNil(tph))
		assert.Equal(t, errEmptyAddress, err)
	})
	t.Run("nil logger", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsTopologyHandler()
		args.Log = nil
		tph, err := NewTopologyHandler(args)

		assert.True(t, check.IfNil(tph))
		assert.Equal(t, errNilLogger, err)
	})
	t.Run("nil address converter", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsTopologyHandler()
		args.AddressConverter = nil
		tph, err := NewTopologyHandler(args)

		assert.True(t, check.IfNil(tph))
		assert.Equal(t, errNilAddressConverter, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsTopologyHandler()
		tph, err := NewTopologyHandler(args)

		assert.False(t, check.IfNil(tph))
		assert.Nil(t, err)
		assert.False(t, tph.IsInterfaceNil()) // IsInterfaceNIl

		assert.True(t, args.PublicKeysProvider == tph.publicKeysProvider) // pointer testing
		assert.Equal(t, args.Timer, tph.timer)
		assert.Equal(t, args.IntervalForLeader, tph.intervalForLeader)
		assert.Equal(t, args.AddressBytes, tph.addressBytes)
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
		args.AddressBytes = bytes.Repeat([]byte("3"), 32)
		tph, _ := NewTopologyHandler(args)

		assert.False(t, tph.MyTurnAsLeader())
	})

	t.Run("leader", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsTopologyHandler()
		tph, _ := NewTopologyHandler(args)

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
	addressConverter, _ := converters.NewAddressConverter()
	return ArgsTopologyHandler{
		PublicKeysProvider: &testsCommon.BroadcasterStub{
			SortedPublicKeysCalled: func() [][]byte {
				return [][]byte{
					bytes.Repeat([]byte("1"), 32),
					bytes.Repeat([]byte("2"), 32),
				}
			},
		},
		Timer:             createTimerStubWithUnixValue(0),
		IntervalForLeader: duration,
		AddressBytes:      bytes.Repeat([]byte("1"), 32),
		Log:               logger.GetOrCreate("test"),
		AddressConverter:  addressConverter,
	}
}
