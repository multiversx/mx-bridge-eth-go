package timer

import (
	"github.com/multiversx/mx-chain-go/ntp"
)

func newNTPTimerWithInnerSyncTimer(ntpSyncTimer ntp.SyncTimer) *ntpTimer {
	return &ntpTimer{
		ntpSyncTimer: ntpSyncTimer,
	}
}
