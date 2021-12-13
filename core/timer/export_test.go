package timer

import (
	"github.com/ElrondNetwork/elrond-go/ntp"
)

func newNTPTimerWithInnerSyncTimer(ntpSyncTimer ntp.SyncTimer) *ntpTimer {
	return &ntpTimer{
		ntpSyncTimer: ntpSyncTimer,
	}
}
