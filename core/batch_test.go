package core

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDepositTransfer_Clone(t *testing.T) {
	t.Parallel()

	dt := &DepositTransfer{
		Nonce:                 112334,
		ToBytes:               []byte("to"),
		DisplayableTo:         "to",
		FromBytes:             []byte("from"),
		DisplayableFrom:       "from",
		SourceTokenBytes:      []byte("source token"),
		DisplayableToken:      "token",
		Amount:                big.NewInt(7463),
		DestinationTokenBytes: []byte("destination token"),
		Data:                  []byte("tx data"),
	}

	cloned := dt.Clone()

	assert.Equal(t, dt, cloned)
	assert.False(t, dt == cloned) // pointer testing
}

func TestDepositTransfer_String(t *testing.T) {
	t.Parallel()

	dt := &DepositTransfer{
		Nonce:            112334,
		ToBytes:          []byte("to"),
		DisplayableTo:    "to",
		FromBytes:        []byte("from"),
		DisplayableFrom:  "from",
		SourceTokenBytes: []byte("source token"),
		DisplayableToken: "token",
		Amount:           big.NewInt(7463),
	}

	expectedString := "to: to, from: from, token address: token, amount: 7463, deposit nonce: 112334, data: "
	assert.Equal(t, expectedString, dt.String())
}

func TestTransferBatch_Clone(t *testing.T) {
	t.Parallel()

	tb := &TransferBatch{
		ID: 2243,
		Deposits: []*DepositTransfer{
			{
				Nonce:                 1,
				ToBytes:               []byte("to1"),
				DisplayableTo:         "to1",
				FromBytes:             []byte("from1"),
				DisplayableFrom:       "from1",
				SourceTokenBytes:      []byte("source token1"),
				DisplayableToken:      "token1",
				Amount:                big.NewInt(3344),
				DestinationTokenBytes: []byte("destination token1"),
				Data:                  []byte("tx data"),
			},
			{
				Nonce:                 2,
				ToBytes:               []byte("to2"),
				DisplayableTo:         "to2",
				FromBytes:             []byte("from2"),
				DisplayableFrom:       "from2",
				SourceTokenBytes:      []byte("source token2"),
				DisplayableToken:      "token2",
				Amount:                big.NewInt(5566),
				DestinationTokenBytes: []byte("destination token2"),
				Data:                  []byte("tx data"),
			},
		},
		Statuses: []byte{Executed, Rejected},
	}

	cloned := tb.Clone()

	assert.Equal(t, tb, cloned)
	assert.False(t, tb == cloned) // pointer testing
}

func TestTransferBatch_String(t *testing.T) {
	t.Parallel()

	tb := &TransferBatch{
		ID: 2243,
		Deposits: []*DepositTransfer{
			{
				Nonce:            1,
				ToBytes:          []byte("to1"),
				DisplayableTo:    "to1",
				FromBytes:        []byte("from1"),
				DisplayableFrom:  "from1",
				SourceTokenBytes: []byte("source token1"),
				DisplayableToken: "token1",
				Amount:           big.NewInt(3344),
			},
			{
				Nonce:            2,
				ToBytes:          []byte("to2"),
				DisplayableTo:    "to2",
				FromBytes:        []byte("from2"),
				DisplayableFrom:  "from2",
				SourceTokenBytes: []byte("source token2"),
				DisplayableToken: "token2",
				Amount:           big.NewInt(5566),
			},
		},
		Statuses: []byte{Executed, Rejected},
	}

	expectedString := `Batch id 2243:
  to: to1, from: from1, token address: token1, amount: 3344, deposit nonce: 1, data: 
  to: to2, from: from2, token address: token2, amount: 5566, deposit nonce: 2, data: 
Statuses: 0304`
	assert.Equal(t, expectedString, tb.String())
}

func TestTransferBatch_ResolveNewDeposits(t *testing.T) {
	t.Parallel()

	batch := &TransferBatch{
		Deposits: []*DepositTransfer{
			{
				DisplayableTo: "to1",
			},
			{
				DisplayableTo: "to2",
			},
		},
		Statuses: make([]byte, 2),
	}

	t.Run("less new deposits", func(t *testing.T) {
		t.Parallel()

		workingBatch := batch.Clone()
		workingBatch.ResolveNewDeposits(0)
		assert.Equal(t, []byte{Rejected, Rejected}, workingBatch.Statuses)

		workingBatch = batch.Clone()
		workingBatch.ResolveNewDeposits(1)
		assert.Equal(t, []byte{0, Rejected}, workingBatch.Statuses)
	})
	t.Run("equal new deposits", func(t *testing.T) {
		t.Parallel()

		workingBatch := batch.Clone()
		workingBatch.ResolveNewDeposits(2)
		assert.Equal(t, []byte{0, 0}, workingBatch.Statuses)
	})
	t.Run("more new deposits", func(t *testing.T) {
		t.Parallel()

		workingBatch := batch.Clone()
		workingBatch.ResolveNewDeposits(3)
		assert.Equal(t, []byte{0, 0, Rejected}, workingBatch.Statuses)
	})
}
