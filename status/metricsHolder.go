package status

import (
	"fmt"
	"sort"
	"sync"

	"github.com/multiversx/mx-bridge-eth-go/core"
)

type metricsHolder struct {
	mut            sync.RWMutex
	statusHandlers map[string]core.StatusHandler
}

// NewMetricsHolder returns a new instance of the component able to hold status handlers
func NewMetricsHolder() *metricsHolder {
	return &metricsHolder{
		statusHandlers: make(map[string]core.StatusHandler),
	}
}

// AddStatusHandler adds the new status handler, if it does not exist
func (mh *metricsHolder) AddStatusHandler(sh core.StatusHandler) error {
	mh.mut.Lock()
	defer mh.mut.Unlock()

	name := sh.Name()
	_, exists := mh.statusHandlers[name]
	if exists {
		return fmt.Errorf("%w for %s", ErrStatusHandlerExists, name)
	}

	mh.statusHandlers[name] = sh

	return nil
}

// GetAvailableStatusHandlers returns a list of all available status handlers
func (mh *metricsHolder) GetAvailableStatusHandlers() []string {
	mh.mut.RLock()
	defer mh.mut.RUnlock()

	names := make([]string, 0, len(mh.statusHandlers))
	for name := range mh.statusHandlers {
		names = append(names, name)
	}

	sort.Slice(names, func(i, j int) bool {
		return names[i] < names[j]
	})

	return names
}

// GetAllMetrics returns all metrics from a specified status handler
func (mh *metricsHolder) GetAllMetrics(name string) (core.GeneralMetrics, error) {
	mh.mut.RLock()
	defer mh.mut.RUnlock()

	sh, exists := mh.statusHandlers[name]
	if !exists {
		return nil, ErrMissingStatusHandler
	}

	return sh.GetAllMetrics(), nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (mh *metricsHolder) IsInterfaceNil() bool {
	return mh == nil
}
