package framework

import (
	"math/big"
)

// IssueTokenParams the parameters when issuing a new token
type IssueTokenParams struct {
	InitialSupplyParams
	AbstractTokenIdentifier string

	// MultiversX
	NumOfDecimalsUniversal           int
	NumOfDecimalsChainSpecific       byte
	MvxUniversalTokenTicker          string
	MvxChainSpecificTokenTicker      string
	MvxUniversalTokenDisplayName     string
	MvxChainSpecificTokenDisplayName string
	ValueToMintOnMvx                 string
	IsMintBurnOnMvX                  bool
	IsNativeOnMvX                    bool
	HasChainSpecificToken            bool

	// Peer chain
	PeerChainTokenName     string
	PeerChainTokenSymbol   string
	ValueToMintOnPeerChain string
	IsMintBurnOnPeerChain  bool
	IsNativeOnPeerChain    bool
	PeerChainType          ChainType
	IsLocked               bool
}

// InitialSupplyParams represents the initial supply parameters
type InitialSupplyParams struct {
	InitialSupplyValue string
}

// TokenOperations defines a token operation in a test. Usually this can define one or to deposits in a batch
type TokenOperations struct {
	ValueToTransferToMvx *big.Int
	ValueToSendFromMvX   *big.Int
	MvxSCCallData        []byte
	MvxFaultySCCall      bool
	MvxForceSCCall       bool
}

// TestTokenParams defines a token collection of operations in one or 2 batches
type TestTokenParams struct {
	IssueTokenParams
	TestOperations                []TokenOperations
	ESDTSafeExtraBalance          *big.Int
	PeerChainTestAddrExtraBalance *big.Int
}

// TokenData represents a test token data
type TokenData struct {
	AbstractTokenIdentifier string

	MvxUniversalTokenTicker     string
	MvxChainSpecificTokenTicker string
	PeerChainTokenName          string
	PeerChainTokenSymbol        string

	MvxUniversalToken     string
	MvxChainSpecificToken string
	PeerChainTokenAddress []byte
	PeerChainTokenInfo    interface{}
}

type EthTokenInfo struct {
	Contract ERC20Contract
}

type SuiTokenInfo struct {
	CoinPackageId  string
	TreasuryId     string
	CoinMetadataId string
	IsLocked       bool
}

type ChainType string

const (
	ChainTypeEthereum ChainType = "ethereum"
	ChainTypeSui      ChainType = "sui"
)
