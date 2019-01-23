// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package did

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

// IdentityContractABI is the input ABI used to generate the binding from.
const IdentityContractABI = "[{\"constant\":true,\"inputs\":[{\"name\":\"_key\",\"type\":\"bytes32\"}],\"name\":\"getKey\",\"outputs\":[{\"name\":\"key\",\"type\":\"bytes32\"},{\"name\":\"purposes\",\"type\":\"uint256[]\"},{\"name\":\"revokedAt\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_key\",\"type\":\"bytes32\"},{\"name\":\"_purposes\",\"type\":\"uint256[]\"},{\"name\":\"_keyType\",\"type\":\"uint256\"}],\"name\":\"addMultiPurposeKey\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_key\",\"type\":\"bytes32\"},{\"name\":\"_purpose\",\"type\":\"uint256\"},{\"name\":\"_keyType\",\"type\":\"uint256\"}],\"name\":\"addKey\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_key\",\"type\":\"bytes32\"}],\"name\":\"revokeKey\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"addressToKey\",\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\"}],\"payable\":false,\"stateMutability\":\"pure\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_purpose\",\"type\":\"uint256\"}],\"name\":\"getKeysByPurpose\",\"outputs\":[{\"name\":\"\",\"type\":\"bytes32[]\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_key\",\"type\":\"bytes32\"},{\"name\":\"_purpose\",\"type\":\"uint256\"}],\"name\":\"keyHasPurpose\",\"outputs\":[{\"name\":\"found\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"name\":\"owner\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"key\",\"type\":\"bytes32\"},{\"indexed\":true,\"name\":\"purpose\",\"type\":\"uint256\"},{\"indexed\":true,\"name\":\"keyType\",\"type\":\"uint256\"}],\"name\":\"KeyAdded\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"key\",\"type\":\"bytes32\"},{\"indexed\":true,\"name\":\"revokedAt\",\"type\":\"uint256\"},{\"indexed\":true,\"name\":\"keyType\",\"type\":\"uint256\"}],\"name\":\"KeyRevoked\",\"type\":\"event\"},{\"constant\":false,\"inputs\":[{\"name\":\"_to\",\"type\":\"address\"},{\"name\":\"_value\",\"type\":\"uint256\"},{\"name\":\"_data\",\"type\":\"bytes\"}],\"name\":\"execute\",\"outputs\":[{\"name\":\"success\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_toSign\",\"type\":\"bytes32\"},{\"name\":\"_signature\",\"type\":\"bytes\"},{\"name\":\"_purpose\",\"type\":\"uint256\"}],\"name\":\"isSignedWithPurpose\",\"outputs\":[{\"name\":\"valid\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"}]"

// IdentityContract is an auto generated Go binding around an Ethereum contract.
type IdentityContract struct {
	IdentityContractCaller     // Read-only binding to the contract
	IdentityContractTransactor // Write-only binding to the contract
	IdentityContractFilterer   // Log filterer for contract events
}

// IdentityContractCaller is an auto generated read-only Go binding around an Ethereum contract.
type IdentityContractCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// IdentityContractTransactor is an auto generated write-only Go binding around an Ethereum contract.
type IdentityContractTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// IdentityContractFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type IdentityContractFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// IdentityContractSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type IdentityContractSession struct {
	Contract     *IdentityContract // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// IdentityContractCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type IdentityContractCallerSession struct {
	Contract *IdentityContractCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts           // Call options to use throughout this session
}

// IdentityContractTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type IdentityContractTransactorSession struct {
	Contract     *IdentityContractTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts           // Transaction auth options to use throughout this session
}

// IdentityContractRaw is an auto generated low-level Go binding around an Ethereum contract.
type IdentityContractRaw struct {
	Contract *IdentityContract // Generic contract binding to access the raw methods on
}

// IdentityContractCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type IdentityContractCallerRaw struct {
	Contract *IdentityContractCaller // Generic read-only contract binding to access the raw methods on
}

// IdentityContractTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type IdentityContractTransactorRaw struct {
	Contract *IdentityContractTransactor // Generic write-only contract binding to access the raw methods on
}

