package timer

import (
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/ntp"
)

var defaultNTPConfig = config.NTPConfig{
	Hosts:               []string{"time.google.com", "time.cloudflare.com", "time.apple.com", "time.windows.com"},
	Port:                123,
	Version:             0,
	TimeoutMilliseconds: 100,
	SyncPeriodSeconds:   3600,
}

type ntpTimer struct {
	ntpSyncTimer ntp.SyncTimer
}

// NewNTPTimer will create a new NTP timer
func NewNTPTimer() *ntpTimer {
	return &ntpTimer{
		ntpSyncTimer: ntp.NewSyncTime(defaultNTPConfig, nil),
	}
}

// NowUnix will return the Unix time
func (n *ntpTimer) NowUnix() int64 {
	return n.ntpSyncTimer.CurrentTime().Unix()
}

// Start will start the inner NTP timer
func (n *ntpTimer) Start() {
	n.ntpSyncTimer.StartSyncingTime()
}

// Close will close the inner NTP timer
func (n *ntpTimer) Close() error {
	return n.ntpSyncTimer.Close()
}

// IsInterfaceNil returns true if there is no value under the interface
func (n *ntpTimer) IsInterfaceNil() bool {
	return n == nil
}
