package status

import (
	"testing"

	"github.com/ElrondNetwork/elrond-eth-bridge/core"
	"github.com/ElrondNetwork/elrond-eth-bridge/testsCommon"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/stretchr/testify/assert"
)

func TestNewStatusHandler(t *testing.T) {
	t.Parallel()

	t.Run("empty name should error", func(t *testing.T) {
		sh, err := NewStatusHandler("")
		assert.Equal(t, ErrEmptyName, err)
		assert.True(t, check.IfNil(sh))
	})
	t.Run("should work", func(t *testing.T) {
		name := "test"
		sh, err := NewStatusHandler(name)
		assert.Nil(t, err)
		assert.False(t, check.IfNil(sh))
		assert.Equal(t, name, sh.Name())
	})
}

func TestStatusHandler_IntMetrics(t *testing.T) {
	t.Parallel()

	sh, _ := NewStatusHandler("test int metrics")
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

	sh, _ := NewStatusHandler("test string metrics")
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

func TestStatusHandler_GetAllMetrics(t *testing.T) {
	t.Parallel()

	sh, _ := NewStatusHandler("test get all metrics")
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
