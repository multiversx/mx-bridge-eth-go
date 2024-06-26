package balanceValidator

import (
	"context"
	"errors"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/multiversx/mx-bridge-eth-go/clients"
	"github.com/multiversx/mx-bridge-eth-go/core/batchProcessor"
	"github.com/multiversx/mx-bridge-eth-go/testsCommon/bridge"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/stretchr/testify/assert"
)

var (
	ethToken = common.BytesToAddress([]byte("eth token"))
	mvxToken = []byte("mvx token")
	amount   = big.NewInt(100)
	amount2  = big.NewInt(200)
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

	lastExecutedEthBatch       uint64
	pendingMvxBatchId          uint64
	amountsOnMvxPendingBatches map[uint64][]*big.Int
	amountsOnEthPendingBatches map[uint64][]*big.Int
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
		t.Run("on computeEthAmount, GetLastExecutedEthBatchID", func(t *testing.T) {
			cfg := testConfiguration{
				direction:       batchProcessor.FromMultiversX,
				isNativeOnMvx:   true,
				isMintBurnOnEth: true,
				errorsOnCalls: map[string]error{
					"GetLastExecutedEthBatchIDMvx": expectedError,
				},
			}
			result := validatorTester(cfg)
			assert.Equal(t, expectedError, result.error)
			assert.True(t, result.checkRequiredBalanceOnEthCalled)
			assert.False(t, result.checkRequiredBalanceOnMvxCalled)
		})
		t.Run("on computeEthAmount, GetBatch", func(t *testing.T) {
			cfg := testConfiguration{
				direction:       batchProcessor.FromMultiversX,
				isNativeOnMvx:   true,
				isMintBurnOnEth: true,
				errorsOnCalls: map[string]error{
					"GetBatchEth": expectedError,
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
		t.Run("on computeMvxAmount, GetPendingBatch", func(t *testing.T) {
			cfg := testConfiguration{
				direction:       batchProcessor.ToMultiversX,
				isMintBurnOnMvx: true,
				isNativeOnEth:   true,
				errorsOnCalls: map[string]error{
					"GetPendingBatchMvx": expectedError,
				},
			}
			result := validatorTester(cfg)
			assert.Equal(t, expectedError, result.error)
			assert.False(t, result.checkRequiredBalanceOnEthCalled)
			assert.True(t, result.checkRequiredBalanceOnMvxCalled)
		})
		t.Run("on computeMvxAmount, GetBatch", func(t *testing.T) {
			cfg := testConfiguration{
				direction:       batchProcessor.ToMultiversX,
				isMintBurnOnMvx: true,
				isNativeOnEth:   true,
				errorsOnCalls: map[string]error{
					"GetBatchMvx": expectedError,
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
	t.Run("scenarios", func(t *testing.T) {
		t.Parallel()

		t.Run("Ethereum -> MultiversX", func(t *testing.T) {
			t.Parallel()

			t.Run("native on MultiversX, mint-burn on Ethereum, ok values, no next pending batches", func(t *testing.T) {
				t.Parallel()

				cfg := testConfiguration{
					direction:          batchProcessor.ToMultiversX,
					isMintBurnOnEth:    true,
					isNativeOnMvx:      true,
					burnBalancesOnEth:  big.NewInt(1100),  // initial burn (1000) + burn from this transfer (100)
					mintBalancesOnEth:  big.NewInt(11000), // minted (10000) + initial burn (1000)
					totalBalancesOnMvx: big.NewInt(10000),
					amount:             amount,
					amountsOnEthPendingBatches: map[uint64][]*big.Int{
						1: {amount},
					},
					mvxToken: mvxToken,
					ethToken: ethToken,
				}

				result := validatorTester(cfg)
				assert.Nil(t, result.error)
				assert.False(t, result.checkRequiredBalanceOnEthCalled)
				assert.True(t, result.checkRequiredBalanceOnMvxCalled)

				t.Run("mismatch should error", func(t *testing.T) {
					cfg.burnBalancesOnEth.Add(cfg.burnBalancesOnEth, big.NewInt(1))
					result = validatorTester(cfg)
					assert.ErrorIs(t, result.error, ErrBalanceMismatch)
				})
			})
			t.Run("native on MultiversX, mint-burn on Ethereum, ok values, with next pending batches", func(t *testing.T) {
				t.Parallel()

				cfg := testConfiguration{
					direction:          batchProcessor.ToMultiversX,
					isMintBurnOnEth:    true,
					isNativeOnMvx:      true,
					burnBalancesOnEth:  big.NewInt(1220),  // initial burn (1000) + burn from this transfer (100) + burn from next batches (120)
					mintBalancesOnEth:  big.NewInt(11000), // minted (10000) + initial burn (1000)
					totalBalancesOnMvx: big.NewInt(10000),
					amount:             amount,
					amountsOnEthPendingBatches: map[uint64][]*big.Int{
						1: {amount},
						2: {big.NewInt(30), big.NewInt(40)},
						3: {big.NewInt(50)},
					},
					mvxToken: mvxToken,
					ethToken: ethToken,
				}

				result := validatorTester(cfg)
				assert.Nil(t, result.error)
				assert.False(t, result.checkRequiredBalanceOnEthCalled)
				assert.True(t, result.checkRequiredBalanceOnMvxCalled)

				t.Run("mismatch should error", func(t *testing.T) {
					cfg.burnBalancesOnEth.Add(cfg.burnBalancesOnEth, big.NewInt(1))
					result = validatorTester(cfg)
					assert.ErrorIs(t, result.error, ErrBalanceMismatch)
				})
			})
			t.Run("native on MultiversX but with mint-burn, mint-burn on Ethereum, ok values, no next pending batches", func(t *testing.T) {
				t.Parallel()

				cfg := testConfiguration{
					direction:         batchProcessor.ToMultiversX,
					isMintBurnOnEth:   true,
					isNativeOnMvx:     true,
					isMintBurnOnMvx:   true,
					burnBalancesOnEth: big.NewInt(1100),  // initial burn (1000) + burn from this transfer (100)
					mintBalancesOnEth: big.NewInt(11000), // minted (10000) + initial burn (1000)
					burnBalancesOnMvx: big.NewInt(12000),
					mintBalancesOnMvx: big.NewInt(2000), // burn - mint on Mvx === mint - burn on Eth
					amount:            amount,
					amountsOnEthPendingBatches: map[uint64][]*big.Int{
						1: {amount},
					},
					mvxToken: mvxToken,
					ethToken: ethToken,
				}

				result := validatorTester(cfg)
				assert.Nil(t, result.error)
				assert.False(t, result.checkRequiredBalanceOnEthCalled)
				assert.True(t, result.checkRequiredBalanceOnMvxCalled)

				t.Run("mismatch should error", func(t *testing.T) {
					cfg.burnBalancesOnEth.Add(cfg.burnBalancesOnEth, big.NewInt(1))
					result = validatorTester(cfg)
					assert.ErrorIs(t, result.error, ErrBalanceMismatch)
				})
			})
			t.Run("native on MultiversX but with mint-burn, mint-burn on Ethereum, ok values, with next pending batches", func(t *testing.T) {
				t.Parallel()

				cfg := testConfiguration{
					direction:         batchProcessor.ToMultiversX,
					isMintBurnOnEth:   true,
					isNativeOnMvx:     true,
					isMintBurnOnMvx:   true,
					burnBalancesOnEth: big.NewInt(1220),  // initial burn (1000) + burn from this transfer (100) + next batches (120)
					mintBalancesOnEth: big.NewInt(11000), // minted (10000) + initial burn (1000)
					burnBalancesOnMvx: big.NewInt(12000),
					mintBalancesOnMvx: big.NewInt(2000), // burn - mint on Mvx === mint - burn on Eth
					amount:            amount,
					amountsOnEthPendingBatches: map[uint64][]*big.Int{
						1: {amount},
						2: {big.NewInt(30), big.NewInt(40)},
						3: {big.NewInt(50)},
					},
					mvxToken: mvxToken,
					ethToken: ethToken,
				}

				result := validatorTester(cfg)
				assert.Nil(t, result.error)
				assert.False(t, result.checkRequiredBalanceOnEthCalled)
				assert.True(t, result.checkRequiredBalanceOnMvxCalled)

				t.Run("mismatch should error", func(t *testing.T) {
					cfg.burnBalancesOnEth.Add(cfg.burnBalancesOnEth, big.NewInt(1))
					result = validatorTester(cfg)
					assert.ErrorIs(t, result.error, ErrBalanceMismatch)
				})
			})
			t.Run("native on Ethereum, mint-burn on MultiversX, ok values, no next pending batches", func(t *testing.T) {
				t.Parallel()

				cfg := testConfiguration{
					direction:          batchProcessor.ToMultiversX,
					isMintBurnOnMvx:    true,
					isNativeOnEth:      true,
					burnBalancesOnMvx:  big.NewInt(1000),  // initial burn (1000)
					mintBalancesOnMvx:  big.NewInt(11000), // minted (10000) + initial burn (1000)
					totalBalancesOnEth: big.NewInt(10100), // initial (10000) + locked from this transfer (100)
					amount:             amount,
					amountsOnEthPendingBatches: map[uint64][]*big.Int{
						1: {amount},
					},
					mvxToken: mvxToken,
					ethToken: ethToken,
				}

				result := validatorTester(cfg)
				assert.Nil(t, result.error)
				assert.False(t, result.checkRequiredBalanceOnEthCalled)
				assert.True(t, result.checkRequiredBalanceOnMvxCalled)

				t.Run("mismatch should error", func(t *testing.T) {
					cfg.burnBalancesOnMvx.Add(cfg.burnBalancesOnMvx, big.NewInt(1))
					result = validatorTester(cfg)
					assert.ErrorIs(t, result.error, ErrBalanceMismatch)
				})
			})
			t.Run("native on Ethereum, mint-burn on MultiversX, ok values, with next pending batches", func(t *testing.T) {
				t.Parallel()

				cfg := testConfiguration{
					direction:          batchProcessor.ToMultiversX,
					isMintBurnOnMvx:    true,
					isNativeOnEth:      true,
					burnBalancesOnMvx:  big.NewInt(1000),  // initial burn (1000)
					mintBalancesOnMvx:  big.NewInt(11000), // minted (10000) + initial burn (1000)
					totalBalancesOnEth: big.NewInt(10220), // initial (10000) + locked from this transfer (100) + next batches (120)
					amount:             amount,
					amountsOnEthPendingBatches: map[uint64][]*big.Int{
						1: {amount},
						2: {big.NewInt(30), big.NewInt(40)},
						3: {big.NewInt(50)},
					},
					mvxToken: mvxToken,
					ethToken: ethToken,
				}

				result := validatorTester(cfg)
				assert.Nil(t, result.error)
				assert.False(t, result.checkRequiredBalanceOnEthCalled)
				assert.True(t, result.checkRequiredBalanceOnMvxCalled)

				t.Run("mismatch should error", func(t *testing.T) {
					cfg.burnBalancesOnMvx.Add(cfg.burnBalancesOnMvx, big.NewInt(1))
					result = validatorTester(cfg)
					assert.ErrorIs(t, result.error, ErrBalanceMismatch)
				})
			})
			t.Run("native on Ethereum but with mint-burn, mint-burn on MultiversX, ok values, no next pending batches", func(t *testing.T) {
				t.Parallel()

				cfg := testConfiguration{
					direction:         batchProcessor.ToMultiversX,
					isMintBurnOnMvx:   true,
					isNativeOnEth:     true,
					isMintBurnOnEth:   true,
					burnBalancesOnMvx: big.NewInt(1000),  // initial burn (1000)
					mintBalancesOnMvx: big.NewInt(11000), // minted (10000) + initial burn (1000)
					burnBalancesOnEth: big.NewInt(12100),
					mintBalancesOnEth: big.NewInt(2000), // burn - mint - transfer on Eth === mint - burn on Mvx
					amount:            amount,
					amountsOnEthPendingBatches: map[uint64][]*big.Int{
						1: {amount},
					},
					mvxToken: mvxToken,
					ethToken: ethToken,
				}

				result := validatorTester(cfg)
				assert.Nil(t, result.error)
				assert.False(t, result.checkRequiredBalanceOnEthCalled)
				assert.True(t, result.checkRequiredBalanceOnMvxCalled)

				t.Run("mismatch should error", func(t *testing.T) {
					cfg.burnBalancesOnMvx.Add(cfg.burnBalancesOnMvx, big.NewInt(1))
					result = validatorTester(cfg)
					assert.ErrorIs(t, result.error, ErrBalanceMismatch)
				})
			})
			t.Run("native on Ethereum but with mint-burn, mint-burn on MultiversX, ok values, with next pending batches", func(t *testing.T) {
				t.Parallel()

				cfg := testConfiguration{
					direction:         batchProcessor.ToMultiversX,
					isMintBurnOnMvx:   true,
					isNativeOnEth:     true,
					isMintBurnOnEth:   true,
					burnBalancesOnMvx: big.NewInt(1000),  // initial burn (1000)
					mintBalancesOnMvx: big.NewInt(11000), // minted (10000) + initial burn (1000)
					burnBalancesOnEth: big.NewInt(12220),
					mintBalancesOnEth: big.NewInt(2000), // burn - mint - transfer on Eth - next transfers === mint - burn on Mvx
					amount:            amount,
					amountsOnEthPendingBatches: map[uint64][]*big.Int{
						1: {amount},
						2: {big.NewInt(30), big.NewInt(40)},
						3: {big.NewInt(50)},
					},
					mvxToken: mvxToken,
					ethToken: ethToken,
				}

				result := validatorTester(cfg)
				assert.Nil(t, result.error)
				assert.False(t, result.checkRequiredBalanceOnEthCalled)
				assert.True(t, result.checkRequiredBalanceOnMvxCalled)

				t.Run("mismatch should error", func(t *testing.T) {
					cfg.burnBalancesOnMvx.Add(cfg.burnBalancesOnMvx, big.NewInt(1))
					result = validatorTester(cfg)
					assert.ErrorIs(t, result.error, ErrBalanceMismatch)
				})
			})
		})

		t.Run("MultiversX -> Ethereum", func(t *testing.T) {
			t.Parallel()

			t.Run("native on MultiversX, mint-burn on Ethereum, ok values, no next pending batches", func(t *testing.T) {
				t.Parallel()

				cfg := testConfiguration{
					direction:          batchProcessor.FromMultiversX,
					isMintBurnOnEth:    true,
					isNativeOnMvx:      true,
					burnBalancesOnEth:  big.NewInt(1000),  // initial burn (1000)
					mintBalancesOnEth:  big.NewInt(11000), // minted (10000) + initial burn (1000)
					totalBalancesOnMvx: big.NewInt(10100), // initial (10000) + transfer from this batch (100)
					amount:             amount,
					amountsOnMvxPendingBatches: map[uint64][]*big.Int{
						1: {amount},
					},
					mvxToken: mvxToken,
					ethToken: ethToken,
				}

				result := validatorTester(cfg)
				assert.Nil(t, result.error)
				assert.True(t, result.checkRequiredBalanceOnEthCalled)
				assert.False(t, result.checkRequiredBalanceOnMvxCalled)

				t.Run("mismatch should error", func(t *testing.T) {
					cfg.burnBalancesOnEth.Add(cfg.burnBalancesOnEth, big.NewInt(1))
					result = validatorTester(cfg)
					assert.ErrorIs(t, result.error, ErrBalanceMismatch)
				})
			})
			t.Run("native on MultiversX, mint-burn on Ethereum, ok values, with next pending batches", func(t *testing.T) {
				t.Parallel()

				cfg := testConfiguration{
					direction:          batchProcessor.FromMultiversX,
					isMintBurnOnEth:    true,
					isNativeOnMvx:      true,
					burnBalancesOnEth:  big.NewInt(1000),  // initial burn (1000)
					mintBalancesOnEth:  big.NewInt(11000), // minted (10000) + initial burn (1000)
					totalBalancesOnMvx: big.NewInt(10220), // initial (10000) + transfer from this batch (100) + next batches (120)
					amount:             amount,
					amountsOnMvxPendingBatches: map[uint64][]*big.Int{
						1: {amount},
						2: {big.NewInt(30), big.NewInt(40)},
						3: {big.NewInt(50)},
					},
					mvxToken: mvxToken,
					ethToken: ethToken,
				}

				result := validatorTester(cfg)
				assert.Nil(t, result.error)
				assert.True(t, result.checkRequiredBalanceOnEthCalled)
				assert.False(t, result.checkRequiredBalanceOnMvxCalled)

				t.Run("mismatch should error", func(t *testing.T) {
					cfg.burnBalancesOnEth.Add(cfg.burnBalancesOnEth, big.NewInt(1))
					result = validatorTester(cfg)
					assert.ErrorIs(t, result.error, ErrBalanceMismatch)
				})
			})
			t.Run("native on MultiversX but with mint-burn, mint-burn on Ethereum, ok values, no next pending batches", func(t *testing.T) {
				t.Parallel()

				cfg := testConfiguration{
					direction:         batchProcessor.FromMultiversX,
					isMintBurnOnEth:   true,
					isNativeOnMvx:     true,
					isMintBurnOnMvx:   true,
					burnBalancesOnEth: big.NewInt(1000),  // initial burn (1000)
					mintBalancesOnEth: big.NewInt(11000), // minted (10000) + initial burn (1000)
					burnBalancesOnMvx: big.NewInt(12100),
					mintBalancesOnMvx: big.NewInt(2000), // burn - mint - transfer on Mvx === mint - burn on Eth
					amount:            amount,
					amountsOnMvxPendingBatches: map[uint64][]*big.Int{
						1: {amount},
					},
					mvxToken: mvxToken,
					ethToken: ethToken,
				}

				result := validatorTester(cfg)
				assert.Nil(t, result.error)
				assert.True(t, result.checkRequiredBalanceOnEthCalled)
				assert.False(t, result.checkRequiredBalanceOnMvxCalled)

				t.Run("mismatch should error", func(t *testing.T) {
					cfg.burnBalancesOnEth.Add(cfg.burnBalancesOnEth, big.NewInt(1))
					result = validatorTester(cfg)
					assert.ErrorIs(t, result.error, ErrBalanceMismatch)
				})
			})
			t.Run("native on MultiversX but with mint-burn, mint-burn on Ethereum, ok values, with next pending batches", func(t *testing.T) {
				t.Parallel()

				cfg := testConfiguration{
					direction:         batchProcessor.FromMultiversX,
					isMintBurnOnEth:   true,
					isNativeOnMvx:     true,
					isMintBurnOnMvx:   true,
					burnBalancesOnEth: big.NewInt(1000),  // initial burn (1000)
					mintBalancesOnEth: big.NewInt(11000), // minted (10000) + initial burn (1000)
					burnBalancesOnMvx: big.NewInt(12220),
					mintBalancesOnMvx: big.NewInt(2000), // burn - mint - transfer - next batches on Mvx === mint - burn on Eth
					amount:            amount,
					amountsOnMvxPendingBatches: map[uint64][]*big.Int{
						1: {amount},
						2: {big.NewInt(30), big.NewInt(40)},
						3: {big.NewInt(50)},
					},
					mvxToken: mvxToken,
					ethToken: ethToken,
				}

				result := validatorTester(cfg)
				assert.Nil(t, result.error)
				assert.True(t, result.checkRequiredBalanceOnEthCalled)
				assert.False(t, result.checkRequiredBalanceOnMvxCalled)

				t.Run("mismatch should error", func(t *testing.T) {
					cfg.burnBalancesOnEth.Add(cfg.burnBalancesOnEth, big.NewInt(1))
					result = validatorTester(cfg)
					assert.ErrorIs(t, result.error, ErrBalanceMismatch)
				})
			})
			t.Run("native on Ethereum, mint-burn on MultiversX, ok values, no next pending batches", func(t *testing.T) {
				t.Parallel()

				cfg := testConfiguration{
					direction:          batchProcessor.FromMultiversX,
					isMintBurnOnMvx:    true,
					isNativeOnEth:      true,
					burnBalancesOnMvx:  big.NewInt(1100),  // initial burn (1000) + transfer from this batch (100)
					mintBalancesOnMvx:  big.NewInt(11000), // minted (10000) + initial burn (1000)
					totalBalancesOnEth: big.NewInt(10000), // initial (10000)
					amount:             amount,
					amountsOnMvxPendingBatches: map[uint64][]*big.Int{
						1: {amount},
					},
					mvxToken: mvxToken,
					ethToken: ethToken,
				}

				result := validatorTester(cfg)
				assert.Nil(t, result.error)
				assert.True(t, result.checkRequiredBalanceOnEthCalled)
				assert.False(t, result.checkRequiredBalanceOnMvxCalled)

				t.Run("mismatch should error", func(t *testing.T) {
					cfg.burnBalancesOnMvx.Add(cfg.burnBalancesOnMvx, big.NewInt(1))
					result = validatorTester(cfg)
					assert.ErrorIs(t, result.error, ErrBalanceMismatch)
				})
			})
			t.Run("native on Ethereum, mint-burn on MultiversX, ok values, with next pending batches", func(t *testing.T) {
				t.Parallel()

				cfg := testConfiguration{
					direction:          batchProcessor.FromMultiversX,
					isMintBurnOnMvx:    true,
					isNativeOnEth:      true,
					burnBalancesOnMvx:  big.NewInt(1220),  // initial burn (1000) + transfer from this batch (100) + next batches (120)
					mintBalancesOnMvx:  big.NewInt(11000), // minted (10000) + initial burn (1000)
					totalBalancesOnEth: big.NewInt(10000), // initial (10000)
					amount:             amount,
					amountsOnMvxPendingBatches: map[uint64][]*big.Int{
						1: {amount},
						2: {big.NewInt(30), big.NewInt(40)},
						3: {big.NewInt(50)},
					},
					mvxToken: mvxToken,
					ethToken: ethToken,
				}

				result := validatorTester(cfg)
				assert.Nil(t, result.error)
				assert.True(t, result.checkRequiredBalanceOnEthCalled)
				assert.False(t, result.checkRequiredBalanceOnMvxCalled)

				t.Run("mismatch should error", func(t *testing.T) {
					cfg.burnBalancesOnMvx.Add(cfg.burnBalancesOnMvx, big.NewInt(1))
					result = validatorTester(cfg)
					assert.ErrorIs(t, result.error, ErrBalanceMismatch)
				})
			})
			t.Run("native on Ethereum but with mint-burn, mint-burn on MultiversX, ok values, no next pending batches", func(t *testing.T) {
				t.Parallel()

				cfg := testConfiguration{
					direction:         batchProcessor.FromMultiversX,
					isMintBurnOnMvx:   true,
					isNativeOnEth:     true,
					isMintBurnOnEth:   true,
					burnBalancesOnMvx: big.NewInt(1100),  // initial burn (1000) + transfer from this batch (100)
					mintBalancesOnMvx: big.NewInt(11000), // minted (10000) + initial burn (1000)
					burnBalancesOnEth: big.NewInt(12000),
					mintBalancesOnEth: big.NewInt(2000), // burn - mint on Eth === mint - burn - transfer on Mvx
					amount:            amount,
					amountsOnMvxPendingBatches: map[uint64][]*big.Int{
						1: {amount},
					},
					mvxToken: mvxToken,
					ethToken: ethToken,
				}

				result := validatorTester(cfg)
				assert.Nil(t, result.error)
				assert.True(t, result.checkRequiredBalanceOnEthCalled)
				assert.False(t, result.checkRequiredBalanceOnMvxCalled)

				t.Run("mismatch should error", func(t *testing.T) {
					cfg.burnBalancesOnMvx.Add(cfg.burnBalancesOnMvx, big.NewInt(1))
					result = validatorTester(cfg)
					assert.ErrorIs(t, result.error, ErrBalanceMismatch)
				})
			})
			t.Run("native on Ethereum but with mint-burn, mint-burn on MultiversX, ok values, with next pending batches", func(t *testing.T) {
				t.Parallel()

				cfg := testConfiguration{
					direction:         batchProcessor.FromMultiversX,
					isMintBurnOnMvx:   true,
					isNativeOnEth:     true,
					isMintBurnOnEth:   true,
					burnBalancesOnMvx: big.NewInt(1220),  // initial burn (1000) + transfer from this batch (100) + transfer from next batches
					mintBalancesOnMvx: big.NewInt(11000), // minted (10000) + initial burn (1000)
					burnBalancesOnEth: big.NewInt(12000),
					mintBalancesOnEth: big.NewInt(2000), // burn - mint on Eth === mint - burn - transfer - next batches on Mvx
					amount:            amount,
					amountsOnMvxPendingBatches: map[uint64][]*big.Int{
						1: {amount},
						2: {big.NewInt(30), big.NewInt(40)},
						3: {big.NewInt(50)},
					},
					mvxToken: mvxToken,
					ethToken: ethToken,
				}

				result := validatorTester(cfg)
				assert.Nil(t, result.error)
				assert.True(t, result.checkRequiredBalanceOnEthCalled)
				assert.False(t, result.checkRequiredBalanceOnMvxCalled)

				t.Run("mismatch should error", func(t *testing.T) {
					cfg.burnBalancesOnMvx.Add(cfg.burnBalancesOnMvx, big.NewInt(1))
					result = validatorTester(cfg)
					assert.ErrorIs(t, result.error, ErrBalanceMismatch)
				})
			})
		})

		t.Run("MultiversX <-> Ethereum", func(t *testing.T) {
			t.Parallel()

			t.Run("from Ethereum: native on MultiversX, mint-burn on Ethereum, ok values, with next pending batches", func(t *testing.T) {
				t.Parallel()

				cfg := testConfiguration{
					direction:          batchProcessor.ToMultiversX,
					isMintBurnOnEth:    true,
					isNativeOnMvx:      true,
					burnBalancesOnEth:  big.NewInt(1220),  // initial burn (1000) + burn from this transfer (100) + burn from next batches (120)
					mintBalancesOnEth:  big.NewInt(11000), // minted (10000) + initial burn (1000)
					totalBalancesOnMvx: big.NewInt(10440),
					amount:             amount,
					amountsOnEthPendingBatches: map[uint64][]*big.Int{
						1: {amount},
						2: {big.NewInt(30), big.NewInt(40)},
						3: {big.NewInt(50)},
					},
					amountsOnMvxPendingBatches: map[uint64][]*big.Int{
						1: {amount2},
						2: {big.NewInt(60), big.NewInt(80)},
						3: {big.NewInt(100)},
					},
					mvxToken: mvxToken,
					ethToken: ethToken,
				}

				result := validatorTester(cfg)
				assert.Nil(t, result.error)
				assert.False(t, result.checkRequiredBalanceOnEthCalled)
				assert.True(t, result.checkRequiredBalanceOnMvxCalled)

				t.Run("mismatch should error", func(t *testing.T) {
					cfg.burnBalancesOnEth.Add(cfg.burnBalancesOnEth, big.NewInt(1))
					result = validatorTester(cfg)
					assert.ErrorIs(t, result.error, ErrBalanceMismatch)
				})
			})
			t.Run("from Ethereum: native on MultiversX but with mint-burn, mint-burn on Ethereum, ok values, with next pending batches", func(t *testing.T) {
				t.Parallel()

				cfg := testConfiguration{
					direction:         batchProcessor.ToMultiversX,
					isMintBurnOnEth:   true,
					isNativeOnMvx:     true,
					isMintBurnOnMvx:   true,
					burnBalancesOnEth: big.NewInt(1220),  // initial burn (1000) + burn from this transfer (100) + next batches (120)
					mintBalancesOnEth: big.NewInt(11000), // minted (10000) + initial burn (1000)
					burnBalancesOnMvx: big.NewInt(12440),
					mintBalancesOnMvx: big.NewInt(2000), // burn - mint on Mvx === mint - burn on Eth
					amount:            amount,
					amountsOnEthPendingBatches: map[uint64][]*big.Int{
						1: {amount},
						2: {big.NewInt(30), big.NewInt(40)},
						3: {big.NewInt(50)},
					},
					amountsOnMvxPendingBatches: map[uint64][]*big.Int{
						1: {amount2},
						2: {big.NewInt(60), big.NewInt(80)},
						3: {big.NewInt(100)},
					},
					mvxToken: mvxToken,
					ethToken: ethToken,
				}

				result := validatorTester(cfg)
				assert.Nil(t, result.error)
				assert.False(t, result.checkRequiredBalanceOnEthCalled)
				assert.True(t, result.checkRequiredBalanceOnMvxCalled)

				t.Run("mismatch should error", func(t *testing.T) {
					cfg.burnBalancesOnEth.Add(cfg.burnBalancesOnEth, big.NewInt(1))
					result = validatorTester(cfg)
					assert.ErrorIs(t, result.error, ErrBalanceMismatch)
				})
			})
			t.Run("from Ethereum: native on Ethereum, mint-burn on MultiversX, ok values, with next pending batches", func(t *testing.T) {
				t.Parallel()

				cfg := testConfiguration{
					direction:          batchProcessor.ToMultiversX,
					isMintBurnOnMvx:    true,
					isNativeOnEth:      true,
					burnBalancesOnMvx:  big.NewInt(1440),  // initial burn (1000)
					mintBalancesOnMvx:  big.NewInt(11000), // minted (10000) + initial burn (1000)
					totalBalancesOnEth: big.NewInt(10220), // initial (10000) + locked from this transfer (100) + next batches (120)
					amount:             amount,
					amountsOnEthPendingBatches: map[uint64][]*big.Int{
						1: {amount},
						2: {big.NewInt(30), big.NewInt(40)},
						3: {big.NewInt(50)},
					},
					amountsOnMvxPendingBatches: map[uint64][]*big.Int{
						1: {amount2},
						2: {big.NewInt(60), big.NewInt(80)},
						3: {big.NewInt(100)},
					},
					mvxToken: mvxToken,
					ethToken: ethToken,
				}

				result := validatorTester(cfg)
				assert.Nil(t, result.error)
				assert.False(t, result.checkRequiredBalanceOnEthCalled)
				assert.True(t, result.checkRequiredBalanceOnMvxCalled)

				t.Run("mismatch should error", func(t *testing.T) {
					cfg.burnBalancesOnMvx.Add(cfg.burnBalancesOnMvx, big.NewInt(1))
					result = validatorTester(cfg)
					assert.ErrorIs(t, result.error, ErrBalanceMismatch)
				})
			})
			t.Run("from Ethereum: native on Ethereum but with mint-burn, mint-burn on MultiversX, ok values, with next pending batches", func(t *testing.T) {
				t.Parallel()

				cfg := testConfiguration{
					direction:         batchProcessor.ToMultiversX,
					isMintBurnOnMvx:   true,
					isNativeOnEth:     true,
					isMintBurnOnEth:   true,
					burnBalancesOnMvx: big.NewInt(1440),  // initial burn (1000)
					mintBalancesOnMvx: big.NewInt(11000), // minted (10000) + initial burn (1000)
					burnBalancesOnEth: big.NewInt(12220),
					mintBalancesOnEth: big.NewInt(2000), // burn - mint - transfer on Eth - next transfers === mint - burn on Mvx
					amount:            amount,
					amountsOnEthPendingBatches: map[uint64][]*big.Int{
						1: {amount},
						2: {big.NewInt(30), big.NewInt(40)},
						3: {big.NewInt(50)},
					},
					amountsOnMvxPendingBatches: map[uint64][]*big.Int{
						1: {amount2},
						2: {big.NewInt(60), big.NewInt(80)},
						3: {big.NewInt(100)},
					},
					mvxToken: mvxToken,
					ethToken: ethToken,
				}

				result := validatorTester(cfg)
				assert.Nil(t, result.error)
				assert.False(t, result.checkRequiredBalanceOnEthCalled)
				assert.True(t, result.checkRequiredBalanceOnMvxCalled)

				t.Run("mismatch should error", func(t *testing.T) {
					cfg.burnBalancesOnMvx.Add(cfg.burnBalancesOnMvx, big.NewInt(1))
					result = validatorTester(cfg)
					assert.ErrorIs(t, result.error, ErrBalanceMismatch)
				})
			})
			t.Run("from MultiversX: native on MultiversX, mint-burn on Ethereum, ok values, with next pending batches on both chains", func(t *testing.T) {
				t.Parallel()

				cfg := testConfiguration{
					direction:          batchProcessor.FromMultiversX,
					isMintBurnOnEth:    true,
					isNativeOnMvx:      true,
					burnBalancesOnEth:  big.NewInt(1440),  // initial burn (1000) + amount2 + transfers from eth
					mintBalancesOnEth:  big.NewInt(11000), // minted (10000) + initial burn (1000)
					totalBalancesOnMvx: big.NewInt(10220), // initial (10000) + transfer from this batch (100) + next batches (120)
					amount:             amount,
					amountsOnMvxPendingBatches: map[uint64][]*big.Int{
						1: {amount},
						2: {big.NewInt(30), big.NewInt(40)},
						3: {big.NewInt(50)},
					},
					amountsOnEthPendingBatches: map[uint64][]*big.Int{
						1: {amount2},
						2: {big.NewInt(60), big.NewInt(80)},
						3: {big.NewInt(100)},
					},
					mvxToken: mvxToken,
					ethToken: ethToken,
				}

				result := validatorTester(cfg)
				assert.Nil(t, result.error)
				assert.True(t, result.checkRequiredBalanceOnEthCalled)
				assert.False(t, result.checkRequiredBalanceOnMvxCalled)

				t.Run("mismatch should error", func(t *testing.T) {
					cfg.burnBalancesOnEth.Add(cfg.burnBalancesOnEth, big.NewInt(1))
					result = validatorTester(cfg)
					assert.ErrorIs(t, result.error, ErrBalanceMismatch)
				})
			})
			t.Run("from MultiversX: native on MultiversX but with mint-burn, mint-burn on Ethereum, ok values, with next pending batches", func(t *testing.T) {
				t.Parallel()

				cfg := testConfiguration{
					direction:         batchProcessor.FromMultiversX,
					isMintBurnOnEth:   true,
					isNativeOnMvx:     true,
					isMintBurnOnMvx:   true,
					burnBalancesOnEth: big.NewInt(1440),  // initial burn (1000)
					mintBalancesOnEth: big.NewInt(11000), // minted (10000) + initial burn (1000)
					burnBalancesOnMvx: big.NewInt(12220),
					mintBalancesOnMvx: big.NewInt(2000), // burn - mint - transfer - next batches on Mvx === mint - burn on Eth
					amount:            amount,
					amountsOnMvxPendingBatches: map[uint64][]*big.Int{
						1: {amount},
						2: {big.NewInt(30), big.NewInt(40)},
						3: {big.NewInt(50)},
					},
					amountsOnEthPendingBatches: map[uint64][]*big.Int{
						1: {amount2},
						2: {big.NewInt(60), big.NewInt(80)},
						3: {big.NewInt(100)},
					},
					mvxToken: mvxToken,
					ethToken: ethToken,
				}

				result := validatorTester(cfg)
				assert.Nil(t, result.error)
				assert.True(t, result.checkRequiredBalanceOnEthCalled)
				assert.False(t, result.checkRequiredBalanceOnMvxCalled)

				t.Run("mismatch should error", func(t *testing.T) {
					cfg.burnBalancesOnEth.Add(cfg.burnBalancesOnEth, big.NewInt(1))
					result = validatorTester(cfg)
					assert.ErrorIs(t, result.error, ErrBalanceMismatch)
				})
			})
			t.Run("from MultiversX: native on Ethereum, mint-burn on MultiversX, ok values, with next pending batches", func(t *testing.T) {
				t.Parallel()

				cfg := testConfiguration{
					direction:          batchProcessor.FromMultiversX,
					isMintBurnOnMvx:    true,
					isNativeOnEth:      true,
					burnBalancesOnMvx:  big.NewInt(1220),  // initial burn (1000) + transfer from this batch (100) + next batches (120)
					mintBalancesOnMvx:  big.NewInt(11000), // minted (10000) + initial burn (1000)
					totalBalancesOnEth: big.NewInt(10440), // initial (10000) + pending batches on eth
					amount:             amount,
					amountsOnMvxPendingBatches: map[uint64][]*big.Int{
						1: {amount},
						2: {big.NewInt(30), big.NewInt(40)},
						3: {big.NewInt(50)},
					},
					amountsOnEthPendingBatches: map[uint64][]*big.Int{
						1: {amount2},
						2: {big.NewInt(60), big.NewInt(80)},
						3: {big.NewInt(100)},
					},
					mvxToken: mvxToken,
					ethToken: ethToken,
				}

				result := validatorTester(cfg)
				assert.Nil(t, result.error)
				assert.True(t, result.checkRequiredBalanceOnEthCalled)
				assert.False(t, result.checkRequiredBalanceOnMvxCalled)

				t.Run("mismatch should error", func(t *testing.T) {
					cfg.burnBalancesOnMvx.Add(cfg.burnBalancesOnMvx, big.NewInt(1))
					result = validatorTester(cfg)
					assert.ErrorIs(t, result.error, ErrBalanceMismatch)
				})
			})
			t.Run("from MultiversX: native on Ethereum but with mint-burn, mint-burn on MultiversX, ok values, with next pending batches", func(t *testing.T) {
				t.Parallel()

				cfg := testConfiguration{
					direction:         batchProcessor.FromMultiversX,
					isMintBurnOnMvx:   true,
					isNativeOnEth:     true,
					isMintBurnOnEth:   true,
					burnBalancesOnMvx: big.NewInt(1220),  // initial burn (1000) + transfer from this batch (100) + transfer from next batches
					mintBalancesOnMvx: big.NewInt(11000), // minted (10000) + initial burn (1000)
					burnBalancesOnEth: big.NewInt(12440),
					mintBalancesOnEth: big.NewInt(2000), // burn - mint on Eth === mint - burn - transfer - next batches on Mvx
					amount:            amount,
					amountsOnMvxPendingBatches: map[uint64][]*big.Int{
						1: {amount},
						2: {big.NewInt(30), big.NewInt(40)},
						3: {big.NewInt(50)},
					},
					amountsOnEthPendingBatches: map[uint64][]*big.Int{
						1: {amount2},
						2: {big.NewInt(60), big.NewInt(80)},
						3: {big.NewInt(100)},
					},
					mvxToken: mvxToken,
					ethToken: ethToken,
				}

				result := validatorTester(cfg)
				assert.Nil(t, result.error)
				assert.True(t, result.checkRequiredBalanceOnEthCalled)
				assert.False(t, result.checkRequiredBalanceOnMvxCalled)

				t.Run("mismatch should error", func(t *testing.T) {
					cfg.burnBalancesOnMvx.Add(cfg.burnBalancesOnMvx, big.NewInt(1))
					result = validatorTester(cfg)
					assert.ErrorIs(t, result.error, ErrBalanceMismatch)
				})
			})
		})

	})
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

			return returnBigIntOrZeroIfNil(cfg.totalBalancesOnMvx), nil
		},
		MintBalancesCalled: func(ctx context.Context, token []byte) (*big.Int, error) {
			err := cfg.errorsOnCalls["MintBalancesMvx"]
			if err != nil {
				return nil, err
			}

			return returnBigIntOrZeroIfNil(cfg.mintBalancesOnMvx), nil
		},
		BurnBalancesCalled: func(ctx context.Context, token []byte) (*big.Int, error) {
			err := cfg.errorsOnCalls["BurnBalancesMvx"]
			if err != nil {
				return nil, err
			}

			return returnBigIntOrZeroIfNil(cfg.burnBalancesOnMvx), nil
		},
		GetPendingBatchCalled: func(ctx context.Context) (*clients.TransferBatch, error) {
			err := cfg.errorsOnCalls["GetPendingBatchMvx"]
			if err != nil {
				return nil, err
			}

			batch := &clients.TransferBatch{
				ID: cfg.pendingMvxBatchId,
			}
			applyDummyFromMvxDepositsToBatch(cfg, batch)

			return batch, nil
		},
		GetBatchCalled: func(ctx context.Context, batchID uint64) (*clients.TransferBatch, error) {
			err := cfg.errorsOnCalls["GetBatchMvx"]
			if err != nil {
				return nil, err
			}

			if batchID > getMaxMvxPendingBatchID(cfg) {
				return nil, clients.ErrNoBatchAvailable
			}
			batch := &clients.TransferBatch{
				ID: batchID,
			}
			applyDummyFromMvxDepositsToBatch(cfg, batch)

			return batch, nil
		},
		GetLastExecutedEthBatchIDCalled: func(ctx context.Context) (uint64, error) {
			err := cfg.errorsOnCalls["GetLastExecutedEthBatchIDMvx"]
			if err != nil {
				return 0, err
			}

			return cfg.lastExecutedEthBatch, nil
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

			return returnBigIntOrZeroIfNil(cfg.totalBalancesOnEth), nil
		},
		MintBalancesCalled: func(ctx context.Context, account common.Address) (*big.Int, error) {
			err := cfg.errorsOnCalls["MintBalancesEth"]
			if err != nil {
				return nil, err
			}

			return returnBigIntOrZeroIfNil(cfg.mintBalancesOnEth), nil
		},
		BurnBalancesCalled: func(ctx context.Context, account common.Address) (*big.Int, error) {
			err := cfg.errorsOnCalls["BurnBalancesEth"]
			if err != nil {
				return nil, err
			}

			return returnBigIntOrZeroIfNil(cfg.burnBalancesOnEth), nil
		},
		GetBatchCalled: func(ctx context.Context, nonce uint64) (*clients.TransferBatch, error) {
			err := cfg.errorsOnCalls["GetBatchEth"]
			if err != nil {
				return nil, err
			}

			batch := &clients.TransferBatch{
				ID: nonce,
			}
			applyDummyFromEthDepositsToBatch(cfg, batch)

			return batch, nil
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

func applyDummyFromMvxDepositsToBatch(cfg testConfiguration, batch *clients.TransferBatch) {
	if cfg.amountsOnMvxPendingBatches != nil {
		values, found := cfg.amountsOnMvxPendingBatches[batch.ID]
		if found {
			depositCounter := uint64(0)

			for _, deposit := range values {
				batch.Deposits = append(batch.Deposits, &clients.DepositTransfer{
					Nonce:            depositCounter,
					Amount:           deposit,
					SourceTokenBytes: mvxToken,
				})
			}
		}
	}
}

func applyDummyFromEthDepositsToBatch(cfg testConfiguration, batch *clients.TransferBatch) {
	if cfg.amountsOnEthPendingBatches != nil {
		values, found := cfg.amountsOnEthPendingBatches[batch.ID]
		if found {
			depositCounter := uint64(0)

			for _, deposit := range values {
				batch.Deposits = append(batch.Deposits, &clients.DepositTransfer{
					Nonce:            depositCounter,
					Amount:           deposit,
					SourceTokenBytes: ethToken.Bytes(),
				})
			}
		}
	}
}

func getMaxMvxPendingBatchID(cfg testConfiguration) uint64 {
	if cfg.amountsOnMvxPendingBatches == nil {
		return 0
	}

	maxBatchIDFound := uint64(0)
	for batchID := range cfg.amountsOnMvxPendingBatches {
		if batchID > maxBatchIDFound {
			maxBatchIDFound = batchID
		}
	}

	return maxBatchIDFound
}

func returnBigIntOrZeroIfNil(value *big.Int) *big.Int {
	if value == nil {
		return big.NewInt(0)
	}

	return value
}
