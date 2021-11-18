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

func TestTrimWhiteSpaceCharacters(t *testing.T) {
	t.Parallel()

	data := "aaII139HSAh32q782!$#*$(nc"

	input := " " + data
	assert.Equal(t, data, TrimWhiteSpaceCharacters(input))

	input = "\t " + data
	assert.Equal(t, data, TrimWhiteSpaceCharacters(input))

	input = "\t " + data + "\n"
	assert.Equal(t, data, TrimWhiteSpaceCharacters(input))

	input = "\t\n " + data + "\n\n\n\n\t"
	assert.Equal(t, data, TrimWhiteSpaceCharacters(input))
}
