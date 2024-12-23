package framework

import (
	"bytes"
	"context"
	"math/big"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient/simulated"
	"github.com/multiversx/mx-bridge-eth-go/clients/ethereum"
	"github.com/multiversx/mx-bridge-eth-go/clients/ethereum/contract"
	"github.com/multiversx/mx-bridge-eth-go/clients/ethereum/wrappers"
	"github.com/multiversx/mx-bridge-eth-go/core/converters"
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

	erc20SafeABI          = SlowTestWorkingDir + "testdata/contracts/eth/ERC20Safe.abi.json"
	erc20SafeBytecode     = SlowTestWorkingDir + "testdata/contracts/eth/ERC20Safe.hex"
	bridgeABI             = SlowTestWorkingDir + "testdata/contracts/eth/Bridge.abi.json"
	bridgeBytecode        = SlowTestWorkingDir + "testdata/contracts/eth/Bridge.hex"
	genericERC20ABI       = SlowTestWorkingDir + "testdata/contracts/eth/GenericERC20.abi.json"
	genericERC20Bytecode  = SlowTestWorkingDir + "testdata/contracts/eth/GenericERC20.hex"
	mintBurnERC20ABI      = SlowTestWorkingDir + "testdata/contracts/eth/MintBurnERC20.abi.json"
	mintBurnERC20Bytecode = SlowTestWorkingDir + "testdata/contracts/eth/MintBurnERC20.hex"
	proxyABI              = SlowTestWorkingDir + "testdata/contracts/eth/Proxy.abi.json"
	proxyBytecode         = SlowTestWorkingDir + "testdata/contracts/eth/Proxy.hex"

	proxyInitializeFunction = "initialize"
)

// EthereumHandler will handle all the operations on the Ethereum side
type EthereumHandler struct {
	testing.TB
	*KeysStore
	TokensRegistry        TokensRegistry
	Quorum                string
	MvxCalleeScAddress    core.AddressHandler
	SimulatedChain        *simulated.Backend
	SimulatedChainWrapper EthereumBlockchainClient
	ChainID               *big.Int
	SafeAddress           common.Address
	SafeContract          *contract.ERC20Safe
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
	addr := make(map[common.Address]types.Account, len(walletsToFundOnEthereum))
	for _, address := range walletsToFundOnEthereum {
		addr[address] = types.Account{Balance: new(big.Int).Lsh(big.NewInt(1), 100)}
	}
	alloc := types.GenesisAlloc(addr)
	handler.SimulatedChain = simulated.NewBackend(alloc,
		simulated.WithBlockGasLimit(ethSimulatedGasLimit),
	)

	handler.SimulatedChainWrapper = handler.SimulatedChain.Client()
	handler.ChainID, _ = handler.SimulatedChainWrapper.ChainID(ctx)

	var err error
	handler.Erc20ContractsHolder, err = ethereum.NewErc20SafeContractsHolder(ethereum.ArgsErc20SafeContractsHolder{
		EthClient:              handler.SimulatedChain.Client(),
		EthClientStatusHandler: &testsCommon.StatusHandlerStub{},
	})
	require.NoError(tb, err)

	return handler
}

// DeployContracts will deploy all required contracts on Ethereum side
func (handler *EthereumHandler) DeployContracts(ctx context.Context) {
	// deploy safe
	handler.SafeAddress = handler.DeployUpgradeableContract(ctx, erc20SafeABI, erc20SafeBytecode)
	ethSafeContract, err := contract.NewERC20Safe(handler.SafeAddress, handler.SimulatedChain.Client())
	require.NoError(handler, err)
	handler.SafeContract = ethSafeContract

	// deploy bridge
	ethRelayersAddresses := make([]common.Address, 0, len(handler.RelayersKeys))
	for _, relayerKeys := range handler.RelayersKeys {
		ethRelayersAddresses = append(ethRelayersAddresses, relayerKeys.EthAddress)
	}
	quorumInt, _ := big.NewInt(0).SetString(handler.Quorum, 10)
	handler.BridgeAddress = handler.DeployUpgradeableContract(ctx, bridgeABI, bridgeBytecode, ethRelayersAddresses, quorumInt, handler.SafeAddress)
	handler.BridgeContract, err = contract.NewBridge(handler.BridgeAddress, handler.SimulatedChain.Client())
	require.NoError(handler, err)

	// set bridge on safe
	auth, _ := bind.NewKeyedTransactorWithChainID(handler.OwnerKeys.EthSK, handler.ChainID)
	tx, err := ethSafeContract.SetBridge(auth, handler.BridgeAddress)

	require.NoError(handler, err)
	handler.SimulatedChain.Commit()
	handler.checkEthTxResult(ctx, tx.Hash())

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
	contractAddress, tx, _, err := bind.DeployContract(contractAuth, parsed, common.FromHex(converters.TrimWhiteSpaceCharacters(string(contractBytes))), handler.SimulatedChain.Client(), params...)
	require.NoError(handler, err)
	handler.SimulatedChain.Commit()

	handler.checkEthTxResult(ctx, tx.Hash())

	log.Info("deployed eth contract", "from file", bytecodeFile, "address", contractAddress.Hex())

	return contractAddress
}

