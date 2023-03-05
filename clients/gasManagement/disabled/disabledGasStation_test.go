package disabled

import (
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/stretchr/testify/assert"
)

func TestNewDisabledGasStation(t *testing.T) {
	dgs := &DisabledGasStation{}

	assert.False(t, check.IfNil(dgs))

	gasPrice, err := dgs.GetCurrentGasPrice()
	assert.Equal(t, big.NewInt(0), gasPrice)
	assert.Nil(t, err)
}
