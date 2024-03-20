package mock

import (
	"strings"
	"sync"
)

type mockLogObserver struct {
	expectedStringInLog string
	logFoundChan        chan struct{}
	hash                string
	mutHash             sync.RWMutex
}

// NewMockLogObserver returns a new instance of mockLogObserver
func NewMockLogObserver(expectedStringInLog string) *mockLogObserver {
	return &mockLogObserver{
		expectedStringInLog: expectedStringInLog,
		logFoundChan:        make(chan struct{}, 1),
	}
}

// Write is called by the logger
func (observer *mockLogObserver) Write(log []byte) (n int, err error) {
	if strings.Contains(string(log), observer.expectedStringInLog) {
		observer.logFoundChan <- struct{}{}
		observer.mutHash.Lock()
		observer.hash = string(log)
		observer.mutHash.Unlock()
	}

	return 0, nil
}

func (observer *mockLogObserver) GetLog() string {

	observer.mutHash.RLock()
	defer observer.mutHash.RUnlock()
	return observer.hash
}

// LogFoundChan returns the internal chan
func (observer *mockLogObserver) LogFoundChan() chan struct{} {
	return observer.logFoundChan
}