// NewIdentityContract creates a new instance of IdentityContract, bound to a specific deployed contract.
func NewIdentityContract(address common.Address, backend bind.ContractBackend) (*IdentityContract, error) {
	contract, err := bindIdentityContract(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &IdentityContract{IdentityContractCaller: IdentityContractCaller{contract: contract}, IdentityContractTransactor: IdentityContractTransactor{contract: contract}, IdentityContractFilterer: IdentityContractFilterer{contract: contract}}, nil
}

// NewIdentityContractCaller creates a new read-only instance of IdentityContract, bound to a specific deployed contract.
func NewIdentityContractCaller(address common.Address, caller bind.ContractCaller) (*IdentityContractCaller, error) {
	contract, err := bindIdentityContract(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &IdentityContractCaller{contract: contract}, nil
}

// NewIdentityContractTransactor creates a new write-only instance of IdentityContract, bound to a specific deployed contract.
func NewIdentityContractTransactor(address common.Address, transactor bind.ContractTransactor) (*IdentityContractTransactor, error) {
	contract, err := bindIdentityContract(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &IdentityContractTransactor{contract: contract}, nil
}

// NewIdentityContractFilterer creates a new log filterer instance of IdentityContract, bound to a specific deployed contract.
func NewIdentityContractFilterer(address common.Address, filterer bind.ContractFilterer) (*IdentityContractFilterer, error) {
	contract, err := bindIdentityContract(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &IdentityContractFilterer{contract: contract}, nil
}

// bindIdentityContract binds a generic wrapper to an already deployed contract.
func bindIdentityContract(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(IdentityContractABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_IdentityContract *IdentityContractRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _IdentityContract.Contract.IdentityContractCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_IdentityContract *IdentityContractRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _IdentityContract.Contract.IdentityContractTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_IdentityContract *IdentityContractRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _IdentityContract.Contract.IdentityContractTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_IdentityContract *IdentityContractCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _IdentityContract.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_IdentityContract *IdentityContractTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _IdentityContract.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_IdentityContract *IdentityContractTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _IdentityContract.Contract.contract.Transact(opts, method, params...)
}

// AddressToKey is a free data retrieval call binding the contract method 0x574363c8.
//
// Solidity: function addressToKey(addr address) constant returns(bytes32)
func (_IdentityContract *IdentityContractCaller) AddressToKey(opts *bind.CallOpts, addr common.Address) ([32]byte, error) {
	var (
		ret0 = new([32]byte)
	)
	out := ret0
	err := _IdentityContract.contract.Call(opts, out, "addressToKey", addr)
	return *ret0, err
}

// AddressToKey is a free data retrieval call binding the contract method 0x574363c8.
//
// Solidity: function addressToKey(addr address) constant returns(bytes32)
func (_IdentityContract *IdentityContractSession) AddressToKey(addr common.Address) ([32]byte, error) {
	return _IdentityContract.Contract.AddressToKey(&_IdentityContract.CallOpts, addr)
}

// AddressToKey is a free data retrieval call binding the contract method 0x574363c8.
//
// Solidity: function addressToKey(addr address) constant returns(bytes32)
func (_IdentityContract *IdentityContractCallerSession) AddressToKey(addr common.Address) ([32]byte, error) {
	return _IdentityContract.Contract.AddressToKey(&_IdentityContract.CallOpts, addr)
}

// GetKey is a free data retrieval call binding the contract method 0x12aaac70.
//
// Solidity: function getKey(_key bytes32) constant returns(key bytes32, purposes uint256[], revokedAt uint256)
func (_IdentityContract *IdentityContractCaller) GetKey(opts *bind.CallOpts, _key [32]byte) (struct {
	Key       [32]byte
	Purposes  []*big.Int
	RevokedAt *big.Int
}, error) {
	ret := new(struct {
		Key       [32]byte
		Purposes  []*big.Int
		RevokedAt *big.Int
	})
	out := ret
	err := _IdentityContract.contract.Call(opts, out, "getKey", _key)
	return *ret, err
}

// GetKey is a free data retrieval call binding the contract method 0x12aaac70.
//
// Solidity: function getKey(_key bytes32) constant returns(key bytes32, purposes uint256[], revokedAt uint256)
func (_IdentityContract *IdentityContractSession) GetKey(_key [32]byte) (struct {
	Key       [32]byte
	Purposes  []*big.Int
	RevokedAt *big.Int
}, error) {
	return _IdentityContract.Contract.GetKey(&_IdentityContract.CallOpts, _key)
}

// GetKey is a free data retrieval call binding the contract method 0x12aaac70.
//
// Solidity: function getKey(_key bytes32) constant returns(key bytes32, purposes uint256[], revokedAt uint256)
func (_IdentityContract *IdentityContractCallerSession) GetKey(_key [32]byte) (struct {
	Key       [32]byte
	Purposes  []*big.Int
	RevokedAt *big.Int
}, error) {
	return _IdentityContract.Contract.GetKey(&_IdentityContract.CallOpts, _key)
}

// GetKeysByPurpose is a free data retrieval call binding the contract method 0x9010f726.
//
// Solidity: function getKeysByPurpose(_purpose uint256) constant returns(bytes32[])
func (_IdentityContract *IdentityContractCaller) GetKeysByPurpose(opts *bind.CallOpts, _purpose *big.Int) ([][32]byte, error) {
	var (
		ret0 = new([][32]byte)
	)
	out := ret0
	err := _IdentityContract.contract.Call(opts, out, "getKeysByPurpose", _purpose)
	return *ret0, err
}

// GetKeysByPurpose is a free data retrieval call binding the contract method 0x9010f726.
//
// Solidity: function getKeysByPurpose(_purpose uint256) constant returns(bytes32[])
func (_IdentityContract *IdentityContractSession) GetKeysByPurpose(_purpose *big.Int) ([][32]byte, error) {
	return _IdentityContract.Contract.GetKeysByPurpose(&_IdentityContract.CallOpts, _purpose)
}

// GetKeysByPurpose is a free data retrieval call binding the contract method 0x9010f726.
//
// Solidity: function getKeysByPurpose(_purpose uint256) constant returns(bytes32[])
func (_IdentityContract *IdentityContractCallerSession) GetKeysByPurpose(_purpose *big.Int) ([][32]byte, error) {
	return _IdentityContract.Contract.GetKeysByPurpose(&_IdentityContract.CallOpts, _purpose)
}

// IsSignedWithPurpose is a free data retrieval call binding the contract method 0x8699fd2b.
//
// Solidity: function isSignedWithPurpose(_toSign bytes32, _signature bytes, _purpose uint256) constant returns(valid bool)
func (_IdentityContract *IdentityContractCaller) IsSignedWithPurpose(opts *bind.CallOpts, _toSign [32]byte, _signature []byte, _purpose *big.Int) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _IdentityContract.contract.Call(opts, out, "isSignedWithPurpose", _toSign, _signature, _purpose)
	return *ret0, err
}

// IsSignedWithPurpose is a free data retrieval call binding the contract method 0x8699fd2b.
//
// Solidity: function isSignedWithPurpose(_toSign bytes32, _signature bytes, _purpose uint256) constant returns(valid bool)
func (_IdentityContract *IdentityContractSession) IsSignedWithPurpose(_toSign [32]byte, _signature []byte, _purpose *big.Int) (bool, error) {
	return _IdentityContract.Contract.IsSignedWithPurpose(&_IdentityContract.CallOpts, _toSign, _signature, _purpose)
}

// IsSignedWithPurpose is a free data retrieval call binding the contract method 0x8699fd2b.
//
// Solidity: function isSignedWithPurpose(_toSign bytes32, _signature bytes, _purpose uint256) constant returns(valid bool)
func (_IdentityContract *IdentityContractCallerSession) IsSignedWithPurpose(_toSign [32]byte, _signature []byte, _purpose *big.Int) (bool, error) {
	return _IdentityContract.Contract.IsSignedWithPurpose(&_IdentityContract.CallOpts, _toSign, _signature, _purpose)
}

// KeyHasPurpose is a free data retrieval call binding the contract method 0xd202158d.
//
// Solidity: function keyHasPurpose(_key bytes32, _purpose uint256) constant returns(found bool)
func (_IdentityContract *IdentityContractCaller) KeyHasPurpose(opts *bind.CallOpts, _key [32]byte, _purpose *big.Int) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _IdentityContract.contract.Call(opts, out, "keyHasPurpose", _key, _purpose)
	return *ret0, err
}

// KeyHasPurpose is a free data retrieval call binding the contract method 0xd202158d.
//
// Solidity: function keyHasPurpose(_key bytes32, _purpose uint256) constant returns(found bool)
func (_IdentityContract *IdentityContractSession) KeyHasPurpose(_key [32]byte, _purpose *big.Int) (bool, error) {
	return _IdentityContract.Contract.KeyHasPurpose(&_IdentityContract.CallOpts, _key, _purpose)
}

// KeyHasPurpose is a free data retrieval call binding the contract method 0xd202158d.
//
// Solidity: function keyHasPurpose(_key bytes32, _purpose uint256) constant returns(found bool)
func (_IdentityContract *IdentityContractCallerSession) KeyHasPurpose(_key [32]byte, _purpose *big.Int) (bool, error) {
	return _IdentityContract.Contract.KeyHasPurpose(&_IdentityContract.CallOpts, _key, _purpose)
}

// AddKey is a paid mutator transaction binding the contract method 0x1d381240.
//
// Solidity: function addKey(_key bytes32, _purpose uint256, _keyType uint256) returns()
func (_IdentityContract *IdentityContractTransactor) AddKey(opts *bind.TransactOpts, _key [32]byte, _purpose *big.Int, _keyType *big.Int) (*types.Transaction, error) {
	return _IdentityContract.contract.Transact(opts, "addKey", _key, _purpose, _keyType)
}

// AddKey is a paid mutator transaction binding the contract method 0x1d381240.
//
// Solidity: function addKey(_key bytes32, _purpose uint256, _keyType uint256) returns()
func (_IdentityContract *IdentityContractSession) AddKey(_key [32]byte, _purpose *big.Int, _keyType *big.Int) (*types.Transaction, error) {
	return _IdentityContract.Contract.AddKey(&_IdentityContract.TransactOpts, _key, _purpose, _keyType)
}

// AddKey is a paid mutator transaction binding the contract method 0x1d381240.
//
// Solidity: function addKey(_key bytes32, _purpose uint256, _keyType uint256) returns()
func (_IdentityContract *IdentityContractTransactorSession) AddKey(_key [32]byte, _purpose *big.Int, _keyType *big.Int) (*types.Transaction, error) {
	return _IdentityContract.Contract.AddKey(&_IdentityContract.TransactOpts, _key, _purpose, _keyType)
}

// AddMultiPurposeKey is a paid mutator transaction binding the contract method 0x173d2616.
//
// Solidity: function addMultiPurposeKey(_key bytes32, _purposes uint256[], _keyType uint256) returns()
func (_IdentityContract *IdentityContractTransactor) AddMultiPurposeKey(opts *bind.TransactOpts, _key [32]byte, _purposes []*big.Int, _keyType *big.Int) (*types.Transaction, error) {
	return _IdentityContract.contract.Transact(opts, "addMultiPurposeKey", _key, _purposes, _keyType)
}

// AddMultiPurposeKey is a paid mutator transaction binding the contract method 0x173d2616.
//
// Solidity: function addMultiPurposeKey(_key bytes32, _purposes uint256[], _keyType uint256) returns()
func (_IdentityContract *IdentityContractSession) AddMultiPurposeKey(_key [32]byte, _purposes []*big.Int, _keyType *big.Int) (*types.Transaction, error) {
	return _IdentityContract.Contract.AddMultiPurposeKey(&_IdentityContract.TransactOpts, _key, _purposes, _keyType)
}

// AddMultiPurposeKey is a paid mutator transaction binding the contract method 0x173d2616.
//
// Solidity: function addMultiPurposeKey(_key bytes32, _purposes uint256[], _keyType uint256) returns()
func (_IdentityContract *IdentityContractTransactorSession) AddMultiPurposeKey(_key [32]byte, _purposes []*big.Int, _keyType *big.Int) (*types.Transaction, error) {
	return _IdentityContract.Contract.AddMultiPurposeKey(&_IdentityContract.TransactOpts, _key, _purposes, _keyType)
}

// Execute is a paid mutator transaction binding the contract method 0xb61d27f6.
//
// Solidity: function execute(_to address, _value uint256, _data bytes) returns(success bool)
func (_IdentityContract *IdentityContractTransactor) Execute(opts *bind.TransactOpts, _to common.Address, _value *big.Int, _data []byte) (*types.Transaction, error) {
	return _IdentityContract.contract.Transact(opts, "execute", _to, _value, _data)
}

// Execute is a paid mutator transaction binding the contract method 0xb61d27f6.
//
// Solidity: function execute(_to address, _value uint256, _data bytes) returns(success bool)
func (_IdentityContract *IdentityContractSession) Execute(_to common.Address, _value *big.Int, _data []byte) (*types.Transaction, error) {
	return _IdentityContract.Contract.Execute(&_IdentityContract.TransactOpts, _to, _value, _data)
}

// Execute is a paid mutator transaction binding the contract method 0xb61d27f6.
//
// Solidity: function execute(_to address, _value uint256, _data bytes) returns(success bool)
func (_IdentityContract *IdentityContractTransactorSession) Execute(_to common.Address, _value *big.Int, _data []byte) (*types.Transaction, error) {
	return _IdentityContract.Contract.Execute(&_IdentityContract.TransactOpts, _to, _value, _data)
}

// RevokeKey is a paid mutator transaction binding the contract method 0x572f2210.
//
// Solidity: function revokeKey(_key bytes32) returns()
func (_IdentityContract *IdentityContractTransactor) RevokeKey(opts *bind.TransactOpts, _key [32]byte) (*types.Transaction, error) {
	return _IdentityContract.contract.Transact(opts, "revokeKey", _key)
}

// RevokeKey is a paid mutator transaction binding the contract method 0x572f2210.
//
// Solidity: function revokeKey(_key bytes32) returns()
func (_IdentityContract *IdentityContractSession) RevokeKey(_key [32]byte) (*types.Transaction, error) {
	return _IdentityContract.Contract.RevokeKey(&_IdentityContract.TransactOpts, _key)
}

// RevokeKey is a paid mutator transaction binding the contract method 0x572f2210.
//
// Solidity: function revokeKey(_key bytes32) returns()
func (_IdentityContract *IdentityContractTransactorSession) RevokeKey(_key [32]byte) (*types.Transaction, error) {
	return _IdentityContract.Contract.RevokeKey(&_IdentityContract.TransactOpts, _key)
}

// IdentityContractKeyAddedIterator is returned from FilterKeyAdded and is used to iterate over the raw logs and unpacked data for KeyAdded events raised by the IdentityContract contract.
type IdentityContractKeyAddedIterator struct {
	Event *IdentityContractKeyAdded // Event containing the contract specifics and raw log

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
func (it *IdentityContractKeyAddedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(IdentityContractKeyAdded)
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
		it.Event = new(IdentityContractKeyAdded)
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
func (it *IdentityContractKeyAddedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *IdentityContractKeyAddedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// IdentityContractKeyAdded represents a KeyAdded event raised by the IdentityContract contract.
type IdentityContractKeyAdded struct {
	Key     [32]byte
	Purpose *big.Int
	KeyType *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterKeyAdded is a free log retrieval operation binding the contract event 0x480000bb1edad8ca1470381cc334b1917fbd51c6531f3a623ea8e0ec7e38a6e9.
//
// Solidity: e KeyAdded(key indexed bytes32, purpose indexed uint256, keyType indexed uint256)
func (_IdentityContract *IdentityContractFilterer) FilterKeyAdded(opts *bind.FilterOpts, key [][32]byte, purpose []*big.Int, keyType []*big.Int) (*IdentityContractKeyAddedIterator, error) {

	var keyRule []interface{}
	for _, keyItem := range key {
		keyRule = append(keyRule, keyItem)
	}
	var purposeRule []interface{}
	for _, purposeItem := range purpose {
		purposeRule = append(purposeRule, purposeItem)
	}
	var keyTypeRule []interface{}
	for _, keyTypeItem := range keyType {
		keyTypeRule = append(keyTypeRule, keyTypeItem)
	}

	logs, sub, err := _IdentityContract.contract.FilterLogs(opts, "KeyAdded", keyRule, purposeRule, keyTypeRule)
	if err != nil {
		return nil, err
	}
	return &IdentityContractKeyAddedIterator{contract: _IdentityContract.contract, event: "KeyAdded", logs: logs, sub: sub}, nil
}

// WatchKeyAdded is a free log subscription operation binding the contract event 0x480000bb1edad8ca1470381cc334b1917fbd51c6531f3a623ea8e0ec7e38a6e9.
//
// Solidity: e KeyAdded(key indexed bytes32, purpose indexed uint256, keyType indexed uint256)
func (_IdentityContract *IdentityContractFilterer) WatchKeyAdded(opts *bind.WatchOpts, sink chan<- *IdentityContractKeyAdded, key [][32]byte, purpose []*big.Int, keyType []*big.Int) (event.Subscription, error) {

	var keyRule []interface{}
	for _, keyItem := range key {
		keyRule = append(keyRule, keyItem)
	}
	var purposeRule []interface{}
	for _, purposeItem := range purpose {
		purposeRule = append(purposeRule, purposeItem)
	}
	var keyTypeRule []interface{}
	for _, keyTypeItem := range keyType {
		keyTypeRule = append(keyTypeRule, keyTypeItem)
	}

	logs, sub, err := _IdentityContract.contract.WatchLogs(opts, "KeyAdded", keyRule, purposeRule, keyTypeRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(IdentityContractKeyAdded)
				if err := _IdentityContract.contract.UnpackLog(event, "KeyAdded", log); err != nil {
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

// IdentityContractKeyRevokedIterator is returned from FilterKeyRevoked and is used to iterate over the raw logs and unpacked data for KeyRevoked events raised by the IdentityContract contract.
type IdentityContractKeyRevokedIterator struct {
	Event *IdentityContractKeyRevoked // Event containing the contract specifics and raw log

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
func (it *IdentityContractKeyRevokedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(IdentityContractKeyRevoked)
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
		it.Event = new(IdentityContractKeyRevoked)
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
func (it *IdentityContractKeyRevokedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *IdentityContractKeyRevokedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// IdentityContractKeyRevoked represents a KeyRevoked event raised by the IdentityContract contract.
type IdentityContractKeyRevoked struct {
	Key       [32]byte
	RevokedAt *big.Int
	KeyType   *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterKeyRevoked is a free log retrieval operation binding the contract event 0x8004a857c5cbc7c7c693a7c6a2852c373b5d03f882b57a8ee22dd2d4492331b1.
//
// Solidity: e KeyRevoked(key indexed bytes32, revokedAt indexed uint256, keyType indexed uint256)
func (_IdentityContract *IdentityContractFilterer) FilterKeyRevoked(opts *bind.FilterOpts, key [][32]byte, revokedAt []*big.Int, keyType []*big.Int) (*IdentityContractKeyRevokedIterator, error) {

	var keyRule []interface{}
	for _, keyItem := range key {
		keyRule = append(keyRule, keyItem)
	}
	var revokedAtRule []interface{}
	for _, revokedAtItem := range revokedAt {
		revokedAtRule = append(revokedAtRule, revokedAtItem)
	}
	var keyTypeRule []interface{}
	for _, keyTypeItem := range keyType {
		keyTypeRule = append(keyTypeRule, keyTypeItem)
	}

	logs, sub, err := _IdentityContract.contract.FilterLogs(opts, "KeyRevoked", keyRule, revokedAtRule, keyTypeRule)
	if err != nil {
		return nil, err
	}
	return &IdentityContractKeyRevokedIterator{contract: _IdentityContract.contract, event: "KeyRevoked", logs: logs, sub: sub}, nil
}

// WatchKeyRevoked is a free log subscription operation binding the contract event 0x8004a857c5cbc7c7c693a7c6a2852c373b5d03f882b57a8ee22dd2d4492331b1.
//
// Solidity: e KeyRevoked(key indexed bytes32, revokedAt indexed uint256, keyType indexed uint256)
func (_IdentityContract *IdentityContractFilterer) WatchKeyRevoked(opts *bind.WatchOpts, sink chan<- *IdentityContractKeyRevoked, key [][32]byte, revokedAt []*big.Int, keyType []*big.Int) (event.Subscription, error) {

	var keyRule []interface{}
	for _, keyItem := range key {
		keyRule = append(keyRule, keyItem)
	}
	var revokedAtRule []interface{}
	for _, revokedAtItem := range revokedAt {
		revokedAtRule = append(revokedAtRule, revokedAtItem)
	}
	var keyTypeRule []interface{}
	for _, keyTypeItem := range keyType {
		keyTypeRule = append(keyTypeRule, keyTypeItem)
	}

	logs, sub, err := _IdentityContract.contract.WatchLogs(opts, "KeyRevoked", keyRule, revokedAtRule, keyTypeRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(IdentityContractKeyRevoked)
				if err := _IdentityContract.contract.UnpackLog(event, "KeyRevoked", log); err != nil {
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
