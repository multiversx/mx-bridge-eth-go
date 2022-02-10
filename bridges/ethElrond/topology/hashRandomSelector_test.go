package topology

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHashRandomSelector_randomInt(t *testing.T) {
	t.Parallel()

	selector := &hashRandomSelector{}
	seedValue := uint64(1641988500)

	assert.Equal(t, uint64(0), selector.randomInt(0, 0))
	assert.Equal(t, uint64(0), selector.randomInt(seedValue, 0))

	assert.Equal(t, uint64(9), selector.randomInt(seedValue, 10))
	assert.Equal(t, uint64(0), selector.randomInt(seedValue, 1))

	assert.Equal(t, uint64(4), selector.randomInt(seedValue+12, 10))
	assert.Equal(t, uint64(0), selector.randomInt(seedValue+12, 1))

	assert.Equal(t, uint64(1), selector.randomInt(seedValue+24, 10))
	assert.Equal(t, uint64(2), selector.randomInt(seedValue+36, 10))
	assert.Equal(t, uint64(0), selector.randomInt(seedValue+48, 10))
	assert.Equal(t, uint64(0), selector.randomInt(seedValue+60, 10))
	assert.Equal(t, uint64(1), selector.randomInt(seedValue+72, 10))
	assert.Equal(t, uint64(1), selector.randomInt(seedValue+84, 10))
	assert.Equal(t, uint64(6), selector.randomInt(seedValue+96, 10))
	assert.Equal(t, uint64(5), selector.randomInt(seedValue+108, 10))
	assert.Equal(t, uint64(9), selector.randomInt(seedValue+120, 10))
}

func TestHashRandomSelector_randomIntTemp(t *testing.T) {
	t.Parallel()

	selector := &hashRandomSelector{}
	seedValue := uint64(1641988500)

	for i := 0; i < 10000; i++ {
		fmt.Println(selector.randomInt((seedValue + uint64(i)/120), 10))
	}
}

func TestHashRandomSelector_randomIntDistribution(t *testing.T) {
	t.Parallel()

	selector := &hashRandomSelector{}
	seedValue := uint64(1641988500)

	setSize := 100000
	values := make(map[uint64]int)
	maxValue := uint64(10)
	for i := 0; i < setSize; i++ {
		value := selector.randomInt(seedValue, maxValue)
		seedValue += 12

		values[value]++
	}

	assert.Equal(t, int(maxValue), len(values))
	fmt.Println(values)
	minCounter := 9800
	maxCounter := 11000
	for _, counter := range values {
		assert.True(t, counter > minCounter)
		assert.True(t, counter < maxCounter)
	}
}
