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

// EthereumIdentityFactoryContractABI is the input ABI used to generate the binding from.
const EthereumIdentityFactoryContractABI = "[{\"constant\":false,\"inputs\":[{\"name\":\"_centrifugeId\",\"type\":\"bytes32\"}],\"name\":\"createIdentity\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"name\":\"_registry\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"centrifugeId\",\"type\":\"bytes32\"},{\"indexed\":false,\"name\":\"identity\",\"type\":\"address\"}],\"name\":\"IdentityCreated\",\"type\":\"event\"}]"

// EthereumIdentityFactoryContract is an auto generated Go binding around an Ethereum contract.
type EthereumIdentityFactoryContract struct {
	EthereumIdentityFactoryContractCaller     // Read-only binding to the contract
	EthereumIdentityFactoryContractTransactor // Write-only binding to the contract
	EthereumIdentityFactoryContractFilterer   // Log filterer for contract events
}

// EthereumIdentityFactoryContractCaller is an auto generated read-only Go binding around an Ethereum contract.
type EthereumIdentityFactoryContractCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// EthereumIdentityFactoryContractTransactor is an auto generated write-only Go binding around an Ethereum contract.
type EthereumIdentityFactoryContractTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// EthereumIdentityFactoryContractFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type EthereumIdentityFactoryContractFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// EthereumIdentityFactoryContractSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type EthereumIdentityFactoryContractSession struct {
	Contract     *EthereumIdentityFactoryContract // Generic contract binding to set the session for
	CallOpts     bind.CallOpts                    // Call options to use throughout this session
	TransactOpts bind.TransactOpts                // Transaction auth options to use throughout this session
}

// EthereumIdentityFactoryContractCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type EthereumIdentityFactoryContractCallerSession struct {
	Contract *EthereumIdentityFactoryContractCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts                          // Call options to use throughout this session
}

// EthereumIdentityFactoryContractTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type EthereumIdentityFactoryContractTransactorSession struct {
	Contract     *EthereumIdentityFactoryContractTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts                          // Transaction auth options to use throughout this session
}

// EthereumIdentityFactoryContractRaw is an auto generated low-level Go binding around an Ethereum contract.
type EthereumIdentityFactoryContractRaw struct {
	Contract *EthereumIdentityFactoryContract // Generic contract binding to access the raw methods on
}

// EthereumIdentityFactoryContractCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type EthereumIdentityFactoryContractCallerRaw struct {
	Contract *EthereumIdentityFactoryContractCaller // Generic read-only contract binding to access the raw methods on
}

// EthereumIdentityFactoryContractTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type EthereumIdentityFactoryContractTransactorRaw struct {
	Contract *EthereumIdentityFactoryContractTransactor // Generic write-only contract binding to access the raw methods on
}

