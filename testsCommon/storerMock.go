package testsCommon

import (
	"errors"
	"sync"
)

// StorerMock -
type StorerMock struct {
	mut  sync.RWMutex
	data map[string][]byte
}

// NewStorerMock -
func NewStorerMock() *StorerMock {
	return &StorerMock{
		data: make(map[string][]byte),
	}
}

// Put -
func (sm *StorerMock) Put(key, data []byte) error {
	sm.mut.Lock()
	defer sm.mut.Unlock()

	sm.data[string(key)] = data

	return nil
}

// Get -
func (sm *StorerMock) Get(key []byte) ([]byte, error) {
	sm.mut.RLock()
	defer sm.mut.RUnlock()

	val, found := sm.data[string(key)]
	if !found {
		return nil, errors.New("key not found")
	}

	return val, nil
}

// Close -
func (sm *StorerMock) Close() error {
	return nil
}

// IsInterfaceNil -
func (sm *StorerMock) IsInterfaceNil() bool {
	return sm == nil
}