// DeployUpgradeableContract can deploy an upgradeable Ethereum contract
func (handler *EthereumHandler) DeployUpgradeableContract(
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
	contractAddress, tx, _, err := bind.DeployContract(contractAuth, parsed, common.FromHex(converters.TrimWhiteSpaceCharacters(string(contractBytes))), handler.SimulatedChain.Client()) // no parameters on the logic contract constructor
	require.NoError(handler, err)
	handler.SimulatedChain.Commit()

	handler.checkEthTxResult(ctx, tx.Hash())

	log.Info("deployed eth logic contract", "from file", bytecodeFile, "address", contractAddress.Hex())

	packedParams, err := parsed.Pack(proxyInitializeFunction, params...)
	require.NoError(handler, err)
	proxyParams := []interface{}{
		contractAddress,
		handler.OwnerKeys.EthAddress, // make the owner of the logic contract the admin for the proxy
		packedParams,
	}
	proxyAddress := handler.DeployContract(ctx, proxyABI, proxyBytecode, proxyParams...)

	log.Info("deployed proxy contract", "address", proxyAddress.Hex())

	return proxyAddress // return the proxy to test that it behaves just the same as the logic contract
}

func (handler *EthereumHandler) checkEthTxResult(ctx context.Context, hash common.Hash) {
	receipt, err := handler.SimulatedChain.Client().TransactionReceipt(ctx, hash)
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

	if params.PreventWhitelist {
		return
	}

	// whitelist eth token
	auth, _ := bind.NewKeyedTransactorWithChainID(handler.OwnerKeys.EthSK, handler.ChainID)
	tx, err := handler.SafeContract.WhitelistToken(
		auth,
		erc20Address,
		big.NewInt(ethMinAmountAllowedToTransfer),
		big.NewInt(ethMaxAmountAllowedToTransfer),
		params.IsMintBurnOnEth,
		params.IsNativeOnEth,
		zeroValueBigInt,
		zeroValueBigInt,
		zeroValueBigInt,
	)
	require.NoError(handler, err)
	handler.SimulatedChain.Commit()
	handler.checkEthTxResult(ctx, tx.Hash())

	if len(params.InitialSupplyValue) > 0 {
		initialSupply, ok := big.NewInt(0).SetString(params.InitialSupplyValue, 10)
		require.True(handler, ok)

		if params.IsMintBurnOnEth {
			mintAmount := big.NewInt(0)
			burnAmount := big.NewInt(0)

			if params.IsNativeOnEth {
				burnAmount = initialSupply
			} else {
				mintAmount = initialSupply
			}

			tx, err = handler.SafeContract.InitSupplyMintBurn(auth, erc20Address, mintAmount, burnAmount)
			require.NoError(handler, err)
			handler.SimulatedChain.Commit()
			handler.checkEthTxResult(ctx, tx.Hash())
		} else {
			// reset the tokens value for the safe contract, so it will "know" about the balance that it has in the ERC20 contract
			tx, err = handler.SafeContract.ResetTotalBalance(auth, erc20Address)
			require.NoError(handler, err)
			handler.SimulatedChain.Commit()
			handler.checkEthTxResult(ctx, tx.Hash())
		}
	}
}