// NewEthereumIdentityFactoryContract creates a new instance of EthereumIdentityFactoryContract, bound to a specific deployed contract.
func NewEthereumIdentityFactoryContract(address common.Address, backend bind.ContractBackend) (*EthereumIdentityFactoryContract, error) {
	contract, err := bindEthereumIdentityFactoryContract(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &EthereumIdentityFactoryContract{EthereumIdentityFactoryContractCaller: EthereumIdentityFactoryContractCaller{contract: contract}, EthereumIdentityFactoryContractTransactor: EthereumIdentityFactoryContractTransactor{contract: contract}, EthereumIdentityFactoryContractFilterer: EthereumIdentityFactoryContractFilterer{contract: contract}}, nil
}

// NewEthereumIdentityFactoryContractCaller creates a new read-only instance of EthereumIdentityFactoryContract, bound to a specific deployed contract.
func NewEthereumIdentityFactoryContractCaller(address common.Address, caller bind.ContractCaller) (*EthereumIdentityFactoryContractCaller, error) {
	contract, err := bindEthereumIdentityFactoryContract(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &EthereumIdentityFactoryContractCaller{contract: contract}, nil
}

// NewEthereumIdentityFactoryContractTransactor creates a new write-only instance of EthereumIdentityFactoryContract, bound to a specific deployed contract.
func NewEthereumIdentityFactoryContractTransactor(address common.Address, transactor bind.ContractTransactor) (*EthereumIdentityFactoryContractTransactor, error) {
	contract, err := bindEthereumIdentityFactoryContract(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &EthereumIdentityFactoryContractTransactor{contract: contract}, nil
}

// NewEthereumIdentityFactoryContractFilterer creates a new log filterer instance of EthereumIdentityFactoryContract, bound to a specific deployed contract.
func NewEthereumIdentityFactoryContractFilterer(address common.Address, filterer bind.ContractFilterer) (*EthereumIdentityFactoryContractFilterer, error) {
	contract, err := bindEthereumIdentityFactoryContract(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &EthereumIdentityFactoryContractFilterer{contract: contract}, nil
}

// bindEthereumIdentityFactoryContract binds a generic wrapper to an already deployed contract.
func bindEthereumIdentityFactoryContract(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(EthereumIdentityFactoryContractABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_EthereumIdentityFactoryContract *EthereumIdentityFactoryContractRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _EthereumIdentityFactoryContract.Contract.EthereumIdentityFactoryContractCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_EthereumIdentityFactoryContract *EthereumIdentityFactoryContractRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _EthereumIdentityFactoryContract.Contract.EthereumIdentityFactoryContractTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_EthereumIdentityFactoryContract *EthereumIdentityFactoryContractRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _EthereumIdentityFactoryContract.Contract.EthereumIdentityFactoryContractTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_EthereumIdentityFactoryContract *EthereumIdentityFactoryContractCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _EthereumIdentityFactoryContract.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_EthereumIdentityFactoryContract *EthereumIdentityFactoryContractTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _EthereumIdentityFactoryContract.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_EthereumIdentityFactoryContract *EthereumIdentityFactoryContractTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _EthereumIdentityFactoryContract.Contract.contract.Transact(opts, method, params...)
}

// CreateIdentity is a paid mutator transaction binding the contract method 0x216b0089.
//
// Solidity: function createIdentity(_centrifugeId bytes32) returns()
func (_EthereumIdentityFactoryContract *EthereumIdentityFactoryContractTransactor) CreateIdentity(opts *bind.TransactOpts, _centrifugeId [32]byte) (*types.Transaction, error) {
	return _EthereumIdentityFactoryContract.contract.Transact(opts, "createIdentity", _centrifugeId)
}

// CreateIdentity is a paid mutator transaction binding the contract method 0x216b0089.
//
// Solidity: function createIdentity(_centrifugeId bytes32) returns()
func (_EthereumIdentityFactoryContract *EthereumIdentityFactoryContractSession) CreateIdentity(_centrifugeId [32]byte) (*types.Transaction, error) {
	return _EthereumIdentityFactoryContract.Contract.CreateIdentity(&_EthereumIdentityFactoryContract.TransactOpts, _centrifugeId)
}

// CreateIdentity is a paid mutator transaction binding the contract method 0x216b0089.
//
// Solidity: function createIdentity(_centrifugeId bytes32) returns()
func (_EthereumIdentityFactoryContract *EthereumIdentityFactoryContractTransactorSession) CreateIdentity(_centrifugeId [32]byte) (*types.Transaction, error) {
	return _EthereumIdentityFactoryContract.Contract.CreateIdentity(&_EthereumIdentityFactoryContract.TransactOpts, _centrifugeId)
}

// EthereumIdentityFactoryContractIdentityCreatedIterator is returned from FilterIdentityCreated and is used to iterate over the raw logs and unpacked data for IdentityCreated events raised by the EthereumIdentityFactoryContract contract.
type EthereumIdentityFactoryContractIdentityCreatedIterator struct {
	Event *EthereumIdentityFactoryContractIdentityCreated // Event containing the contract specifics and raw log

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
func (it *EthereumIdentityFactoryContractIdentityCreatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(EthereumIdentityFactoryContractIdentityCreated)
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
		it.Event = new(EthereumIdentityFactoryContractIdentityCreated)
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
func (it *EthereumIdentityFactoryContractIdentityCreatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *EthereumIdentityFactoryContractIdentityCreatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// EthereumIdentityFactoryContractIdentityCreated represents a IdentityCreated event raised by the EthereumIdentityFactoryContract contract.
type EthereumIdentityFactoryContractIdentityCreated struct {
	CentrifugeId [32]byte
	Identity     common.Address
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterIdentityCreated is a free log retrieval operation binding the contract event 0xd5413e953e9014ac81206e92bce8c06461ad70cfc75b747d1e8ec20cf95b68d9.
//
// Solidity: event IdentityCreated(centrifugeId indexed bytes32, identity address)
func (_EthereumIdentityFactoryContract *EthereumIdentityFactoryContractFilterer) FilterIdentityCreated(opts *bind.FilterOpts, centrifugeId [][32]byte) (*EthereumIdentityFactoryContractIdentityCreatedIterator, error) {

	var centrifugeIdRule []interface{}
	for _, centrifugeIdItem := range centrifugeId {
		centrifugeIdRule = append(centrifugeIdRule, centrifugeIdItem)
	}

	logs, sub, err := _EthereumIdentityFactoryContract.contract.FilterLogs(opts, "IdentityCreated", centrifugeIdRule)
	if err != nil {
		return nil, err
	}
	return &EthereumIdentityFactoryContractIdentityCreatedIterator{contract: _EthereumIdentityFactoryContract.contract, event: "IdentityCreated", logs: logs, sub: sub}, nil
}

// WatchIdentityCreated is a free log subscription operation binding the contract event 0xd5413e953e9014ac81206e92bce8c06461ad70cfc75b747d1e8ec20cf95b68d9.
//
// Solidity: event IdentityCreated(centrifugeId indexed bytes32, identity address)
func (_EthereumIdentityFactoryContract *EthereumIdentityFactoryContractFilterer) WatchIdentityCreated(opts *bind.WatchOpts, sink chan<- *EthereumIdentityFactoryContractIdentityCreated, centrifugeId [][32]byte) (event.Subscription, error) {

	var centrifugeIdRule []interface{}
	for _, centrifugeIdItem := range centrifugeId {
		centrifugeIdRule = append(centrifugeIdRule, centrifugeIdItem)
	}

	logs, sub, err := _EthereumIdentityFactoryContract.contract.WatchLogs(opts, "IdentityCreated", centrifugeIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(EthereumIdentityFactoryContractIdentityCreated)
				if err := _EthereumIdentityFactoryContract.contract.UnpackLog(event, "IdentityCreated", log); err != nil {
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
