package mock

import "strings"

type mockLogObserver struct {
	expectedStringInLog string
	logFoundChan        chan struct{}
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
	}

	return 0, nil
}

// LogFoundChan returns the internal chan
func (observer *mockLogObserver) LogFoundChan() chan struct{} {
	return observer.logFoundChan
}
