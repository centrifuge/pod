// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package identity

import (
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// EthereumIdentityRegistryContractABI is the input ABI used to generate the binding from.
const EthereumIdentityRegistryContractABI = "[{\"constant\":false,\"inputs\":[{\"name\":\"_centrifugeId\",\"type\":\"bytes32\"},{\"name\":\"_identity\",\"type\":\"address\"}],\"name\":\"updateIdentityAddress\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_centrifugeId\",\"type\":\"bytes32\"},{\"name\":\"_identity\",\"type\":\"address\"}],\"name\":\"registerIdentity\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_centrifugeId\",\"type\":\"bytes32\"}],\"name\":\"getIdentityByCentrifugeId\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"centrifugeId\",\"type\":\"bytes32\"},{\"indexed\":false,\"name\":\"identity\",\"type\":\"address\"}],\"name\":\"IdentityRegistered\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"centrifugeId\",\"type\":\"bytes32\"},{\"indexed\":false,\"name\":\"identity\",\"type\":\"address\"}],\"name\":\"IdentityUpdated\",\"type\":\"event\"}]"

// EthereumIdentityRegistryContract is an auto generated Go binding around an Ethereum contract.
type EthereumIdentityRegistryContract struct {
	EthereumIdentityRegistryContractCaller     // Read-only binding to the contract
	EthereumIdentityRegistryContractTransactor // Write-only binding to the contract
	EthereumIdentityRegistryContractFilterer   // Log filterer for contract events
}

// EthereumIdentityRegistryContractCaller is an auto generated read-only Go binding around an Ethereum contract.
type EthereumIdentityRegistryContractCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// EthereumIdentityRegistryContractTransactor is an auto generated write-only Go binding around an Ethereum contract.
type EthereumIdentityRegistryContractTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// EthereumIdentityRegistryContractFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type EthereumIdentityRegistryContractFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// EthereumIdentityRegistryContractSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type EthereumIdentityRegistryContractSession struct {
	Contract     *EthereumIdentityRegistryContract // Generic contract binding to set the session for
	CallOpts     bind.CallOpts                     // Call options to use throughout this session
	TransactOpts bind.TransactOpts                 // Transaction auth options to use throughout this session
}

// EthereumIdentityRegistryContractCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type EthereumIdentityRegistryContractCallerSession struct {
	Contract *EthereumIdentityRegistryContractCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts                           // Call options to use throughout this session
}

// EthereumIdentityRegistryContractTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type EthereumIdentityRegistryContractTransactorSession struct {
	Contract     *EthereumIdentityRegistryContractTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts                           // Transaction auth options to use throughout this session
}

// EthereumIdentityRegistryContractRaw is an auto generated low-level Go binding around an Ethereum contract.
type EthereumIdentityRegistryContractRaw struct {
	Contract *EthereumIdentityRegistryContract // Generic contract binding to access the raw methods on
}

// EthereumIdentityRegistryContractCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type EthereumIdentityRegistryContractCallerRaw struct {
	Contract *EthereumIdentityRegistryContractCaller // Generic read-only contract binding to access the raw methods on
}

// EthereumIdentityRegistryContractTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type EthereumIdentityRegistryContractTransactorRaw struct {
	Contract *EthereumIdentityRegistryContractTransactor // Generic write-only contract binding to access the raw methods on
}

