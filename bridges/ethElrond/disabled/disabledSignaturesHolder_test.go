package disabled

import (
	"fmt"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/stretchr/testify/assert"
)

func TestDisabledSignaturesHolder_MethodsShouldNotPanic(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, fmt.Sprintf("should have not panicked %v", r))
		}
	}()

	disabled := NewDisabledSignaturesHolder()
	assert.False(t, check.IfNil(disabled))
	disabled.ClearStoredSignatures()

	sigs := disabled.Signatures(nil)
	assert.Empty(t, sigs)
}
