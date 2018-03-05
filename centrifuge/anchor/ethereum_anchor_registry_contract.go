// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package anchor

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

// EthereumAnchorRegistryContractABI is the input ABI used to generate the binding from.
const EthereumAnchorRegistryContractABI = "[{\"constant\":true,\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"anchors\",\"outputs\":[{\"name\":\"identifier\",\"type\":\"bytes32\"},{\"name\":\"merkleRoot\",\"type\":\"bytes32\"},{\"name\":\"timestamp\",\"type\":\"bytes32\"},{\"name\":\"schemaVersion\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"identifier\",\"type\":\"bytes32\"},{\"indexed\":true,\"name\":\"rootHash\",\"type\":\"bytes32\"},{\"indexed\":false,\"name\":\"timestamp\",\"type\":\"bytes32\"},{\"indexed\":false,\"name\":\"anchorSchemaVersion\",\"type\":\"uint256\"}],\"name\":\"AnchorRegistered\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"constant\":false,\"inputs\":[{\"name\":\"identifier\",\"type\":\"bytes32\"},{\"name\":\"merkleRoot\",\"type\":\"bytes32\"},{\"name\":\"anchorSchemaVersion\",\"type\":\"uint256\"}],\"name\":\"registerAnchor\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"identifier\",\"type\":\"bytes32\"}],\"name\":\"getAnchorById\",\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\"},{\"name\":\"\",\"type\":\"bytes32\"},{\"name\":\"\",\"type\":\"bytes32\"},{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"}]"

// EthereumAnchorRegistryContract is an auto generated Go binding around an Ethereum contract.
type EthereumAnchorRegistryContract struct {
	EthereumAnchorRegistryContractCaller     // Read-only binding to the contract
	EthereumAnchorRegistryContractTransactor // Write-only binding to the contract
	EthereumAnchorRegistryContractFilterer   // Log filterer for contract events
}

// EthereumAnchorRegistryContractCaller is an auto generated read-only Go binding around an Ethereum contract.
type EthereumAnchorRegistryContractCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// EthereumAnchorRegistryContractTransactor is an auto generated write-only Go binding around an Ethereum contract.
type EthereumAnchorRegistryContractTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// EthereumAnchorRegistryContractFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type EthereumAnchorRegistryContractFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// EthereumAnchorRegistryContractSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type EthereumAnchorRegistryContractSession struct {
	Contract     *EthereumAnchorRegistryContract // Generic contract binding to set the session for
	CallOpts     bind.CallOpts                   // Call options to use throughout this session
	TransactOpts bind.TransactOpts               // Transaction auth options to use throughout this session
}

// EthereumAnchorRegistryContractCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type EthereumAnchorRegistryContractCallerSession struct {
	Contract *EthereumAnchorRegistryContractCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts                         // Call options to use throughout this session
}

// EthereumAnchorRegistryContractTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type EthereumAnchorRegistryContractTransactorSession struct {
	Contract     *EthereumAnchorRegistryContractTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts                         // Transaction auth options to use throughout this session
}

// EthereumAnchorRegistryContractRaw is an auto generated low-level Go binding around an Ethereum contract.
type EthereumAnchorRegistryContractRaw struct {
	Contract *EthereumAnchorRegistryContract // Generic contract binding to access the raw methods on
}

// EthereumAnchorRegistryContractCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type EthereumAnchorRegistryContractCallerRaw struct {
	Contract *EthereumAnchorRegistryContractCaller // Generic read-only contract binding to access the raw methods on
}

// EthereumAnchorRegistryContractTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type EthereumAnchorRegistryContractTransactorRaw struct {
	Contract *EthereumAnchorRegistryContractTransactor // Generic write-only contract binding to access the raw methods on
}

