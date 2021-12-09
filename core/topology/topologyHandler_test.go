package topology

import (
    "testing"
    "time"

    "github.com/ElrondNetwork/elrond-eth-bridge/core"
    "github.com/ElrondNetwork/elrond-eth-bridge/testsCommon"
    "github.com/stretchr/testify/assert"
)

var duration = time.Second

func TestNewTopologyHandler(t *testing.T) {
    t.Parallel()

    t.Run("should work", func(t *testing.T) {
        providedSortedPublicKeys := [][]byte{[]byte("aaa"), []byte("bbb")}
        providedTimer := createTimerStubWithUnixValue(123)
        providedAddress := []byte("aaa")
        providedStepDuration := duration
        tph, err := createTopologyHandler(providedSortedPublicKeys, providedTimer, providedStepDuration, providedAddress)
        assert.NotNil(t, tph) // pointer check
        assert.Nil(t, err)
        assert.False(t, tph.IsInterfaceNil()) // IsInterfaceNIl

        assert.Equal(t, providedSortedPublicKeys, tph.sortedPublicKeys)
        assert.Equal(t, providedTimer, tph.timer)
        assert.Equal(t, providedAddress, tph.address)
        assert.Equal(t, providedStepDuration, tph.stepDuration)
    })

    t.Run("nil providedSortedPublicKeys", func(t *testing.T) {
        expectedErr := errNilSortedPublicKeys
        tph, err := createTopologyHandler(nil, nil, 0, nil)
        assert.Nil(t, tph)
        assert.Equal(t, expectedErr, err)
    })

    t.Run("nil timer", func(t *testing.T) {
        expectedErr := errNilTimer
        tph, err := createTopologyHandler([][]byte{[]byte("abc")}, nil, 0, nil)
        assert.Nil(t, tph)
        assert.Equal(t, expectedErr, err)
    })

    t.Run("invalid step duration", func(t *testing.T) {
        expectedErr := errInvalidStepDuration
        tph, err := createTopologyHandler([][]byte{[]byte("abc")}, createTimerStubWithUnixValue(0), time.Duration(12345), nil)
        assert.Nil(t, tph)
        assert.Equal(t, expectedErr, err)
    })

    t.Run("nil address", func(t *testing.T) {
        expectedErr := errNilAddress
        tph, err := createTopologyHandler([][]byte{[]byte("abc")}, createTimerStubWithUnixValue(0), duration, nil)
        assert.Nil(t, tph)
        assert.Equal(t, expectedErr, err)
    })
}

func TestMyTurnAsLeader(t *testing.T) {
    t.Parallel()

    t.Run("not leader - SortedPublicKeys empty", func(t *testing.T) {
        tph, _ := createTopologyHandler([][]byte{}, createTimerStubWithUnixValue(0), duration, []byte("abc"))
        assert.False(t, tph.MyTurnAsLeader())
    })

    t.Run("not leader", func(t *testing.T) {
        tph, _ := createTopologyHandler([][]byte{[]byte("aaa"), []byte("bbb")}, createTimerStubWithUnixValue(0), duration, []byte("abc"))

        // 0/1%2=0 -> providedSortedPublicKeys[0] != providedAddress -> not leader
        assert.False(t, tph.MyTurnAsLeader())
    })

    t.Run("leader", func(t *testing.T) {
        expectedLeader := []byte("aaa")
        tph, _ := createTopologyHandler([][]byte{expectedLeader, []byte("bbb")}, createTimerStubWithUnixValue(0), duration, expectedLeader)

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

func createTopologyHandler(pk [][]byte, timer core.Timer, duration time.Duration, address []byte) (*topologyHandler, error) {
    args := ArgsTopologyHandler{
        SortedPublicKeys: pk,
        Timer:            timer,
        StepDuration:     duration,
        Address:          address,
    }
    return NewTopologyHandler(args)
}
