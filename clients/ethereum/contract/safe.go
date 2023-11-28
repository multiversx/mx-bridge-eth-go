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

// ContractMetaData contains all meta data concerning the Contract contract.
var ContractMetaData = &bind.MetaData{
	ABI: "[{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousAdmin\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newAdmin\",\"type\":\"address\"}],\"name\":\"AdminRoleTransferred\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousBridge\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newBridge\",\"type\":\"address\"}],\"name\":\"BridgeTransferred\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint112\",\"name\":\"depositNonce\",\"type\":\"uint112\"},{\"indexed\":false,\"internalType\":\"uint112\",\"name\":\"batchId\",\"type\":\"uint112\"}],\"name\":\"ERC20Deposit\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"admin\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"batchBlockLimit\",\"outputs\":[{\"internalType\":\"uint8\",\"name\":\"\",\"type\":\"uint8\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"batchDeposits\",\"outputs\":[{\"internalType\":\"uint112\",\"name\":\"nonce\",\"type\":\"uint112\"},{\"internalType\":\"address\",\"name\":\"tokenAddress\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"depositor\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"recipient\",\"type\":\"bytes32\"},{\"internalType\":\"enumDepositStatus\",\"name\":\"status\",\"type\":\"uint8\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"batchSettleLimit\",\"outputs\":[{\"internalType\":\"uint8\",\"name\":\"\",\"type\":\"uint8\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"batchSize\",\"outputs\":[{\"internalType\":\"uint16\",\"name\":\"\",\"type\":\"uint16\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"batches\",\"outputs\":[{\"internalType\":\"uint112\",\"name\":\"nonce\",\"type\":\"uint112\"},{\"internalType\":\"uint64\",\"name\":\"blockNumber\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"lastUpdatedBlockNumber\",\"type\":\"uint64\"},{\"internalType\":\"uint16\",\"name\":\"depositsCount\",\"type\":\"uint16\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"batchesCount\",\"outputs\":[{\"internalType\":\"uint64\",\"name\":\"\",\"type\":\"uint64\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"bridge\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"tokenAddress\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"bytes32\",\"name\":\"recipientAddress\",\"type\":\"bytes32\"}],\"name\":\"deposit\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"depositsCount\",\"outputs\":[{\"internalType\":\"uint64\",\"name\":\"\",\"type\":\"uint64\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"batchNonce\",\"type\":\"uint256\"}],\"name\":\"getBatch\",\"outputs\":[{\"components\":[{\"internalType\":\"uint112\",\"name\":\"nonce\",\"type\":\"uint112\"},{\"internalType\":\"uint64\",\"name\":\"blockNumber\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"lastUpdatedBlockNumber\",\"type\":\"uint64\"},{\"internalType\":\"uint16\",\"name\":\"depositsCount\",\"type\":\"uint16\"}],\"internalType\":\"structBatch\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"batchNonce\",\"type\":\"uint256\"}],\"name\":\"getDeposits\",\"outputs\":[{\"components\":[{\"internalType\":\"uint112\",\"name\":\"nonce\",\"type\":\"uint112\"},{\"internalType\":\"address\",\"name\":\"tokenAddress\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"depositor\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"recipient\",\"type\":\"bytes32\"},{\"internalType\":\"enumDepositStatus\",\"name\":\"status\",\"type\":\"uint8\"}],\"internalType\":\"structDeposit[]\",\"name\":\"\",\"type\":\"tuple[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"}],\"name\":\"getTokenMaxLimit\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"}],\"name\":\"getTokenMinLimit\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"tokenAddress\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"initSupply\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"isAnyBatchInProgress\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"}],\"name\":\"isTokenWhitelisted\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"pause\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"paused\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"tokenAddress\",\"type\":\"address\"}],\"name\":\"recoverLostFunds\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"}],\"name\":\"removeTokenFromWhitelist\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceAdmin\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint8\",\"name\":\"newBatchBlockLimit\",\"type\":\"uint8\"}],\"name\":\"setBatchBlockLimit\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint8\",\"name\":\"newBatchSettleLimit\",\"type\":\"uint8\"}],\"name\":\"setBatchSettleLimit\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint16\",\"name\":\"newBatchSize\",\"type\":\"uint16\"}],\"name\":\"setBatchSize\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newBridge\",\"type\":\"address\"}],\"name\":\"setBridge\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"setTokenMaxLimit\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"setTokenMinLimit\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"tokenBalances\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"tokenMaxLimits\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"tokenMinLimits\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"tokenMintedBalances\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"tokenAddress\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"recipientAddress\",\"type\":\"address\"}],\"name\":\"transfer\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newAdmin\",\"type\":\"address\"}],\"name\":\"transferAdmin\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"unpause\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"minimumAmount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"maximumAmount\",\"type\":\"uint256\"},{\"internalType\":\"bool\",\"name\":\"mintBurn\",\"type\":\"bool\"}],\"name\":\"whitelistToken\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"whitelistedTokens\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"whitelistedTokensMintBurn\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]",
}

// ContractABI is the input ABI used to generate the binding from.
// Deprecated: Use ContractMetaData.ABI instead.
var ContractABI = ContractMetaData.ABI

// Contract is an auto generated Go binding around an Ethereum contract.
type Contract struct {
	ContractCaller     // Read-only binding to the contract
	ContractTransactor // Write-only binding to the contract
	ContractFilterer   // Log filterer for contract events
}

// ContractCaller is an auto generated read-only Go binding around an Ethereum contract.
type ContractCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ContractTransactor is an auto generated write-only Go binding around an Ethereum contract.
type ContractTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ContractFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type ContractFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ContractSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ContractSession struct {
	Contract     *Contract         // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// ContractCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ContractCallerSession struct {
	Contract *ContractCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts   // Call options to use throughout this session
}

// ContractTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ContractTransactorSession struct {
	Contract     *ContractTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts   // Transaction auth options to use throughout this session
}

// ContractRaw is an auto generated low-level Go binding around an Ethereum contract.
type ContractRaw struct {
	Contract *Contract // Generic contract binding to access the raw methods on
}

// ContractCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ContractCallerRaw struct {
	Contract *ContractCaller // Generic read-only contract binding to access the raw methods on
}

// ContractTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ContractTransactorRaw struct {
	Contract *ContractTransactor // Generic write-only contract binding to access the raw methods on
}

