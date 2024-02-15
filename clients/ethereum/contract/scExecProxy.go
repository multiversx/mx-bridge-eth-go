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

// SCExecProxyMetaData contains all meta data concerning the SCExecProxy contract.
var SCExecProxyMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"contractERC20Safe\",\"name\":\"erc20Safe\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"target\",\"type\":\"address\"}],\"name\":\"AddressEmptyCode\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"AddressInsufficientBalance\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"FailedInnerCall\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"}],\"name\":\"SafeERC20FailedOperation\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousAdmin\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newAdmin\",\"type\":\"address\"}],\"name\":\"AdminRoleTransferred\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint64\",\"name\":\"batchNonce\",\"type\":\"uint64\"},{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"depositNonce\",\"type\":\"uint64\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"callData\",\"type\":\"string\"}],\"name\":\"ERC20SCDeposit\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"admin\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"tokenAddress\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"bytes32\",\"name\":\"recipientAddress\",\"type\":\"bytes32\"},{\"internalType\":\"string\",\"name\":\"callData\",\"type\":\"string\"}],\"name\":\"deposit\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"isSafePaused\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceAdmin\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"safe\",\"outputs\":[{\"internalType\":\"contractERC20Safe\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"contractERC20Safe\",\"name\":\"erc20Safe\",\"type\":\"address\"}],\"name\":\"setSafe\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newAdmin\",\"type\":\"address\"}],\"name\":\"transferAdmin\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
}

// SCExecProxyABI is the input ABI used to generate the binding from.
// Deprecated: Use SCExecProxyMetaData.ABI instead.
var SCExecProxyABI = SCExecProxyMetaData.ABI

// SCExecProxy is an auto generated Go binding around an Ethereum contract.
type SCExecProxy struct {
	SCExecProxyCaller     // Read-only binding to the contract
	SCExecProxyTransactor // Write-only binding to the contract
	SCExecProxyFilterer   // Log filterer for contract events
}

// SCExecProxyCaller is an auto generated read-only Go binding around an Ethereum contract.
type SCExecProxyCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SCExecProxyTransactor is an auto generated write-only Go binding around an Ethereum contract.
type SCExecProxyTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SCExecProxyFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type SCExecProxyFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SCExecProxySession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type SCExecProxySession struct {
	Contract     *SCExecProxy      // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// SCExecProxyCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type SCExecProxyCallerSession struct {
	Contract *SCExecProxyCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts      // Call options to use throughout this session
}

// SCExecProxyTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type SCExecProxyTransactorSession struct {
	Contract     *SCExecProxyTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts      // Transaction auth options to use throughout this session
}

// SCExecProxyRaw is an auto generated low-level Go binding around an Ethereum contract.
type SCExecProxyRaw struct {
	Contract *SCExecProxy // Generic contract binding to access the raw methods on
}

// SCExecProxyCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type SCExecProxyCallerRaw struct {
	Contract *SCExecProxyCaller // Generic read-only contract binding to access the raw methods on
}

// SCExecProxyTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type SCExecProxyTransactorRaw struct {
	Contract *SCExecProxyTransactor // Generic write-only contract binding to access the raw methods on
}

