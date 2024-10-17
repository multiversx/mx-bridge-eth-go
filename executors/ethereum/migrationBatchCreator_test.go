package ethereum

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/multiversx/mx-bridge-eth-go/testsCommon/bridge"
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/stretchr/testify/assert"
)

var safeContractAddress = common.HexToAddress(strings.Repeat("9", 40))
var tkn1Erc20Address = bytes.Repeat([]byte("2"), 20)
var tkn2Erc20Address = bytes.Repeat([]byte("3"), 20)
var balanceOfTkn1 = big.NewInt(19)
var balanceOfTkn2 = big.NewInt(38)
var expectedErr = errors.New("expected error")

func createMockArgsForMigrationBatchCreator() ArgsMigrationBatchCreator {
	return ArgsMigrationBatchCreator{
		MvxDataGetter: &bridge.DataGetterStub{
			GetAllKnownTokensCalled: func(ctx context.Context) ([][]byte, error) {
				return [][]byte{
					[]byte("tkn1"),
					[]byte("tkn2"),
				}, nil
			},
			GetERC20AddressForTokenIdCalled: func(ctx context.Context, tokenId []byte) ([][]byte, error) {
				return [][]byte{[]byte("erc 20 address")}, nil
			},
		},
		Erc20ContractsHolder: &bridge.ERC20ContractsHolderStub{},
		SafeContractAddress:  safeContractAddress,
		Logger:               &testscommon.LoggerStub{},
		EthereumChainWrapper: &bridge.EthereumClientWrapperStub{},
	}
}

func TestNewMigrationBatchCreator(t *testing.T) {
	t.Parallel()

	t.Run("nil mvx data getter should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsForMigrationBatchCreator()
		args.MvxDataGetter = nil

		creator, err := NewMigrationBatchCreator(args)
		assert.Nil(t, creator)
		assert.Equal(t, errNilMvxDataGetter, err)
	})
	t.Run("nil erc20 contracts holder should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsForMigrationBatchCreator()
		args.Erc20ContractsHolder = nil

		creator, err := NewMigrationBatchCreator(args)
		assert.Nil(t, creator)
		assert.Equal(t, errNilErc20ContractsHolder, err)
	})
	t.Run("nil logger should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsForMigrationBatchCreator()
		args.Logger = nil

		creator, err := NewMigrationBatchCreator(args)
		assert.Nil(t, creator)
		assert.Equal(t, errNilLogger, err)
	})
	t.Run("nil Ethereum chain wrapper should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsForMigrationBatchCreator()
		args.EthereumChainWrapper = nil

		creator, err := NewMigrationBatchCreator(args)
		assert.Nil(t, creator)
		assert.Equal(t, errNilEthereumChainWrapper, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsForMigrationBatchCreator()

		creator, err := NewMigrationBatchCreator(args)
		assert.NotNil(t, creator)
		assert.Nil(t, err)
	})
}

