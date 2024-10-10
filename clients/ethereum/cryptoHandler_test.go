package ethereum

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

func TestNewCryptoHandler(t *testing.T) {
	t.Parallel()

	t.Run("invalid file should error", func(t *testing.T) {
		t.Parallel()

		handler, err := NewCryptoHandler("missing file")
		assert.Nil(t, handler)
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "open missing file: no such file or directory")
	})
	t.Run("invalid private key file", func(t *testing.T) {
		t.Parallel()

		handler, err := NewCryptoHandler("./testdata/nok-ethereum-key")
		assert.Nil(t, handler)
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "invalid hex data for private key")
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		handler, err := NewCryptoHandler("./testdata/ok-ethereum-key")
		assert.NotNil(t, handler)
		assert.Nil(t, err)
	})
}

func TestCryptoHandler_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var instance *cryptoHandler
	assert.True(t, instance.IsInterfaceNil())

	instance = &cryptoHandler{}
	assert.False(t, instance.IsInterfaceNil())
}

func TestCryptoHandler_Sign(t *testing.T) {
	t.Parallel()

	t.Run("test 1", func(t *testing.T) {
		expectedSig := "b556014dd984183e4662dc3204e522a5a92093fd6f64bb2da9c1b66b8d5ad12d774e05728b83c76bf09bb91af93ede4118f59aa949c7d02c86051dd0fa140c9900"
		msgHash := common.HexToHash("c99286352d865e33f1747761cbd440a7906b9bd8a5261cb6909e5ba18dd19b08")

		handler, _ := NewCryptoHandler("./testdata/ok-ethereum-key")
		sig, err := handler.Sign(msgHash)
		assert.Nil(t, err)
		assert.Equal(t, expectedSig, hex.EncodeToString(sig))
	})
	t.Run("test 2", func(t *testing.T) {
		expectedSig := "9abff5ecad356a82855f3ecc816cad5d19315ab812f1affeed7f8020accf01127d4c41ed56ff1b3053b64957a19aa1c6fd7dd1b5aa53065b0df231f517bfe89f01"
		msgHash := common.HexToHash("c99286352d865e33f1747761cbd440a7906b9bd8a5261cb6909e5ba18dd19b09")

		handler, _ := NewCryptoHandler("./testdata/ok-ethereum-key")
		sig, err := handler.Sign(msgHash)
		assert.Nil(t, err)
		assert.Equal(t, expectedSig, hex.EncodeToString(sig))
	})
}

func TestCryptoHandler_GetAddress(t *testing.T) {
	t.Parallel()

	handler, _ := NewCryptoHandler("./testdata/ok-ethereum-key")
	expectedAddress := common.HexToAddress("0x3FE464Ac5aa562F7948322F92020F2b668D543d8")

	assert.Equal(t, expectedAddress, handler.GetAddress())
}

func TestCryptoHandler_CreateKeyedTransactor(t *testing.T) {
	t.Parallel()

	t.Run("nil chain ID should error", func(t *testing.T) {
		t.Parallel()

		handler, _ := NewCryptoHandler("./testdata/ok-ethereum-key")
		opts, err := handler.CreateKeyedTransactor(nil)
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "no chain id specified")
		assert.Nil(t, opts)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		handler, _ := NewCryptoHandler("./testdata/ok-ethereum-key")
		opts, err := handler.CreateKeyedTransactor(big.NewInt(1))
		assert.Nil(t, err)
		assert.NotNil(t, opts)
	})
}
