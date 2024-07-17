package framework

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"math/big"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/abi/bind/backends"
	"github.com/ethereum/go-ethereum/common"
	ethCore "github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/multiversx/mx-bridge-eth-go/clients/ethereum"
	"github.com/multiversx/mx-bridge-eth-go/clients/ethereum/contract"
	"github.com/multiversx/mx-bridge-eth-go/clients/ethereum/wrappers"
	"github.com/multiversx/mx-bridge-eth-go/core/converters"
	"github.com/multiversx/mx-bridge-eth-go/integrationTests"
	"github.com/multiversx/mx-bridge-eth-go/parsers"
	"github.com/multiversx/mx-bridge-eth-go/testsCommon"
	"github.com/multiversx/mx-sdk-go/core"
	"github.com/stretchr/testify/require"
)

const (
	ethSimulatedGasLimit          = 9000000
	ethStatusSuccess              = uint64(1)
	minterRoleString              = "MINTER_ROLE"
	ethMinAmountAllowedToTransfer = 25
	ethMaxAmountAllowedToTransfer = 500000

	erc20SafeABI          = "testdata/contracts/eth/ERC20Safe.abi.json"
	erc20SafeBytecode     = "testdata/contracts/eth/ERC20Safe.hex"
	bridgeABI             = "testdata/contracts/eth/Bridge.abi.json"
	bridgeBytecode        = "testdata/contracts/eth/Bridge.hex"
	scExecProxyABI        = "testdata/contracts/eth/SCExecProxy.abi.json"
	scExecProxyBytecode   = "testdata/contracts/eth/SCExecProxy.hex"
	genericERC20ABI       = "testdata/contracts/eth/GenericERC20.abi.json"
	genericERC20Bytecode  = "testdata/contracts/eth/GenericERC20.hex"
	mintBurnERC20ABI      = "testdata/contracts/eth/MintBurnERC20.abi.json"
	mintBurnERC20Bytecode = "testdata/contracts/eth/MintBurnERC20.hex"
)

// EthereumHandler will handle all the operations on the Ethereum side
type EthereumHandler struct {
	testing.TB
	*KeysStore
	TokensRegistry        TokensRegistry
	Quorum                string
	MvxTestCallerAddress  core.AddressHandler
	SimulatedChain        *backends.SimulatedBackend
	SimulatedChainWrapper EthereumBlockchainClient
	ChainID               *big.Int
	SafeAddress           common.Address
	SafeContract          *contract.ERC20Safe
	SCProxyAddress        common.Address
	SCProxyContract       *contract.SCExecProxy
	BridgeAddress         common.Address
	BridgeContract        *contract.Bridge
	Erc20ContractsHolder  ethereum.Erc20ContractsHolder
	EthChainWrapper       ethereum.ClientWrapper
}

// NewEthereumHandler will create the handler that will adapt all test operations on Ethereum
func NewEthereumHandler(
	tb testing.TB,
	ctx context.Context,
	keysStore *KeysStore,
	tokensRegistry TokensRegistry,
	quorum string,
) *EthereumHandler {
	handler := &EthereumHandler{
		TB:             tb,
		KeysStore:      keysStore,
		TokensRegistry: tokensRegistry,
		Quorum:         quorum,
	}

	walletsToFundOnEthereum := handler.WalletsToFundOnEthereum()
	addr := make(map[common.Address]ethCore.GenesisAccount, len(walletsToFundOnEthereum))
	for _, address := range walletsToFundOnEthereum {
		addr[address] = ethCore.GenesisAccount{Balance: new(big.Int).Lsh(big.NewInt(1), 100)}
	}
	alloc := ethCore.GenesisAlloc(addr)
	handler.SimulatedChain = backends.NewSimulatedBackend(alloc, ethSimulatedGasLimit)

	handler.SimulatedChainWrapper = integrationTests.NewSimulatedETHChainWrapper(handler.SimulatedChain)
	handler.ChainID, _ = handler.SimulatedChainWrapper.ChainID(ctx)

	var err error
	handler.Erc20ContractsHolder, err = ethereum.NewErc20SafeContractsHolder(ethereum.ArgsErc20SafeContractsHolder{
		EthClient:              handler.SimulatedChain,
		EthClientStatusHandler: &testsCommon.StatusHandlerStub{},
	})
	require.NoError(tb, err)

	return handler
}

