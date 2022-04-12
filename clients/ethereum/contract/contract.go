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
)

// Batch is an auto generated low-level Go binding around an user-defined struct.
type Batch struct {
	Nonce                *big.Int
	Timestamp            *big.Int
	LastUpdatedTimestamp *big.Int
	DepositsCount        uint8
}

// Deposit is an auto generated low-level Go binding around an user-defined struct.
type Deposit struct {
	Nonce        *big.Int
	TokenAddress common.Address
	Amount       *big.Int
	Depositor    common.Address
	Recipient    [32]byte
	Status       uint8
}

// BridgeMetaData contains all meta data concerning the Bridge contract.
var BridgeMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"address[]\",\"name\":\"board\",\"type\":\"address[]\"},{\"internalType\":\"uint256\",\"name\":\"initialQuorum\",\"type\":\"uint256\"},{\"internalType\":\"contractERC20Safe\",\"name\":\"erc20Safe\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousAdmin\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newAdmin\",\"type\":\"address\"}],\"name\":\"AdminRoleTransferred\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"quorum\",\"type\":\"uint256\"}],\"name\":\"QuorumChanged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"}],\"name\":\"RelayerAdded\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"}],\"name\":\"RelayerRemoved\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"addRelayer\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"admin\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"batchSettleBlockCount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"crossTransferStatuses\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"createdBlockNumber\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address[]\",\"name\":\"tokens\",\"type\":\"address[]\"},{\"internalType\":\"address[]\",\"name\":\"recipients\",\"type\":\"address[]\"},{\"internalType\":\"uint256[]\",\"name\":\"amounts\",\"type\":\"uint256[]\"},{\"internalType\":\"uint256[]\",\"name\":\"depositNonces\",\"type\":\"uint256[]\"},{\"internalType\":\"uint256\",\"name\":\"batchNonceElrondETH\",\"type\":\"uint256\"},{\"internalType\":\"bytes[]\",\"name\":\"signatures\",\"type\":\"bytes[]\"}],\"name\":\"executeTransfer\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"executedBatches\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"batchNonce\",\"type\":\"uint256\"}],\"name\":\"getBatch\",\"outputs\":[{\"components\":[{\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"lastUpdatedTimestamp\",\"type\":\"uint256\"},{\"internalType\":\"uint8\",\"name\":\"depositsCount\",\"type\":\"uint8\"}],\"internalType\":\"structBatch\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"batchNonce\",\"type\":\"uint256\"}],\"name\":\"getBatchDeposits\",\"outputs\":[{\"components\":[{\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"tokenAddress\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"depositor\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"recipient\",\"type\":\"bytes32\"},{\"internalType\":\"enumDepositStatus\",\"name\":\"status\",\"type\":\"uint8\"}],\"internalType\":\"structDeposit[]\",\"name\":\"\",\"type\":\"tuple[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"index\",\"type\":\"uint256\"}],\"name\":\"getRelayer\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getRelayers\",\"outputs\":[{\"internalType\":\"address[]\",\"name\":\"\",\"type\":\"address[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getRelayersCount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"batchNonceElrondETH\",\"type\":\"uint256\"}],\"name\":\"getStatusesAfterExecution\",\"outputs\":[{\"internalType\":\"enumDepositStatus[]\",\"name\":\"\",\"type\":\"uint8[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"isRelayer\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"quorum\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"removeRelayer\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceAdmin\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"renounceRelayer\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"newQuorum\",\"type\":\"uint256\"}],\"name\":\"setQuorum\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newAdmin\",\"type\":\"address\"}],\"name\":\"transferAdmin\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"batchNonceElrondETH\",\"type\":\"uint256\"}],\"name\":\"wasBatchExecuted\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]",
}

// BridgeABI is the input ABI used to generate the binding from.
// Deprecated: Use BridgeMetaData.ABI instead.
var BridgeABI = BridgeMetaData.ABI

// Bridge is an auto generated Go binding around an Ethereum contract.
type Bridge struct {
	BridgeCaller     // Read-only binding to the contract
	BridgeTransactor // Write-only binding to the contract
	BridgeFilterer   // Log filterer for contract events
}

// BridgeCaller is an auto generated read-only Go binding around an Ethereum contract.
type BridgeCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BridgeTransactor is an auto generated write-only Go binding around an Ethereum contract.
type BridgeTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BridgeFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type BridgeFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BridgeSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type BridgeSession struct {
	Contract     *Bridge           // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// BridgeCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type BridgeCallerSession struct {
	Contract *BridgeCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts // Call options to use throughout this session
}

// BridgeTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type BridgeTransactorSession struct {
	Contract     *BridgeTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// BridgeRaw is an auto generated low-level Go binding around an Ethereum contract.
type BridgeRaw struct {
	Contract *Bridge // Generic contract binding to access the raw methods on
}

// BridgeCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type BridgeCallerRaw struct {
	Contract *BridgeCaller // Generic read-only contract binding to access the raw methods on
}

// BridgeTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type BridgeTransactorRaw struct {
	Contract *BridgeTransactor // Generic write-only contract binding to access the raw methods on
}

