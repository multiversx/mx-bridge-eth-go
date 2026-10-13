package batchRedirect

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	bridgeCore "github.com/multiversx/mx-bridge-eth-go/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	testRecovery = common.HexToAddress("0x1234567890abcdef1234567890abcdef12345678")
	bscUSDC      = common.HexToAddress("0x8ac76a51cc950d9822d68b83fe1ad97b32cd580d")
)

func deposit(nonce uint64, to []byte, amount *big.Int) *bridgeCore.DepositTransfer {
	return &bridgeCore.DepositTransfer{
		Nonce:                 nonce,
		FromBytes:             redirectedSender,
		ToBytes:               to,
		DisplayableTo:         common.BytesToAddress(to).Hex(),
		SourceTokenBytes:      []byte("BSCUSDC-887875"),
		DestinationTokenBytes: bscUSDC.Bytes(),
		DisplayableToken:      "BSCUSDC-887875",
		Amount:                amount,
	}
}

func batch5956() *bridgeCore.TransferBatch {
	return &bridgeCore.TransferBatch{
		ID:       5956,
		Deposits: []*bridgeCore.DepositTransfer{deposit(7717, redirectedRecipient, usdc(49_999)), deposit(7718, redirectedRecipient, usdc(30_440))},
		Statuses: make([]byte, 2),
	}
}

func TestNewRedirector(t *testing.T) {
	t.Parallel()

	for _, value := range []string{"", "  "} {
		redirector, err := NewRedirector(value)
		require.Nil(t, err)
		assert.Equal(t, "", redirector.RecoveryAddress())
	}

	redirector, err := NewRedirector(testRecovery.Hex())
	require.Nil(t, err)
	assert.Equal(t, testRecovery.Hex(), redirector.RecoveryAddress())

	redirector, err = NewRedirector("1234567890abcdef1234567890abcdef12345678")
	require.Nil(t, err)
	assert.Equal(t, testRecovery.Hex(), redirector.RecoveryAddress())

	for _, invalid := range []string{"0x1234", "not-hex", "0x0000000000000000000000000000000000000000", "0x717fb1435f5fa0cc1c7e541e5d92ef71342fcd41"} {
		_, err = NewRedirector(invalid)
		assert.ErrorIs(t, err, ErrInvalidRecoveryAddress, invalid)
	}
}

func TestRedirector_Apply(t *testing.T) {
	t.Parallel()

	t.Run("redirects both listed deposits without touching the input batch", func(t *testing.T) {
		t.Parallel()

		redirector, _ := NewRedirector(testRecovery.Hex())
		input := batch5956()
		result, numRedirected, err := redirector.Apply(input)
		require.Nil(t, err)
		assert.Equal(t, 2, numRedirected)
		for i, dt := range result.Deposits {
			assert.Equal(t, testRecovery.Bytes(), dt.ToBytes)
			assert.Equal(t, testRecovery.Hex(), dt.DisplayableTo)
			assert.Equal(t, input.Deposits[i].Amount, dt.Amount)
			assert.Equal(t, input.Deposits[i].Nonce, dt.Nonce)
			assert.Nil(t, dt.Data)
			assert.Equal(t, redirectedRecipient, input.Deposits[i].ToBytes)
		}
		assert.Equal(t, input.ID, result.ID)
		assert.Equal(t, input.Statuses, result.Statuses)
	})
	t.Run("refuses batch 5956 when no recovery address is configured", func(t *testing.T) {
		t.Parallel()

		redirector, _ := NewRedirector("")
		_, _, err := redirector.Apply(batch5956())
		assert.ErrorIs(t, err, ErrRecoveryAddressNotSet)

		var nilRedirector *Redirector
		_, _, err = nilRedirector.Apply(batch5956())
		assert.ErrorIs(t, err, ErrRecoveryAddressNotSet)
	})
	t.Run("refuses batch 5956 when a deposit pays the redirected recipient with different data", func(t *testing.T) {
		t.Parallel()

		redirector, _ := NewRedirector(testRecovery.Hex())
		batch := batch5956()
		batch.Deposits[1].Amount = usdc(30_441)
		_, _, err := redirector.Apply(batch)
		assert.ErrorIs(t, err, ErrUnexpectedDeposit)

		batch = batch5956()
		batch.Deposits = append(batch.Deposits, deposit(7719, redirectedRecipient, usdc(1)))
		_, _, err = redirector.Apply(batch)
		assert.ErrorIs(t, err, ErrUnexpectedDeposit)
	})
	t.Run("leaves other batches unchanged, with or without a recovery address", func(t *testing.T) {
		t.Parallel()

		other := batch5956()
		other.ID = 5957
		user := common.HexToAddress("0x00000000000000000000000000000000000000aa").Bytes()
		other.Deposits = append(other.Deposits, deposit(7720, user, usdc(10)))
		for _, configured := range []string{"", testRecovery.Hex()} {
			redirector, _ := NewRedirector(configured)
			result, numRedirected, err := redirector.Apply(other)
			require.Nil(t, err)
			assert.Equal(t, 0, numRedirected)
			assert.Same(t, other, result)
		}
	})
	t.Run("nil batch", func(t *testing.T) {
		t.Parallel()

		redirector, _ := NewRedirector("")
		result, numRedirected, err := redirector.Apply(nil)
		assert.Nil(t, err)
		assert.Nil(t, result)
		assert.Equal(t, 0, numRedirected)
	})
}
