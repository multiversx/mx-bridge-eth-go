package bridge

import (
	"errors"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDepositTransaction_String(t *testing.T) {
	t.Parallel()

	dt := &DepositTransaction{
		To:            "to",
		DisplayableTo: "displayableTo",
		From:          "from",
		TokenAddress:  "tokenAddress",
		Amount:        big.NewInt(1122),
		DepositNonce:  1133,
		BlockNonce:    1144,
		Status:        Rejected,
		Error:         errors.New("error"),
	}

	expected := "to: displayableTo, from: from, token address: tokenAddress, amount: 1122, deposit nonce: 1133, " +
		"block nonce: 1144, status: 4, error: error"

	require.Equal(t, expected, dt.String())
}

func TestDepositTransaction_Clone(t *testing.T) {
	t.Parallel()

	dt := &DepositTransaction{
		To:            "to",
		DisplayableTo: "displayableTo",
		From:          "from",
		TokenAddress:  "tokenAddress",
		Amount:        big.NewInt(1122),
		DepositNonce:  1133,
		BlockNonce:    1144,
		Status:        Rejected,
		Error:         errors.New("error"),
	}

	clonedDt := dt.Clone()

	assert.Equal(t, dt, clonedDt)
	assert.False(t, dt == clonedDt) // pointer testing
}

func TestBatch_SetStatusOnAllTransactions(t *testing.T) {
	t.Parallel()

	b := Batch{
		Transactions: []*DepositTransaction{
			{
				Status: 1,
			},
			{
				Status: 2,
			},
			{
				Status: 3,
			},
		},
	}

	err := errors.New("an error")
	b.SetStatusOnAllTransactions(Rejected, err)

	for _, dt := range b.Transactions {
		assert.Equal(t, Rejected, dt.Status)
		assert.Equal(t, err, dt.Error)
	}
}

func TestBatch_Clone(t *testing.T) {
	t.Parallel()

	t.Run("nil batch should return nil", func(t *testing.T) {
		var batch *Batch

		assert.Nil(t, batch.Clone())
	})
	t.Run("should clone deeply the batch", func(t *testing.T) {
		batch := &Batch{
			ID: 1132,
			Transactions: []*DepositTransaction{
				{
					To:            "to1",
					DisplayableTo: "to1",
					From:          "from1",
					TokenAddress:  "token1",
					Amount:        big.NewInt(132782),
					DepositNonce:  242343,
					BlockNonce:    22478,
					Status:        54,
					Error:         errors.New("error 1"),
				},
				{
					To:            "to2",
					DisplayableTo: "to2",
					From:          "from2",
					TokenAddress:  "token2",
					Amount:        big.NewInt(1322),
					DepositNonce:  2423,
					BlockNonce:    228,
					Status:        5,
					Error:         errors.New("error 2"),
				},
			},
		}

		clonedBatch := batch.Clone()
		assert.Equal(t, batch, clonedBatch)
		assert.False(t, batch == clonedBatch) // pointer testing
		for i, tx := range batch.Transactions {
			clonedTx := clonedBatch.Transactions[i]
			assert.Equal(t, clonedTx, tx)
			assert.False(t, tx == clonedTx) // pointer testing
		}
	})
}
