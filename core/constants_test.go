package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClientStatus(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "Available", Available.String())
	assert.Equal(t, "Unavailable", Unavailable.String())
	assert.Equal(t, "Invalid status 56", ClientStatus(56).String())
}
