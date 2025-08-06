package framework

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math/big"
	"os"
	"strings"
	"testing"

	"github.com/block-vision/sui-go-sdk/models"
	suiSdk "github.com/block-vision/sui-go-sdk/sui"
	"github.com/block-vision/sui-go-sdk/utils"
	"github.com/multiversx/mx-sdk-go/core"
	"github.com/stretchr/testify/require"
)

const (
	suiSafeBytecode          = "testdata/contracts/sui/safe.mv"
	suiBridgeBytecode        = "testdata/contracts/sui/bridge.mv"
	suiEventsBytecode        = "testdata/contracts/sui/events.mv"
	suiPausableBytecode      = "testdata/contracts/sui/pausable.mv"
	suiRolesBytecode         = "testdata/contracts/sui/roles.mv"
	suiSharedStructsBytecode = "testdata/contracts/sui/shared_structs.mv"
	suiUtilsBytecode         = "testdata/contracts/sui/utils.mv"
	suiTestCoinBytecode      = "testdata/contracts/sui/test_coin.mv"
)

// SuiHandler will handle all the operations on the Sui side
type SuiHandler struct {
	testing.TB
	*KeysStore
	TokensRegistry             TokensRegistry
	Quorum                     string
	MvxTestCallerAddress       core.AddressHandler
	SuiChainSimulator          *suiChainSimulatorWrapper
	SuiProxy                   suiSdk.ISuiAPI
	PackageID                  string
	BridgeObjectID             string
	BridgeCap                  string
	AdminCap                   string
	SafeObjectID               string
	BridgeInitialSharedVersion uint64
	SafeInitialSharedVersion   uint64
}

// NewSuiHandler will create the handler that will adapt all test operations on Sui
func NewSuiHandler(
	tb testing.TB,
	_ context.Context, // ctx is unused
	keysStore *KeysStore,
	tokensRegistry TokensRegistry,
	quorum string,
) *SuiHandler {
	handler := &SuiHandler{
		TB:             tb,
		KeysStore:      keysStore,
		TokensRegistry: tokensRegistry,
		Quorum:         quorum,
	}

	// TODO: chain simulator
	walletsToFundOnSui := handler.WalletsToFundOnSui()
	handler.FundWallets(walletsToFundOnSui)
	handler.SuiProxy = suiSdk.NewSuiClient("http://localhost:9000")

	return handler
}

func (handler *SuiHandler) DeployContracts(ctx context.Context) {
	txMeta, err := handler.SuiProxy.Publish(ctx, models.PublishRequest{
		Sender:          string(handler.OwnerKeys.SuiAddress),
		CompiledModules: handler.getEncodedModules(),
		Dependencies: []string{
			"0x1", // Sui Framework
			"0x2", // Move Standard Library
		},
		GasBudget: "500000000",
	})
	require.NoError(handler, err)

	resp := handler.signAndExecuteTxReturnResult(ctx, txMeta, handler.OwnerKeys.SuiSK)

	for _, obj := range resp.ObjectChanges {
		if obj.Type == "created" {
			if strings.Contains(obj.ObjectType, "::roles::BridgeCap") {
				handler.BridgeCap = obj.ObjectId
			}
			if strings.Contains(obj.ObjectType, "::roles::AdminCap") {
				handler.AdminCap = obj.ObjectId
			}
			if strings.Contains(obj.ObjectType, "::safe::BridgeSafe") {
				handler.SafeObjectID = obj.ObjectId
			}

		} else if obj.Type == "published" {
			handler.PackageID = obj.PackageId
		}
	}

	suiRelayersAddresses := make([]string, 0, len(handler.RelayersKeys))
	suiRelayersPubKeys := make([]ed25519.PublicKey, 0, len(handler.RelayersKeys))
	for _, relayerKeys := range handler.RelayersKeys {
		suiRelayersAddresses = append(suiRelayersAddresses, string(relayerKeys.SuiAddress))
		suiRelayersPubKeys = append(suiRelayersPubKeys, relayerKeys.SuiSK.Public().(ed25519.PublicKey))
	}

	bridgeIdBytes := handler.DeployContract(
		ctx,
		"bridge",
		"initialize",
		suiRelayersAddresses,
		suiRelayersPubKeys,
		handler.Quorum,
		handler.SafeObjectID,
		handler.BridgeCap,
	)
	handler.BridgeObjectID = string(bridgeIdBytes)
}

