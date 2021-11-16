package bridge

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestActionID(t *testing.T) {
	t.Parallel()

	value := int64(15)
	id := ActionID(value)

	require.Equal(t, value, id.Int64())
	require.Equal(t, "15", id.String())
	require.Equal(t, []byte{15}, id.Bytes())
}
