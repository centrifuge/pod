// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package ethid

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
const EthereumIdentityContractABI = "[{\"constant\":true,\"inputs\":[{\"name\":\"_key\",\"type\":\"bytes32\"}],\"name\":\"getKey\",\"outputs\":[{\"name\":\"key\",\"type\":\"bytes32\"},{\"name\":\"purposes\",\"type\":\"uint256[]\"},{\"name\":\"revokedAt\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_key\",\"type\":\"bytes32\"},{\"name\":\"_purpose\",\"type\":\"uint256\"}],\"name\":\"addKey\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"centrifugeId\",\"outputs\":[{\"name\":\"\",\"type\":\"uint48\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_key\",\"type\":\"bytes32\"}],\"name\":\"revokeKey\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_purpose\",\"type\":\"uint256\"}],\"name\":\"getKeysByPurpose\",\"outputs\":[{\"name\":\"\",\"type\":\"bytes32[]\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_key\",\"type\":\"bytes32\"},{\"name\":\"_purposes\",\"type\":\"uint256[]\"}],\"name\":\"addMultiPurposeKey\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_key\",\"type\":\"bytes32\"},{\"name\":\"_purpose\",\"type\":\"uint256\"}],\"name\":\"keyHasPurpose\",\"outputs\":[{\"name\":\"found\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"name\":\"_centrifugeId\",\"type\":\"uint48\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"key\",\"type\":\"bytes32\"},{\"indexed\":true,\"name\":\"purpose\",\"type\":\"uint256\"}],\"name\":\"KeyAdded\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"key\",\"type\":\"bytes32\"},{\"indexed\":true,\"name\":\"revokedAt\",\"type\":\"uint256\"}],\"name\":\"KeyRevoked\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"previousOwner\",\"type\":\"address\"}],\"name\":\"OwnershipRenounced\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"constant\":true,\"inputs\":[{\"name\":\"_toSign\",\"type\":\"bytes32\"},{\"name\":\"_signature\",\"type\":\"bytes\"}],\"name\":\"isSignatureValid\",\"outputs\":[{\"name\":\"valid\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"}]"

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

// CentrifugeID is a free data retrieval call binding the contract method 0x41a43c38.
//
// Solidity: function centrifugeId() constant returns(uint48)
func (_EthereumIdentityContract *EthereumIdentityContractCaller) CentrifugeId(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _EthereumIdentityContract.contract.Call(opts, out, "centrifugeId")
	return *ret0, err
}

// CentrifugeID is a free data retrieval call binding the contract method 0x41a43c38.
//
// Solidity: function centrifugeId() constant returns(uint48)
func (_EthereumIdentityContract *EthereumIdentityContractSession) CentrifugeId() (*big.Int, error) {
	return _EthereumIdentityContract.Contract.CentrifugeId(&_EthereumIdentityContract.CallOpts)
}

// CentrifugeID is a free data retrieval call binding the contract method 0x41a43c38.
//
// Solidity: function centrifugeId() constant returns(uint48)
func (_EthereumIdentityContract *EthereumIdentityContractCallerSession) CentrifugeId() (*big.Int, error) {
	return _EthereumIdentityContract.Contract.CentrifugeId(&_EthereumIdentityContract.CallOpts)
}

