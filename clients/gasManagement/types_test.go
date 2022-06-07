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

	response := createMockGasStationResponse()
	expectedTrueString := `{
        "status":"1",
        "message":"OK-Missing/Invalid API Key, rate limit of 1/5sec applied",
        "result":{
            "LastBlock":"14836699",
            "SafeGasPrice":"81",
            "ProposeGasPrice":"82",
            "FastGasPrice":"83",
            "suggestBaseFee":"80.856621497",
            "gasUsedRatio":"0.0422401857919075,0.636178148305543,0.399708304558626,0.212555933333333,0.645151576152554"
        }
}`

	assert.Equal(t, stripSpaces(expectedTrueString), stripSpaces(response.String()))
}
