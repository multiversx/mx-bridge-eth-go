package batchValidatorManagement

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-eth-bridge/clients"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/stretchr/testify/assert"
)

func createMockArgsBatchValidator() ArgsBatchValidator {
	return ArgsBatchValidator{
		SourceChain:      clients.Ethereum,
		DestinationChain: clients.Elrond,
		RequestURL:       "",
		RequestTime:      time.Second,
	}
}

func TestNewGasStation(t *testing.T) {
	t.Parallel()

	t.Run("invalid SourceChain", func(t *testing.T) {
		args := createMockArgsBatchValidator()
		args.SourceChain = ""

		gs, err := NewBatchValidator(args)
		assert.True(t, check.IfNil(gs))
		assert.True(t, errors.Is(err, clients.ErrInvalidValue))
	})
	t.Run("invalid DestinationChain", func(t *testing.T) {
		args := createMockArgsBatchValidator()
		args.DestinationChain = ""

		gs, err := NewBatchValidator(args)
		assert.True(t, check.IfNil(gs))
		assert.True(t, errors.Is(err, clients.ErrInvalidValue))
	})
	t.Run("invalid request time", func(t *testing.T) {
		args := createMockArgsBatchValidator()
		args.RequestTime = time.Duration(minRequestTime.Nanoseconds() - 1)

		gs, err := NewBatchValidator(args)
		assert.True(t, check.IfNil(gs))
		assert.True(t, errors.Is(err, clients.ErrInvalidValue))
		assert.True(t, strings.Contains(err.Error(), "checkArgs for value RequestTime"))
	})
	t.Run("should work", func(t *testing.T) {
		args := createMockArgsBatchValidator()

		gs, err := NewBatchValidator(args)
		assert.False(t, check.IfNil(gs))
		assert.Nil(t, err)
	})
}