// GetKey is a free data retrieval call binding the contract method 0x12aaac70.
//
// Solidity: function getKey(_key bytes32) constant returns(key bytes32, purposes uint256[], revokedAt uint256)
func (_EthereumIdentityContract *EthereumIdentityContractCaller) GetKey(opts *bind.CallOpts, _key [32]byte) (struct {
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
	err := _EthereumIdentityContract.contract.Call(opts, out, "getKey", _key)
	return *ret, err
}

// GetKey is a free data retrieval call binding the contract method 0x12aaac70.
//
// Solidity: function getKey(_key bytes32) constant returns(key bytes32, purposes uint256[], revokedAt uint256)
func (_EthereumIdentityContract *EthereumIdentityContractSession) GetKey(_key [32]byte) (struct {
	Key       [32]byte
	Purposes  []*big.Int
	RevokedAt *big.Int
}, error) {
	return _EthereumIdentityContract.Contract.GetKey(&_EthereumIdentityContract.CallOpts, _key)
}

// GetKey is a free data retrieval call binding the contract method 0x12aaac70.
//
// Solidity: function getKey(_key bytes32) constant returns(key bytes32, purposes uint256[], revokedAt uint256)
func (_EthereumIdentityContract *EthereumIdentityContractCallerSession) GetKey(_key [32]byte) (struct {
	Key       [32]byte
	Purposes  []*big.Int
	RevokedAt *big.Int
}, error) {
	return _EthereumIdentityContract.Contract.GetKey(&_EthereumIdentityContract.CallOpts, _key)
}

// GetKeysByPurpose is a free data retrieval call binding the contract method 0x9010f726.
//
// Solidity: function getKeysByPurpose(_purpose uint256) constant returns(bytes32[])
func (_EthereumIdentityContract *EthereumIdentityContractCaller) GetKeysByPurpose(opts *bind.CallOpts, _purpose *big.Int) ([][32]byte, error) {
	var (
		ret0 = new([][32]byte)
	)
	out := ret0
	err := _EthereumIdentityContract.contract.Call(opts, out, "getKeysByPurpose", _purpose)
	return *ret0, err
}

// GetKeysByPurpose is a free data retrieval call binding the contract method 0x9010f726.
//
// Solidity: function getKeysByPurpose(_purpose uint256) constant returns(bytes32[])
func (_EthereumIdentityContract *EthereumIdentityContractSession) GetKeysByPurpose(_purpose *big.Int) ([][32]byte, error) {
	return _EthereumIdentityContract.Contract.GetKeysByPurpose(&_EthereumIdentityContract.CallOpts, _purpose)
}

// GetKeysByPurpose is a free data retrieval call binding the contract method 0x9010f726.
//
// Solidity: function getKeysByPurpose(_purpose uint256) constant returns(bytes32[])
func (_EthereumIdentityContract *EthereumIdentityContractCallerSession) GetKeysByPurpose(_purpose *big.Int) ([][32]byte, error) {
	return _EthereumIdentityContract.Contract.GetKeysByPurpose(&_EthereumIdentityContract.CallOpts, _purpose)
}

// IsSignatureValid is a free data retrieval call binding the contract method 0x0892beb7.
//
// Solidity: function isSignatureValid(_toSign bytes32, _signature bytes) constant returns(valid bool)
func (_EthereumIdentityContract *EthereumIdentityContractCaller) IsSignatureValid(opts *bind.CallOpts, _toSign [32]byte, _signature []byte) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _EthereumIdentityContract.contract.Call(opts, out, "isSignatureValid", _toSign, _signature)
	return *ret0, err
}

// IsSignatureValid is a free data retrieval call binding the contract method 0x0892beb7.
//
// Solidity: function isSignatureValid(_toSign bytes32, _signature bytes) constant returns(valid bool)
func (_EthereumIdentityContract *EthereumIdentityContractSession) IsSignatureValid(_toSign [32]byte, _signature []byte) (bool, error) {
	return _EthereumIdentityContract.Contract.IsSignatureValid(&_EthereumIdentityContract.CallOpts, _toSign, _signature)
}

// IsSignatureValid is a free data retrieval call binding the contract method 0x0892beb7.
//
// Solidity: function isSignatureValid(_toSign bytes32, _signature bytes) constant returns(valid bool)
func (_EthereumIdentityContract *EthereumIdentityContractCallerSession) IsSignatureValid(_toSign [32]byte, _signature []byte) (bool, error) {
	return _EthereumIdentityContract.Contract.IsSignatureValid(&_EthereumIdentityContract.CallOpts, _toSign, _signature)
}

