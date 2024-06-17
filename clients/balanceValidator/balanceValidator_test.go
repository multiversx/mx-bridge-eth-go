package balanceValidator

import (
	"context"
	"errors"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/multiversx/mx-bridge-eth-go/core/batchProcessor"
	"github.com/multiversx/mx-bridge-eth-go/testsCommon/bridge"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/stretchr/testify/assert"
)

// TODO(jls): use these in next PR
//var (
//	ethToken = common.BytesToAddress([]byte("eth token"))
//	mvxToken = []byte("mvx token")
//	amount   = big.NewInt(37)
//)

func createMockArgsBalanceValidator() ArgsBalanceValidator {
	return ArgsBalanceValidator{
		Log:              &testscommon.LoggerStub{},
		MultiversXClient: &bridge.MultiversXClientStub{},
		EthereumClient:   &bridge.EthereumClientStub{},
	}
}

type testConfiguration struct {
	isNativeOnEth      bool
	isMintBurnOnEth    bool
	totalBalancesOnEth *big.Int
	burnBalancesOnEth  *big.Int
	mintBalancesOnEth  *big.Int

	isNativeOnMvx      bool
	isMintBurnOnMvx    bool
	totalBalancesOnMvx *big.Int
	burnBalancesOnMvx  *big.Int
	mintBalancesOnMvx  *big.Int

	errorsOnCalls map[string]error

	ethToken  common.Address
	mvxToken  []byte
	amount    *big.Int
	direction batchProcessor.Direction
}

type testResult struct {
	checkRequiredBalanceOnEthCalled bool
	checkRequiredBalanceOnMvxCalled bool
	error                           error
}

func TestNewBalanceValidator(t *testing.T) {
	t.Parallel()

	t.Run("nil logger should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsBalanceValidator()
		args.Log = nil
		instance, err := NewBalanceValidator(args)
		assert.Nil(t, instance)
		assert.Equal(t, ErrNilLogger, err)
	})
	t.Run("nil MultiversX client should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsBalanceValidator()
		args.MultiversXClient = nil
		instance, err := NewBalanceValidator(args)
		assert.Nil(t, instance)
		assert.Equal(t, ErrNilMultiversXClient, err)
	})
	t.Run("nil Ethereum client should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsBalanceValidator()
		args.EthereumClient = nil
		instance, err := NewBalanceValidator(args)
		assert.Nil(t, instance)
		assert.Equal(t, ErrNilEthereumClient, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsBalanceValidator()
		instance, err := NewBalanceValidator(args)
		assert.NotNil(t, instance)
		assert.Nil(t, err)
	})
}

func TestBalanceValidator_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var instance *balanceValidator
	assert.True(t, instance.IsInterfaceNil())

	instance = &balanceValidator{}
	assert.False(t, instance.IsInterfaceNil())
}

