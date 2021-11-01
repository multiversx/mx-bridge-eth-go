package testsCommon

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

var fullPathTimerStub = "github.com/ElrondNetwork/elrond-eth-bridge/testsCommon.(*TimerStub)."

// TimerStub -
type TimerStub struct {
	functionCalledCounter map[string]int
	mut                   sync.RWMutex

	AfterCalled   func(d time.Duration) <-chan time.Time
	NowUnixCalled func() int64
	StartCalled   func()
	CloseCalled   func() error
}

// NewTimerStub -
func NewTimerStub() *TimerStub {
	return &TimerStub{
		functionCalledCounter: make(map[string]int),
	}
}

// After -
func (stub *TimerStub) After(d time.Duration) <-chan time.Time {
	stub.incrementFunctionCounter()
	if stub.AfterCalled != nil {
		return stub.AfterCalled(d)
	}

	ch := make(chan time.Time)
	close(ch)

	return ch
}

// NowUnix -
func (stub *TimerStub) NowUnix() int64 {
	stub.incrementFunctionCounter()
	if stub.NowUnixCalled != nil {
		return stub.NowUnixCalled()
	}

	return 0
}

// Start -
func (stub *TimerStub) Start() {
	stub.incrementFunctionCounter()
	if stub.StartCalled != nil {
		stub.StartCalled()
	}
}

// Close -
func (stub *TimerStub) Close() error {
	stub.incrementFunctionCounter()
	if stub.CloseCalled != nil {
		return stub.CloseCalled()
	}

	return nil
}

// incrementFunctionCounter increments the counter for the function that called it
func (stub *TimerStub) incrementFunctionCounter() {
	stub.mut.Lock()
	defer stub.mut.Unlock()

	pc, _, _, _ := runtime.Caller(1)
	fmt.Printf("TimerStub: called %s\n", runtime.FuncForPC(pc).Name())
	stub.functionCalledCounter[runtime.FuncForPC(pc).Name()]++
}

// GetFunctionCounter returns the called counter of a given function
func (stub *TimerStub) GetFunctionCounter(function string) int {
	stub.mut.Lock()
	defer stub.mut.Unlock()

	return stub.functionCalledCounter[fullPathTimerStub+function]
}

// IsInterfaceNil returns true if there is no value under the interface
func (stub *TimerStub) IsInterfaceNil() bool {
	return stub == nil
}