// NewEthereumAnchorRegistryContract creates a new instance of EthereumAnchorRegistryContract, bound to a specific deployed contract.
func NewEthereumAnchorRegistryContract(address common.Address, backend bind.ContractBackend) (*EthereumAnchorRegistryContract, error) {
	contract, err := bindEthereumAnchorRegistryContract(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &EthereumAnchorRegistryContract{EthereumAnchorRegistryContractCaller: EthereumAnchorRegistryContractCaller{contract: contract}, EthereumAnchorRegistryContractTransactor: EthereumAnchorRegistryContractTransactor{contract: contract}, EthereumAnchorRegistryContractFilterer: EthereumAnchorRegistryContractFilterer{contract: contract}}, nil
}

// NewEthereumAnchorRegistryContractCaller creates a new read-only instance of EthereumAnchorRegistryContract, bound to a specific deployed contract.
func NewEthereumAnchorRegistryContractCaller(address common.Address, caller bind.ContractCaller) (*EthereumAnchorRegistryContractCaller, error) {
	contract, err := bindEthereumAnchorRegistryContract(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &EthereumAnchorRegistryContractCaller{contract: contract}, nil
}

// NewEthereumAnchorRegistryContractTransactor creates a new write-only instance of EthereumAnchorRegistryContract, bound to a specific deployed contract.
func NewEthereumAnchorRegistryContractTransactor(address common.Address, transactor bind.ContractTransactor) (*EthereumAnchorRegistryContractTransactor, error) {
	contract, err := bindEthereumAnchorRegistryContract(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &EthereumAnchorRegistryContractTransactor{contract: contract}, nil
}

// NewEthereumAnchorRegistryContractFilterer creates a new log filterer instance of EthereumAnchorRegistryContract, bound to a specific deployed contract.
func NewEthereumAnchorRegistryContractFilterer(address common.Address, filterer bind.ContractFilterer) (*EthereumAnchorRegistryContractFilterer, error) {
	contract, err := bindEthereumAnchorRegistryContract(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &EthereumAnchorRegistryContractFilterer{contract: contract}, nil
}

// bindEthereumAnchorRegistryContract binds a generic wrapper to an already deployed contract.
func bindEthereumAnchorRegistryContract(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(EthereumAnchorRegistryContractABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_EthereumAnchorRegistryContract *EthereumAnchorRegistryContractRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _EthereumAnchorRegistryContract.Contract.EthereumAnchorRegistryContractCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_EthereumAnchorRegistryContract *EthereumAnchorRegistryContractRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _EthereumAnchorRegistryContract.Contract.EthereumAnchorRegistryContractTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_EthereumAnchorRegistryContract *EthereumAnchorRegistryContractRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _EthereumAnchorRegistryContract.Contract.EthereumAnchorRegistryContractTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_EthereumAnchorRegistryContract *EthereumAnchorRegistryContractCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _EthereumAnchorRegistryContract.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_EthereumAnchorRegistryContract *EthereumAnchorRegistryContractTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _EthereumAnchorRegistryContract.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_EthereumAnchorRegistryContract *EthereumAnchorRegistryContractTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _EthereumAnchorRegistryContract.Contract.contract.Transact(opts, method, params...)
}

// Anchors is a free data retrieval call binding the contract method 0xb01b6d53.
//
// Solidity: function anchors( bytes32) constant returns(identifier bytes32, merkleRoot bytes32, timestamp bytes32, schemaVersion uint256)
func (_EthereumAnchorRegistryContract *EthereumAnchorRegistryContractCaller) Anchors(opts *bind.CallOpts, arg0 [32]byte) (struct {
	Identifier    [32]byte
	MerkleRoot    [32]byte
	Timestamp     [32]byte
	SchemaVersion *big.Int
}, error) {
	ret := new(struct {
		Identifier    [32]byte
		MerkleRoot    [32]byte
		Timestamp     [32]byte
		SchemaVersion *big.Int
	})
	out := ret
	err := _EthereumAnchorRegistryContract.contract.Call(opts, out, "anchors", arg0)
	return *ret, err
}

// Anchors is a free data retrieval call binding the contract method 0xb01b6d53.
//
// Solidity: function anchors( bytes32) constant returns(identifier bytes32, merkleRoot bytes32, timestamp bytes32, schemaVersion uint256)
func (_EthereumAnchorRegistryContract *EthereumAnchorRegistryContractSession) Anchors(arg0 [32]byte) (struct {
	Identifier    [32]byte
	MerkleRoot    [32]byte
	Timestamp     [32]byte
	SchemaVersion *big.Int
}, error) {
	return _EthereumAnchorRegistryContract.Contract.Anchors(&_EthereumAnchorRegistryContract.CallOpts, arg0)
}

// Anchors is a free data retrieval call binding the contract method 0xb01b6d53.
//
// Solidity: function anchors( bytes32) constant returns(identifier bytes32, merkleRoot bytes32, timestamp bytes32, schemaVersion uint256)
func (_EthereumAnchorRegistryContract *EthereumAnchorRegistryContractCallerSession) Anchors(arg0 [32]byte) (struct {
	Identifier    [32]byte
	MerkleRoot    [32]byte
	Timestamp     [32]byte
	SchemaVersion *big.Int
}, error) {
	return _EthereumAnchorRegistryContract.Contract.Anchors(&_EthereumAnchorRegistryContract.CallOpts, arg0)
}

// GetAnchorById is a free data retrieval call binding the contract method 0x04d466b2.
//
// Solidity: function getAnchorById(identifier bytes32) constant returns(bytes32, bytes32, bytes32, uint256)
func (_EthereumAnchorRegistryContract *EthereumAnchorRegistryContractCaller) GetAnchorById(opts *bind.CallOpts, identifier [32]byte) ([32]byte, [32]byte, [32]byte, *big.Int, error) {
	var (
		ret0 = new([32]byte)
		ret1 = new([32]byte)
		ret2 = new([32]byte)
		ret3 = new(*big.Int)
	)
	out := &[]interface{}{
		ret0,
		ret1,
		ret2,
		ret3,
	}
	err := _EthereumAnchorRegistryContract.contract.Call(opts, out, "getAnchorById", identifier)
	return *ret0, *ret1, *ret2, *ret3, err
}

// GetAnchorById is a free data retrieval call binding the contract method 0x04d466b2.
//
// Solidity: function getAnchorById(identifier bytes32) constant returns(bytes32, bytes32, bytes32, uint256)
func (_EthereumAnchorRegistryContract *EthereumAnchorRegistryContractSession) GetAnchorById(identifier [32]byte) ([32]byte, [32]byte, [32]byte, *big.Int, error) {
	return _EthereumAnchorRegistryContract.Contract.GetAnchorById(&_EthereumAnchorRegistryContract.CallOpts, identifier)
}

// GetAnchorById is a free data retrieval call binding the contract method 0x04d466b2.
//
// Solidity: function getAnchorById(identifier bytes32) constant returns(bytes32, bytes32, bytes32, uint256)
func (_EthereumAnchorRegistryContract *EthereumAnchorRegistryContractCallerSession) GetAnchorById(identifier [32]byte) ([32]byte, [32]byte, [32]byte, *big.Int, error) {
	return _EthereumAnchorRegistryContract.Contract.GetAnchorById(&_EthereumAnchorRegistryContract.CallOpts, identifier)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() constant returns(address)
func (_EthereumAnchorRegistryContract *EthereumAnchorRegistryContractCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _EthereumAnchorRegistryContract.contract.Call(opts, out, "owner")
	return *ret0, err
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() constant returns(address)
func (_EthereumAnchorRegistryContract *EthereumAnchorRegistryContractSession) Owner() (common.Address, error) {
	return _EthereumAnchorRegistryContract.Contract.Owner(&_EthereumAnchorRegistryContract.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() constant returns(address)
func (_EthereumAnchorRegistryContract *EthereumAnchorRegistryContractCallerSession) Owner() (common.Address, error) {
	return _EthereumAnchorRegistryContract.Contract.Owner(&_EthereumAnchorRegistryContract.CallOpts)
}

// RegisterAnchor is a paid mutator transaction binding the contract method 0x77bfeb8b.
//
// Solidity: function registerAnchor(identifier bytes32, merkleRoot bytes32, anchorSchemaVersion uint256) returns()
func (_EthereumAnchorRegistryContract *EthereumAnchorRegistryContractTransactor) RegisterAnchor(opts *bind.TransactOpts, identifier [32]byte, merkleRoot [32]byte, anchorSchemaVersion *big.Int) (*types.Transaction, error) {
	return _EthereumAnchorRegistryContract.contract.Transact(opts, "registerAnchor", identifier, merkleRoot, anchorSchemaVersion)
}

// RegisterAnchor is a paid mutator transaction binding the contract method 0x77bfeb8b.
//
// Solidity: function registerAnchor(identifier bytes32, merkleRoot bytes32, anchorSchemaVersion uint256) returns()
func (_EthereumAnchorRegistryContract *EthereumAnchorRegistryContractSession) RegisterAnchor(identifier [32]byte, merkleRoot [32]byte, anchorSchemaVersion *big.Int) (*types.Transaction, error) {
	return _EthereumAnchorRegistryContract.Contract.RegisterAnchor(&_EthereumAnchorRegistryContract.TransactOpts, identifier, merkleRoot, anchorSchemaVersion)
}

// RegisterAnchor is a paid mutator transaction binding the contract method 0x77bfeb8b.
//
// Solidity: function registerAnchor(identifier bytes32, merkleRoot bytes32, anchorSchemaVersion uint256) returns()
func (_EthereumAnchorRegistryContract *EthereumAnchorRegistryContractTransactorSession) RegisterAnchor(identifier [32]byte, merkleRoot [32]byte, anchorSchemaVersion *big.Int) (*types.Transaction, error) {
	return _EthereumAnchorRegistryContract.Contract.RegisterAnchor(&_EthereumAnchorRegistryContract.TransactOpts, identifier, merkleRoot, anchorSchemaVersion)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(newOwner address) returns()
func (_EthereumAnchorRegistryContract *EthereumAnchorRegistryContractTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _EthereumAnchorRegistryContract.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(newOwner address) returns()
func (_EthereumAnchorRegistryContract *EthereumAnchorRegistryContractSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _EthereumAnchorRegistryContract.Contract.TransferOwnership(&_EthereumAnchorRegistryContract.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(newOwner address) returns()
func (_EthereumAnchorRegistryContract *EthereumAnchorRegistryContractTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _EthereumAnchorRegistryContract.Contract.TransferOwnership(&_EthereumAnchorRegistryContract.TransactOpts, newOwner)
}

// EthereumAnchorRegistryContractAnchorRegisteredIterator is returned from FilterAnchorRegistered and is used to iterate over the raw logs and unpacked data for AnchorRegistered events raised by the EthereumAnchorRegistryContract contract.
type EthereumAnchorRegistryContractAnchorRegisteredIterator struct {
	Event *EthereumAnchorRegistryContractAnchorRegistered // Event containing the contract specifics and raw log

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
func (it *EthereumAnchorRegistryContractAnchorRegisteredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(EthereumAnchorRegistryContractAnchorRegistered)
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
		it.Event = new(EthereumAnchorRegistryContractAnchorRegistered)
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
func (it *EthereumAnchorRegistryContractAnchorRegisteredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *EthereumAnchorRegistryContractAnchorRegisteredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// EthereumAnchorRegistryContractAnchorRegistered represents a AnchorRegistered event raised by the EthereumAnchorRegistryContract contract.
type EthereumAnchorRegistryContractAnchorRegistered struct {
	From                common.Address
	Identifier          [32]byte
	RootHash            [32]byte
	Timestamp           [32]byte
	AnchorSchemaVersion *big.Int
	Raw                 types.Log // Blockchain specific contextual infos
}

// FilterAnchorRegistered is a free log retrieval operation binding the contract event 0xb5b2dfde590bcc9ed7e31e337bc1780f2c3e052da6e82cb35e7232a916ac3550.
//
// Solidity: event AnchorRegistered(from indexed address, identifier indexed bytes32, rootHash indexed bytes32, timestamp bytes32, anchorSchemaVersion uint256)
func (_EthereumAnchorRegistryContract *EthereumAnchorRegistryContractFilterer) FilterAnchorRegistered(opts *bind.FilterOpts, from []common.Address, identifier [][32]byte, rootHash [][32]byte) (*EthereumAnchorRegistryContractAnchorRegisteredIterator, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var identifierRule []interface{}
	for _, identifierItem := range identifier {
		identifierRule = append(identifierRule, identifierItem)
	}
	var rootHashRule []interface{}
	for _, rootHashItem := range rootHash {
		rootHashRule = append(rootHashRule, rootHashItem)
	}

	logs, sub, err := _EthereumAnchorRegistryContract.contract.FilterLogs(opts, "AnchorRegistered", fromRule, identifierRule, rootHashRule)
	if err != nil {
		return nil, err
	}
	return &EthereumAnchorRegistryContractAnchorRegisteredIterator{contract: _EthereumAnchorRegistryContract.contract, event: "AnchorRegistered", logs: logs, sub: sub}, nil
}

// WatchAnchorRegistered is a free log subscription operation binding the contract event 0xb5b2dfde590bcc9ed7e31e337bc1780f2c3e052da6e82cb35e7232a916ac3550.
//
// Solidity: event AnchorRegistered(from indexed address, identifier indexed bytes32, rootHash indexed bytes32, timestamp bytes32, anchorSchemaVersion uint256)
func (_EthereumAnchorRegistryContract *EthereumAnchorRegistryContractFilterer) WatchAnchorRegistered(opts *bind.WatchOpts, sink chan<- *EthereumAnchorRegistryContractAnchorRegistered, from []common.Address, identifier [][32]byte, rootHash [][32]byte) (event.Subscription, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var identifierRule []interface{}
	for _, identifierItem := range identifier {
		identifierRule = append(identifierRule, identifierItem)
	}
	var rootHashRule []interface{}
	for _, rootHashItem := range rootHash {
		rootHashRule = append(rootHashRule, rootHashItem)
	}

	logs, sub, err := _EthereumAnchorRegistryContract.contract.WatchLogs(opts, "AnchorRegistered", fromRule, identifierRule, rootHashRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(EthereumAnchorRegistryContractAnchorRegistered)
				if err := _EthereumAnchorRegistryContract.contract.UnpackLog(event, "AnchorRegistered", log); err != nil {
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

// EthereumAnchorRegistryContractOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the EthereumAnchorRegistryContract contract.
type EthereumAnchorRegistryContractOwnershipTransferredIterator struct {
	Event *EthereumAnchorRegistryContractOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *EthereumAnchorRegistryContractOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(EthereumAnchorRegistryContractOwnershipTransferred)
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
		it.Event = new(EthereumAnchorRegistryContractOwnershipTransferred)
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
func (it *EthereumAnchorRegistryContractOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *EthereumAnchorRegistryContractOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// EthereumAnchorRegistryContractOwnershipTransferred represents a OwnershipTransferred event raised by the EthereumAnchorRegistryContract contract.
type EthereumAnchorRegistryContractOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(previousOwner indexed address, newOwner indexed address)
func (_EthereumAnchorRegistryContract *EthereumAnchorRegistryContractFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*EthereumAnchorRegistryContractOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _EthereumAnchorRegistryContract.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &EthereumAnchorRegistryContractOwnershipTransferredIterator{contract: _EthereumAnchorRegistryContract.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(previousOwner indexed address, newOwner indexed address)
func (_EthereumAnchorRegistryContract *EthereumAnchorRegistryContractFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *EthereumAnchorRegistryContractOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _EthereumAnchorRegistryContract.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(EthereumAnchorRegistryContractOwnershipTransferred)
				if err := _EthereumAnchorRegistryContract.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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
