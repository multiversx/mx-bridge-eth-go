package framework

import (
	"context"
	"crypto/ed25519"
	"fmt"
	"math/big"
	"testing"

	"github.com/block-vision/sui-go-sdk/models"
	"github.com/block-vision/sui-go-sdk/sui"
	suiClient "github.com/multiversx/mx-bridge-eth-go/clients/sui"
	"github.com/stretchr/testify/require"
)

const (
	networkUrl = "http://localhost:9000"
	faucetUrl  = "http://localhost:9123"
)

type ArgsSuiChainSimulatorWrapper struct {
	TB    testing.TB
	Owner KeysHolder
}

type suiChainSimulatorWrapper struct {
	testing.TB
	proxy suiClient.Proxy
	owner KeysHolder
}

func CreateSuiChainSimulatorWrapper(args ArgsSuiChainSimulatorWrapper) *suiChainSimulatorWrapper {
	wrapper := &suiChainSimulatorWrapper{
		TB:    args.TB,
		proxy: sui.NewSuiClient(networkUrl),
		owner: args.Owner,
	}
	return wrapper
}

func (s *suiChainSimulatorWrapper) PublishPackage(ctx context.Context, request models.PublishRequest, signer KeysHolder) models.SuiTransactionBlockResponse {
	txMeta, err := s.proxy.Publish(ctx, request)
	require.NoError(s, err)

	return s.signAndExecuteTxReturnResult(ctx, txMeta, signer.SuiSK)
}

func (s *suiChainSimulatorWrapper) MoveCall(ctx context.Context, request models.MoveCallRequest, signer KeysHolder) models.SuiTransactionBlockResponse {
	txMeta, err := s.proxy.MoveCall(ctx, request)
	require.NoError(s, err)

	return s.signAndExecuteTxReturnResult(ctx, txMeta, signer.SuiSK)
}

func (s *suiChainSimulatorWrapper) GetCoinBalance(ctx context.Context, owner string, coinType string) *big.Int {
	balance, err := s.proxy.SuiXGetBalance(ctx, models.SuiXGetBalanceRequest{
		Owner:    owner,
		CoinType: coinType,
	})
	require.NoError(s, err)

	bigIntBalance, ok := big.NewInt(0).SetString(balance.TotalBalance, 10)
	require.True(s, ok)

	return bigIntBalance
}

func (s *suiChainSimulatorWrapper) GetCoins(ctx context.Context, owner string, coinType string) []models.CoinData {
	coins, err := s.proxy.SuiXGetCoins(ctx, models.SuiXGetCoinsRequest{
		Owner:    owner,
		CoinType: coinType,
	})
	require.NoError(s, err)
	return coins.Data
}

func (s *suiChainSimulatorWrapper) SplitCoin(ctx context.Context, request models.SplitCoinRequest, signer KeysHolder) models.SuiTransactionBlockResponse {
	txMeta, err := s.proxy.SplitCoin(ctx, request)
	require.NoError(s, err)

	return s.signAndExecuteTxReturnResult(ctx, txMeta, signer.SuiSK)
}

func (s *suiChainSimulatorWrapper) FundWallets(wallets [][]byte) {
	for _, wallet := range wallets {
		header := map[string]string{}
		err := sui.RequestSuiFromFaucet(faucetUrl, string(wallet), header)
		if err != nil {
			log.Error("error in suiChainSimulatorWrapper.FundWallets", "error", err)
		}
		log.Info("Funded wallet: " + string(wallet))
	}
}

func (s *suiChainSimulatorWrapper) GenerateBlocks(ctx context.Context, numBlocks int) {
	for i := 0; i < numBlocks; i++ {
		address := string(s.owner.SuiAddress)

		coins, err := s.proxy.SuiXGetCoins(ctx, models.SuiXGetCoinsRequest{
			Owner:    address,
			CoinType: "0x2::sui::SUI",
			Limit:    5,
		})
		require.NoError(s, err)
		require.True(s, len(coins.Data) > 0, "No coins found for address: "+address)

		pay, err := s.proxy.Pay(ctx, models.PayRequest{
			Signer:      address,
			SuiObjectId: []string{coins.Data[0].CoinObjectId},
			Recipient:   []string{address},
			Amount:      []string{"100"},
			GasBudget:   "10000000",
		})
		require.NoError(s, err)

		resp, err := s.proxy.SignAndExecuteTransactionBlock(
			ctx,
			models.SignAndExecuteTransactionBlockRequest{
				TxnMetaData: pay,
				PriKey:      s.owner.SuiSK,
				Options:     models.SuiTransactionBlockOptions{ShowEffects: true},
				RequestType: "WaitForLocalExecution",
			},
		)
		require.NoError(s, err)
		require.Equal(s, "success", resp.Effects.Status.Status)
	}
}

func (s *suiChainSimulatorWrapper) signAndExecuteTxReturnResult(
	ctx context.Context,
	txMeta models.TxnMetaData,
	signerPriKey ed25519.PrivateKey,
) models.SuiTransactionBlockResponse {
	exec, err := s.proxy.SignAndExecuteTransactionBlock(ctx, models.SignAndExecuteTransactionBlockRequest{
		TxnMetaData: txMeta,
		PriKey:      signerPriKey,
		Options: models.SuiTransactionBlockOptions{
			ShowEffects:       true,
			ShowObjectChanges: true,
		},
		RequestType: "WaitForLocalExecution",
	})

	require.Nil(s, err)
	require.Equal(s, "success", exec.Effects.Status.Status, fmt.Sprintf("Error: %s", exec.Effects.Status.Error))

	return exec
}
