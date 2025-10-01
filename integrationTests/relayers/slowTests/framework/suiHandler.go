package framework

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"fmt"
	"math/big"
	"os"
	"strings"
	"testing"

	"github.com/block-vision/sui-go-sdk/models"
	"github.com/multiversx/mx-sdk-go/core"
	"github.com/stretchr/testify/require"
)

const (
	suiSafeBytecode             = "testdata/contracts/sui/safe.mv"
	suiBridgeBytecode           = "testdata/contracts/sui/bridge.mv"
	suiEventsBytecode           = "testdata/contracts/sui/events.mv"
	suiPausableBytecode         = "testdata/contracts/sui/pausable.mv"
	suiRolesBytecode            = "testdata/contracts/sui/bridge_roles.mv"
	suiSharedStructsBytecode    = "testdata/contracts/sui/shared_structs.mv"
	suiUtilsBytecode            = "testdata/contracts/sui/utils.mv"
	suiTestCoinBytecode         = "testdata/contracts/sui/test_coin.mv"
	suiBridgeTokenBytecode      = "testdata/contracts/sui/bridge_token.mv"
	suiTokenRolesBytecode       = "testdata/contracts/sui/lk_roles.mv"
	suiTokenTreasuryBytecode    = "testdata/contracts/sui/treasury.mv"
	suiExtensionServiceBytecode = "testdata/contracts/sui/two_step_role.mv"
	suiUpgradeBytecode          = "testdata/contracts/sui/upgrade_service.mv"

	suiFrameworkId = "0x1"
	moveStdLibId   = "0x2"
	clockId        = "0x6"

	suiPubKeyLength = 32
)

// SuiHandler will handle all the operations on the Sui side
type SuiHandler struct {
	testing.TB
	*KeysStore
	TokensRegistry             TokensRegistry
	Quorum                     string
	MvxTestCallerAddress       core.AddressHandler
	SuiChainSimulator          *suiChainSimulatorWrapper
	PackageID                  string
	BridgeObjectID             string
	BridgeCap                  string
	AdminCap                   string
	SafeObjectID               string
	BridgeInitialSharedVersion uint64
	SafeInitialSharedVersion   uint64
	TokenPackageID             string
	TokenType                  string
	TokenTreasuryCapId         string
	TokenPolicyCapId           string
	TokenManagerId             string
	FromCoinCap                string
}

// NewSuiHandler will create the handler that will adapt all test operations on Sui
func NewSuiHandler(
	tb testing.TB,
	keysStore *KeysStore,
	tokensRegistry TokensRegistry,
	chainSimulator *suiChainSimulatorWrapper,
	quorum string,
) *SuiHandler {
	handler := &SuiHandler{
		TB:                tb,
		KeysStore:         keysStore,
		TokensRegistry:    tokensRegistry,
		Quorum:            quorum,
		SuiChainSimulator: chainSimulator,
	}

	walletsToFundOnSui := handler.WalletsToFundOnSui()
	handler.SuiChainSimulator.FundWallets(walletsToFundOnSui)

	return handler
}