func (handler *SuiHandler) getEncodedModules() []string {
	var modules []string

	mv := handler.readModuleBytes(suiSafeBytecode)
	modules = append(modules, base64.StdEncoding.EncodeToString(mv))

	mv = handler.readModuleBytes(suiBridgeBytecode)
	modules = append(modules, base64.StdEncoding.EncodeToString(mv))

	mv = handler.readModuleBytes(suiEventsBytecode)
	modules = append(modules, base64.StdEncoding.EncodeToString(mv))

	mv = handler.readModuleBytes(suiPausableBytecode)
	modules = append(modules, base64.StdEncoding.EncodeToString(mv))

	mv = handler.readModuleBytes(suiRolesBytecode)
	modules = append(modules, base64.StdEncoding.EncodeToString(mv))

	mv = handler.readModuleBytes(suiSharedStructsBytecode)
	modules = append(modules, base64.StdEncoding.EncodeToString(mv))

	mv = handler.readModuleBytes(suiUtilsBytecode)
	modules = append(modules, base64.StdEncoding.EncodeToString(mv))

	return modules
}

func (handler *SuiHandler) readModuleBytes(path string) []byte {
	b, err := os.ReadFile(path)
	require.NoError(handler, err)
	return b
}

func (handler *SuiHandler) DeployContract(
	ctx context.Context,
	params ...interface{},
) []byte {
	module := params[0].(string)
	function := params[1].(string)
	suiRelayerAddresses := params[2].([]string)
	suiRelayersPubKeys := params[3].([]ed25519.PublicKey)
	quorumStr := params[4].(string)
	safeObjectID := params[5].(string)
	adminCap := params[6].(string)

	txMeta, err := handler.SuiProxy.MoveCall(ctx, models.MoveCallRequest{
		Signer:          string(handler.OwnerKeys.SuiAddress),
		PackageObjectId: handler.PackageID,
		Module:          module,
		Function:        function,
		TypeArguments:   []interface{}{},
		Arguments: []interface{}{
			suiRelayerAddresses,
			suiRelayersPubKeys,
			quorumStr,
			safeObjectID,
			adminCap,
		},
		GasBudget: "100000000",
	})
	require.Nil(handler, err)

	resp := handler.signAndExecuteTxReturnResult(ctx, txMeta, handler.OwnerKeys.SuiSK)
	for _, obj := range resp.ObjectChanges {
		if obj.Type == "created" {
			if strings.Contains(obj.ObjectType, "::bridge::Bridge") {
				return []byte(obj.ObjectId)
			}
		}
	}
	return nil
}

// DeployUpgradeableContract not implemented on Sui chain
func (handler *SuiHandler) DeployUpgradeableContract(_ context.Context, _ ...interface{}) []byte {
	panic("Not implemented for Sui")
}

// GetBalance returns the receiver's balance
func (handler *SuiHandler) GetBalance(ctx context.Context, receiver []byte, abstractTokenIdentifier string) *big.Int {
	token := handler.TokensRegistry.GetTokenData(abstractTokenIdentifier)
	require.NotNil(handler, token)
	require.NotNil(handler, token.PeerChainTokenAddress)

	coinType := fmt.Sprintf("0x%s::test_coin::TEST_COIN", hex.EncodeToString(token.PeerChainTokenAddress))

	balance, err := handler.SuiProxy.SuiXGetBalance(ctx, models.SuiXGetBalanceRequest{
		Owner:    string(receiver),
		CoinType: coinType,
	})
	require.NoError(handler, err)

	bigIntBalance, ok := big.NewInt(0).SetString(balance.TotalBalance, 10)
	require.True(handler, ok)

	return bigIntBalance
}

