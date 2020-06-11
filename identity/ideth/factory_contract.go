// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package ideth

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

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
)

// FactoryContractABI is the input ABI used to generate the binding from.
const FactoryContractABI = "[{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"identity\",\"type\":\"address\"}],\"name\":\"IdentityCreated\",\"type\":\"event\"},{\"constant\":false,\"inputs\":[],\"name\":\"createIdentity\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"manager\",\"type\":\"address\"},{\"name\":\"keys\",\"type\":\"bytes32[]\"},{\"name\":\"purposes\",\"type\":\"uint256[]\"}],\"name\":\"createIdentityFor\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"identityAddr\",\"type\":\"address\"}],\"name\":\"createdIdentity\",\"outputs\":[{\"name\":\"valid\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"}]"

// FactoryContract is an auto generated Go binding around an Ethereum contract.
type FactoryContract struct {
	FactoryContractCaller     // Read-only binding to the contract
	FactoryContractTransactor // Write-only binding to the contract
	FactoryContractFilterer   // Log filterer for contract events
}

// FactoryContractCaller is an auto generated read-only Go binding around an Ethereum contract.
type FactoryContractCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// FactoryContractTransactor is an auto generated write-only Go binding around an Ethereum contract.
type FactoryContractTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// FactoryContractFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type FactoryContractFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// FactoryContractSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type FactoryContractSession struct {
	Contract     *FactoryContract  // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// FactoryContractCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type FactoryContractCallerSession struct {
	Contract *FactoryContractCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts          // Call options to use throughout this session
}

// FactoryContractTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type FactoryContractTransactorSession struct {
	Contract     *FactoryContractTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts          // Transaction auth options to use throughout this session
}

// FactoryContractRaw is an auto generated low-level Go binding around an Ethereum contract.
type FactoryContractRaw struct {
	Contract *FactoryContract // Generic contract binding to access the raw methods on
}

// FactoryContractCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type FactoryContractCallerRaw struct {
	Contract *FactoryContractCaller // Generic read-only contract binding to access the raw methods on
}

// FactoryContractTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type FactoryContractTransactorRaw struct {
	Contract *FactoryContractTransactor // Generic write-only contract binding to access the raw methods on
}

