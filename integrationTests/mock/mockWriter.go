package mock

import "strings"

type mockLogObserver struct {
	expectedStringInLog string
	exceptStrings       []string
	logFoundChan        chan struct{}
}

// NewMockLogObserver returns a new instance of mockLogObserver
func NewMockLogObserver(expectedStringInLog string, exceptStrings ...string) *mockLogObserver {
	return &mockLogObserver{
		expectedStringInLog: expectedStringInLog,
		exceptStrings:       exceptStrings,
		logFoundChan:        make(chan struct{}, 1),
	}
}

// Write is called by the logger
func (observer *mockLogObserver) Write(log []byte) (n int, err error) {
	str := string(log)
	if !strings.Contains(str, observer.expectedStringInLog) {
		return 0, nil
	}
	if observer.stringIsExcepted(str) {
		return 0, nil
	}

	observer.logFoundChan <- struct{}{}

	return 0, nil
}

func (observer *mockLogObserver) stringIsExcepted(str string) bool {
	for _, exceptString := range observer.exceptStrings {
		if strings.Contains(str, exceptString) {
			return true
		}
	}

	return false
}

// LogFoundChan returns the internal chan
func (observer *mockLogObserver) LogFoundChan() chan struct{} {
	return observer.logFoundChan
}