// NewContract creates a new instance of Contract, bound to a specific deployed contract.
func NewContract(address common.Address, backend bind.ContractBackend) (*Contract, error) {
	contract, err := bindContract(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Contract{ContractCaller: ContractCaller{contract: contract}, ContractTransactor: ContractTransactor{contract: contract}, ContractFilterer: ContractFilterer{contract: contract}}, nil
}

// NewContractCaller creates a new read-only instance of Contract, bound to a specific deployed contract.
func NewContractCaller(address common.Address, caller bind.ContractCaller) (*ContractCaller, error) {
	contract, err := bindContract(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ContractCaller{contract: contract}, nil
}

// NewContractTransactor creates a new write-only instance of Contract, bound to a specific deployed contract.
func NewContractTransactor(address common.Address, transactor bind.ContractTransactor) (*ContractTransactor, error) {
	contract, err := bindContract(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ContractTransactor{contract: contract}, nil
}

// NewContractFilterer creates a new log filterer instance of Contract, bound to a specific deployed contract.
func NewContractFilterer(address common.Address, filterer bind.ContractFilterer) (*ContractFilterer, error) {
	contract, err := bindContract(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ContractFilterer{contract: contract}, nil
}

// bindContract binds a generic wrapper to an already deployed contract.
func bindContract(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := ContractMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Contract *ContractRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Contract.Contract.ContractCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Contract *ContractRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Contract.Contract.ContractTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Contract *ContractRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Contract.Contract.ContractTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Contract *ContractCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Contract.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Contract *ContractTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Contract.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Contract *ContractTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Contract.Contract.contract.Transact(opts, method, params...)
}

// Admin is a free data retrieval call binding the contract method 0xf851a440.
//
// Solidity: function admin() view returns(address)
func (_Contract *ContractCaller) Admin(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "admin")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Admin is a free data retrieval call binding the contract method 0xf851a440.
//
// Solidity: function admin() view returns(address)
func (_Contract *ContractSession) Admin() (common.Address, error) {
	return _Contract.Contract.Admin(&_Contract.CallOpts)
}

// Admin is a free data retrieval call binding the contract method 0xf851a440.
//
// Solidity: function admin() view returns(address)
func (_Contract *ContractCallerSession) Admin() (common.Address, error) {
	return _Contract.Contract.Admin(&_Contract.CallOpts)
}

// BatchBlockLimit is a free data retrieval call binding the contract method 0x9ab7cfaa.
//
// Solidity: function batchBlockLimit() view returns(uint8)
func (_Contract *ContractCaller) BatchBlockLimit(opts *bind.CallOpts) (uint8, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "batchBlockLimit")

	if err != nil {
		return *new(uint8), err
	}

	out0 := *abi.ConvertType(out[0], new(uint8)).(*uint8)

	return out0, err

}

// BatchBlockLimit is a free data retrieval call binding the contract method 0x9ab7cfaa.
//
// Solidity: function batchBlockLimit() view returns(uint8)
func (_Contract *ContractSession) BatchBlockLimit() (uint8, error) {
	return _Contract.Contract.BatchBlockLimit(&_Contract.CallOpts)
}

// BatchBlockLimit is a free data retrieval call binding the contract method 0x9ab7cfaa.
//
// Solidity: function batchBlockLimit() view returns(uint8)
func (_Contract *ContractCallerSession) BatchBlockLimit() (uint8, error) {
	return _Contract.Contract.BatchBlockLimit(&_Contract.CallOpts)
}

// BatchDeposits is a free data retrieval call binding the contract method 0x284c0c44.
//
// Solidity: function batchDeposits(uint256 , uint256 ) view returns(uint112 nonce, address tokenAddress, uint256 amount, address depositor, bytes32 recipient, uint8 status)
func (_Contract *ContractCaller) BatchDeposits(opts *bind.CallOpts, arg0 *big.Int, arg1 *big.Int) (struct {
	Nonce        *big.Int
	TokenAddress common.Address
	Amount       *big.Int
	Depositor    common.Address
	Recipient    [32]byte
	Status       uint8
}, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "batchDeposits", arg0, arg1)

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
func (_Contract *ContractSession) BatchDeposits(arg0 *big.Int, arg1 *big.Int) (struct {
	Nonce        *big.Int
	TokenAddress common.Address
	Amount       *big.Int
	Depositor    common.Address
	Recipient    [32]byte
	Status       uint8
}, error) {
	return _Contract.Contract.BatchDeposits(&_Contract.CallOpts, arg0, arg1)
}

// BatchDeposits is a free data retrieval call binding the contract method 0x284c0c44.
//
// Solidity: function batchDeposits(uint256 , uint256 ) view returns(uint112 nonce, address tokenAddress, uint256 amount, address depositor, bytes32 recipient, uint8 status)
func (_Contract *ContractCallerSession) BatchDeposits(arg0 *big.Int, arg1 *big.Int) (struct {
	Nonce        *big.Int
	TokenAddress common.Address
	Amount       *big.Int
	Depositor    common.Address
	Recipient    [32]byte
	Status       uint8
}, error) {
	return _Contract.Contract.BatchDeposits(&_Contract.CallOpts, arg0, arg1)
}

// BatchSettleLimit is a free data retrieval call binding the contract method 0x2325b5f7.
//
// Solidity: function batchSettleLimit() view returns(uint8)
func (_Contract *ContractCaller) BatchSettleLimit(opts *bind.CallOpts) (uint8, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "batchSettleLimit")

	if err != nil {
		return *new(uint8), err
	}

	out0 := *abi.ConvertType(out[0], new(uint8)).(*uint8)

	return out0, err

}

// BatchSettleLimit is a free data retrieval call binding the contract method 0x2325b5f7.
//
// Solidity: function batchSettleLimit() view returns(uint8)
func (_Contract *ContractSession) BatchSettleLimit() (uint8, error) {
	return _Contract.Contract.BatchSettleLimit(&_Contract.CallOpts)
}

// BatchSettleLimit is a free data retrieval call binding the contract method 0x2325b5f7.
//
// Solidity: function batchSettleLimit() view returns(uint8)
func (_Contract *ContractCallerSession) BatchSettleLimit() (uint8, error) {
	return _Contract.Contract.BatchSettleLimit(&_Contract.CallOpts)
}

// BatchSize is a free data retrieval call binding the contract method 0xf4daaba1.
//
// Solidity: function batchSize() view returns(uint16)
func (_Contract *ContractCaller) BatchSize(opts *bind.CallOpts) (uint16, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "batchSize")

	if err != nil {
		return *new(uint16), err
	}

	out0 := *abi.ConvertType(out[0], new(uint16)).(*uint16)

	return out0, err

}

// BatchSize is a free data retrieval call binding the contract method 0xf4daaba1.
//
// Solidity: function batchSize() view returns(uint16)
func (_Contract *ContractSession) BatchSize() (uint16, error) {
	return _Contract.Contract.BatchSize(&_Contract.CallOpts)
}

// BatchSize is a free data retrieval call binding the contract method 0xf4daaba1.
//
// Solidity: function batchSize() view returns(uint16)
func (_Contract *ContractCallerSession) BatchSize() (uint16, error) {
	return _Contract.Contract.BatchSize(&_Contract.CallOpts)
}

// Batches is a free data retrieval call binding the contract method 0xb32c4d8d.
//
// Solidity: function batches(uint256 ) view returns(uint112 nonce, uint64 blockNumber, uint64 lastUpdatedBlockNumber, uint16 depositsCount)
func (_Contract *ContractCaller) Batches(opts *bind.CallOpts, arg0 *big.Int) (struct {
	Nonce                  *big.Int
	BlockNumber            uint64
	LastUpdatedBlockNumber uint64
	DepositsCount          uint16
}, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "batches", arg0)

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
func (_Contract *ContractSession) Batches(arg0 *big.Int) (struct {
	Nonce                  *big.Int
	BlockNumber            uint64
	LastUpdatedBlockNumber uint64
	DepositsCount          uint16
}, error) {
	return _Contract.Contract.Batches(&_Contract.CallOpts, arg0)
}

// Batches is a free data retrieval call binding the contract method 0xb32c4d8d.
//
// Solidity: function batches(uint256 ) view returns(uint112 nonce, uint64 blockNumber, uint64 lastUpdatedBlockNumber, uint16 depositsCount)
func (_Contract *ContractCallerSession) Batches(arg0 *big.Int) (struct {
	Nonce                  *big.Int
	BlockNumber            uint64
	LastUpdatedBlockNumber uint64
	DepositsCount          uint16
}, error) {
	return _Contract.Contract.Batches(&_Contract.CallOpts, arg0)
}

// BatchesCount is a free data retrieval call binding the contract method 0x87ea0961.
//
// Solidity: function batchesCount() view returns(uint64)
func (_Contract *ContractCaller) BatchesCount(opts *bind.CallOpts) (uint64, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "batchesCount")

	if err != nil {
		return *new(uint64), err
	}

	out0 := *abi.ConvertType(out[0], new(uint64)).(*uint64)

	return out0, err

}

