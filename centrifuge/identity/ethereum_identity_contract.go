// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package identity

import (
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// EthereumIdentityContractABI is the input ABI used to generate the binding from.
const EthereumIdentityContractABI = "[{\"constant\":false,\"inputs\":[{\"name\":\"_key\",\"type\":\"bytes32\"},{\"name\":\"_kType\",\"type\":\"uint256\"}],\"name\":\"addKey\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"centrifugeId\",\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_kType\",\"type\":\"uint256\"}],\"name\":\"getKeysByType\",\"outputs\":[{\"name\":\"\",\"type\":\"bytes32[]\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"name\":\"_centrifugeId\",\"type\":\"bytes32\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"kType\",\"type\":\"uint256\"},{\"indexed\":true,\"name\":\"key\",\"type\":\"bytes32\"}],\"name\":\"KeyRegistered\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"}]"

// EthereumIdentityContract is an auto generated Go binding around an Ethereum contract.
type EthereumIdentityContract struct {
	EthereumIdentityContractCaller     // Read-only binding to the contract
	EthereumIdentityContractTransactor // Write-only binding to the contract
	EthereumIdentityContractFilterer   // Log filterer for contract events
}

// EthereumIdentityContractCaller is an auto generated read-only Go binding around an Ethereum contract.
type EthereumIdentityContractCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// EthereumIdentityContractTransactor is an auto generated write-only Go binding around an Ethereum contract.
type EthereumIdentityContractTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// EthereumIdentityContractFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type EthereumIdentityContractFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// EthereumIdentityContractSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type EthereumIdentityContractSession struct {
	Contract     *EthereumIdentityContract // Generic contract binding to set the session for
	CallOpts     bind.CallOpts             // Call options to use throughout this session
	TransactOpts bind.TransactOpts         // Transaction auth options to use throughout this session
}

// EthereumIdentityContractCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type EthereumIdentityContractCallerSession struct {
	Contract *EthereumIdentityContractCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts                   // Call options to use throughout this session
}

// EthereumIdentityContractTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type EthereumIdentityContractTransactorSession struct {
	Contract     *EthereumIdentityContractTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts                   // Transaction auth options to use throughout this session
}

// EthereumIdentityContractRaw is an auto generated low-level Go binding around an Ethereum contract.
type EthereumIdentityContractRaw struct {
	Contract *EthereumIdentityContract // Generic contract binding to access the raw methods on
}

// EthereumIdentityContractCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type EthereumIdentityContractCallerRaw struct {
	Contract *EthereumIdentityContractCaller // Generic read-only contract binding to access the raw methods on
}

// EthereumIdentityContractTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type EthereumIdentityContractTransactorRaw struct {
	Contract *EthereumIdentityContractTransactor // Generic write-only contract binding to access the raw methods on
}