// KeyHasPurpose is a free data retrieval call binding the contract method 0xd202158d.
//
// Solidity: function keyHasPurpose(_key bytes32, _purpose uint256) constant returns(found bool)
func (_EthereumIdentityContract *EthereumIdentityContractCaller) KeyHasPurpose(opts *bind.CallOpts, _key [32]byte, _purpose *big.Int) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _EthereumIdentityContract.contract.Call(opts, out, "keyHasPurpose", _key, _purpose)
	return *ret0, err
}

// KeyHasPurpose is a free data retrieval call binding the contract method 0xd202158d.
//
// Solidity: function keyHasPurpose(_key bytes32, _purpose uint256) constant returns(found bool)
func (_EthereumIdentityContract *EthereumIdentityContractSession) KeyHasPurpose(_key [32]byte, _purpose *big.Int) (bool, error) {
	return _EthereumIdentityContract.Contract.KeyHasPurpose(&_EthereumIdentityContract.CallOpts, _key, _purpose)
}

// KeyHasPurpose is a free data retrieval call binding the contract method 0xd202158d.
//
// Solidity: function keyHasPurpose(_key bytes32, _purpose uint256) constant returns(found bool)
func (_EthereumIdentityContract *EthereumIdentityContractCallerSession) KeyHasPurpose(_key [32]byte, _purpose *big.Int) (bool, error) {
	return _EthereumIdentityContract.Contract.KeyHasPurpose(&_EthereumIdentityContract.CallOpts, _key, _purpose)
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
// Solidity: function addKey(_key bytes32, _purpose uint256) returns()
func (_EthereumIdentityContract *EthereumIdentityContractTransactor) AddKey(opts *bind.TransactOpts, _key [32]byte, _purpose *big.Int) (*types.Transaction, error) {
	return _EthereumIdentityContract.contract.Transact(opts, "addKey", _key, _purpose)
}

// AddKey is a paid mutator transaction binding the contract method 0x4103ef4c.
//
// Solidity: function addKey(_key bytes32, _purpose uint256) returns()
func (_EthereumIdentityContract *EthereumIdentityContractSession) AddKey(_key [32]byte, _purpose *big.Int) (*types.Transaction, error) {
	return _EthereumIdentityContract.Contract.AddKey(&_EthereumIdentityContract.TransactOpts, _key, _purpose)
}

// AddKey is a paid mutator transaction binding the contract method 0x4103ef4c.
//
// Solidity: function addKey(_key bytes32, _purpose uint256) returns()
func (_EthereumIdentityContract *EthereumIdentityContractTransactorSession) AddKey(_key [32]byte, _purpose *big.Int) (*types.Transaction, error) {
	return _EthereumIdentityContract.Contract.AddKey(&_EthereumIdentityContract.TransactOpts, _key, _purpose)
}

// AddMultiPurposeKey is a paid mutator transaction binding the contract method 0xcb757f7f.
//
// Solidity: function addMultiPurposeKey(_key bytes32, _purposes uint256[]) returns()
func (_EthereumIdentityContract *EthereumIdentityContractTransactor) AddMultiPurposeKey(opts *bind.TransactOpts, _key [32]byte, _purposes []*big.Int) (*types.Transaction, error) {
	return _EthereumIdentityContract.contract.Transact(opts, "addMultiPurposeKey", _key, _purposes)
}

// AddMultiPurposeKey is a paid mutator transaction binding the contract method 0xcb757f7f.
//
// Solidity: function addMultiPurposeKey(_key bytes32, _purposes uint256[]) returns()
func (_EthereumIdentityContract *EthereumIdentityContractSession) AddMultiPurposeKey(_key [32]byte, _purposes []*big.Int) (*types.Transaction, error) {
	return _EthereumIdentityContract.Contract.AddMultiPurposeKey(&_EthereumIdentityContract.TransactOpts, _key, _purposes)
}

// AddMultiPurposeKey is a paid mutator transaction binding the contract method 0xcb757f7f.
//
// Solidity: function addMultiPurposeKey(_key bytes32, _purposes uint256[]) returns()
func (_EthereumIdentityContract *EthereumIdentityContractTransactorSession) AddMultiPurposeKey(_key [32]byte, _purposes []*big.Int) (*types.Transaction, error) {
	return _EthereumIdentityContract.Contract.AddMultiPurposeKey(&_EthereumIdentityContract.TransactOpts, _key, _purposes)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_EthereumIdentityContract *EthereumIdentityContractTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _EthereumIdentityContract.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_EthereumIdentityContract *EthereumIdentityContractSession) RenounceOwnership() (*types.Transaction, error) {
	return _EthereumIdentityContract.Contract.RenounceOwnership(&_EthereumIdentityContract.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_EthereumIdentityContract *EthereumIdentityContractTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _EthereumIdentityContract.Contract.RenounceOwnership(&_EthereumIdentityContract.TransactOpts)
}

// RevokeKey is a paid mutator transaction binding the contract method 0x572f2210.
//
// Solidity: function revokeKey(_key bytes32) returns()
func (_EthereumIdentityContract *EthereumIdentityContractTransactor) RevokeKey(opts *bind.TransactOpts, _key [32]byte) (*types.Transaction, error) {
	return _EthereumIdentityContract.contract.Transact(opts, "revokeKey", _key)
}

// RevokeKey is a paid mutator transaction binding the contract method 0x572f2210.
//
// Solidity: function revokeKey(_key bytes32) returns()
func (_EthereumIdentityContract *EthereumIdentityContractSession) RevokeKey(_key [32]byte) (*types.Transaction, error) {
	return _EthereumIdentityContract.Contract.RevokeKey(&_EthereumIdentityContract.TransactOpts, _key)
}

// RevokeKey is a paid mutator transaction binding the contract method 0x572f2210.
//
// Solidity: function revokeKey(_key bytes32) returns()
func (_EthereumIdentityContract *EthereumIdentityContractTransactorSession) RevokeKey(_key [32]byte) (*types.Transaction, error) {
	return _EthereumIdentityContract.Contract.RevokeKey(&_EthereumIdentityContract.TransactOpts, _key)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(_newOwner address) returns()
func (_EthereumIdentityContract *EthereumIdentityContractTransactor) TransferOwnership(opts *bind.TransactOpts, _newOwner common.Address) (*types.Transaction, error) {
	return _EthereumIdentityContract.contract.Transact(opts, "transferOwnership", _newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(_newOwner address) returns()
func (_EthereumIdentityContract *EthereumIdentityContractSession) TransferOwnership(_newOwner common.Address) (*types.Transaction, error) {
	return _EthereumIdentityContract.Contract.TransferOwnership(&_EthereumIdentityContract.TransactOpts, _newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(_newOwner address) returns()
func (_EthereumIdentityContract *EthereumIdentityContractTransactorSession) TransferOwnership(_newOwner common.Address) (*types.Transaction, error) {
	return _EthereumIdentityContract.Contract.TransferOwnership(&_EthereumIdentityContract.TransactOpts, _newOwner)
}

// EthereumIdentityContractKeyAddedIterator is returned from FilterKeyAdded and is used to iterate over the raw logs and unpacked data for KeyAdded events raised by the EthereumIdentityContract contract.
type EthereumIdentityContractKeyAddedIterator struct {
	Event *EthereumIdentityContractKeyAdded // Event containing the contract specifics and raw log

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
func (it *EthereumIdentityContractKeyAddedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(EthereumIdentityContractKeyAdded)
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
		it.Event = new(EthereumIdentityContractKeyAdded)
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
func (it *EthereumIdentityContractKeyAddedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *EthereumIdentityContractKeyAddedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// EthereumIdentityContractKeyAdded represents a KeyAdded event raised by the EthereumIdentityContract contract.
type EthereumIdentityContractKeyAdded struct {
	Key     [32]byte
	Purpose *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterKeyAdded is a free log retrieval operation binding the contract event 0x90cf26894583787fe4a16185f0128c58652d10939f4b0185a951efc8452bcaa8.
//
// Solidity: e KeyAdded(key indexed bytes32, purpose indexed uint256)
func (_EthereumIdentityContract *EthereumIdentityContractFilterer) FilterKeyAdded(opts *bind.FilterOpts, key [][32]byte, purpose []*big.Int) (*EthereumIdentityContractKeyAddedIterator, error) {

	var keyRule []interface{}
	for _, keyItem := range key {
		keyRule = append(keyRule, keyItem)
	}
	var purposeRule []interface{}
	for _, purposeItem := range purpose {
		purposeRule = append(purposeRule, purposeItem)
	}

	logs, sub, err := _EthereumIdentityContract.contract.FilterLogs(opts, "KeyAdded", keyRule, purposeRule)
	if err != nil {
		return nil, err
	}
	return &EthereumIdentityContractKeyAddedIterator{contract: _EthereumIdentityContract.contract, event: "KeyAdded", logs: logs, sub: sub}, nil
}

// WatchKeyAdded is a free log subscription operation binding the contract event 0x90cf26894583787fe4a16185f0128c58652d10939f4b0185a951efc8452bcaa8.
//
// Solidity: e KeyAdded(key indexed bytes32, purpose indexed uint256)
func (_EthereumIdentityContract *EthereumIdentityContractFilterer) WatchKeyAdded(opts *bind.WatchOpts, sink chan<- *EthereumIdentityContractKeyAdded, key [][32]byte, purpose []*big.Int) (event.Subscription, error) {

	var keyRule []interface{}
	for _, keyItem := range key {
		keyRule = append(keyRule, keyItem)
	}
	var purposeRule []interface{}
	for _, purposeItem := range purpose {
		purposeRule = append(purposeRule, purposeItem)
	}

	logs, sub, err := _EthereumIdentityContract.contract.WatchLogs(opts, "KeyAdded", keyRule, purposeRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(EthereumIdentityContractKeyAdded)
				if err := _EthereumIdentityContract.contract.UnpackLog(event, "KeyAdded", log); err != nil {
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

// EthereumIdentityContractKeyRevokedIterator is returned from FilterKeyRevoked and is used to iterate over the raw logs and unpacked data for KeyRevoked events raised by the EthereumIdentityContract contract.
type EthereumIdentityContractKeyRevokedIterator struct {
	Event *EthereumIdentityContractKeyRevoked // Event containing the contract specifics and raw log

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
func (it *EthereumIdentityContractKeyRevokedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(EthereumIdentityContractKeyRevoked)
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
		it.Event = new(EthereumIdentityContractKeyRevoked)
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
func (it *EthereumIdentityContractKeyRevokedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *EthereumIdentityContractKeyRevokedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// EthereumIdentityContractKeyRevoked represents a KeyRevoked event raised by the EthereumIdentityContract contract.
type EthereumIdentityContractKeyRevoked struct {
	Key       [32]byte
	RevokedAt *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterKeyRevoked is a free log retrieval operation binding the contract event 0x3aa2212f5d3ddf6cf452b1611d7ea62ac572afe7d5e4310c09185edd915c686e.
//
// Solidity: e KeyRevoked(key indexed bytes32, revokedAt indexed uint256)
func (_EthereumIdentityContract *EthereumIdentityContractFilterer) FilterKeyRevoked(opts *bind.FilterOpts, key [][32]byte, revokedAt []*big.Int) (*EthereumIdentityContractKeyRevokedIterator, error) {

	var keyRule []interface{}
	for _, keyItem := range key {
		keyRule = append(keyRule, keyItem)
	}
	var revokedAtRule []interface{}
	for _, revokedAtItem := range revokedAt {
		revokedAtRule = append(revokedAtRule, revokedAtItem)
	}

	logs, sub, err := _EthereumIdentityContract.contract.FilterLogs(opts, "KeyRevoked", keyRule, revokedAtRule)
	if err != nil {
		return nil, err
	}
	return &EthereumIdentityContractKeyRevokedIterator{contract: _EthereumIdentityContract.contract, event: "KeyRevoked", logs: logs, sub: sub}, nil
}

// WatchKeyRevoked is a free log subscription operation binding the contract event 0x3aa2212f5d3ddf6cf452b1611d7ea62ac572afe7d5e4310c09185edd915c686e.
//
// Solidity: e KeyRevoked(key indexed bytes32, revokedAt indexed uint256)
func (_EthereumIdentityContract *EthereumIdentityContractFilterer) WatchKeyRevoked(opts *bind.WatchOpts, sink chan<- *EthereumIdentityContractKeyRevoked, key [][32]byte, revokedAt []*big.Int) (event.Subscription, error) {

	var keyRule []interface{}
	for _, keyItem := range key {
		keyRule = append(keyRule, keyItem)
	}
	var revokedAtRule []interface{}
	for _, revokedAtItem := range revokedAt {
		revokedAtRule = append(revokedAtRule, revokedAtItem)
	}

	logs, sub, err := _EthereumIdentityContract.contract.WatchLogs(opts, "KeyRevoked", keyRule, revokedAtRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(EthereumIdentityContractKeyRevoked)
				if err := _EthereumIdentityContract.contract.UnpackLog(event, "KeyRevoked", log); err != nil {
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

// EthereumIdentityContractOwnershipRenouncedIterator is returned from FilterOwnershipRenounced and is used to iterate over the raw logs and unpacked data for OwnershipRenounced events raised by the EthereumIdentityContract contract.
type EthereumIdentityContractOwnershipRenouncedIterator struct {
	Event *EthereumIdentityContractOwnershipRenounced // Event containing the contract specifics and raw log

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
func (it *EthereumIdentityContractOwnershipRenouncedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(EthereumIdentityContractOwnershipRenounced)
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
		it.Event = new(EthereumIdentityContractOwnershipRenounced)
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
func (it *EthereumIdentityContractOwnershipRenouncedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *EthereumIdentityContractOwnershipRenouncedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// EthereumIdentityContractOwnershipRenounced represents a OwnershipRenounced event raised by the EthereumIdentityContract contract.
type EthereumIdentityContractOwnershipRenounced struct {
	PreviousOwner common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipRenounced is a free log retrieval operation binding the contract event 0xf8df31144d9c2f0f6b59d69b8b98abd5459d07f2742c4df920b25aae33c64820.
//
// Solidity: e OwnershipRenounced(previousOwner indexed address)
func (_EthereumIdentityContract *EthereumIdentityContractFilterer) FilterOwnershipRenounced(opts *bind.FilterOpts, previousOwner []common.Address) (*EthereumIdentityContractOwnershipRenouncedIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}

	logs, sub, err := _EthereumIdentityContract.contract.FilterLogs(opts, "OwnershipRenounced", previousOwnerRule)
	if err != nil {
		return nil, err
	}
	return &EthereumIdentityContractOwnershipRenouncedIterator{contract: _EthereumIdentityContract.contract, event: "OwnershipRenounced", logs: logs, sub: sub}, nil
}

// WatchOwnershipRenounced is a free log subscription operation binding the contract event 0xf8df31144d9c2f0f6b59d69b8b98abd5459d07f2742c4df920b25aae33c64820.
//
// Solidity: e OwnershipRenounced(previousOwner indexed address)
func (_EthereumIdentityContract *EthereumIdentityContractFilterer) WatchOwnershipRenounced(opts *bind.WatchOpts, sink chan<- *EthereumIdentityContractOwnershipRenounced, previousOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}

	logs, sub, err := _EthereumIdentityContract.contract.WatchLogs(opts, "OwnershipRenounced", previousOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(EthereumIdentityContractOwnershipRenounced)
				if err := _EthereumIdentityContract.contract.UnpackLog(event, "OwnershipRenounced", log); err != nil {
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
// Solidity: e OwnershipTransferred(previousOwner indexed address, newOwner indexed address)
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
// Solidity: e OwnershipTransferred(previousOwner indexed address, newOwner indexed address)
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
