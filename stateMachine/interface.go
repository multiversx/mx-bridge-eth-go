package stateMachine

import "time"

// Timer defines operations related to time
type Timer interface {
	After(d time.Duration) <-chan time.Time
	NowUnix() int64
	Start()
	Close() error
}
