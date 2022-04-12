package batchValidatorManagement

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMicroserviceResponse_String(t *testing.T) {
	t.Parallel()

	response := &microserviceResponse{
		Valid: true,
	}
	expectedTrueString := `{"valid":true}`
	assert.Equal(t, expectedTrueString, response.String())

	response.Valid = false
	expectedFalseString := `{"valid":false}`
	assert.Equal(t, expectedFalseString, response.String())
}