func TestFindAnUsableBatchID(t *testing.T) {
	unreachableBatchID := uint64(math.MaxUint64)

	t.Run("was batch used errors, should error", func(t *testing.T) {
		t.Parallel()

		result, err := testFindAnUsableBatchID(t, 1367, 1)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, expectedErr)
		assert.Contains(t, err.Error(), "on batch 1")

		result, err = testFindAnUsableBatchID(t, 1367, 100000)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, expectedErr)
		assert.Contains(t, err.Error(), "on batch 100000")

		result, err = testFindAnUsableBatchID(t, 1367, 50000)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, expectedErr)
		assert.Contains(t, err.Error(), "on batch 50000")
	})
	t.Run("should resolve in an optimum number of steps", func(t *testing.T) {
		t.Parallel()

		result, err := testFindAnUsableBatchID(t, 1367, unreachableBatchID)
		assert.Nil(t, err)
		expectedMap := map[uint64]int{
			1:      1, // initial low
			100000: 1, // initial high
			50000:  1, // (1 + 100000) / 2 = 50000
			25000:  1, // (1 + 50000) / 2 = 25000
			12500:  1, // (1 + 25000) / 2 = 12500
			6250:   1, // (1 + 12500) / 2 = 6250
			3125:   1, // (1 + 6250) / 2 = 3125
			1563:   1, // (1 + 3125) / 2 = 1563
			782:    1, // (1 + 1563) / 2 = 782
			1172:   1, // (782 + 1563) / 2 = 1172
			1367:   1, // (1172 + 1563) / 2 = 1367
			1269:   1, // (1172 + 1367) / 2 = 1269
			1318:   1, // (1269 + 1367) / 2 = 1318
			1342:   1, // (1318 + 1367) / 2 = 1342
			1354:   1, // (1342 + 1367) / 2 = 1354
			1360:   1, // (1354 + 1367) / 2 = 1360
			1363:   1, // (1360 + 1367) / 2 = 1363
			1365:   1, // (1363 + 1367) / 2 = 1365
			1366:   1, // (1365 + 1367) / 2 = 1366
		}

		assert.Equal(t, expectedMap, result)
	})
	t.Run("should resolve in an optimum number of steps if initial higher is not enough", func(t *testing.T) {
		t.Parallel()

		result, err := testFindAnUsableBatchID(t, 175000, unreachableBatchID)
		assert.Nil(t, err)
		expectedMap := map[uint64]int{
			1:      1, // initial low
			100000: 1, // initial high
			200000: 1, // added 100000 to the initial high
			150000: 1, // (100000 + 200000) / 2 = 150000
			175000: 1, // (150000 + 200000) / 2 = 175000
			162500: 1, // (150000 + 175000) / 2 = 162500
			168750: 1, // (162500 + 175000) / 2 = 168750
			171875: 1, // (168750 + 175000) / 2 = 171875
			173437: 1, // (171875 + 175000) / 2 = 173437
			174218: 1, // (173437 + 175000) / 2 = 174218
			174609: 1, // (174218 + 175000) / 2 = 174609
			174804: 1, // (174609 + 175000) / 2 = 174804
			174902: 1, // (174804 + 175000) / 2 = 174902
			174951: 1, // (174902 + 175000) / 2 = 174951
			174975: 1, // (174951 + 175000) / 2 = 174975
			174987: 1, // (174975 + 175000) / 2 = 174987
			174993: 1, // (174987 + 175000) / 2 = 174993
			174996: 1, // (174993 + 175000) / 2 = 174996
			174998: 1, // (174996 + 175000) / 2 = 174998
			174999: 1, // (174998 + 175000) / 2 = 174999
		}

		assert.Equal(t, expectedMap, result)
	})
	t.Run("should resolve in an optimum number of steps on 1", func(t *testing.T) {
		t.Parallel()

		result, err := testFindAnUsableBatchID(t, 1, unreachableBatchID)
		assert.Nil(t, err)
		expectedMap := map[uint64]int{
			1:      1, // initial low
			100000: 1, // initial high
		}

		assert.Equal(t, expectedMap, result)
	})
	t.Run("should resolve in an optimum number of steps on 2", func(t *testing.T) {
		t.Parallel()

		result, err := testFindAnUsableBatchID(t, 2, unreachableBatchID)
		assert.Nil(t, err)
		expectedMap := map[uint64]int{
			1:      1, // initial low
			100000: 1, // initial high
			50000:  1, // (1 + 100000) / 2 = 50000
			25000:  1, // (1 + 50000) / 2 = 25000
			12500:  1, // (1 + 25000) / 2 = 12500
			6250:   1, // (1 + 12500) / 2 = 6250
			3125:   1, // (1 + 6250) / 2 = 3125
			1563:   1, // (1 + 3125) / 2 = 1563
			782:    1, // (1 + 1563) / 2 = 782
			391:    1, // (1 + 782) / 2 = 391
			196:    1, // (1 + 391) / 2 = 196
			98:     1, // (1 + 196) / 2 = 98
			49:     1, // (1 + 98) / 2 = 49
			25:     1, // (1 + 49) / 2 = 25
			13:     1, // (1 + 25) / 2 = 13
			7:      1, // (1 + 13) / 2 = 7
			4:      1, // (1 + 7) / 2 = 4
			2:      1, // (1 + 4) / 2 = 2
		}

		assert.Equal(t, expectedMap, result)
	})
	t.Run("should resolve in an optimum number of steps on 100000", func(t *testing.T) {
		t.Parallel()

		result, err := testFindAnUsableBatchID(t, 100000, unreachableBatchID)
		assert.Nil(t, err)
		expectedMap := map[uint64]int{
			1:      1, // initial low
			100000: 1, // initial high
			50000:  1, // (1 + 100000) / 2 = 50000
			75000:  1, // (50000 + 100000) / 2 = 75000
			87500:  1, // (75000 + 100000) / 2 = 87500
			93750:  1, // (87500 + 100000) / 2 = 93750
			96875:  1, // (93750 + 100000) / 2 = 96875
			98437:  1, // (96875 + 100000) / 2 = 98437
			99218:  1, // (98437 + 100000) / 2 = 99218
			99609:  1, // (99218 + 100000) / 2 = 99609
			99804:  1, // (99609 + 100000) / 2 = 99804
			99902:  1, // (99804 + 100000) / 2 = 99902
			99951:  1, // (99902 + 100000) / 2 = 99951
			99975:  1, // (99951 + 100000) / 2 = 99975
			99987:  1, // (99975 + 100000) / 2 = 99987
			99993:  1, // (99987 + 100000) / 2 = 99993
			99996:  1, // (99993 + 100000) / 2 = 99996
			99998:  1, // (99996 + 100000) / 2 = 99998
			99999:  1, // (99998 + 100000) / 2 = 99999
		}

		assert.Equal(t, expectedMap, result)
	})
}

