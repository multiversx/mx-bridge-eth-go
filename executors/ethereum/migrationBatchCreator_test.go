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
	"github.com/stretchr/testify/assert"
)

var safeContractAddress = common.HexToAddress(strings.Repeat("9", 40))
var tkn1Erc20Address = bytes.Repeat([]byte("2"), 20)
var tkn2Erc20Address = bytes.Repeat([]byte("3"), 20)
var balanceOfTkn1 = big.NewInt(37)
var balanceOfTkn2 = big.NewInt(38)

func createMockArgsForMigrationBatchCreator() ArgsMigrationBatchCreator {
	return ArgsMigrationBatchCreator{
		TokensList:           []string{"tkn1", "tkn2"},
		TokensMapper:         &bridge.TokensMapperStub{},
		Erc20ContractsHolder: &bridge.ERC20ContractsHolderStub{},
		SafeContractAddress:  safeContractAddress,
		SafeContractWrapper:  &bridge.SafeContractWrapperStub{},
	}
}

func TestNewMigrationBatchCreator(t *testing.T) {
	t.Parallel()

	t.Run("nil or empty tokens list should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsForMigrationBatchCreator()
		args.TokensList = nil

		creator, err := NewMigrationBatchCreator(args)
		assert.Nil(t, creator)
		assert.Equal(t, errEmptyTokensList, err)

		args = createMockArgsForMigrationBatchCreator()
		args.TokensList = make([]string, 0)

		creator, err = NewMigrationBatchCreator(args)
		assert.Nil(t, creator)
		assert.Equal(t, errEmptyTokensList, err)
	})
	t.Run("nil tokens mapper should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsForMigrationBatchCreator()
		args.TokensMapper = nil

		creator, err := NewMigrationBatchCreator(args)
		assert.Nil(t, creator)
		assert.Equal(t, errNilTokensMapper, err)
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
	t.Run("ConvertToken errors should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsForMigrationBatchCreator()
		args.TokensMapper = &bridge.TokensMapperStub{
			ConvertTokenCalled: func(ctx context.Context, sourceBytes []byte) ([]byte, error) {
				return nil, expectedErr
			},
		}

		creator, _ := NewMigrationBatchCreator(args)
		batch, err := creator.CreateBatchInfo(context.Background(), newSafeContractAddress)
		assert.Equal(t, expectedErr, err)
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
		args.TokensMapper = &bridge.TokensMapperStub{
			ConvertTokenCalled: func(ctx context.Context, sourceBytes []byte) ([]byte, error) {
				if string(sourceBytes) == "tkn1" {
					return tkn1Erc20Address, nil
				}
				if string(sourceBytes) == "tkn2" {
					return tkn2Erc20Address, nil
				}

				return nil, fmt.Errorf("unexpected source bytes")
			},
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
