package status

import (
	"sync"

	"github.com/multiversx/mx-bridge-eth-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("status")

type statusHandler struct {
	mutStatus     sync.RWMutex
	intMetrics    map[string]int
	stringMetrics map[string]string
	storer        core.Storer
	name          string
}

// NewStatusHandler creates a new instance of the status handler
func NewStatusHandler(name string, storer core.Storer) (*statusHandler, error) {
	if len(name) == 0 {
		return nil, ErrEmptyName
	}
	if check.IfNil(storer) {
		return nil, ErrNilStorer
	}

	sh := &statusHandler{
		storer:        storer,
		intMetrics:    make(map[string]int),
		stringMetrics: make(map[string]string),
		name:          name,
	}
	sh.tryLoadPersistedData()

	return sh, nil
}

// SetIntMetric will set the metric from an int value
func (sh *statusHandler) SetIntMetric(metric string, value int) {
	sh.mutStatus.Lock()
	defer sh.mutStatus.Unlock()

	sh.intMetrics[metric] = value
	sh.persistChanges(metric)
}

// AddIntMetric will update the provided metric with the delta value
func (sh *statusHandler) AddIntMetric(metric string, delta int) {
	sh.mutStatus.Lock()
	defer sh.mutStatus.Unlock()

	sh.intMetrics[metric] += delta
	sh.persistChanges(metric)
}

// SetStringMetric will update the provided metric with string value
func (sh *statusHandler) SetStringMetric(metric string, val string) {
	sh.mutStatus.Lock()
	defer sh.mutStatus.Unlock()

	sh.stringMetrics[metric] = val
	sh.persistChanges(metric)
}

// GetStringMetrics returns the string metrics
func (sh *statusHandler) GetStringMetrics() core.StringMetrics {
	metrics := make(core.StringMetrics)

	sh.mutStatus.RLock()
	defer sh.mutStatus.RUnlock()

	for metric, value := range sh.stringMetrics {
		metrics[metric] = value
	}

	return metrics
}

// GetIntMetrics returns the int metrics
func (sh *statusHandler) GetIntMetrics() core.IntMetrics {
	metrics := make(core.IntMetrics)

	sh.mutStatus.RLock()
	defer sh.mutStatus.RUnlock()

	for metric, value := range sh.intMetrics {
		metrics[metric] = value
	}

	return metrics
}

// GetAllMetrics returns all contained metrics as objects map.
func (sh *statusHandler) GetAllMetrics() core.GeneralMetrics {
	sh.mutStatus.RLock()
	defer sh.mutStatus.RUnlock()

	generalMetrics := make(core.GeneralMetrics)
	for key, val := range sh.intMetrics {
		generalMetrics[key] = val
	}
	for key, val := range sh.stringMetrics {
		generalMetrics[key] = val
	}

	return generalMetrics
}

// Name returns the status handler's name
func (sh *statusHandler) Name() string {
	return sh.name
}

func (sh *statusHandler) tryLoadPersistedData() {
	if check.IfNil(sh.storer) {
		log.Debug("no persister provided, using in-memory caches")
		return
	}

	data, err := sh.storer.Get([]byte(sh.name))
	if err != nil {
		log.Debug("statusHandler.tryLoadPersistedData reading from storer", "name", sh.name, "error", err)
		return
	}

	persistence, err := loadFromBuff(data)
	if err != nil {
		log.Debug("statusHandler.tryLoadPersistedData loading from buffer", "name", sh.name, "error", err)
		return
	}

	for key, val := range persistence.IntMetrics {
		sh.intMetrics[key] = val
	}
	for key, val := range persistence.StringMetrics {
		sh.stringMetrics[key] = val
	}

	loadedMetrics := len(sh.intMetrics) + len(sh.stringMetrics)
	log.Debug("statusHandler.tryLoadPersistedData loaded data", "name", sh.name, "num metrics", loadedMetrics)
}

func (sh *statusHandler) persistChanges(metric string) {
	if !shouldPersistMetric(metric) {
		return
	}

	// it is safe to simply copy the map pointers because we are still under the mutex and after the call to save end,
	// no one will keep those pointers but the statusHandler
	persistence := &statusHandlerPersistenceData{
		IntMetrics:    sh.intMetrics,
		StringMetrics: sh.stringMetrics,
	}

	buff, num, err := convertToBuff(persistence)
	if err != nil {
		log.Debug("statusHandler.persistChanges save to buffer", "name", sh.name, "error", err)
		return
	}

	err = sh.storer.Put([]byte(sh.name), buff)
	if err != nil {
		log.Debug("statusHandler.persistChanges writing to storer", "name", sh.name, "error", err)
		return
	}

	log.Trace("statusHandler.persistChanges saved data", "name", sh.name, "num metrics", num)
}

// IsInterfaceNil returns true if there is no value under the interface
func (sh *statusHandler) IsInterfaceNil() bool {
	return sh == nil
}
