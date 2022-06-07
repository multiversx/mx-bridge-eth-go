package ethereum

import (
	"context"
	"errors"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-eth-bridge/clients"
	"github.com/ElrondNetwork/elrond-eth-bridge/testsCommon"
	bridgeTests "github.com/ElrondNetwork/elrond-eth-bridge/testsCommon/bridge"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ethereum/go-ethereum"
	"github.com/stretchr/testify/assert"
)

func createMockArgsContractsHolder() ArgsErc20SafeContractsHolder {

	args := ArgsErc20SafeContractsHolder{
		EthClient:              &bridgeTests.ContractBackendStub{},
		EthClientStatusHandler: &testsCommon.StatusHandlerStub{},
	}

	return args
}

func TestNewErc20SafeContractsHolder(t *testing.T) {
	t.Parallel()

	t.Run("nil EthClient", func(t *testing.T) {
		args := createMockArgsContractsHolder()
		args.EthClient = nil

		ch, err := NewErc20SafeContractsHolder(args)
		assert.Nil(t, ch)
		assert.Equal(t, errNilEthClient, err)
	})
	t.Run("nil status handler", func(t *testing.T) {
		args := createMockArgsContractsHolder()
		args.EthClientStatusHandler = nil

		ch, err := NewErc20SafeContractsHolder(args)
		assert.Nil(t, ch)
		assert.Equal(t, clients.ErrNilStatusHandler, err)
	})
	t.Run("should work", func(t *testing.T) {
		args := createMockArgsContractsHolder()

		ch, err := NewErc20SafeContractsHolder(args)
		assert.Nil(t, err)
		assert.False(t, check.IfNil(ch))
		assert.Equal(t, 0, len(ch.contracts))
	})
}

func TestBalanceOf(t *testing.T) {
	t.Parallel()

	t.Run("address does not exists on map nor blockchain", func(t *testing.T) {
		expectedError := errors.New("no contract code at given address")
		args := createMockArgsContractsHolder()
		args.EthClient = &bridgeTests.ContractBackendStub{
			CallContractCalled: func(ctx context.Context, call ethereum.CallMsg, blockNumber *big.Int) ([]byte, error) {
				return nil, expectedError
			},
		}
		ch, err := NewErc20SafeContractsHolder(args)
		assert.Nil(t, err)
		assert.False(t, check.IfNil(ch))
		assert.Equal(t, 0, len(ch.contracts))

		result, err := ch.BalanceOf(context.Background(), testsCommon.CreateRandomEthereumAddress(), testsCommon.CreateRandomEthereumAddress())
		assert.Equal(t, expectedError, err)
		assert.Nil(t, result)
		assert.Equal(t, 1, len(ch.contracts))
	})
	t.Run("address exists only on blockchain", func(t *testing.T) {
		var returnedBalance int64 = 1000
		args := createMockArgsContractsHolder()
		args.EthClient = &bridgeTests.ContractBackendStub{
			CallContractCalled: func(ctx context.Context, call ethereum.CallMsg, blockNumber *big.Int) ([]byte, error) {
				return convertBigToAbiCompatible(big.NewInt(returnedBalance)), nil
			},
		}
		ch, err := NewErc20SafeContractsHolder(args)
		assert.Nil(t, err)
		assert.False(t, check.IfNil(ch))
		assert.Equal(t, 0, len(ch.contracts))

		result, err := ch.BalanceOf(context.Background(), testsCommon.CreateRandomEthereumAddress(), testsCommon.CreateRandomEthereumAddress())
		assert.Nil(t, err)
		assert.Equal(t, big.NewInt(returnedBalance), result)
		assert.Equal(t, 1, len(ch.contracts))
	})
	t.Run("address exists also in contracts map", func(t *testing.T) {
		var returnedBalance int64 = 1000
		args := createMockArgsContractsHolder()
		args.EthClient = &bridgeTests.ContractBackendStub{
			CallContractCalled: func(ctx context.Context, call ethereum.CallMsg, blockNumber *big.Int) ([]byte, error) {
				return convertBigToAbiCompatible(big.NewInt(returnedBalance)), nil
			},
		}
		ch, err := NewErc20SafeContractsHolder(args)
		contractAddress := testsCommon.CreateRandomEthereumAddress()
		address1 := testsCommon.CreateRandomEthereumAddress()
		address2 := testsCommon.CreateRandomEthereumAddress()
		assert.Nil(t, err)
		assert.False(t, check.IfNil(ch))
		assert.Equal(t, 0, len(ch.contracts))

		result, err := ch.BalanceOf(context.Background(), contractAddress, address1)
		// first time the contract does not exists in the map, so it should add it
		assert.Nil(t, err)
		assert.Equal(t, big.NewInt(returnedBalance), result)
		assert.Equal(t, 1, len(ch.contracts))

		result, err = ch.BalanceOf(context.Background(), contractAddress, address2)
		// second time the contract already exists in the map, so it should just use it
		assert.Nil(t, err)
		assert.Equal(t, big.NewInt(returnedBalance), result)
		assert.Equal(t, 1, len(ch.contracts))
	})

	t.Run("new contract addres while another contracts already exists", func(t *testing.T) {
		var returnedBalance int64 = 1000
		args := createMockArgsContractsHolder()
		args.EthClient = &bridgeTests.ContractBackendStub{
			CallContractCalled: func(ctx context.Context, call ethereum.CallMsg, blockNumber *big.Int) ([]byte, error) {
				return convertBigToAbiCompatible(big.NewInt(returnedBalance)), nil
			},
		}
		ch, err := NewErc20SafeContractsHolder(args)
		contractAddress1 := testsCommon.CreateRandomEthereumAddress()
		contractAddress2 := testsCommon.CreateRandomEthereumAddress()
		address1 := testsCommon.CreateRandomEthereumAddress()
		address2 := testsCommon.CreateRandomEthereumAddress()
		assert.Nil(t, err)
		assert.False(t, check.IfNil(ch))
		assert.Equal(t, 0, len(ch.contracts))

		result, err := ch.BalanceOf(context.Background(), contractAddress1, address1)
		assert.Nil(t, err)
		assert.Equal(t, big.NewInt(returnedBalance), result)
		assert.Equal(t, 1, len(ch.contracts))

		result, err = ch.BalanceOf(context.Background(), contractAddress1, address2)
		assert.Nil(t, err)
		assert.Equal(t, big.NewInt(returnedBalance), result)
		assert.Equal(t, 1, len(ch.contracts))

		result, err = ch.BalanceOf(context.Background(), contractAddress2, address2)
		assert.Nil(t, err)
		assert.Equal(t, big.NewInt(returnedBalance), result)
		assert.Equal(t, 2, len(ch.contracts))

		result, err = ch.BalanceOf(context.Background(), contractAddress2, address1)
		assert.Nil(t, err)
		assert.Equal(t, big.NewInt(returnedBalance), result)
		assert.Equal(t, 2, len(ch.contracts))
	})
}

func convertBigToAbiCompatible(number *big.Int) []byte {
	numberAsBytes := number.Bytes()
	size := len(numberAsBytes)
	sizeBuffer := size + 32 - size%32
	bs := make([]byte, sizeBuffer)
	for i := 0; i < size; i++ {
		bs[sizeBuffer-i-1] = numberAsBytes[size-i-1]
	}
	return bs
}