func (handler *EthereumHandler) deployTestERC20Contract(ctx context.Context, params IssueTokenParams) (common.Address, ERC20Contract) {
	if params.IsMintBurnOnEth {
		ethMintBurnAddress := handler.DeployUpgradeableContract(
			ctx,
			mintBurnERC20ABI,
			mintBurnERC20Bytecode,
			params.EthTokenName,
			params.EthTokenSymbol,
			params.NumOfDecimalsChainSpecific,
		)

		ethMintBurnContract, err := contract.NewMintBurnERC20(ethMintBurnAddress, handler.SimulatedChain.Client())
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

	ethGenericTokenContract, err := contract.NewGenericERC20(ethGenericTokenAddress, handler.SimulatedChain.Client())
	require.NoError(handler, err)

	// mint to the depositor
	handler.mintTokens(ctx, ethGenericTokenContract, params.ValueToMintOnEth, handler.DepositorKeys.EthAddress)
	if len(params.InitialSupplyValue) > 0 {
		handler.mintTokens(ctx, ethGenericTokenContract, params.InitialSupplyValue, handler.SafeAddress)
	}

	return ethGenericTokenAddress, ethGenericTokenContract
}

func (handler *EthereumHandler) mintTokens(
	ctx context.Context,
	ethGenericTokenContract *contract.GenericERC20,
	value string,
	recipientAddress common.Address,
) {
	auth, _ := bind.NewKeyedTransactorWithChainID(handler.DepositorKeys.EthSK, handler.ChainID)

	mintAmount, ok := big.NewInt(0).SetString(value, 10)
	require.True(handler, ok)

	tx, err := ethGenericTokenContract.Mint(auth, recipientAddress, mintAmount)
	require.NoError(handler, err)
	handler.SimulatedChain.Commit()
	handler.checkEthTxResult(ctx, tx.Hash())

	balance, err := ethGenericTokenContract.BalanceOf(nil, recipientAddress)
	require.NoError(handler, err)
	require.Equal(handler, mintAmount.String(), balance.String())
}

// ApproveForToken will approve the spender to spend the amount of tokens on the behalf of the approver
func (handler *EthereumHandler) ApproveForToken(
	ctx context.Context,
	token *TokenData,
	approver KeysHolder,
	spender common.Address,
	amount *big.Int,
) {
	auth, _ := bind.NewKeyedTransactorWithChainID(approver.EthSK, handler.ChainID)
	tx, err := token.EthErc20Contract.Approve(auth, spender, amount)
	require.NoError(handler, err)
	handler.SimulatedChain.Commit()
	handler.checkEthTxResult(ctx, tx.Hash())
}

// SendDepositTransactionFromEthereum will send a deposit transaction from Ethereum to MultiversX
func (handler *EthereumHandler) SendDepositTransactionFromEthereum(
	ctx context.Context,
	from KeysHolder,
	to KeysHolder,
	targetSCAddress core.AddressHandler,
	token *TokenData,
	operation TokenOperations,
) {
	if operation.ValueToTransferToMvx == nil {
		return
	}

	auth, _ := bind.NewKeyedTransactorWithChainID(from.EthSK, handler.ChainID)

	var tx *types.Transaction
	var err error
	if len(operation.MvxSCCallData) > 0 || operation.MvxForceSCCall {
		tx, err = handler.SafeContract.DepositWithSCExecution(
			auth,
			token.EthErc20Address,
			operation.ValueToTransferToMvx,
			targetSCAddress.AddressSlice(),
			operation.MvxSCCallData,
		)
	} else {
		tx, err = handler.SafeContract.Deposit(auth, token.EthErc20Address, operation.ValueToTransferToMvx, to.MvxAddress.AddressSlice())
	}

	if operation.IsFaultyDeposit {
		require.NotNil(handler, err)
		return
	}
	require.NoError(handler, err)
	handler.SimulatedChain.Commit()
	handler.checkEthTxResult(ctx, tx.Hash())
}

// SettleBatchOnEthereum commits as many blocks as needed to settle the batch on Ethereum
func (handler *EthereumHandler) SettleBatchOnEthereum() {
	batchSettleLimit, _ := handler.SafeContract.BatchSettleLimit(nil)
	for i := uint8(0); i < batchSettleLimit+1; i++ {
		handler.SimulatedChain.Commit()
	}
}

// TransferToken will transfer the amount of tokens from one address to another
func (handler *EthereumHandler) TransferToken(
	ctx context.Context,
	params TestTokenParams,
	from KeysHolder,
	to KeysHolder,
	amount *big.Int,
) {
	tkData := handler.TokensRegistry.GetTokenData(params.AbstractTokenIdentifier)

	auth, _ := bind.NewKeyedTransactorWithChainID(from.EthSK, handler.ChainID)
	tx, err := tkData.EthErc20Contract.Transfer(auth, to.EthAddress, amount)
	require.NoError(handler, err)
	handler.SimulatedChain.Commit()
	handler.checkEthTxResult(ctx, tx.Hash())
}

// GetTotalBalancesForToken will return the total locked balance for the provided token
func (handler *EthereumHandler) GetTotalBalancesForToken(ctx context.Context, address common.Address) *big.Int {
	opts := &bind.CallOpts{Context: ctx}
	balance, err := handler.SafeContract.TotalBalances(opts, address)
	require.NoError(handler, err)
	return balance
}

// GetBurnBalanceForToken will return burn balance for the provided token
func (handler *EthereumHandler) GetBurnBalanceForToken(ctx context.Context, address common.Address) *big.Int {
	opts := &bind.CallOpts{Context: ctx}
	balance, err := handler.SafeContract.BurnBalances(opts, address)
	require.NoError(handler, err)
	return balance
}

// GetMintBalanceForToken will return mint balance for the provided token
func (handler *EthereumHandler) GetMintBalanceForToken(ctx context.Context, address common.Address) *big.Int {
	opts := &bind.CallOpts{Context: ctx}
	balance, err := handler.SafeContract.MintBalances(opts, address)
	require.NoError(handler, err)
	return balance
}

// Close will close the resources allocated
func (handler *EthereumHandler) Close() error {
	return handler.SimulatedChain.Close()
}