// NewBridge creates a new instance of Bridge, bound to a specific deployed contract.
func NewBridge(address common.Address, backend bind.ContractBackend) (*Bridge, error) {
	contract, err := bindBridge(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Bridge{BridgeCaller: BridgeCaller{contract: contract}, BridgeTransactor: BridgeTransactor{contract: contract}, BridgeFilterer: BridgeFilterer{contract: contract}}, nil
}

// NewBridgeCaller creates a new read-only instance of Bridge, bound to a specific deployed contract.
func NewBridgeCaller(address common.Address, caller bind.ContractCaller) (*BridgeCaller, error) {
	contract, err := bindBridge(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &BridgeCaller{contract: contract}, nil
}

// NewBridgeTransactor creates a new write-only instance of Bridge, bound to a specific deployed contract.
func NewBridgeTransactor(address common.Address, transactor bind.ContractTransactor) (*BridgeTransactor, error) {
	contract, err := bindBridge(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &BridgeTransactor{contract: contract}, nil
}

// NewBridgeFilterer creates a new log filterer instance of Bridge, bound to a specific deployed contract.
func NewBridgeFilterer(address common.Address, filterer bind.ContractFilterer) (*BridgeFilterer, error) {
	contract, err := bindBridge(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &BridgeFilterer{contract: contract}, nil
}

// bindBridge binds a generic wrapper to an already deployed contract.
func bindBridge(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(BridgeABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Bridge *BridgeRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Bridge.Contract.BridgeCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Bridge *BridgeRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Bridge.Contract.BridgeTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Bridge *BridgeRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Bridge.Contract.BridgeTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Bridge *BridgeCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Bridge.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Bridge *BridgeTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Bridge.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Bridge *BridgeTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Bridge.Contract.contract.Transact(opts, method, params...)
}

// Admin is a free data retrieval call binding the contract method 0xf851a440.
//
// Solidity: function admin() view returns(address)
func (_Bridge *BridgeCaller) Admin(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Bridge.contract.Call(opts, &out, "admin")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Admin is a free data retrieval call binding the contract method 0xf851a440.
//
// Solidity: function admin() view returns(address)
func (_Bridge *BridgeSession) Admin() (common.Address, error) {
	return _Bridge.Contract.Admin(&_Bridge.CallOpts)
}

// Admin is a free data retrieval call binding the contract method 0xf851a440.
//
// Solidity: function admin() view returns(address)
func (_Bridge *BridgeCallerSession) Admin() (common.Address, error) {
	return _Bridge.Contract.Admin(&_Bridge.CallOpts)
}

// BatchSettleBlockCount is a free data retrieval call binding the contract method 0x4ab3867f.
//
// Solidity: function batchSettleBlockCount() view returns(uint256)
func (_Bridge *BridgeCaller) BatchSettleBlockCount(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Bridge.contract.Call(opts, &out, "batchSettleBlockCount")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// BatchSettleBlockCount is a free data retrieval call binding the contract method 0x4ab3867f.
//
// Solidity: function batchSettleBlockCount() view returns(uint256)
func (_Bridge *BridgeSession) BatchSettleBlockCount() (*big.Int, error) {
	return _Bridge.Contract.BatchSettleBlockCount(&_Bridge.CallOpts)
}

// BatchSettleBlockCount is a free data retrieval call binding the contract method 0x4ab3867f.
//
// Solidity: function batchSettleBlockCount() view returns(uint256)
func (_Bridge *BridgeCallerSession) BatchSettleBlockCount() (*big.Int, error) {
	return _Bridge.Contract.BatchSettleBlockCount(&_Bridge.CallOpts)
}

// CrossTransferStatuses is a free data retrieval call binding the contract method 0xb2c79ca3.
//
// Solidity: function crossTransferStatuses(uint256 ) view returns(uint256 createdBlockNumber)
func (_Bridge *BridgeCaller) CrossTransferStatuses(opts *bind.CallOpts, arg0 *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _Bridge.contract.Call(opts, &out, "crossTransferStatuses", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// CrossTransferStatuses is a free data retrieval call binding the contract method 0xb2c79ca3.
//
// Solidity: function crossTransferStatuses(uint256 ) view returns(uint256 createdBlockNumber)
func (_Bridge *BridgeSession) CrossTransferStatuses(arg0 *big.Int) (*big.Int, error) {
	return _Bridge.Contract.CrossTransferStatuses(&_Bridge.CallOpts, arg0)
}

// CrossTransferStatuses is a free data retrieval call binding the contract method 0xb2c79ca3.
//
// Solidity: function crossTransferStatuses(uint256 ) view returns(uint256 createdBlockNumber)
func (_Bridge *BridgeCallerSession) CrossTransferStatuses(arg0 *big.Int) (*big.Int, error) {
	return _Bridge.Contract.CrossTransferStatuses(&_Bridge.CallOpts, arg0)
}

// ExecutedBatches is a free data retrieval call binding the contract method 0x7039e21b.
//
// Solidity: function executedBatches(uint256 ) view returns(bool)
func (_Bridge *BridgeCaller) ExecutedBatches(opts *bind.CallOpts, arg0 *big.Int) (bool, error) {
	var out []interface{}
	err := _Bridge.contract.Call(opts, &out, "executedBatches", arg0)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// ExecutedBatches is a free data retrieval call binding the contract method 0x7039e21b.
//
// Solidity: function executedBatches(uint256 ) view returns(bool)
func (_Bridge *BridgeSession) ExecutedBatches(arg0 *big.Int) (bool, error) {
	return _Bridge.Contract.ExecutedBatches(&_Bridge.CallOpts, arg0)
}

// ExecutedBatches is a free data retrieval call binding the contract method 0x7039e21b.
//
// Solidity: function executedBatches(uint256 ) view returns(bool)
func (_Bridge *BridgeCallerSession) ExecutedBatches(arg0 *big.Int) (bool, error) {
	return _Bridge.Contract.ExecutedBatches(&_Bridge.CallOpts, arg0)
}

// GetBatch is a free data retrieval call binding the contract method 0x5ac44282.
//
// Solidity: function getBatch(uint256 batchNonce) view returns((uint256,uint256,uint256,uint8))
func (_Bridge *BridgeCaller) GetBatch(opts *bind.CallOpts, batchNonce *big.Int) (Batch, error) {
	var out []interface{}
	err := _Bridge.contract.Call(opts, &out, "getBatch", batchNonce)

	if err != nil {
		return *new(Batch), err
	}

	out0 := *abi.ConvertType(out[0], new(Batch)).(*Batch)

	return out0, err

}

// GetBatch is a free data retrieval call binding the contract method 0x5ac44282.
//
// Solidity: function getBatch(uint256 batchNonce) view returns((uint256,uint256,uint256,uint8))
func (_Bridge *BridgeSession) GetBatch(batchNonce *big.Int) (Batch, error) {
	return _Bridge.Contract.GetBatch(&_Bridge.CallOpts, batchNonce)
}

// GetBatch is a free data retrieval call binding the contract method 0x5ac44282.
//
// Solidity: function getBatch(uint256 batchNonce) view returns((uint256,uint256,uint256,uint8))
func (_Bridge *BridgeCallerSession) GetBatch(batchNonce *big.Int) (Batch, error) {
	return _Bridge.Contract.GetBatch(&_Bridge.CallOpts, batchNonce)
}

// GetBatchDeposits is a free data retrieval call binding the contract method 0x90924da7.
//
// Solidity: function getBatchDeposits(uint256 batchNonce) view returns((uint256,address,uint256,address,bytes32,uint8)[])
func (_Bridge *BridgeCaller) GetBatchDeposits(opts *bind.CallOpts, batchNonce *big.Int) ([]Deposit, error) {
	var out []interface{}
	err := _Bridge.contract.Call(opts, &out, "getBatchDeposits", batchNonce)

	if err != nil {
		return *new([]Deposit), err
	}

	out0 := *abi.ConvertType(out[0], new([]Deposit)).(*[]Deposit)

	return out0, err

}

// GetBatchDeposits is a free data retrieval call binding the contract method 0x90924da7.
//
// Solidity: function getBatchDeposits(uint256 batchNonce) view returns((uint256,address,uint256,address,bytes32,uint8)[])
func (_Bridge *BridgeSession) GetBatchDeposits(batchNonce *big.Int) ([]Deposit, error) {
	return _Bridge.Contract.GetBatchDeposits(&_Bridge.CallOpts, batchNonce)
}

// GetBatchDeposits is a free data retrieval call binding the contract method 0x90924da7.
//
// Solidity: function getBatchDeposits(uint256 batchNonce) view returns((uint256,address,uint256,address,bytes32,uint8)[])
func (_Bridge *BridgeCallerSession) GetBatchDeposits(batchNonce *big.Int) ([]Deposit, error) {
	return _Bridge.Contract.GetBatchDeposits(&_Bridge.CallOpts, batchNonce)
}

// GetRelayer is a free data retrieval call binding the contract method 0xbee2e4dd.
//
// Solidity: function getRelayer(uint256 index) view returns(address)
func (_Bridge *BridgeCaller) GetRelayer(opts *bind.CallOpts, index *big.Int) (common.Address, error) {
	var out []interface{}
	err := _Bridge.contract.Call(opts, &out, "getRelayer", index)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetRelayer is a free data retrieval call binding the contract method 0xbee2e4dd.
//
// Solidity: function getRelayer(uint256 index) view returns(address)
func (_Bridge *BridgeSession) GetRelayer(index *big.Int) (common.Address, error) {
	return _Bridge.Contract.GetRelayer(&_Bridge.CallOpts, index)
}

// GetRelayer is a free data retrieval call binding the contract method 0xbee2e4dd.
//
// Solidity: function getRelayer(uint256 index) view returns(address)
func (_Bridge *BridgeCallerSession) GetRelayer(index *big.Int) (common.Address, error) {
	return _Bridge.Contract.GetRelayer(&_Bridge.CallOpts, index)
}

// GetRelayers is a free data retrieval call binding the contract method 0x179ff4b2.
//
// Solidity: function getRelayers() view returns(address[])
func (_Bridge *BridgeCaller) GetRelayers(opts *bind.CallOpts) ([]common.Address, error) {
	var out []interface{}
	err := _Bridge.contract.Call(opts, &out, "getRelayers")

	if err != nil {
		return *new([]common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new([]common.Address)).(*[]common.Address)

	return out0, err

}

// GetRelayers is a free data retrieval call binding the contract method 0x179ff4b2.
//
// Solidity: function getRelayers() view returns(address[])
func (_Bridge *BridgeSession) GetRelayers() ([]common.Address, error) {
	return _Bridge.Contract.GetRelayers(&_Bridge.CallOpts)
}

// GetRelayers is a free data retrieval call binding the contract method 0x179ff4b2.
//
// Solidity: function getRelayers() view returns(address[])
func (_Bridge *BridgeCallerSession) GetRelayers() ([]common.Address, error) {
	return _Bridge.Contract.GetRelayers(&_Bridge.CallOpts)
}

// GetRelayersCount is a free data retrieval call binding the contract method 0xd3d9ec01.
//
// Solidity: function getRelayersCount() view returns(uint256)
func (_Bridge *BridgeCaller) GetRelayersCount(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Bridge.contract.Call(opts, &out, "getRelayersCount")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetRelayersCount is a free data retrieval call binding the contract method 0xd3d9ec01.
//
// Solidity: function getRelayersCount() view returns(uint256)
func (_Bridge *BridgeSession) GetRelayersCount() (*big.Int, error) {
	return _Bridge.Contract.GetRelayersCount(&_Bridge.CallOpts)
}

// GetRelayersCount is a free data retrieval call binding the contract method 0xd3d9ec01.
//
// Solidity: function getRelayersCount() view returns(uint256)
func (_Bridge *BridgeCallerSession) GetRelayersCount() (*big.Int, error) {
	return _Bridge.Contract.GetRelayersCount(&_Bridge.CallOpts)
}

// GetStatusesAfterExecution is a free data retrieval call binding the contract method 0xdb626c2d.
//
// Solidity: function getStatusesAfterExecution(uint256 batchNonceElrondETH) view returns(uint8[])
func (_Bridge *BridgeCaller) GetStatusesAfterExecution(opts *bind.CallOpts, batchNonceElrondETH *big.Int) ([]uint8, error) {
	var out []interface{}
	err := _Bridge.contract.Call(opts, &out, "getStatusesAfterExecution", batchNonceElrondETH)

	if err != nil {
		return *new([]uint8), err
	}

	out0 := *abi.ConvertType(out[0], new([]uint8)).(*[]uint8)

	return out0, err

}

// GetStatusesAfterExecution is a free data retrieval call binding the contract method 0xdb626c2d.
//
// Solidity: function getStatusesAfterExecution(uint256 batchNonceElrondETH) view returns(uint8[])
func (_Bridge *BridgeSession) GetStatusesAfterExecution(batchNonceElrondETH *big.Int) ([]uint8, error) {
	return _Bridge.Contract.GetStatusesAfterExecution(&_Bridge.CallOpts, batchNonceElrondETH)
}

// GetStatusesAfterExecution is a free data retrieval call binding the contract method 0xdb626c2d.
//
// Solidity: function getStatusesAfterExecution(uint256 batchNonceElrondETH) view returns(uint8[])
func (_Bridge *BridgeCallerSession) GetStatusesAfterExecution(batchNonceElrondETH *big.Int) ([]uint8, error) {
	return _Bridge.Contract.GetStatusesAfterExecution(&_Bridge.CallOpts, batchNonceElrondETH)
}

// IsRelayer is a free data retrieval call binding the contract method 0x541d5548.
//
// Solidity: function isRelayer(address account) view returns(bool)
func (_Bridge *BridgeCaller) IsRelayer(opts *bind.CallOpts, account common.Address) (bool, error) {
	var out []interface{}
	err := _Bridge.contract.Call(opts, &out, "isRelayer", account)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsRelayer is a free data retrieval call binding the contract method 0x541d5548.
//
// Solidity: function isRelayer(address account) view returns(bool)
func (_Bridge *BridgeSession) IsRelayer(account common.Address) (bool, error) {
	return _Bridge.Contract.IsRelayer(&_Bridge.CallOpts, account)
}

// IsRelayer is a free data retrieval call binding the contract method 0x541d5548.
//
// Solidity: function isRelayer(address account) view returns(bool)
func (_Bridge *BridgeCallerSession) IsRelayer(account common.Address) (bool, error) {
	return _Bridge.Contract.IsRelayer(&_Bridge.CallOpts, account)
}

// Quorum is a free data retrieval call binding the contract method 0x1703a018.
//
// Solidity: function quorum() view returns(uint256)
func (_Bridge *BridgeCaller) Quorum(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Bridge.contract.Call(opts, &out, "quorum")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Quorum is a free data retrieval call binding the contract method 0x1703a018.
//
// Solidity: function quorum() view returns(uint256)
func (_Bridge *BridgeSession) Quorum() (*big.Int, error) {
	return _Bridge.Contract.Quorum(&_Bridge.CallOpts)
}

// Quorum is a free data retrieval call binding the contract method 0x1703a018.
//
// Solidity: function quorum() view returns(uint256)
func (_Bridge *BridgeCallerSession) Quorum() (*big.Int, error) {
	return _Bridge.Contract.Quorum(&_Bridge.CallOpts)
}

// WasBatchExecuted is a free data retrieval call binding the contract method 0x84aa1ad0.
//
// Solidity: function wasBatchExecuted(uint256 batchNonceElrondETH) view returns(bool)
func (_Bridge *BridgeCaller) WasBatchExecuted(opts *bind.CallOpts, batchNonceElrondETH *big.Int) (bool, error) {
	var out []interface{}
	err := _Bridge.contract.Call(opts, &out, "wasBatchExecuted", batchNonceElrondETH)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// WasBatchExecuted is a free data retrieval call binding the contract method 0x84aa1ad0.
//
// Solidity: function wasBatchExecuted(uint256 batchNonceElrondETH) view returns(bool)
func (_Bridge *BridgeSession) WasBatchExecuted(batchNonceElrondETH *big.Int) (bool, error) {
	return _Bridge.Contract.WasBatchExecuted(&_Bridge.CallOpts, batchNonceElrondETH)
}

// WasBatchExecuted is a free data retrieval call binding the contract method 0x84aa1ad0.
//
// Solidity: function wasBatchExecuted(uint256 batchNonceElrondETH) view returns(bool)
func (_Bridge *BridgeCallerSession) WasBatchExecuted(batchNonceElrondETH *big.Int) (bool, error) {
	return _Bridge.Contract.WasBatchExecuted(&_Bridge.CallOpts, batchNonceElrondETH)
}

// AddRelayer is a paid mutator transaction binding the contract method 0xdd39f00d.
//
// Solidity: function addRelayer(address account) returns()
func (_Bridge *BridgeTransactor) AddRelayer(opts *bind.TransactOpts, account common.Address) (*types.Transaction, error) {
	return _Bridge.contract.Transact(opts, "addRelayer", account)
}

// AddRelayer is a paid mutator transaction binding the contract method 0xdd39f00d.
//
// Solidity: function addRelayer(address account) returns()
func (_Bridge *BridgeSession) AddRelayer(account common.Address) (*types.Transaction, error) {
	return _Bridge.Contract.AddRelayer(&_Bridge.TransactOpts, account)
}

// AddRelayer is a paid mutator transaction binding the contract method 0xdd39f00d.
//
// Solidity: function addRelayer(address account) returns()
func (_Bridge *BridgeTransactorSession) AddRelayer(account common.Address) (*types.Transaction, error) {
	return _Bridge.Contract.AddRelayer(&_Bridge.TransactOpts, account)
}

// ExecuteTransfer is a paid mutator transaction binding the contract method 0x51db0518.
//
// Solidity: function executeTransfer(address[] tokens, address[] recipients, uint256[] amounts, uint256[] depositNonces, uint256 batchNonceElrondETH, bytes[] signatures) returns()
func (_Bridge *BridgeTransactor) ExecuteTransfer(opts *bind.TransactOpts, tokens []common.Address, recipients []common.Address, amounts []*big.Int, depositNonces []*big.Int, batchNonceElrondETH *big.Int, signatures [][]byte) (*types.Transaction, error) {
	return _Bridge.contract.Transact(opts, "executeTransfer", tokens, recipients, amounts, depositNonces, batchNonceElrondETH, signatures)
}

// ExecuteTransfer is a paid mutator transaction binding the contract method 0x51db0518.
//
// Solidity: function executeTransfer(address[] tokens, address[] recipients, uint256[] amounts, uint256[] depositNonces, uint256 batchNonceElrondETH, bytes[] signatures) returns()
func (_Bridge *BridgeSession) ExecuteTransfer(tokens []common.Address, recipients []common.Address, amounts []*big.Int, depositNonces []*big.Int, batchNonceElrondETH *big.Int, signatures [][]byte) (*types.Transaction, error) {
	return _Bridge.Contract.ExecuteTransfer(&_Bridge.TransactOpts, tokens, recipients, amounts, depositNonces, batchNonceElrondETH, signatures)
}

// ExecuteTransfer is a paid mutator transaction binding the contract method 0x51db0518.
//
// Solidity: function executeTransfer(address[] tokens, address[] recipients, uint256[] amounts, uint256[] depositNonces, uint256 batchNonceElrondETH, bytes[] signatures) returns()
func (_Bridge *BridgeTransactorSession) ExecuteTransfer(tokens []common.Address, recipients []common.Address, amounts []*big.Int, depositNonces []*big.Int, batchNonceElrondETH *big.Int, signatures [][]byte) (*types.Transaction, error) {
	return _Bridge.Contract.ExecuteTransfer(&_Bridge.TransactOpts, tokens, recipients, amounts, depositNonces, batchNonceElrondETH, signatures)
}

// RemoveRelayer is a paid mutator transaction binding the contract method 0x60f0a5ac.
//
// Solidity: function removeRelayer(address account) returns()
func (_Bridge *BridgeTransactor) RemoveRelayer(opts *bind.TransactOpts, account common.Address) (*types.Transaction, error) {
	return _Bridge.contract.Transact(opts, "removeRelayer", account)
}

// RemoveRelayer is a paid mutator transaction binding the contract method 0x60f0a5ac.
//
// Solidity: function removeRelayer(address account) returns()
func (_Bridge *BridgeSession) RemoveRelayer(account common.Address) (*types.Transaction, error) {
	return _Bridge.Contract.RemoveRelayer(&_Bridge.TransactOpts, account)
}

// RemoveRelayer is a paid mutator transaction binding the contract method 0x60f0a5ac.
//
// Solidity: function removeRelayer(address account) returns()
func (_Bridge *BridgeTransactorSession) RemoveRelayer(account common.Address) (*types.Transaction, error) {
	return _Bridge.Contract.RemoveRelayer(&_Bridge.TransactOpts, account)
}

// RenounceAdmin is a paid mutator transaction binding the contract method 0x8bad0c0a.
//
// Solidity: function renounceAdmin() returns()
func (_Bridge *BridgeTransactor) RenounceAdmin(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Bridge.contract.Transact(opts, "renounceAdmin")
}

// RenounceAdmin is a paid mutator transaction binding the contract method 0x8bad0c0a.
//
// Solidity: function renounceAdmin() returns()
func (_Bridge *BridgeSession) RenounceAdmin() (*types.Transaction, error) {
	return _Bridge.Contract.RenounceAdmin(&_Bridge.TransactOpts)
}

// RenounceAdmin is a paid mutator transaction binding the contract method 0x8bad0c0a.
//
// Solidity: function renounceAdmin() returns()
func (_Bridge *BridgeTransactorSession) RenounceAdmin() (*types.Transaction, error) {
	return _Bridge.Contract.RenounceAdmin(&_Bridge.TransactOpts)
}

// RenounceRelayer is a paid mutator transaction binding the contract method 0x475ed4d0.
//
// Solidity: function renounceRelayer(address account) returns()
func (_Bridge *BridgeTransactor) RenounceRelayer(opts *bind.TransactOpts, account common.Address) (*types.Transaction, error) {
	return _Bridge.contract.Transact(opts, "renounceRelayer", account)
}

// RenounceRelayer is a paid mutator transaction binding the contract method 0x475ed4d0.
//
// Solidity: function renounceRelayer(address account) returns()
func (_Bridge *BridgeSession) RenounceRelayer(account common.Address) (*types.Transaction, error) {
	return _Bridge.Contract.RenounceRelayer(&_Bridge.TransactOpts, account)
}

// RenounceRelayer is a paid mutator transaction binding the contract method 0x475ed4d0.
//
// Solidity: function renounceRelayer(address account) returns()
func (_Bridge *BridgeTransactorSession) RenounceRelayer(account common.Address) (*types.Transaction, error) {
	return _Bridge.Contract.RenounceRelayer(&_Bridge.TransactOpts, account)
}

// SetQuorum is a paid mutator transaction binding the contract method 0xc1ba4e59.
//
// Solidity: function setQuorum(uint256 newQuorum) returns()
func (_Bridge *BridgeTransactor) SetQuorum(opts *bind.TransactOpts, newQuorum *big.Int) (*types.Transaction, error) {
	return _Bridge.contract.Transact(opts, "setQuorum", newQuorum)
}

// SetQuorum is a paid mutator transaction binding the contract method 0xc1ba4e59.
//
// Solidity: function setQuorum(uint256 newQuorum) returns()
func (_Bridge *BridgeSession) SetQuorum(newQuorum *big.Int) (*types.Transaction, error) {
	return _Bridge.Contract.SetQuorum(&_Bridge.TransactOpts, newQuorum)
}

// SetQuorum is a paid mutator transaction binding the contract method 0xc1ba4e59.
//
// Solidity: function setQuorum(uint256 newQuorum) returns()
func (_Bridge *BridgeTransactorSession) SetQuorum(newQuorum *big.Int) (*types.Transaction, error) {
	return _Bridge.Contract.SetQuorum(&_Bridge.TransactOpts, newQuorum)
}

// TransferAdmin is a paid mutator transaction binding the contract method 0x75829def.
//
// Solidity: function transferAdmin(address newAdmin) returns()
func (_Bridge *BridgeTransactor) TransferAdmin(opts *bind.TransactOpts, newAdmin common.Address) (*types.Transaction, error) {
	return _Bridge.contract.Transact(opts, "transferAdmin", newAdmin)
}

// TransferAdmin is a paid mutator transaction binding the contract method 0x75829def.
//
// Solidity: function transferAdmin(address newAdmin) returns()
func (_Bridge *BridgeSession) TransferAdmin(newAdmin common.Address) (*types.Transaction, error) {
	return _Bridge.Contract.TransferAdmin(&_Bridge.TransactOpts, newAdmin)
}

// TransferAdmin is a paid mutator transaction binding the contract method 0x75829def.
//
// Solidity: function transferAdmin(address newAdmin) returns()
func (_Bridge *BridgeTransactorSession) TransferAdmin(newAdmin common.Address) (*types.Transaction, error) {
	return _Bridge.Contract.TransferAdmin(&_Bridge.TransactOpts, newAdmin)
}

// BridgeAdminRoleTransferredIterator is returned from FilterAdminRoleTransferred and is used to iterate over the raw logs and unpacked data for AdminRoleTransferred events raised by the Bridge contract.
type BridgeAdminRoleTransferredIterator struct {
	Event *BridgeAdminRoleTransferred // Event containing the contract specifics and raw log

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
func (it *BridgeAdminRoleTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BridgeAdminRoleTransferred)
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
		it.Event = new(BridgeAdminRoleTransferred)
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
func (it *BridgeAdminRoleTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BridgeAdminRoleTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BridgeAdminRoleTransferred represents a AdminRoleTransferred event raised by the Bridge contract.
type BridgeAdminRoleTransferred struct {
	PreviousAdmin common.Address
	NewAdmin      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterAdminRoleTransferred is a free log retrieval operation binding the contract event 0xe379ac64de02d8184ca1a871ac486cb8137de77e485ede140e97057b9c765ffd.
//
// Solidity: event AdminRoleTransferred(address indexed previousAdmin, address indexed newAdmin)
func (_Bridge *BridgeFilterer) FilterAdminRoleTransferred(opts *bind.FilterOpts, previousAdmin []common.Address, newAdmin []common.Address) (*BridgeAdminRoleTransferredIterator, error) {

	var previousAdminRule []interface{}
	for _, previousAdminItem := range previousAdmin {
		previousAdminRule = append(previousAdminRule, previousAdminItem)
	}
	var newAdminRule []interface{}
	for _, newAdminItem := range newAdmin {
		newAdminRule = append(newAdminRule, newAdminItem)
	}

	logs, sub, err := _Bridge.contract.FilterLogs(opts, "AdminRoleTransferred", previousAdminRule, newAdminRule)
	if err != nil {
		return nil, err
	}
	return &BridgeAdminRoleTransferredIterator{contract: _Bridge.contract, event: "AdminRoleTransferred", logs: logs, sub: sub}, nil
}

// WatchAdminRoleTransferred is a free log subscription operation binding the contract event 0xe379ac64de02d8184ca1a871ac486cb8137de77e485ede140e97057b9c765ffd.
//
// Solidity: event AdminRoleTransferred(address indexed previousAdmin, address indexed newAdmin)
func (_Bridge *BridgeFilterer) WatchAdminRoleTransferred(opts *bind.WatchOpts, sink chan<- *BridgeAdminRoleTransferred, previousAdmin []common.Address, newAdmin []common.Address) (event.Subscription, error) {

	var previousAdminRule []interface{}
	for _, previousAdminItem := range previousAdmin {
		previousAdminRule = append(previousAdminRule, previousAdminItem)
	}
	var newAdminRule []interface{}
	for _, newAdminItem := range newAdmin {
		newAdminRule = append(newAdminRule, newAdminItem)
	}

	logs, sub, err := _Bridge.contract.WatchLogs(opts, "AdminRoleTransferred", previousAdminRule, newAdminRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BridgeAdminRoleTransferred)
				if err := _Bridge.contract.UnpackLog(event, "AdminRoleTransferred", log); err != nil {
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
func (_Bridge *BridgeFilterer) ParseAdminRoleTransferred(log types.Log) (*BridgeAdminRoleTransferred, error) {
	event := new(BridgeAdminRoleTransferred)
	if err := _Bridge.contract.UnpackLog(event, "AdminRoleTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BridgeQuorumChangedIterator is returned from FilterQuorumChanged and is used to iterate over the raw logs and unpacked data for QuorumChanged events raised by the Bridge contract.
type BridgeQuorumChangedIterator struct {
	Event *BridgeQuorumChanged // Event containing the contract specifics and raw log

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
func (it *BridgeQuorumChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BridgeQuorumChanged)
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
		it.Event = new(BridgeQuorumChanged)
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
func (it *BridgeQuorumChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BridgeQuorumChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BridgeQuorumChanged represents a QuorumChanged event raised by the Bridge contract.
type BridgeQuorumChanged struct {
	Quorum *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterQuorumChanged is a free log retrieval operation binding the contract event 0x027863d12a407097e086a48e36475bfc859d0b200b7e6f65b5fd3b218e46632e.
//
// Solidity: event QuorumChanged(uint256 quorum)
func (_Bridge *BridgeFilterer) FilterQuorumChanged(opts *bind.FilterOpts) (*BridgeQuorumChangedIterator, error) {

	logs, sub, err := _Bridge.contract.FilterLogs(opts, "QuorumChanged")
	if err != nil {
		return nil, err
	}
	return &BridgeQuorumChangedIterator{contract: _Bridge.contract, event: "QuorumChanged", logs: logs, sub: sub}, nil
}

// WatchQuorumChanged is a free log subscription operation binding the contract event 0x027863d12a407097e086a48e36475bfc859d0b200b7e6f65b5fd3b218e46632e.
//
// Solidity: event QuorumChanged(uint256 quorum)
func (_Bridge *BridgeFilterer) WatchQuorumChanged(opts *bind.WatchOpts, sink chan<- *BridgeQuorumChanged) (event.Subscription, error) {

	logs, sub, err := _Bridge.contract.WatchLogs(opts, "QuorumChanged")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BridgeQuorumChanged)
				if err := _Bridge.contract.UnpackLog(event, "QuorumChanged", log); err != nil {
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

// ParseQuorumChanged is a log parse operation binding the contract event 0x027863d12a407097e086a48e36475bfc859d0b200b7e6f65b5fd3b218e46632e.
//
// Solidity: event QuorumChanged(uint256 quorum)
func (_Bridge *BridgeFilterer) ParseQuorumChanged(log types.Log) (*BridgeQuorumChanged, error) {
	event := new(BridgeQuorumChanged)
	if err := _Bridge.contract.UnpackLog(event, "QuorumChanged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BridgeRelayerAddedIterator is returned from FilterRelayerAdded and is used to iterate over the raw logs and unpacked data for RelayerAdded events raised by the Bridge contract.
type BridgeRelayerAddedIterator struct {
	Event *BridgeRelayerAdded // Event containing the contract specifics and raw log

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
func (it *BridgeRelayerAddedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BridgeRelayerAdded)
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
		it.Event = new(BridgeRelayerAdded)
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
func (it *BridgeRelayerAddedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BridgeRelayerAddedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BridgeRelayerAdded represents a RelayerAdded event raised by the Bridge contract.
type BridgeRelayerAdded struct {
	Account common.Address
	Sender  common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterRelayerAdded is a free log retrieval operation binding the contract event 0xd756b9aee10d6f2c80dc42c5031beb0e0847f6e1d6ba50199bdfc3f0de5cc0cc.
//
// Solidity: event RelayerAdded(address indexed account, address indexed sender)
func (_Bridge *BridgeFilterer) FilterRelayerAdded(opts *bind.FilterOpts, account []common.Address, sender []common.Address) (*BridgeRelayerAddedIterator, error) {

	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}
	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}

	logs, sub, err := _Bridge.contract.FilterLogs(opts, "RelayerAdded", accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return &BridgeRelayerAddedIterator{contract: _Bridge.contract, event: "RelayerAdded", logs: logs, sub: sub}, nil
}

// WatchRelayerAdded is a free log subscription operation binding the contract event 0xd756b9aee10d6f2c80dc42c5031beb0e0847f6e1d6ba50199bdfc3f0de5cc0cc.
//
// Solidity: event RelayerAdded(address indexed account, address indexed sender)
func (_Bridge *BridgeFilterer) WatchRelayerAdded(opts *bind.WatchOpts, sink chan<- *BridgeRelayerAdded, account []common.Address, sender []common.Address) (event.Subscription, error) {

	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}
	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}

	logs, sub, err := _Bridge.contract.WatchLogs(opts, "RelayerAdded", accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BridgeRelayerAdded)
				if err := _Bridge.contract.UnpackLog(event, "RelayerAdded", log); err != nil {
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

// ParseRelayerAdded is a log parse operation binding the contract event 0xd756b9aee10d6f2c80dc42c5031beb0e0847f6e1d6ba50199bdfc3f0de5cc0cc.
//
// Solidity: event RelayerAdded(address indexed account, address indexed sender)
func (_Bridge *BridgeFilterer) ParseRelayerAdded(log types.Log) (*BridgeRelayerAdded, error) {
	event := new(BridgeRelayerAdded)
	if err := _Bridge.contract.UnpackLog(event, "RelayerAdded", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BridgeRelayerRemovedIterator is returned from FilterRelayerRemoved and is used to iterate over the raw logs and unpacked data for RelayerRemoved events raised by the Bridge contract.
type BridgeRelayerRemovedIterator struct {
	Event *BridgeRelayerRemoved // Event containing the contract specifics and raw log

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
func (it *BridgeRelayerRemovedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BridgeRelayerRemoved)
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
		it.Event = new(BridgeRelayerRemoved)
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
func (it *BridgeRelayerRemovedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BridgeRelayerRemovedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BridgeRelayerRemoved represents a RelayerRemoved event raised by the Bridge contract.
type BridgeRelayerRemoved struct {
	Account common.Address
	Sender  common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterRelayerRemoved is a free log retrieval operation binding the contract event 0x0bdcf1d6f29aa87af8131cc81dcbb295fcf98d71cfcdc79cc5d965317bae1d0a.
//
// Solidity: event RelayerRemoved(address indexed account, address indexed sender)
func (_Bridge *BridgeFilterer) FilterRelayerRemoved(opts *bind.FilterOpts, account []common.Address, sender []common.Address) (*BridgeRelayerRemovedIterator, error) {

	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}
	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}

	logs, sub, err := _Bridge.contract.FilterLogs(opts, "RelayerRemoved", accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return &BridgeRelayerRemovedIterator{contract: _Bridge.contract, event: "RelayerRemoved", logs: logs, sub: sub}, nil
}

// WatchRelayerRemoved is a free log subscription operation binding the contract event 0x0bdcf1d6f29aa87af8131cc81dcbb295fcf98d71cfcdc79cc5d965317bae1d0a.
//
// Solidity: event RelayerRemoved(address indexed account, address indexed sender)
func (_Bridge *BridgeFilterer) WatchRelayerRemoved(opts *bind.WatchOpts, sink chan<- *BridgeRelayerRemoved, account []common.Address, sender []common.Address) (event.Subscription, error) {

	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}
	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}

	logs, sub, err := _Bridge.contract.WatchLogs(opts, "RelayerRemoved", accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BridgeRelayerRemoved)
				if err := _Bridge.contract.UnpackLog(event, "RelayerRemoved", log); err != nil {
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

// ParseRelayerRemoved is a log parse operation binding the contract event 0x0bdcf1d6f29aa87af8131cc81dcbb295fcf98d71cfcdc79cc5d965317bae1d0a.
//
// Solidity: event RelayerRemoved(address indexed account, address indexed sender)
func (_Bridge *BridgeFilterer) ParseRelayerRemoved(log types.Log) (*BridgeRelayerRemoved, error) {
	event := new(BridgeRelayerRemoved)
	if err := _Bridge.contract.UnpackLog(event, "RelayerRemoved", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