// NewEthereumIdentityRegistryContract creates a new instance of EthereumIdentityRegistryContract, bound to a specific deployed contract.
func NewEthereumIdentityRegistryContract(address common.Address, backend bind.ContractBackend) (*EthereumIdentityRegistryContract, error) {
	contract, err := bindEthereumIdentityRegistryContract(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &EthereumIdentityRegistryContract{EthereumIdentityRegistryContractCaller: EthereumIdentityRegistryContractCaller{contract: contract}, EthereumIdentityRegistryContractTransactor: EthereumIdentityRegistryContractTransactor{contract: contract}, EthereumIdentityRegistryContractFilterer: EthereumIdentityRegistryContractFilterer{contract: contract}}, nil
}

// NewEthereumIdentityRegistryContractCaller creates a new read-only instance of EthereumIdentityRegistryContract, bound to a specific deployed contract.
func NewEthereumIdentityRegistryContractCaller(address common.Address, caller bind.ContractCaller) (*EthereumIdentityRegistryContractCaller, error) {
	contract, err := bindEthereumIdentityRegistryContract(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &EthereumIdentityRegistryContractCaller{contract: contract}, nil
}

// NewEthereumIdentityRegistryContractTransactor creates a new write-only instance of EthereumIdentityRegistryContract, bound to a specific deployed contract.
func NewEthereumIdentityRegistryContractTransactor(address common.Address, transactor bind.ContractTransactor) (*EthereumIdentityRegistryContractTransactor, error) {
	contract, err := bindEthereumIdentityRegistryContract(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &EthereumIdentityRegistryContractTransactor{contract: contract}, nil
}

// NewEthereumIdentityRegistryContractFilterer creates a new log filterer instance of EthereumIdentityRegistryContract, bound to a specific deployed contract.
func NewEthereumIdentityRegistryContractFilterer(address common.Address, filterer bind.ContractFilterer) (*EthereumIdentityRegistryContractFilterer, error) {
	contract, err := bindEthereumIdentityRegistryContract(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &EthereumIdentityRegistryContractFilterer{contract: contract}, nil
}

// bindEthereumIdentityRegistryContract binds a generic wrapper to an already deployed contract.
func bindEthereumIdentityRegistryContract(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(EthereumIdentityRegistryContractABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_EthereumIdentityRegistryContract *EthereumIdentityRegistryContractRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _EthereumIdentityRegistryContract.Contract.EthereumIdentityRegistryContractCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_EthereumIdentityRegistryContract *EthereumIdentityRegistryContractRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _EthereumIdentityRegistryContract.Contract.EthereumIdentityRegistryContractTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_EthereumIdentityRegistryContract *EthereumIdentityRegistryContractRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _EthereumIdentityRegistryContract.Contract.EthereumIdentityRegistryContractTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_EthereumIdentityRegistryContract *EthereumIdentityRegistryContractCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _EthereumIdentityRegistryContract.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_EthereumIdentityRegistryContract *EthereumIdentityRegistryContractTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _EthereumIdentityRegistryContract.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_EthereumIdentityRegistryContract *EthereumIdentityRegistryContractTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _EthereumIdentityRegistryContract.Contract.contract.Transact(opts, method, params...)
}

// GetIdentityByCentrifugeId is a free data retrieval call binding the contract method 0xdf448e56.
//
// Solidity: function getIdentityByCentrifugeId(_centrifugeId bytes32) constant returns(address)
func (_EthereumIdentityRegistryContract *EthereumIdentityRegistryContractCaller) GetIdentityByCentrifugeId(opts *bind.CallOpts, _centrifugeId [32]byte) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _EthereumIdentityRegistryContract.contract.Call(opts, out, "getIdentityByCentrifugeId", _centrifugeId)
	return *ret0, err
}

// GetIdentityByCentrifugeId is a free data retrieval call binding the contract method 0xdf448e56.
//
// Solidity: function getIdentityByCentrifugeId(_centrifugeId bytes32) constant returns(address)
func (_EthereumIdentityRegistryContract *EthereumIdentityRegistryContractSession) GetIdentityByCentrifugeId(_centrifugeId [32]byte) (common.Address, error) {
	return _EthereumIdentityRegistryContract.Contract.GetIdentityByCentrifugeId(&_EthereumIdentityRegistryContract.CallOpts, _centrifugeId)
}

// GetIdentityByCentrifugeId is a free data retrieval call binding the contract method 0xdf448e56.
//
// Solidity: function getIdentityByCentrifugeId(_centrifugeId bytes32) constant returns(address)
func (_EthereumIdentityRegistryContract *EthereumIdentityRegistryContractCallerSession) GetIdentityByCentrifugeId(_centrifugeId [32]byte) (common.Address, error) {
	return _EthereumIdentityRegistryContract.Contract.GetIdentityByCentrifugeId(&_EthereumIdentityRegistryContract.CallOpts, _centrifugeId)
}

// RegisterIdentity is a paid mutator transaction binding the contract method 0xa134531c.
//
// Solidity: function registerIdentity(_centrifugeId bytes32, _identity address) returns()
func (_EthereumIdentityRegistryContract *EthereumIdentityRegistryContractTransactor) RegisterIdentity(opts *bind.TransactOpts, _centrifugeId [32]byte, _identity common.Address) (*types.Transaction, error) {
	return _EthereumIdentityRegistryContract.contract.Transact(opts, "registerIdentity", _centrifugeId, _identity)
}

// RegisterIdentity is a paid mutator transaction binding the contract method 0xa134531c.
//
// Solidity: function registerIdentity(_centrifugeId bytes32, _identity address) returns()
func (_EthereumIdentityRegistryContract *EthereumIdentityRegistryContractSession) RegisterIdentity(_centrifugeId [32]byte, _identity common.Address) (*types.Transaction, error) {
	return _EthereumIdentityRegistryContract.Contract.RegisterIdentity(&_EthereumIdentityRegistryContract.TransactOpts, _centrifugeId, _identity)
}

// RegisterIdentity is a paid mutator transaction binding the contract method 0xa134531c.
//
// Solidity: function registerIdentity(_centrifugeId bytes32, _identity address) returns()
func (_EthereumIdentityRegistryContract *EthereumIdentityRegistryContractTransactorSession) RegisterIdentity(_centrifugeId [32]byte, _identity common.Address) (*types.Transaction, error) {
	return _EthereumIdentityRegistryContract.Contract.RegisterIdentity(&_EthereumIdentityRegistryContract.TransactOpts, _centrifugeId, _identity)
}

// UpdateIdentityAddress is a paid mutator transaction binding the contract method 0x5c1b60bb.
//
// Solidity: function updateIdentityAddress(_centrifugeId bytes32, _identity address) returns()
func (_EthereumIdentityRegistryContract *EthereumIdentityRegistryContractTransactor) UpdateIdentityAddress(opts *bind.TransactOpts, _centrifugeId [32]byte, _identity common.Address) (*types.Transaction, error) {
	return _EthereumIdentityRegistryContract.contract.Transact(opts, "updateIdentityAddress", _centrifugeId, _identity)
}

// UpdateIdentityAddress is a paid mutator transaction binding the contract method 0x5c1b60bb.
//
// Solidity: function updateIdentityAddress(_centrifugeId bytes32, _identity address) returns()
func (_EthereumIdentityRegistryContract *EthereumIdentityRegistryContractSession) UpdateIdentityAddress(_centrifugeId [32]byte, _identity common.Address) (*types.Transaction, error) {
	return _EthereumIdentityRegistryContract.Contract.UpdateIdentityAddress(&_EthereumIdentityRegistryContract.TransactOpts, _centrifugeId, _identity)
}

// UpdateIdentityAddress is a paid mutator transaction binding the contract method 0x5c1b60bb.
//
// Solidity: function updateIdentityAddress(_centrifugeId bytes32, _identity address) returns()
func (_EthereumIdentityRegistryContract *EthereumIdentityRegistryContractTransactorSession) UpdateIdentityAddress(_centrifugeId [32]byte, _identity common.Address) (*types.Transaction, error) {
	return _EthereumIdentityRegistryContract.Contract.UpdateIdentityAddress(&_EthereumIdentityRegistryContract.TransactOpts, _centrifugeId, _identity)
}

// EthereumIdentityRegistryContractIdentityRegisteredIterator is returned from FilterIdentityRegistered and is used to iterate over the raw logs and unpacked data for IdentityRegistered events raised by the EthereumIdentityRegistryContract contract.
type EthereumIdentityRegistryContractIdentityRegisteredIterator struct {
	Event *EthereumIdentityRegistryContractIdentityRegistered // Event containing the contract specifics and raw log

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
func (it *EthereumIdentityRegistryContractIdentityRegisteredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(EthereumIdentityRegistryContractIdentityRegistered)
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
		it.Event = new(EthereumIdentityRegistryContractIdentityRegistered)
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
func (it *EthereumIdentityRegistryContractIdentityRegisteredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *EthereumIdentityRegistryContractIdentityRegisteredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// EthereumIdentityRegistryContractIdentityRegistered represents a IdentityRegistered event raised by the EthereumIdentityRegistryContract contract.
type EthereumIdentityRegistryContractIdentityRegistered struct {
	CentrifugeId [32]byte
	Identity     common.Address
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterIdentityRegistered is a free log retrieval operation binding the contract event 0xd0c90c8049cdcd414cc4e521cc98d9c70f84ba695cde160691c0aebc48df058a.
//
// Solidity: event IdentityRegistered(centrifugeId indexed bytes32, identity address)
func (_EthereumIdentityRegistryContract *EthereumIdentityRegistryContractFilterer) FilterIdentityRegistered(opts *bind.FilterOpts, centrifugeId [][32]byte) (*EthereumIdentityRegistryContractIdentityRegisteredIterator, error) {

	var centrifugeIdRule []interface{}
	for _, centrifugeIdItem := range centrifugeId {
		centrifugeIdRule = append(centrifugeIdRule, centrifugeIdItem)
	}

	logs, sub, err := _EthereumIdentityRegistryContract.contract.FilterLogs(opts, "IdentityRegistered", centrifugeIdRule)
	if err != nil {
		return nil, err
	}
	return &EthereumIdentityRegistryContractIdentityRegisteredIterator{contract: _EthereumIdentityRegistryContract.contract, event: "IdentityRegistered", logs: logs, sub: sub}, nil
}

// WatchIdentityRegistered is a free log subscription operation binding the contract event 0xd0c90c8049cdcd414cc4e521cc98d9c70f84ba695cde160691c0aebc48df058a.
//
// Solidity: event IdentityRegistered(centrifugeId indexed bytes32, identity address)
func (_EthereumIdentityRegistryContract *EthereumIdentityRegistryContractFilterer) WatchIdentityRegistered(opts *bind.WatchOpts, sink chan<- *EthereumIdentityRegistryContractIdentityRegistered, centrifugeId [][32]byte) (event.Subscription, error) {

	var centrifugeIdRule []interface{}
	for _, centrifugeIdItem := range centrifugeId {
		centrifugeIdRule = append(centrifugeIdRule, centrifugeIdItem)
	}

	logs, sub, err := _EthereumIdentityRegistryContract.contract.WatchLogs(opts, "IdentityRegistered", centrifugeIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(EthereumIdentityRegistryContractIdentityRegistered)
				if err := _EthereumIdentityRegistryContract.contract.UnpackLog(event, "IdentityRegistered", log); err != nil {
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

// EthereumIdentityRegistryContractIdentityUpdatedIterator is returned from FilterIdentityUpdated and is used to iterate over the raw logs and unpacked data for IdentityUpdated events raised by the EthereumIdentityRegistryContract contract.
type EthereumIdentityRegistryContractIdentityUpdatedIterator struct {
	Event *EthereumIdentityRegistryContractIdentityUpdated // Event containing the contract specifics and raw log

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
func (it *EthereumIdentityRegistryContractIdentityUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(EthereumIdentityRegistryContractIdentityUpdated)
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
		it.Event = new(EthereumIdentityRegistryContractIdentityUpdated)
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
func (it *EthereumIdentityRegistryContractIdentityUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *EthereumIdentityRegistryContractIdentityUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// EthereumIdentityRegistryContractIdentityUpdated represents a IdentityUpdated event raised by the EthereumIdentityRegistryContract contract.
type EthereumIdentityRegistryContractIdentityUpdated struct {
	CentrifugeId [32]byte
	Identity     common.Address
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterIdentityUpdated is a free log retrieval operation binding the contract event 0x043d2bc4475f97a36335b76dbd57eb79e7a7f8f052230b56e5e0473605b975b5.
//
// Solidity: event IdentityUpdated(centrifugeId indexed bytes32, identity address)
func (_EthereumIdentityRegistryContract *EthereumIdentityRegistryContractFilterer) FilterIdentityUpdated(opts *bind.FilterOpts, centrifugeId [][32]byte) (*EthereumIdentityRegistryContractIdentityUpdatedIterator, error) {

	var centrifugeIdRule []interface{}
	for _, centrifugeIdItem := range centrifugeId {
		centrifugeIdRule = append(centrifugeIdRule, centrifugeIdItem)
	}

	logs, sub, err := _EthereumIdentityRegistryContract.contract.FilterLogs(opts, "IdentityUpdated", centrifugeIdRule)
	if err != nil {
		return nil, err
	}
	return &EthereumIdentityRegistryContractIdentityUpdatedIterator{contract: _EthereumIdentityRegistryContract.contract, event: "IdentityUpdated", logs: logs, sub: sub}, nil
}

// WatchIdentityUpdated is a free log subscription operation binding the contract event 0x043d2bc4475f97a36335b76dbd57eb79e7a7f8f052230b56e5e0473605b975b5.
//
// Solidity: event IdentityUpdated(centrifugeId indexed bytes32, identity address)
func (_EthereumIdentityRegistryContract *EthereumIdentityRegistryContractFilterer) WatchIdentityUpdated(opts *bind.WatchOpts, sink chan<- *EthereumIdentityRegistryContractIdentityUpdated, centrifugeId [][32]byte) (event.Subscription, error) {

	var centrifugeIdRule []interface{}
	for _, centrifugeIdItem := range centrifugeId {
		centrifugeIdRule = append(centrifugeIdRule, centrifugeIdItem)
	}

	logs, sub, err := _EthereumIdentityRegistryContract.contract.WatchLogs(opts, "IdentityUpdated", centrifugeIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(EthereumIdentityRegistryContractIdentityUpdated)
				if err := _EthereumIdentityRegistryContract.contract.UnpackLog(event, "IdentityUpdated", log); err != nil {
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
