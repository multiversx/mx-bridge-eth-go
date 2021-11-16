package bridge

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNonce(t *testing.T) {
	t.Parallel()

	value := int64(15)
	n := Nonce(value)

	require.Equal(t, value, n.Int64())
	require.Equal(t, "15", n.String())
	require.Equal(t, []byte{15}, n.Bytes())
}
