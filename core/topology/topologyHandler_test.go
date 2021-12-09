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
        args := createMockArgsTopologyHandler()
        tph, err := NewTopologyHandler(args)

        assert.NotNil(t, tph) // pointer check
        assert.Nil(t, err)
        assert.False(t, tph.IsInterfaceNil()) // IsInterfaceNIl

        assert.Equal(t, args.SortedPublicKeys, tph.sortedPublicKeys)
        assert.Equal(t, args.Timer, tph.timer)
        assert.Equal(t, args.StepDuration, tph.stepDuration)
        assert.Equal(t, args.Address, tph.address)
    })

    t.Run("nil providedSortedPublicKeys", func(t *testing.T) {
        args := createMockArgsTopologyHandler()
        args.SortedPublicKeys = nil
        tph, err := NewTopologyHandler(args)

        assert.Nil(t, tph)
        assert.Equal(t, errNilSortedPublicKeys, err)
    })

    t.Run("nil timer", func(t *testing.T) {
        args := createMockArgsTopologyHandler()
        args.Timer = nil
        tph, err := NewTopologyHandler(args)

        assert.Nil(t, tph)
        assert.Equal(t, errNilTimer, err)
    })

    t.Run("invalid step duration", func(t *testing.T) {
        args := createMockArgsTopologyHandler()
        args.StepDuration = time.Duration(12345)
        tph, err := NewTopologyHandler(args)

        assert.Nil(t, tph)
        assert.Equal(t, errInvalidStepDuration, err)
    })

    t.Run("nil address", func(t *testing.T) {
        args := createMockArgsTopologyHandler()
        args.Address = nil
        tph, err := NewTopologyHandler(args)

        assert.Nil(t, tph)
        assert.Equal(t, errNilAddress, err)
    })
}

func TestMyTurnAsLeader(t *testing.T) {
    t.Parallel()

    t.Run("not leader - SortedPublicKeys empty", func(t *testing.T) {
        args := createMockArgsTopologyHandler()
        args.SortedPublicKeys = [][]byte{}
        tph, _ := NewTopologyHandler(args)

        assert.False(t, tph.MyTurnAsLeader())
    })

    t.Run("not leader", func(t *testing.T) {
        args := createMockArgsTopologyHandler()
        args.Address = []byte("abc")
        tph, _ := NewTopologyHandler(args)

        // 0/1%2=0 -> providedSortedPublicKeys[0] != providedAddress -> not leader
        assert.False(t, tph.MyTurnAsLeader())
    })

    t.Run("leader", func(t *testing.T) {
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
        SortedPublicKeys: [][]byte{[]byte("aaa"), []byte("bbb")},
        Timer:            createTimerStubWithUnixValue(0),
        StepDuration:     duration,
        Address:          []byte("aaa"),
    }
}
