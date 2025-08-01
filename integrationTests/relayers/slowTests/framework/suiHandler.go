package framework

import (
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	sdkCore "github.com/multiversx/mx-sdk-go/core"
	"github.com/stretchr/testify/require"
)

// SuiHandler will handle all the operations on the Sui side
type SuiHandler struct {
	testing.TB
	*KeysStore
	TokensRegistry TokensRegistry

	Quorum               string
	MvxTestCallerAddress sdkCore.AddressHandler
	SuiClient            SuiBlockchainClient
	BridgePackageID      []byte
	BridgeObjectID       []byte
	TokenContractsHolder map[string]MoveContract
	PublicKey            ed25519.PublicKey
	PrivateKey           ed25519.PrivateKey
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
		TB:                   tb,
		KeysStore:            keysStore,
		TokensRegistry:       tokensRegistry,
		Quorum:               quorum,
		TokenContractsHolder: make(map[string]MoveContract),
	}

	publicKey, privateKey, err := ed25519.GenerateKey(nil)
	require.NoError(tb, err)

	handler.PublicKey = publicKey
	handler.PrivateKey = privateKey

	handler.SuiClient = newMockSuiClient()

	return handler
}

func (handler *SuiHandler) DeployContracts(ctx context.Context) {
	handler.deployBridgePackage(ctx)
	handler.initializeBridge(ctx)
}

func (handler *SuiHandler) GetBalance(receiver interface{}, abstractTokenIdentifier string) *big.Int {
	var receiverBytes []byte

	switch addr := receiver.(type) {
	case common.Address:
		receiverBytes = addr.Bytes()
	case []byte:
		receiverBytes = addr
	case string:
		var err error
		receiverBytes, err = hex.DecodeString(addr)
		require.NoError(handler, err)
	default:
		handler.Fatalf("Unsupported receiver type: %T", receiver)
		return big.NewInt(0)
	}

	token := handler.TokensRegistry.GetTokenData(abstractTokenIdentifier)
	require.NotNil(handler, token)

	coinType := handler.getCoinType(abstractTokenIdentifier)
	balance, err := handler.SuiClient.GetBalance(context.Background(), receiverBytes, coinType)
	require.NoError(handler, err)

	return balance
}

func (handler *SuiHandler) IssueAndWhitelistToken(ctx context.Context, params IssueTokenParams) {
	tokenContract := handler.deployTokenContract(ctx, params)

	handler.TokensRegistry.AddToken(params)

	handler.TokenContractsHolder[params.AbstractTokenIdentifier] = tokenContract

	// Register the token contract with generic interface method
	handler.TokensRegistry.RegisterPeerChainAddressAndContract(
		params.AbstractTokenIdentifier,
		handler.PublicKey, // Use PublicKey as address for Sui
		tokenContract,
	)

	handler.whitelistToken(ctx, params.AbstractTokenIdentifier)
}

// CreateBatch creates a batch on Sui
func (handler *SuiHandler) CreateBatch(
	ctx context.Context,
	mvxTestCallerAddress sdkCore.AddressHandler,
	tokensParams ...TestTokenParams,
) {
	for _, params := range tokensParams {
		handler.createBatchOnSui(ctx, CreateBatchParams{
			TestTokenParams:   params,
			TestCallerAddress: mvxTestCallerAddress,
		})
	}
}

// SendFromPeerChainToMultiversX sends from Sui to MultiversX
func (handler *SuiHandler) SendFromPeerChainToMultiversX(
	ctx context.Context,
	mvxTestCallerAddress sdkCore.AddressHandler,
	tokensParams ...TestTokenParams,
) {
	for _, params := range tokensParams {
		handler.sendFromSuiToMultiversX(ctx, TestTransferParams{
			TestTokenParams:   params,
			TestCallerAddress: mvxTestCallerAddress,
		})
	}
}

func (handler *SuiHandler) Mint(ctx context.Context, params TestTokenParams, valueToMint *big.Int) {
	token := handler.TokensRegistry.GetTokenData(params.AbstractTokenIdentifier)
	require.NotNil(handler, token)

	contract, exists := handler.TokenContractsHolder[params.AbstractTokenIdentifier]
	require.True(handler, exists, "Token contract not found")

	receiverBytes := handler.PublicKey

	_, err := contract.Mint(ctx, receiverBytes, valueToMint)
	require.NoError(handler, err)
}

func (handler *SuiHandler) PauseContractsForTokenChanges(ctx context.Context) {
	handler.pauseBridge(ctx)
}

func (handler *SuiHandler) UnPauseContractsAfterTokenChanges(ctx context.Context) {
	handler.unpauseBridge(ctx)
}

func (handler *SuiHandler) Close() error {
	return nil
}

func (handler *SuiHandler) GetChainType() string {
	return string(ChainTypeSui)
}

// Private helper methods

func (handler *SuiHandler) deployBridgePackage(_ context.Context) {
	handler.BridgePackageID = []byte("mock_bridge_package_id")
	handler.BridgeObjectID = []byte("mock_bridge_object_id")
}

func (handler *SuiHandler) initializeBridge(_ context.Context) {
	// Mock implementation - initialize bridge with relayers and quorum
}

func (handler *SuiHandler) deployTokenContract(_ context.Context, params IssueTokenParams) MoveContract {
	return &mockMoveContract{
		tokenIdentifier: params.AbstractTokenIdentifier,
		handler:         handler,
	}
}

func (handler *SuiHandler) whitelistToken(_ context.Context, _ string) {
	// Mock whitelisting token in bridge
}

func (handler *SuiHandler) createBatchOnSui(_ context.Context, _ CreateBatchParams) {
	// Mock batch creation on Sui
}

func (handler *SuiHandler) sendFromSuiToMultiversX(_ context.Context, _ TestTransferParams) {
	// Mock transfer from Sui to MultiversX
}

func (handler *SuiHandler) pauseBridge(_ context.Context) {
	// Mock pause bridge
}

func (handler *SuiHandler) unpauseBridge(_ context.Context) {
	// Mock unpause bridge
}

func (handler *SuiHandler) getCoinType(_ string) string {
	return "0x2::sui::SUI" // Default SUI coin type
}