func TestBridgeExecutor_CheckToken(t *testing.T) {
	t.Parallel()

	expectedError := errors.New("expected error")
	t.Run("unknown direction should error", func(t *testing.T) {
		t.Parallel()

		cfg := testConfiguration{
			direction: "",
		}
		result := validatorTester(cfg)
		assert.ErrorIs(t, result.error, ErrInvalidDirection)
	})
	t.Run("query operations error", func(t *testing.T) {
		t.Parallel()

		t.Run("on isMintBurnOnEthereum", func(t *testing.T) {
			cfg := testConfiguration{
				direction: batchProcessor.FromMultiversX,
				errorsOnCalls: map[string]error{
					"MintBurnTokensEth": expectedError,
				},
			}
			result := validatorTester(cfg)
			assert.Equal(t, expectedError, result.error)
			assert.True(t, result.checkRequiredBalanceOnEthCalled)
			assert.False(t, result.checkRequiredBalanceOnMvxCalled)
		})
		t.Run("on isMintBurnOnMultiversX", func(t *testing.T) {
			cfg := testConfiguration{
				direction: batchProcessor.ToMultiversX,
				errorsOnCalls: map[string]error{
					"IsMintBurnTokenMvx": expectedError,
				},
			}
			result := validatorTester(cfg)
			assert.Equal(t, expectedError, result.error)
			assert.False(t, result.checkRequiredBalanceOnEthCalled)
			assert.True(t, result.checkRequiredBalanceOnMvxCalled)
		})
		t.Run("on isNativeOnEthereum", func(t *testing.T) {
			cfg := testConfiguration{
				direction: batchProcessor.ToMultiversX,
				errorsOnCalls: map[string]error{
					"NativeTokensEth": expectedError,
				},
			}
			result := validatorTester(cfg)
			assert.Equal(t, expectedError, result.error)
			assert.False(t, result.checkRequiredBalanceOnEthCalled)
			assert.True(t, result.checkRequiredBalanceOnMvxCalled)
		})
		t.Run("on isNativeOnMultiversX", func(t *testing.T) {
			cfg := testConfiguration{
				direction: batchProcessor.FromMultiversX,
				errorsOnCalls: map[string]error{
					"IsNativeTokenMvx": expectedError,
				},
			}
			result := validatorTester(cfg)
			assert.Equal(t, expectedError, result.error)
			assert.True(t, result.checkRequiredBalanceOnEthCalled)
			assert.False(t, result.checkRequiredBalanceOnMvxCalled)
		})
		t.Run("on computeEthAmount, TotalBalances", func(t *testing.T) {
			cfg := testConfiguration{
				direction:       batchProcessor.FromMultiversX,
				isMintBurnOnMvx: true,
				isNativeOnEth:   true,
				errorsOnCalls: map[string]error{
					"TotalBalancesEth": expectedError,
				},
			}
			result := validatorTester(cfg)
			assert.Equal(t, expectedError, result.error)
			assert.True(t, result.checkRequiredBalanceOnEthCalled)
			assert.False(t, result.checkRequiredBalanceOnMvxCalled)
		})
		t.Run("on computeEthAmount, BurnBalances", func(t *testing.T) {
			cfg := testConfiguration{
				direction:       batchProcessor.FromMultiversX,
				isNativeOnMvx:   true,
				isMintBurnOnEth: true,
				errorsOnCalls: map[string]error{
					"BurnBalancesEth": expectedError,
				},
			}
			result := validatorTester(cfg)
			assert.Equal(t, expectedError, result.error)
			assert.True(t, result.checkRequiredBalanceOnEthCalled)
			assert.False(t, result.checkRequiredBalanceOnMvxCalled)
		})
		t.Run("on computeEthAmount, MintBalances", func(t *testing.T) {
			cfg := testConfiguration{
				direction:       batchProcessor.FromMultiversX,
				isNativeOnMvx:   true,
				isMintBurnOnEth: true,
				errorsOnCalls: map[string]error{
					"MintBalancesEth": expectedError,
				},
			}
			result := validatorTester(cfg)
			assert.Equal(t, expectedError, result.error)
			assert.True(t, result.checkRequiredBalanceOnEthCalled)
			assert.False(t, result.checkRequiredBalanceOnMvxCalled)
		})
		t.Run("on computeMvxAmount, TotalBalances", func(t *testing.T) {
			cfg := testConfiguration{
				direction:       batchProcessor.ToMultiversX,
				isNativeOnMvx:   true,
				isMintBurnOnEth: true,
				errorsOnCalls: map[string]error{
					"TotalBalancesMvx": expectedError,
				},
			}
			result := validatorTester(cfg)
			assert.Equal(t, expectedError, result.error)
			assert.False(t, result.checkRequiredBalanceOnEthCalled)
			assert.True(t, result.checkRequiredBalanceOnMvxCalled)
		})
		t.Run("on computeMvxAmount, BurnBalances", func(t *testing.T) {
			cfg := testConfiguration{
				direction:       batchProcessor.ToMultiversX,
				isMintBurnOnMvx: true,
				isNativeOnEth:   true,
				errorsOnCalls: map[string]error{
					"BurnBalancesMvx": expectedError,
				},
			}
			result := validatorTester(cfg)
			assert.Equal(t, expectedError, result.error)
			assert.False(t, result.checkRequiredBalanceOnEthCalled)
			assert.True(t, result.checkRequiredBalanceOnMvxCalled)
		})
		t.Run("on computeMvxAmount, MintBalances", func(t *testing.T) {
			cfg := testConfiguration{
				direction:       batchProcessor.ToMultiversX,
				isMintBurnOnMvx: true,
				isNativeOnEth:   true,
				errorsOnCalls: map[string]error{
					"MintBalancesMvx": expectedError,
				},
			}
			result := validatorTester(cfg)
			assert.Equal(t, expectedError, result.error)
			assert.False(t, result.checkRequiredBalanceOnEthCalled)
			assert.True(t, result.checkRequiredBalanceOnMvxCalled)
		})
	})
	t.Run("invalid setup", func(t *testing.T) {
		t.Parallel()

		t.Run("on Ethereum is not native nor is mint/burn, should error", func(t *testing.T) {
			cfg := testConfiguration{
				direction:       batchProcessor.ToMultiversX,
				isMintBurnOnMvx: true,
			}
			result := validatorTester(cfg)
			assert.ErrorIs(t, result.error, ErrInvalidSetup)
			assert.Contains(t, result.error.Error(), "isNativeOnEthereum = false, isMintBurnOnEthereum = false")
			assert.False(t, result.checkRequiredBalanceOnEthCalled)
			assert.True(t, result.checkRequiredBalanceOnMvxCalled)
		})
		t.Run("on MultiversX is not native nor is mint/burn, should error", func(t *testing.T) {
			cfg := testConfiguration{
				direction:     batchProcessor.ToMultiversX,
				isNativeOnEth: true,
			}
			result := validatorTester(cfg)
			assert.ErrorIs(t, result.error, ErrInvalidSetup)
			assert.Contains(t, result.error.Error(), "isNativeOnMultiversX = false, isMintBurnOnMultiversX = false")
			assert.False(t, result.checkRequiredBalanceOnEthCalled)
			assert.True(t, result.checkRequiredBalanceOnMvxCalled)
		})
		t.Run("native on both chains, should error", func(t *testing.T) {
			cfg := testConfiguration{
				direction:     batchProcessor.ToMultiversX,
				isNativeOnEth: true,
				isNativeOnMvx: true,
			}
			result := validatorTester(cfg)
			assert.ErrorIs(t, result.error, ErrInvalidSetup)
			assert.Contains(t, result.error.Error(), "isNativeOnEthereum = true, isNativeOnMultiversX = true")
			assert.False(t, result.checkRequiredBalanceOnEthCalled)
			assert.True(t, result.checkRequiredBalanceOnMvxCalled)
		})
	})
	t.Run("bad values on mint & burn", func(t *testing.T) {
		t.Parallel()

		t.Run("on Ethereum, native", func(t *testing.T) {
			t.Parallel()

			cfg := testConfiguration{
				direction:         batchProcessor.ToMultiversX,
				isMintBurnOnEth:   true,
				isNativeOnEth:     true,
				isMintBurnOnMvx:   true,
				burnBalancesOnEth: big.NewInt(37),
				mintBalancesOnEth: big.NewInt(38),
			}
			result := validatorTester(cfg)
			assert.ErrorIs(t, result.error, ErrNegativeAmount)
			assert.Contains(t, result.error.Error(), "ethAmount: -1")
			assert.False(t, result.checkRequiredBalanceOnEthCalled)
			assert.True(t, result.checkRequiredBalanceOnMvxCalled)
		})
		t.Run("on Ethereum, non-native", func(t *testing.T) {
			t.Parallel()

			cfg := testConfiguration{
				direction:         batchProcessor.ToMultiversX,
				isMintBurnOnEth:   true,
				isNativeOnMvx:     true,
				burnBalancesOnEth: big.NewInt(38),
				mintBalancesOnEth: big.NewInt(37),
			}
			result := validatorTester(cfg)
			assert.ErrorIs(t, result.error, ErrNegativeAmount)
			assert.Contains(t, result.error.Error(), "ethAmount: -1")
			assert.False(t, result.checkRequiredBalanceOnEthCalled)
			assert.True(t, result.checkRequiredBalanceOnMvxCalled)
		})
		t.Run("on MultiversX, native", func(t *testing.T) {
			t.Parallel()

			cfg := testConfiguration{
				direction:         batchProcessor.ToMultiversX,
				isMintBurnOnEth:   true,
				isMintBurnOnMvx:   true,
				isNativeOnMvx:     true,
				burnBalancesOnMvx: big.NewInt(37),
				mintBalancesOnMvx: big.NewInt(38),
			}
			result := validatorTester(cfg)
			assert.ErrorIs(t, result.error, ErrNegativeAmount)
			assert.Contains(t, result.error.Error(), "mvxAmount: -1")
			assert.False(t, result.checkRequiredBalanceOnEthCalled)
			assert.True(t, result.checkRequiredBalanceOnMvxCalled)
		})
		t.Run("on MultiversX, non-native", func(t *testing.T) {
			t.Parallel()

			cfg := testConfiguration{
				direction:         batchProcessor.ToMultiversX,
				isNativeOnEth:     true,
				isMintBurnOnMvx:   true,
				burnBalancesOnMvx: big.NewInt(38),
				mintBalancesOnMvx: big.NewInt(37),
			}
			result := validatorTester(cfg)
			assert.ErrorIs(t, result.error, ErrNegativeAmount)
			assert.Contains(t, result.error.Error(), "mvxAmount: -1")
			assert.False(t, result.checkRequiredBalanceOnEthCalled)
			assert.True(t, result.checkRequiredBalanceOnMvxCalled)
		})
	})
	// TODO(jls): add tests here for the correct balance computation when all pending batches are considered
}

