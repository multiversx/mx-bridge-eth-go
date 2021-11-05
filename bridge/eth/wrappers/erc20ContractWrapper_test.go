package wrappers

import (
	"context"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-eth-bridge/core"
	"github.com/ElrondNetwork/elrond-eth-bridge/testsCommon"
	"github.com/ElrondNetwork/elrond-eth-bridge/testsCommon/interactors"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

func createMockArgsErc20ContractWrapper() (ArgsErc20ContractWrapper, *testsCommon.StatusHandlerMock) {
	statusHandler := testsCommon.NewStatusHandlerMock("mock")

	return ArgsErc20ContractWrapper{
		Erc20Contract: &interactors.GenericErc20ContractStub{},
		StatusHandler: statusHandler,
	}, statusHandler
}

func TestNewErc20ContractWrapper(t *testing.T) {
	t.Parallel()

	t.Run("erc20 contract is nil", func(t *testing.T) {
		args, _ := createMockArgsErc20ContractWrapper()
		args.Erc20Contract = nil

		wrapper, err := NewErc20ContractWrapper(args)
		assert.Equal(t, ErrNilErc20Contract, err)
		assert.True(t, check.IfNil(wrapper))
	})
	t.Run("nil status handler", func(t *testing.T) {
		args, _ := createMockArgsErc20ContractWrapper()
		args.StatusHandler = nil

		wrapper, err := NewErc20ContractWrapper(args)
		assert.True(t, check.IfNil(wrapper))
		assert.Equal(t, ErrNilStatusHandler, err)
	})
	t.Run("should work", func(t *testing.T) {
		args, _ := createMockArgsErc20ContractWrapper()

		wrapper, err := NewErc20ContractWrapper(args)
		assert.False(t, check.IfNil(wrapper))
		assert.Nil(t, err)
	})
}

func TestErc20ContractWrapper_BalanceOf(t *testing.T) {
	t.Parallel()

	args, statusHandler := createMockArgsErc20ContractWrapper()
	handlerCalled := false
	args.Erc20Contract = &interactors.GenericErc20ContractStub{
		BalanceOfCalled: func(account common.Address) (*big.Int, error) {
			handlerCalled = true
			return nil, nil
		},
	}
	wrapper, _ := NewErc20ContractWrapper(args)
	balance, err := wrapper.BalanceOf(context.TODO(), common.Address{})
	assert.Nil(t, err)
	assert.Nil(t, balance)
	assert.True(t, handlerCalled)
	assert.Equal(t, 1, statusHandler.GetIntMetric(core.MetricNumEthClientRequests))
}
