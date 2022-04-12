package disabled

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/stretchr/testify/assert"
)

func TestNewDisabledGasStation(t *testing.T) {
	dbv := &DisabledBatchValidator{}

	assert.False(t, check.IfNil(dbv))

	isValid, err := dbv.ValidateBatch(nil)
	assert.True(t, isValid)
	assert.Nil(t, err)
}
