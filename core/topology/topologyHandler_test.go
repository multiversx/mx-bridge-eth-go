package topology

import (
    "testing"
    "time"

    "github.com/ElrondNetwork/elrond-eth-bridge/testsCommon"
    "github.com/stretchr/testify/assert"
)

var duration = time.Duration(1000000000) // one sec

func TestNewTopologyHandler(t *testing.T) {
    t.Parallel()

    t.Run("should work", func(t *testing.T) {
        providedSortedPublicKeys := [][]byte{[]byte("aaa"), []byte("bbb")}
        providedTimer := createTimerStubWithUnixValue(123)
        providedAddress := []byte("aaa")
        providedStepDuration := duration
        args := ArgsTopologyHandler{
            SortedPublicKeys: providedSortedPublicKeys,
            Timer:            providedTimer,
            Address:          providedAddress,
            StepDuration:     providedStepDuration,
        }
        tph, err := NewTopologyHandler(args)
        assert.NotNil(t, tph) // pointer check
        assert.Nil(t, err)
        assert.False(t, tph.IsInterfaceNil()) // IsInterfaceNIl

        assert.Equal(t, providedSortedPublicKeys, tph.sortedPublicKeys)
        assert.Equal(t, providedTimer, tph.timer)
        assert.Equal(t, providedAddress, tph.address)
        assert.Equal(t, providedStepDuration, tph.stepDuration)
    })

    t.Run("nil providedSortedPublicKeys", func(t *testing.T) {
        expectedErr := ErrNilSortedPublicKeys
        args := ArgsTopologyHandler{}
        tph, err := NewTopologyHandler(args)
        assert.Nil(t, tph)
        assert.Equal(t, expectedErr, err)
    })

    t.Run("nil timer", func(t *testing.T) {
        expectedErr := ErrNilTimer
        args := ArgsTopologyHandler{}
        args.SortedPublicKeys = [][]byte{[]byte("abc")}
        tph, err := NewTopologyHandler(args)
        assert.Nil(t, tph)
        assert.Equal(t, expectedErr, err)
    })

    t.Run("invalid step duration", func(t *testing.T) {
        expectedErr := ErrInvalidStepDuration
        args := ArgsTopologyHandler{
            SortedPublicKeys: [][]byte{[]byte("abc")},
            Timer:            createTimerStubWithUnixValue(0),
            StepDuration:     time.Duration(12345),
        }
        tph, err := NewTopologyHandler(args)
        assert.Nil(t, tph)
        assert.Equal(t, expectedErr, err)
    })

    t.Run("nil address", func(t *testing.T) {
        expectedErr := ErrNilAddress
        args := ArgsTopologyHandler{
            SortedPublicKeys: [][]byte{[]byte("abc")},
            Timer:            createTimerStubWithUnixValue(0),
            StepDuration:     duration,
        }
        tph, err := NewTopologyHandler(args)
        assert.Nil(t, tph)
        assert.Equal(t, expectedErr, err)
    })
}

func TestMyTurnAsLeader(t *testing.T) {
    t.Parallel()

    t.Run("not leader - SortedPublicKeys empty", func(t *testing.T) {
        args := ArgsTopologyHandler{
            SortedPublicKeys: [][]byte{},
            Timer:            createTimerStubWithUnixValue(0),
            Address:          []byte("abc"),
            StepDuration:     duration,
        }
        tph, _ := NewTopologyHandler(args)

        assert.False(t, tph.MyTurnAsLeader())
    })

    t.Run("not leader", func(t *testing.T) {
        args := ArgsTopologyHandler{
            SortedPublicKeys: [][]byte{[]byte("aaa"), []byte("bbb")}, // numberOfPeers = 2
            Timer:            createTimerStubWithUnixValue(0),
            Address:          []byte("ccc"),
            StepDuration:     duration,
        }
        tph, _ := NewTopologyHandler(args)

        // 0/1%2=0 -> providedSortedPublicKeys[0] != providedAddress -> not leader
        assert.False(t, tph.MyTurnAsLeader())
    })

    t.Run("leader", func(t *testing.T) {
        expectedLeader := []byte("aaa")
        args := ArgsTopologyHandler{
            SortedPublicKeys: [][]byte{expectedLeader, []byte("bbb")}, // numberOfPeers = 2
            Timer:            createTimerStubWithUnixValue(0),
            Address:          expectedLeader,
            StepDuration:     duration,
        }
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