// DeployContracts will deploy all required contracts on Ethereum side
func (handler *EthereumHandler) DeployContracts(ctx context.Context) {
	// deploy safe
	handler.SafeAddress = handler.DeployContract(ctx, erc20SafeABI, erc20SafeBytecode)
	ethSafeContract, err := contract.NewERC20Safe(handler.SafeAddress, handler.SimulatedChain)
	require.NoError(handler, err)
	handler.SafeContract = ethSafeContract

	// deploy bridge
	ethRelayersAddresses := make([]common.Address, 0, len(handler.RelayersKeys))
	for _, relayerKeys := range handler.RelayersKeys {
		ethRelayersAddresses = append(ethRelayersAddresses, relayerKeys.EthAddress)
	}
	quorumInt, _ := big.NewInt(0).SetString(handler.Quorum, 10)
	handler.BridgeAddress = handler.DeployContract(ctx, bridgeABI, bridgeBytecode, ethRelayersAddresses, quorumInt, handler.SafeAddress)
	handler.BridgeContract, err = contract.NewBridge(handler.BridgeAddress, handler.SimulatedChain)
	require.NoError(handler, err)

	// set bridge on safe
	auth, _ := bind.NewKeyedTransactorWithChainID(handler.OwnerKeys.EthSK, handler.ChainID)
	tx, err := ethSafeContract.SetBridge(auth, handler.BridgeAddress)
	require.NoError(handler, err)
	handler.SimulatedChain.Commit()
	handler.checkEthTxResult(ctx, tx.Hash())

	// deploy exec-proxy
	handler.SCProxyAddress = handler.DeployContract(ctx, scExecProxyABI, scExecProxyBytecode, handler.SafeAddress)
	scProxyContract, err := contract.NewSCExecProxy(handler.SCProxyAddress, handler.SimulatedChain)
	require.NoError(handler, err)
	handler.SCProxyContract = scProxyContract

	handler.EthChainWrapper, err = wrappers.NewEthereumChainWrapper(wrappers.ArgsEthereumChainWrapper{
		StatusHandler:    &testsCommon.StatusHandlerStub{},
		MultiSigContract: handler.BridgeContract,
		SafeContract:     handler.SafeContract,
		BlockchainClient: handler.SimulatedChainWrapper,
	})
	require.NoError(handler, err)

	handler.UnPauseContractsAfterTokenChanges(ctx)
}

// DeployContract can deploy an Ethereum contract
func (handler *EthereumHandler) DeployContract(
	ctx context.Context,
	abiFile string,
	bytecodeFile string,
	params ...interface{},
) common.Address {
	abiBytes, err := os.ReadFile(abiFile)
	require.NoError(handler, err)
	parsed, err := abi.JSON(bytes.NewReader(abiBytes))
	require.NoError(handler, err)

	contractBytes, err := os.ReadFile(bytecodeFile)
	require.NoError(handler, err)

	contractAuth, _ := bind.NewKeyedTransactorWithChainID(handler.OwnerKeys.EthSK, handler.ChainID)
	contractAddress, tx, _, err := bind.DeployContract(contractAuth, parsed, common.FromHex(converters.TrimWhiteSpaceCharacters(string(contractBytes))), handler.SimulatedChain, params...)
	require.NoError(handler, err)
	handler.SimulatedChain.Commit()

	handler.checkEthTxResult(ctx, tx.Hash())

	log.Info("deployed eth contract", "from file", bytecodeFile, "address", contractAddress.Hex())

	return contractAddress
}

