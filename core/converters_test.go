package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConvertFromByteSliceToArray(t *testing.T) {
	t.Parallel()

	buff := []byte("12345678901234567890123456789012")

	result := ConvertFromByteSliceToArray(buff)
	assert.Equal(t, buff, result[:])
}
