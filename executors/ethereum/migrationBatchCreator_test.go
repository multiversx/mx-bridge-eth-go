package ethereum

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/multiversx/mx-bridge-eth-go/testsCommon/bridge"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/stretchr/testify/assert"
)

var safeContractAddress = common.HexToAddress(strings.Repeat("9", 40))
var tkn1Erc20Address = bytes.Repeat([]byte("2"), 20)
var tkn2Erc20Address = bytes.Repeat([]byte("3"), 20)
var balanceOfTkn1 = big.NewInt(37)
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
		batch, err := creator.CreateBatchInfo(context.Background(), newSafeContractAddress)
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
		batch, err := creator.CreateBatchInfo(context.Background(), newSafeContractAddress)
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
		batch, err := creator.CreateBatchInfo(context.Background(), newSafeContractAddress)
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
		batch, err := creator.CreateBatchInfo(context.Background(), newSafeContractAddress)
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
		batch, err := creator.CreateBatchInfo(context.Background(), newSafeContractAddress)
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
		batch, err := creator.CreateBatchInfo(context.Background(), newSafeContractAddress)
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
		batch, err := creator.CreateBatchInfo(context.Background(), newSafeContractAddress)
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
				depositCountValue := depositCount
				depositCount++

				return depositCountValue, nil

			},
			BatchesCountCalled: func(opts *bind.CallOpts) (uint64, error) {
				return batchesCount, nil
			},
		}
		creator, _ := NewMigrationBatchCreator(args)

		expectedBatch := &BatchInfo{
			OldSafeContractAddress: safeContractAddress.String(),
			NewSafeContractAddress: newSafeContractAddress.String(),
			BatchID:                2245,
			MessageHash:            common.HexToHash("0x93915c0bea665553dfc85ec3cdf4b883100929f22d6cbbfc44db2f0ee71b3b56"),
			DepositsInfo: []*DepositInfo{
				{
					DepositNonce:    40,
					Token:           "tkn1",
					ContractAddress: common.BytesToAddress(tkn1Erc20Address).String(),
					contractAddress: common.BytesToAddress(tkn1Erc20Address),
					amount:          big.NewInt(37),
					Amount:          "37",
				},
				{
					DepositNonce:    41,
					Token:           "tkn2",
					ContractAddress: common.BytesToAddress(tkn2Erc20Address).String(),
					contractAddress: common.BytesToAddress(tkn2Erc20Address),
					amount:          big.NewInt(38),
					Amount:          "38",
				},
			},
		}

		batch, err := creator.CreateBatchInfo(context.Background(), newSafeContractAddress)
		assert.Nil(t, err)
		assert.Equal(t, expectedBatch, batch)
	})

}