// NewEthereumIdentityContract creates a new instance of EthereumIdentityContract, bound to a specific deployed contract.
func NewEthereumIdentityContract(address common.Address, backend bind.ContractBackend) (*EthereumIdentityContract, error) {
	contract, err := bindEthereumIdentityContract(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &EthereumIdentityContract{EthereumIdentityContractCaller: EthereumIdentityContractCaller{contract: contract}, EthereumIdentityContractTransactor: EthereumIdentityContractTransactor{contract: contract}, EthereumIdentityContractFilterer: EthereumIdentityContractFilterer{contract: contract}}, nil
}

// NewEthereumIdentityContractCaller creates a new read-only instance of EthereumIdentityContract, bound to a specific deployed contract.
func NewEthereumIdentityContractCaller(address common.Address, caller bind.ContractCaller) (*EthereumIdentityContractCaller, error) {
	contract, err := bindEthereumIdentityContract(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &EthereumIdentityContractCaller{contract: contract}, nil
}

// NewEthereumIdentityContractTransactor creates a new write-only instance of EthereumIdentityContract, bound to a specific deployed contract.
func NewEthereumIdentityContractTransactor(address common.Address, transactor bind.ContractTransactor) (*EthereumIdentityContractTransactor, error) {
	contract, err := bindEthereumIdentityContract(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &EthereumIdentityContractTransactor{contract: contract}, nil
}

// NewEthereumIdentityContractFilterer creates a new log filterer instance of EthereumIdentityContract, bound to a specific deployed contract.
func NewEthereumIdentityContractFilterer(address common.Address, filterer bind.ContractFilterer) (*EthereumIdentityContractFilterer, error) {
	contract, err := bindEthereumIdentityContract(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &EthereumIdentityContractFilterer{contract: contract}, nil
}

// bindEthereumIdentityContract binds a generic wrapper to an already deployed contract.
func bindEthereumIdentityContract(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(EthereumIdentityContractABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_EthereumIdentityContract *EthereumIdentityContractRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _EthereumIdentityContract.Contract.EthereumIdentityContractCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_EthereumIdentityContract *EthereumIdentityContractRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _EthereumIdentityContract.Contract.EthereumIdentityContractTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_EthereumIdentityContract *EthereumIdentityContractRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _EthereumIdentityContract.Contract.EthereumIdentityContractTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_EthereumIdentityContract *EthereumIdentityContractCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _EthereumIdentityContract.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_EthereumIdentityContract *EthereumIdentityContractTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _EthereumIdentityContract.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_EthereumIdentityContract *EthereumIdentityContractTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _EthereumIdentityContract.Contract.contract.Transact(opts, method, params...)
}

// CentrifugeId is a free data retrieval call binding the contract method 0x41a43c38.
//
// Solidity: function centrifugeId() constant returns(bytes32)
func (_EthereumIdentityContract *EthereumIdentityContractCaller) CentrifugeId(opts *bind.CallOpts) ([32]byte, error) {
	var (
		ret0 = new([32]byte)
	)
	out := ret0
	err := _EthereumIdentityContract.contract.Call(opts, out, "centrifugeId")
	return *ret0, err
}

// CentrifugeId is a free data retrieval call binding the contract method 0x41a43c38.
//
// Solidity: function centrifugeId() constant returns(bytes32)
func (_EthereumIdentityContract *EthereumIdentityContractSession) CentrifugeId() ([32]byte, error) {
	return _EthereumIdentityContract.Contract.CentrifugeId(&_EthereumIdentityContract.CallOpts)
}

// CentrifugeId is a free data retrieval call binding the contract method 0x41a43c38.
//
// Solidity: function centrifugeId() constant returns(bytes32)
func (_EthereumIdentityContract *EthereumIdentityContractCallerSession) CentrifugeId() ([32]byte, error) {
	return _EthereumIdentityContract.Contract.CentrifugeId(&_EthereumIdentityContract.CallOpts)
}

// GetKeysByType is a free data retrieval call binding the contract method 0x41cbfc7b.
//
// Solidity: function getKeysByType(_kType uint256) constant returns(bytes32[])
func (_EthereumIdentityContract *EthereumIdentityContractCaller) GetKeysByType(opts *bind.CallOpts, _kType *big.Int) ([][32]byte, error) {
	var (
		ret0 = new([][32]byte)
	)
	out := ret0
	err := _EthereumIdentityContract.contract.Call(opts, out, "getKeysByType", _kType)
	return *ret0, err
}

// GetKeysByType is a free data retrieval call binding the contract method 0x41cbfc7b.
//
// Solidity: function getKeysByType(_kType uint256) constant returns(bytes32[])
func (_EthereumIdentityContract *EthereumIdentityContractSession) GetKeysByType(_kType *big.Int) ([][32]byte, error) {
	return _EthereumIdentityContract.Contract.GetKeysByType(&_EthereumIdentityContract.CallOpts, _kType)
}

// GetKeysByType is a free data retrieval call binding the contract method 0x41cbfc7b.
//
// Solidity: function getKeysByType(_kType uint256) constant returns(bytes32[])
func (_EthereumIdentityContract *EthereumIdentityContractCallerSession) GetKeysByType(_kType *big.Int) ([][32]byte, error) {
	return _EthereumIdentityContract.Contract.GetKeysByType(&_EthereumIdentityContract.CallOpts, _kType)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() constant returns(address)
func (_EthereumIdentityContract *EthereumIdentityContractCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _EthereumIdentityContract.contract.Call(opts, out, "owner")
	return *ret0, err
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() constant returns(address)
func (_EthereumIdentityContract *EthereumIdentityContractSession) Owner() (common.Address, error) {
	return _EthereumIdentityContract.Contract.Owner(&_EthereumIdentityContract.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() constant returns(address)
func (_EthereumIdentityContract *EthereumIdentityContractCallerSession) Owner() (common.Address, error) {
	return _EthereumIdentityContract.Contract.Owner(&_EthereumIdentityContract.CallOpts)
}

// AddKey is a paid mutator transaction binding the contract method 0x4103ef4c.
//
// Solidity: function addKey(_key bytes32, _kType uint256) returns()
func (_EthereumIdentityContract *EthereumIdentityContractTransactor) AddKey(opts *bind.TransactOpts, _key [32]byte, _kType *big.Int) (*types.Transaction, error) {
	return _EthereumIdentityContract.contract.Transact(opts, "addKey", _key, _kType)
}

// AddKey is a paid mutator transaction binding the contract method 0x4103ef4c.
//
// Solidity: function addKey(_key bytes32, _kType uint256) returns()
func (_EthereumIdentityContract *EthereumIdentityContractSession) AddKey(_key [32]byte, _kType *big.Int) (*types.Transaction, error) {
	return _EthereumIdentityContract.Contract.AddKey(&_EthereumIdentityContract.TransactOpts, _key, _kType)
}

// AddKey is a paid mutator transaction binding the contract method 0x4103ef4c.
//
// Solidity: function addKey(_key bytes32, _kType uint256) returns()
func (_EthereumIdentityContract *EthereumIdentityContractTransactorSession) AddKey(_key [32]byte, _kType *big.Int) (*types.Transaction, error) {
	return _EthereumIdentityContract.Contract.AddKey(&_EthereumIdentityContract.TransactOpts, _key, _kType)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(newOwner address) returns()
func (_EthereumIdentityContract *EthereumIdentityContractTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _EthereumIdentityContract.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(newOwner address) returns()
func (_EthereumIdentityContract *EthereumIdentityContractSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _EthereumIdentityContract.Contract.TransferOwnership(&_EthereumIdentityContract.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(newOwner address) returns()
func (_EthereumIdentityContract *EthereumIdentityContractTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _EthereumIdentityContract.Contract.TransferOwnership(&_EthereumIdentityContract.TransactOpts, newOwner)
}

// EthereumIdentityContractKeyRegisteredIterator is returned from FilterKeyRegistered and is used to iterate over the raw logs and unpacked data for KeyRegistered events raised by the EthereumIdentityContract contract.
type EthereumIdentityContractKeyRegisteredIterator struct {
	Event *EthereumIdentityContractKeyRegistered // Event containing the contract specifics and raw log

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
func (it *EthereumIdentityContractKeyRegisteredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(EthereumIdentityContractKeyRegistered)
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
		it.Event = new(EthereumIdentityContractKeyRegistered)
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
func (it *EthereumIdentityContractKeyRegisteredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *EthereumIdentityContractKeyRegisteredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// EthereumIdentityContractKeyRegistered represents a KeyRegistered event raised by the EthereumIdentityContract contract.
type EthereumIdentityContractKeyRegistered struct {
	KType *big.Int
	Key   [32]byte
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterKeyRegistered is a free log retrieval operation binding the contract event 0xf45ca858c8bbe25c93f4ceefe73db33816426cb24fe5b73ff350e236b76005db.
//
// Solidity: event KeyRegistered(kType indexed uint256, key indexed bytes32)
func (_EthereumIdentityContract *EthereumIdentityContractFilterer) FilterKeyRegistered(opts *bind.FilterOpts, kType []*big.Int, key [][32]byte) (*EthereumIdentityContractKeyRegisteredIterator, error) {

	var kTypeRule []interface{}
	for _, kTypeItem := range kType {
		kTypeRule = append(kTypeRule, kTypeItem)
	}
	var keyRule []interface{}
	for _, keyItem := range key {
		keyRule = append(keyRule, keyItem)
	}

	logs, sub, err := _EthereumIdentityContract.contract.FilterLogs(opts, "KeyRegistered", kTypeRule, keyRule)
	if err != nil {
		return nil, err
	}
	return &EthereumIdentityContractKeyRegisteredIterator{contract: _EthereumIdentityContract.contract, event: "KeyRegistered", logs: logs, sub: sub}, nil
}

// WatchKeyRegistered is a free log subscription operation binding the contract event 0xf45ca858c8bbe25c93f4ceefe73db33816426cb24fe5b73ff350e236b76005db.
//
// Solidity: event KeyRegistered(kType indexed uint256, key indexed bytes32)
func (_EthereumIdentityContract *EthereumIdentityContractFilterer) WatchKeyRegistered(opts *bind.WatchOpts, sink chan<- *EthereumIdentityContractKeyRegistered, kType []*big.Int, key [][32]byte) (event.Subscription, error) {

	var kTypeRule []interface{}
	for _, kTypeItem := range kType {
		kTypeRule = append(kTypeRule, kTypeItem)
	}
	var keyRule []interface{}
	for _, keyItem := range key {
		keyRule = append(keyRule, keyItem)
	}

	logs, sub, err := _EthereumIdentityContract.contract.WatchLogs(opts, "KeyRegistered", kTypeRule, keyRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(EthereumIdentityContractKeyRegistered)
				if err := _EthereumIdentityContract.contract.UnpackLog(event, "KeyRegistered", log); err != nil {
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

// EthereumIdentityContractOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the EthereumIdentityContract contract.
type EthereumIdentityContractOwnershipTransferredIterator struct {
	Event *EthereumIdentityContractOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *EthereumIdentityContractOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(EthereumIdentityContractOwnershipTransferred)
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
		it.Event = new(EthereumIdentityContractOwnershipTransferred)
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
func (it *EthereumIdentityContractOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *EthereumIdentityContractOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// EthereumIdentityContractOwnershipTransferred represents a OwnershipTransferred event raised by the EthereumIdentityContract contract.
type EthereumIdentityContractOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(previousOwner indexed address, newOwner indexed address)
func (_EthereumIdentityContract *EthereumIdentityContractFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*EthereumIdentityContractOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _EthereumIdentityContract.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &EthereumIdentityContractOwnershipTransferredIterator{contract: _EthereumIdentityContract.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(previousOwner indexed address, newOwner indexed address)
func (_EthereumIdentityContract *EthereumIdentityContractFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *EthereumIdentityContractOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _EthereumIdentityContract.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(EthereumIdentityContractOwnershipTransferred)
				if err := _EthereumIdentityContract.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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
