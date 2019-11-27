// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package anchors

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
	_ = abi.U256
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
)

// AnchorContractABI is the input ABI used to generate the binding from.
const AnchorContractABI = "[{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"anchorId\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"documentRoot\",\"type\":\"bytes32\"},{\"indexed\":false,\"name\":\"blockHeight\",\"type\":\"uint32\"}],\"name\":\"AnchorCommitted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"anchorId\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"blockHeight\",\"type\":\"uint32\"}],\"name\":\"AnchorPreCommitted\",\"type\":\"event\"},{\"constant\":false,\"inputs\":[{\"name\":\"anchorId\",\"type\":\"uint256\"},{\"name\":\"signingRoot\",\"type\":\"bytes32\"}],\"name\":\"preCommit\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"anchorIdPreImage\",\"type\":\"uint256\"},{\"name\":\"documentRoot\",\"type\":\"bytes32\"},{\"name\":\"proof\",\"type\":\"bytes32\"}],\"name\":\"commit\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"id\",\"type\":\"uint256\"}],\"name\":\"getAnchorById\",\"outputs\":[{\"name\":\"anchorId\",\"type\":\"uint256\"},{\"name\":\"documentRoot\",\"type\":\"bytes32\"},{\"name\":\"blockNumber\",\"type\":\"uint32\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"anchorId\",\"type\":\"uint256\"}],\"name\":\"hasValidPreCommit\",\"outputs\":[{\"name\":\"valid\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"}]"

// AnchorContract is an auto generated Go binding around an Ethereum contract.
type AnchorContract struct {
	AnchorContractCaller     // Read-only binding to the contract
	AnchorContractTransactor // Write-only binding to the contract
	AnchorContractFilterer   // Log filterer for contract events
}

// AnchorContractCaller is an auto generated read-only Go binding around an Ethereum contract.
type AnchorContractCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AnchorContractTransactor is an auto generated write-only Go binding around an Ethereum contract.
type AnchorContractTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AnchorContractFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type AnchorContractFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AnchorContractSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type AnchorContractSession struct {
	Contract     *AnchorContract   // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// AnchorContractCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type AnchorContractCallerSession struct {
	Contract *AnchorContractCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts         // Call options to use throughout this session
}

// AnchorContractTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type AnchorContractTransactorSession struct {
	Contract     *AnchorContractTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts         // Transaction auth options to use throughout this session
}

// AnchorContractRaw is an auto generated low-level Go binding around an Ethereum contract.
type AnchorContractRaw struct {
	Contract *AnchorContract // Generic contract binding to access the raw methods on
}

// AnchorContractCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type AnchorContractCallerRaw struct {
	Contract *AnchorContractCaller // Generic read-only contract binding to access the raw methods on
}

// AnchorContractTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type AnchorContractTransactorRaw struct {
	Contract *AnchorContractTransactor // Generic write-only contract binding to access the raw methods on
}

