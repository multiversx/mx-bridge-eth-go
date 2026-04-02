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
	"github.com/multiversx/mx-bridge-eth-go/config"
	"github.com/multiversx/mx-sdk-go/core"
	"github.com/stretchr/testify/require"
)

const (
	// sui_extensions modules
	suiExtensionServiceBytecode = "testdata/contracts/sui/two_step_role.mv"
	suiUpgradeBytecode          = "testdata/contracts/sui/upgrade_service.mv"

	// locked_token modules
	suiBridgeTokenBytecode         = "testdata/contracts/sui/token/bridge_token.mv"
	suiTokenRolesBytecode          = "testdata/contracts/sui/token/lk_roles.mv"
	suiTokenTreasuryBytecode       = "testdata/contracts/sui/token/treasury.mv"
	suiTokenVersionControlBytecode = "testdata/contracts/sui/token/token_version_control.mv"
	suiUpgradeServiceTokenBytecode = "testdata/contracts/sui/token/upgrade_service_token.mv"

	// stablecoin (treasury) modules
	suiStablecoinVersionControlBytecode = "testdata/contracts/sui/stablecoin/version_control.mv"
	suiStablecoinMintAllowanceBytecode  = "testdata/contracts/sui/stablecoin/mint_allowance.mv"
	suiStablecoinRolesBytecode          = "testdata/contracts/sui/stablecoin/roles.mv"
	suiStablecoinTreasuryBytecode       = "testdata/contracts/sui/stablecoin/treasury.mv"
	suiStablecoinEntryBytecode          = "testdata/contracts/sui/stablecoin/entry.mv"
	suiXmnBytecode                      = "testdata/contracts/sui/stablecoin/xmn.mv"

	// bridge modules
	suiSafeBytecode                 = "testdata/contracts/sui/bridge/safe.mv"
	suiBridgeBytecode               = "testdata/contracts/sui/bridge/bridge.mv"
	suiEventsBytecode               = "testdata/contracts/sui/bridge/events.mv"
	suiPausableBytecode             = "testdata/contracts/sui/bridge/pausable.mv"
	suiRolesBytecode                = "testdata/contracts/sui/bridge/bridge_roles.mv"
	suiSharedStructsBytecode        = "testdata/contracts/sui/bridge/shared_structs.mv"
	suiUtilsBytecode                = "testdata/contracts/sui/bridge/utils.mv"
	suiBridgeVersionControlBytecode = "testdata/contracts/sui/bridge/bridge_version_control.mv"
	suiUpgradeManagerBytecode       = "testdata/contracts/sui/bridge/upgrade_manager.mv"
	suiUpgradeServiceBridgeBytecode = "testdata/contracts/sui/bridge/upgrade_service_bridge.mv"
	suiXmnMintCapAdapterBytecode    = "testdata/contracts/sui/bridge/xmn_mint_cap_adapter.mv"

	// test coin
	suiTestCoinBytecode = "testdata/contracts/sui/coin/test_coin.mv"

	suiFrameworkId               = "0x1"
	moveStdLibId                 = "0x2"
	clockId                      = "0x6"
	denyListObjectId             = "0x0000000000000000000000000000000000000000000000000000000000000403"
	denyListInitialSharedVersion = uint64(1)

	suiPubKeyLength = 32
)

// SuiHandler will handle all the operations on the Sui side
type SuiHandler struct {
	testing.TB
	*KeysStore
	TokensRegistry                  TokensRegistry
	Quorum                          string
	MvxTestCallerAddress            core.AddressHandler
	SuiChainSimulator               *suiChainSimulatorWrapper
	PackageID                       string
	BridgeObjectID                  string
	BridgeInitialSharedVersion      uint64
	SafeObjectID                    string
	SafeInitialSharedVersion        uint64
	TreasuryId                      string
	TreasuryInitialSharedVersion    uint64
	XmnTreasuryId                   string
	XmnTreasuryInitialSharedVersion uint64
	MintBurnAdapterInfos            map[string]SuiMintBurnAdapterInfo
	BridgeCap                       string
	TokenType                       string
	FromCoinCap                     string
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
		TB:                   tb,
		KeysStore:            keysStore,
		TokensRegistry:       tokensRegistry,
		Quorum:               quorum,
		SuiChainSimulator:    chainSimulator,
		MintBurnAdapterInfos: make(map[string]SuiMintBurnAdapterInfo),
	}

	walletsToFundOnSui := handler.WalletsToFundOnSui()
	handler.SuiChainSimulator.FundWallets(walletsToFundOnSui)

	return handler
}