// BatchesCount is a free data retrieval call binding the contract method 0x87ea0961.
//
// Solidity: function batchesCount() view returns(uint64)
func (_Contract *ContractSession) BatchesCount() (uint64, error) {
	return _Contract.Contract.BatchesCount(&_Contract.CallOpts)
}

// BatchesCount is a free data retrieval call binding the contract method 0x87ea0961.
//
// Solidity: function batchesCount() view returns(uint64)
func (_Contract *ContractCallerSession) BatchesCount() (uint64, error) {
	return _Contract.Contract.BatchesCount(&_Contract.CallOpts)
}

// Bridge is a free data retrieval call binding the contract method 0xe78cea92.
//
// Solidity: function bridge() view returns(address)
func (_Contract *ContractCaller) Bridge(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "bridge")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Bridge is a free data retrieval call binding the contract method 0xe78cea92.
//
// Solidity: function bridge() view returns(address)
func (_Contract *ContractSession) Bridge() (common.Address, error) {
	return _Contract.Contract.Bridge(&_Contract.CallOpts)
}

// Bridge is a free data retrieval call binding the contract method 0xe78cea92.
//
// Solidity: function bridge() view returns(address)
func (_Contract *ContractCallerSession) Bridge() (common.Address, error) {
	return _Contract.Contract.Bridge(&_Contract.CallOpts)
}

// DepositsCount is a free data retrieval call binding the contract method 0x4506e935.
//
// Solidity: function depositsCount() view returns(uint64)
func (_Contract *ContractCaller) DepositsCount(opts *bind.CallOpts) (uint64, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "depositsCount")

	if err != nil {
		return *new(uint64), err
	}

	out0 := *abi.ConvertType(out[0], new(uint64)).(*uint64)

	return out0, err

}

// DepositsCount is a free data retrieval call binding the contract method 0x4506e935.
//
// Solidity: function depositsCount() view returns(uint64)
func (_Contract *ContractSession) DepositsCount() (uint64, error) {
	return _Contract.Contract.DepositsCount(&_Contract.CallOpts)
}

// DepositsCount is a free data retrieval call binding the contract method 0x4506e935.
//
// Solidity: function depositsCount() view returns(uint64)
func (_Contract *ContractCallerSession) DepositsCount() (uint64, error) {
	return _Contract.Contract.DepositsCount(&_Contract.CallOpts)
}

// GetBatch is a free data retrieval call binding the contract method 0x5ac44282.
//
// Solidity: function getBatch(uint256 batchNonce) view returns((uint112,uint64,uint64,uint16))
func (_Contract *ContractCaller) GetBatch(opts *bind.CallOpts, batchNonce *big.Int) (Batch, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "getBatch", batchNonce)

	if err != nil {
		return *new(Batch), err
	}

	out0 := *abi.ConvertType(out[0], new(Batch)).(*Batch)

	return out0, err

}

// GetBatch is a free data retrieval call binding the contract method 0x5ac44282.
//
// Solidity: function getBatch(uint256 batchNonce) view returns((uint112,uint64,uint64,uint16))
func (_Contract *ContractSession) GetBatch(batchNonce *big.Int) (Batch, error) {
	return _Contract.Contract.GetBatch(&_Contract.CallOpts, batchNonce)
}

// GetBatch is a free data retrieval call binding the contract method 0x5ac44282.
//
// Solidity: function getBatch(uint256 batchNonce) view returns((uint112,uint64,uint64,uint16))
func (_Contract *ContractCallerSession) GetBatch(batchNonce *big.Int) (Batch, error) {
	return _Contract.Contract.GetBatch(&_Contract.CallOpts, batchNonce)
}

// GetDeposits is a free data retrieval call binding the contract method 0x085c967f.
//
// Solidity: function getDeposits(uint256 batchNonce) view returns((uint112,address,uint256,address,bytes32,uint8)[])
func (_Contract *ContractCaller) GetDeposits(opts *bind.CallOpts, batchNonce *big.Int) ([]Deposit, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "getDeposits", batchNonce)

	if err != nil {
		return *new([]Deposit), err
	}

	out0 := *abi.ConvertType(out[0], new([]Deposit)).(*[]Deposit)

	return out0, err

}

// GetDeposits is a free data retrieval call binding the contract method 0x085c967f.
//
// Solidity: function getDeposits(uint256 batchNonce) view returns((uint112,address,uint256,address,bytes32,uint8)[])
func (_Contract *ContractSession) GetDeposits(batchNonce *big.Int) ([]Deposit, error) {
	return _Contract.Contract.GetDeposits(&_Contract.CallOpts, batchNonce)
}

// GetDeposits is a free data retrieval call binding the contract method 0x085c967f.
//
// Solidity: function getDeposits(uint256 batchNonce) view returns((uint112,address,uint256,address,bytes32,uint8)[])
func (_Contract *ContractCallerSession) GetDeposits(batchNonce *big.Int) ([]Deposit, error) {
	return _Contract.Contract.GetDeposits(&_Contract.CallOpts, batchNonce)
}

// GetTokenMaxLimit is a free data retrieval call binding the contract method 0xc652a0b5.
//
// Solidity: function getTokenMaxLimit(address token) view returns(uint256)
func (_Contract *ContractCaller) GetTokenMaxLimit(opts *bind.CallOpts, token common.Address) (*big.Int, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "getTokenMaxLimit", token)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetTokenMaxLimit is a free data retrieval call binding the contract method 0xc652a0b5.
//
// Solidity: function getTokenMaxLimit(address token) view returns(uint256)
func (_Contract *ContractSession) GetTokenMaxLimit(token common.Address) (*big.Int, error) {
	return _Contract.Contract.GetTokenMaxLimit(&_Contract.CallOpts, token)
}

// GetTokenMaxLimit is a free data retrieval call binding the contract method 0xc652a0b5.
//
// Solidity: function getTokenMaxLimit(address token) view returns(uint256)
func (_Contract *ContractCallerSession) GetTokenMaxLimit(token common.Address) (*big.Int, error) {
	return _Contract.Contract.GetTokenMaxLimit(&_Contract.CallOpts, token)
}

// GetTokenMinLimit is a free data retrieval call binding the contract method 0x9f0ebb93.
//
// Solidity: function getTokenMinLimit(address token) view returns(uint256)
func (_Contract *ContractCaller) GetTokenMinLimit(opts *bind.CallOpts, token common.Address) (*big.Int, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "getTokenMinLimit", token)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetTokenMinLimit is a free data retrieval call binding the contract method 0x9f0ebb93.
//
// Solidity: function getTokenMinLimit(address token) view returns(uint256)
func (_Contract *ContractSession) GetTokenMinLimit(token common.Address) (*big.Int, error) {
	return _Contract.Contract.GetTokenMinLimit(&_Contract.CallOpts, token)
}

