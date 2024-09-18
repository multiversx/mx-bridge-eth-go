package batchProcessor

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	bridgeCore "github.com/multiversx/mx-bridge-eth-go/core"
	"github.com/stretchr/testify/assert"
)

func TestExtractListEthToMvx(t *testing.T) {
	t.Parallel()

	testBatch := &bridgeCore.TransferBatch{
		ID: 37,
		Deposits: []*bridgeCore.DepositTransfer{
			{
				Nonce:                 1,
				ToBytes:               []byte("to 1"),
				FromBytes:             []byte("from 1"),
				SourceTokenBytes:      []byte("source token 1"),
				DestinationTokenBytes: []byte("destination token 1"),
				Amount:                big.NewInt(11),
			},
			{
				Nonce:                 2,
				ToBytes:               []byte("to 2"),
				FromBytes:             []byte("from 2"),
				SourceTokenBytes:      []byte("source token 2"),
				DestinationTokenBytes: []byte("destination token 2"),
				Amount:                big.NewInt(22),
			},
		},
		Statuses: nil,
	}

	args := ExtractListEthToMvx(testBatch)

	expectedEthTokens := []common.Address{
		common.BytesToAddress([]byte("source token 1")),
		common.BytesToAddress([]byte("source token 2")),
	}
	assert.Equal(t, expectedEthTokens, args.EthTokens)

	expectedRecipients := []common.Address{
		common.BytesToAddress([]byte("to 1")),
		common.BytesToAddress([]byte("to 2")),
	}
	assert.Equal(t, expectedRecipients, args.Recipients)

	expectedMvxTokenBytes := [][]byte{
		[]byte("destination token 1"),
		[]byte("destination token 2"),
	}
	assert.Equal(t, expectedMvxTokenBytes, args.MvxTokenBytes)

	expectedAmounts := []*big.Int{
		big.NewInt(11),
		big.NewInt(22),
	}
	assert.Equal(t, expectedAmounts, args.Amounts)

	expectedNonces := []*big.Int{
		big.NewInt(1),
		big.NewInt(2),
	}
	assert.Equal(t, expectedNonces, args.Nonces)
}

func TestExtractListMvxToEth(t *testing.T) {
	t.Parallel()

	t.Run("invalid length for the From field should error", func(t *testing.T) {
		t.Parallel()

		testBatch := &bridgeCore.TransferBatch{
			ID: 37,
			Deposits: []*bridgeCore.DepositTransfer{
				{
					Nonce:                 1,
					ToBytes:               []byte("to 1"),
					FromBytes:             []byte("from 1"),
					SourceTokenBytes:      []byte("source token 1"),
					DestinationTokenBytes: []byte("destination token 1"),
					Amount:                big.NewInt(11),
				},
			},
			Statuses: nil,
		}

		args, err := ExtractListMvxToEth(testBatch)
		assert.Nil(t, args)
		assert.ErrorIs(t, err, errInternalErrorValidatingLength)
		assert.Contains(t, err.Error(), "expected 32, got 6")

		testBatch = &bridgeCore.TransferBatch{
			ID: 37,
			Deposits: []*bridgeCore.DepositTransfer{
				{
					Nonce:                 1,
					ToBytes:               []byte("to 1"),
					FromBytes:             bytes.Repeat([]byte("1"), 33),
					SourceTokenBytes:      []byte("source token 1"),
					DestinationTokenBytes: []byte("destination token 1"),
					Amount:                big.NewInt(11),
				},
			},
			Statuses: nil,
		}

		args, err = ExtractListMvxToEth(testBatch)
		assert.Nil(t, args)
		assert.ErrorIs(t, err, errInternalErrorValidatingLength)
		assert.Contains(t, err.Error(), "expected 32, got 33")
	})
	t.Run("should work", func(t *testing.T) {
		testBatch := &bridgeCore.TransferBatch{
			ID: 37,
			Deposits: []*bridgeCore.DepositTransfer{
				{
					Nonce:                 1,
					ToBytes:               []byte("to 1"),
					FromBytes:             bytes.Repeat([]byte("1"), 32),
					SourceTokenBytes:      []byte("source token 1"),
					DestinationTokenBytes: []byte("destination token 1"),
					Amount:                big.NewInt(11),
					Data:                  []byte("data1"),
				},
				{
					Nonce:                 2,
					ToBytes:               []byte("to 2"),
					FromBytes:             bytes.Repeat([]byte("2"), 32),
					SourceTokenBytes:      []byte("source token 2"),
					DestinationTokenBytes: []byte("destination token 2"),
					Amount:                big.NewInt(22),
					Data:                  []byte("data2"),
				},
			},
			Statuses: nil,
		}

		args, err := ExtractListMvxToEth(testBatch)
		assert.Nil(t, err)

		expectedEthTokens := []common.Address{
			common.BytesToAddress([]byte("destination token 1")),
			common.BytesToAddress([]byte("destination token 2")),
		}
		assert.Equal(t, expectedEthTokens, args.EthTokens)

		expectedRecipients := []common.Address{
			common.BytesToAddress([]byte("to 1")),
			common.BytesToAddress([]byte("to 2")),
		}
		assert.Equal(t, expectedRecipients, args.Recipients)

		expectedMvxTokenBytes := [][]byte{
			[]byte("source token 1"),
			[]byte("source token 2"),
		}
		assert.Equal(t, expectedMvxTokenBytes, args.MvxTokenBytes)

		expectedAmounts := []*big.Int{
			big.NewInt(11),
			big.NewInt(22),
		}
		assert.Equal(t, expectedAmounts, args.Amounts)

		expectedNonces := []*big.Int{
			big.NewInt(1),
			big.NewInt(2),
		}
		assert.Equal(t, expectedNonces, args.Nonces)

		sender1 := [32]byte(bytes.Repeat([]byte("1"), 32))
		sender2 := [32]byte(bytes.Repeat([]byte("2"), 32))

		expectedSenders := [][32]byte{sender1, sender2}
		assert.Equal(t, expectedSenders, args.Senders)

		expectedData := [][]byte{
			testBatch.Deposits[0].Data,
			testBatch.Deposits[1].Data,
		}
		assert.Equal(t, expectedData, args.ScCalls)
	})
}