func (handler *EthereumHandler) checkEthTxResult(ctx context.Context, hash common.Hash) {
	receipt, err := handler.SimulatedChain.TransactionReceipt(ctx, hash)
	require.NoError(handler, err)
	require.Equal(handler, ethStatusSuccess, receipt.Status)
}

// GetBalance returns the receiver's balance
func (handler *EthereumHandler) GetBalance(receiver common.Address, abstractTokenIdentifier string) *big.Int {
	token := handler.TokensRegistry.GetTokenData(abstractTokenIdentifier)
	require.NotNil(handler, token)
	require.NotNil(handler, token.EthErc20Address)

	balance, err := token.EthErc20Contract.BalanceOf(nil, receiver)
	require.NoError(handler, err)

	return balance
}

// UnPauseContractsAfterTokenChanges can unpause contracts after token changes
func (handler *EthereumHandler) UnPauseContractsAfterTokenChanges(ctx context.Context) {
	auth, _ := bind.NewKeyedTransactorWithChainID(handler.OwnerKeys.EthSK, handler.ChainID)

	// unpause bridge contract
	tx, err := handler.BridgeContract.Unpause(auth)
	require.NoError(handler, err)
	handler.SimulatedChain.Commit()
	handler.checkEthTxResult(ctx, tx.Hash())

	// unpause safe contract
	tx, err = handler.SafeContract.Unpause(auth)
	require.NoError(handler, err)
	handler.SimulatedChain.Commit()
	handler.checkEthTxResult(ctx, tx.Hash())
}

// PauseContractsForTokenChanges can pause contracts for token changes
func (handler *EthereumHandler) PauseContractsForTokenChanges(ctx context.Context) {
	auth, _ := bind.NewKeyedTransactorWithChainID(handler.OwnerKeys.EthSK, handler.ChainID)

	// pause bridge contract
	tx, err := handler.BridgeContract.Pause(auth)
	require.NoError(handler, err)
	handler.SimulatedChain.Commit()
	handler.checkEthTxResult(ctx, tx.Hash())

	// pause safe contract
	tx, err = handler.SafeContract.Pause(auth)
	require.NoError(handler, err)
	handler.SimulatedChain.Commit()
	handler.checkEthTxResult(ctx, tx.Hash())
}

// IssueAndWhitelistToken will issue and whitelist the token on Ethereum
func (handler *EthereumHandler) IssueAndWhitelistToken(ctx context.Context, params IssueTokenParams) {
	erc20Address, erc20ContractInstance := handler.deployTestERC20Contract(ctx, params)

	handler.TokensRegistry.RegisterEthAddressAndContract(params.AbstractTokenIdentifier, erc20Address, erc20ContractInstance)

	// whitelist eth token
	auth, _ := bind.NewKeyedTransactorWithChainID(handler.OwnerKeys.EthSK, handler.ChainID)
	tx, err := handler.SafeContract.WhitelistToken(auth, erc20Address, big.NewInt(ethMinAmountAllowedToTransfer), big.NewInt(ethMaxAmountAllowedToTransfer), params.IsMintBurnOnEth, params.IsNativeOnEth)
	require.NoError(handler, err)
	handler.SimulatedChain.Commit()
	handler.checkEthTxResult(ctx, tx.Hash())
}