// NewAnchorContract creates a new instance of AnchorContract, bound to a specific deployed contract.
func NewAnchorContract(address common.Address, backend bind.ContractBackend) (*AnchorContract, error) {
	contract, err := bindAnchorContract(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &AnchorContract{AnchorContractCaller: AnchorContractCaller{contract: contract}, AnchorContractTransactor: AnchorContractTransactor{contract: contract}, AnchorContractFilterer: AnchorContractFilterer{contract: contract}}, nil
}

// NewAnchorContractCaller creates a new read-only instance of AnchorContract, bound to a specific deployed contract.
func NewAnchorContractCaller(address common.Address, caller bind.ContractCaller) (*AnchorContractCaller, error) {
	contract, err := bindAnchorContract(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &AnchorContractCaller{contract: contract}, nil
}

// NewAnchorContractTransactor creates a new write-only instance of AnchorContract, bound to a specific deployed contract.
func NewAnchorContractTransactor(address common.Address, transactor bind.ContractTransactor) (*AnchorContractTransactor, error) {
	contract, err := bindAnchorContract(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &AnchorContractTransactor{contract: contract}, nil
}

// NewAnchorContractFilterer creates a new log filterer instance of AnchorContract, bound to a specific deployed contract.
func NewAnchorContractFilterer(address common.Address, filterer bind.ContractFilterer) (*AnchorContractFilterer, error) {
	contract, err := bindAnchorContract(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &AnchorContractFilterer{contract: contract}, nil
}

// bindAnchorContract binds a generic wrapper to an already deployed contract.
func bindAnchorContract(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(AnchorContractABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_AnchorContract *AnchorContractRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _AnchorContract.Contract.AnchorContractCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_AnchorContract *AnchorContractRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _AnchorContract.Contract.AnchorContractTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_AnchorContract *AnchorContractRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _AnchorContract.Contract.AnchorContractTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_AnchorContract *AnchorContractCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _AnchorContract.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_AnchorContract *AnchorContractTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _AnchorContract.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_AnchorContract *AnchorContractTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _AnchorContract.Contract.contract.Transact(opts, method, params...)
}

// GetAnchorByID is a free data retrieval call binding the contract method 0x32bf361b.
//
// Solidity: function getAnchorById(uint256 id) constant returns(uint256 anchorId, bytes32 documentRoot, uint32 blockNumber)
func (_AnchorContract *AnchorContractCaller) GetAnchorById(opts *bind.CallOpts, id *big.Int) (struct {
	AnchorId     *big.Int
	DocumentRoot [32]byte
	BlockNumber  uint32
}, error) {
	ret := new(struct {
		AnchorId     *big.Int
		DocumentRoot [32]byte
		BlockNumber  uint32
	})
	out := ret
	err := _AnchorContract.contract.Call(opts, out, "getAnchorById", id)
	return *ret, err
}

// GetAnchorByID is a free data retrieval call binding the contract method 0x32bf361b.
//
// Solidity: function getAnchorById(uint256 id) constant returns(uint256 anchorId, bytes32 documentRoot, uint32 blockNumber)
func (_AnchorContract *AnchorContractSession) GetAnchorById(id *big.Int) (struct {
	AnchorId     *big.Int
	DocumentRoot [32]byte
	BlockNumber  uint32
}, error) {
	return _AnchorContract.Contract.GetAnchorById(&_AnchorContract.CallOpts, id)
}

// GetAnchorByID is a free data retrieval call binding the contract method 0x32bf361b.
//
// Solidity: function getAnchorById(uint256 id) constant returns(uint256 anchorId, bytes32 documentRoot, uint32 blockNumber)
func (_AnchorContract *AnchorContractCallerSession) GetAnchorById(id *big.Int) (struct {
	AnchorId     *big.Int
	DocumentRoot [32]byte
	BlockNumber  uint32
}, error) {
	return _AnchorContract.Contract.GetAnchorById(&_AnchorContract.CallOpts, id)
}

// HasValidPreCommit is a free data retrieval call binding the contract method 0xb5c7d034.
//
// Solidity: function hasValidPreCommit(uint256 anchorId) constant returns(bool valid)
func (_AnchorContract *AnchorContractCaller) HasValidPreCommit(opts *bind.CallOpts, anchorId *big.Int) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _AnchorContract.contract.Call(opts, out, "hasValidPreCommit", anchorId)
	return *ret0, err
}

// HasValidPreCommit is a free data retrieval call binding the contract method 0xb5c7d034.
//
// Solidity: function hasValidPreCommit(uint256 anchorId) constant returns(bool valid)
func (_AnchorContract *AnchorContractSession) HasValidPreCommit(anchorId *big.Int) (bool, error) {
	return _AnchorContract.Contract.HasValidPreCommit(&_AnchorContract.CallOpts, anchorId)
}

// HasValidPreCommit is a free data retrieval call binding the contract method 0xb5c7d034.
//
// Solidity: function hasValidPreCommit(uint256 anchorId) constant returns(bool valid)
func (_AnchorContract *AnchorContractCallerSession) HasValidPreCommit(anchorId *big.Int) (bool, error) {
	return _AnchorContract.Contract.HasValidPreCommit(&_AnchorContract.CallOpts, anchorId)
}

// Commit is a paid mutator transaction binding the contract method 0x1cd98bce.
//
// Solidity: function commit(uint256 anchorIdPreImage, bytes32 documentRoot, bytes32 proof) returns()
func (_AnchorContract *AnchorContractTransactor) Commit(opts *bind.TransactOpts, anchorIdPreImage *big.Int, documentRoot [32]byte, proof [32]byte) (*types.Transaction, error) {
	return _AnchorContract.contract.Transact(opts, "commit", anchorIdPreImage, documentRoot, proof)
}

// Commit is a paid mutator transaction binding the contract method 0x1cd98bce.
//
// Solidity: function commit(uint256 anchorIdPreImage, bytes32 documentRoot, bytes32 proof) returns()
func (_AnchorContract *AnchorContractSession) Commit(anchorIdPreImage *big.Int, documentRoot [32]byte, proof [32]byte) (*types.Transaction, error) {
	return _AnchorContract.Contract.Commit(&_AnchorContract.TransactOpts, anchorIdPreImage, documentRoot, proof)
}

// Commit is a paid mutator transaction binding the contract method 0x1cd98bce.
//
// Solidity: function commit(uint256 anchorIdPreImage, bytes32 documentRoot, bytes32 proof) returns()
func (_AnchorContract *AnchorContractTransactorSession) Commit(anchorIdPreImage *big.Int, documentRoot [32]byte, proof [32]byte) (*types.Transaction, error) {
	return _AnchorContract.Contract.Commit(&_AnchorContract.TransactOpts, anchorIdPreImage, documentRoot, proof)
}

// PreCommit is a paid mutator transaction binding the contract method 0x53b36015.
//
// Solidity: function preCommit(uint256 anchorId, bytes32 signingRoot) returns()
func (_AnchorContract *AnchorContractTransactor) PreCommit(opts *bind.TransactOpts, anchorId *big.Int, signingRoot [32]byte) (*types.Transaction, error) {
	return _AnchorContract.contract.Transact(opts, "preCommit", anchorId, signingRoot)
}

// PreCommit is a paid mutator transaction binding the contract method 0x53b36015.
//
// Solidity: function preCommit(uint256 anchorId, bytes32 signingRoot) returns()
func (_AnchorContract *AnchorContractSession) PreCommit(anchorId *big.Int, signingRoot [32]byte) (*types.Transaction, error) {
	return _AnchorContract.Contract.PreCommit(&_AnchorContract.TransactOpts, anchorId, signingRoot)
}

// PreCommit is a paid mutator transaction binding the contract method 0x53b36015.
//
// Solidity: function preCommit(uint256 anchorId, bytes32 signingRoot) returns()
func (_AnchorContract *AnchorContractTransactorSession) PreCommit(anchorId *big.Int, signingRoot [32]byte) (*types.Transaction, error) {
	return _AnchorContract.Contract.PreCommit(&_AnchorContract.TransactOpts, anchorId, signingRoot)
}

// AnchorContractAnchorCommittedIterator is returned from FilterAnchorCommitted and is used to iterate over the raw logs and unpacked data for AnchorCommitted events raised by the AnchorContract contract.
type AnchorContractAnchorCommittedIterator struct {
	Event *AnchorContractAnchorCommitted // Event containing the contract specifics and raw log

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
func (it *AnchorContractAnchorCommittedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AnchorContractAnchorCommitted)
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
		it.Event = new(AnchorContractAnchorCommitted)
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
func (it *AnchorContractAnchorCommittedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AnchorContractAnchorCommittedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AnchorContractAnchorCommitted represents a AnchorCommitted event raised by the AnchorContract contract.
type AnchorContractAnchorCommitted struct {
	From         common.Address
	AnchorId     *big.Int
	DocumentRoot [32]byte
	BlockHeight  uint32
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterAnchorCommitted is a free log retrieval operation binding the contract event 0xd1eb81d62e07e99a310f0f4c9a107a644e475be1f4b7eaa3d5c731c140195ee9.
//
// Solidity: event AnchorCommitted(address indexed from, uint256 indexed anchorId, bytes32 documentRoot, uint32 blockHeight)
func (_AnchorContract *AnchorContractFilterer) FilterAnchorCommitted(opts *bind.FilterOpts, from []common.Address, anchorId []*big.Int) (*AnchorContractAnchorCommittedIterator, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var anchorIdRule []interface{}
	for _, anchorIdItem := range anchorId {
		anchorIdRule = append(anchorIdRule, anchorIdItem)
	}

	logs, sub, err := _AnchorContract.contract.FilterLogs(opts, "AnchorCommitted", fromRule, anchorIdRule)
	if err != nil {
		return nil, err
	}
	return &AnchorContractAnchorCommittedIterator{contract: _AnchorContract.contract, event: "AnchorCommitted", logs: logs, sub: sub}, nil
}

// WatchAnchorCommitted is a free log subscription operation binding the contract event 0xd1eb81d62e07e99a310f0f4c9a107a644e475be1f4b7eaa3d5c731c140195ee9.
//
// Solidity: event AnchorCommitted(address indexed from, uint256 indexed anchorId, bytes32 documentRoot, uint32 blockHeight)
func (_AnchorContract *AnchorContractFilterer) WatchAnchorCommitted(opts *bind.WatchOpts, sink chan<- *AnchorContractAnchorCommitted, from []common.Address, anchorId []*big.Int) (event.Subscription, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var anchorIdRule []interface{}
	for _, anchorIdItem := range anchorId {
		anchorIdRule = append(anchorIdRule, anchorIdItem)
	}

	logs, sub, err := _AnchorContract.contract.WatchLogs(opts, "AnchorCommitted", fromRule, anchorIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AnchorContractAnchorCommitted)
				if err := _AnchorContract.contract.UnpackLog(event, "AnchorCommitted", log); err != nil {
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

// AnchorContractAnchorPreCommittedIterator is returned from FilterAnchorPreCommitted and is used to iterate over the raw logs and unpacked data for AnchorPreCommitted events raised by the AnchorContract contract.
type AnchorContractAnchorPreCommittedIterator struct {
	Event *AnchorContractAnchorPreCommitted // Event containing the contract specifics and raw log

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
func (it *AnchorContractAnchorPreCommittedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AnchorContractAnchorPreCommitted)
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
		it.Event = new(AnchorContractAnchorPreCommitted)
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
func (it *AnchorContractAnchorPreCommittedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AnchorContractAnchorPreCommittedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AnchorContractAnchorPreCommitted represents a AnchorPreCommitted event raised by the AnchorContract contract.
type AnchorContractAnchorPreCommitted struct {
	From        common.Address
	AnchorId    *big.Int
	BlockHeight uint32
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterAnchorPreCommitted is a free log retrieval operation binding the contract event 0xaa2928be4e330731bc1f0289edebfc72ccb9979ffc703a3de4edd8ea760462da.
//
// Solidity: event AnchorPreCommitted(address indexed from, uint256 indexed anchorId, uint32 blockHeight)
func (_AnchorContract *AnchorContractFilterer) FilterAnchorPreCommitted(opts *bind.FilterOpts, from []common.Address, anchorId []*big.Int) (*AnchorContractAnchorPreCommittedIterator, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var anchorIdRule []interface{}
	for _, anchorIdItem := range anchorId {
		anchorIdRule = append(anchorIdRule, anchorIdItem)
	}

	logs, sub, err := _AnchorContract.contract.FilterLogs(opts, "AnchorPreCommitted", fromRule, anchorIdRule)
	if err != nil {
		return nil, err
	}
	return &AnchorContractAnchorPreCommittedIterator{contract: _AnchorContract.contract, event: "AnchorPreCommitted", logs: logs, sub: sub}, nil
}

// WatchAnchorPreCommitted is a free log subscription operation binding the contract event 0xaa2928be4e330731bc1f0289edebfc72ccb9979ffc703a3de4edd8ea760462da.
//
// Solidity: event AnchorPreCommitted(address indexed from, uint256 indexed anchorId, uint32 blockHeight)
func (_AnchorContract *AnchorContractFilterer) WatchAnchorPreCommitted(opts *bind.WatchOpts, sink chan<- *AnchorContractAnchorPreCommitted, from []common.Address, anchorId []*big.Int) (event.Subscription, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var anchorIdRule []interface{}
	for _, anchorIdItem := range anchorId {
		anchorIdRule = append(anchorIdRule, anchorIdItem)
	}

	logs, sub, err := _AnchorContract.contract.WatchLogs(opts, "AnchorPreCommitted", fromRule, anchorIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AnchorContractAnchorPreCommitted)
				if err := _AnchorContract.contract.UnpackLog(event, "AnchorPreCommitted", log); err != nil {
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