func validatorTester(cfg testConfiguration) testResult {
	args := createMockArgsBalanceValidator()

	result := testResult{}

	args.MultiversXClient = &bridge.MultiversXClientStub{
		CheckRequiredBalanceCalled: func(ctx context.Context, token []byte, value *big.Int) error {
			result.checkRequiredBalanceOnMvxCalled = true
			return nil
		},
		IsMintBurnTokenCalled: func(ctx context.Context, token []byte) (bool, error) {
			err := cfg.errorsOnCalls["IsMintBurnTokenMvx"]
			if err != nil {
				return false, err
			}

			return cfg.isMintBurnOnMvx, nil
		},
		IsNativeTokenCalled: func(ctx context.Context, token []byte) (bool, error) {
			err := cfg.errorsOnCalls["IsNativeTokenMvx"]
			if err != nil {
				return false, err
			}

			return cfg.isNativeOnMvx, nil
		},
		TotalBalancesCalled: func(ctx context.Context, token []byte) (*big.Int, error) {
			err := cfg.errorsOnCalls["TotalBalancesMvx"]
			if err != nil {
				return nil, err
			}

			value := cfg.totalBalancesOnMvx
			if value == nil {
				value = big.NewInt(0)
			}

			return value, nil
		},
		MintBalancesCalled: func(ctx context.Context, token []byte) (*big.Int, error) {
			err := cfg.errorsOnCalls["MintBalancesMvx"]
			if err != nil {
				return nil, err
			}

			value := cfg.mintBalancesOnMvx
			if value == nil {
				value = big.NewInt(0)
			}

			return value, nil
		},
		BurnBalancesCalled: func(ctx context.Context, token []byte) (*big.Int, error) {
			err := cfg.errorsOnCalls["BurnBalancesMvx"]
			if err != nil {
				return nil, err
			}

			value := cfg.burnBalancesOnMvx
			if value == nil {
				value = big.NewInt(0)
			}

			return value, nil
		},
	}
	args.EthereumClient = &bridge.EthereumClientStub{
		CheckRequiredBalanceCalled: func(ctx context.Context, erc20Address common.Address, value *big.Int) error {
			result.checkRequiredBalanceOnEthCalled = true
			return nil
		},
		MintBurnTokensCalled: func(ctx context.Context, account common.Address) (bool, error) {
			err := cfg.errorsOnCalls["MintBurnTokensEth"]
			if err != nil {
				return false, err
			}

			return cfg.isMintBurnOnEth, nil
		},
		NativeTokensCalled: func(ctx context.Context, account common.Address) (bool, error) {
			err := cfg.errorsOnCalls["NativeTokensEth"]
			if err != nil {
				return false, err
			}

			return cfg.isNativeOnEth, nil
		},
		TotalBalancesCalled: func(ctx context.Context, account common.Address) (*big.Int, error) {
			err := cfg.errorsOnCalls["TotalBalancesEth"]
			if err != nil {
				return nil, err
			}

			value := cfg.totalBalancesOnEth
			if value == nil {
				value = big.NewInt(0)
			}

			return value, nil
		},
		MintBalancesCalled: func(ctx context.Context, account common.Address) (*big.Int, error) {
			err := cfg.errorsOnCalls["MintBalancesEth"]
			if err != nil {
				return nil, err
			}

			value := cfg.mintBalancesOnEth
			if value == nil {
				value = big.NewInt(0)
			}

			return value, nil
		},
		BurnBalancesCalled: func(ctx context.Context, account common.Address) (*big.Int, error) {
			err := cfg.errorsOnCalls["BurnBalancesEth"]
			if err != nil {
				return nil, err
			}

			value := cfg.burnBalancesOnEth
			if value == nil {
				value = big.NewInt(0)
			}

			return value, nil
		},
	}

	validator, err := NewBalanceValidator(args)
	if err != nil {
		result.error = err
		return result
	}

	result.error = validator.CheckToken(context.Background(), cfg.ethToken, cfg.mvxToken, cfg.amount, cfg.direction)

	return result
}