// NewSCExecProxy creates a new instance of SCExecProxy, bound to a specific deployed contract.
func NewSCExecProxy(address common.Address, backend bind.ContractBackend) (*SCExecProxy, error) {
	contract, err := bindSCExecProxy(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &SCExecProxy{SCExecProxyCaller: SCExecProxyCaller{contract: contract}, SCExecProxyTransactor: SCExecProxyTransactor{contract: contract}, SCExecProxyFilterer: SCExecProxyFilterer{contract: contract}}, nil
}

// NewSCExecProxyCaller creates a new read-only instance of SCExecProxy, bound to a specific deployed contract.
func NewSCExecProxyCaller(address common.Address, caller bind.ContractCaller) (*SCExecProxyCaller, error) {
	contract, err := bindSCExecProxy(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &SCExecProxyCaller{contract: contract}, nil
}

// NewSCExecProxyTransactor creates a new write-only instance of SCExecProxy, bound to a specific deployed contract.
func NewSCExecProxyTransactor(address common.Address, transactor bind.ContractTransactor) (*SCExecProxyTransactor, error) {
	contract, err := bindSCExecProxy(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &SCExecProxyTransactor{contract: contract}, nil
}

// NewSCExecProxyFilterer creates a new log filterer instance of SCExecProxy, bound to a specific deployed contract.
func NewSCExecProxyFilterer(address common.Address, filterer bind.ContractFilterer) (*SCExecProxyFilterer, error) {
	contract, err := bindSCExecProxy(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &SCExecProxyFilterer{contract: contract}, nil
}

// bindSCExecProxy binds a generic wrapper to an already deployed contract.
func bindSCExecProxy(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := SCExecProxyMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_SCExecProxy *SCExecProxyRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _SCExecProxy.Contract.SCExecProxyCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_SCExecProxy *SCExecProxyRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SCExecProxy.Contract.SCExecProxyTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_SCExecProxy *SCExecProxyRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _SCExecProxy.Contract.SCExecProxyTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_SCExecProxy *SCExecProxyCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _SCExecProxy.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_SCExecProxy *SCExecProxyTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SCExecProxy.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_SCExecProxy *SCExecProxyTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _SCExecProxy.Contract.contract.Transact(opts, method, params...)
}

// Admin is a free data retrieval call binding the contract method 0xf851a440.
//
// Solidity: function admin() view returns(address)
func (_SCExecProxy *SCExecProxyCaller) Admin(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _SCExecProxy.contract.Call(opts, &out, "admin")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Admin is a free data retrieval call binding the contract method 0xf851a440.
//
// Solidity: function admin() view returns(address)
func (_SCExecProxy *SCExecProxySession) Admin() (common.Address, error) {
	return _SCExecProxy.Contract.Admin(&_SCExecProxy.CallOpts)
}

// Admin is a free data retrieval call binding the contract method 0xf851a440.
//
// Solidity: function admin() view returns(address)
func (_SCExecProxy *SCExecProxyCallerSession) Admin() (common.Address, error) {
	return _SCExecProxy.Contract.Admin(&_SCExecProxy.CallOpts)
}

// IsSafePaused is a free data retrieval call binding the contract method 0xa0579640.
//
// Solidity: function isSafePaused() view returns(bool)
func (_SCExecProxy *SCExecProxyCaller) IsSafePaused(opts *bind.CallOpts) (bool, error) {
	var out []interface{}
	err := _SCExecProxy.contract.Call(opts, &out, "isSafePaused")

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsSafePaused is a free data retrieval call binding the contract method 0xa0579640.
//
// Solidity: function isSafePaused() view returns(bool)
func (_SCExecProxy *SCExecProxySession) IsSafePaused() (bool, error) {
	return _SCExecProxy.Contract.IsSafePaused(&_SCExecProxy.CallOpts)
}

// IsSafePaused is a free data retrieval call binding the contract method 0xa0579640.
//
// Solidity: function isSafePaused() view returns(bool)
func (_SCExecProxy *SCExecProxyCallerSession) IsSafePaused() (bool, error) {
	return _SCExecProxy.Contract.IsSafePaused(&_SCExecProxy.CallOpts)
}

// Safe is a free data retrieval call binding the contract method 0x186f0354.
//
// Solidity: function safe() view returns(address)
func (_SCExecProxy *SCExecProxyCaller) Safe(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _SCExecProxy.contract.Call(opts, &out, "safe")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Safe is a free data retrieval call binding the contract method 0x186f0354.
//
// Solidity: function safe() view returns(address)
func (_SCExecProxy *SCExecProxySession) Safe() (common.Address, error) {
	return _SCExecProxy.Contract.Safe(&_SCExecProxy.CallOpts)
}

// Safe is a free data retrieval call binding the contract method 0x186f0354.
//
// Solidity: function safe() view returns(address)
func (_SCExecProxy *SCExecProxyCallerSession) Safe() (common.Address, error) {
	return _SCExecProxy.Contract.Safe(&_SCExecProxy.CallOpts)
}

// Deposit is a paid mutator transaction binding the contract method 0x7583e9fd.
//
// Solidity: function deposit(address tokenAddress, uint256 amount, bytes32 recipientAddress, string callData) returns()
func (_SCExecProxy *SCExecProxyTransactor) Deposit(opts *bind.TransactOpts, tokenAddress common.Address, amount *big.Int, recipientAddress [32]byte, callData string) (*types.Transaction, error) {
	return _SCExecProxy.contract.Transact(opts, "deposit", tokenAddress, amount, recipientAddress, callData)
}

// Deposit is a paid mutator transaction binding the contract method 0x7583e9fd.
//
// Solidity: function deposit(address tokenAddress, uint256 amount, bytes32 recipientAddress, string callData) returns()
func (_SCExecProxy *SCExecProxySession) Deposit(tokenAddress common.Address, amount *big.Int, recipientAddress [32]byte, callData string) (*types.Transaction, error) {
	return _SCExecProxy.Contract.Deposit(&_SCExecProxy.TransactOpts, tokenAddress, amount, recipientAddress, callData)
}

// Deposit is a paid mutator transaction binding the contract method 0x7583e9fd.
//
// Solidity: function deposit(address tokenAddress, uint256 amount, bytes32 recipientAddress, string callData) returns()
func (_SCExecProxy *SCExecProxyTransactorSession) Deposit(tokenAddress common.Address, amount *big.Int, recipientAddress [32]byte, callData string) (*types.Transaction, error) {
	return _SCExecProxy.Contract.Deposit(&_SCExecProxy.TransactOpts, tokenAddress, amount, recipientAddress, callData)
}

// RenounceAdmin is a paid mutator transaction binding the contract method 0x8bad0c0a.
//
// Solidity: function renounceAdmin() returns()
func (_SCExecProxy *SCExecProxyTransactor) RenounceAdmin(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SCExecProxy.contract.Transact(opts, "renounceAdmin")
}

// RenounceAdmin is a paid mutator transaction binding the contract method 0x8bad0c0a.
//
// Solidity: function renounceAdmin() returns()
func (_SCExecProxy *SCExecProxySession) RenounceAdmin() (*types.Transaction, error) {
	return _SCExecProxy.Contract.RenounceAdmin(&_SCExecProxy.TransactOpts)
}

// RenounceAdmin is a paid mutator transaction binding the contract method 0x8bad0c0a.
//
// Solidity: function renounceAdmin() returns()
func (_SCExecProxy *SCExecProxyTransactorSession) RenounceAdmin() (*types.Transaction, error) {
	return _SCExecProxy.Contract.RenounceAdmin(&_SCExecProxy.TransactOpts)
}

// SetSafe is a paid mutator transaction binding the contract method 0x5db0cb94.
//
// Solidity: function setSafe(address erc20Safe) returns()
func (_SCExecProxy *SCExecProxyTransactor) SetSafe(opts *bind.TransactOpts, erc20Safe common.Address) (*types.Transaction, error) {
	return _SCExecProxy.contract.Transact(opts, "setSafe", erc20Safe)
}

// SetSafe is a paid mutator transaction binding the contract method 0x5db0cb94.
//
// Solidity: function setSafe(address erc20Safe) returns()
func (_SCExecProxy *SCExecProxySession) SetSafe(erc20Safe common.Address) (*types.Transaction, error) {
	return _SCExecProxy.Contract.SetSafe(&_SCExecProxy.TransactOpts, erc20Safe)
}

// SetSafe is a paid mutator transaction binding the contract method 0x5db0cb94.
//
// Solidity: function setSafe(address erc20Safe) returns()
func (_SCExecProxy *SCExecProxyTransactorSession) SetSafe(erc20Safe common.Address) (*types.Transaction, error) {
	return _SCExecProxy.Contract.SetSafe(&_SCExecProxy.TransactOpts, erc20Safe)
}

// TransferAdmin is a paid mutator transaction binding the contract method 0x75829def.
//
// Solidity: function transferAdmin(address newAdmin) returns()
func (_SCExecProxy *SCExecProxyTransactor) TransferAdmin(opts *bind.TransactOpts, newAdmin common.Address) (*types.Transaction, error) {
	return _SCExecProxy.contract.Transact(opts, "transferAdmin", newAdmin)
}

// TransferAdmin is a paid mutator transaction binding the contract method 0x75829def.
//
// Solidity: function transferAdmin(address newAdmin) returns()
func (_SCExecProxy *SCExecProxySession) TransferAdmin(newAdmin common.Address) (*types.Transaction, error) {
	return _SCExecProxy.Contract.TransferAdmin(&_SCExecProxy.TransactOpts, newAdmin)
}

// TransferAdmin is a paid mutator transaction binding the contract method 0x75829def.
//
// Solidity: function transferAdmin(address newAdmin) returns()
func (_SCExecProxy *SCExecProxyTransactorSession) TransferAdmin(newAdmin common.Address) (*types.Transaction, error) {
	return _SCExecProxy.Contract.TransferAdmin(&_SCExecProxy.TransactOpts, newAdmin)
}

// SCExecProxyAdminRoleTransferredIterator is returned from FilterAdminRoleTransferred and is used to iterate over the raw logs and unpacked data for AdminRoleTransferred events raised by the SCExecProxy contract.
type SCExecProxyAdminRoleTransferredIterator struct {
	Event *SCExecProxyAdminRoleTransferred // Event containing the contract specifics and raw log

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
func (it *SCExecProxyAdminRoleTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SCExecProxyAdminRoleTransferred)
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
		it.Event = new(SCExecProxyAdminRoleTransferred)
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
func (it *SCExecProxyAdminRoleTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SCExecProxyAdminRoleTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SCExecProxyAdminRoleTransferred represents a AdminRoleTransferred event raised by the SCExecProxy contract.
type SCExecProxyAdminRoleTransferred struct {
	PreviousAdmin common.Address
	NewAdmin      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterAdminRoleTransferred is a free log retrieval operation binding the contract event 0xe379ac64de02d8184ca1a871ac486cb8137de77e485ede140e97057b9c765ffd.
//
// Solidity: event AdminRoleTransferred(address indexed previousAdmin, address indexed newAdmin)
func (_SCExecProxy *SCExecProxyFilterer) FilterAdminRoleTransferred(opts *bind.FilterOpts, previousAdmin []common.Address, newAdmin []common.Address) (*SCExecProxyAdminRoleTransferredIterator, error) {

	var previousAdminRule []interface{}
	for _, previousAdminItem := range previousAdmin {
		previousAdminRule = append(previousAdminRule, previousAdminItem)
	}
	var newAdminRule []interface{}
	for _, newAdminItem := range newAdmin {
		newAdminRule = append(newAdminRule, newAdminItem)
	}

	logs, sub, err := _SCExecProxy.contract.FilterLogs(opts, "AdminRoleTransferred", previousAdminRule, newAdminRule)
	if err != nil {
		return nil, err
	}
	return &SCExecProxyAdminRoleTransferredIterator{contract: _SCExecProxy.contract, event: "AdminRoleTransferred", logs: logs, sub: sub}, nil
}

// WatchAdminRoleTransferred is a free log subscription operation binding the contract event 0xe379ac64de02d8184ca1a871ac486cb8137de77e485ede140e97057b9c765ffd.
//
// Solidity: event AdminRoleTransferred(address indexed previousAdmin, address indexed newAdmin)
func (_SCExecProxy *SCExecProxyFilterer) WatchAdminRoleTransferred(opts *bind.WatchOpts, sink chan<- *SCExecProxyAdminRoleTransferred, previousAdmin []common.Address, newAdmin []common.Address) (event.Subscription, error) {

	var previousAdminRule []interface{}
	for _, previousAdminItem := range previousAdmin {
		previousAdminRule = append(previousAdminRule, previousAdminItem)
	}
	var newAdminRule []interface{}
	for _, newAdminItem := range newAdmin {
		newAdminRule = append(newAdminRule, newAdminItem)
	}

	logs, sub, err := _SCExecProxy.contract.WatchLogs(opts, "AdminRoleTransferred", previousAdminRule, newAdminRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SCExecProxyAdminRoleTransferred)
				if err := _SCExecProxy.contract.UnpackLog(event, "AdminRoleTransferred", log); err != nil {
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
func (_SCExecProxy *SCExecProxyFilterer) ParseAdminRoleTransferred(log types.Log) (*SCExecProxyAdminRoleTransferred, error) {
	event := new(SCExecProxyAdminRoleTransferred)
	if err := _SCExecProxy.contract.UnpackLog(event, "AdminRoleTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SCExecProxyERC20SCDepositIterator is returned from FilterERC20SCDeposit and is used to iterate over the raw logs and unpacked data for ERC20SCDeposit events raised by the SCExecProxy contract.
type SCExecProxyERC20SCDepositIterator struct {
	Event *SCExecProxyERC20SCDeposit // Event containing the contract specifics and raw log

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
func (it *SCExecProxyERC20SCDepositIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SCExecProxyERC20SCDeposit)
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
		it.Event = new(SCExecProxyERC20SCDeposit)
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
func (it *SCExecProxyERC20SCDepositIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SCExecProxyERC20SCDepositIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SCExecProxyERC20SCDeposit represents a ERC20SCDeposit event raised by the SCExecProxy contract.
type SCExecProxyERC20SCDeposit struct {
	BatchNonce   uint64
	DepositNonce uint64
	CallData     string
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterERC20SCDeposit is a free log retrieval operation binding the contract event 0x78b7f0ae2aed1debc707b7fdfc571d40c238bf57db740c833098f00dfddb6e9b.
//
// Solidity: event ERC20SCDeposit(uint64 indexed batchNonce, uint64 depositNonce, string callData)
func (_SCExecProxy *SCExecProxyFilterer) FilterERC20SCDeposit(opts *bind.FilterOpts, batchNonce []uint64) (*SCExecProxyERC20SCDepositIterator, error) {

	var batchNonceRule []interface{}
	for _, batchNonceItem := range batchNonce {
		batchNonceRule = append(batchNonceRule, batchNonceItem)
	}

	logs, sub, err := _SCExecProxy.contract.FilterLogs(opts, "ERC20SCDeposit", batchNonceRule)
	if err != nil {
		return nil, err
	}
	return &SCExecProxyERC20SCDepositIterator{contract: _SCExecProxy.contract, event: "ERC20SCDeposit", logs: logs, sub: sub}, nil
}

// WatchERC20SCDeposit is a free log subscription operation binding the contract event 0x78b7f0ae2aed1debc707b7fdfc571d40c238bf57db740c833098f00dfddb6e9b.
//
// Solidity: event ERC20SCDeposit(uint64 indexed batchNonce, uint64 depositNonce, string callData)
func (_SCExecProxy *SCExecProxyFilterer) WatchERC20SCDeposit(opts *bind.WatchOpts, sink chan<- *SCExecProxyERC20SCDeposit, batchNonce []uint64) (event.Subscription, error) {

	var batchNonceRule []interface{}
	for _, batchNonceItem := range batchNonce {
		batchNonceRule = append(batchNonceRule, batchNonceItem)
	}

	logs, sub, err := _SCExecProxy.contract.WatchLogs(opts, "ERC20SCDeposit", batchNonceRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SCExecProxyERC20SCDeposit)
				if err := _SCExecProxy.contract.UnpackLog(event, "ERC20SCDeposit", log); err != nil {
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

// ParseERC20SCDeposit is a log parse operation binding the contract event 0x78b7f0ae2aed1debc707b7fdfc571d40c238bf57db740c833098f00dfddb6e9b.
//
// Solidity: event ERC20SCDeposit(uint64 indexed batchNonce, uint64 depositNonce, string callData)
func (_SCExecProxy *SCExecProxyFilterer) ParseERC20SCDeposit(log types.Log) (*SCExecProxyERC20SCDeposit, error) {
	event := new(SCExecProxyERC20SCDeposit)
	if err := _SCExecProxy.contract.UnpackLog(event, "ERC20SCDeposit", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
