package gasManagement

import (
	"strings"
	"testing"
	"unicode"

	"github.com/stretchr/testify/assert"
)

func stripSpaces(str string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			// if the character is a space, drop it
			return -1
		}
		// else keep it in the string
		return r
	}, str)
}

func TestGasStationResponse_String(t *testing.T) {
	t.Parallel()

	response := &gasStationResponse{
		Fast:        370,
		Fastest:     400,
		SafeLow:     270,
		Average:     300,
		BlockTime:   15.380281690140846,
		BlockNum:    14460250,
		Speed:       0.5719409845737478,
		SafeLowWait: 17.5,
		AvgWait:     3.1,
		FastWait:    0.5,
		FastestWait: 0.5,
	}
	expectedTrueString := `{
		"fast": 370,
		"fastest": 400,
		"safeLow": 270,
		"average": 300,
		"block_time": 15.380281690140846,
		"blockNum": 14460250,
		"speed": 0.5719409845737478,
		"safeLowWait": 17.5,
		"avgWait": 3.1,
		"fastWait": 0.5,
		"fastestWait": 0.5
	}`

	assert.Equal(t, stripSpaces(expectedTrueString), stripSpaces(response.String()))
}
