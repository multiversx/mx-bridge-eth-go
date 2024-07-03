package parsers

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/multiversx/mx-sdk-go/data"
	"github.com/stretchr/testify/assert"
)

func TestCallData_String(t *testing.T) {
	t.Parallel()

	t.Run("with no arguments should work", func(t *testing.T) {
		t.Parallel()

		callData := CallData{
			Type:     1,
			Function: "callMe",
			GasLimit: 50000,
		}
		expectedString := "type: 1, function: callMe, gas limit: 50000, no arguments"

		assert.Equal(t, expectedString, callData.String())
	})
	t.Run("with arguments should work", func(t *testing.T) {
		t.Parallel()

		callData := CallData{
			Type:      1,
			Function:  "callMe",
			GasLimit:  50000,
			Arguments: []string{"arg1", "arg2"},
		}
		expectedString := "type: 1, function: callMe, gas limit: 50000, arguments: arg1, arg2"

		assert.Equal(t, expectedString, callData.String())
	})
}

func TestProxySCCompleteCallData_String(t *testing.T) {
	t.Parallel()

	t.Run("nil fields should work", func(t *testing.T) {
		t.Parallel()

		callData := ProxySCCompleteCallData{
			CallData: CallData{
				Type:     1,
				Function: "callMe",
				GasLimit: 50000,
			},
			From:  common.Address{},
			Token: "tkn",
			Nonce: 1,
		}

		expectedString := "Eth address: 0x0000000000000000000000000000000000000000, MvX address: <nil>, token: tkn, amount: <nil>, nonce: 1, type: 1, function: callMe, gas limit: 50000, no arguments"
		assert.Equal(t, expectedString, callData.String())
	})
	t.Run("not a Valid MvX address should work", func(t *testing.T) {
		t.Parallel()

		callData := ProxySCCompleteCallData{
			CallData: CallData{
				Type:     1,
				Function: "callMe",
				GasLimit: 50000,
			},
			From:  common.Address{},
			To:    data.NewAddressFromBytes([]byte{0x1, 0x2}),
			Token: "tkn",
			Nonce: 1,
		}

		expectedString := "Eth address: 0x0000000000000000000000000000000000000000, MvX address: <err>, token: tkn, amount: <nil>, nonce: 1, type: 1, function: callMe, gas limit: 50000, no arguments"
		assert.Equal(t, expectedString, callData.String())
	})
	t.Run("with valid data should work", func(t *testing.T) {
		t.Parallel()

		callData := ProxySCCompleteCallData{
			CallData: CallData{
				Type:     1,
				Function: "callMe",
				GasLimit: 50000,
			},
			From:   common.Address{},
			Token:  "tkn",
			Amount: big.NewInt(37),
			Nonce:  1,
		}
		ethUnhexed, _ := hex.DecodeString("880ec53af800b5cd051531672ef4fc4de233bd5d")
		callData.From.SetBytes(ethUnhexed)
		callData.To, _ = data.NewAddressFromBech32String("erd1qqqqqqqqqqqqqpgqsudu3a3n9yu62k5qkgcpy4j9ywl2x2gl5smsy7t4uv")

		expectedString := "Eth address: 0x880EC53Af800b5Cd051531672EF4fc4De233bD5d, MvX address: erd1qqqqqqqqqqqqqpgqsudu3a3n9yu62k5qkgcpy4j9ywl2x2gl5smsy7t4uv, token: tkn, amount: 37, nonce: 1, type: 1, function: callMe, gas limit: 50000, no arguments"
		assert.Equal(t, expectedString, callData.String())
	})
}