// GetTokenMinLimit is a free data retrieval call binding the contract method 0x9f0ebb93.
//
// Solidity: function getTokenMinLimit(address token) view returns(uint256)
func (_Contract *ContractCallerSession) GetTokenMinLimit(token common.Address) (*big.Int, error) {
	return _Contract.Contract.GetTokenMinLimit(&_Contract.CallOpts, token)
}

// IsAnyBatchInProgress is a free data retrieval call binding the contract method 0x82146138.
//
// Solidity: function isAnyBatchInProgress() view returns(bool)
func (_Contract *ContractCaller) IsAnyBatchInProgress(opts *bind.CallOpts) (bool, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "isAnyBatchInProgress")

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsAnyBatchInProgress is a free data retrieval call binding the contract method 0x82146138.
//
// Solidity: function isAnyBatchInProgress() view returns(bool)
func (_Contract *ContractSession) IsAnyBatchInProgress() (bool, error) {
	return _Contract.Contract.IsAnyBatchInProgress(&_Contract.CallOpts)
}

// IsAnyBatchInProgress is a free data retrieval call binding the contract method 0x82146138.
//
// Solidity: function isAnyBatchInProgress() view returns(bool)
func (_Contract *ContractCallerSession) IsAnyBatchInProgress() (bool, error) {
	return _Contract.Contract.IsAnyBatchInProgress(&_Contract.CallOpts)
}

// IsTokenWhitelisted is a free data retrieval call binding the contract method 0xb5af090f.
//
// Solidity: function isTokenWhitelisted(address token) view returns(bool)
func (_Contract *ContractCaller) IsTokenWhitelisted(opts *bind.CallOpts, token common.Address) (bool, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "isTokenWhitelisted", token)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsTokenWhitelisted is a free data retrieval call binding the contract method 0xb5af090f.
//
// Solidity: function isTokenWhitelisted(address token) view returns(bool)
func (_Contract *ContractSession) IsTokenWhitelisted(token common.Address) (bool, error) {
	return _Contract.Contract.IsTokenWhitelisted(&_Contract.CallOpts, token)
}

// IsTokenWhitelisted is a free data retrieval call binding the contract method 0xb5af090f.
//
// Solidity: function isTokenWhitelisted(address token) view returns(bool)
func (_Contract *ContractCallerSession) IsTokenWhitelisted(token common.Address) (bool, error) {
	return _Contract.Contract.IsTokenWhitelisted(&_Contract.CallOpts, token)
}

// Paused is a free data retrieval call binding the contract method 0x5c975abb.
//
// Solidity: function paused() view returns(bool)
func (_Contract *ContractCaller) Paused(opts *bind.CallOpts) (bool, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "paused")

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// Paused is a free data retrieval call binding the contract method 0x5c975abb.
//
// Solidity: function paused() view returns(bool)
func (_Contract *ContractSession) Paused() (bool, error) {
	return _Contract.Contract.Paused(&_Contract.CallOpts)
}

// Paused is a free data retrieval call binding the contract method 0x5c975abb.
//
// Solidity: function paused() view returns(bool)
func (_Contract *ContractCallerSession) Paused() (bool, error) {
	return _Contract.Contract.Paused(&_Contract.CallOpts)
}