func testFindAnUsableBatchID(t *testing.T, firstFreeBatchId uint64, errorBatchID uint64) (map[uint64]int, error) {
	args := createMockArgsForMigrationBatchCreator()
	checkedMap := make(map[uint64]int)
	args.EthereumChainWrapper = &bridge.EthereumClientWrapperStub{
		WasBatchExecutedCalled: func(ctx context.Context, batchNonce *big.Int) (bool, error) {
			batchNonceUint64 := batchNonce.Uint64()

			if batchNonceUint64 == errorBatchID {
				return false, fmt.Errorf("%w on batch %d", expectedErr, batchNonceUint64)
			}

			checkedMap[batchNonceUint64]++
			return batchNonceUint64 < firstFreeBatchId, nil
		},
	}

	creator, _ := NewMigrationBatchCreator(args)
	batchID, err := creator.findAnUsableBatchID(context.Background(), 0)
	if err != nil {
		return nil, err
	}

	assert.Equal(t, firstFreeBatchId, batchID)

	return checkedMap, nil
}

func TestMigrationBatchCreator_CreateBatchInfo(t *testing.T) {
	t.Parallel()

	newSafeContractAddress := common.HexToAddress(strings.Repeat("8", 40))
	firstFreeBatchId := uint64(1367)
	t.Run("findAnUsableBatchID errors should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsForMigrationBatchCreator()
		args.EthereumChainWrapper = &bridge.EthereumClientWrapperStub{
			WasBatchExecutedCalled: func(ctx context.Context, batchNonce *big.Int) (bool, error) {
				return false, expectedErr
			},
		}

		creator, _ := NewMigrationBatchCreator(args)
		batch, err := creator.CreateBatchInfo(context.Background(), newSafeContractAddress, core.OptionalUint64{})
		assert.Equal(t, expectedErr, err)
		assert.Nil(t, batch)
	})
	t.Run("get all known tokens errors should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsForMigrationBatchCreator()
		args.EthereumChainWrapper = &bridge.EthereumClientWrapperStub{
			WasBatchExecutedCalled: func(ctx context.Context, batchNonce *big.Int) (bool, error) {
				return batchNonce.Uint64() < firstFreeBatchId, nil
			},
		}
		args.MvxDataGetter = &bridge.DataGetterStub{
			GetAllKnownTokensCalled: func(ctx context.Context) ([][]byte, error) {
				return nil, expectedErr
			},
		}

		creator, _ := NewMigrationBatchCreator(args)
		batch, err := creator.CreateBatchInfo(context.Background(), newSafeContractAddress, core.OptionalUint64{})
		assert.Equal(t, expectedErr, err)
		assert.Nil(t, batch)
	})
	t.Run("get all known tokens returns 0 tokens should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsForMigrationBatchCreator()
		args.EthereumChainWrapper = &bridge.EthereumClientWrapperStub{
			WasBatchExecutedCalled: func(ctx context.Context, batchNonce *big.Int) (bool, error) {
				return batchNonce.Uint64() < firstFreeBatchId, nil
			},
		}
		args.MvxDataGetter = &bridge.DataGetterStub{
			GetAllKnownTokensCalled: func(ctx context.Context) ([][]byte, error) {
				return make([][]byte, 0), nil
			},
		}

		creator, _ := NewMigrationBatchCreator(args)
		batch, err := creator.CreateBatchInfo(context.Background(), newSafeContractAddress, core.OptionalUint64{})
		assert.ErrorIs(t, err, errEmptyTokensList)
		assert.Nil(t, batch)
	})
	t.Run("GetERC20AddressForTokenId errors should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsForMigrationBatchCreator()
		args.EthereumChainWrapper = &bridge.EthereumClientWrapperStub{
			WasBatchExecutedCalled: func(ctx context.Context, batchNonce *big.Int) (bool, error) {
				return batchNonce.Uint64() < firstFreeBatchId, nil
			},
		}
		args.MvxDataGetter.(*bridge.DataGetterStub).GetERC20AddressForTokenIdCalled = func(ctx context.Context, sourceBytes []byte) ([][]byte, error) {
			return nil, expectedErr
		}

		creator, _ := NewMigrationBatchCreator(args)
		batch, err := creator.CreateBatchInfo(context.Background(), newSafeContractAddress, core.OptionalUint64{})
		assert.Equal(t, expectedErr, err)
		assert.Nil(t, batch)
	})
	t.Run("GetERC20AddressForTokenId returns empty list should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsForMigrationBatchCreator()
		args.EthereumChainWrapper = &bridge.EthereumClientWrapperStub{
			WasBatchExecutedCalled: func(ctx context.Context, batchNonce *big.Int) (bool, error) {
				return batchNonce.Uint64() < firstFreeBatchId, nil
			},
		}
		args.MvxDataGetter.(*bridge.DataGetterStub).GetERC20AddressForTokenIdCalled = func(ctx context.Context, sourceBytes []byte) ([][]byte, error) {
			return make([][]byte, 0), nil
		}

		creator, _ := NewMigrationBatchCreator(args)
		batch, err := creator.CreateBatchInfo(context.Background(), newSafeContractAddress, core.OptionalUint64{})
		assert.ErrorIs(t, err, errWrongERC20AddressResponse)
		assert.Nil(t, batch)
	})
	t.Run("BalanceOf errors should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsForMigrationBatchCreator()
		args.EthereumChainWrapper = &bridge.EthereumClientWrapperStub{
			WasBatchExecutedCalled: func(ctx context.Context, batchNonce *big.Int) (bool, error) {
				return batchNonce.Uint64() < firstFreeBatchId, nil
			},
		}
		args.Erc20ContractsHolder = &bridge.ERC20ContractsHolderStub{
			BalanceOfCalled: func(ctx context.Context, erc20Address common.Address, address common.Address) (*big.Int, error) {
				return nil, expectedErr
			},
		}

		creator, _ := NewMigrationBatchCreator(args)
		batch, err := creator.CreateBatchInfo(context.Background(), newSafeContractAddress, core.OptionalUint64{})
		assert.ErrorIs(t, err, expectedErr)
		assert.Nil(t, batch)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsForMigrationBatchCreator()
		args.MvxDataGetter.(*bridge.DataGetterStub).GetERC20AddressForTokenIdCalled = func(ctx context.Context, sourceBytes []byte) ([][]byte, error) {
			if string(sourceBytes) == "tkn1" {
				return [][]byte{tkn1Erc20Address}, nil
			}
			if string(sourceBytes) == "tkn2" {
				return [][]byte{tkn2Erc20Address}, nil
			}

			return nil, fmt.Errorf("unexpected source bytes")
		}
		args.Erc20ContractsHolder = &bridge.ERC20ContractsHolderStub{
			BalanceOfCalled: func(ctx context.Context, erc20Address common.Address, address common.Address) (*big.Int, error) {
				assert.Equal(t, address.String(), safeContractAddress.String())

				if string(erc20Address.Bytes()) == string(tkn1Erc20Address) {
					return balanceOfTkn1, nil
				}
				if string(erc20Address.Bytes()) == string(tkn2Erc20Address) {
					return balanceOfTkn2, nil
				}

				return nil, fmt.Errorf("unexpected ERC20 contract address")
			},
		}
		args.EthereumChainWrapper = &bridge.EthereumClientWrapperStub{
			WasBatchExecutedCalled: func(ctx context.Context, batchNonce *big.Int) (bool, error) {
				return batchNonce.Uint64() < firstFreeBatchId, nil
			},
		}
		creator, _ := NewMigrationBatchCreator(args)

		t.Run("without trim", func(t *testing.T) {
			expectedBatch := &BatchInfo{
				OldSafeContractAddress: safeContractAddress.String(),
				NewSafeContractAddress: newSafeContractAddress.String(),
				BatchID:                firstFreeBatchId,
				MessageHash:            common.HexToHash("0x5650b9dcc3283c624328422a6a41dd3305f86c5456762f63110a2fc4e23f5162"),
				DepositsInfo: []*DepositInfo{
					{
						DepositNonce:          1,
						Token:                 "tkn1",
						ContractAddressString: common.BytesToAddress(tkn1Erc20Address).String(),
						ContractAddress:       common.BytesToAddress(tkn1Erc20Address),
						Amount:                big.NewInt(19),
						AmountString:          "19",
					},
					{
						DepositNonce:          2,
						Token:                 "tkn2",
						ContractAddressString: common.BytesToAddress(tkn2Erc20Address).String(),
						ContractAddress:       common.BytesToAddress(tkn2Erc20Address),
						Amount:                big.NewInt(38),
						AmountString:          "38",
					},
				},
			}

			batch, err := creator.CreateBatchInfo(context.Background(), newSafeContractAddress, core.OptionalUint64{})
			assert.Nil(t, err)
			assert.Equal(t, expectedBatch, batch)
		})
		t.Run("with trim", func(t *testing.T) {
			expectedBatch := &BatchInfo{
				OldSafeContractAddress: safeContractAddress.String(),
				NewSafeContractAddress: newSafeContractAddress.String(),
				BatchID:                firstFreeBatchId,
				MessageHash:            common.HexToHash("0x7dc48af8b0431d100adefaed79bacd0c33ab0fdcc11723de6eaa3f158595a097"),
				DepositsInfo: []*DepositInfo{
					{
						DepositNonce:          1,
						Token:                 "tkn1",
						ContractAddressString: common.BytesToAddress(tkn1Erc20Address).String(),
						ContractAddress:       common.BytesToAddress(tkn1Erc20Address),
						Amount:                big.NewInt(19),
						AmountString:          "19",
					},
					{
						DepositNonce:          2,
						Token:                 "tkn2",
						ContractAddressString: common.BytesToAddress(tkn2Erc20Address).String(),
						ContractAddress:       common.BytesToAddress(tkn2Erc20Address),
						Amount:                big.NewInt(20),
						AmountString:          "20",
					},
				},
			}

			trimValue := core.OptionalUint64{
				Value:    20,
				HasValue: true,
			}
			batch, err := creator.CreateBatchInfo(context.Background(), newSafeContractAddress, trimValue)
			assert.Nil(t, err)
			assert.Equal(t, expectedBatch, batch)
		})
	})

}