func (handler *EthereumHandler) deployTestERC20Contract(ctx context.Context, params IssueTokenParams) (common.Address, ERC20Contract) {
	if params.IsMintBurnOnEth {
		ethMintBurnAddress := handler.DeployContract(
			ctx,
			mintBurnERC20ABI,
			mintBurnERC20Bytecode,
			params.EthTokenName,
			params.EthTokenSymbol,
			params.NumOfDecimalsChainSpecific,
		)

		ethMintBurnContract, err := contract.NewMintBurnERC20(ethMintBurnAddress, handler.SimulatedChain)
		require.NoError(handler, err)

		ownerAuth, _ := bind.NewKeyedTransactorWithChainID(handler.OwnerKeys.EthSK, handler.ChainID)
		minterRoleBytes := [32]byte(crypto.Keccak256([]byte(minterRoleString)))

		// grant mint role to the depositor address for the initial mint
		txGrantRole, err := ethMintBurnContract.GrantRole(ownerAuth, minterRoleBytes, handler.DepositorKeys.EthAddress)
		require.NoError(handler, err)
		handler.SimulatedChain.Commit()
		handler.checkEthTxResult(ctx, txGrantRole.Hash())

		// grant mint role to the safe contract
		txGrantRole, err = ethMintBurnContract.GrantRole(ownerAuth, minterRoleBytes, handler.SafeAddress)
		require.NoError(handler, err)
		handler.SimulatedChain.Commit()
		handler.checkEthTxResult(ctx, txGrantRole.Hash())

		// mint generic token on the behalf of the depositor
		auth, _ := bind.NewKeyedTransactorWithChainID(handler.DepositorKeys.EthSK, handler.ChainID)

		mintAmount, ok := big.NewInt(0).SetString(params.ValueToMintOnEth, 10)
		require.True(handler, ok)
		tx, err := ethMintBurnContract.Mint(auth, handler.DepositorKeys.EthAddress, mintAmount)
		require.NoError(handler, err)
		handler.SimulatedChain.Commit()
		handler.checkEthTxResult(ctx, tx.Hash())

		balance, err := ethMintBurnContract.BalanceOf(nil, handler.DepositorKeys.EthAddress)
		require.NoError(handler, err)
		require.Equal(handler, mintAmount.String(), balance.String())

		return ethMintBurnAddress, ethMintBurnContract
	}

	// deploy generic eth token
	ethGenericTokenAddress := handler.DeployContract(
		ctx,
		genericERC20ABI,
		genericERC20Bytecode,
		params.EthTokenName,
		params.EthTokenSymbol,
		params.NumOfDecimalsChainSpecific,
	)

	ethGenericTokenContract, err := contract.NewGenericERC20(ethGenericTokenAddress, handler.SimulatedChain)
	require.NoError(handler, err)

	// mint the address that will create the transfers
	auth, _ := bind.NewKeyedTransactorWithChainID(handler.DepositorKeys.EthSK, handler.ChainID)

	mintAmount, ok := big.NewInt(0).SetString(params.ValueToMintOnEth, 10)
	require.True(handler, ok)
	tx, err := ethGenericTokenContract.Mint(auth, handler.TestKeys.EthAddress, mintAmount)
	require.NoError(handler, err)
	handler.SimulatedChain.Commit()
	handler.checkEthTxResult(ctx, tx.Hash())

	balance, err := ethGenericTokenContract.BalanceOf(nil, handler.TestKeys.EthAddress)
	require.NoError(handler, err)
	require.Equal(handler, mintAmount.String(), balance.String())

	return ethGenericTokenAddress, ethGenericTokenContract
}

// CreateBatchOnEthereum will create a batch on Ethereum using the provided tokens parameters list
func (handler *EthereumHandler) CreateBatchOnEthereum(
	ctx context.Context,
	mvxTestCallerAddress core.AddressHandler,
	tokensParams ...TestTokenParams,
) {
	for _, params := range tokensParams {
		handler.createDepositsOnEthereumForToken(ctx, params, handler.TestKeys.EthSK, mvxTestCallerAddress)
	}

	// wait until batch is settled
	batchSettleLimit, _ := handler.SafeContract.BatchSettleLimit(nil)
	for i := uint8(0); i < batchSettleLimit+1; i++ {
		handler.SimulatedChain.Commit()
	}
}