// TokenBalances is a free data retrieval call binding the contract method 0x523fba7f.
//
// Solidity: function tokenBalances(address ) view returns(uint256)
func (_Contract *ContractCaller) TokenBalances(opts *bind.CallOpts, arg0 common.Address) (*big.Int, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "tokenBalances", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// TokenBalances is a free data retrieval call binding the contract method 0x523fba7f.
//
// Solidity: function tokenBalances(address ) view returns(uint256)
func (_Contract *ContractSession) TokenBalances(arg0 common.Address) (*big.Int, error) {
	return _Contract.Contract.TokenBalances(&_Contract.CallOpts, arg0)
}

// TokenBalances is a free data retrieval call binding the contract method 0x523fba7f.
//
// Solidity: function tokenBalances(address ) view returns(uint256)
func (_Contract *ContractCallerSession) TokenBalances(arg0 common.Address) (*big.Int, error) {
	return _Contract.Contract.TokenBalances(&_Contract.CallOpts, arg0)
}

// TokenMaxLimits is a free data retrieval call binding the contract method 0xc639651d.
//
// Solidity: function tokenMaxLimits(address ) view returns(uint256)
func (_Contract *ContractCaller) TokenMaxLimits(opts *bind.CallOpts, arg0 common.Address) (*big.Int, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "tokenMaxLimits", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// TokenMaxLimits is a free data retrieval call binding the contract method 0xc639651d.
//
// Solidity: function tokenMaxLimits(address ) view returns(uint256)
func (_Contract *ContractSession) TokenMaxLimits(arg0 common.Address) (*big.Int, error) {
	return _Contract.Contract.TokenMaxLimits(&_Contract.CallOpts, arg0)
}

// TokenMaxLimits is a free data retrieval call binding the contract method 0xc639651d.
//
// Solidity: function tokenMaxLimits(address ) view returns(uint256)
func (_Contract *ContractCallerSession) TokenMaxLimits(arg0 common.Address) (*big.Int, error) {
	return _Contract.Contract.TokenMaxLimits(&_Contract.CallOpts, arg0)
}

// TokenMinLimits is a free data retrieval call binding the contract method 0xf6246ea1.
//
// Solidity: function tokenMinLimits(address ) view returns(uint256)
func (_Contract *ContractCaller) TokenMinLimits(opts *bind.CallOpts, arg0 common.Address) (*big.Int, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "tokenMinLimits", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// TokenMinLimits is a free data retrieval call binding the contract method 0xf6246ea1.
//
// Solidity: function tokenMinLimits(address ) view returns(uint256)
func (_Contract *ContractSession) TokenMinLimits(arg0 common.Address) (*big.Int, error) {
	return _Contract.Contract.TokenMinLimits(&_Contract.CallOpts, arg0)
}

// TokenMinLimits is a free data retrieval call binding the contract method 0xf6246ea1.
//
// Solidity: function tokenMinLimits(address ) view returns(uint256)
func (_Contract *ContractCallerSession) TokenMinLimits(arg0 common.Address) (*big.Int, error) {
	return _Contract.Contract.TokenMinLimits(&_Contract.CallOpts, arg0)
}

// TokenMintedBalances is a free data retrieval call binding the contract method 0x34ae3850.
//
// Solidity: function tokenMintedBalances(address ) view returns(uint256)
func (_Contract *ContractCaller) TokenMintedBalances(opts *bind.CallOpts, arg0 common.Address) (*big.Int, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "tokenMintedBalances", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// TokenMintedBalances is a free data retrieval call binding the contract method 0x34ae3850.
//
// Solidity: function tokenMintedBalances(address ) view returns(uint256)
func (_Contract *ContractSession) TokenMintedBalances(arg0 common.Address) (*big.Int, error) {
	return _Contract.Contract.TokenMintedBalances(&_Contract.CallOpts, arg0)
}

// TokenMintedBalances is a free data retrieval call binding the contract method 0x34ae3850.
//
// Solidity: function tokenMintedBalances(address ) view returns(uint256)
func (_Contract *ContractCallerSession) TokenMintedBalances(arg0 common.Address) (*big.Int, error) {
	return _Contract.Contract.TokenMintedBalances(&_Contract.CallOpts, arg0)
}

// WhitelistedTokens is a free data retrieval call binding the contract method 0xdaf9c210.
//
// Solidity: function whitelistedTokens(address ) view returns(bool)
func (_Contract *ContractCaller) WhitelistedTokens(opts *bind.CallOpts, arg0 common.Address) (bool, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "whitelistedTokens", arg0)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// WhitelistedTokens is a free data retrieval call binding the contract method 0xdaf9c210.
//
// Solidity: function whitelistedTokens(address ) view returns(bool)
func (_Contract *ContractSession) WhitelistedTokens(arg0 common.Address) (bool, error) {
	return _Contract.Contract.WhitelistedTokens(&_Contract.CallOpts, arg0)
}

// WhitelistedTokens is a free data retrieval call binding the contract method 0xdaf9c210.
//
// Solidity: function whitelistedTokens(address ) view returns(bool)
func (_Contract *ContractCallerSession) WhitelistedTokens(arg0 common.Address) (bool, error) {
	return _Contract.Contract.WhitelistedTokens(&_Contract.CallOpts, arg0)
}

// WhitelistedTokensMintBurn is a free data retrieval call binding the contract method 0x48db2fca.
//
// Solidity: function whitelistedTokensMintBurn(address ) view returns(bool)
func (_Contract *ContractCaller) WhitelistedTokensMintBurn(opts *bind.CallOpts, arg0 common.Address) (bool, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "whitelistedTokensMintBurn", arg0)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// WhitelistedTokensMintBurn is a free data retrieval call binding the contract method 0x48db2fca.
//
// Solidity: function whitelistedTokensMintBurn(address ) view returns(bool)
func (_Contract *ContractSession) WhitelistedTokensMintBurn(arg0 common.Address) (bool, error) {
	return _Contract.Contract.WhitelistedTokensMintBurn(&_Contract.CallOpts, arg0)
}

// WhitelistedTokensMintBurn is a free data retrieval call binding the contract method 0x48db2fca.
//
// Solidity: function whitelistedTokensMintBurn(address ) view returns(bool)
func (_Contract *ContractCallerSession) WhitelistedTokensMintBurn(arg0 common.Address) (bool, error) {
	return _Contract.Contract.WhitelistedTokensMintBurn(&_Contract.CallOpts, arg0)
}

// Deposit is a paid mutator transaction binding the contract method 0x26b3293f.
//
// Solidity: function deposit(address tokenAddress, uint256 amount, bytes32 recipientAddress) returns()
func (_Contract *ContractTransactor) Deposit(opts *bind.TransactOpts, tokenAddress common.Address, amount *big.Int, recipientAddress [32]byte) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "deposit", tokenAddress, amount, recipientAddress)
}

// Deposit is a paid mutator transaction binding the contract method 0x26b3293f.
//
// Solidity: function deposit(address tokenAddress, uint256 amount, bytes32 recipientAddress) returns()
func (_Contract *ContractSession) Deposit(tokenAddress common.Address, amount *big.Int, recipientAddress [32]byte) (*types.Transaction, error) {
	return _Contract.Contract.Deposit(&_Contract.TransactOpts, tokenAddress, amount, recipientAddress)
}

// Deposit is a paid mutator transaction binding the contract method 0x26b3293f.
//
// Solidity: function deposit(address tokenAddress, uint256 amount, bytes32 recipientAddress) returns()
func (_Contract *ContractTransactorSession) Deposit(tokenAddress common.Address, amount *big.Int, recipientAddress [32]byte) (*types.Transaction, error) {
	return _Contract.Contract.Deposit(&_Contract.TransactOpts, tokenAddress, amount, recipientAddress)
}

// InitSupply is a paid mutator transaction binding the contract method 0x4013c89c.
//
// Solidity: function initSupply(address tokenAddress, uint256 amount) returns()
func (_Contract *ContractTransactor) InitSupply(opts *bind.TransactOpts, tokenAddress common.Address, amount *big.Int) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "initSupply", tokenAddress, amount)
}

// InitSupply is a paid mutator transaction binding the contract method 0x4013c89c.
//
// Solidity: function initSupply(address tokenAddress, uint256 amount) returns()
func (_Contract *ContractSession) InitSupply(tokenAddress common.Address, amount *big.Int) (*types.Transaction, error) {
	return _Contract.Contract.InitSupply(&_Contract.TransactOpts, tokenAddress, amount)
}

// InitSupply is a paid mutator transaction binding the contract method 0x4013c89c.
//
// Solidity: function initSupply(address tokenAddress, uint256 amount) returns()
func (_Contract *ContractTransactorSession) InitSupply(tokenAddress common.Address, amount *big.Int) (*types.Transaction, error) {
	return _Contract.Contract.InitSupply(&_Contract.TransactOpts, tokenAddress, amount)
}

// Pause is a paid mutator transaction binding the contract method 0x8456cb59.
//
// Solidity: function pause() returns()
func (_Contract *ContractTransactor) Pause(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "pause")
}

// Pause is a paid mutator transaction binding the contract method 0x8456cb59.
//
// Solidity: function pause() returns()
func (_Contract *ContractSession) Pause() (*types.Transaction, error) {
	return _Contract.Contract.Pause(&_Contract.TransactOpts)
}

// Pause is a paid mutator transaction binding the contract method 0x8456cb59.
//
// Solidity: function pause() returns()
func (_Contract *ContractTransactorSession) Pause() (*types.Transaction, error) {
	return _Contract.Contract.Pause(&_Contract.TransactOpts)
}

// RecoverLostFunds is a paid mutator transaction binding the contract method 0x770be784.
//
// Solidity: function recoverLostFunds(address tokenAddress) returns()
func (_Contract *ContractTransactor) RecoverLostFunds(opts *bind.TransactOpts, tokenAddress common.Address) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "recoverLostFunds", tokenAddress)
}

// RecoverLostFunds is a paid mutator transaction binding the contract method 0x770be784.
//
// Solidity: function recoverLostFunds(address tokenAddress) returns()
func (_Contract *ContractSession) RecoverLostFunds(tokenAddress common.Address) (*types.Transaction, error) {
	return _Contract.Contract.RecoverLostFunds(&_Contract.TransactOpts, tokenAddress)
}

// RecoverLostFunds is a paid mutator transaction binding the contract method 0x770be784.
//
// Solidity: function recoverLostFunds(address tokenAddress) returns()
func (_Contract *ContractTransactorSession) RecoverLostFunds(tokenAddress common.Address) (*types.Transaction, error) {
	return _Contract.Contract.RecoverLostFunds(&_Contract.TransactOpts, tokenAddress)
}

// RemoveTokenFromWhitelist is a paid mutator transaction binding the contract method 0x306275be.
//
// Solidity: function removeTokenFromWhitelist(address token) returns()
func (_Contract *ContractTransactor) RemoveTokenFromWhitelist(opts *bind.TransactOpts, token common.Address) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "removeTokenFromWhitelist", token)
}

// RemoveTokenFromWhitelist is a paid mutator transaction binding the contract method 0x306275be.
//
// Solidity: function removeTokenFromWhitelist(address token) returns()
func (_Contract *ContractSession) RemoveTokenFromWhitelist(token common.Address) (*types.Transaction, error) {
	return _Contract.Contract.RemoveTokenFromWhitelist(&_Contract.TransactOpts, token)
}