func (handler *SuiHandler) DeployContracts(ctx context.Context) {
	resp := handler.SuiChainSimulator.PublishPackage(ctx, models.PublishRequest{
		Sender:          string(handler.OwnerKeys.SuiAddress),
		CompiledModules: handler.getBridgeEncodedModules(),
		Dependencies: []string{
			suiFrameworkId,
			moveStdLibId,
		},
		GasBudget: "500000000",
	}, handler.OwnerKeys)

	for _, obj := range resp.ObjectChanges {
		switch obj.Type {
		case "created":
			if strings.Contains(obj.ObjectType, "0x2::coin::TreasuryCap") {
				handler.TokenType = extractInnerType(obj.ObjectType)
			}
			if strings.Contains(obj.ObjectType, "::treasury::Treasury<") {
				isv := extractInitialSharedVersion(obj.Owner)
				if strings.Contains(obj.ObjectType, "::xmn::XMN") {
					handler.XmnTreasuryId = obj.ObjectId
					handler.XmnTreasuryInitialSharedVersion = isv
				} else {
					handler.TreasuryId = obj.ObjectId
					handler.TreasuryInitialSharedVersion = isv
				}
			}
		case "published":
			handler.PackageID = obj.PackageId
		}
	}

	handler.transferFromCoinCapToOwner(ctx)
	handler.initSafe(ctx)

	suiRelayersPubKeys := make([][suiPubKeyLength]byte, 0, len(handler.RelayersKeys))
	for _, relayerKeys := range handler.RelayersKeys {
		pubKeyBytes := relayerKeys.SuiSK.Public().(ed25519.PublicKey)
		var pk [suiPubKeyLength]byte
		copy(pk[:], pubKeyBytes)
		suiRelayersPubKeys = append(suiRelayersPubKeys, pk)
	}

	bridgeIdBytes := handler.DeployContract(
		ctx,
		"bridge",
		"initialize",
		suiRelayersPubKeys,
		handler.Quorum,
		handler.SafeObjectID,
		handler.BridgeCap,
	)
	handler.BridgeObjectID = string(bridgeIdBytes)
}

func (handler *SuiHandler) getBridgeEncodedModules() []string {
	// Module order follows topological dependency order:
	// sui_extensions → locked_token → stablecoin (treasury + xmn) → bridge modules

	paths := []string{
		// sui_extensions
		suiExtensionServiceBytecode,
		suiUpgradeBytecode,
		// locked_token
		suiBridgeTokenBytecode,
		suiTokenRolesBytecode,
		suiTokenTreasuryBytecode,
		suiTokenVersionControlBytecode,
		suiUpgradeServiceTokenBytecode,
		// stablecoin treasury
		suiStablecoinVersionControlBytecode,
		suiStablecoinMintAllowanceBytecode,
		suiStablecoinRolesBytecode,
		suiStablecoinTreasuryBytecode,
		suiStablecoinEntryBytecode,
		// xmn coin
		suiXmnBytecode,
		// bridge modules
		suiSafeBytecode,
		suiBridgeBytecode,
		suiEventsBytecode,
		suiPausableBytecode,
		suiRolesBytecode,
		suiSharedStructsBytecode,
		suiUtilsBytecode,
		suiBridgeVersionControlBytecode,
		suiUpgradeManagerBytecode,
		suiUpgradeServiceBridgeBytecode,
		suiXmnMintCapAdapterBytecode,
	}

	modules := make([]string, 0, len(paths))
	for _, p := range paths {
		mv := handler.readModuleBytes(p)
		modules = append(modules, base64.StdEncoding.EncodeToString(mv))
	}

	return modules
}

