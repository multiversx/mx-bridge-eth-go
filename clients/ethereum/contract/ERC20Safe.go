// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package contract

import (
	"errors"
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
	_ = abi.ConvertType
)

// ERC20SafeMetaData contains all meta data concerning the ERC20Safe contract.
var ERC20SafeMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"address\",\"name\":\"target\",\"type\":\"address\"}],\"name\":\"AddressEmptyCode\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"AddressInsufficientBalance\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"FailedInnerCall\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidInitialization\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotInitializing\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"}],\"name\":\"SafeERC20FailedOperation\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousAdmin\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newAdmin\",\"type\":\"address\"}],\"name\":\"AdminRoleTransferred\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousBridge\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newBridge\",\"type\":\"address\"}],\"name\":\"BridgeTransferred\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint112\",\"name\":\"batchId\",\"type\":\"uint112\"},{\"indexed\":false,\"internalType\":\"uint112\",\"name\":\"depositNonce\",\"type\":\"uint112\"}],\"name\":\"ERC20Deposit\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint112\",\"name\":\"batchId\",\"type\":\"uint112\"},{\"indexed\":false,\"internalType\":\"uint112\",\"name\":\"depositNonce\",\"type\":\"uint112\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"callData\",\"type\":\"bytes\"}],\"name\":\"ERC20SCDeposit\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"version\",\"type\":\"uint64\"}],\"name\":\"Initialized\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"bool\",\"name\":\"isPause\",\"type\":\"bool\"}],\"name\":\"Pause\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"admin\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"batchBlockLimit\",\"outputs\":[{\"internalType\":\"uint8\",\"name\":\"\",\"type\":\"uint8\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"batchDeposits\",\"outputs\":[{\"internalType\":\"uint112\",\"name\":\"nonce\",\"type\":\"uint112\"},{\"internalType\":\"address\",\"name\":\"tokenAddress\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"depositor\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"recipient\",\"type\":\"bytes32\"},{\"internalType\":\"enumDepositStatus\",\"name\":\"status\",\"type\":\"uint8\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"batchSettleLimit\",\"outputs\":[{\"internalType\":\"uint8\",\"name\":\"\",\"type\":\"uint8\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"batchSize\",\"outputs\":[{\"internalType\":\"uint16\",\"name\":\"\",\"type\":\"uint16\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"batches\",\"outputs\":[{\"internalType\":\"uint112\",\"name\":\"nonce\",\"type\":\"uint112\"},{\"internalType\":\"uint64\",\"name\":\"blockNumber\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"lastUpdatedBlockNumber\",\"type\":\"uint64\"},{\"internalType\":\"uint16\",\"name\":\"depositsCount\",\"type\":\"uint16\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"batchesCount\",\"outputs\":[{\"internalType\":\"uint64\",\"name\":\"\",\"type\":\"uint64\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"bridge\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"burnBalances\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"tokenAddress\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"bytes32\",\"name\":\"recipientAddress\",\"type\":\"bytes32\"}],\"name\":\"deposit\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"tokenAddress\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"bytes32\",\"name\":\"recipientAddress\",\"type\":\"bytes32\"},{\"internalType\":\"bytes\",\"name\":\"callData\",\"type\":\"bytes\"}],\"name\":\"depositWithSCExecution\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"depositsCount\",\"outputs\":[{\"internalType\":\"uint64\",\"name\":\"\",\"type\":\"uint64\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"batchNonce\",\"type\":\"uint256\"}],\"name\":\"getBatch\",\"outputs\":[{\"components\":[{\"internalType\":\"uint112\",\"name\":\"nonce\",\"type\":\"uint112\"},{\"internalType\":\"uint64\",\"name\":\"blockNumber\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"lastUpdatedBlockNumber\",\"type\":\"uint64\"},{\"internalType\":\"uint16\",\"name\":\"depositsCount\",\"type\":\"uint16\"}],\"internalType\":\"structBatch\",\"name\":\"\",\"type\":\"tuple\"},{\"internalType\":\"bool\",\"name\":\"isBatchFinal\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"batchNonce\",\"type\":\"uint256\"}],\"name\":\"getDeposits\",\"outputs\":[{\"components\":[{\"internalType\":\"uint112\",\"name\":\"nonce\",\"type\":\"uint112\"},{\"internalType\":\"address\",\"name\":\"tokenAddress\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"depositor\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"recipient\",\"type\":\"bytes32\"},{\"internalType\":\"enumDepositStatus\",\"name\":\"status\",\"type\":\"uint8\"}],\"internalType\":\"structDeposit[]\",\"name\":\"\",\"type\":\"tuple[]\"},{\"internalType\":\"bool\",\"name\":\"areDepositsFinal\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"}],\"name\":\"getTokenMaxLimit\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"}],\"name\":\"getTokenMinLimit\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"tokenAddress\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"initSupply\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"tokenAddress\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"burnAmount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"mintAmount\",\"type\":\"uint256\"}],\"name\":\"initSupplyMintBurn\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"initialize\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"isAnyBatchInProgress\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"}],\"name\":\"isTokenWhitelisted\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"mintBalances\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"mintBurnTokens\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"nativeTokens\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"pause\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"paused\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"tokenAddress\",\"type\":\"address\"}],\"name\":\"recoverLostFunds\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"}],\"name\":\"removeTokenFromWhitelist\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceAdmin\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"tokenAddress\",\"type\":\"address\"}],\"name\":\"resetTotalBalance\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint8\",\"name\":\"newBatchBlockLimit\",\"type\":\"uint8\"}],\"name\":\"setBatchBlockLimit\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint8\",\"name\":\"newBatchSettleLimit\",\"type\":\"uint8\"}],\"name\":\"setBatchSettleLimit\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint16\",\"name\":\"newBatchSize\",\"type\":\"uint16\"}],\"name\":\"setBatchSize\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newBridge\",\"type\":\"address\"}],\"name\":\"setBridge\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"setTokenMaxLimit\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"setTokenMinLimit\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"tokenMaxLimits\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"tokenMinLimits\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"totalBalances\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"tokenAddress\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"recipientAddress\",\"type\":\"address\"}],\"name\":\"transfer\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newAdmin\",\"type\":\"address\"}],\"name\":\"transferAdmin\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"unpause\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"minimumAmount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"maximumAmount\",\"type\":\"uint256\"},{\"internalType\":\"bool\",\"name\":\"mintBurn\",\"type\":\"bool\"},{\"internalType\":\"bool\",\"name\":\"native\",\"type\":\"bool\"}],\"name\":\"whitelistToken\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"whitelistedTokens\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]",
}

// ERC20SafeABI is the input ABI used to generate the binding from.
// Deprecated: Use ERC20SafeMetaData.ABI instead.
var ERC20SafeABI = ERC20SafeMetaData.ABI

// ERC20Safe is an auto generated Go binding around an Ethereum contract.
type ERC20Safe struct {
	ERC20SafeCaller     // Read-only binding to the contract
	ERC20SafeTransactor // Write-only binding to the contract
	ERC20SafeFilterer   // Log filterer for contract events
}

// ERC20SafeCaller is an auto generated read-only Go binding around an Ethereum contract.
type ERC20SafeCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ERC20SafeTransactor is an auto generated write-only Go binding around an Ethereum contract.
type ERC20SafeTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ERC20SafeFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type ERC20SafeFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ERC20SafeSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ERC20SafeSession struct {
	Contract     *ERC20Safe        // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// ERC20SafeCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ERC20SafeCallerSession struct {
	Contract *ERC20SafeCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts    // Call options to use throughout this session
}

// ERC20SafeTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ERC20SafeTransactorSession struct {
	Contract     *ERC20SafeTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts    // Transaction auth options to use throughout this session
}

// ERC20SafeRaw is an auto generated low-level Go binding around an Ethereum contract.
type ERC20SafeRaw struct {
	Contract *ERC20Safe // Generic contract binding to access the raw methods on
}

// ERC20SafeCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ERC20SafeCallerRaw struct {
	Contract *ERC20SafeCaller // Generic read-only contract binding to access the raw methods on
}

// ERC20SafeTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ERC20SafeTransactorRaw struct {
	Contract *ERC20SafeTransactor // Generic write-only contract binding to access the raw methods on
}

