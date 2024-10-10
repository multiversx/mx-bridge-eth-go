package ethereum

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
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
		SafeContractWrapper:  &bridge.SafeContractWrapperStub{},
		Logger:               &testscommon.LoggerStub{},
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
	t.Run("nil safe contract wrapper should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsForMigrationBatchCreator()
		args.SafeContractWrapper = nil

		creator, err := NewMigrationBatchCreator(args)
		assert.Nil(t, creator)
		assert.Equal(t, errNilSafeContractWrapper, err)
	})
	t.Run("nil logger should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsForMigrationBatchCreator()
		args.Logger = nil

		creator, err := NewMigrationBatchCreator(args)
		assert.Nil(t, creator)
		assert.Equal(t, errNilLogger, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsForMigrationBatchCreator()

		creator, err := NewMigrationBatchCreator(args)
		assert.NotNil(t, creator)
		assert.Nil(t, err)
	})
}

func TestMigrationBatchCreator_CreateBatchInfo(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("expected error")
	newSafeContractAddress := common.HexToAddress(strings.Repeat("8", 40))
	t.Run("BatchesCount errors should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsForMigrationBatchCreator()
		args.SafeContractWrapper = &bridge.SafeContractWrapperStub{
			BatchesCountCalled: func(opts *bind.CallOpts) (uint64, error) {
				return 0, expectedErr
			},
		}

		creator, _ := NewMigrationBatchCreator(args)
		batch, err := creator.CreateBatchInfo(context.Background(), newSafeContractAddress, core.OptionalUint64{})
		assert.Equal(t, expectedErr, err)
		assert.Nil(t, batch)
	})
	t.Run("DepositsCount errors should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsForMigrationBatchCreator()
		args.SafeContractWrapper = &bridge.SafeContractWrapperStub{
			DepositsCountCalled: func(opts *bind.CallOpts) (uint64, error) {
				return 0, expectedErr
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

		depositCountStart := uint64(39)
		depositCount := depositCountStart
		batchesCount := uint64(2244)
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
		args.SafeContractWrapper = &bridge.SafeContractWrapperStub{
			DepositsCountCalled: func(opts *bind.CallOpts) (uint64, error) {
				depositCountValue := atomic.AddUint64(&depositCount, 1)

				return depositCountValue, nil

			},
			BatchesCountCalled: func(opts *bind.CallOpts) (uint64, error) {
				return batchesCount, nil
			},
		}
		creator, _ := NewMigrationBatchCreator(args)

		t.Run("without trim", func(t *testing.T) {
			atomic.StoreUint64(&depositCount, depositCountStart)
			expectedBatch := &BatchInfo{
				OldSafeContractAddress: safeContractAddress.String(),
				NewSafeContractAddress: newSafeContractAddress.String(),
				BatchID:                2245,
				MessageHash:            common.HexToHash("0xe87c7ee013d37956c0023c6a07dce7941a3932293d1b98ab3f00cbde5eae93be"),
				DepositsInfo: []*DepositInfo{
					{
						DepositNonce:          41,
						Token:                 "tkn1",
						ContractAddressString: common.BytesToAddress(tkn1Erc20Address).String(),
						ContractAddress:       common.BytesToAddress(tkn1Erc20Address),
						Amount:                big.NewInt(19),
						AmountString:          "19",
					},
					{
						DepositNonce:          42,
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
			atomic.StoreUint64(&depositCount, depositCountStart)
			expectedBatch := &BatchInfo{
				OldSafeContractAddress: safeContractAddress.String(),
				NewSafeContractAddress: newSafeContractAddress.String(),
				BatchID:                2245,
				MessageHash:            common.HexToHash("0xfa4c46fc0d0b75460d376a03723b2543aac07d64c47f5322b1a506663bcd266d"),
				DepositsInfo: []*DepositInfo{
					{
						DepositNonce:          41,
						Token:                 "tkn1",
						ContractAddressString: common.BytesToAddress(tkn1Erc20Address).String(),
						ContractAddress:       common.BytesToAddress(tkn1Erc20Address),
						Amount:                big.NewInt(19),
						AmountString:          "19",
					},
					{
						DepositNonce:          42,
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