// RemoveTokenFromWhitelist is a paid mutator transaction binding the contract method 0x306275be.
//
// Solidity: function removeTokenFromWhitelist(address token) returns()
func (_Contract *ContractTransactorSession) RemoveTokenFromWhitelist(token common.Address) (*types.Transaction, error) {
	return _Contract.Contract.RemoveTokenFromWhitelist(&_Contract.TransactOpts, token)
}

// RenounceAdmin is a paid mutator transaction binding the contract method 0x8bad0c0a.
//
// Solidity: function renounceAdmin() returns()
func (_Contract *ContractTransactor) RenounceAdmin(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "renounceAdmin")
}

// RenounceAdmin is a paid mutator transaction binding the contract method 0x8bad0c0a.
//
// Solidity: function renounceAdmin() returns()
func (_Contract *ContractSession) RenounceAdmin() (*types.Transaction, error) {
	return _Contract.Contract.RenounceAdmin(&_Contract.TransactOpts)
}

// RenounceAdmin is a paid mutator transaction binding the contract method 0x8bad0c0a.
//
// Solidity: function renounceAdmin() returns()
func (_Contract *ContractTransactorSession) RenounceAdmin() (*types.Transaction, error) {
	return _Contract.Contract.RenounceAdmin(&_Contract.TransactOpts)
}

// SetBatchBlockLimit is a paid mutator transaction binding the contract method 0xe8a70ee2.
//
// Solidity: function setBatchBlockLimit(uint8 newBatchBlockLimit) returns()
func (_Contract *ContractTransactor) SetBatchBlockLimit(opts *bind.TransactOpts, newBatchBlockLimit uint8) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "setBatchBlockLimit", newBatchBlockLimit)
}

// SetBatchBlockLimit is a paid mutator transaction binding the contract method 0xe8a70ee2.
//
// Solidity: function setBatchBlockLimit(uint8 newBatchBlockLimit) returns()
func (_Contract *ContractSession) SetBatchBlockLimit(newBatchBlockLimit uint8) (*types.Transaction, error) {
	return _Contract.Contract.SetBatchBlockLimit(&_Contract.TransactOpts, newBatchBlockLimit)
}

// SetBatchBlockLimit is a paid mutator transaction binding the contract method 0xe8a70ee2.
//
// Solidity: function setBatchBlockLimit(uint8 newBatchBlockLimit) returns()
func (_Contract *ContractTransactorSession) SetBatchBlockLimit(newBatchBlockLimit uint8) (*types.Transaction, error) {
	return _Contract.Contract.SetBatchBlockLimit(&_Contract.TransactOpts, newBatchBlockLimit)
}

// SetBatchSettleLimit is a paid mutator transaction binding the contract method 0xf2e0ec48.
//
// Solidity: function setBatchSettleLimit(uint8 newBatchSettleLimit) returns()
func (_Contract *ContractTransactor) SetBatchSettleLimit(opts *bind.TransactOpts, newBatchSettleLimit uint8) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "setBatchSettleLimit", newBatchSettleLimit)
}

// SetBatchSettleLimit is a paid mutator transaction binding the contract method 0xf2e0ec48.
//
// Solidity: function setBatchSettleLimit(uint8 newBatchSettleLimit) returns()
func (_Contract *ContractSession) SetBatchSettleLimit(newBatchSettleLimit uint8) (*types.Transaction, error) {
	return _Contract.Contract.SetBatchSettleLimit(&_Contract.TransactOpts, newBatchSettleLimit)
}

// SetBatchSettleLimit is a paid mutator transaction binding the contract method 0xf2e0ec48.
//
// Solidity: function setBatchSettleLimit(uint8 newBatchSettleLimit) returns()
func (_Contract *ContractTransactorSession) SetBatchSettleLimit(newBatchSettleLimit uint8) (*types.Transaction, error) {
	return _Contract.Contract.SetBatchSettleLimit(&_Contract.TransactOpts, newBatchSettleLimit)
}

// SetBatchSize is a paid mutator transaction binding the contract method 0xd4673de9.
//
// Solidity: function setBatchSize(uint16 newBatchSize) returns()
func (_Contract *ContractTransactor) SetBatchSize(opts *bind.TransactOpts, newBatchSize uint16) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "setBatchSize", newBatchSize)
}

// SetBatchSize is a paid mutator transaction binding the contract method 0xd4673de9.
//
// Solidity: function setBatchSize(uint16 newBatchSize) returns()
func (_Contract *ContractSession) SetBatchSize(newBatchSize uint16) (*types.Transaction, error) {
	return _Contract.Contract.SetBatchSize(&_Contract.TransactOpts, newBatchSize)
}

// SetBatchSize is a paid mutator transaction binding the contract method 0xd4673de9.
//
// Solidity: function setBatchSize(uint16 newBatchSize) returns()
func (_Contract *ContractTransactorSession) SetBatchSize(newBatchSize uint16) (*types.Transaction, error) {
	return _Contract.Contract.SetBatchSize(&_Contract.TransactOpts, newBatchSize)
}

// SetBridge is a paid mutator transaction binding the contract method 0x8dd14802.
//
// Solidity: function setBridge(address newBridge) returns()
func (_Contract *ContractTransactor) SetBridge(opts *bind.TransactOpts, newBridge common.Address) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "setBridge", newBridge)
}

// SetBridge is a paid mutator transaction binding the contract method 0x8dd14802.
//
// Solidity: function setBridge(address newBridge) returns()
func (_Contract *ContractSession) SetBridge(newBridge common.Address) (*types.Transaction, error) {
	return _Contract.Contract.SetBridge(&_Contract.TransactOpts, newBridge)
}

// SetBridge is a paid mutator transaction binding the contract method 0x8dd14802.
//
// Solidity: function setBridge(address newBridge) returns()
func (_Contract *ContractTransactorSession) SetBridge(newBridge common.Address) (*types.Transaction, error) {
	return _Contract.Contract.SetBridge(&_Contract.TransactOpts, newBridge)
}

// SetTokenMaxLimit is a paid mutator transaction binding the contract method 0x7d7763ce.
//
// Solidity: function setTokenMaxLimit(address token, uint256 amount) returns()
func (_Contract *ContractTransactor) SetTokenMaxLimit(opts *bind.TransactOpts, token common.Address, amount *big.Int) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "setTokenMaxLimit", token, amount)
}

// SetTokenMaxLimit is a paid mutator transaction binding the contract method 0x7d7763ce.
//
// Solidity: function setTokenMaxLimit(address token, uint256 amount) returns()
func (_Contract *ContractSession) SetTokenMaxLimit(token common.Address, amount *big.Int) (*types.Transaction, error) {
	return _Contract.Contract.SetTokenMaxLimit(&_Contract.TransactOpts, token, amount)
}

// SetTokenMaxLimit is a paid mutator transaction binding the contract method 0x7d7763ce.
//
// Solidity: function setTokenMaxLimit(address token, uint256 amount) returns()
func (_Contract *ContractTransactorSession) SetTokenMaxLimit(token common.Address, amount *big.Int) (*types.Transaction, error) {
	return _Contract.Contract.SetTokenMaxLimit(&_Contract.TransactOpts, token, amount)
}