func (handler *EthereumHandler) createDepositsOnEthereumForToken(
	ctx context.Context,
	params TestTokenParams,
	from *ecdsa.PrivateKey,
	mvxTestCallerAddress core.AddressHandler,
) {
	// add allowance for the sender
	auth, _ := bind.NewKeyedTransactorWithChainID(from, handler.ChainID)

	token := handler.TokensRegistry.GetTokenData(params.AbstractTokenIdentifier)
	require.NotNil(handler, token)
	require.NotNil(handler, token.EthErc20Contract)

	allowanceValueForSafe := big.NewInt(0)
	allowanceValueForScProxy := big.NewInt(0)
	for _, operation := range params.TestOperations {
		if operation.ValueToTransferToMvx == nil {
			continue
		}
		if len(operation.MvxSCCallMethod) > 0 {
			allowanceValueForScProxy.Add(allowanceValueForScProxy, operation.ValueToTransferToMvx)
		} else {
			allowanceValueForSafe.Add(allowanceValueForSafe, operation.ValueToTransferToMvx)
		}
	}

	if allowanceValueForSafe.Cmp(zeroValueBigInt) > 0 {
		tx, err := token.EthErc20Contract.Approve(auth, handler.SafeAddress, allowanceValueForSafe)
		require.NoError(handler, err)
		handler.SimulatedChain.Commit()
		handler.checkEthTxResult(ctx, tx.Hash())
	}
	if allowanceValueForScProxy.Cmp(zeroValueBigInt) > 0 {
		tx, err := token.EthErc20Contract.Approve(auth, handler.SCProxyAddress, allowanceValueForScProxy)
		require.NoError(handler, err)
		handler.SimulatedChain.Commit()
		handler.checkEthTxResult(ctx, tx.Hash())
	}

	var err error
	for _, operation := range params.TestOperations {
		if operation.ValueToTransferToMvx == nil {
			continue
		}

		var tx *types.Transaction
		if len(operation.MvxSCCallMethod) > 0 {
			codec := parsers.MultiversxCodec{}
			callData := parsers.CallData{
				Type:      parsers.DataPresentProtocolMarker,
				Function:  operation.MvxSCCallMethod,
				GasLimit:  operation.MvxSCCallGasLimit,
				Arguments: operation.MvxSCCallArguments,
			}

			buff := codec.EncodeCallData(callData)

			tx, err = handler.SCProxyContract.Deposit(
				auth,
				token.EthErc20Address,
				operation.ValueToTransferToMvx,
				mvxTestCallerAddress.AddressSlice(),
				string(buff),
			)
		} else {
			tx, err = handler.SafeContract.Deposit(auth, token.EthErc20Address, operation.ValueToTransferToMvx, handler.TestKeys.MvxAddress.AddressSlice())
		}

		require.NoError(handler, err)
		handler.SimulatedChain.Commit()
		handler.checkEthTxResult(ctx, tx.Hash())
	}
}

// SendFromEthereumToMultiversX will create the deposit transactions on the Ethereum side
func (handler *EthereumHandler) SendFromEthereumToMultiversX(
	ctx context.Context,
	mvxTestCallerAddress core.AddressHandler,
	tokensParams ...TestTokenParams,
) {
	for _, params := range tokensParams {
		handler.createDepositsOnEthereumForToken(ctx, params, handler.TestKeys.EthSK, mvxTestCallerAddress)
	}
}

// Mint will mint the provided token on Ethereum with the provided value on the behalf of the Depositor address
func (handler *EthereumHandler) Mint(ctx context.Context, params TestTokenParams, valueToMint *big.Int) {
	token := handler.TokensRegistry.GetTokenData(params.AbstractTokenIdentifier)
	require.NotNil(handler, token)
	require.NotNil(handler, token.EthErc20Contract)

	// mint erc20 token into eth safe
	auth, _ := bind.NewKeyedTransactorWithChainID(handler.DepositorKeys.EthSK, handler.ChainID)
	tx, err := token.EthErc20Contract.Mint(auth, handler.SafeAddress, valueToMint)
	require.NoError(handler, err)
	handler.SimulatedChain.Commit()
	handler.checkEthTxResult(ctx, tx.Hash())
}

// Close will close the resources allocated
func (handler *EthereumHandler) Close() error {
	return handler.SimulatedChain.Close()
}
