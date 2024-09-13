package framework

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
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
	EthSCCallData        []byte
}

// TestTokenParams defines a token collection of operations in one or 2 batches
type TestTokenParams struct {
	IssueTokenParams
	TestOperations          []TokenOperations
	ESDTSafeExtraBalance    *big.Int
	EthTestAddrExtraBalance *big.Int
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
