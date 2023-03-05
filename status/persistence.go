package status

import (
	"github.com/multiversx/mx-bridge-eth-go/core"
	"github.com/multiversx/mx-chain-core-go/marshal"
)

// only json marshaller is supported because we used maps
var marshaller = &marshal.JsonMarshalizer{}

type statusHandlerPersistenceData struct {
	IntMetrics    core.IntMetrics    `json:"intMetrics"`
	StringMetrics core.StringMetrics `json:"stringMetrics"`
}

func loadFromBuff(buff []byte) (*statusHandlerPersistenceData, error) {
	data := &statusHandlerPersistenceData{}
	err := marshaller.Unmarshal(data, buff)
	if err != nil {
		return nil, err
	}

	neededData := &statusHandlerPersistenceData{
		IntMetrics:    make(core.IntMetrics),
		StringMetrics: make(core.StringMetrics),
	}
	for key, val := range data.IntMetrics {
		if !shouldPersistMetric(key) {
			continue
		}

		neededData.IntMetrics[key] = val
	}
	for key, val := range data.StringMetrics {
		if !shouldPersistMetric(key) {
			continue
		}

		neededData.StringMetrics[key] = val
	}

	return neededData, nil
}

func convertToBuff(persistence *statusHandlerPersistenceData) ([]byte, int, error) {
	neededData := &statusHandlerPersistenceData{
		IntMetrics:    make(core.IntMetrics),
		StringMetrics: make(core.StringMetrics),
	}
	for key, val := range persistence.IntMetrics {
		if !shouldPersistMetric(key) {
			continue
		}

		neededData.IntMetrics[key] = val
	}
	for key, val := range persistence.StringMetrics {
		if !shouldPersistMetric(key) {
			continue
		}

		neededData.StringMetrics[key] = val
	}

	numMetrics := len(neededData.StringMetrics) + len(neededData.IntMetrics)
	buff, err := marshaller.Marshal(neededData)

	return buff, numMetrics, err
}

func shouldPersistMetric(metric string) bool {
	for _, persistedMetric := range core.PersistedMetrics {
		if persistedMetric == metric {
			return true
		}
	}

	return false
}