// UnPauseContractsAfterTokenChanges can unpause contracts after token changes
func (handler *SuiHandler) UnPauseContractsAfterTokenChanges(ctx context.Context) {
	// unpause bridge contract
	txMeta, err := handler.SuiProxy.MoveCall(ctx, models.MoveCallRequest{
		Signer:          string(handler.OwnerKeys.SuiAddress),
		PackageObjectId: handler.PackageID,
		Module:          "bridge",
		Function:        "unpause_contract",
		TypeArguments:   []interface{}{},
		Arguments: []interface{}{
			handler.BridgeObjectID,
			handler.AdminCap,
		},
		GasBudget: "10000000",
	})
	require.NoError(handler, err)

	handler.signAndExecuteTxReturnResult(ctx, txMeta, handler.OwnerKeys.SuiSK)

	// unpause safe contract
	txMeta, err = handler.SuiProxy.MoveCall(ctx, models.MoveCallRequest{
		Signer:          string(handler.OwnerKeys.SuiAddress),
		PackageObjectId: handler.PackageID,
		Module:          "safe",
		Function:        "unpause_contract",
		TypeArguments:   []interface{}{},
		Arguments: []interface{}{
			handler.SafeObjectID,
			handler.AdminCap,
		},
		GasBudget: "10000000",
	})
	require.NoError(handler, err)

	handler.signAndExecuteTxReturnResult(ctx, txMeta, handler.OwnerKeys.SuiSK)
}

// PauseContractsForTokenChanges can pause contracts for token changes
func (handler *SuiHandler) PauseContractsForTokenChanges(ctx context.Context) {
	// unpause bridge contract
	txMeta, err := handler.SuiProxy.MoveCall(ctx, models.MoveCallRequest{
		Signer:          string(handler.OwnerKeys.SuiAddress),
		PackageObjectId: handler.PackageID,
		Module:          "bridge",
		Function:        "pause_contract",
		TypeArguments:   []interface{}{},
		Arguments: []interface{}{
			handler.BridgeObjectID,
			handler.AdminCap,
		},
		GasBudget: "10000000",
	})
	require.NoError(handler, err)

	handler.signAndExecuteTxReturnResult(ctx, txMeta, handler.OwnerKeys.SuiSK)

	// unpause safe contract
	txMeta, err = handler.SuiProxy.MoveCall(ctx, models.MoveCallRequest{
		Signer:          string(handler.OwnerKeys.SuiAddress),
		PackageObjectId: handler.PackageID,
		Module:          "safe",
		Function:        "pause_contract",
		TypeArguments:   []interface{}{},
		Arguments: []interface{}{
			handler.SafeObjectID,
			handler.AdminCap,
		},
		GasBudget: "10000000",
	})
	require.NoError(handler, err)

	handler.signAndExecuteTxReturnResult(ctx, txMeta, handler.OwnerKeys.SuiSK)
}

