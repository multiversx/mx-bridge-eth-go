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

var zero = big.NewInt(0)

func createMockArgsBalanceValidator() ArgsBalanceValidator {
	return ArgsBalanceValidator{
		Log:              &testscommon.LoggerStub{},
		MultiversXClient: &bridge.MultiversXClientStub{},
		EthereumClient:   &bridge.EthereumClientStub{},
	}
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

	direction := batchProcessor.FromMultiversX
	var ethIsNative, ethIsMintBurn, mvxIsNative, mvxIsMintBurn bool
	t.Run("from MultiversX", func(t *testing.T) {
		ethIsNative = false
		t.Run("ethIsNative = false", func(t *testing.T) {
			ethIsMintBurn = false
			t.Run("ethIsMintBurn = false", func(t *testing.T) {
				// mvxIsNative & mvxIsMintBurn does not matter
				testBridge(ethIsMintBurn, ethIsNative, mvxIsMintBurn, mvxIsNative, zero, zero, zero, zero, zero, zero, direction, ErrInvalidSetup, []string{"isNativeOnEthereum", "isMintBurnOnEthereum"})(t)
			})

			ethIsMintBurn = true
			t.Run("ethIsMintBurn = true", func(t *testing.T) {
				mvxIsNative = false
				t.Run("mvxIsNative = false", func(t *testing.T) {
					mvxIsMintBurn = false
					t.Run("mvxIsMintBurn = false", func(t *testing.T) {
						testBridge(ethIsMintBurn, ethIsNative, mvxIsMintBurn, mvxIsNative, zero, zero, zero, zero, zero, zero, direction, ErrInvalidSetup, []string{"isNativeOnMultiversX", "isMintBurnOnMultiversX"})(t)
					})

					mvxIsMintBurn = true
					t.Run("mvxIsMintBurn = true", func(t *testing.T) {
						testBridge(ethIsMintBurn, ethIsNative, mvxIsMintBurn, mvxIsNative, zero, zero, zero, zero, zero, zero, direction, ErrInvalidSetup, []string{"isNativeOnEthereum", "isNativeOnMultiversX"})(t)
					})
				})

				mvxIsNative = true
				//TODO(jls): fix this
				//t.Run("mvxIsNative = true", func(t *testing.T) {
				//	mvxIsMintBurn = false
				//	t.Run("mvxIsMintBurn = false", func(t *testing.T) {
				//		ethMintBalance := big.NewInt(0)
				//		ethBurnBalance := big.NewInt(0)
				//		mvxTotalBalances := big.NewInt(99)
				//
				//		testBridge(ethIsMintBurn, ethIsNative, mvxIsMintBurn, mvxIsNative, zero, ethMintBalance, ethBurnBalance, mvxTotalBalances, zero, zero, direction, ErrBalanceMismatch, []string{})(t)
				//		mvxTotalBalances = big.NewInt(100)
				//		testBridge(ethIsMintBurn, ethIsNative, mvxIsMintBurn, mvxIsNative, zero, ethMintBalance, ethBurnBalance, mvxTotalBalances, zero, zero, direction, nil, []string{})(t)
				//	})
				//	mvxIsMintBurn = true
				//	t.Run("mvxIsMintBurn = true", func(t *testing.T) {
				//		ethMintBalance := big.NewInt(0)
				//		ethBurnBalance := big.NewInt(0)
				//		mvxMintBalance := big.NewInt(0)
				//		mvxBurnBalance := big.NewInt(99)
				//
				//		testBridge(ethIsMintBurn, ethIsNative, mvxIsMintBurn, mvxIsNative, zero, ethMintBalance, ethBurnBalance, zero, mvxMintBalance, mvxBurnBalance, direction, ErrBalanceMismatch, []string{})(t)
				//		mvxBurnBalance = big.NewInt(100)
				//		testBridge(ethIsMintBurn, ethIsNative, mvxIsMintBurn, mvxIsNative, zero, ethMintBalance, ethBurnBalance, zero, mvxMintBalance, mvxBurnBalance, direction, nil, []string{})(t)
				//	})
				//})
			})
		})

		ethIsNative = true
		t.Run("ethIsNative = true", func(t *testing.T) {
			ethIsMintBurn = false
			t.Run("ethIsMintBurn = false", func(t *testing.T) {
				mvxIsNative = false
				t.Run("mvxIsNative = false", func(t *testing.T) {
					mvxIsMintBurn = false
					t.Run("mvxIsMintBurn = false", func(t *testing.T) {
						testBridge(ethIsMintBurn, ethIsNative, mvxIsMintBurn, mvxIsNative, zero, zero, zero, zero, zero, zero, direction, ErrInvalidSetup, []string{"isNativeOnMultiversX", "isMintBurnOnMultiversX"})(t)
					})
					mvxIsMintBurn = true
					//TODO(jls): fix this
					//t.Run("mvxIsMintBurn = true", func(t *testing.T) {
					//	ethTotalBalances := big.NewInt(100)
					//	mvxMintBalance := big.NewInt(100)
					//	mvxBurnBalance := big.NewInt(99)
					//
					//	testBridge(ethIsMintBurn, ethIsNative, mvxIsMintBurn, mvxIsNative, ethTotalBalances, zero, zero, zero, mvxMintBalance, mvxBurnBalance, direction, ErrBalanceMismatch, []string{})(t)
					//	mvxBurnBalance = big.NewInt(100)
					//	testBridge(ethIsMintBurn, ethIsNative, mvxIsMintBurn, mvxIsNative, ethTotalBalances, zero, zero, zero, mvxMintBalance, mvxBurnBalance, direction, nil, []string{})(t)
					//})
				})
				mvxIsNative = true
				// mvxIsMintBurn does not matter
				testBridge(ethIsMintBurn, ethIsNative, mvxIsMintBurn, mvxIsNative, zero, zero, zero, zero, zero, zero, direction, ErrInvalidSetup, []string{"isNativeOnEthereum", "isNativeOnMultiversX"})(t)
			})

			ethIsMintBurn = true
			t.Run("ethIsMintBurn = true", func(t *testing.T) {
				mvxIsNative = false
				t.Run("mvxIsNative = false", func(t *testing.T) {
					mvxIsMintBurn = false
					t.Run("mvxIsMintBurn = false", func(t *testing.T) {
						testBridge(ethIsMintBurn, ethIsNative, mvxIsMintBurn, mvxIsNative, zero, zero, zero, zero, zero, zero, direction, ErrInvalidSetup, []string{"isNativeOnMultiversX", "isMintBurnOnMultiversX"})(t)
					})

					mvxIsMintBurn = true
					//TODO(jls): fix this
					//t.Run("mvxIsMintBurn = true", func(t *testing.T) {
					//	ethMintBalance := big.NewInt(0)
					//	ethBurnBalance := big.NewInt(100)
					//	mvxMintBalance := big.NewInt(100)
					//	mvxBurnBalance := big.NewInt(99)
					//
					//	testBridge(ethIsMintBurn, ethIsNative, mvxIsMintBurn, mvxIsNative, zero, ethMintBalance, ethBurnBalance, zero, mvxMintBalance, mvxBurnBalance, direction, ErrBalanceMismatch, []string{})(t)
					//	mvxBurnBalance = big.NewInt(100)
					//	testBridge(ethIsMintBurn, ethIsNative, mvxIsMintBurn, mvxIsNative, zero, ethMintBalance, ethBurnBalance, zero, mvxMintBalance, mvxBurnBalance, direction, nil, []string{})(t)
					//})
				})
				mvxIsNative = true
				// mvxIsMintBurn does not matter
				testBridge(ethIsMintBurn, ethIsNative, mvxIsMintBurn, mvxIsNative, zero, zero, zero, zero, zero, zero, direction, ErrInvalidSetup, []string{"isNativeOnEthereum", "isNativeOnMultiversX"})(t)
			})
		})
	})

	direction = batchProcessor.ToMultiversX
	t.Run("to MultiversX", func(t *testing.T) {
		ethIsNative = false
		t.Run("ethIsNative = false", func(t *testing.T) {
			ethIsMintBurn = false
			t.Run("ethIsMintBurn = false", func(t *testing.T) {
				// mvxIsNative && mvxIsMintBurn does not matter
				testBridge(ethIsMintBurn, ethIsNative, mvxIsMintBurn, mvxIsNative, zero, zero, zero, zero, zero, zero, direction, ErrInvalidSetup, []string{"isNativeOnEthereum", "isMintBurnOnEthereum"})(t)
			})

			ethIsMintBurn = true
			t.Run("ethIsMintBurn = true", func(t *testing.T) {
				mvxIsNative = false
				t.Run("mvxIsNative = false", func(t *testing.T) {
					// mvxIsMintBurn does not matter
					testBridge(ethIsMintBurn, ethIsNative, mvxIsMintBurn, mvxIsNative, zero, zero, zero, zero, zero, zero, direction, ErrInvalidSetup, []string{"isNativeOnEthereum", "isNativeOnMultiversX"})(t)
				})

				mvxIsNative = true
				t.Run("mvxIsNative = true", func(t *testing.T) {
					mvxIsMintBurn = false
					//TODO(jls): fix this
					//t.Run("mvxIsMintBurn = false", func(t *testing.T) {
					//	ethMintBalance := big.NewInt(0)
					//	ethBurnBalance := big.NewInt(0)
					//	mvxTotalBalances := big.NewInt(99)
					//
					//	testBridge(ethIsMintBurn, ethIsNative, mvxIsMintBurn, mvxIsNative, zero, ethMintBalance, ethBurnBalance, mvxTotalBalances, zero, zero, direction, ErrBalanceMismatch, []string{})(t)
					//	mvxTotalBalances = big.NewInt(100)
					//	testBridge(ethIsMintBurn, ethIsNative, mvxIsMintBurn, mvxIsNative, zero, ethMintBalance, ethBurnBalance, mvxTotalBalances, zero, zero, direction, nil, []string{})(t)
					//})
					mvxIsMintBurn = true
					//TODO(jls): fix this
					//t.Run("mvxIsMintBurn = true", func(t *testing.T) {
					//	ethMintBalance := big.NewInt(0)
					//	ethBurnBalance := big.NewInt(0)
					//	mvxBurnBalance := big.NewInt(99)
					//	mvxMintBalance := big.NewInt(0)
					//
					//	testBridge(ethIsMintBurn, ethIsNative, mvxIsMintBurn, mvxIsNative, zero, ethMintBalance, ethBurnBalance, zero, mvxMintBalance, mvxBurnBalance, direction, ErrBalanceMismatch, []string{})(t)
					//	mvxBurnBalance = big.NewInt(100)
					//	testBridge(ethIsMintBurn, ethIsNative, mvxIsMintBurn, mvxIsNative, zero, ethMintBalance, ethBurnBalance, zero, mvxMintBalance, mvxBurnBalance, direction, nil, []string{})(t)
					//})
				})
			})
		})

		ethIsNative = true
		t.Run("ethIsNative = true", func(t *testing.T) {
			ethIsMintBurn = false
			t.Run("ethIsMintBurn = false", func(t *testing.T) {
				mvxIsNative = false
				t.Run("mvxIsNative = false", func(t *testing.T) {
					mvxIsMintBurn = false
					t.Run("mvxIsMintBurn = false", func(t *testing.T) {
						testBridge(ethIsMintBurn, ethIsNative, mvxIsMintBurn, mvxIsNative, zero, zero, zero, zero, zero, zero, direction, ErrInvalidSetup, []string{"isNativeOnMultiversX", "isMintBurnOnMultiversX"})(t)
					})
					mvxIsMintBurn = true
					//TODO(jls): fix this
					//t.Run("mvxIsMintBurn = true", func(t *testing.T) {
					//	ethTotalBalances := big.NewInt(99)
					//	mvxBurnBalance := big.NewInt(0)
					//	mvxMintBalance := big.NewInt(0)
					//
					//	testBridge(ethIsMintBurn, ethIsNative, mvxIsMintBurn, mvxIsNative, ethTotalBalances, zero, zero, zero, mvxMintBalance, mvxBurnBalance, direction, ErrBalanceMismatch, []string{})(t)
					//	ethTotalBalances = big.NewInt(100)
					//	testBridge(ethIsMintBurn, ethIsNative, mvxIsMintBurn, mvxIsNative, ethTotalBalances, zero, zero, zero, mvxMintBalance, mvxBurnBalance, direction, nil, []string{})(t)
					//})
				})
				mvxIsNative = true
				// mvxIsMintBurn does not matter
				testBridge(ethIsMintBurn, ethIsNative, mvxIsMintBurn, mvxIsNative, zero, zero, zero, zero, zero, zero, direction, ErrInvalidSetup, []string{"isNativeOnEthereum", "isNativeOnMultiversX"})(t)
			})

			ethIsMintBurn = true
			t.Run("ethIsMintBurn = true", func(t *testing.T) {
				mvxIsNative = false
				t.Run("mvxIsNative = false", func(t *testing.T) {
					mvxIsMintBurn = false
					t.Run("mvxIsMintBurn = false", func(t *testing.T) {
						testBridge(ethIsMintBurn, ethIsNative, mvxIsMintBurn, mvxIsNative, zero, zero, zero, zero, zero, zero, direction, ErrInvalidSetup, []string{"isNativeOnMultiversX", "isMintBurnOnMultiversX"})(t)
					})

					mvxIsMintBurn = true
					//TODO(jls): fix this
					//t.Run("mvxIsMintBurn = true", func(t *testing.T) {
					//	ethMintBalance := big.NewInt(0)
					//	ethBurnBalance := big.NewInt(100)
					//	mvxBurnBalance := big.NewInt(0)
					//	mvxMintBalance := big.NewInt(1)
					//
					//	testBridge(ethIsMintBurn, ethIsNative, mvxIsMintBurn, mvxIsNative, zero, ethMintBalance, ethBurnBalance, zero, mvxMintBalance, mvxBurnBalance, direction, ErrBalanceMismatch, []string{})(t)
					//	mvxMintBalance = big.NewInt(0)
					//	testBridge(ethIsMintBurn, ethIsNative, mvxIsMintBurn, mvxIsNative, zero, ethMintBalance, ethBurnBalance, zero, mvxMintBalance, mvxBurnBalance, direction, nil, []string{})(t)
					//})
				})
				mvxIsNative = true
				// mvxIsMintBurn does not matter
				testBridge(ethIsMintBurn, ethIsNative, mvxIsMintBurn, mvxIsNative, zero, zero, zero, zero, zero, zero, direction, ErrInvalidSetup, []string{"isNativeOnEthereum", "isNativeOnMultiversX"})(t)
			})
		})
	})
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