func (handler *SuiHandler) DeployContracts(ctx context.Context) {
	resp := handler.SuiChainSimulator.PublishPackage(ctx, models.PublishRequest{
		Sender:          string(handler.OwnerKeys.SuiAddress),
		CompiledModules: handler.getBridgeTokenEncodedModules(),
		Dependencies: []string{
			suiFrameworkId,
			moveStdLibId,
		},
		GasBudget: "500000000",
	}, handler.OwnerKeys)

	for _, obj := range resp.ObjectChanges {
		if obj.Type == "created" {
			if strings.Contains(obj.ObjectType, "0x2::coin::TreasuryCap") {
				handler.TokenType = extractInnerType(obj.ObjectType)
			}
			if strings.Contains(obj.ObjectType, "::treasury::Treasury<") {
				handler.TokenManagerId = obj.ObjectId
			}
		} else if obj.Type == "published" {
			handler.TokenPackageID = obj.PackageId
		}
	}

	handler.transferFromCoinCapToOwner(ctx)

	// publish bridge package
	resp = handler.SuiChainSimulator.PublishPackage(ctx, models.PublishRequest{
		Sender:          string(handler.OwnerKeys.SuiAddress),
		CompiledModules: handler.getBridgeEncodedModules(),
		Dependencies: []string{
			suiFrameworkId,
			moveStdLibId,
			handler.TokenPackageID, // TODO
		},
		GasBudget: "500000000",
	}, handler.OwnerKeys)

	for _, obj := range resp.ObjectChanges {
		if obj.Type == "published" {
			handler.PackageID = obj.PackageId
		}
	}

	handler.initSafe(ctx)

	suiRelayersAddresses := make([]string, 0, len(handler.RelayersKeys))
	suiRelayersPubKeys := make([][suiPubKeyLength]byte, 0, len(handler.RelayersKeys))
	for _, relayerKeys := range handler.RelayersKeys {
		suiRelayersAddresses = append(suiRelayersAddresses, string(relayerKeys.SuiAddress))

		pubKeyBytes := relayerKeys.SuiSK.Public().(ed25519.PublicKey)
		var pk [suiPubKeyLength]byte
		copy(pk[:], pubKeyBytes)
		suiRelayersPubKeys = append(suiRelayersPubKeys, pk)
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

func (handler *SuiHandler) getBridgeEncodedModules() []string {
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

func (handler *SuiHandler) getBridgeTokenEncodedModules() []string {
	var modules []string

	mv := handler.readModuleBytes(suiBridgeTokenBytecode)
	modules = append(modules, base64.StdEncoding.EncodeToString(mv))

	mv = handler.readModuleBytes(suiTokenRolesBytecode)
	modules = append(modules, base64.StdEncoding.EncodeToString(mv))

	mv = handler.readModuleBytes(suiTokenTreasuryBytecode)
	modules = append(modules, base64.StdEncoding.EncodeToString(mv))

	mv = handler.readModuleBytes(suiExtensionServiceBytecode)
	modules = append(modules, base64.StdEncoding.EncodeToString(mv))

	mv = handler.readModuleBytes(suiUpgradeBytecode)
	modules = append(modules, base64.StdEncoding.EncodeToString(mv))

	return modules
}

//func (handler *SuiHandler) foo() []string {
//	var modules []string
//
//	mv := handler.readModuleBytes(suiExtensionServiceBytecode)
//	modules = append(modules, base64.StdEncoding.EncodeToString(mv))
//
//	mv = handler.readModuleBytes(suiUpgradeBytecode)
//	modules = append(modules, base64.StdEncoding.EncodeToString(mv))
//
//	return modules
//}

func (handler *SuiHandler) readModuleBytes(path string) []byte {
	b, err := os.ReadFile(path)
	require.NoError(handler, err)
	return b
}

func extractInnerType(s string) string {
	start := strings.Index(s, "<")
	end := strings.LastIndex(s, ">")
	if start == -1 || end == -1 || start >= end {
		return ""
	}
	return s[start+1 : end]
}

func (handler *SuiHandler) transferFromCoinCapToOwner(ctx context.Context) {
	resp := handler.SuiChainSimulator.MoveCall(ctx, models.MoveCallRequest{
		Signer:          string(handler.OwnerKeys.SuiAddress),
		PackageObjectId: handler.TokenPackageID,
		Module:          "treasury",
		Function:        "transfer_from_coin_cap",
		TypeArguments: []interface{}{
			handler.TokenType,
		},
		Arguments: []interface{}{
			handler.TokenManagerId,
			handler.OwnerKeys.SuiAddress,
		},
		GasBudget: "10000000",
	}, handler.OwnerKeys)

	for _, obj := range resp.ObjectChanges {
		if obj.Type == "created" {
			if strings.Contains(obj.ObjectType, "::treasury::FromCoinCap") {
				handler.FromCoinCap = obj.ObjectId
			}
		}
	}
}

func (handler *SuiHandler) initSafe(ctx context.Context) {
	resp := handler.SuiChainSimulator.MoveCall(ctx, models.MoveCallRequest{
		Signer:          string(handler.OwnerKeys.SuiAddress),
		PackageObjectId: handler.PackageID,
		Module:          "safe",
		Function:        "initialize",
		TypeArguments:   []interface{}{},
		Arguments: []interface{}{
			handler.FromCoinCap,
		},
		GasBudget: "10000000",
	}, handler.OwnerKeys)

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
				if ownerMap, ok := obj.Owner.(map[string]interface{}); ok {
					if shared, exists := ownerMap["Shared"]; exists {
						if sharedMap, ok := shared.(map[string]interface{}); ok {
							if version, exists := sharedMap["initial_shared_version"]; exists {
								if versionFloat, ok := version.(float64); ok {
									handler.SafeInitialSharedVersion = uint64(versionFloat)
								}
							}
						}
					}
				}
			}
		}
	}
}

