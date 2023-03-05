package status

import (
	"testing"

	"github.com/multiversx/mx-bridge-eth-go/core"
	"github.com/multiversx/mx-bridge-eth-go/testsCommon"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewStatusHandler(t *testing.T) {
	t.Parallel()

	t.Run("empty name should error", func(t *testing.T) {
		sh, err := NewStatusHandler("", testsCommon.NewStorerMock())
		assert.Equal(t, ErrEmptyName, err)
		assert.True(t, check.IfNil(sh))
	})
	t.Run("nil storer should error", func(t *testing.T) {
		name := "test"
		sh, err := NewStatusHandler(name, nil)
		assert.Equal(t, ErrNilStorer, err)
		assert.True(t, check.IfNil(sh))
	})
	t.Run("with storer but not containing data", func(t *testing.T) {
		name := "test"
		storer := testsCommon.NewStorerMock()

		sh, err := NewStatusHandler(name, storer)
		require.Nil(t, err)

		expected := make(core.GeneralMetrics)

		assert.Equal(t, expected, sh.GetAllMetrics())
		assert.Equal(t, name, sh.Name())
	})
	t.Run("with storer but containing garbage", func(t *testing.T) {
		name := "test"
		storer := testsCommon.NewStorerMock()

		_ = storer.Put([]byte(name), []byte("garbage"))

		sh, err := NewStatusHandler(name, storer)
		require.Nil(t, err)

		expected := make(core.GeneralMetrics)

		assert.Equal(t, expected, sh.GetAllMetrics())
	})
	t.Run("with storer should load with autocleanup", func(t *testing.T) {
		name := "test"
		storer := testsCommon.NewStorerMock()

		existent := &statusHandlerPersistenceData{
			IntMetrics: map[string]int{
				"not-persistent-int":  5,
				core.MetricNumBatches: 6,
			},
			StringMetrics: map[string]string{
				"not-persistent-string":                   "value1",
				core.MetricLastQueriedEthereumBlockNumber: "value2",
			},
		}

		buffExistent, err := marshaller.Marshal(existent)
		require.Nil(t, err)

		_ = storer.Put([]byte(name), buffExistent)

		sh, err := NewStatusHandler(name, storer)
		require.Nil(t, err)

		expected := make(core.GeneralMetrics)
		expected[core.MetricNumBatches] = 6
		expected[core.MetricLastQueriedEthereumBlockNumber] = "value2"

		assert.Equal(t, expected, sh.GetAllMetrics())
	})
}

func TestStatusHandler_IntMetrics(t *testing.T) {
	t.Parallel()

	sh, _ := NewStatusHandler("test int metrics", testsCommon.NewStorerMock())
	assert.True(t, testsCommon.EqualIntMetrics(make(core.IntMetrics), sh.GetIntMetrics()))

	metric1 := "metric1"
	int1 := 1232
	sh.SetIntMetric(metric1, int1)

	expected := make(core.IntMetrics)
	expected[metric1] = int1
	assert.True(t, testsCommon.EqualIntMetrics(expected, sh.GetIntMetrics()))

	sh.AddIntMetric(metric1, 1)
	expected = make(core.IntMetrics)
	expected[metric1] = int1 + 1
	assert.True(t, testsCommon.EqualIntMetrics(expected, sh.GetIntMetrics()))

	metric2 := "metric2"
	int2 := 75846
	sh.SetIntMetric(metric2, int2)

	expected = make(core.IntMetrics)
	expected[metric1] = int1 + 1
	expected[metric2] = int2
	assert.True(t, testsCommon.EqualIntMetrics(expected, sh.GetIntMetrics()))

	sh.SetIntMetric(metric1, int1)
	expected = make(core.IntMetrics)
	expected[metric1] = int1
	expected[metric2] = int2
	assert.True(t, testsCommon.EqualIntMetrics(expected, sh.GetIntMetrics()))
}

func TestStatusHandler_StringMetrics(t *testing.T) {
	t.Parallel()

	sh, _ := NewStatusHandler("test string metrics", testsCommon.NewStorerMock())
	assert.True(t, testsCommon.EqualIntMetrics(make(core.IntMetrics), sh.GetIntMetrics()))

	metric1 := "metric1"
	value1 := "value1"
	sh.SetStringMetric(metric1, value1)

	expected := make(core.StringMetrics)
	expected[metric1] = value1
	assert.True(t, testsCommon.EqualStringMetrics(expected, sh.GetStringMetrics()))

	metric2 := "metric2"
	value2 := "value2"
	sh.SetStringMetric(metric2, value2)

	expected = make(core.StringMetrics)
	expected[metric1] = value1
	expected[metric2] = value2
	assert.True(t, testsCommon.EqualStringMetrics(expected, sh.GetStringMetrics()))

	newValue1 := "new value 1"
	sh.SetStringMetric(metric1, newValue1)
	expected = make(core.StringMetrics)
	expected[metric1] = newValue1
	expected[metric2] = value2
	assert.True(t, testsCommon.EqualStringMetrics(expected, sh.GetStringMetrics()))
}

func TestStatusHandler_SetMetricsWithStorer(t *testing.T) {
	t.Parallel()

	name := "test"
	storer := testsCommon.NewStorerMock()
	sh, _ := NewStatusHandler(name, storer)

	sh.AddIntMetric("not persistent", 1)
	buff, err := storer.Get([]byte(name))
	assert.NotNil(t, err)
	assert.Nil(t, buff)

	sh.SetStringMetric("not persistent", "22")
	buff, err = storer.Get([]byte(name))
	assert.NotNil(t, err)
	assert.Nil(t, buff)

	sh.AddIntMetric(core.MetricNumBatches, 1)
	sh.SetStringMetric(core.MetricNumEthClientRequests, "22")
	buff, err = storer.Get([]byte(name))
	assert.Nil(t, err)

	persistence := &statusHandlerPersistenceData{}
	err = marshaller.Unmarshal(persistence, buff)
	assert.Nil(t, err)

	expectedPeristence := &statusHandlerPersistenceData{
		IntMetrics: map[string]int{
			core.MetricNumBatches: 1,
		},
		StringMetrics: map[string]string{
			core.MetricNumEthClientRequests: "22",
		},
	}

	assert.Equal(t, expectedPeristence, persistence)
}

func TestStatusHandler_GetAllMetrics(t *testing.T) {
	t.Parallel()

	sh, _ := NewStatusHandler("test get all metrics", testsCommon.NewStorerMock())
	metric1 := "metric1"
	value1 := "value1"
	sh.SetStringMetric(metric1, value1)

	metric2 := "metric2"
	value2 := 4
	sh.SetIntMetric(metric2, value2)

	expectedMap := make(core.GeneralMetrics)
	expectedMap[metric1] = value1
	expectedMap[metric2] = value2

	assert.Equal(t, expectedMap, sh.GetAllMetrics())
}