func (handler *SuiHandler) readModuleBytes(path string) []byte {
	b, err := os.ReadFile(path)
	require.NoError(handler, err)
	return b
}

func extractInitialSharedVersion(owner interface{}) uint64 {
	if ownerMap, ok := owner.(map[string]interface{}); ok {
		if shared, exists := ownerMap["Shared"]; exists {
			if sharedMap, ok := shared.(map[string]interface{}); ok {
				if version, exists := sharedMap["initial_shared_version"]; exists {
					if versionFloat, ok := version.(float64); ok {
						return uint64(versionFloat)
					}
				}
			}
		}
	}
	return 0
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
		PackageObjectId: handler.PackageID,
		Module:          "treasury",
		Function:        "transfer_from_coin_cap",
		TypeArguments: []interface{}{
			handler.TokenType,
		},
		Arguments: []interface{}{
			handler.TreasuryId,
			string(handler.OwnerKeys.SuiAddress),
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
			if strings.Contains(obj.ObjectType, "::bridge_roles::BridgeCap") {
				handler.BridgeCap = obj.ObjectId
			}
			if strings.Contains(obj.ObjectType, "::safe::BridgeSafe") {
				handler.SafeObjectID = obj.ObjectId
				handler.SafeInitialSharedVersion = extractInitialSharedVersion(obj.Owner)
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
	suiRelayersPubKeys := params[2].([][suiPubKeyLength]byte)
	quorumStr := params[3].(string)
	safeObjectID := params[4].(string)
	adminCap := params[5].(string)

	resp := handler.SuiChainSimulator.MoveCall(ctx, models.MoveCallRequest{
		Signer:          string(handler.OwnerKeys.SuiAddress),
		PackageObjectId: handler.PackageID,
		Module:          module,
		Function:        function,
		TypeArguments:   []interface{}{},
		Arguments: []interface{}{
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
			handler.SafeObjectID,
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
			handler.SafeObjectID,
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
		},
		GasBudget: "10000000",
	}, handler.OwnerKeys)
}

// IssueAndWhitelistToken will issue and whitelist the token on Sui
func (handler *SuiHandler) IssueAndWhitelistToken(ctx context.Context, params IssueTokenParams) {
	isMintBurnAdapter := params.IsMintBurnOnPeerChain && !params.IsLocked

	var coinType string
	var suiTokenInfo SuiTokenInfo

	switch {
	case isMintBurnAdapter:
		coinType = fmt.Sprintf("%s::xmn::XMN", handler.PackageID)
		suiTokenInfo = SuiTokenInfo{
			CoinPackageId: handler.PackageID,
			TreasuryId:    handler.XmnTreasuryId,
			IsMintBurn:    true,
		}
	case params.IsLocked:
		suiTokenInfo = SuiTokenInfo{
			CoinPackageId: handler.PackageID,
			IsLocked:      true,
		}
		coinType = handler.TokenType
	default:
		coinPackageId, treasuryId, metadataId := handler.deployCoinContract(ctx)
		suiTokenInfo = SuiTokenInfo{
			CoinPackageId:  coinPackageId,
			TreasuryId:     treasuryId,
			CoinMetadataId: metadataId,
		}
		coinType = fmt.Sprintf("%s::test_coin::TEST_COIN", coinPackageId)
	}

	handler.TokensRegistry.RegisterPeerChainAddressAndInfo(params.AbstractTokenIdentifier, []byte(coinType), suiTokenInfo)

	if !params.IsLocked && !isMintBurnAdapter {
		handler.updateMetadata(ctx, params)
	}

	mintAmount, ok := big.NewInt(0).SetString(params.ValueToMintOnPeerChain, 10)
	require.True(handler, ok)

	switch {
	case isMintBurnAdapter:
		mintCapId := handler.setupMintCapForAdapter(ctx, coinType)
		// Mint to test user before moving MintCap into the safe
		if mintAmount.Sign() > 0 {
			handler.mintXmn(ctx, mintCapId, coinType, string(handler.TestKeys.SuiAddress), mintAmount)
		}
		handler.whitelistMintBurnToken(ctx, coinType, mintCapId)
		handler.MintBurnAdapterInfos[params.AbstractTokenIdentifier] = SuiMintBurnAdapterInfo{
			XmnTreasuryId:                   handler.XmnTreasuryId,
			XmnTreasuryInitialSharedVersion: handler.XmnTreasuryInitialSharedVersion,
		}
	case params.IsLocked:
		handler.mintLocked(ctx, params, string(handler.TestKeys.SuiAddress), mintAmount)
		handler.whitelistSafeToken(ctx, coinType, params)
	default:
		handler.mint(ctx, params, string(handler.TestKeys.SuiAddress), mintAmount)
		handler.whitelistSafeToken(ctx, coinType, params)
	}

	if len(params.InitialSupplyValue) > 0 && !isMintBurnAdapter {
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

func (handler *SuiHandler) whitelistSafeToken(ctx context.Context, coinType string, params IssueTokenParams) {
	handler.SuiChainSimulator.MoveCall(ctx, models.MoveCallRequest{
		Signer:          string(handler.OwnerKeys.SuiAddress),
		PackageObjectId: handler.PackageID,
		Module:          "safe",
		Function:        "whitelist_token",
		TypeArguments:   []interface{}{coinType},
		Arguments: []interface{}{
			handler.SafeObjectID,
			"25",
			"500000",
			params.IsNativeOnPeerChain,
			params.IsLocked,
		},
		GasBudget: "100000000",
	}, handler.OwnerKeys)
}

func (handler *SuiHandler) whitelistMintBurnToken(ctx context.Context, coinType string, mintCapId string) {
	handler.SuiChainSimulator.MoveCall(ctx, models.MoveCallRequest{
		Signer:          string(handler.OwnerKeys.SuiAddress),
		PackageObjectId: handler.PackageID,
		Module:          "xmn_mint_cap_adapter",
		Function:        "whitelist_token",
		TypeArguments:   []interface{}{coinType},
		Arguments: []interface{}{
			handler.SafeObjectID,
			"25",
			"500000000000000",
			mintCapId,
			handler.XmnTreasuryId,
		},
		GasBudget: "100000000",
	}, handler.OwnerKeys)
}

// setupMintCapForAdapter creates a MintCap for the XMN treasury and sets its allowance.
// Returns the MintCap object ID (still owned by owner at this point).
func (handler *SuiHandler) setupMintCapForAdapter(ctx context.Context, coinType string) string {
	resp := handler.SuiChainSimulator.MoveCall(ctx, models.MoveCallRequest{
		Signer:          string(handler.OwnerKeys.SuiAddress),
		PackageObjectId: handler.PackageID,
		Module:          "treasury",
		Function:        "configure_new_controller",
		TypeArguments:   []interface{}{coinType},
		Arguments: []interface{}{
			handler.XmnTreasuryId,
			string(handler.OwnerKeys.SuiAddress),
			string(handler.OwnerKeys.SuiAddress),
		},
		GasBudget: "100000000",
	}, handler.OwnerKeys)

	var mintCapId string
	for _, obj := range resp.ObjectChanges {
		if obj.Type == "created" && strings.Contains(obj.ObjectType, "::treasury::MintCap<") {
			mintCapId = obj.ObjectId
			break
		}
	}
	require.NotEmpty(handler, mintCapId, "MintCap not found in ObjectChanges after configure_new_controller")

	// Set a large allowance so the adapter can mint freely in tests
	handler.SuiChainSimulator.MoveCall(ctx, models.MoveCallRequest{
		Signer:          string(handler.OwnerKeys.SuiAddress),
		PackageObjectId: handler.PackageID,
		Module:          "treasury",
		Function:        "configure_minter",
		TypeArguments:   []interface{}{coinType},
		Arguments: []interface{}{
			handler.XmnTreasuryId,
			denyListObjectId,
			"18446744073709551615",
		},
		GasBudget: "100000000",
	}, handler.OwnerKeys)

	return mintCapId
}

func (handler *SuiHandler) mintXmn(ctx context.Context, mintCapId string, coinType string, receiver string, amount *big.Int) {
	handler.SuiChainSimulator.MoveCall(ctx, models.MoveCallRequest{
		Signer:          string(handler.OwnerKeys.SuiAddress),
		PackageObjectId: handler.PackageID,
		Module:          "treasury",
		Function:        "mint",
		TypeArguments:   []interface{}{coinType},
		Arguments: []interface{}{
			handler.XmnTreasuryId,
			mintCapId,
			denyListObjectId,
			amount.String(),
			receiver,
		},
		GasBudget: "100000000",
	}, handler.OwnerKeys)
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
		switch obj.Type {
		case "created":
			if strings.Contains(obj.ObjectType, "0x2::coin::TreasuryCap") {
				treasuryId = obj.ObjectId
			}
			if strings.Contains(obj.ObjectType, "0x2::coin::CoinMetadata") {
				metadataId = obj.ObjectId
			}
		case "published":
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
		PackageObjectId: handler.PackageID,
		Module:          "treasury",
		Function:        "mint_coin_to_receiver",
		TypeArguments: []interface{}{
			handler.TokenType,
		},
		Arguments: []interface{}{
			handler.TreasuryId,
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
			handler.SafeObjectID,
			coinObjId,
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

	suiTokenInfo := token.PeerChainTokenInfo.(SuiTokenInfo)
	coinType := string(token.PeerChainTokenAddress)

	for _, operation := range params.TestOperations {
		if operation.ValueToTransferToMvx == nil {
			continue
		}

		coinObjId := handler.getCoinObjectIdForToken(ctx, token.PeerChainTokenAddress, operation.ValueToTransferToMvx, handler.TestKeys)

		if suiTokenInfo.IsMintBurn {
			adapterInfo := handler.MintBurnAdapterInfos[params.AbstractTokenIdentifier]
			handler.SuiChainSimulator.MoveCall(ctx, models.MoveCallRequest{
				Signer:          string(from.SuiAddress),
				PackageObjectId: handler.PackageID,
				Module:          "xmn_mint_cap_adapter",
				Function:        "deposit",
				TypeArguments:   []interface{}{coinType},
				Arguments: []interface{}{
					handler.SafeObjectID,
					coinObjId,
					handler.TestKeys.MvxAddress.AddressSlice(),
					clockId,
					adapterInfo.XmnTreasuryId,
					denyListObjectId,
				},
				GasBudget: "10000000",
			}, from)
		} else {
			handler.SuiChainSimulator.MoveCall(ctx, models.MoveCallRequest{
				Signer:          string(from.SuiAddress),
				PackageObjectId: handler.PackageID,
				Module:          "safe",
				Function:        "deposit",
				TypeArguments:   []interface{}{coinType},
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

			coinTypeStr := string(coinAddress)
			for _, obj := range resp.ObjectChanges {
				if obj.Type == "created" && strings.Contains(obj.ObjectType, coinTypeStr) {
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

// buildTokenAdapterConfigs builds the SuiTokenAdapterConfig slice for all registered mint-burn tokens.
func (handler *SuiHandler) buildTokenAdapterConfigs() []config.SuiTokenAdapterConfig {
	var result []config.SuiTokenAdapterConfig
	for abstractTokenId, adapterInfo := range handler.MintBurnAdapterInfos {
		token := handler.TokensRegistry.GetTokenData(abstractTokenId)
		if token == nil {
			continue
		}
		coinType := string(token.PeerChainTokenAddress)
		result = append(result, config.SuiTokenAdapterConfig{
			CoinType:      coinType,
			AdapterModule: "xmn_mint_cap_adapter",
			AdapterObjects: []config.SuiAdapterObjectConfig{
				{
					ObjectId:             adapterInfo.XmnTreasuryId,
					InitialSharedVersion: adapterInfo.XmnTreasuryInitialSharedVersion,
					Mutable:              true,
				},
				{
					ObjectId:             denyListObjectId,
					InitialSharedVersion: denyListInitialSharedVersion,
					Mutable:              false,
				},
				{
					ObjectId:             clockId,
					InitialSharedVersion: 1,
					Mutable:              false,
				},
			},
		})
	}
	return result
}

func (handler *SuiHandler) Close() error {
	return nil
}
