package topology

import (
    "testing"
    "time"

    "github.com/ElrondNetwork/elrond-eth-bridge/core"
    "github.com/ElrondNetwork/elrond-eth-bridge/testsCommon"
    "github.com/stretchr/testify/assert"
)

func TestNewTopologyHandler(t *testing.T) {
    t.Parallel()

    providedSortedPublicKeys := make([][]byte, 0)
    providedTimer := core.Timer(nil)
    providedAddress := make([]byte, 0)
    providedStepDuration := time.Duration(0)
    args := ArgsTopologyHandler{
        SortedPublicKeys: providedSortedPublicKeys,
        Timer:            providedTimer,
        Address:          providedAddress,
        StepDuration:     providedStepDuration,
    }
    tph := NewTopologyHandler(args)
    assert.NotNil(t, tph)                 // pointer check
    assert.False(t, tph.IsInterfaceNil()) // IsInterfaceNIl

    assert.Equal(t, providedSortedPublicKeys, tph.sortedPublicKeys)
    assert.Equal(t, providedTimer, tph.timer)
    assert.Equal(t, providedAddress, tph.address)
    assert.Equal(t, providedStepDuration, tph.stepDuration)
}

func TestMyTurnAsLeader(t *testing.T) {
    t.Parallel()

    t.Run("not leader - SortedPublicKeys empty", func(t *testing.T) {
        args := ArgsTopologyHandler{
            SortedPublicKeys: nil,
            Timer:            nil,
            Address:          nil,
            StepDuration:     0,
        }
        tph := NewTopologyHandler(args)

        assert.False(t, tph.MyTurnAsLeader())
    })

    t.Run("not leader", func(t *testing.T) {
        providedSortedPublicKeys := [][]byte{[]byte("aaa"), []byte("bbb")} // numberOfPeers = 2
        providedTimer := testsCommon.NewTimerStub()
        providedTimer.NowUnixCalled = func() int64 {
            return 0
        }
        providedAddress := []byte("ccc")
        providedStepDuration := time.Duration(1000000000) // one second
        args := ArgsTopologyHandler{
            SortedPublicKeys: providedSortedPublicKeys,
            Timer:            providedTimer,
            Address:          providedAddress,
            StepDuration:     providedStepDuration,
        }
        tph := NewTopologyHandler(args)

        // 0/1%2=0 -> providedSortedPublicKeys[0] != providedAddress -> not leader
        assert.False(t, tph.MyTurnAsLeader())
    })

    t.Run("leader", func(t *testing.T) {
        expectedLeader := []byte("aaa")
        providedSortedPublicKeys := [][]byte{expectedLeader, []byte("bbb")} // numberOfPeers = 2
        providedTimer := testsCommon.NewTimerStub()
        providedTimer.NowUnixCalled = func() int64 {
            return 0
        }
        providedAddress := expectedLeader
        providedStepDuration := time.Duration(1000000000) // one second
        args := ArgsTopologyHandler{
            SortedPublicKeys: providedSortedPublicKeys,
            Timer:            providedTimer,
            Address:          providedAddress,
            StepDuration:     providedStepDuration,
        }
        tph := NewTopologyHandler(args)

        // index=0/1%2=0 -> providedSortedPublicKeys[0] == providedAddress -> leader
        assert.True(t, tph.MyTurnAsLeader())
    })
}
