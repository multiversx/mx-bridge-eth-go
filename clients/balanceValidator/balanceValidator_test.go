package balanceValidator

import (
	"context"
	"errors"
	"math/big"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/multiversx/mx-bridge-eth-go/core/batchProcessor"
	"github.com/multiversx/mx-bridge-eth-go/testsCommon/bridge"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/stretchr/testify/assert"
)

var (
	ethToken = common.BytesToAddress([]byte("eth token"))
	mvxToken = []byte("mvx token")
	amount   = big.NewInt(37)
)

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

		t.Run("on ethereum, native", func(t *testing.T) {
			t.Parallel()

		})
	})

	//direction := batchProcessor.FromMultiversX
	//var ethIsNative, ethIsMintBurn, mvxIsNative, mvxIsMintBurn bool
	//t.Run("from MultiversX", func(t *testing.T) {
	//	ethIsNative = false
	//	t.Run("ethIsNative = false", func(t *testing.T) {
	//		ethIsMintBurn = false
	//		t.Run("ethIsMintBurn = false", func(t *testing.T) {
	//			// mvxIsNative & mvxIsMintBurn does not matter
	//			testBridge(ethIsMintBurn, ethIsNative, mvxIsMintBurn, mvxIsNative, zero, zero, zero, zero, zero, zero, direction, ErrInvalidSetup, []string{"isNativeOnEthereum", "isMintBurnOnEthereum"})(t)
	//		})
	//
	//		ethIsMintBurn = true
	//		t.Run("ethIsMintBurn = true", func(t *testing.T) {
	//			mvxIsNative = false
	//			t.Run("mvxIsNative = false", func(t *testing.T) {
	//				mvxIsMintBurn = false
	//				t.Run("mvxIsMintBurn = false", func(t *testing.T) {
	//					testBridge(ethIsMintBurn, ethIsNative, mvxIsMintBurn, mvxIsNative, zero, zero, zero, zero, zero, zero, direction, ErrInvalidSetup, []string{"isNativeOnMultiversX", "isMintBurnOnMultiversX"})(t)
	//				})
	//
	//				mvxIsMintBurn = true
	//				t.Run("mvxIsMintBurn = true", func(t *testing.T) {
	//					testBridge(ethIsMintBurn, ethIsNative, mvxIsMintBurn, mvxIsNative, zero, zero, zero, zero, zero, zero, direction, ErrInvalidSetup, []string{"isNativeOnEthereum", "isNativeOnMultiversX"})(t)
	//				})
	//			})
	//
	//			mvxIsNative = true
	//			//TODO(jls): fix this
	//			//t.Run("mvxIsNative = true", func(t *testing.T) {
	//			//	mvxIsMintBurn = false
	//			//	t.Run("mvxIsMintBurn = false", func(t *testing.T) {
	//			//		ethMintBalance := big.NewInt(0)
	//			//		ethBurnBalance := big.NewInt(0)
	//			//		mvxTotalBalances := big.NewInt(99)
	//			//
	//			//		testBridge(ethIsMintBurn, ethIsNative, mvxIsMintBurn, mvxIsNative, zero, ethMintBalance, ethBurnBalance, mvxTotalBalances, zero, zero, direction, ErrBalanceMismatch, []string{})(t)
	//			//		mvxTotalBalances = big.NewInt(100)
	//			//		testBridge(ethIsMintBurn, ethIsNative, mvxIsMintBurn, mvxIsNative, zero, ethMintBalance, ethBurnBalance, mvxTotalBalances, zero, zero, direction, nil, []string{})(t)
	//			//	})
	//			//	mvxIsMintBurn = true
	//			//	t.Run("mvxIsMintBurn = true", func(t *testing.T) {
	//			//		ethMintBalance := big.NewInt(0)
	//			//		ethBurnBalance := big.NewInt(0)
	//			//		mvxMintBalance := big.NewInt(0)
	//			//		mvxBurnBalance := big.NewInt(99)
	//			//
	//			//		testBridge(ethIsMintBurn, ethIsNative, mvxIsMintBurn, mvxIsNative, zero, ethMintBalance, ethBurnBalance, zero, mvxMintBalance, mvxBurnBalance, direction, ErrBalanceMismatch, []string{})(t)
	//			//		mvxBurnBalance = big.NewInt(100)
	//			//		testBridge(ethIsMintBurn, ethIsNative, mvxIsMintBurn, mvxIsNative, zero, ethMintBalance, ethBurnBalance, zero, mvxMintBalance, mvxBurnBalance, direction, nil, []string{})(t)
	//			//	})
	//			//})
	//		})
	//	})
	//
	//	ethIsNative = true
	//	t.Run("ethIsNative = true", func(t *testing.T) {
	//		ethIsMintBurn = false
	//		t.Run("ethIsMintBurn = false", func(t *testing.T) {
	//			mvxIsNative = false
	//			t.Run("mvxIsNative = false", func(t *testing.T) {
	//				mvxIsMintBurn = false
	//				t.Run("mvxIsMintBurn = false", func(t *testing.T) {
	//					testBridge(ethIsMintBurn, ethIsNative, mvxIsMintBurn, mvxIsNative, zero, zero, zero, zero, zero, zero, direction, ErrInvalidSetup, []string{"isNativeOnMultiversX", "isMintBurnOnMultiversX"})(t)
	//				})
	//				mvxIsMintBurn = true
	//				//TODO(jls): fix this
	//				//t.Run("mvxIsMintBurn = true", func(t *testing.T) {
	//				//	ethTotalBalances := big.NewInt(100)
	//				//	mvxMintBalance := big.NewInt(100)
	//				//	mvxBurnBalance := big.NewInt(99)
	//				//
	//				//	testBridge(ethIsMintBurn, ethIsNative, mvxIsMintBurn, mvxIsNative, ethTotalBalances, zero, zero, zero, mvxMintBalance, mvxBurnBalance, direction, ErrBalanceMismatch, []string{})(t)
	//				//	mvxBurnBalance = big.NewInt(100)
	//				//	testBridge(ethIsMintBurn, ethIsNative, mvxIsMintBurn, mvxIsNative, ethTotalBalances, zero, zero, zero, mvxMintBalance, mvxBurnBalance, direction, nil, []string{})(t)
	//				//})
	//			})
	//			mvxIsNative = true
	//			// mvxIsMintBurn does not matter
	//			testBridge(ethIsMintBurn, ethIsNative, mvxIsMintBurn, mvxIsNative, zero, zero, zero, zero, zero, zero, direction, ErrInvalidSetup, []string{"isNativeOnEthereum", "isNativeOnMultiversX"})(t)
	//		})
	//
	//		ethIsMintBurn = true
	//		t.Run("ethIsMintBurn = true", func(t *testing.T) {
	//			mvxIsNative = false
	//			t.Run("mvxIsNative = false", func(t *testing.T) {
	//				mvxIsMintBurn = false
	//				t.Run("mvxIsMintBurn = false", func(t *testing.T) {
	//					testBridge(ethIsMintBurn, ethIsNative, mvxIsMintBurn, mvxIsNative, zero, zero, zero, zero, zero, zero, direction, ErrInvalidSetup, []string{"isNativeOnMultiversX", "isMintBurnOnMultiversX"})(t)
	//				})
	//
	//				mvxIsMintBurn = true
	//				//TODO(jls): fix this
	//				//t.Run("mvxIsMintBurn = true", func(t *testing.T) {
	//				//	ethMintBalance := big.NewInt(0)
	//				//	ethBurnBalance := big.NewInt(100)
	//				//	mvxMintBalance := big.NewInt(100)
	//				//	mvxBurnBalance := big.NewInt(99)
	//				//
	//				//	testBridge(ethIsMintBurn, ethIsNative, mvxIsMintBurn, mvxIsNative, zero, ethMintBalance, ethBurnBalance, zero, mvxMintBalance, mvxBurnBalance, direction, ErrBalanceMismatch, []string{})(t)
	//				//	mvxBurnBalance = big.NewInt(100)
	//				//	testBridge(ethIsMintBurn, ethIsNative, mvxIsMintBurn, mvxIsNative, zero, ethMintBalance, ethBurnBalance, zero, mvxMintBalance, mvxBurnBalance, direction, nil, []string{})(t)
	//				//})
	//			})
	//			mvxIsNative = true
	//			// mvxIsMintBurn does not matter
	//			testBridge(ethIsMintBurn, ethIsNative, mvxIsMintBurn, mvxIsNative, zero, zero, zero, zero, zero, zero, direction, ErrInvalidSetup, []string{"isNativeOnEthereum", "isNativeOnMultiversX"})(t)
	//		})
	//	})
	//})
	//
	//direction = batchProcessor.ToMultiversX
	//t.Run("to MultiversX", func(t *testing.T) {
	//	ethIsNative = false
	//	t.Run("ethIsNative = false", func(t *testing.T) {
	//		ethIsMintBurn = false
	//		t.Run("ethIsMintBurn = false", func(t *testing.T) {
	//			// mvxIsNative && mvxIsMintBurn does not matter
	//			testBridge(ethIsMintBurn, ethIsNative, mvxIsMintBurn, mvxIsNative, zero, zero, zero, zero, zero, zero, direction, ErrInvalidSetup, []string{"isNativeOnEthereum", "isMintBurnOnEthereum"})(t)
	//		})
	//
	//		ethIsMintBurn = true
	//		t.Run("ethIsMintBurn = true", func(t *testing.T) {
	//			mvxIsNative = false
	//			t.Run("mvxIsNative = false", func(t *testing.T) {
	//				// mvxIsMintBurn does not matter
	//				testBridge(ethIsMintBurn, ethIsNative, mvxIsMintBurn, mvxIsNative, zero, zero, zero, zero, zero, zero, direction, ErrInvalidSetup, []string{"isNativeOnEthereum", "isNativeOnMultiversX"})(t)
	//			})
	//
	//			mvxIsNative = true
	//			t.Run("mvxIsNative = true", func(t *testing.T) {
	//				mvxIsMintBurn = false
	//				//TODO(jls): fix this
	//				//t.Run("mvxIsMintBurn = false", func(t *testing.T) {
	//				//	ethMintBalance := big.NewInt(0)
	//				//	ethBurnBalance := big.NewInt(0)
	//				//	mvxTotalBalances := big.NewInt(99)
	//				//
	//				//	testBridge(ethIsMintBurn, ethIsNative, mvxIsMintBurn, mvxIsNative, zero, ethMintBalance, ethBurnBalance, mvxTotalBalances, zero, zero, direction, ErrBalanceMismatch, []string{})(t)
	//				//	mvxTotalBalances = big.NewInt(100)
	//				//	testBridge(ethIsMintBurn, ethIsNative, mvxIsMintBurn, mvxIsNative, zero, ethMintBalance, ethBurnBalance, mvxTotalBalances, zero, zero, direction, nil, []string{})(t)
	//				//})
	//				mvxIsMintBurn = true
	//				//TODO(jls): fix this
	//				//t.Run("mvxIsMintBurn = true", func(t *testing.T) {
	//				//	ethMintBalance := big.NewInt(0)
	//				//	ethBurnBalance := big.NewInt(0)
	//				//	mvxBurnBalance := big.NewInt(99)
	//				//	mvxMintBalance := big.NewInt(0)
	//				//
	//				//	testBridge(ethIsMintBurn, ethIsNative, mvxIsMintBurn, mvxIsNative, zero, ethMintBalance, ethBurnBalance, zero, mvxMintBalance, mvxBurnBalance, direction, ErrBalanceMismatch, []string{})(t)
	//				//	mvxBurnBalance = big.NewInt(100)
	//				//	testBridge(ethIsMintBurn, ethIsNative, mvxIsMintBurn, mvxIsNative, zero, ethMintBalance, ethBurnBalance, zero, mvxMintBalance, mvxBurnBalance, direction, nil, []string{})(t)
	//				//})
	//			})
	//		})
	//	})
	//
	//	ethIsNative = true
	//	t.Run("ethIsNative = true", func(t *testing.T) {
	//		ethIsMintBurn = false
	//		t.Run("ethIsMintBurn = false", func(t *testing.T) {
	//			mvxIsNative = false
	//			t.Run("mvxIsNative = false", func(t *testing.T) {
	//				mvxIsMintBurn = false
	//				t.Run("mvxIsMintBurn = false", func(t *testing.T) {
	//					testBridge(ethIsMintBurn, ethIsNative, mvxIsMintBurn, mvxIsNative, zero, zero, zero, zero, zero, zero, direction, ErrInvalidSetup, []string{"isNativeOnMultiversX", "isMintBurnOnMultiversX"})(t)
	//				})
	//				mvxIsMintBurn = true
	//				//TODO(jls): fix this
	//				//t.Run("mvxIsMintBurn = true", func(t *testing.T) {
	//				//	ethTotalBalances := big.NewInt(99)
	//				//	mvxBurnBalance := big.NewInt(0)
	//				//	mvxMintBalance := big.NewInt(0)
	//				//
	//				//	testBridge(ethIsMintBurn, ethIsNative, mvxIsMintBurn, mvxIsNative, ethTotalBalances, zero, zero, zero, mvxMintBalance, mvxBurnBalance, direction, ErrBalanceMismatch, []string{})(t)
	//				//	ethTotalBalances = big.NewInt(100)
	//				//	testBridge(ethIsMintBurn, ethIsNative, mvxIsMintBurn, mvxIsNative, ethTotalBalances, zero, zero, zero, mvxMintBalance, mvxBurnBalance, direction, nil, []string{})(t)
	//				//})
	//			})
	//			mvxIsNative = true
	//			// mvxIsMintBurn does not matter
	//			testBridge(ethIsMintBurn, ethIsNative, mvxIsMintBurn, mvxIsNative, zero, zero, zero, zero, zero, zero, direction, ErrInvalidSetup, []string{"isNativeOnEthereum", "isNativeOnMultiversX"})(t)
	//		})
	//
	//		ethIsMintBurn = true
	//		t.Run("ethIsMintBurn = true", func(t *testing.T) {
	//			mvxIsNative = false
	//			t.Run("mvxIsNative = false", func(t *testing.T) {
	//				mvxIsMintBurn = false
	//				t.Run("mvxIsMintBurn = false", func(t *testing.T) {
	//					testBridge(ethIsMintBurn, ethIsNative, mvxIsMintBurn, mvxIsNative, zero, zero, zero, zero, zero, zero, direction, ErrInvalidSetup, []string{"isNativeOnMultiversX", "isMintBurnOnMultiversX"})(t)
	//				})
	//
	//				mvxIsMintBurn = true
	//				//TODO(jls): fix this
	//				//t.Run("mvxIsMintBurn = true", func(t *testing.T) {
	//				//	ethMintBalance := big.NewInt(0)
	//				//	ethBurnBalance := big.NewInt(100)
	//				//	mvxBurnBalance := big.NewInt(0)
	//				//	mvxMintBalance := big.NewInt(1)
	//				//
	//				//	testBridge(ethIsMintBurn, ethIsNative, mvxIsMintBurn, mvxIsNative, zero, ethMintBalance, ethBurnBalance, zero, mvxMintBalance, mvxBurnBalance, direction, ErrBalanceMismatch, []string{})(t)
	//				//	mvxMintBalance = big.NewInt(0)
	//				//	testBridge(ethIsMintBurn, ethIsNative, mvxIsMintBurn, mvxIsNative, zero, ethMintBalance, ethBurnBalance, zero, mvxMintBalance, mvxBurnBalance, direction, nil, []string{})(t)
	//				//})
	//			})
	//			mvxIsNative = true
	//			// mvxIsMintBurn does not matter
	//			testBridge(ethIsMintBurn, ethIsNative, mvxIsMintBurn, mvxIsNative, zero, zero, zero, zero, zero, zero, direction, ErrInvalidSetup, []string{"isNativeOnEthereum", "isNativeOnMultiversX"})(t)
	//		})
	//	})
	//})
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

