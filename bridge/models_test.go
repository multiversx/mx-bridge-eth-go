package bridge

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBatch_ResolveNewDeposits(t *testing.T) {
	t.Parallel()

	batch := &Batch{
		Transactions: []*DepositTransaction{
			{
				To: "to1",
			},
			{
				To: "to2",
			},
		},
		Statuses: make([]byte, 2),
	}

	for i := 0; i < 3; i++ {
		batch.ResolveNewDeposits(i)
		assert.Equal(t, 2, len(batch.Statuses))
	}

	batch.ResolveNewDeposits(3)
	assert.Equal(t, 3, len(batch.Statuses))
	assert.Equal(t, Rejected, batch.Statuses[2])
	assert.Equal(t, byte(0), batch.Statuses[0]+batch.Statuses[1])
}