// SetTokenMinLimit is a paid mutator transaction binding the contract method 0x920b0308.
//
// Solidity: function setTokenMinLimit(address token, uint256 amount) returns()
func (_Contract *ContractTransactor) SetTokenMinLimit(opts *bind.TransactOpts, token common.Address, amount *big.Int) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "setTokenMinLimit", token, amount)
}

// SetTokenMinLimit is a paid mutator transaction binding the contract method 0x920b0308.
//
// Solidity: function setTokenMinLimit(address token, uint256 amount) returns()
func (_Contract *ContractSession) SetTokenMinLimit(token common.Address, amount *big.Int) (*types.Transaction, error) {
	return _Contract.Contract.SetTokenMinLimit(&_Contract.TransactOpts, token, amount)
}

// SetTokenMinLimit is a paid mutator transaction binding the contract method 0x920b0308.
//
// Solidity: function setTokenMinLimit(address token, uint256 amount) returns()
func (_Contract *ContractTransactorSession) SetTokenMinLimit(token common.Address, amount *big.Int) (*types.Transaction, error) {
	return _Contract.Contract.SetTokenMinLimit(&_Contract.TransactOpts, token, amount)
}

// Transfer is a paid mutator transaction binding the contract method 0xdbba0f01.
//
// Solidity: function transfer(address tokenAddress, uint256 amount, address recipientAddress) returns(bool)
func (_Contract *ContractTransactor) Transfer(opts *bind.TransactOpts, tokenAddress common.Address, amount *big.Int, recipientAddress common.Address) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "transfer", tokenAddress, amount, recipientAddress)
}

// Transfer is a paid mutator transaction binding the contract method 0xdbba0f01.
//
// Solidity: function transfer(address tokenAddress, uint256 amount, address recipientAddress) returns(bool)
func (_Contract *ContractSession) Transfer(tokenAddress common.Address, amount *big.Int, recipientAddress common.Address) (*types.Transaction, error) {
	return _Contract.Contract.Transfer(&_Contract.TransactOpts, tokenAddress, amount, recipientAddress)
}

// Transfer is a paid mutator transaction binding the contract method 0xdbba0f01.
//
// Solidity: function transfer(address tokenAddress, uint256 amount, address recipientAddress) returns(bool)
func (_Contract *ContractTransactorSession) Transfer(tokenAddress common.Address, amount *big.Int, recipientAddress common.Address) (*types.Transaction, error) {
	return _Contract.Contract.Transfer(&_Contract.TransactOpts, tokenAddress, amount, recipientAddress)
}

// TransferAdmin is a paid mutator transaction binding the contract method 0x75829def.
//
// Solidity: function transferAdmin(address newAdmin) returns()
func (_Contract *ContractTransactor) TransferAdmin(opts *bind.TransactOpts, newAdmin common.Address) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "transferAdmin", newAdmin)
}

// TransferAdmin is a paid mutator transaction binding the contract method 0x75829def.
//
// Solidity: function transferAdmin(address newAdmin) returns()
func (_Contract *ContractSession) TransferAdmin(newAdmin common.Address) (*types.Transaction, error) {
	return _Contract.Contract.TransferAdmin(&_Contract.TransactOpts, newAdmin)
}

// TransferAdmin is a paid mutator transaction binding the contract method 0x75829def.
//
// Solidity: function transferAdmin(address newAdmin) returns()
func (_Contract *ContractTransactorSession) TransferAdmin(newAdmin common.Address) (*types.Transaction, error) {
	return _Contract.Contract.TransferAdmin(&_Contract.TransactOpts, newAdmin)
}

// Unpause is a paid mutator transaction binding the contract method 0x3f4ba83a.
//
// Solidity: function unpause() returns()
func (_Contract *ContractTransactor) Unpause(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "unpause")
}

// Unpause is a paid mutator transaction binding the contract method 0x3f4ba83a.
//
// Solidity: function unpause() returns()
func (_Contract *ContractSession) Unpause() (*types.Transaction, error) {
	return _Contract.Contract.Unpause(&_Contract.TransactOpts)
}

// Unpause is a paid mutator transaction binding the contract method 0x3f4ba83a.
//
// Solidity: function unpause() returns()
func (_Contract *ContractTransactorSession) Unpause() (*types.Transaction, error) {
	return _Contract.Contract.Unpause(&_Contract.TransactOpts)
}

// WhitelistToken is a paid mutator transaction binding the contract method 0x5fd94707.
//
// Solidity: function whitelistToken(address token, uint256 minimumAmount, uint256 maximumAmount, bool mintBurn) returns()
func (_Contract *ContractTransactor) WhitelistToken(opts *bind.TransactOpts, token common.Address, minimumAmount *big.Int, maximumAmount *big.Int, mintBurn bool) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "whitelistToken", token, minimumAmount, maximumAmount, mintBurn)
}

// WhitelistToken is a paid mutator transaction binding the contract method 0x5fd94707.
//
// Solidity: function whitelistToken(address token, uint256 minimumAmount, uint256 maximumAmount, bool mintBurn) returns()
func (_Contract *ContractSession) WhitelistToken(token common.Address, minimumAmount *big.Int, maximumAmount *big.Int, mintBurn bool) (*types.Transaction, error) {
	return _Contract.Contract.WhitelistToken(&_Contract.TransactOpts, token, minimumAmount, maximumAmount, mintBurn)
}

// WhitelistToken is a paid mutator transaction binding the contract method 0x5fd94707.
//
// Solidity: function whitelistToken(address token, uint256 minimumAmount, uint256 maximumAmount, bool mintBurn) returns()
func (_Contract *ContractTransactorSession) WhitelistToken(token common.Address, minimumAmount *big.Int, maximumAmount *big.Int, mintBurn bool) (*types.Transaction, error) {
	return _Contract.Contract.WhitelistToken(&_Contract.TransactOpts, token, minimumAmount, maximumAmount, mintBurn)
}

