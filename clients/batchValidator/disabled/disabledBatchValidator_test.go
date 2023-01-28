package disabled

import (
	"context"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/stretchr/testify/assert"
)

func TestNewDisabledBatchValidator(t *testing.T) {
	dbv := NewDisabledBatchValidator()

	assert.False(t, check.IfNil(dbv))

	isValid, err := dbv.ValidateBatch(context.Background(), nil)
	assert.True(t, isValid)
	assert.Nil(t, err)
}
