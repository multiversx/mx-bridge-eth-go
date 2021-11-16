package bridge

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBatchID(t *testing.T) {
	t.Parallel()

	value := int64(15)
	b := BatchID(15)

	require.Equal(t, value, b.Int64())
	require.Equal(t, "15", b.String())
	require.Equal(t, []byte{15}, b.Bytes())
}
