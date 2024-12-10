package framework

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// DeltaBalancesOnKeys represents a map of ExtraBalancesHolder where the map's key is username
type DeltaBalancesOnKeys map[string]*DeltaBalanceHolder

// IssueTokenParams the parameters when issuing a new token
type IssueTokenParams struct {
	InitialSupplyParams
	AbstractTokenIdentifier string
	PreventWhitelist        bool

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

	// Ethereum
	EthTokenName     string
	EthTokenSymbol   string
	ValueToMintOnEth string
	IsMintBurnOnEth  bool
	IsNativeOnEth    bool
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
	IsFaultyDeposit      bool
	InvalidReceiver      []byte
}

// TestTokenParams defines a token collection of operations in one or 2 batches
type TestTokenParams struct {
	IssueTokenParams
	TestOperations []TokenOperations
	DeltaBalances  map[HalfBridgeIdentifier]DeltaBalancesOnKeys
	MintBurnChecks *MintBurnBalances
}

// TokenData represents a test token data
type TokenData struct {
	AbstractTokenIdentifier string

	MvxUniversalTokenTicker     string
	MvxChainSpecificTokenTicker string
	EthTokenName                string
	EthTokenSymbol              string

	MvxUniversalToken     string
	MvxChainSpecificToken string
	EthErc20Address       common.Address
	EthErc20Contract      ERC20Contract
}

// DeltaBalanceHolder holds the delta balances for a specific address
type DeltaBalanceHolder struct {
	OnEth    *big.Int
	OnMvx    *big.Int
	MvxToken TokenBalanceType
}

// MintBurnBalances holds the mint/burn tokens balances for a test token
type MintBurnBalances struct {
	TotalUniversalMint     *big.Int
	TotalChainSpecificMint *big.Int
	TotalUniversalBurn     *big.Int
	TotalChainSpecificBurn *big.Int
	SafeMintValue          *big.Int
	SafeBurnValue          *big.Int
}

// ESDTSupply represents the DTO that holds the supply values for a token
type ESDTSupply struct {
	Supply string `json:"supply"`
	Minted string `json:"minted"`
	Burned string `json:"burned"`
}
