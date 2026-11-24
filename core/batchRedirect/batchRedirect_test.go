package batchRedirect

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	bridgeCore "github.com/multiversx/mx-bridge-eth-go/core"
	sdkData "github.com/multiversx/mx-sdk-go/data"
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

func batch6185() *bridgeCore.TransferBatch {
	otherSender := bytes.Repeat([]byte{0x11}, 32)

	redirected := deposit(7961, batch6185Recipient, usdcMicro(8_853_225_198))
	redirected.FromBytes = batch6185Sender

	otherUSDC := deposit(7962, common.HexToAddress("0x7f341040a9708bdbbecffa1ac997653447bdbe1b").Bytes(), usdcMicro(120_692_637))
	otherUSDC.FromBytes = otherSender

	otherTADA := deposit(7964, common.HexToAddress("0x97ce2c128ce3cb2705ead8d1df95528a3cf1476e").Bytes(), usdc(964_222))
	otherTADA.FromBytes = otherSender
	otherTADA.SourceTokenBytes = []byte("BSCTADA-f64757")
	otherTADA.DisplayableToken = "BSCTADA-f64757"
	otherTADA.DestinationTokenBytes = common.HexToAddress("0x9b26e318bc6a2c8b45f5daea2cc14697e0e0f8b5").Bytes()

	return &bridgeCore.TransferBatch{
		ID:       6185,
		Deposits: []*bridgeCore.DepositTransfer{redirected, otherUSDC, otherTADA},
		Statuses: make([]byte, 3),
	}
}

func TestRedirects_Batch6185EntryMatchesTheMainnetDeposit(t *testing.T) {
	t.Parallel()

	sender, err := sdkData.NewAddressFromBech32String("erd168swpgnf4ecqg58gmk395n33mfs4fuap7h36md8lawzkat68kggqngyvpm")
	require.Nil(t, err)

	entry := Redirects[len(Redirects)-1]
	assert.Equal(t, uint64(6185), entry.BatchID)
	assert.Equal(t, uint64(7961), entry.DepositNonce)
	assert.Equal(t, sender.AddressBytes(), entry.From)
	assert.Equal(t, common.HexToAddress("0x3099ba48d4daf3fad0b874b40f1bb55f8a6b4107").Bytes(), entry.OriginalTo)
	assert.Equal(t, "BSCUSDC-887875", entry.SourceToken)
	assert.Equal(t, "8853225198000000000000", entry.Amount.String())
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

	for _, invalid := range []string{"0x1234", "not-hex", "0x0000000000000000000000000000000000000000", "0x717fb1435f5fa0cc1c7e541e5d92ef71342fcd41", "0x3099ba48d4daf3fad0b874b40f1bb55f8a6b4107"} {
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
	t.Run("redirects only deposit 7961 of batch 6185", func(t *testing.T) {
		t.Parallel()

		redirector, _ := NewRedirector(testRecovery.Hex())
		input := batch6185()
		result, numRedirected, err := redirector.Apply(input)
		require.Nil(t, err)
		assert.Equal(t, 1, numRedirected)
		require.Len(t, result.Deposits, 3)
		for i, before := range input.Deposits {
			t.Logf("nonce %d %s amount %s: %s -> %s", before.Nonce, before.DisplayableToken, before.Amount.String(),
				common.BytesToAddress(before.ToBytes).Hex(), common.BytesToAddress(result.Deposits[i].ToBytes).Hex())
		}

		assert.Equal(t, testRecovery.Bytes(), result.Deposits[0].ToBytes)
		assert.Equal(t, testRecovery.Hex(), result.Deposits[0].DisplayableTo)
		assert.Equal(t, uint64(7961), result.Deposits[0].Nonce)
		assert.Equal(t, input.Deposits[0].Amount, result.Deposits[0].Amount)
		assert.Equal(t, input.Deposits[0].SourceTokenBytes, result.Deposits[0].SourceTokenBytes)
		assert.Equal(t, input.Deposits[0].DestinationTokenBytes, result.Deposits[0].DestinationTokenBytes)
		assert.Nil(t, result.Deposits[0].Data)
		assert.Equal(t, batch6185Recipient, input.Deposits[0].ToBytes)

		assert.Same(t, input.Deposits[1], result.Deposits[1])
		assert.Same(t, input.Deposits[2], result.Deposits[2])
		assert.Equal(t, input.ID, result.ID)
		assert.Equal(t, input.Statuses, result.Statuses)
	})
	t.Run("refuses batch 6185 when no recovery address is configured", func(t *testing.T) {
		t.Parallel()

		redirector, _ := NewRedirector("")
		_, _, err := redirector.Apply(batch6185())
		assert.ErrorIs(t, err, ErrRecoveryAddressNotSet)
	})
	t.Run("refuses batch 6185 when a deposit pays the redirected recipient with different data", func(t *testing.T) {
		t.Parallel()

		redirector, _ := NewRedirector(testRecovery.Hex())

		batch := batch6185()
		batch.Deposits[0].Amount = usdcMicro(8_853_225_199)
		_, _, err := redirector.Apply(batch)
		assert.ErrorIs(t, err, ErrUnexpectedDeposit)

		batch = batch6185()
		batch.Deposits[0].FromBytes = bytes.Repeat([]byte{0x22}, 32)
		_, _, err = redirector.Apply(batch)
		assert.ErrorIs(t, err, ErrUnexpectedDeposit)

		batch = batch6185()
		extra := deposit(7977, batch6185Recipient, usdc(1))
		extra.FromBytes = batch6185Sender
		batch.Deposits = append(batch.Deposits, extra)
		_, _, err = redirector.Apply(batch)
		assert.ErrorIs(t, err, ErrUnexpectedDeposit)
	})
	t.Run("does not redirect the batch 6185 deposit when it appears in another batch", func(t *testing.T) {
		t.Parallel()

		other := batch6185()
		other.ID = 6186
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
