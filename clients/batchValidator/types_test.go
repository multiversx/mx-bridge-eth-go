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

func TestMicroserviceBadRequestBody_String(t *testing.T) {
	t.Parallel()

	response := &microserviceBadRequestBody{
		StatusCode: 400,
		Message:    "Cannot read properties of undefined (reading 'length')",
		Error:      "Bad Request",
	}
	expectedString := `{"statusCode":400,"message":"Cannot read properties of undefined (reading 'length')","error":"Bad Request"}`
	assert.Equal(t, expectedString, response.String())

}