// NewERC20Safe creates a new instance of ERC20Safe, bound to a specific deployed contract.
func NewERC20Safe(address common.Address, backend bind.ContractBackend) (*ERC20Safe, error) {
	contract, err := bindERC20Safe(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &ERC20Safe{ERC20SafeCaller: ERC20SafeCaller{contract: contract}, ERC20SafeTransactor: ERC20SafeTransactor{contract: contract}, ERC20SafeFilterer: ERC20SafeFilterer{contract: contract}}, nil
}

// NewERC20SafeCaller creates a new read-only instance of ERC20Safe, bound to a specific deployed contract.
func NewERC20SafeCaller(address common.Address, caller bind.ContractCaller) (*ERC20SafeCaller, error) {
	contract, err := bindERC20Safe(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ERC20SafeCaller{contract: contract}, nil
}

// NewERC20SafeTransactor creates a new write-only instance of ERC20Safe, bound to a specific deployed contract.
func NewERC20SafeTransactor(address common.Address, transactor bind.ContractTransactor) (*ERC20SafeTransactor, error) {
	contract, err := bindERC20Safe(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ERC20SafeTransactor{contract: contract}, nil
}

// NewERC20SafeFilterer creates a new log filterer instance of ERC20Safe, bound to a specific deployed contract.
func NewERC20SafeFilterer(address common.Address, filterer bind.ContractFilterer) (*ERC20SafeFilterer, error) {
	contract, err := bindERC20Safe(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ERC20SafeFilterer{contract: contract}, nil
}

// bindERC20Safe binds a generic wrapper to an already deployed contract.
func bindERC20Safe(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := ERC20SafeMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ERC20Safe *ERC20SafeRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ERC20Safe.Contract.ERC20SafeCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ERC20Safe *ERC20SafeRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ERC20Safe.Contract.ERC20SafeTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ERC20Safe *ERC20SafeRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ERC20Safe.Contract.ERC20SafeTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ERC20Safe *ERC20SafeCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ERC20Safe.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ERC20Safe *ERC20SafeTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ERC20Safe.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ERC20Safe *ERC20SafeTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ERC20Safe.Contract.contract.Transact(opts, method, params...)
}

// Admin is a free data retrieval call binding the contract method 0xf851a440.
//
// Solidity: function admin() view returns(address)
func (_ERC20Safe *ERC20SafeCaller) Admin(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _ERC20Safe.contract.Call(opts, &out, "admin")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Admin is a free data retrieval call binding the contract method 0xf851a440.
//
// Solidity: function admin() view returns(address)
func (_ERC20Safe *ERC20SafeSession) Admin() (common.Address, error) {
	return _ERC20Safe.Contract.Admin(&_ERC20Safe.CallOpts)
}

// Admin is a free data retrieval call binding the contract method 0xf851a440.
//
// Solidity: function admin() view returns(address)
func (_ERC20Safe *ERC20SafeCallerSession) Admin() (common.Address, error) {
	return _ERC20Safe.Contract.Admin(&_ERC20Safe.CallOpts)
}

// BatchBlockLimit is a free data retrieval call binding the contract method 0x9ab7cfaa.
//
// Solidity: function batchBlockLimit() view returns(uint8)
func (_ERC20Safe *ERC20SafeCaller) BatchBlockLimit(opts *bind.CallOpts) (uint8, error) {
	var out []interface{}
	err := _ERC20Safe.contract.Call(opts, &out, "batchBlockLimit")

	if err != nil {
		return *new(uint8), err
	}

	out0 := *abi.ConvertType(out[0], new(uint8)).(*uint8)

	return out0, err

}

// BatchBlockLimit is a free data retrieval call binding the contract method 0x9ab7cfaa.
//
// Solidity: function batchBlockLimit() view returns(uint8)
func (_ERC20Safe *ERC20SafeSession) BatchBlockLimit() (uint8, error) {
	return _ERC20Safe.Contract.BatchBlockLimit(&_ERC20Safe.CallOpts)
}

// BatchBlockLimit is a free data retrieval call binding the contract method 0x9ab7cfaa.
//
// Solidity: function batchBlockLimit() view returns(uint8)
func (_ERC20Safe *ERC20SafeCallerSession) BatchBlockLimit() (uint8, error) {
	return _ERC20Safe.Contract.BatchBlockLimit(&_ERC20Safe.CallOpts)
}

// BatchDeposits is a free data retrieval call binding the contract method 0x284c0c44.
//
// Solidity: function batchDeposits(uint256 , uint256 ) view returns(uint112 nonce, address tokenAddress, uint256 amount, address depositor, bytes32 recipient, uint8 status)
func (_ERC20Safe *ERC20SafeCaller) BatchDeposits(opts *bind.CallOpts, arg0 *big.Int, arg1 *big.Int) (struct {
	Nonce        *big.Int
	TokenAddress common.Address
	Amount       *big.Int
	Depositor    common.Address
	Recipient    [32]byte
	Status       uint8
}, error) {
	var out []interface{}
	err := _ERC20Safe.contract.Call(opts, &out, "batchDeposits", arg0, arg1)

	outstruct := new(struct {
		Nonce        *big.Int
		TokenAddress common.Address
		Amount       *big.Int
		Depositor    common.Address
		Recipient    [32]byte
		Status       uint8
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Nonce = *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	outstruct.TokenAddress = *abi.ConvertType(out[1], new(common.Address)).(*common.Address)
	outstruct.Amount = *abi.ConvertType(out[2], new(*big.Int)).(**big.Int)
	outstruct.Depositor = *abi.ConvertType(out[3], new(common.Address)).(*common.Address)
	outstruct.Recipient = *abi.ConvertType(out[4], new([32]byte)).(*[32]byte)
	outstruct.Status = *abi.ConvertType(out[5], new(uint8)).(*uint8)

	return *outstruct, err

}

// BatchDeposits is a free data retrieval call binding the contract method 0x284c0c44.
//
// Solidity: function batchDeposits(uint256 , uint256 ) view returns(uint112 nonce, address tokenAddress, uint256 amount, address depositor, bytes32 recipient, uint8 status)
func (_ERC20Safe *ERC20SafeSession) BatchDeposits(arg0 *big.Int, arg1 *big.Int) (struct {
	Nonce        *big.Int
	TokenAddress common.Address
	Amount       *big.Int
	Depositor    common.Address
	Recipient    [32]byte
	Status       uint8
}, error) {
	return _ERC20Safe.Contract.BatchDeposits(&_ERC20Safe.CallOpts, arg0, arg1)
}

// BatchDeposits is a free data retrieval call binding the contract method 0x284c0c44.
//
// Solidity: function batchDeposits(uint256 , uint256 ) view returns(uint112 nonce, address tokenAddress, uint256 amount, address depositor, bytes32 recipient, uint8 status)
func (_ERC20Safe *ERC20SafeCallerSession) BatchDeposits(arg0 *big.Int, arg1 *big.Int) (struct {
	Nonce        *big.Int
	TokenAddress common.Address
	Amount       *big.Int
	Depositor    common.Address
	Recipient    [32]byte
	Status       uint8
}, error) {
	return _ERC20Safe.Contract.BatchDeposits(&_ERC20Safe.CallOpts, arg0, arg1)
}

// BatchSettleLimit is a free data retrieval call binding the contract method 0x2325b5f7.
//
// Solidity: function batchSettleLimit() view returns(uint8)
func (_ERC20Safe *ERC20SafeCaller) BatchSettleLimit(opts *bind.CallOpts) (uint8, error) {
	var out []interface{}
	err := _ERC20Safe.contract.Call(opts, &out, "batchSettleLimit")

	if err != nil {
		return *new(uint8), err
	}

	out0 := *abi.ConvertType(out[0], new(uint8)).(*uint8)

	return out0, err

}

// BatchSettleLimit is a free data retrieval call binding the contract method 0x2325b5f7.
//
// Solidity: function batchSettleLimit() view returns(uint8)
func (_ERC20Safe *ERC20SafeSession) BatchSettleLimit() (uint8, error) {
	return _ERC20Safe.Contract.BatchSettleLimit(&_ERC20Safe.CallOpts)
}

// BatchSettleLimit is a free data retrieval call binding the contract method 0x2325b5f7.
//
// Solidity: function batchSettleLimit() view returns(uint8)
func (_ERC20Safe *ERC20SafeCallerSession) BatchSettleLimit() (uint8, error) {
	return _ERC20Safe.Contract.BatchSettleLimit(&_ERC20Safe.CallOpts)
}

// BatchSize is a free data retrieval call binding the contract method 0xf4daaba1.
//
// Solidity: function batchSize() view returns(uint16)
func (_ERC20Safe *ERC20SafeCaller) BatchSize(opts *bind.CallOpts) (uint16, error) {
	var out []interface{}
	err := _ERC20Safe.contract.Call(opts, &out, "batchSize")

	if err != nil {
		return *new(uint16), err
	}

	out0 := *abi.ConvertType(out[0], new(uint16)).(*uint16)

	return out0, err

}

// BatchSize is a free data retrieval call binding the contract method 0xf4daaba1.
//
// Solidity: function batchSize() view returns(uint16)
func (_ERC20Safe *ERC20SafeSession) BatchSize() (uint16, error) {
	return _ERC20Safe.Contract.BatchSize(&_ERC20Safe.CallOpts)
}

// BatchSize is a free data retrieval call binding the contract method 0xf4daaba1.
//
// Solidity: function batchSize() view returns(uint16)
func (_ERC20Safe *ERC20SafeCallerSession) BatchSize() (uint16, error) {
	return _ERC20Safe.Contract.BatchSize(&_ERC20Safe.CallOpts)
}

// Batches is a free data retrieval call binding the contract method 0xb32c4d8d.
//
// Solidity: function batches(uint256 ) view returns(uint112 nonce, uint64 blockNumber, uint64 lastUpdatedBlockNumber, uint16 depositsCount)
func (_ERC20Safe *ERC20SafeCaller) Batches(opts *bind.CallOpts, arg0 *big.Int) (struct {
	Nonce                  *big.Int
	BlockNumber            uint64
	LastUpdatedBlockNumber uint64
	DepositsCount          uint16
}, error) {
	var out []interface{}
	err := _ERC20Safe.contract.Call(opts, &out, "batches", arg0)

	outstruct := new(struct {
		Nonce                  *big.Int
		BlockNumber            uint64
		LastUpdatedBlockNumber uint64
		DepositsCount          uint16
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Nonce = *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	outstruct.BlockNumber = *abi.ConvertType(out[1], new(uint64)).(*uint64)
	outstruct.LastUpdatedBlockNumber = *abi.ConvertType(out[2], new(uint64)).(*uint64)
	outstruct.DepositsCount = *abi.ConvertType(out[3], new(uint16)).(*uint16)

	return *outstruct, err

}

// Batches is a free data retrieval call binding the contract method 0xb32c4d8d.
//
// Solidity: function batches(uint256 ) view returns(uint112 nonce, uint64 blockNumber, uint64 lastUpdatedBlockNumber, uint16 depositsCount)
func (_ERC20Safe *ERC20SafeSession) Batches(arg0 *big.Int) (struct {
	Nonce                  *big.Int
	BlockNumber            uint64
	LastUpdatedBlockNumber uint64
	DepositsCount          uint16
}, error) {
	return _ERC20Safe.Contract.Batches(&_ERC20Safe.CallOpts, arg0)
}

// Batches is a free data retrieval call binding the contract method 0xb32c4d8d.
//
// Solidity: function batches(uint256 ) view returns(uint112 nonce, uint64 blockNumber, uint64 lastUpdatedBlockNumber, uint16 depositsCount)
func (_ERC20Safe *ERC20SafeCallerSession) Batches(arg0 *big.Int) (struct {
	Nonce                  *big.Int
	BlockNumber            uint64
	LastUpdatedBlockNumber uint64
	DepositsCount          uint16
}, error) {
	return _ERC20Safe.Contract.Batches(&_ERC20Safe.CallOpts, arg0)
}

// BatchesCount is a free data retrieval call binding the contract method 0x87ea0961.
//
// Solidity: function batchesCount() view returns(uint64)
func (_ERC20Safe *ERC20SafeCaller) BatchesCount(opts *bind.CallOpts) (uint64, error) {
	var out []interface{}
	err := _ERC20Safe.contract.Call(opts, &out, "batchesCount")

	if err != nil {
		return *new(uint64), err
	}

	out0 := *abi.ConvertType(out[0], new(uint64)).(*uint64)

	return out0, err

}

// BatchesCount is a free data retrieval call binding the contract method 0x87ea0961.
//
// Solidity: function batchesCount() view returns(uint64)
func (_ERC20Safe *ERC20SafeSession) BatchesCount() (uint64, error) {
	return _ERC20Safe.Contract.BatchesCount(&_ERC20Safe.CallOpts)
}

// BatchesCount is a free data retrieval call binding the contract method 0x87ea0961.
//
// Solidity: function batchesCount() view returns(uint64)
func (_ERC20Safe *ERC20SafeCallerSession) BatchesCount() (uint64, error) {
	return _ERC20Safe.Contract.BatchesCount(&_ERC20Safe.CallOpts)
}

// Bridge is a free data retrieval call binding the contract method 0xe78cea92.
//
// Solidity: function bridge() view returns(address)
func (_ERC20Safe *ERC20SafeCaller) Bridge(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _ERC20Safe.contract.Call(opts, &out, "bridge")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Bridge is a free data retrieval call binding the contract method 0xe78cea92.
//
// Solidity: function bridge() view returns(address)
func (_ERC20Safe *ERC20SafeSession) Bridge() (common.Address, error) {
	return _ERC20Safe.Contract.Bridge(&_ERC20Safe.CallOpts)
}

// Bridge is a free data retrieval call binding the contract method 0xe78cea92.
//
// Solidity: function bridge() view returns(address)
func (_ERC20Safe *ERC20SafeCallerSession) Bridge() (common.Address, error) {
	return _ERC20Safe.Contract.Bridge(&_ERC20Safe.CallOpts)
}

// BurnBalances is a free data retrieval call binding the contract method 0xcf6682a2.
//
// Solidity: function burnBalances(address ) view returns(uint256)
func (_ERC20Safe *ERC20SafeCaller) BurnBalances(opts *bind.CallOpts, arg0 common.Address) (*big.Int, error) {
	var out []interface{}
	err := _ERC20Safe.contract.Call(opts, &out, "burnBalances", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// BurnBalances is a free data retrieval call binding the contract method 0xcf6682a2.
//
// Solidity: function burnBalances(address ) view returns(uint256)
func (_ERC20Safe *ERC20SafeSession) BurnBalances(arg0 common.Address) (*big.Int, error) {
	return _ERC20Safe.Contract.BurnBalances(&_ERC20Safe.CallOpts, arg0)
}

// BurnBalances is a free data retrieval call binding the contract method 0xcf6682a2.
//
// Solidity: function burnBalances(address ) view returns(uint256)
func (_ERC20Safe *ERC20SafeCallerSession) BurnBalances(arg0 common.Address) (*big.Int, error) {
	return _ERC20Safe.Contract.BurnBalances(&_ERC20Safe.CallOpts, arg0)
}

// DepositsCount is a free data retrieval call binding the contract method 0x4506e935.
//
// Solidity: function depositsCount() view returns(uint64)
func (_ERC20Safe *ERC20SafeCaller) DepositsCount(opts *bind.CallOpts) (uint64, error) {
	var out []interface{}
	err := _ERC20Safe.contract.Call(opts, &out, "depositsCount")

	if err != nil {
		return *new(uint64), err
	}

	out0 := *abi.ConvertType(out[0], new(uint64)).(*uint64)

	return out0, err

}

// DepositsCount is a free data retrieval call binding the contract method 0x4506e935.
//
// Solidity: function depositsCount() view returns(uint64)
func (_ERC20Safe *ERC20SafeSession) DepositsCount() (uint64, error) {
	return _ERC20Safe.Contract.DepositsCount(&_ERC20Safe.CallOpts)
}

// DepositsCount is a free data retrieval call binding the contract method 0x4506e935.
//
// Solidity: function depositsCount() view returns(uint64)
func (_ERC20Safe *ERC20SafeCallerSession) DepositsCount() (uint64, error) {
	return _ERC20Safe.Contract.DepositsCount(&_ERC20Safe.CallOpts)
}

// GetBatch is a free data retrieval call binding the contract method 0x5ac44282.
//
// Solidity: function getBatch(uint256 batchNonce) view returns((uint112,uint64,uint64,uint16), bool isBatchFinal)
func (_ERC20Safe *ERC20SafeCaller) GetBatch(opts *bind.CallOpts, batchNonce *big.Int) (Batch, bool, error) {
	var out []interface{}
	err := _ERC20Safe.contract.Call(opts, &out, "getBatch", batchNonce)

	if err != nil {
		return *new(Batch), *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(Batch)).(*Batch)
	out1 := *abi.ConvertType(out[1], new(bool)).(*bool)

	return out0, out1, err

}

// GetBatch is a free data retrieval call binding the contract method 0x5ac44282.
//
// Solidity: function getBatch(uint256 batchNonce) view returns((uint112,uint64,uint64,uint16), bool isBatchFinal)
func (_ERC20Safe *ERC20SafeSession) GetBatch(batchNonce *big.Int) (Batch, bool, error) {
	return _ERC20Safe.Contract.GetBatch(&_ERC20Safe.CallOpts, batchNonce)
}

// GetBatch is a free data retrieval call binding the contract method 0x5ac44282.
//
// Solidity: function getBatch(uint256 batchNonce) view returns((uint112,uint64,uint64,uint16), bool isBatchFinal)
func (_ERC20Safe *ERC20SafeCallerSession) GetBatch(batchNonce *big.Int) (Batch, bool, error) {
	return _ERC20Safe.Contract.GetBatch(&_ERC20Safe.CallOpts, batchNonce)
}

// GetDeposits is a free data retrieval call binding the contract method 0x085c967f.
//
// Solidity: function getDeposits(uint256 batchNonce) view returns((uint112,address,uint256,address,bytes32,uint8)[], bool areDepositsFinal)
func (_ERC20Safe *ERC20SafeCaller) GetDeposits(opts *bind.CallOpts, batchNonce *big.Int) ([]Deposit, bool, error) {
	var out []interface{}
	err := _ERC20Safe.contract.Call(opts, &out, "getDeposits", batchNonce)

	if err != nil {
		return *new([]Deposit), *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new([]Deposit)).(*[]Deposit)
	out1 := *abi.ConvertType(out[1], new(bool)).(*bool)

	return out0, out1, err

}

// GetDeposits is a free data retrieval call binding the contract method 0x085c967f.
//
// Solidity: function getDeposits(uint256 batchNonce) view returns((uint112,address,uint256,address,bytes32,uint8)[], bool areDepositsFinal)
func (_ERC20Safe *ERC20SafeSession) GetDeposits(batchNonce *big.Int) ([]Deposit, bool, error) {
	return _ERC20Safe.Contract.GetDeposits(&_ERC20Safe.CallOpts, batchNonce)
}

// GetDeposits is a free data retrieval call binding the contract method 0x085c967f.
//
// Solidity: function getDeposits(uint256 batchNonce) view returns((uint112,address,uint256,address,bytes32,uint8)[], bool areDepositsFinal)
func (_ERC20Safe *ERC20SafeCallerSession) GetDeposits(batchNonce *big.Int) ([]Deposit, bool, error) {
	return _ERC20Safe.Contract.GetDeposits(&_ERC20Safe.CallOpts, batchNonce)
}

// GetTokenMaxLimit is a free data retrieval call binding the contract method 0xc652a0b5.
//
// Solidity: function getTokenMaxLimit(address token) view returns(uint256)
func (_ERC20Safe *ERC20SafeCaller) GetTokenMaxLimit(opts *bind.CallOpts, token common.Address) (*big.Int, error) {
	var out []interface{}
	err := _ERC20Safe.contract.Call(opts, &out, "getTokenMaxLimit", token)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetTokenMaxLimit is a free data retrieval call binding the contract method 0xc652a0b5.
//
// Solidity: function getTokenMaxLimit(address token) view returns(uint256)
func (_ERC20Safe *ERC20SafeSession) GetTokenMaxLimit(token common.Address) (*big.Int, error) {
	return _ERC20Safe.Contract.GetTokenMaxLimit(&_ERC20Safe.CallOpts, token)
}

// GetTokenMaxLimit is a free data retrieval call binding the contract method 0xc652a0b5.
//
// Solidity: function getTokenMaxLimit(address token) view returns(uint256)
func (_ERC20Safe *ERC20SafeCallerSession) GetTokenMaxLimit(token common.Address) (*big.Int, error) {
	return _ERC20Safe.Contract.GetTokenMaxLimit(&_ERC20Safe.CallOpts, token)
}

// GetTokenMinLimit is a free data retrieval call binding the contract method 0x9f0ebb93.
//
// Solidity: function getTokenMinLimit(address token) view returns(uint256)
func (_ERC20Safe *ERC20SafeCaller) GetTokenMinLimit(opts *bind.CallOpts, token common.Address) (*big.Int, error) {
	var out []interface{}
	err := _ERC20Safe.contract.Call(opts, &out, "getTokenMinLimit", token)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetTokenMinLimit is a free data retrieval call binding the contract method 0x9f0ebb93.
//
// Solidity: function getTokenMinLimit(address token) view returns(uint256)
func (_ERC20Safe *ERC20SafeSession) GetTokenMinLimit(token common.Address) (*big.Int, error) {
	return _ERC20Safe.Contract.GetTokenMinLimit(&_ERC20Safe.CallOpts, token)
}

// GetTokenMinLimit is a free data retrieval call binding the contract method 0x9f0ebb93.
//
// Solidity: function getTokenMinLimit(address token) view returns(uint256)
func (_ERC20Safe *ERC20SafeCallerSession) GetTokenMinLimit(token common.Address) (*big.Int, error) {
	return _ERC20Safe.Contract.GetTokenMinLimit(&_ERC20Safe.CallOpts, token)
}

// IsAnyBatchInProgress is a free data retrieval call binding the contract method 0x82146138.
//
// Solidity: function isAnyBatchInProgress() view returns(bool)
func (_ERC20Safe *ERC20SafeCaller) IsAnyBatchInProgress(opts *bind.CallOpts) (bool, error) {
	var out []interface{}
	err := _ERC20Safe.contract.Call(opts, &out, "isAnyBatchInProgress")

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsAnyBatchInProgress is a free data retrieval call binding the contract method 0x82146138.
//
// Solidity: function isAnyBatchInProgress() view returns(bool)
func (_ERC20Safe *ERC20SafeSession) IsAnyBatchInProgress() (bool, error) {
	return _ERC20Safe.Contract.IsAnyBatchInProgress(&_ERC20Safe.CallOpts)
}

// IsAnyBatchInProgress is a free data retrieval call binding the contract method 0x82146138.
//
// Solidity: function isAnyBatchInProgress() view returns(bool)
func (_ERC20Safe *ERC20SafeCallerSession) IsAnyBatchInProgress() (bool, error) {
	return _ERC20Safe.Contract.IsAnyBatchInProgress(&_ERC20Safe.CallOpts)
}

// IsTokenWhitelisted is a free data retrieval call binding the contract method 0xb5af090f.
//
// Solidity: function isTokenWhitelisted(address token) view returns(bool)
func (_ERC20Safe *ERC20SafeCaller) IsTokenWhitelisted(opts *bind.CallOpts, token common.Address) (bool, error) {
	var out []interface{}
	err := _ERC20Safe.contract.Call(opts, &out, "isTokenWhitelisted", token)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsTokenWhitelisted is a free data retrieval call binding the contract method 0xb5af090f.
//
// Solidity: function isTokenWhitelisted(address token) view returns(bool)
func (_ERC20Safe *ERC20SafeSession) IsTokenWhitelisted(token common.Address) (bool, error) {
	return _ERC20Safe.Contract.IsTokenWhitelisted(&_ERC20Safe.CallOpts, token)
}

// IsTokenWhitelisted is a free data retrieval call binding the contract method 0xb5af090f.
//
// Solidity: function isTokenWhitelisted(address token) view returns(bool)
func (_ERC20Safe *ERC20SafeCallerSession) IsTokenWhitelisted(token common.Address) (bool, error) {
	return _ERC20Safe.Contract.IsTokenWhitelisted(&_ERC20Safe.CallOpts, token)
}

// MintBalances is a free data retrieval call binding the contract method 0xbc56602f.
//
// Solidity: function mintBalances(address ) view returns(uint256)
func (_ERC20Safe *ERC20SafeCaller) MintBalances(opts *bind.CallOpts, arg0 common.Address) (*big.Int, error) {
	var out []interface{}
	err := _ERC20Safe.contract.Call(opts, &out, "mintBalances", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MintBalances is a free data retrieval call binding the contract method 0xbc56602f.
//
// Solidity: function mintBalances(address ) view returns(uint256)
func (_ERC20Safe *ERC20SafeSession) MintBalances(arg0 common.Address) (*big.Int, error) {
	return _ERC20Safe.Contract.MintBalances(&_ERC20Safe.CallOpts, arg0)
}

// MintBalances is a free data retrieval call binding the contract method 0xbc56602f.
//
// Solidity: function mintBalances(address ) view returns(uint256)
func (_ERC20Safe *ERC20SafeCallerSession) MintBalances(arg0 common.Address) (*big.Int, error) {
	return _ERC20Safe.Contract.MintBalances(&_ERC20Safe.CallOpts, arg0)
}

// MintBurnTokens is a free data retrieval call binding the contract method 0x90e0cfcb.
//
// Solidity: function mintBurnTokens(address ) view returns(bool)
func (_ERC20Safe *ERC20SafeCaller) MintBurnTokens(opts *bind.CallOpts, arg0 common.Address) (bool, error) {
	var out []interface{}
	err := _ERC20Safe.contract.Call(opts, &out, "mintBurnTokens", arg0)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// MintBurnTokens is a free data retrieval call binding the contract method 0x90e0cfcb.
//
// Solidity: function mintBurnTokens(address ) view returns(bool)
func (_ERC20Safe *ERC20SafeSession) MintBurnTokens(arg0 common.Address) (bool, error) {
	return _ERC20Safe.Contract.MintBurnTokens(&_ERC20Safe.CallOpts, arg0)
}

// MintBurnTokens is a free data retrieval call binding the contract method 0x90e0cfcb.
//
// Solidity: function mintBurnTokens(address ) view returns(bool)
func (_ERC20Safe *ERC20SafeCallerSession) MintBurnTokens(arg0 common.Address) (bool, error) {
	return _ERC20Safe.Contract.MintBurnTokens(&_ERC20Safe.CallOpts, arg0)
}

// NativeTokens is a free data retrieval call binding the contract method 0xc86726f6.
//
// Solidity: function nativeTokens(address ) view returns(bool)
func (_ERC20Safe *ERC20SafeCaller) NativeTokens(opts *bind.CallOpts, arg0 common.Address) (bool, error) {
	var out []interface{}
	err := _ERC20Safe.contract.Call(opts, &out, "nativeTokens", arg0)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// NativeTokens is a free data retrieval call binding the contract method 0xc86726f6.
//
// Solidity: function nativeTokens(address ) view returns(bool)
func (_ERC20Safe *ERC20SafeSession) NativeTokens(arg0 common.Address) (bool, error) {
	return _ERC20Safe.Contract.NativeTokens(&_ERC20Safe.CallOpts, arg0)
}

// NativeTokens is a free data retrieval call binding the contract method 0xc86726f6.
//
// Solidity: function nativeTokens(address ) view returns(bool)
func (_ERC20Safe *ERC20SafeCallerSession) NativeTokens(arg0 common.Address) (bool, error) {
	return _ERC20Safe.Contract.NativeTokens(&_ERC20Safe.CallOpts, arg0)
}

// Paused is a free data retrieval call binding the contract method 0x5c975abb.
//
// Solidity: function paused() view returns(bool)
func (_ERC20Safe *ERC20SafeCaller) Paused(opts *bind.CallOpts) (bool, error) {
	var out []interface{}
	err := _ERC20Safe.contract.Call(opts, &out, "paused")

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// Paused is a free data retrieval call binding the contract method 0x5c975abb.
//
// Solidity: function paused() view returns(bool)
func (_ERC20Safe *ERC20SafeSession) Paused() (bool, error) {
	return _ERC20Safe.Contract.Paused(&_ERC20Safe.CallOpts)
}

// Paused is a free data retrieval call binding the contract method 0x5c975abb.
//
// Solidity: function paused() view returns(bool)
func (_ERC20Safe *ERC20SafeCallerSession) Paused() (bool, error) {
	return _ERC20Safe.Contract.Paused(&_ERC20Safe.CallOpts)
}

// TokenMaxLimits is a free data retrieval call binding the contract method 0xc639651d.
//
// Solidity: function tokenMaxLimits(address ) view returns(uint256)
func (_ERC20Safe *ERC20SafeCaller) TokenMaxLimits(opts *bind.CallOpts, arg0 common.Address) (*big.Int, error) {
	var out []interface{}
	err := _ERC20Safe.contract.Call(opts, &out, "tokenMaxLimits", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// TokenMaxLimits is a free data retrieval call binding the contract method 0xc639651d.
//
// Solidity: function tokenMaxLimits(address ) view returns(uint256)
func (_ERC20Safe *ERC20SafeSession) TokenMaxLimits(arg0 common.Address) (*big.Int, error) {
	return _ERC20Safe.Contract.TokenMaxLimits(&_ERC20Safe.CallOpts, arg0)
}

// TokenMaxLimits is a free data retrieval call binding the contract method 0xc639651d.
//
// Solidity: function tokenMaxLimits(address ) view returns(uint256)
func (_ERC20Safe *ERC20SafeCallerSession) TokenMaxLimits(arg0 common.Address) (*big.Int, error) {
	return _ERC20Safe.Contract.TokenMaxLimits(&_ERC20Safe.CallOpts, arg0)
}

// TokenMinLimits is a free data retrieval call binding the contract method 0xf6246ea1.
//
// Solidity: function tokenMinLimits(address ) view returns(uint256)
func (_ERC20Safe *ERC20SafeCaller) TokenMinLimits(opts *bind.CallOpts, arg0 common.Address) (*big.Int, error) {
	var out []interface{}
	err := _ERC20Safe.contract.Call(opts, &out, "tokenMinLimits", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// TokenMinLimits is a free data retrieval call binding the contract method 0xf6246ea1.
//
// Solidity: function tokenMinLimits(address ) view returns(uint256)
func (_ERC20Safe *ERC20SafeSession) TokenMinLimits(arg0 common.Address) (*big.Int, error) {
	return _ERC20Safe.Contract.TokenMinLimits(&_ERC20Safe.CallOpts, arg0)
}

// TokenMinLimits is a free data retrieval call binding the contract method 0xf6246ea1.
//
// Solidity: function tokenMinLimits(address ) view returns(uint256)
func (_ERC20Safe *ERC20SafeCallerSession) TokenMinLimits(arg0 common.Address) (*big.Int, error) {
	return _ERC20Safe.Contract.TokenMinLimits(&_ERC20Safe.CallOpts, arg0)
}

// TotalBalances is a free data retrieval call binding the contract method 0xaee9c872.
//
// Solidity: function totalBalances(address ) view returns(uint256)
func (_ERC20Safe *ERC20SafeCaller) TotalBalances(opts *bind.CallOpts, arg0 common.Address) (*big.Int, error) {
	var out []interface{}
	err := _ERC20Safe.contract.Call(opts, &out, "totalBalances", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// TotalBalances is a free data retrieval call binding the contract method 0xaee9c872.
//
// Solidity: function totalBalances(address ) view returns(uint256)
func (_ERC20Safe *ERC20SafeSession) TotalBalances(arg0 common.Address) (*big.Int, error) {
	return _ERC20Safe.Contract.TotalBalances(&_ERC20Safe.CallOpts, arg0)
}

// TotalBalances is a free data retrieval call binding the contract method 0xaee9c872.
//
// Solidity: function totalBalances(address ) view returns(uint256)
func (_ERC20Safe *ERC20SafeCallerSession) TotalBalances(arg0 common.Address) (*big.Int, error) {
	return _ERC20Safe.Contract.TotalBalances(&_ERC20Safe.CallOpts, arg0)
}

// WhitelistedTokens is a free data retrieval call binding the contract method 0xdaf9c210.
//
// Solidity: function whitelistedTokens(address ) view returns(bool)
func (_ERC20Safe *ERC20SafeCaller) WhitelistedTokens(opts *bind.CallOpts, arg0 common.Address) (bool, error) {
	var out []interface{}
	err := _ERC20Safe.contract.Call(opts, &out, "whitelistedTokens", arg0)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// WhitelistedTokens is a free data retrieval call binding the contract method 0xdaf9c210.
//
// Solidity: function whitelistedTokens(address ) view returns(bool)
func (_ERC20Safe *ERC20SafeSession) WhitelistedTokens(arg0 common.Address) (bool, error) {
	return _ERC20Safe.Contract.WhitelistedTokens(&_ERC20Safe.CallOpts, arg0)
}

// WhitelistedTokens is a free data retrieval call binding the contract method 0xdaf9c210.
//
// Solidity: function whitelistedTokens(address ) view returns(bool)
func (_ERC20Safe *ERC20SafeCallerSession) WhitelistedTokens(arg0 common.Address) (bool, error) {
	return _ERC20Safe.Contract.WhitelistedTokens(&_ERC20Safe.CallOpts, arg0)
}

// Deposit is a paid mutator transaction binding the contract method 0x26b3293f.
//
// Solidity: function deposit(address tokenAddress, uint256 amount, bytes32 recipientAddress) returns()
func (_ERC20Safe *ERC20SafeTransactor) Deposit(opts *bind.TransactOpts, tokenAddress common.Address, amount *big.Int, recipientAddress [32]byte) (*types.Transaction, error) {
	return _ERC20Safe.contract.Transact(opts, "deposit", tokenAddress, amount, recipientAddress)
}

// Deposit is a paid mutator transaction binding the contract method 0x26b3293f.
//
// Solidity: function deposit(address tokenAddress, uint256 amount, bytes32 recipientAddress) returns()
func (_ERC20Safe *ERC20SafeSession) Deposit(tokenAddress common.Address, amount *big.Int, recipientAddress [32]byte) (*types.Transaction, error) {
	return _ERC20Safe.Contract.Deposit(&_ERC20Safe.TransactOpts, tokenAddress, amount, recipientAddress)
}

// Deposit is a paid mutator transaction binding the contract method 0x26b3293f.
//
// Solidity: function deposit(address tokenAddress, uint256 amount, bytes32 recipientAddress) returns()
func (_ERC20Safe *ERC20SafeTransactorSession) Deposit(tokenAddress common.Address, amount *big.Int, recipientAddress [32]byte) (*types.Transaction, error) {
	return _ERC20Safe.Contract.Deposit(&_ERC20Safe.TransactOpts, tokenAddress, amount, recipientAddress)
}

// DepositWithSCExecution is a paid mutator transaction binding the contract method 0xc859b3fe.
//
// Solidity: function depositWithSCExecution(address tokenAddress, uint256 amount, bytes32 recipientAddress, bytes callData) returns()
func (_ERC20Safe *ERC20SafeTransactor) DepositWithSCExecution(opts *bind.TransactOpts, tokenAddress common.Address, amount *big.Int, recipientAddress [32]byte, callData []byte) (*types.Transaction, error) {
	return _ERC20Safe.contract.Transact(opts, "depositWithSCExecution", tokenAddress, amount, recipientAddress, callData)
}

// DepositWithSCExecution is a paid mutator transaction binding the contract method 0xc859b3fe.
//
// Solidity: function depositWithSCExecution(address tokenAddress, uint256 amount, bytes32 recipientAddress, bytes callData) returns()
func (_ERC20Safe *ERC20SafeSession) DepositWithSCExecution(tokenAddress common.Address, amount *big.Int, recipientAddress [32]byte, callData []byte) (*types.Transaction, error) {
	return _ERC20Safe.Contract.DepositWithSCExecution(&_ERC20Safe.TransactOpts, tokenAddress, amount, recipientAddress, callData)
}

// DepositWithSCExecution is a paid mutator transaction binding the contract method 0xc859b3fe.
//
// Solidity: function depositWithSCExecution(address tokenAddress, uint256 amount, bytes32 recipientAddress, bytes callData) returns()
func (_ERC20Safe *ERC20SafeTransactorSession) DepositWithSCExecution(tokenAddress common.Address, amount *big.Int, recipientAddress [32]byte, callData []byte) (*types.Transaction, error) {
	return _ERC20Safe.Contract.DepositWithSCExecution(&_ERC20Safe.TransactOpts, tokenAddress, amount, recipientAddress, callData)
}

// InitSupply is a paid mutator transaction binding the contract method 0x4013c89c.
//
// Solidity: function initSupply(address tokenAddress, uint256 amount) returns()
func (_ERC20Safe *ERC20SafeTransactor) InitSupply(opts *bind.TransactOpts, tokenAddress common.Address, amount *big.Int) (*types.Transaction, error) {
	return _ERC20Safe.contract.Transact(opts, "initSupply", tokenAddress, amount)
}

// InitSupply is a paid mutator transaction binding the contract method 0x4013c89c.
//
// Solidity: function initSupply(address tokenAddress, uint256 amount) returns()
func (_ERC20Safe *ERC20SafeSession) InitSupply(tokenAddress common.Address, amount *big.Int) (*types.Transaction, error) {
	return _ERC20Safe.Contract.InitSupply(&_ERC20Safe.TransactOpts, tokenAddress, amount)
}

// InitSupply is a paid mutator transaction binding the contract method 0x4013c89c.
//
// Solidity: function initSupply(address tokenAddress, uint256 amount) returns()
func (_ERC20Safe *ERC20SafeTransactorSession) InitSupply(tokenAddress common.Address, amount *big.Int) (*types.Transaction, error) {
	return _ERC20Safe.Contract.InitSupply(&_ERC20Safe.TransactOpts, tokenAddress, amount)
}

// InitSupplyMintBurn is a paid mutator transaction binding the contract method 0xe9935b4a.
//
// Solidity: function initSupplyMintBurn(address tokenAddress, uint256 burnAmount, uint256 mintAmount) returns()
func (_ERC20Safe *ERC20SafeTransactor) InitSupplyMintBurn(opts *bind.TransactOpts, tokenAddress common.Address, burnAmount *big.Int, mintAmount *big.Int) (*types.Transaction, error) {
	return _ERC20Safe.contract.Transact(opts, "initSupplyMintBurn", tokenAddress, burnAmount, mintAmount)
}

// InitSupplyMintBurn is a paid mutator transaction binding the contract method 0xe9935b4a.
//
// Solidity: function initSupplyMintBurn(address tokenAddress, uint256 burnAmount, uint256 mintAmount) returns()
func (_ERC20Safe *ERC20SafeSession) InitSupplyMintBurn(tokenAddress common.Address, burnAmount *big.Int, mintAmount *big.Int) (*types.Transaction, error) {
	return _ERC20Safe.Contract.InitSupplyMintBurn(&_ERC20Safe.TransactOpts, tokenAddress, burnAmount, mintAmount)
}

// InitSupplyMintBurn is a paid mutator transaction binding the contract method 0xe9935b4a.
//
// Solidity: function initSupplyMintBurn(address tokenAddress, uint256 burnAmount, uint256 mintAmount) returns()
func (_ERC20Safe *ERC20SafeTransactorSession) InitSupplyMintBurn(tokenAddress common.Address, burnAmount *big.Int, mintAmount *big.Int) (*types.Transaction, error) {
	return _ERC20Safe.Contract.InitSupplyMintBurn(&_ERC20Safe.TransactOpts, tokenAddress, burnAmount, mintAmount)
}

// Initialize is a paid mutator transaction binding the contract method 0x8129fc1c.
//
// Solidity: function initialize() returns()
func (_ERC20Safe *ERC20SafeTransactor) Initialize(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ERC20Safe.contract.Transact(opts, "initialize")
}

// Initialize is a paid mutator transaction binding the contract method 0x8129fc1c.
//
// Solidity: function initialize() returns()
func (_ERC20Safe *ERC20SafeSession) Initialize() (*types.Transaction, error) {
	return _ERC20Safe.Contract.Initialize(&_ERC20Safe.TransactOpts)
}

// Initialize is a paid mutator transaction binding the contract method 0x8129fc1c.
//
// Solidity: function initialize() returns()
func (_ERC20Safe *ERC20SafeTransactorSession) Initialize() (*types.Transaction, error) {
	return _ERC20Safe.Contract.Initialize(&_ERC20Safe.TransactOpts)
}

// Pause is a paid mutator transaction binding the contract method 0x8456cb59.
//
// Solidity: function pause() returns()
func (_ERC20Safe *ERC20SafeTransactor) Pause(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ERC20Safe.contract.Transact(opts, "pause")
}

// Pause is a paid mutator transaction binding the contract method 0x8456cb59.
//
// Solidity: function pause() returns()
func (_ERC20Safe *ERC20SafeSession) Pause() (*types.Transaction, error) {
	return _ERC20Safe.Contract.Pause(&_ERC20Safe.TransactOpts)
}

// Pause is a paid mutator transaction binding the contract method 0x8456cb59.
//
// Solidity: function pause() returns()
func (_ERC20Safe *ERC20SafeTransactorSession) Pause() (*types.Transaction, error) {
	return _ERC20Safe.Contract.Pause(&_ERC20Safe.TransactOpts)
}

// RecoverLostFunds is a paid mutator transaction binding the contract method 0x770be784.
//
// Solidity: function recoverLostFunds(address tokenAddress) returns()
func (_ERC20Safe *ERC20SafeTransactor) RecoverLostFunds(opts *bind.TransactOpts, tokenAddress common.Address) (*types.Transaction, error) {
	return _ERC20Safe.contract.Transact(opts, "recoverLostFunds", tokenAddress)
}

// RecoverLostFunds is a paid mutator transaction binding the contract method 0x770be784.
//
// Solidity: function recoverLostFunds(address tokenAddress) returns()
func (_ERC20Safe *ERC20SafeSession) RecoverLostFunds(tokenAddress common.Address) (*types.Transaction, error) {
	return _ERC20Safe.Contract.RecoverLostFunds(&_ERC20Safe.TransactOpts, tokenAddress)
}

// RecoverLostFunds is a paid mutator transaction binding the contract method 0x770be784.
//
// Solidity: function recoverLostFunds(address tokenAddress) returns()
func (_ERC20Safe *ERC20SafeTransactorSession) RecoverLostFunds(tokenAddress common.Address) (*types.Transaction, error) {
	return _ERC20Safe.Contract.RecoverLostFunds(&_ERC20Safe.TransactOpts, tokenAddress)
}

// RemoveTokenFromWhitelist is a paid mutator transaction binding the contract method 0x306275be.
//
// Solidity: function removeTokenFromWhitelist(address token) returns()
func (_ERC20Safe *ERC20SafeTransactor) RemoveTokenFromWhitelist(opts *bind.TransactOpts, token common.Address) (*types.Transaction, error) {
	return _ERC20Safe.contract.Transact(opts, "removeTokenFromWhitelist", token)
}

// RemoveTokenFromWhitelist is a paid mutator transaction binding the contract method 0x306275be.
//
// Solidity: function removeTokenFromWhitelist(address token) returns()
func (_ERC20Safe *ERC20SafeSession) RemoveTokenFromWhitelist(token common.Address) (*types.Transaction, error) {
	return _ERC20Safe.Contract.RemoveTokenFromWhitelist(&_ERC20Safe.TransactOpts, token)
}

// RemoveTokenFromWhitelist is a paid mutator transaction binding the contract method 0x306275be.
//
// Solidity: function removeTokenFromWhitelist(address token) returns()
func (_ERC20Safe *ERC20SafeTransactorSession) RemoveTokenFromWhitelist(token common.Address) (*types.Transaction, error) {
	return _ERC20Safe.Contract.RemoveTokenFromWhitelist(&_ERC20Safe.TransactOpts, token)
}

// RenounceAdmin is a paid mutator transaction binding the contract method 0x8bad0c0a.
//
// Solidity: function renounceAdmin() returns()
func (_ERC20Safe *ERC20SafeTransactor) RenounceAdmin(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ERC20Safe.contract.Transact(opts, "renounceAdmin")
}

// RenounceAdmin is a paid mutator transaction binding the contract method 0x8bad0c0a.
//
// Solidity: function renounceAdmin() returns()
func (_ERC20Safe *ERC20SafeSession) RenounceAdmin() (*types.Transaction, error) {
	return _ERC20Safe.Contract.RenounceAdmin(&_ERC20Safe.TransactOpts)
}

// RenounceAdmin is a paid mutator transaction binding the contract method 0x8bad0c0a.
//
// Solidity: function renounceAdmin() returns()
func (_ERC20Safe *ERC20SafeTransactorSession) RenounceAdmin() (*types.Transaction, error) {
	return _ERC20Safe.Contract.RenounceAdmin(&_ERC20Safe.TransactOpts)
}

// ResetTotalBalance is a paid mutator transaction binding the contract method 0xd2763186.
//
// Solidity: function resetTotalBalance(address tokenAddress) returns()
func (_ERC20Safe *ERC20SafeTransactor) ResetTotalBalance(opts *bind.TransactOpts, tokenAddress common.Address) (*types.Transaction, error) {
	return _ERC20Safe.contract.Transact(opts, "resetTotalBalance", tokenAddress)
}

// ResetTotalBalance is a paid mutator transaction binding the contract method 0xd2763186.
//
// Solidity: function resetTotalBalance(address tokenAddress) returns()
func (_ERC20Safe *ERC20SafeSession) ResetTotalBalance(tokenAddress common.Address) (*types.Transaction, error) {
	return _ERC20Safe.Contract.ResetTotalBalance(&_ERC20Safe.TransactOpts, tokenAddress)
}

// ResetTotalBalance is a paid mutator transaction binding the contract method 0xd2763186.
//
// Solidity: function resetTotalBalance(address tokenAddress) returns()
func (_ERC20Safe *ERC20SafeTransactorSession) ResetTotalBalance(tokenAddress common.Address) (*types.Transaction, error) {
	return _ERC20Safe.Contract.ResetTotalBalance(&_ERC20Safe.TransactOpts, tokenAddress)
}

// SetBatchBlockLimit is a paid mutator transaction binding the contract method 0xe8a70ee2.
//
// Solidity: function setBatchBlockLimit(uint8 newBatchBlockLimit) returns()
func (_ERC20Safe *ERC20SafeTransactor) SetBatchBlockLimit(opts *bind.TransactOpts, newBatchBlockLimit uint8) (*types.Transaction, error) {
	return _ERC20Safe.contract.Transact(opts, "setBatchBlockLimit", newBatchBlockLimit)
}

// SetBatchBlockLimit is a paid mutator transaction binding the contract method 0xe8a70ee2.
//
// Solidity: function setBatchBlockLimit(uint8 newBatchBlockLimit) returns()
func (_ERC20Safe *ERC20SafeSession) SetBatchBlockLimit(newBatchBlockLimit uint8) (*types.Transaction, error) {
	return _ERC20Safe.Contract.SetBatchBlockLimit(&_ERC20Safe.TransactOpts, newBatchBlockLimit)
}

// SetBatchBlockLimit is a paid mutator transaction binding the contract method 0xe8a70ee2.
//
// Solidity: function setBatchBlockLimit(uint8 newBatchBlockLimit) returns()
func (_ERC20Safe *ERC20SafeTransactorSession) SetBatchBlockLimit(newBatchBlockLimit uint8) (*types.Transaction, error) {
	return _ERC20Safe.Contract.SetBatchBlockLimit(&_ERC20Safe.TransactOpts, newBatchBlockLimit)
}

// SetBatchSettleLimit is a paid mutator transaction binding the contract method 0xf2e0ec48.
//
// Solidity: function setBatchSettleLimit(uint8 newBatchSettleLimit) returns()
func (_ERC20Safe *ERC20SafeTransactor) SetBatchSettleLimit(opts *bind.TransactOpts, newBatchSettleLimit uint8) (*types.Transaction, error) {
	return _ERC20Safe.contract.Transact(opts, "setBatchSettleLimit", newBatchSettleLimit)
}

// SetBatchSettleLimit is a paid mutator transaction binding the contract method 0xf2e0ec48.
//
// Solidity: function setBatchSettleLimit(uint8 newBatchSettleLimit) returns()
func (_ERC20Safe *ERC20SafeSession) SetBatchSettleLimit(newBatchSettleLimit uint8) (*types.Transaction, error) {
	return _ERC20Safe.Contract.SetBatchSettleLimit(&_ERC20Safe.TransactOpts, newBatchSettleLimit)
}

// SetBatchSettleLimit is a paid mutator transaction binding the contract method 0xf2e0ec48.
//
// Solidity: function setBatchSettleLimit(uint8 newBatchSettleLimit) returns()
func (_ERC20Safe *ERC20SafeTransactorSession) SetBatchSettleLimit(newBatchSettleLimit uint8) (*types.Transaction, error) {
	return _ERC20Safe.Contract.SetBatchSettleLimit(&_ERC20Safe.TransactOpts, newBatchSettleLimit)
}

// SetBatchSize is a paid mutator transaction binding the contract method 0xd4673de9.
//
// Solidity: function setBatchSize(uint16 newBatchSize) returns()
func (_ERC20Safe *ERC20SafeTransactor) SetBatchSize(opts *bind.TransactOpts, newBatchSize uint16) (*types.Transaction, error) {
	return _ERC20Safe.contract.Transact(opts, "setBatchSize", newBatchSize)
}

// SetBatchSize is a paid mutator transaction binding the contract method 0xd4673de9.
//
// Solidity: function setBatchSize(uint16 newBatchSize) returns()
func (_ERC20Safe *ERC20SafeSession) SetBatchSize(newBatchSize uint16) (*types.Transaction, error) {
	return _ERC20Safe.Contract.SetBatchSize(&_ERC20Safe.TransactOpts, newBatchSize)
}

// SetBatchSize is a paid mutator transaction binding the contract method 0xd4673de9.
//
// Solidity: function setBatchSize(uint16 newBatchSize) returns()
func (_ERC20Safe *ERC20SafeTransactorSession) SetBatchSize(newBatchSize uint16) (*types.Transaction, error) {
	return _ERC20Safe.Contract.SetBatchSize(&_ERC20Safe.TransactOpts, newBatchSize)
}

// SetBridge is a paid mutator transaction binding the contract method 0x8dd14802.
//
// Solidity: function setBridge(address newBridge) returns()
func (_ERC20Safe *ERC20SafeTransactor) SetBridge(opts *bind.TransactOpts, newBridge common.Address) (*types.Transaction, error) {
	return _ERC20Safe.contract.Transact(opts, "setBridge", newBridge)
}

// SetBridge is a paid mutator transaction binding the contract method 0x8dd14802.
//
// Solidity: function setBridge(address newBridge) returns()
func (_ERC20Safe *ERC20SafeSession) SetBridge(newBridge common.Address) (*types.Transaction, error) {
	return _ERC20Safe.Contract.SetBridge(&_ERC20Safe.TransactOpts, newBridge)
}

// SetBridge is a paid mutator transaction binding the contract method 0x8dd14802.
//
// Solidity: function setBridge(address newBridge) returns()
func (_ERC20Safe *ERC20SafeTransactorSession) SetBridge(newBridge common.Address) (*types.Transaction, error) {
	return _ERC20Safe.Contract.SetBridge(&_ERC20Safe.TransactOpts, newBridge)
}

// SetTokenMaxLimit is a paid mutator transaction binding the contract method 0x7d7763ce.
//
// Solidity: function setTokenMaxLimit(address token, uint256 amount) returns()
func (_ERC20Safe *ERC20SafeTransactor) SetTokenMaxLimit(opts *bind.TransactOpts, token common.Address, amount *big.Int) (*types.Transaction, error) {
	return _ERC20Safe.contract.Transact(opts, "setTokenMaxLimit", token, amount)
}

// SetTokenMaxLimit is a paid mutator transaction binding the contract method 0x7d7763ce.
//
// Solidity: function setTokenMaxLimit(address token, uint256 amount) returns()
func (_ERC20Safe *ERC20SafeSession) SetTokenMaxLimit(token common.Address, amount *big.Int) (*types.Transaction, error) {
	return _ERC20Safe.Contract.SetTokenMaxLimit(&_ERC20Safe.TransactOpts, token, amount)
}

// SetTokenMaxLimit is a paid mutator transaction binding the contract method 0x7d7763ce.
//
// Solidity: function setTokenMaxLimit(address token, uint256 amount) returns()
func (_ERC20Safe *ERC20SafeTransactorSession) SetTokenMaxLimit(token common.Address, amount *big.Int) (*types.Transaction, error) {
	return _ERC20Safe.Contract.SetTokenMaxLimit(&_ERC20Safe.TransactOpts, token, amount)
}

// SetTokenMinLimit is a paid mutator transaction binding the contract method 0x920b0308.
//
// Solidity: function setTokenMinLimit(address token, uint256 amount) returns()
func (_ERC20Safe *ERC20SafeTransactor) SetTokenMinLimit(opts *bind.TransactOpts, token common.Address, amount *big.Int) (*types.Transaction, error) {
	return _ERC20Safe.contract.Transact(opts, "setTokenMinLimit", token, amount)
}

// SetTokenMinLimit is a paid mutator transaction binding the contract method 0x920b0308.
//
// Solidity: function setTokenMinLimit(address token, uint256 amount) returns()
func (_ERC20Safe *ERC20SafeSession) SetTokenMinLimit(token common.Address, amount *big.Int) (*types.Transaction, error) {
	return _ERC20Safe.Contract.SetTokenMinLimit(&_ERC20Safe.TransactOpts, token, amount)
}

// SetTokenMinLimit is a paid mutator transaction binding the contract method 0x920b0308.
//
// Solidity: function setTokenMinLimit(address token, uint256 amount) returns()
func (_ERC20Safe *ERC20SafeTransactorSession) SetTokenMinLimit(token common.Address, amount *big.Int) (*types.Transaction, error) {
	return _ERC20Safe.Contract.SetTokenMinLimit(&_ERC20Safe.TransactOpts, token, amount)
}

// Transfer is a paid mutator transaction binding the contract method 0xdbba0f01.
//
// Solidity: function transfer(address tokenAddress, uint256 amount, address recipientAddress) returns(bool)
func (_ERC20Safe *ERC20SafeTransactor) Transfer(opts *bind.TransactOpts, tokenAddress common.Address, amount *big.Int, recipientAddress common.Address) (*types.Transaction, error) {
	return _ERC20Safe.contract.Transact(opts, "transfer", tokenAddress, amount, recipientAddress)
}

// Transfer is a paid mutator transaction binding the contract method 0xdbba0f01.
//
// Solidity: function transfer(address tokenAddress, uint256 amount, address recipientAddress) returns(bool)
func (_ERC20Safe *ERC20SafeSession) Transfer(tokenAddress common.Address, amount *big.Int, recipientAddress common.Address) (*types.Transaction, error) {
	return _ERC20Safe.Contract.Transfer(&_ERC20Safe.TransactOpts, tokenAddress, amount, recipientAddress)
}

// Transfer is a paid mutator transaction binding the contract method 0xdbba0f01.
//
// Solidity: function transfer(address tokenAddress, uint256 amount, address recipientAddress) returns(bool)
func (_ERC20Safe *ERC20SafeTransactorSession) Transfer(tokenAddress common.Address, amount *big.Int, recipientAddress common.Address) (*types.Transaction, error) {
	return _ERC20Safe.Contract.Transfer(&_ERC20Safe.TransactOpts, tokenAddress, amount, recipientAddress)
}

// TransferAdmin is a paid mutator transaction binding the contract method 0x75829def.
//
// Solidity: function transferAdmin(address newAdmin) returns()
func (_ERC20Safe *ERC20SafeTransactor) TransferAdmin(opts *bind.TransactOpts, newAdmin common.Address) (*types.Transaction, error) {
	return _ERC20Safe.contract.Transact(opts, "transferAdmin", newAdmin)
}

// TransferAdmin is a paid mutator transaction binding the contract method 0x75829def.
//
// Solidity: function transferAdmin(address newAdmin) returns()
func (_ERC20Safe *ERC20SafeSession) TransferAdmin(newAdmin common.Address) (*types.Transaction, error) {
	return _ERC20Safe.Contract.TransferAdmin(&_ERC20Safe.TransactOpts, newAdmin)
}

// TransferAdmin is a paid mutator transaction binding the contract method 0x75829def.
//
// Solidity: function transferAdmin(address newAdmin) returns()
func (_ERC20Safe *ERC20SafeTransactorSession) TransferAdmin(newAdmin common.Address) (*types.Transaction, error) {
	return _ERC20Safe.Contract.TransferAdmin(&_ERC20Safe.TransactOpts, newAdmin)
}

// Unpause is a paid mutator transaction binding the contract method 0x3f4ba83a.
//
// Solidity: function unpause() returns()
func (_ERC20Safe *ERC20SafeTransactor) Unpause(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ERC20Safe.contract.Transact(opts, "unpause")
}

// Unpause is a paid mutator transaction binding the contract method 0x3f4ba83a.
//
// Solidity: function unpause() returns()
func (_ERC20Safe *ERC20SafeSession) Unpause() (*types.Transaction, error) {
	return _ERC20Safe.Contract.Unpause(&_ERC20Safe.TransactOpts)
}

// Unpause is a paid mutator transaction binding the contract method 0x3f4ba83a.
//
// Solidity: function unpause() returns()
func (_ERC20Safe *ERC20SafeTransactorSession) Unpause() (*types.Transaction, error) {
	return _ERC20Safe.Contract.Unpause(&_ERC20Safe.TransactOpts)
}

// WhitelistToken is a paid mutator transaction binding the contract method 0xa7c3a06f.
//
// Solidity: function whitelistToken(address token, uint256 minimumAmount, uint256 maximumAmount, bool mintBurn, bool native) returns()
func (_ERC20Safe *ERC20SafeTransactor) WhitelistToken(opts *bind.TransactOpts, token common.Address, minimumAmount *big.Int, maximumAmount *big.Int, mintBurn bool, native bool) (*types.Transaction, error) {
	return _ERC20Safe.contract.Transact(opts, "whitelistToken", token, minimumAmount, maximumAmount, mintBurn, native)
}

// WhitelistToken is a paid mutator transaction binding the contract method 0xa7c3a06f.
//
// Solidity: function whitelistToken(address token, uint256 minimumAmount, uint256 maximumAmount, bool mintBurn, bool native) returns()
func (_ERC20Safe *ERC20SafeSession) WhitelistToken(token common.Address, minimumAmount *big.Int, maximumAmount *big.Int, mintBurn bool, native bool) (*types.Transaction, error) {
	return _ERC20Safe.Contract.WhitelistToken(&_ERC20Safe.TransactOpts, token, minimumAmount, maximumAmount, mintBurn, native)
}

// WhitelistToken is a paid mutator transaction binding the contract method 0xa7c3a06f.
//
// Solidity: function whitelistToken(address token, uint256 minimumAmount, uint256 maximumAmount, bool mintBurn, bool native) returns()
func (_ERC20Safe *ERC20SafeTransactorSession) WhitelistToken(token common.Address, minimumAmount *big.Int, maximumAmount *big.Int, mintBurn bool, native bool) (*types.Transaction, error) {
	return _ERC20Safe.Contract.WhitelistToken(&_ERC20Safe.TransactOpts, token, minimumAmount, maximumAmount, mintBurn, native)
}

// ERC20SafeAdminRoleTransferredIterator is returned from FilterAdminRoleTransferred and is used to iterate over the raw logs and unpacked data for AdminRoleTransferred events raised by the ERC20Safe contract.
type ERC20SafeAdminRoleTransferredIterator struct {
	Event *ERC20SafeAdminRoleTransferred // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ERC20SafeAdminRoleTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ERC20SafeAdminRoleTransferred)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ERC20SafeAdminRoleTransferred)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ERC20SafeAdminRoleTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ERC20SafeAdminRoleTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ERC20SafeAdminRoleTransferred represents a AdminRoleTransferred event raised by the ERC20Safe contract.
type ERC20SafeAdminRoleTransferred struct {
	PreviousAdmin common.Address
	NewAdmin      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterAdminRoleTransferred is a free log retrieval operation binding the contract event 0xe379ac64de02d8184ca1a871ac486cb8137de77e485ede140e97057b9c765ffd.
//
// Solidity: event AdminRoleTransferred(address indexed previousAdmin, address indexed newAdmin)
func (_ERC20Safe *ERC20SafeFilterer) FilterAdminRoleTransferred(opts *bind.FilterOpts, previousAdmin []common.Address, newAdmin []common.Address) (*ERC20SafeAdminRoleTransferredIterator, error) {

	var previousAdminRule []interface{}
	for _, previousAdminItem := range previousAdmin {
		previousAdminRule = append(previousAdminRule, previousAdminItem)
	}
	var newAdminRule []interface{}
	for _, newAdminItem := range newAdmin {
		newAdminRule = append(newAdminRule, newAdminItem)
	}

	logs, sub, err := _ERC20Safe.contract.FilterLogs(opts, "AdminRoleTransferred", previousAdminRule, newAdminRule)
	if err != nil {
		return nil, err
	}
	return &ERC20SafeAdminRoleTransferredIterator{contract: _ERC20Safe.contract, event: "AdminRoleTransferred", logs: logs, sub: sub}, nil
}

// WatchAdminRoleTransferred is a free log subscription operation binding the contract event 0xe379ac64de02d8184ca1a871ac486cb8137de77e485ede140e97057b9c765ffd.
//
// Solidity: event AdminRoleTransferred(address indexed previousAdmin, address indexed newAdmin)
func (_ERC20Safe *ERC20SafeFilterer) WatchAdminRoleTransferred(opts *bind.WatchOpts, sink chan<- *ERC20SafeAdminRoleTransferred, previousAdmin []common.Address, newAdmin []common.Address) (event.Subscription, error) {

	var previousAdminRule []interface{}
	for _, previousAdminItem := range previousAdmin {
		previousAdminRule = append(previousAdminRule, previousAdminItem)
	}
	var newAdminRule []interface{}
	for _, newAdminItem := range newAdmin {
		newAdminRule = append(newAdminRule, newAdminItem)
	}

	logs, sub, err := _ERC20Safe.contract.WatchLogs(opts, "AdminRoleTransferred", previousAdminRule, newAdminRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ERC20SafeAdminRoleTransferred)
				if err := _ERC20Safe.contract.UnpackLog(event, "AdminRoleTransferred", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseAdminRoleTransferred is a log parse operation binding the contract event 0xe379ac64de02d8184ca1a871ac486cb8137de77e485ede140e97057b9c765ffd.
//
// Solidity: event AdminRoleTransferred(address indexed previousAdmin, address indexed newAdmin)
func (_ERC20Safe *ERC20SafeFilterer) ParseAdminRoleTransferred(log types.Log) (*ERC20SafeAdminRoleTransferred, error) {
	event := new(ERC20SafeAdminRoleTransferred)
	if err := _ERC20Safe.contract.UnpackLog(event, "AdminRoleTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ERC20SafeBridgeTransferredIterator is returned from FilterBridgeTransferred and is used to iterate over the raw logs and unpacked data for BridgeTransferred events raised by the ERC20Safe contract.
type ERC20SafeBridgeTransferredIterator struct {
	Event *ERC20SafeBridgeTransferred // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ERC20SafeBridgeTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ERC20SafeBridgeTransferred)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ERC20SafeBridgeTransferred)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ERC20SafeBridgeTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ERC20SafeBridgeTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ERC20SafeBridgeTransferred represents a BridgeTransferred event raised by the ERC20Safe contract.
type ERC20SafeBridgeTransferred struct {
	PreviousBridge common.Address
	NewBridge      common.Address
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterBridgeTransferred is a free log retrieval operation binding the contract event 0xcca5fddab921a878ddbd4edb737a2cf3ac6df70864f108606647d1b37a5e07a0.
//
// Solidity: event BridgeTransferred(address indexed previousBridge, address indexed newBridge)
func (_ERC20Safe *ERC20SafeFilterer) FilterBridgeTransferred(opts *bind.FilterOpts, previousBridge []common.Address, newBridge []common.Address) (*ERC20SafeBridgeTransferredIterator, error) {

	var previousBridgeRule []interface{}
	for _, previousBridgeItem := range previousBridge {
		previousBridgeRule = append(previousBridgeRule, previousBridgeItem)
	}
	var newBridgeRule []interface{}
	for _, newBridgeItem := range newBridge {
		newBridgeRule = append(newBridgeRule, newBridgeItem)
	}

	logs, sub, err := _ERC20Safe.contract.FilterLogs(opts, "BridgeTransferred", previousBridgeRule, newBridgeRule)
	if err != nil {
		return nil, err
	}
	return &ERC20SafeBridgeTransferredIterator{contract: _ERC20Safe.contract, event: "BridgeTransferred", logs: logs, sub: sub}, nil
}

// WatchBridgeTransferred is a free log subscription operation binding the contract event 0xcca5fddab921a878ddbd4edb737a2cf3ac6df70864f108606647d1b37a5e07a0.
//
// Solidity: event BridgeTransferred(address indexed previousBridge, address indexed newBridge)
func (_ERC20Safe *ERC20SafeFilterer) WatchBridgeTransferred(opts *bind.WatchOpts, sink chan<- *ERC20SafeBridgeTransferred, previousBridge []common.Address, newBridge []common.Address) (event.Subscription, error) {

	var previousBridgeRule []interface{}
	for _, previousBridgeItem := range previousBridge {
		previousBridgeRule = append(previousBridgeRule, previousBridgeItem)
	}
	var newBridgeRule []interface{}
	for _, newBridgeItem := range newBridge {
		newBridgeRule = append(newBridgeRule, newBridgeItem)
	}

	logs, sub, err := _ERC20Safe.contract.WatchLogs(opts, "BridgeTransferred", previousBridgeRule, newBridgeRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ERC20SafeBridgeTransferred)
				if err := _ERC20Safe.contract.UnpackLog(event, "BridgeTransferred", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseBridgeTransferred is a log parse operation binding the contract event 0xcca5fddab921a878ddbd4edb737a2cf3ac6df70864f108606647d1b37a5e07a0.
//
// Solidity: event BridgeTransferred(address indexed previousBridge, address indexed newBridge)
func (_ERC20Safe *ERC20SafeFilterer) ParseBridgeTransferred(log types.Log) (*ERC20SafeBridgeTransferred, error) {
	event := new(ERC20SafeBridgeTransferred)
	if err := _ERC20Safe.contract.UnpackLog(event, "BridgeTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ERC20SafeERC20DepositIterator is returned from FilterERC20Deposit and is used to iterate over the raw logs and unpacked data for ERC20Deposit events raised by the ERC20Safe contract.
type ERC20SafeERC20DepositIterator struct {
	Event *ERC20SafeERC20Deposit // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ERC20SafeERC20DepositIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ERC20SafeERC20Deposit)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ERC20SafeERC20Deposit)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ERC20SafeERC20DepositIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ERC20SafeERC20DepositIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ERC20SafeERC20Deposit represents a ERC20Deposit event raised by the ERC20Safe contract.
type ERC20SafeERC20Deposit struct {
	BatchId      *big.Int
	DepositNonce *big.Int
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterERC20Deposit is a free log retrieval operation binding the contract event 0x6c15ce44793c685a79cde26a0bd5419ef4f3a337991f156be7b365962001b4a7.
//
// Solidity: event ERC20Deposit(uint112 batchId, uint112 depositNonce)
func (_ERC20Safe *ERC20SafeFilterer) FilterERC20Deposit(opts *bind.FilterOpts) (*ERC20SafeERC20DepositIterator, error) {

	logs, sub, err := _ERC20Safe.contract.FilterLogs(opts, "ERC20Deposit")
	if err != nil {
		return nil, err
	}
	return &ERC20SafeERC20DepositIterator{contract: _ERC20Safe.contract, event: "ERC20Deposit", logs: logs, sub: sub}, nil
}

// WatchERC20Deposit is a free log subscription operation binding the contract event 0x6c15ce44793c685a79cde26a0bd5419ef4f3a337991f156be7b365962001b4a7.
//
// Solidity: event ERC20Deposit(uint112 batchId, uint112 depositNonce)
func (_ERC20Safe *ERC20SafeFilterer) WatchERC20Deposit(opts *bind.WatchOpts, sink chan<- *ERC20SafeERC20Deposit) (event.Subscription, error) {

	logs, sub, err := _ERC20Safe.contract.WatchLogs(opts, "ERC20Deposit")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ERC20SafeERC20Deposit)
				if err := _ERC20Safe.contract.UnpackLog(event, "ERC20Deposit", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseERC20Deposit is a log parse operation binding the contract event 0x6c15ce44793c685a79cde26a0bd5419ef4f3a337991f156be7b365962001b4a7.
//
// Solidity: event ERC20Deposit(uint112 batchId, uint112 depositNonce)
func (_ERC20Safe *ERC20SafeFilterer) ParseERC20Deposit(log types.Log) (*ERC20SafeERC20Deposit, error) {
	event := new(ERC20SafeERC20Deposit)
	if err := _ERC20Safe.contract.UnpackLog(event, "ERC20Deposit", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ERC20SafeERC20SCDepositIterator is returned from FilterERC20SCDeposit and is used to iterate over the raw logs and unpacked data for ERC20SCDeposit events raised by the ERC20Safe contract.
type ERC20SafeERC20SCDepositIterator struct {
	Event *ERC20SafeERC20SCDeposit // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ERC20SafeERC20SCDepositIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ERC20SafeERC20SCDeposit)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ERC20SafeERC20SCDeposit)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ERC20SafeERC20SCDepositIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ERC20SafeERC20SCDepositIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ERC20SafeERC20SCDeposit represents a ERC20SCDeposit event raised by the ERC20Safe contract.
type ERC20SafeERC20SCDeposit struct {
	BatchId      *big.Int
	DepositNonce *big.Int
	CallData     []byte
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterERC20SCDeposit is a free log retrieval operation binding the contract event 0xce848da21487607afba5c5a500c2ad1002d9b8db97ca5512671309df071113b2.
//
// Solidity: event ERC20SCDeposit(uint112 indexed batchId, uint112 depositNonce, bytes callData)
func (_ERC20Safe *ERC20SafeFilterer) FilterERC20SCDeposit(opts *bind.FilterOpts, batchId []*big.Int) (*ERC20SafeERC20SCDepositIterator, error) {

	var batchIdRule []interface{}
	for _, batchIdItem := range batchId {
		batchIdRule = append(batchIdRule, batchIdItem)
	}

	logs, sub, err := _ERC20Safe.contract.FilterLogs(opts, "ERC20SCDeposit", batchIdRule)
	if err != nil {
		return nil, err
	}
	return &ERC20SafeERC20SCDepositIterator{contract: _ERC20Safe.contract, event: "ERC20SCDeposit", logs: logs, sub: sub}, nil
}

// WatchERC20SCDeposit is a free log subscription operation binding the contract event 0xce848da21487607afba5c5a500c2ad1002d9b8db97ca5512671309df071113b2.
//
// Solidity: event ERC20SCDeposit(uint112 indexed batchId, uint112 depositNonce, bytes callData)
func (_ERC20Safe *ERC20SafeFilterer) WatchERC20SCDeposit(opts *bind.WatchOpts, sink chan<- *ERC20SafeERC20SCDeposit, batchId []*big.Int) (event.Subscription, error) {

	var batchIdRule []interface{}
	for _, batchIdItem := range batchId {
		batchIdRule = append(batchIdRule, batchIdItem)
	}

	logs, sub, err := _ERC20Safe.contract.WatchLogs(opts, "ERC20SCDeposit", batchIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ERC20SafeERC20SCDeposit)
				if err := _ERC20Safe.contract.UnpackLog(event, "ERC20SCDeposit", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseERC20SCDeposit is a log parse operation binding the contract event 0xce848da21487607afba5c5a500c2ad1002d9b8db97ca5512671309df071113b2.
//
// Solidity: event ERC20SCDeposit(uint112 indexed batchId, uint112 depositNonce, bytes callData)
func (_ERC20Safe *ERC20SafeFilterer) ParseERC20SCDeposit(log types.Log) (*ERC20SafeERC20SCDeposit, error) {
	event := new(ERC20SafeERC20SCDeposit)
	if err := _ERC20Safe.contract.UnpackLog(event, "ERC20SCDeposit", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ERC20SafeInitializedIterator is returned from FilterInitialized and is used to iterate over the raw logs and unpacked data for Initialized events raised by the ERC20Safe contract.
type ERC20SafeInitializedIterator struct {
	Event *ERC20SafeInitialized // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ERC20SafeInitializedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ERC20SafeInitialized)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ERC20SafeInitialized)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ERC20SafeInitializedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ERC20SafeInitializedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ERC20SafeInitialized represents a Initialized event raised by the ERC20Safe contract.
type ERC20SafeInitialized struct {
	Version uint64
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterInitialized is a free log retrieval operation binding the contract event 0xc7f505b2f371ae2175ee4913f4499e1f2633a7b5936321eed1cdaeb6115181d2.
//
// Solidity: event Initialized(uint64 version)
func (_ERC20Safe *ERC20SafeFilterer) FilterInitialized(opts *bind.FilterOpts) (*ERC20SafeInitializedIterator, error) {

	logs, sub, err := _ERC20Safe.contract.FilterLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return &ERC20SafeInitializedIterator{contract: _ERC20Safe.contract, event: "Initialized", logs: logs, sub: sub}, nil
}

// WatchInitialized is a free log subscription operation binding the contract event 0xc7f505b2f371ae2175ee4913f4499e1f2633a7b5936321eed1cdaeb6115181d2.
//
// Solidity: event Initialized(uint64 version)
func (_ERC20Safe *ERC20SafeFilterer) WatchInitialized(opts *bind.WatchOpts, sink chan<- *ERC20SafeInitialized) (event.Subscription, error) {

	logs, sub, err := _ERC20Safe.contract.WatchLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ERC20SafeInitialized)
				if err := _ERC20Safe.contract.UnpackLog(event, "Initialized", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseInitialized is a log parse operation binding the contract event 0xc7f505b2f371ae2175ee4913f4499e1f2633a7b5936321eed1cdaeb6115181d2.
//
// Solidity: event Initialized(uint64 version)
func (_ERC20Safe *ERC20SafeFilterer) ParseInitialized(log types.Log) (*ERC20SafeInitialized, error) {
	event := new(ERC20SafeInitialized)
	if err := _ERC20Safe.contract.UnpackLog(event, "Initialized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ERC20SafePauseIterator is returned from FilterPause and is used to iterate over the raw logs and unpacked data for Pause events raised by the ERC20Safe contract.
type ERC20SafePauseIterator struct {
	Event *ERC20SafePause // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ERC20SafePauseIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ERC20SafePause)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ERC20SafePause)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ERC20SafePauseIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ERC20SafePauseIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ERC20SafePause represents a Pause event raised by the ERC20Safe contract.
type ERC20SafePause struct {
	IsPause bool
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterPause is a free log retrieval operation binding the contract event 0x9422424b175dda897495a07b091ef74a3ef715cf6d866fc972954c1c7f459304.
//
// Solidity: event Pause(bool isPause)
func (_ERC20Safe *ERC20SafeFilterer) FilterPause(opts *bind.FilterOpts) (*ERC20SafePauseIterator, error) {

	logs, sub, err := _ERC20Safe.contract.FilterLogs(opts, "Pause")
	if err != nil {
		return nil, err
	}
	return &ERC20SafePauseIterator{contract: _ERC20Safe.contract, event: "Pause", logs: logs, sub: sub}, nil
}

// WatchPause is a free log subscription operation binding the contract event 0x9422424b175dda897495a07b091ef74a3ef715cf6d866fc972954c1c7f459304.
//
// Solidity: event Pause(bool isPause)
func (_ERC20Safe *ERC20SafeFilterer) WatchPause(opts *bind.WatchOpts, sink chan<- *ERC20SafePause) (event.Subscription, error) {

	logs, sub, err := _ERC20Safe.contract.WatchLogs(opts, "Pause")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ERC20SafePause)
				if err := _ERC20Safe.contract.UnpackLog(event, "Pause", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParsePause is a log parse operation binding the contract event 0x9422424b175dda897495a07b091ef74a3ef715cf6d866fc972954c1c7f459304.
//
// Solidity: event Pause(bool isPause)
func (_ERC20Safe *ERC20SafeFilterer) ParsePause(log types.Log) (*ERC20SafePause, error) {
	event := new(ERC20SafePause)
	if err := _ERC20Safe.contract.UnpackLog(event, "Pause", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