func (handler *SuiHandler) DeployContract(
	ctx context.Context,
	params ...interface{},
) []byte {
	module := params[0].(string)
	function := params[1].(string)
	suiRelayerAddresses := params[2].([]string)
	suiRelayersPubKeys := params[3].([][suiPubKeyLength]byte)
	quorumStr := params[4].(string)
	safeObjectID := params[5].(string)
	adminCap := params[6].(string)

	resp := handler.SuiChainSimulator.MoveCall(ctx, models.MoveCallRequest{
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
	}, handler.OwnerKeys)

	for _, obj := range resp.ObjectChanges {
		if obj.Type == "created" {
			if strings.Contains(obj.ObjectType, "::bridge::Bridge") {
				if ownerMap, ok := obj.Owner.(map[string]interface{}); ok {
					if shared, exists := ownerMap["Shared"]; exists {
						if sharedMap, ok := shared.(map[string]interface{}); ok {
							if version, exists := sharedMap["initial_shared_version"]; exists {
								if versionFloat, ok := version.(float64); ok {
									handler.BridgeInitialSharedVersion = uint64(versionFloat)
								}
							}
						}
					}
				}

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
	suiTokenInfo := token.PeerChainTokenInfo.(SuiTokenInfo)
	require.NotNil(handler, suiTokenInfo)

	if suiTokenInfo.IsLocked {
		return handler.SuiChainSimulator.GetTokenBalance(ctx, string(receiver), handler.TokenType)
	} else {
		return handler.SuiChainSimulator.GetCoinBalance(ctx, string(receiver), string(token.PeerChainTokenAddress))
	}
}

// UnPauseContractsAfterTokenChanges can unpause contracts after token changes
func (handler *SuiHandler) UnPauseContractsAfterTokenChanges(ctx context.Context) {
	// unpause bridge contract
	handler.SuiChainSimulator.MoveCall(ctx, models.MoveCallRequest{
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
	}, handler.OwnerKeys)

	// unpause safe contract
	handler.SuiChainSimulator.MoveCall(ctx, models.MoveCallRequest{
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
	}, handler.OwnerKeys)
}

// PauseContractsForTokenChanges can pause contracts for token changes
func (handler *SuiHandler) PauseContractsForTokenChanges(ctx context.Context) {
	// unpause bridge contract
	handler.SuiChainSimulator.MoveCall(ctx, models.MoveCallRequest{
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
	}, handler.OwnerKeys)

	// unpause safe contract
	handler.SuiChainSimulator.MoveCall(ctx, models.MoveCallRequest{
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
	}, handler.OwnerKeys)
}

// IssueAndWhitelistToken will issue and whitelist the token on Sui
func (handler *SuiHandler) IssueAndWhitelistToken(ctx context.Context, params IssueTokenParams) {
	var coinType string
	var suiTokenInfo SuiTokenInfo
	if !params.IsLocked {
		coinPackageId, treasuryId, metadataId := handler.deployCoinContract(ctx)
		suiTokenInfo = SuiTokenInfo{
			CoinPackageId:  coinPackageId,
			TreasuryId:     treasuryId,
			CoinMetadataId: metadataId,
			IsLocked:       params.IsLocked,
		}
		coinType = fmt.Sprintf("%s::test_coin::TEST_COIN", coinPackageId)
	} else {
		suiTokenInfo = SuiTokenInfo{
			CoinPackageId:  handler.TokenPackageID,
			TreasuryId:     "",
			CoinMetadataId: "",
			IsLocked:       params.IsLocked,
		}
		coinType = handler.TokenType
	}

	handler.TokensRegistry.RegisterPeerChainAddressAndInfo(params.AbstractTokenIdentifier, []byte(coinType), suiTokenInfo)

	if !params.IsLocked {
		handler.updateMetadata(ctx, params)
		if params.IsLocked {
			handler.setTreasuryCapOnSafe(ctx)
			handler.setPolicyCapOnSafe(ctx)
		}
	}

	// mint token
	mintAmount, ok := big.NewInt(0).SetString(params.ValueToMintOnPeerChain, 10)
	require.True(handler, ok)
	if params.IsLocked {
		handler.mintLocked(ctx, params, string(handler.TestKeys.SuiAddress), mintAmount)
	} else {
		handler.mint(ctx, params, string(handler.TestKeys.SuiAddress), mintAmount)
	}

	// whitelist token
	handler.SuiChainSimulator.MoveCall(ctx, models.MoveCallRequest{
		Signer:          string(handler.OwnerKeys.SuiAddress),
		PackageObjectId: handler.PackageID,
		Module:          "safe",
		Function:        "whitelist_token",
		TypeArguments: []interface{}{
			coinType,
		},
		Arguments: []interface{}{
			handler.SafeObjectID,
			handler.AdminCap,
			"25",
			"500000",
			params.IsNativeOnPeerChain,
			params.IsLocked,
		},
		GasBudget: "100000000",
	}, handler.OwnerKeys)

	if len(params.InitialSupplyValue) > 0 {
		initialSupplyValue, ok := big.NewInt(0).SetString(params.InitialSupplyValue, 10)
		require.True(handler, ok)

		if !params.IsLocked {
			handler.mint(ctx, params, string(handler.OwnerKeys.SuiAddress), initialSupplyValue)
		} else {
			handler.mintLocked(ctx, params, string(handler.OwnerKeys.SuiAddress), initialSupplyValue)
		}
		handler.initSupplyForToken(ctx, params)
	}
}

func (handler *SuiHandler) deployCoinContract(ctx context.Context) (string, string, string) {
	mv := handler.readModuleBytes(suiTestCoinBytecode)

	resp := handler.SuiChainSimulator.PublishPackage(ctx, models.PublishRequest{
		Sender:          string(handler.OwnerKeys.SuiAddress),
		CompiledModules: []string{base64.StdEncoding.EncodeToString(mv)},
		Dependencies: []string{
			suiFrameworkId,
			moveStdLibId,
		},
		GasBudget: "100000000",
	}, handler.OwnerKeys)

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
	handler.SuiChainSimulator.MoveCall(ctx, models.MoveCallRequest{
		Signer:          string(handler.OwnerKeys.SuiAddress),
		PackageObjectId: "0x2",
		Module:          "coin",
		Function:        "update_name",
		TypeArguments: []interface{}{
			string(tokenData.PeerChainTokenAddress),
		},
		Arguments: []interface{}{
			suiTokenInfo.TreasuryId,
			suiTokenInfo.CoinMetadataId,
			params.PeerChainTokenName,
		},
		GasBudget: "100000000",
	}, handler.OwnerKeys)

	// update symbol
	handler.SuiChainSimulator.MoveCall(ctx, models.MoveCallRequest{
		Signer:          string(handler.OwnerKeys.SuiAddress),
		PackageObjectId: "0x2",
		Module:          "coin",
		Function:        "update_symbol",
		TypeArguments: []interface{}{
			string(tokenData.PeerChainTokenAddress),
		},
		Arguments: []interface{}{
			suiTokenInfo.TreasuryId,
			suiTokenInfo.CoinMetadataId,
			params.PeerChainTokenSymbol,
		},
		GasBudget: "100000000",
	}, handler.OwnerKeys)
}

func (handler *SuiHandler) mintLocked(ctx context.Context, params IssueTokenParams, receiver string, amount *big.Int) {
	tokenData := handler.TokensRegistry.GetTokenData(params.AbstractTokenIdentifier)
	suiTokenInfo := tokenData.PeerChainTokenInfo.(SuiTokenInfo)
	require.NotNil(handler, suiTokenInfo)

	handler.SuiChainSimulator.MoveCall(ctx, models.MoveCallRequest{
		Signer:          string(handler.OwnerKeys.SuiAddress),
		PackageObjectId: handler.TokenPackageID,
		Module:          "treasury",
		Function:        "mint_coin_to_receiver",
		TypeArguments: []interface{}{
			handler.TokenType,
		},
		Arguments: []interface{}{
			handler.TokenManagerId,
			amount.String(),
			receiver,
		},
		GasBudget: "100000000",
	}, handler.OwnerKeys)
}

func (handler *SuiHandler) initSupplyForToken(ctx context.Context, params IssueTokenParams) {
	tokenData := handler.TokensRegistry.GetTokenData(params.AbstractTokenIdentifier)
	initSupplyValue, ok := big.NewInt(0).SetString(params.InitialSupplyValue, 10)
	require.True(handler, ok)

	coinObjId := handler.getCoinObjectIdForToken(ctx, tokenData.PeerChainTokenAddress, initSupplyValue, handler.OwnerKeys)

	handler.SuiChainSimulator.MoveCall(ctx, models.MoveCallRequest{
		Signer:          string(handler.OwnerKeys.SuiAddress),
		PackageObjectId: handler.PackageID,
		Module:          "safe",
		Function:        "init_supply",
		TypeArguments: []interface{}{
			string(tokenData.PeerChainTokenAddress),
		},
		Arguments: []interface{}{
			handler.AdminCap,
			handler.SafeObjectID,
			coinObjId,
		},
		GasBudget: "100000000",
	}, handler.OwnerKeys)
}

func (handler *SuiHandler) setTreasuryCapOnSafe(ctx context.Context) {
	handler.SuiChainSimulator.MoveCall(ctx, models.MoveCallRequest{
		Signer:          string(handler.OwnerKeys.SuiAddress),
		PackageObjectId: handler.PackageID,
		Module:          "safe",
		Function:        "set_treasury_cap",
		TypeArguments:   []interface{}{},
		Arguments: []interface{}{
			handler.AdminCap,
			handler.SafeObjectID,
			handler.TokenTreasuryCapId,
		},
		GasBudget: "100000000",
	}, handler.OwnerKeys)
}

func (handler *SuiHandler) setPolicyCapOnSafe(ctx context.Context) {
	handler.SuiChainSimulator.MoveCall(ctx, models.MoveCallRequest{
		Signer:          string(handler.OwnerKeys.SuiAddress),
		PackageObjectId: handler.PackageID,
		Module:          "safe",
		Function:        "set_policy_cap",
		TypeArguments:   []interface{}{},
		Arguments: []interface{}{
			handler.AdminCap,
			handler.SafeObjectID,
			handler.TokenPolicyCapId,
		},
		GasBudget: "100000000",
	}, handler.OwnerKeys)
}

// CreateBatchOnPeerChain will create a batch on Sui using the provided tokens parameters list
func (handler *SuiHandler) CreateBatchOnPeerChain(
	ctx context.Context,
	_ core.AddressHandler,
	tokensParams ...TestTokenParams,
) {
	for _, params := range tokensParams {
		handler.createDepositsOnSuiForToken(ctx, params, handler.TestKeys)
	}

	// Wait until the batch is processed
	handler.SuiChainSimulator.GenerateBlocks(ctx, 50)
}

func (handler *SuiHandler) createDepositsOnSuiForToken(
	ctx context.Context,
	params TestTokenParams,
	from KeysHolder,
) {
	token := handler.TokensRegistry.GetTokenData(params.AbstractTokenIdentifier)
	require.NotNil(handler, token)
	require.NotNil(handler, token.PeerChainTokenAddress)

	for _, operation := range params.TestOperations {
		if operation.ValueToTransferToMvx == nil {
			continue
		}

		coinObjId := handler.getCoinObjectIdForToken(ctx, token.PeerChainTokenAddress, operation.ValueToTransferToMvx, handler.TestKeys)
		coinType := string(token.PeerChainTokenAddress)

		// No sc call data only
		handler.SuiChainSimulator.MoveCall(ctx, models.MoveCallRequest{
			Signer:          string(from.SuiAddress),
			PackageObjectId: handler.PackageID,
			Module:          "safe",

			Function: "deposit",
			TypeArguments: []interface{}{
				coinType,
			},
			Arguments: []interface{}{
				handler.SafeObjectID,
				coinObjId,
				handler.TestKeys.MvxAddress.AddressSlice(),
				clockId,
			},
			GasBudget: "10000000",
		}, from)
	}
}

func (handler *SuiHandler) getCoinObjectIdForToken(ctx context.Context, coinAddress []byte, targetValue *big.Int, signer KeysHolder) string {
	coins := handler.SuiChainSimulator.GetCoins(ctx, string(signer.SuiAddress), string(coinAddress))
	srcCoin := coins[0]
	coinBalance, _ := big.NewInt(0).SetString(srcCoin.Balance, 10)

	var coinToSendId string
	if coinBalance.Cmp(targetValue) == 0 {
		coinToSendId = srcCoin.CoinObjectId
	} else {
		if coinBalance.Cmp(targetValue) > 0 {
			resp := handler.SuiChainSimulator.SplitCoin(ctx, models.SplitCoinRequest{
				Signer:       string(signer.SuiAddress),
				CoinObjectId: srcCoin.CoinObjectId,
				SplitAmounts: []string{targetValue.String()},
				GasBudget:    "10000000",
			}, signer)

			for _, obj := range resp.ObjectChanges {
				if obj.Type == "created" && strings.Contains(obj.ObjectType, "test_coin::TEST_COIN") {
					coinToSendId = obj.ObjectId
					break
				}
			}
		}
	}

	return coinToSendId
}

// SendFromPeerChainToMultiversX will create the deposit transactions on the Sui side
func (handler *SuiHandler) SendFromPeerChainToMultiversX(
	ctx context.Context,
	_ core.AddressHandler,
	tokensParams ...TestTokenParams,
) {
	for _, params := range tokensParams {
		handler.createDepositsOnSuiForToken(ctx, params, handler.TestKeys)
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

	handler.SuiChainSimulator.MoveCall(ctx, models.MoveCallRequest{
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
	}, handler.OwnerKeys)
}

func (handler *SuiHandler) Close() error {
	return nil
}
