package relay

import (
	"time"

	logger "github.com/ElrondNetwork/elrond-go-logger"
)

func setTestLogLevel() {
	_ = logger.SetLogLevel("*:" + logger.LogNone.String())
}

type timerStub struct {
	afterDuration time.Duration
	timeNowUnix   int64
}

func (s *timerStub) after(time.Duration) <-chan time.Time {
	return time.After(s.afterDuration)
}

func (s *timerStub) nowUnix() int64 {
	return s.timeNowUnix
}
