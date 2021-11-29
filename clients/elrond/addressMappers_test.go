package elrond

import (
	"testing"

	"github.com/ElrondNetwork/elrond-eth-bridge/testsCommon/bridgeV2"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/stretchr/testify/assert"
)

func TestNewErc20ToElrondMapper(t *testing.T) {
	t.Parallel()
	t.Run("nil dataGetter", func(t *testing.T) {
		mapper, err := NewErc20ToElrondMapper(nil)
		assert.Equal(t, errNilDataGetter, err)
		assert.True(t, check.IfNil(mapper))
	})
	t.Run("should work", func(t *testing.T) {
		mapper, err := NewErc20ToElrondMapper(&bridgeV2.DataGetterStub{})
		assert.Nil(t, err)
		assert.False(t, check.IfNil(mapper))
	})
}

func TestNewElrondToErc20Mapper(t *testing.T) {
	t.Parallel()
	t.Run("nil dataGetter", func(t *testing.T) {
		mapper, err := NewElrondToErc20Mapper(nil)
		assert.Equal(t, errNilDataGetter, err)
		assert.True(t, check.IfNil(mapper))
	})
	t.Run("should work", func(t *testing.T) {
		mapper, err := NewElrondToErc20Mapper(&bridgeV2.DataGetterStub{})
		assert.Nil(t, err)
		assert.False(t, check.IfNil(mapper))
	})
}
