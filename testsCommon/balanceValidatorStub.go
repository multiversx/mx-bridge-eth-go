package testsCommon

import (
	"context"
	"math/big"

	"github.com/multiversx/mx-bridge-eth-go/core/batchProcessor"
)

// BalanceValidatorStub -
type BalanceValidatorStub struct {
	CheckTokenCalled func(ctx context.Context, token []byte, mvxToken []byte, amount *big.Int, direction batchProcessor.Direction) error
}

// CheckToken -
func (stub *BalanceValidatorStub) CheckToken(ctx context.Context, token []byte, mvxToken []byte, amount *big.Int, direction batchProcessor.Direction) error {
	if stub.CheckTokenCalled != nil {
		return stub.CheckTokenCalled(ctx, token, mvxToken, amount, direction)
	}

	return nil
}

// IsInterfaceNil -
func (stub *BalanceValidatorStub) IsInterfaceNil() bool {
	return stub == nil
}
