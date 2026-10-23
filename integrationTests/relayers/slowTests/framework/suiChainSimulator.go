package framework

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"os/exec"
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

func (s *suiChainSimulatorWrapper) GetTokenBalance(ctx context.Context, owner string, tokenType string) *big.Int {
	resp, err := s.proxy.SuiXGetOwnedObjects(ctx, models.SuiXGetOwnedObjectsRequest{
		Address: owner,
		Limit:   50,
		Query: models.SuiObjectResponseQuery{
			Filter: models.SuiObjectDataFilter{
				"StructType": fmt.Sprintf("0x2::token::Token<%s>", tokenType),
			},
			Options: models.SuiObjectDataOptions{
				ShowContent:       true,
				ShowBcs:           true,
				ShowOwner:         true,
				ShowType:          true,
				ShowDisplay:       true,
				ShowStorageRebate: true,
			},
		},
	})
	require.NoError(s, err)

	log.Info("GetTokenBalance", "owner", owner, "tokenType", tokenType, "numObjects", len(resp.Data))
	totalBalance := big.NewInt(0)

	for _, obj := range resp.Data {
		if obj.Data == nil {
			continue
		}
		log.Info("GetTokenBalance object", "type", obj.Data.Type, "objectId", obj.Data.ObjectId)
		if obj.Data.Content == nil {
			continue
		}
		balanceStr, ok := obj.Data.Content.Fields["balance"].(string)
		if !ok {
			// Token<T> stores balance as Balance<T> struct {"value": "..."}, not a plain string
			if balanceMap, ok2 := obj.Data.Content.Fields["balance"].(map[string]interface{}); ok2 {
				balanceStr, ok = balanceMap["value"].(string)
			}
		}
		if !ok {
			continue
		}
		bigIntBalance, ok := big.NewInt(0).SetString(balanceStr, 10)
		require.True(s, ok)
		totalBalance.Add(totalBalance, bigIntBalance)
	}

	return totalBalance
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

func (s *suiChainSimulatorWrapper) TransferObject(ctx context.Context, request models.TransferObjectRequest) models.SuiTransactionBlockResponse {
	txMeta, err := s.proxy.TransferObject(ctx, request)
	require.NoError(s, err)

	return s.signAndExecuteTxReturnResult(ctx, txMeta, s.owner.SuiSK)
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

// BuildAndPublish compiles a Move package locally (no network) and publishes via JSON-RPC.
// This avoids the gRPC transport issue where sui client publish uses gRPC but the local
// test node only exposes JSON-RPC on port 9000.
func (s *suiChainSimulatorWrapper) BuildAndPublish(ctx context.Context, packageDir string, signer KeysHolder) models.SuiTransactionBlockResponse {
	// Build step: purely local compilation, no network needed.
	// SUI_CONFIG_DIR points to a temp dir with a minimal client.yaml so the CLI doesn't
	// create one interactively. We keep the real HOME so ~/.move package cache is reused.
	tmpCfg, err := os.MkdirTemp("", "sui-cfg-*")
	require.NoError(s, err)
	defer func() { _ = os.RemoveAll(tmpCfg) }()

	keystorePath := tmpCfg + "/sui.keystore"
	err = os.WriteFile(keystorePath, []byte("[]"), 0600)
	require.NoError(s, err)
	clientYaml := fmt.Sprintf("---\nkeystore:\n  File: %s\nenvs:\n  - alias: testnet\n    rpc: \"%s\"\n    ws: ~\n    basic_auth: ~\nactive_env: testnet\nactive_address: \"0x0000000000000000000000000000000000000000000000000000000000000000\"\n", keystorePath, networkUrl)
	err = os.WriteFile(tmpCfg+"/client.yaml", []byte(clientYaml), 0600)
	require.NoError(s, err)

	var stdout, stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, "sui", "move", "build",
		"--dump-bytecode-as-base64", "--no-tree-shaking")
	cmd.Dir = packageDir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	// SUI_CONFIG_DIR overrides config location; keep HOME so ~/.move cache is available
	cmd.Env = append(os.Environ(), "SUI_CONFIG_DIR="+tmpCfg)

	err = cmd.Run()
	require.NoError(s, err, "sui move build failed.\nstdout: %s\nstderr: %s", stdout.String(), stderr.String())

	// Parse the JSON build output. The build progress/warnings go to stderr;
	// only the JSON is on stdout.
	var buildResult struct {
		Modules      []string `json:"modules"`
		Dependencies []string `json:"dependencies"`
	}
	rawJSON := stdout.Bytes()
	if idx := bytes.IndexByte(rawJSON, '{'); idx > 0 {
		rawJSON = rawJSON[idx:]
	}
	err = json.Unmarshal(rawJSON, &buildResult)
	require.NoError(s, err, "failed to parse sui move build JSON output: %s", stdout.String())
	require.NotEmpty(s, buildResult.Modules, "sui move build returned no modules")

	// Publish step: submit via JSON-RPC (the Go SDK path), bypassing gRPC.
	return s.PublishPackage(ctx, models.PublishRequest{
		Sender:          string(signer.SuiAddress),
		CompiledModules: buildResult.Modules,
		Dependencies:    buildResult.Dependencies,
		GasBudget:       "500000000",
	}, signer)
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
