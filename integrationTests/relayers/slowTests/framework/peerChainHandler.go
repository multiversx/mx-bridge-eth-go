package framework

import (
	"context"
	"math/big"
	"testing"

	"github.com/multiversx/mx-sdk-go/core"
)

// ChainType represents the type of peer chain
type ChainType string

const (
	ChainTypeEthereum ChainType = "ethereum"
	ChainTypeSui      ChainType = "sui"
)

// PeerChainHandler defines the common interface for all peer chain handlers (Ethereum, Sui, etc.)
type PeerChainHandler interface {
	// Contract deployment and setup
	DeployContracts(ctx context.Context)

	// Token management
	IssueAndWhitelistToken(ctx context.Context, params IssueTokenParams)
	GetBalance(receiver interface{}, abstractTokenIdentifier string) *big.Int
	Mint(ctx context.Context, params TestTokenParams, valueToMint *big.Int)

	// Contract state management
	PauseContractsForTokenChanges(ctx context.Context)
	UnPauseContractsAfterTokenChanges(ctx context.Context)

	// Transfer operations - Updated method names to be generic
	CreateBatch(ctx context.Context, mvxTestCallerAddress core.AddressHandler, tokensParams ...TestTokenParams)
	SendFromPeerChainToMultiversX(ctx context.Context, mvxTestCallerAddress core.AddressHandler, tokensParams ...TestTokenParams)

	// Resource management
	Close() error

	// Chain identification
	GetChainType() string
}

// PeerChainConfig represents configuration for creating a peer chain handler
type PeerChainConfig struct {
	ChainType      ChainType
	TestingTB      testing.TB
	Context        context.Context
	KeysStore      *KeysStore
	TokensRegistry TokensRegistry
	Quorum         string
}

// NewPeerChainHandler creates a new peer chain handler based on the chain type
func NewPeerChainHandler(config PeerChainConfig) PeerChainHandler {
	switch config.ChainType {
	case ChainTypeEthereum:
		return NewEthereumHandler(config.TestingTB, config.Context, config.KeysStore, config.TokensRegistry, config.Quorum)
	case ChainTypeSui:
		return NewSuiHandler(config.TestingTB, config.Context, config.KeysStore, config.TokensRegistry, config.Quorum)
	default:
		panic("unsupported chain type: " + string(config.ChainType))
	}
}