// IssueAndWhitelistToken will issue and whitelist the token on Sui
func (handler *SuiHandler) IssueAndWhitelistToken(ctx context.Context, params IssueTokenParams) {
	coinPackageId, treasuryId, metadataId := handler.deployCoinContract(ctx)
	suiTokenInfo := SuiTokenInfo{
		CoinPackageId:  coinPackageId,
		TreasuryId:     treasuryId,
		CoinMetadataId: metadataId,
	}
	coinAddr := strings.TrimPrefix(coinPackageId, "0x")
	coinAddrBytes, _ := hex.DecodeString(coinAddr)
	handler.TokensRegistry.RegisterPeerChainAddressAndInfo(params.AbstractTokenIdentifier, coinAddrBytes, suiTokenInfo)
	handler.updateMetadata(ctx, params)

	// mint token
	mintAmount, ok := big.NewInt(0).SetString(params.ValueToMintOnPeerChain, 10)
	require.True(handler, ok)
	handler.mint(ctx, params, string(handler.TestKeys.SuiAddress), mintAmount)

	// whitelist token
	txMeta, err := handler.SuiProxy.MoveCall(ctx, models.MoveCallRequest{
		Signer:          string(handler.OwnerKeys.SuiAddress),
		PackageObjectId: handler.PackageID,
		Module:          "safe",
		Function:        "whitelist_token",
		TypeArguments: []interface{}{
			fmt.Sprintf("%s::test_coin::TEST_COIN", coinPackageId),
		},
		Arguments: []interface{}{
			handler.SafeObjectID,
			handler.AdminCap,
			"25",
			"500000",
			params.IsNativeOnPeerChain,
		},
		GasBudget: "100000000",
	})
	require.Nil(handler, err)

	handler.signAndExecuteTxReturnResult(ctx, txMeta, handler.OwnerKeys.SuiSK)

	//TODO: Initial supply value?
}

func (handler *SuiHandler) deployCoinContract(ctx context.Context) (string, string, string) {
	mv := handler.readModuleBytes(suiTestCoinBytecode)

	txMeta, err := handler.SuiProxy.Publish(ctx, models.PublishRequest{
		Sender:          string(handler.OwnerKeys.SuiAddress),
		CompiledModules: []string{base64.StdEncoding.EncodeToString(mv)},
		Dependencies: []string{
			"0x1", // Sui Framework
			"0x2", // Move Standard Library
		},
		GasBudget: "100000000",
	})
	require.NoError(handler, err)

	resp := handler.signAndExecuteTxReturnResult(ctx, txMeta, handler.OwnerKeys.SuiSK)

	var coinPackageId, treasuryId, metadataId string
	for _, obj := range resp.ObjectChanges {
		if obj.Type == "created" {
			if strings.Contains(obj.ObjectType, "0x2::coin::TreasuryCap") {
				treasuryId = obj.ObjectId
			}
			if strings.Contains(obj.ObjectType, "0x2::coin::CoinMetadata") {
				metadataId = obj.ObjectId
			}
		} else if obj.Type == "published" {
			coinPackageId = obj.PackageId
		}
	}

	return coinPackageId, treasuryId, metadataId

}

func (handler *SuiHandler) updateMetadata(ctx context.Context, params IssueTokenParams) {
	tokenData := handler.TokensRegistry.GetTokenData(params.AbstractTokenIdentifier)
	suiTokenInfo := tokenData.PeerChainTokenInfo.(SuiTokenInfo)

	// update name
	txMeta, err := handler.SuiProxy.MoveCall(ctx, models.MoveCallRequest{
		Signer:          string(handler.OwnerKeys.SuiAddress),
		PackageObjectId: "0x2",
		Module:          "coin",
		Function:        "update_name",
		TypeArguments: []interface{}{
			fmt.Sprintf("%s::test_coin::TEST_COIN", suiTokenInfo.CoinPackageId),
		},
		Arguments: []interface{}{
			suiTokenInfo.TreasuryId,
			suiTokenInfo.CoinMetadataId,
			params.PeerChainTokenName,
		},
		GasBudget: "100000000",
	})
	require.Nil(handler, err)

	handler.signAndExecuteTxReturnResult(ctx, txMeta, handler.OwnerKeys.SuiSK)

	// update symbol
	txMeta, err = handler.SuiProxy.MoveCall(ctx, models.MoveCallRequest{
		Signer:          string(handler.OwnerKeys.SuiAddress),
		PackageObjectId: "0x2",
		Module:          "coin",
		Function:        "update_symbol",
		TypeArguments: []interface{}{
			fmt.Sprintf("%s::test_coin::TEST_COIN", suiTokenInfo.CoinPackageId),
		},
		Arguments: []interface{}{
			suiTokenInfo.TreasuryId,
			suiTokenInfo.CoinMetadataId,
			params.PeerChainTokenSymbol,
		},
		GasBudget: "100000000",
	})
	require.Nil(handler, err)

	handler.signAndExecuteTxReturnResult(ctx, txMeta, handler.OwnerKeys.SuiSK)
}

