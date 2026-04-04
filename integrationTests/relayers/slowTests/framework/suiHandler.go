package framework

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"fmt"
	"io"
	"math/big"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/block-vision/sui-go-sdk/models"
	"github.com/multiversx/mx-bridge-eth-go/config"
	"github.com/multiversx/mx-sdk-go/core"
	"github.com/stretchr/testify/require"
)

const (
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
	TokensRegistry       TokensRegistry
	Quorum               string
	MvxTestCallerAddress core.AddressHandler
	SuiChainSimulator    *suiChainSimulatorWrapper
	// PackageID is the bridge_safe package ID
	PackageID                       string
	LockedTokenPkgId                string
	StablecoinTreasuryPkgId         string
	XmnPkgId                        string
	BridgeObjectID                  string
	BridgeInitialSharedVersion      uint64
	SafeObjectID                    string
	SafeInitialSharedVersion        uint64
	SafeInitCapId                   string
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
	docsAbsPath, err := filepath.Abs("../../../.docs")
	require.NoError(handler, err)

	suiExtDir := filepath.Join(docsAbsPath, "stablecoin-sui", "packages", "sui_extensions")
	lockedTokenSrc := filepath.Join(docsAbsPath, "mx-locked-token-sc-sui")
	treasurySrc := filepath.Join(docsAbsPath, "stablecoin-sui", "packages", "treasury")
	xmnSrc := filepath.Join(docsAbsPath, "stablecoin-sui", "packages", "xmn")
	bridgeSrc := filepath.Join(docsAbsPath, "mx-bridge-sc-sui")

	// Step 1: Deploy sui_extensions — no external package deps.
	// After publishing, set published-at in sui_extensions' own Move.toml so
	// downstream packages can reference it as a local dep without triggering
	// git downloads for its transitive framework dependencies.
	suiExtTmp := makeTempPackageCopy(handler.TB, suiExtDir, nil)
	defer func() { _ = os.RemoveAll(suiExtTmp) }()
	resp1 := handler.SuiChainSimulator.BuildAndPublish(ctx, suiExtTmp, handler.OwnerKeys)
	suiExtPkgId := extractPublishedPkgId(resp1, "two_step_role")
	require.NotEmpty(handler, suiExtPkgId, "suiExtPkgId not found")
	setPackagePublishedAt(handler.TB, suiExtTmp, "sui_extensions", suiExtPkgId)

	// Step 2: Deploy locked_token (depends on sui_extensions).
	lockedTokenToml := makeMoveToml("locked_token", "2024.beta", map[string]depEntry{
		"sui_extensions": {local: suiExtTmp},
	})
	lockedTokenTmp := makeTempPackageCopy(handler.TB, lockedTokenSrc, lockedTokenToml)
	defer func() { _ = os.RemoveAll(lockedTokenTmp) }()
	resp2 := handler.SuiChainSimulator.BuildAndPublish(ctx, lockedTokenTmp, handler.OwnerKeys)
	handler.extractLockedTokenDeployResult(resp2)
	setPackagePublishedAt(handler.TB, lockedTokenTmp, "locked_token", handler.LockedTokenPkgId)

	// Step 3: Deploy stablecoin treasury (depends on sui_extensions).
	treasuryToml := makeMoveToml("treasury", "2024.beta", map[string]depEntry{
		"sui_extensions": {local: suiExtTmp},
	})
	treasuryTmp := makeTempPackageCopy(handler.TB, treasurySrc, treasuryToml)
	defer func() { _ = os.RemoveAll(treasuryTmp) }()
	resp3 := handler.SuiChainSimulator.BuildAndPublish(ctx, treasuryTmp, handler.OwnerKeys)
	handler.StablecoinTreasuryPkgId = extractPublishedPkgId(resp3, "treasury")
	require.NotEmpty(handler, handler.StablecoinTreasuryPkgId, "StablecoinTreasuryPkgId not found")
	setPackagePublishedAt(handler.TB, treasuryTmp, "treasury", handler.StablecoinTreasuryPkgId)

	// Step 4: Deploy xmn (depends on treasury + sui_extensions).
	// Both treasury and sui_extensions tmp dirs now have published-at in their own
	// [package] sections, so the CLI won't try to re-resolve their transitive deps.
	xmnToml := makeMoveToml("xmn", "2024.beta", map[string]depEntry{
		"treasury":       {local: treasuryTmp},
		"sui_extensions": {local: suiExtTmp},
	})
	xmnTmp := makeTempPackageCopy(handler.TB, xmnSrc, xmnToml)
	defer func() { _ = os.RemoveAll(xmnTmp) }()
	resp4 := handler.SuiChainSimulator.BuildAndPublish(ctx, xmnTmp, handler.OwnerKeys)
	handler.extractXmnDeployResult(resp4)

	// Step 5: Deploy bridge_safe (depends on locked_token, treasury, sui_extensions).
	bridgeDir := handler.prepareBridgeDir(bridgeSrc, suiExtTmp, lockedTokenTmp, treasuryTmp)
	defer func() { _ = os.RemoveAll(bridgeDir) }()
	resp5 := handler.SuiChainSimulator.BuildAndPublish(ctx, bridgeDir, handler.OwnerKeys)
	handler.extractBridgePublishResult(resp5)

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

func (handler *SuiHandler) extractLockedTokenDeployResult(resp models.SuiTransactionBlockResponse) {
	for _, obj := range resp.ObjectChanges {
		switch obj.Type {
		case "published":
			handler.LockedTokenPkgId = obj.PackageId
		case "created":
			if strings.Contains(obj.ObjectType, "0x2::coin::TreasuryCap") {
				handler.TokenType = extractInnerType(obj.ObjectType)
			}
			if strings.Contains(obj.ObjectType, "::treasury::Treasury<") {
				handler.TreasuryId = obj.ObjectId
				handler.TreasuryInitialSharedVersion = extractInitialSharedVersion(obj.Owner)
			}
		}
	}
	require.NotEmpty(handler, handler.LockedTokenPkgId, "LockedTokenPkgId not found in publish response")
	require.NotEmpty(handler, handler.TreasuryId, "locked token TreasuryId not found in publish response")
	require.NotEmpty(handler, handler.TokenType, "TokenType not found in publish response")
}

func (handler *SuiHandler) extractXmnDeployResult(resp models.SuiTransactionBlockResponse) {
	for _, obj := range resp.ObjectChanges {
		switch obj.Type {
		case "published":
			handler.XmnPkgId = obj.PackageId
		case "created":
			if strings.Contains(obj.ObjectType, "::treasury::Treasury<") && strings.Contains(obj.ObjectType, "::xmn::XMN") {
				handler.XmnTreasuryId = obj.ObjectId
				handler.XmnTreasuryInitialSharedVersion = extractInitialSharedVersion(obj.Owner)
			}
		}
	}
	require.NotEmpty(handler, handler.XmnPkgId, "XmnPkgId not found in publish response")
	require.NotEmpty(handler, handler.XmnTreasuryId, "XmnTreasuryId not found in publish response")
}

func (handler *SuiHandler) extractBridgePublishResult(resp models.SuiTransactionBlockResponse) {
	for _, obj := range resp.ObjectChanges {
		switch obj.Type {
		case "published":
			for _, m := range obj.Modules {
				if m == "safe" {
					handler.PackageID = obj.PackageId
				}
			}
		case "created":
			if strings.Contains(obj.ObjectType, "::safe::SafeInitCap") {
				handler.SafeInitCapId = obj.ObjectId
			}
		}
	}
	require.NotEmpty(handler, handler.PackageID, "bridge PackageID not found in publish response")
	require.NotEmpty(handler, handler.SafeInitCapId, "SafeInitCapId not found in publish response")
}

// prepareBridgeDir creates a temp copy of the bridge source dir with a Move.toml
// pointing to already-published dependency tmp dirs (each with published-at set in
// their own [package] section).
func (handler *SuiHandler) prepareBridgeDir(bridgeSrc, suiExtTmp, lockedTokenTmp, treasuryTmp string) string {
	toml := makeMoveToml("bridge_safe", "2024.beta", map[string]depEntry{
		"locked_token":   {local: lockedTokenTmp},
		"treasury":       {local: treasuryTmp},
		"sui_extensions": {local: suiExtTmp, override: true},
	})

	return makeTempPackageCopy(handler.TB, bridgeSrc, toml)
}

// depEntry describes a Move dependency entry in Move.toml.
type depEntry struct {
	local    string
	override bool
}

// makeMoveToml generates a Move.toml []byte for a package with the given name, edition, and deps.
// Each dep entry only specifies the local path; published-at must be set in the dep's own
// [package] section (via setPackagePublishedAt) so the Sui CLI recognises it as deployed.
func makeMoveToml(pkgName, edition string, deps map[string]depEntry) []byte {
	var sb strings.Builder
	fmt.Fprintf(&sb, "[package]\nname = %q\nedition = %q\nversion = \"0.0.1\"\n\n[dependencies]\n", pkgName, edition)
	for name, dep := range deps {
		line := fmt.Sprintf("%s = { local = %q", name, dep.local)
		if dep.override {
			line += ", override = true"
		}
		line += " }\n"
		sb.WriteString(line)
	}
	fmt.Fprintf(&sb, "\n[addresses]\n%s = \"0x0\"\n", pkgName)
	return []byte(sb.String())
}

// setPackagePublishedAt updates a Move.toml in tmpDir to mark the package as already
// deployed: inserts published-at = pkgId into the [package] section and updates the
// named address from "0x0" to pkgId. This tells the Sui CLI the package is published
// so it won't try to re-resolve or republish the package's transitive dependencies.
func setPackagePublishedAt(tb testing.TB, tmpDir, pkgName, pkgId string) {
	tomlPath := filepath.Join(tmpDir, "Move.toml")
	content, err := os.ReadFile(tomlPath)
	require.NoError(tb, err)

	s := string(content)

	// Insert or update published-at in the [package] section.
	// The [package] section runs from "[package]" up to the next "[" section header.
	pubAtLineRe := regexp.MustCompile(`(?m)^published-at\s*=\s*"[^"]*"`)
	if pubAtLineRe.MatchString(s) {
		s = pubAtLineRe.ReplaceAllString(s, fmt.Sprintf(`published-at = %q`, pkgId))
	} else {
		pkgSectionRe := regexp.MustCompile(`(\[package\][^\[]*)`)
		s = pkgSectionRe.ReplaceAllStringFunc(s, func(section string) string {
			trimmed := strings.TrimRight(section, "\n")
			return trimmed + fmt.Sprintf("\npublished-at = %q\n", pkgId)
		})
	}

	// Update the named address (e.g. pkgName = "0x0") to the real package ID.
	addrRe := regexp.MustCompile(fmt.Sprintf(`(?m)^%s\s*=\s*"[^"]*"`, regexp.QuoteMeta(pkgName)))
	if addrRe.MatchString(s) {
		s = addrRe.ReplaceAllString(s, fmt.Sprintf(`%s = %q`, pkgName, pkgId))
	}

	err = os.WriteFile(tomlPath, []byte(s), 0644)
	require.NoError(tb, err)
}

// extractPublishedPkgId finds the PackageId from a "published" objectChange that contains the given module name.
func extractPublishedPkgId(resp models.SuiTransactionBlockResponse, moduleName string) string {
	for _, obj := range resp.ObjectChanges {
		if obj.Type != "published" {
			continue
		}
		for _, m := range obj.Modules {
			if m == moduleName {
				return obj.PackageId
			}
		}
	}
	// Fallback: return the first published package ID
	for _, obj := range resp.ObjectChanges {
		if obj.Type == "published" {
			return obj.PackageId
		}
	}
	return ""
}

// makeTempPackageCopy copies a Move package to a temp dir, removes Move.lock, absolutizes local
// dep paths in Move.toml, and optionally replaces Move.toml entirely (when overrideToml != nil).
func makeTempPackageCopy(tb testing.TB, srcDir string, overrideToml []byte) string {
	tmpDir, err := os.MkdirTemp("", "sui-pkg-*")
	require.NoError(tb, err)

	err = copyDirRecursive(srcDir, tmpDir)
	require.NoError(tb, err)

	if overrideToml != nil {
		// We're replacing Move.toml entirely, so the existing Move.lock (if any) would
		// have a stale manifest_digest and stale local dep paths.  Delete it so the CLI
		// regenerates a clean one using the deps' own published-at markers.
		_ = os.Remove(filepath.Join(tmpDir, "Move.lock"))
		err = os.WriteFile(filepath.Join(tmpDir, "Move.toml"), overrideToml, 0644)
		require.NoError(tb, err)
	} else {
		// Keep Move.lock — it pins the framework git revisions so the CLI uses the
		// cached ~/.move packages instead of downloading from git during build.
		// Make all local = "..." paths in Move.toml absolute so the package works from any dir
		tomlPath := filepath.Join(tmpDir, "Move.toml")
		content, readErr := os.ReadFile(tomlPath)
		require.NoError(tb, readErr)
		updated := absolutizeLocalDeps(srcDir, content)
		err = os.WriteFile(tomlPath, updated, 0644)
		require.NoError(tb, err)
	}

	return tmpDir
}

// absolutizeLocalDeps replaces all `local = "..."` relative paths in Move.toml content
// with absolute paths resolved relative to srcDir.
func absolutizeLocalDeps(srcDir string, content []byte) []byte {
	re := regexp.MustCompile(`local\s*=\s*"([^"]*)"`)
	return re.ReplaceAllFunc(content, func(match []byte) []byte {
		sub := re.FindSubmatch(match)
		if len(sub) < 2 {
			return match
		}
		relPath := string(sub[1])
		absPath, absErr := filepath.Abs(filepath.Join(srcDir, relPath))
		if absErr != nil {
			return match
		}
		return []byte(fmt.Sprintf("local = %q", absPath))
	})
}

func copyDirRecursive(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, relErr := filepath.Rel(src, path)
		if relErr != nil {
			return relErr
		}
		dstPath := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}
		return copyFile(path, dstPath)
	})
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() { _ = out.Close() }()

	_, err = io.Copy(out, in)
	return err
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
		PackageObjectId: handler.LockedTokenPkgId,
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
			handler.SafeInitCapId,
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
		tokenBal := handler.SuiChainSimulator.GetTokenBalance(ctx, string(receiver), handler.TokenType)
		coinBal := handler.SuiChainSimulator.GetCoinBalance(ctx, string(receiver), string(token.PeerChainTokenAddress))
		log.Info("GetBalance for locked token", "receiver", string(receiver), "tokenBalance(Token<T>)", tokenBal, "coinBalance(Coin<T>)", coinBal)
		return tokenBal
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
		coinType = fmt.Sprintf("%s::xmn::XMN", handler.XmnPkgId)
		suiTokenInfo = SuiTokenInfo{
			CoinPackageId: handler.XmnPkgId,
			TreasuryId:    handler.XmnTreasuryId,
			IsMintBurn:    true,
		}
	case params.IsLocked:
		suiTokenInfo = SuiTokenInfo{
			CoinPackageId: handler.LockedTokenPkgId,
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
		PackageObjectId: handler.StablecoinTreasuryPkgId,
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
		PackageObjectId: handler.StablecoinTreasuryPkgId,
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
		PackageObjectId: handler.StablecoinTreasuryPkgId,
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
		PackageObjectId: handler.LockedTokenPkgId,
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
