package status

import (
	"errors"
	"testing"

	"github.com/multiversx/mx-bridge-eth-go/core"
	"github.com/multiversx/mx-bridge-eth-go/testsCommon"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/stretchr/testify/assert"
)

func TestNewMetricsHolder(t *testing.T) {
	t.Parallel()

	mh := NewMetricsHolder()
	assert.False(t, check.IfNil(mh))
}

func TestMetricsHolder_AddStatusHandler(t *testing.T) {
	t.Parallel()

	mh := NewMetricsHolder()

	err := mh.AddStatusHandler(testsCommon.NewStatusHandlerMock("mock1"))
	assert.Nil(t, err)
	assert.Equal(t, 1, len(mh.statusHandlers))

	err = mh.AddStatusHandler(testsCommon.NewStatusHandlerMock("mock2"))
	assert.Nil(t, err)
	assert.Equal(t, 2, len(mh.statusHandlers))

	err = mh.AddStatusHandler(testsCommon.NewStatusHandlerMock("mock1"))
	assert.True(t, errors.Is(err, ErrStatusHandlerExists))
	assert.Equal(t, 2, len(mh.statusHandlers))
}

func TestMetricsHolder_GetAvailableStatusHandlers(t *testing.T) {
	t.Parallel()

	mh := NewMetricsHolder()
	assert.Equal(t, make([]string, 0), mh.GetAvailableStatusHandlers())

	_ = mh.AddStatusHandler(testsCommon.NewStatusHandlerMock("mock2"))
	assert.Equal(t, []string{"mock2"}, mh.GetAvailableStatusHandlers())

	_ = mh.AddStatusHandler(testsCommon.NewStatusHandlerMock("mock1"))
	assert.Equal(t, []string{"mock1", "mock2"}, mh.GetAvailableStatusHandlers())
}

func TestMetricsHolder_GetAllMetrics(t *testing.T) {
	t.Parallel()

	mh := NewMetricsHolder()
	metrics, err := mh.GetAllMetrics("not-found")
	assert.Nil(t, metrics)
	assert.True(t, errors.Is(err, ErrMissingStatusHandler))

	expected := make(core.GeneralMetrics)
	sh := testsCommon.NewStatusHandlerMock("mock1")
	sh.AddIntMetric("metric1", 1)
	_ = mh.AddStatusHandler(sh)
	expected["metric1"] = 1

	metrics, err = mh.GetAllMetrics("mock1")
	assert.Nil(t, err)
	assert.Equal(t, expected, metrics)

	metrics, err = mh.GetAllMetrics("mock2")
	assert.Nil(t, metrics)
	assert.True(t, errors.Is(err, ErrMissingStatusHandler))
}