// CreateBatchOnPeerChain will create a batch on Sui using the provided tokens parameters list
func (handler *SuiHandler) CreateBatchOnPeerChain(
	ctx context.Context,
	_ core.AddressHandler,
	tokensParams ...TestTokenParams,
) {
	for _, params := range tokensParams {
		handler.createDepositsOnSuiForToken(ctx, params, handler.TestKeys.SuiSK, handler.TestKeys.SuiAddress)
	}

	// TODO: wait until batch is settled
}

func (handler *SuiHandler) createDepositsOnSuiForToken(
	ctx context.Context,
	params TestTokenParams,
	fromPriKey ed25519.PrivateKey,
	fromAddress []byte,
) {
	token := handler.TokensRegistry.GetTokenData(params.AbstractTokenIdentifier)
	require.NotNil(handler, token)
	require.NotNil(handler, token.PeerChainTokenAddress)

	for _, operation := range params.TestOperations {
		if operation.ValueToTransferToMvx == nil {
			continue
		}

		coinObjId := handler.getCoinObjectIdForToken(ctx, token.PeerChainTokenAddress, operation.ValueToTransferToMvx)
		coinType := fmt.Sprintf("0x%s::test_coin::TEST_COIN", hex.EncodeToString(token.PeerChainTokenAddress))

		// No sc call data only
		txMeta, err := handler.SuiProxy.MoveCall(ctx, models.MoveCallRequest{
			Signer:          string(fromAddress),
			PackageObjectId: handler.PackageID,
			Module:          "safe",
			Function:        "deposit",
			TypeArguments: []interface{}{
				coinType,
			},
			Arguments: []interface{}{
				handler.SafeObjectID,
				coinObjId,
				handler.TestKeys.MvxAddress.AddressSlice(),
				"0x6",
			},
			GasBudget: "10000000",
		})
		require.NoError(handler, err)

		handler.signAndExecuteTxReturnResult(ctx, txMeta, fromPriKey)
	}
}

func (handler *SuiHandler) getCoinObjectIdForToken(ctx context.Context, coinAddress []byte, targetValue *big.Int) string {
	coins, err := handler.SuiProxy.SuiXGetCoins(ctx, models.SuiXGetCoinsRequest{
		Owner:    string(handler.TestKeys.SuiAddress),
		CoinType: fmt.Sprintf("0x%s::test_coin::TEST_COIN", hex.EncodeToString(coinAddress)),
	})
	require.NoError(handler, err)

	utils.PrettyPrint(coins)

	srcCoin := coins.Data[0]
	coinBalance, _ := big.NewInt(0).SetString(srcCoin.Balance, 10)
	var coinToSendId string
	if coinBalance.Cmp(targetValue) > 0 {
		txMeta, err := handler.SuiProxy.SplitCoin(ctx, models.SplitCoinRequest{
			Signer:       string(handler.TestKeys.SuiAddress),
			CoinObjectId: srcCoin.CoinObjectId,
			SplitAmounts: []string{targetValue.String()},
			GasBudget:    "10000000",
		})
		require.NoError(handler, err)

		resp := handler.signAndExecuteTxReturnResult(ctx, txMeta, handler.TestKeys.SuiSK)
		for _, obj := range resp.ObjectChanges {
			if obj.Type == "created" && strings.Contains(obj.ObjectType, "test_coin::TEST_COIN") {
				coinToSendId = obj.ObjectId
				break
			}
		}
	}
	// TODO: maybe merge

	return coinToSendId
}