// ContractAdminRoleTransferredIterator is returned from FilterAdminRoleTransferred and is used to iterate over the raw logs and unpacked data for AdminRoleTransferred events raised by the Contract contract.
type ContractAdminRoleTransferredIterator struct {
	Event *ContractAdminRoleTransferred // Event containing the contract specifics and raw log

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
func (it *ContractAdminRoleTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractAdminRoleTransferred)
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
		it.Event = new(ContractAdminRoleTransferred)
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
func (it *ContractAdminRoleTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractAdminRoleTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractAdminRoleTransferred represents a AdminRoleTransferred event raised by the Contract contract.
type ContractAdminRoleTransferred struct {
	PreviousAdmin common.Address
	NewAdmin      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterAdminRoleTransferred is a free log retrieval operation binding the contract event 0xe379ac64de02d8184ca1a871ac486cb8137de77e485ede140e97057b9c765ffd.
//
// Solidity: event AdminRoleTransferred(address indexed previousAdmin, address indexed newAdmin)
func (_Contract *ContractFilterer) FilterAdminRoleTransferred(opts *bind.FilterOpts, previousAdmin []common.Address, newAdmin []common.Address) (*ContractAdminRoleTransferredIterator, error) {

	var previousAdminRule []interface{}
	for _, previousAdminItem := range previousAdmin {
		previousAdminRule = append(previousAdminRule, previousAdminItem)
	}
	var newAdminRule []interface{}
	for _, newAdminItem := range newAdmin {
		newAdminRule = append(newAdminRule, newAdminItem)
	}

	logs, sub, err := _Contract.contract.FilterLogs(opts, "AdminRoleTransferred", previousAdminRule, newAdminRule)
	if err != nil {
		return nil, err
	}
	return &ContractAdminRoleTransferredIterator{contract: _Contract.contract, event: "AdminRoleTransferred", logs: logs, sub: sub}, nil
}

// WatchAdminRoleTransferred is a free log subscription operation binding the contract event 0xe379ac64de02d8184ca1a871ac486cb8137de77e485ede140e97057b9c765ffd.
//
// Solidity: event AdminRoleTransferred(address indexed previousAdmin, address indexed newAdmin)
func (_Contract *ContractFilterer) WatchAdminRoleTransferred(opts *bind.WatchOpts, sink chan<- *ContractAdminRoleTransferred, previousAdmin []common.Address, newAdmin []common.Address) (event.Subscription, error) {

	var previousAdminRule []interface{}
	for _, previousAdminItem := range previousAdmin {
		previousAdminRule = append(previousAdminRule, previousAdminItem)
	}
	var newAdminRule []interface{}
	for _, newAdminItem := range newAdmin {
		newAdminRule = append(newAdminRule, newAdminItem)
	}

	logs, sub, err := _Contract.contract.WatchLogs(opts, "AdminRoleTransferred", previousAdminRule, newAdminRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractAdminRoleTransferred)
				if err := _Contract.contract.UnpackLog(event, "AdminRoleTransferred", log); err != nil {
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
func (_Contract *ContractFilterer) ParseAdminRoleTransferred(log types.Log) (*ContractAdminRoleTransferred, error) {
	event := new(ContractAdminRoleTransferred)
	if err := _Contract.contract.UnpackLog(event, "AdminRoleTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ContractBridgeTransferredIterator is returned from FilterBridgeTransferred and is used to iterate over the raw logs and unpacked data for BridgeTransferred events raised by the Contract contract.
type ContractBridgeTransferredIterator struct {
	Event *ContractBridgeTransferred // Event containing the contract specifics and raw log

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
func (it *ContractBridgeTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractBridgeTransferred)
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
		it.Event = new(ContractBridgeTransferred)
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
func (it *ContractBridgeTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractBridgeTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractBridgeTransferred represents a BridgeTransferred event raised by the Contract contract.
type ContractBridgeTransferred struct {
	PreviousBridge common.Address
	NewBridge      common.Address
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterBridgeTransferred is a free log retrieval operation binding the contract event 0xcca5fddab921a878ddbd4edb737a2cf3ac6df70864f108606647d1b37a5e07a0.
//
// Solidity: event BridgeTransferred(address indexed previousBridge, address indexed newBridge)
func (_Contract *ContractFilterer) FilterBridgeTransferred(opts *bind.FilterOpts, previousBridge []common.Address, newBridge []common.Address) (*ContractBridgeTransferredIterator, error) {

	var previousBridgeRule []interface{}
	for _, previousBridgeItem := range previousBridge {
		previousBridgeRule = append(previousBridgeRule, previousBridgeItem)
	}
	var newBridgeRule []interface{}
	for _, newBridgeItem := range newBridge {
		newBridgeRule = append(newBridgeRule, newBridgeItem)
	}

	logs, sub, err := _Contract.contract.FilterLogs(opts, "BridgeTransferred", previousBridgeRule, newBridgeRule)
	if err != nil {
		return nil, err
	}
	return &ContractBridgeTransferredIterator{contract: _Contract.contract, event: "BridgeTransferred", logs: logs, sub: sub}, nil
}

// WatchBridgeTransferred is a free log subscription operation binding the contract event 0xcca5fddab921a878ddbd4edb737a2cf3ac6df70864f108606647d1b37a5e07a0.
//
// Solidity: event BridgeTransferred(address indexed previousBridge, address indexed newBridge)
func (_Contract *ContractFilterer) WatchBridgeTransferred(opts *bind.WatchOpts, sink chan<- *ContractBridgeTransferred, previousBridge []common.Address, newBridge []common.Address) (event.Subscription, error) {

	var previousBridgeRule []interface{}
	for _, previousBridgeItem := range previousBridge {
		previousBridgeRule = append(previousBridgeRule, previousBridgeItem)
	}
	var newBridgeRule []interface{}
	for _, newBridgeItem := range newBridge {
		newBridgeRule = append(newBridgeRule, newBridgeItem)
	}

	logs, sub, err := _Contract.contract.WatchLogs(opts, "BridgeTransferred", previousBridgeRule, newBridgeRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractBridgeTransferred)
				if err := _Contract.contract.UnpackLog(event, "BridgeTransferred", log); err != nil {
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
func (_Contract *ContractFilterer) ParseBridgeTransferred(log types.Log) (*ContractBridgeTransferred, error) {
	event := new(ContractBridgeTransferred)
	if err := _Contract.contract.UnpackLog(event, "BridgeTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ContractERC20DepositIterator is returned from FilterERC20Deposit and is used to iterate over the raw logs and unpacked data for ERC20Deposit events raised by the Contract contract.
type ContractERC20DepositIterator struct {
	Event *ContractERC20Deposit // Event containing the contract specifics and raw log

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
func (it *ContractERC20DepositIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractERC20Deposit)
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
		it.Event = new(ContractERC20Deposit)
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
func (it *ContractERC20DepositIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractERC20DepositIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractERC20Deposit represents a ERC20Deposit event raised by the Contract contract.
type ContractERC20Deposit struct {
	DepositNonce *big.Int
	BatchId      *big.Int
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterERC20Deposit is a free log retrieval operation binding the contract event 0x6c15ce44793c685a79cde26a0bd5419ef4f3a337991f156be7b365962001b4a7.
//
// Solidity: event ERC20Deposit(uint112 depositNonce, uint112 batchId)
func (_Contract *ContractFilterer) FilterERC20Deposit(opts *bind.FilterOpts) (*ContractERC20DepositIterator, error) {

	logs, sub, err := _Contract.contract.FilterLogs(opts, "ERC20Deposit")
	if err != nil {
		return nil, err
	}
	return &ContractERC20DepositIterator{contract: _Contract.contract, event: "ERC20Deposit", logs: logs, sub: sub}, nil
}

// WatchERC20Deposit is a free log subscription operation binding the contract event 0x6c15ce44793c685a79cde26a0bd5419ef4f3a337991f156be7b365962001b4a7.
//
// Solidity: event ERC20Deposit(uint112 depositNonce, uint112 batchId)
func (_Contract *ContractFilterer) WatchERC20Deposit(opts *bind.WatchOpts, sink chan<- *ContractERC20Deposit) (event.Subscription, error) {

	logs, sub, err := _Contract.contract.WatchLogs(opts, "ERC20Deposit")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractERC20Deposit)
				if err := _Contract.contract.UnpackLog(event, "ERC20Deposit", log); err != nil {
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
// Solidity: event ERC20Deposit(uint112 depositNonce, uint112 batchId)
func (_Contract *ContractFilterer) ParseERC20Deposit(log types.Log) (*ContractERC20Deposit, error) {
	event := new(ContractERC20Deposit)
	if err := _Contract.contract.UnpackLog(event, "ERC20Deposit", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