func testBridge(
	ethIsMintBurn,
	ethIsNative,
	mvxIsMintBurn,
	mvxIsNative bool,
	ethTotalBalances,
	ethMintBalance,
	ethBurnBalance,
	mvxTotalBalances,
	mvxMintBalance,
	mvxBurnBalance *big.Int,
	direction batchProcessor.Direction,
	expectedErr error,
	expectedStringsInErr []string,
) func(t *testing.T) {
	return func(t *testing.T) {
		ethToken := common.BytesToAddress([]byte("eth token"))
		mvxToken := []byte("mvx token")
		amount := big.NewInt(100)

		args := createMockArgsBalanceValidator()
		args.EthereumClient = &bridge.EthereumClientStub{
			MintBurnTokensCalled: func(ctx context.Context, token common.Address) (bool, error) {
				return ethIsMintBurn, nil
			},
			NativeTokensCalled: func(ctx context.Context, token common.Address) (bool, error) {
				return ethIsNative, nil
			},
			TotalBalancesCalled: func(ctx context.Context, token common.Address) (*big.Int, error) {
				return ethTotalBalances, nil
			},
			MintBalancesCalled: func(ctx context.Context, token common.Address) (*big.Int, error) {
				return ethMintBalance, nil
			},
			BurnBalancesCalled: func(ctx context.Context, token common.Address) (*big.Int, error) {
				return ethBurnBalance, nil
			},
			CheckRequiredBalanceCalled: func(ctx context.Context, token common.Address, amount *big.Int) error {
				return nil
			},
		}

		args.MultiversXClient = &bridge.MultiversXClientStub{
			IsMintBurnTokenCalled: func(ctx context.Context, token []byte) (bool, error) {
				return mvxIsMintBurn, nil
			},
			IsNativeTokenCalled: func(ctx context.Context, token []byte) (bool, error) {
				return mvxIsNative, nil
			},
			TotalBalancesCalled: func(ctx context.Context, token []byte) (*big.Int, error) {
				return mvxTotalBalances, nil
			},
			MintBalancesCalled: func(ctx context.Context, token []byte) (*big.Int, error) {
				return mvxMintBalance, nil
			},
			BurnBalancesCalled: func(ctx context.Context, token []byte) (*big.Int, error) {
				return mvxBurnBalance, nil
			},
			CheckRequiredBalanceCalled: func(ctx context.Context, token []byte, amount *big.Int) error {
				return nil
			},
		}

		executor, _ := NewBalanceValidator(args)
		err := executor.CheckToken(context.Background(), ethToken, mvxToken, amount, direction)

		assert.True(t, errors.Is(err, expectedErr))
		for _, expectedStringInErr := range expectedStringsInErr {
			assert.True(t, strings.Contains(err.Error(), expectedStringInErr))
		}
	}
}