// SendFromPeerChainToMultiversX will create the deposit transactions on the Sui side
func (handler *SuiHandler) SendFromPeerChainToMultiversX(
	ctx context.Context,
	_ core.AddressHandler,
	tokensParams ...TestTokenParams,
) {
	for _, params := range tokensParams {
		handler.createDepositsOnSuiForToken(ctx, params, handler.TestKeys.SuiSK, handler.TestKeys.SuiAddress)
	}
}

// Mint will mint the provided token on Sui with the provided value on the behalf of the Depositor address
func (handler *SuiHandler) Mint(ctx context.Context, params TestTokenParams, valueToMint *big.Int) {
	handler.mint(ctx, params.IssueTokenParams, handler.SafeObjectID, valueToMint)
}

func (handler *SuiHandler) mint(ctx context.Context, params IssueTokenParams, receiver string, valueToMint *big.Int) {
	tokenData := handler.TokensRegistry.GetTokenData(params.AbstractTokenIdentifier)
	require.NotNil(handler, tokenData)
	require.NotNil(handler, tokenData.PeerChainTokenInfo)
	suiTokenInfo := tokenData.PeerChainTokenInfo.(SuiTokenInfo)

	txMeta, err := handler.SuiProxy.MoveCall(ctx, models.MoveCallRequest{
		Signer:          string(handler.OwnerKeys.SuiAddress),
		PackageObjectId: suiTokenInfo.CoinPackageId,
		Module:          "test_coin",
		Function:        "mint",
		TypeArguments:   []interface{}{},
		Arguments: []interface{}{
			suiTokenInfo.TreasuryId,
			valueToMint.String(),
			receiver,
		},
		GasBudget: "100000000",
	})
	require.Nil(handler, err)

	handler.signAndExecuteTxReturnResult(ctx, txMeta, handler.OwnerKeys.SuiSK)
}

func (handler *SuiHandler) signAndExecuteTxReturnResult(
	ctx context.Context,
	txMeta models.TxnMetaData,
	signerPriKey ed25519.PrivateKey,
) models.SuiTransactionBlockResponse {
	exec, err := handler.SuiProxy.SignAndExecuteTransactionBlock(ctx, models.SignAndExecuteTransactionBlockRequest{
		TxnMetaData: txMeta,
		PriKey:      signerPriKey,
		Options: models.SuiTransactionBlockOptions{
			ShowEffects:       true,
			ShowObjectChanges: true,
		},
		RequestType: "WaitForLocalExecution",
	})

	require.Nil(handler, err)
	require.Equal(handler, "success", exec.Effects.Status.Status, fmt.Sprintf("Error: %s", exec.Effects.Status.Error))

	return exec
}

func (handler *SuiHandler) Close() error {
	panic("Close not implemented for SuiHandler")
}

func (handler *SuiHandler) FundWallets(wallets [][]byte) {
	for _, wallet := range wallets {
		//faucetHost, err := suiSdk.GetFaucetHost(constant.SuiLocalnet)
		//if err != nil {
		//	fmt.Println("GetFaucetHost err:", err)
		//	return
		//}

		header := map[string]string{}
		err := suiSdk.RequestSuiFromFaucet("http://127.0.0.1:9123", string(wallet), header)
		if err != nil {
			log.Error("error in suiChainSimulatorWrapper.FundWallets", "error", err.Error())
			continue
		}
		log.Info("Funded wallet: " + string(wallet))
	}
}

func (handler *SuiHandler) GenerateBlocks(ctx context.Context, numBlocks int) {
	for i := 0; i < numBlocks; i++ {
		_, err := handler.SuiProxy.SignAndExecuteTransactionBlock(
			ctx,
			models.SignAndExecuteTransactionBlockRequest{
				TxnMetaData: models.TxnMetaData{},
				PriKey:      handler.OwnerKeys.SuiSK,
				Options:     models.SuiTransactionBlockOptions{},
				RequestType: "WaitForLocalExecution",
			},
		)
		if err != nil {
			log.Error("Failed to generate block", "error", err)
			continue
		}
	}
}