// NewFactoryContract creates a new instance of FactoryContract, bound to a specific deployed contract.
func NewFactoryContract(address common.Address, backend bind.ContractBackend) (*FactoryContract, error) {
	contract, err := bindFactoryContract(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &FactoryContract{FactoryContractCaller: FactoryContractCaller{contract: contract}, FactoryContractTransactor: FactoryContractTransactor{contract: contract}, FactoryContractFilterer: FactoryContractFilterer{contract: contract}}, nil
}

// NewFactoryContractCaller creates a new read-only instance of FactoryContract, bound to a specific deployed contract.
func NewFactoryContractCaller(address common.Address, caller bind.ContractCaller) (*FactoryContractCaller, error) {
	contract, err := bindFactoryContract(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &FactoryContractCaller{contract: contract}, nil
}

// NewFactoryContractTransactor creates a new write-only instance of FactoryContract, bound to a specific deployed contract.
func NewFactoryContractTransactor(address common.Address, transactor bind.ContractTransactor) (*FactoryContractTransactor, error) {
	contract, err := bindFactoryContract(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &FactoryContractTransactor{contract: contract}, nil
}

// NewFactoryContractFilterer creates a new log filterer instance of FactoryContract, bound to a specific deployed contract.
func NewFactoryContractFilterer(address common.Address, filterer bind.ContractFilterer) (*FactoryContractFilterer, error) {
	contract, err := bindFactoryContract(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &FactoryContractFilterer{contract: contract}, nil
}

// bindFactoryContract binds a generic wrapper to an already deployed contract.
func bindFactoryContract(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(FactoryContractABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_FactoryContract *FactoryContractRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _FactoryContract.Contract.FactoryContractCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_FactoryContract *FactoryContractRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _FactoryContract.Contract.FactoryContractTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_FactoryContract *FactoryContractRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _FactoryContract.Contract.FactoryContractTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_FactoryContract *FactoryContractCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _FactoryContract.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_FactoryContract *FactoryContractTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _FactoryContract.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_FactoryContract *FactoryContractTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _FactoryContract.Contract.contract.Transact(opts, method, params...)
}

// CreatedIdentity is a free data retrieval call binding the contract method 0xfc252feb.
//
// Solidity: function createdIdentity(address identityAddr) constant returns(bool valid)
func (_FactoryContract *FactoryContractCaller) CreatedIdentity(opts *bind.CallOpts, identityAddr common.Address) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _FactoryContract.contract.Call(opts, out, "createdIdentity", identityAddr)
	return *ret0, err
}

// CreatedIdentity is a free data retrieval call binding the contract method 0xfc252feb.
//
// Solidity: function createdIdentity(address identityAddr) constant returns(bool valid)
func (_FactoryContract *FactoryContractSession) CreatedIdentity(identityAddr common.Address) (bool, error) {
	return _FactoryContract.Contract.CreatedIdentity(&_FactoryContract.CallOpts, identityAddr)
}

// CreatedIdentity is a free data retrieval call binding the contract method 0xfc252feb.
//
// Solidity: function createdIdentity(address identityAddr) constant returns(bool valid)
func (_FactoryContract *FactoryContractCallerSession) CreatedIdentity(identityAddr common.Address) (bool, error) {
	return _FactoryContract.Contract.CreatedIdentity(&_FactoryContract.CallOpts, identityAddr)
}

// CreateIdentity is a paid mutator transaction binding the contract method 0x59d21ad9.
//
// Solidity: function createIdentity() returns()
func (_FactoryContract *FactoryContractTransactor) CreateIdentity(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _FactoryContract.contract.Transact(opts, "createIdentity")
}

// CreateIdentity is a paid mutator transaction binding the contract method 0x59d21ad9.
//
// Solidity: function createIdentity() returns()
func (_FactoryContract *FactoryContractSession) CreateIdentity() (*types.Transaction, error) {
	return _FactoryContract.Contract.CreateIdentity(&_FactoryContract.TransactOpts)
}

// CreateIdentity is a paid mutator transaction binding the contract method 0x59d21ad9.
//
// Solidity: function createIdentity() returns()
func (_FactoryContract *FactoryContractTransactorSession) CreateIdentity() (*types.Transaction, error) {
	return _FactoryContract.Contract.CreateIdentity(&_FactoryContract.TransactOpts)
}

// CreateIdentityFor is a paid mutator transaction binding the contract method 0xc4ff1c23.
//
// Solidity: function createIdentityFor(address manager, bytes32[] keys, uint256[] purposes) returns()
func (_FactoryContract *FactoryContractTransactor) CreateIdentityFor(opts *bind.TransactOpts, manager common.Address, keys [][32]byte, purposes []*big.Int) (*types.Transaction, error) {
	return _FactoryContract.contract.Transact(opts, "createIdentityFor", manager, keys, purposes)
}

// CreateIdentityFor is a paid mutator transaction binding the contract method 0xc4ff1c23.
//
// Solidity: function createIdentityFor(address manager, bytes32[] keys, uint256[] purposes) returns()
func (_FactoryContract *FactoryContractSession) CreateIdentityFor(manager common.Address, keys [][32]byte, purposes []*big.Int) (*types.Transaction, error) {
	return _FactoryContract.Contract.CreateIdentityFor(&_FactoryContract.TransactOpts, manager, keys, purposes)
}

// CreateIdentityFor is a paid mutator transaction binding the contract method 0xc4ff1c23.
//
// Solidity: function createIdentityFor(address manager, bytes32[] keys, uint256[] purposes) returns()
func (_FactoryContract *FactoryContractTransactorSession) CreateIdentityFor(manager common.Address, keys [][32]byte, purposes []*big.Int) (*types.Transaction, error) {
	return _FactoryContract.Contract.CreateIdentityFor(&_FactoryContract.TransactOpts, manager, keys, purposes)
}

// FactoryContractIdentityCreatedIterator is returned from FilterIdentityCreated and is used to iterate over the raw logs and unpacked data for IdentityCreated events raised by the FactoryContract contract.
type FactoryContractIdentityCreatedIterator struct {
	Event *FactoryContractIdentityCreated // Event containing the contract specifics and raw log

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
func (it *FactoryContractIdentityCreatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(FactoryContractIdentityCreated)
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
		it.Event = new(FactoryContractIdentityCreated)
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
func (it *FactoryContractIdentityCreatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *FactoryContractIdentityCreatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// FactoryContractIdentityCreated represents a IdentityCreated event raised by the FactoryContract contract.
type FactoryContractIdentityCreated struct {
	Identity common.Address
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterIdentityCreated is a free log retrieval operation binding the contract event 0xac993fde3b9423ff59e4a23cded8e89074c9c8740920d1d870f586ba7c5c8cf0.
//
// Solidity: event IdentityCreated(address indexed identity)
func (_FactoryContract *FactoryContractFilterer) FilterIdentityCreated(opts *bind.FilterOpts, identity []common.Address) (*FactoryContractIdentityCreatedIterator, error) {

	var identityRule []interface{}
	for _, identityItem := range identity {
		identityRule = append(identityRule, identityItem)
	}

	logs, sub, err := _FactoryContract.contract.FilterLogs(opts, "IdentityCreated", identityRule)
	if err != nil {
		return nil, err
	}
	return &FactoryContractIdentityCreatedIterator{contract: _FactoryContract.contract, event: "IdentityCreated", logs: logs, sub: sub}, nil
}

// WatchIdentityCreated is a free log subscription operation binding the contract event 0xac993fde3b9423ff59e4a23cded8e89074c9c8740920d1d870f586ba7c5c8cf0.
//
// Solidity: event IdentityCreated(address indexed identity)
func (_FactoryContract *FactoryContractFilterer) WatchIdentityCreated(opts *bind.WatchOpts, sink chan<- *FactoryContractIdentityCreated, identity []common.Address) (event.Subscription, error) {

	var identityRule []interface{}
	for _, identityItem := range identity {
		identityRule = append(identityRule, identityItem)
	}

	logs, sub, err := _FactoryContract.contract.WatchLogs(opts, "IdentityCreated", identityRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(FactoryContractIdentityCreated)
				if err := _FactoryContract.contract.UnpackLog(event, "IdentityCreated", log); err != nil {
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
