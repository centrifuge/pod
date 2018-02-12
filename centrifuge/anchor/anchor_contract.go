// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package anchor

import (
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// EthereumAnchorABI is the input ABI used to generate the binding from.
const EthereumAnchorABI = "[{\"constant\":true,\"inputs\":[{\"name\":\"identifier\",\"type\":\"bytes32\"}],\"name\":\"getAnchorById\",\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\"},{\"name\":\"\",\"type\":\"bytes32\"},{\"name\":\"\",\"type\":\"bytes32\"},{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"identifier\",\"type\":\"bytes32\"},{\"name\":\"merkleRoot\",\"type\":\"bytes32\"},{\"name\":\"anchorSchemaVersion\",\"type\":\"uint256\"}],\"name\":\"registerAnchor\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"anchors\",\"outputs\":[{\"name\":\"identifier\",\"type\":\"bytes32\"},{\"name\":\"merkleRoot\",\"type\":\"bytes32\"},{\"name\":\"timestamp\",\"type\":\"bytes32\"},{\"name\":\"schemaVersion\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"identifier\",\"type\":\"bytes32\"},{\"indexed\":false,\"name\":\"rootHash\",\"type\":\"bytes32\"},{\"indexed\":false,\"name\":\"timestamp\",\"type\":\"bytes32\"},{\"indexed\":false,\"name\":\"anchorSchemaVersion\",\"type\":\"uint256\"}],\"name\":\"AnchorRegistered\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"}]"

// EthereumAnchor is an auto generated Go binding around an Ethereum contract.
type EthereumAnchor struct {
	EthereumAnchorCaller     // Read-only binding to the contract
	EthereumAnchorTransactor // Write-only binding to the contract
}

// EthereumAnchorCaller is an auto generated read-only Go binding around an Ethereum contract.
type EthereumAnchorCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// EthereumAnchorTransactor is an auto generated write-only Go binding around an Ethereum contract.
type EthereumAnchorTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// EthereumAnchorSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type EthereumAnchorSession struct {
	Contract     *EthereumAnchor   // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// EthereumAnchorCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type EthereumAnchorCallerSession struct {
	Contract *EthereumAnchorCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts         // Call options to use throughout this session
}

// EthereumAnchorTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type EthereumAnchorTransactorSession struct {
	Contract     *EthereumAnchorTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts         // Transaction auth options to use throughout this session
}

// EthereumAnchorRaw is an auto generated low-level Go binding around an Ethereum contract.
type EthereumAnchorRaw struct {
	Contract *EthereumAnchor // Generic contract binding to access the raw methods on
}

// EthereumAnchorCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type EthereumAnchorCallerRaw struct {
	Contract *EthereumAnchorCaller // Generic read-only contract binding to access the raw methods on
}

// EthereumAnchorTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type EthereumAnchorTransactorRaw struct {
	Contract *EthereumAnchorTransactor // Generic write-only contract binding to access the raw methods on
}

// NewEthereumAnchor creates a new instance of EthereumAnchor, bound to a specific deployed contract.
func NewEthereumAnchor(address common.Address, backend bind.ContractBackend) (*EthereumAnchor, error) {
	contract, err := bindEthereumAnchor(address, backend, backend)
	if err != nil {
		return nil, err
	}
	return &EthereumAnchor{EthereumAnchorCaller: EthereumAnchorCaller{contract: contract}, EthereumAnchorTransactor: EthereumAnchorTransactor{contract: contract}}, nil
}

// NewEthereumAnchorCaller creates a new read-only instance of EthereumAnchor, bound to a specific deployed contract.
func NewEthereumAnchorCaller(address common.Address, caller bind.ContractCaller) (*EthereumAnchorCaller, error) {
	contract, err := bindEthereumAnchor(address, caller, nil)
	if err != nil {
		return nil, err
	}
	return &EthereumAnchorCaller{contract: contract}, nil
}

// NewEthereumAnchorTransactor creates a new write-only instance of EthereumAnchor, bound to a specific deployed contract.
func NewEthereumAnchorTransactor(address common.Address, transactor bind.ContractTransactor) (*EthereumAnchorTransactor, error) {
	contract, err := bindEthereumAnchor(address, nil, transactor)
	if err != nil {
		return nil, err
	}
	return &EthereumAnchorTransactor{contract: contract}, nil
}

// bindEthereumAnchor binds a generic wrapper to an already deployed contract.
func bindEthereumAnchor(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(EthereumAnchorABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_EthereumAnchor *EthereumAnchorRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _EthereumAnchor.Contract.EthereumAnchorCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_EthereumAnchor *EthereumAnchorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _EthereumAnchor.Contract.EthereumAnchorTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_EthereumAnchor *EthereumAnchorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _EthereumAnchor.Contract.EthereumAnchorTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_EthereumAnchor *EthereumAnchorCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _EthereumAnchor.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_EthereumAnchor *EthereumAnchorTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _EthereumAnchor.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_EthereumAnchor *EthereumAnchorTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _EthereumAnchor.Contract.contract.Transact(opts, method, params...)
}

// Anchors is a free data retrieval call binding the contract method 0xb01b6d53.
//
// Solidity: function anchors( bytes32) constant returns(identifier bytes32, merkleRoot bytes32, timestamp bytes32, schemaVersion uint256)
func (_EthereumAnchor *EthereumAnchorCaller) Anchors(opts *bind.CallOpts, arg0 [32]byte) (struct {
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
	err := _EthereumAnchor.contract.Call(opts, out, "anchors", arg0)
	return *ret, err
}

// Anchors is a free data retrieval call binding the contract method 0xb01b6d53.
//
// Solidity: function anchors( bytes32) constant returns(identifier bytes32, merkleRoot bytes32, timestamp bytes32, schemaVersion uint256)
func (_EthereumAnchor *EthereumAnchorSession) Anchors(arg0 [32]byte) (struct {
	Identifier    [32]byte
	MerkleRoot    [32]byte
	Timestamp     [32]byte
	SchemaVersion *big.Int
}, error) {
	return _EthereumAnchor.Contract.Anchors(&_EthereumAnchor.CallOpts, arg0)
}

// Anchors is a free data retrieval call binding the contract method 0xb01b6d53.
//
// Solidity: function anchors( bytes32) constant returns(identifier bytes32, merkleRoot bytes32, timestamp bytes32, schemaVersion uint256)
func (_EthereumAnchor *EthereumAnchorCallerSession) Anchors(arg0 [32]byte) (struct {
	Identifier    [32]byte
	MerkleRoot    [32]byte
	Timestamp     [32]byte
	SchemaVersion *big.Int
}, error) {
	return _EthereumAnchor.Contract.Anchors(&_EthereumAnchor.CallOpts, arg0)
}

// GetAnchorById is a free data retrieval call binding the contract method 0x04d466b2.
//
// Solidity: function getAnchorById(identifier bytes32) constant returns(bytes32, bytes32, bytes32, uint256)
func (_EthereumAnchor *EthereumAnchorCaller) GetAnchorById(opts *bind.CallOpts, identifier [32]byte) ([32]byte, [32]byte, [32]byte, *big.Int, error) {
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
	err := _EthereumAnchor.contract.Call(opts, out, "getAnchorById", identifier)
	return *ret0, *ret1, *ret2, *ret3, err
}

// GetAnchorById is a free data retrieval call binding the contract method 0x04d466b2.
//
// Solidity: function getAnchorById(identifier bytes32) constant returns(bytes32, bytes32, bytes32, uint256)
func (_EthereumAnchor *EthereumAnchorSession) GetAnchorById(identifier [32]byte) ([32]byte, [32]byte, [32]byte, *big.Int, error) {
	return _EthereumAnchor.Contract.GetAnchorById(&_EthereumAnchor.CallOpts, identifier)
}

// GetAnchorById is a free data retrieval call binding the contract method 0x04d466b2.
//
// Solidity: function getAnchorById(identifier bytes32) constant returns(bytes32, bytes32, bytes32, uint256)
func (_EthereumAnchor *EthereumAnchorCallerSession) GetAnchorById(identifier [32]byte) ([32]byte, [32]byte, [32]byte, *big.Int, error) {
	return _EthereumAnchor.Contract.GetAnchorById(&_EthereumAnchor.CallOpts, identifier)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() constant returns(address)
func (_EthereumAnchor *EthereumAnchorCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _EthereumAnchor.contract.Call(opts, out, "owner")
	return *ret0, err
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() constant returns(address)
func (_EthereumAnchor *EthereumAnchorSession) Owner() (common.Address, error) {
	return _EthereumAnchor.Contract.Owner(&_EthereumAnchor.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() constant returns(address)
func (_EthereumAnchor *EthereumAnchorCallerSession) Owner() (common.Address, error) {
	return _EthereumAnchor.Contract.Owner(&_EthereumAnchor.CallOpts)
}

// RegisterAnchor is a paid mutator transaction binding the contract method 0x77bfeb8b.
//
// Solidity: function registerAnchor(identifier bytes32, merkleRoot bytes32, anchorSchemaVersion uint256) returns()
func (_EthereumAnchor *EthereumAnchorTransactor) RegisterAnchor(opts *bind.TransactOpts, identifier [32]byte, merkleRoot [32]byte, anchorSchemaVersion *big.Int) (*types.Transaction, error) {
	return _EthereumAnchor.contract.Transact(opts, "registerAnchor", identifier, merkleRoot, anchorSchemaVersion)
}

// RegisterAnchor is a paid mutator transaction binding the contract method 0x77bfeb8b.
//
// Solidity: function registerAnchor(identifier bytes32, merkleRoot bytes32, anchorSchemaVersion uint256) returns()
func (_EthereumAnchor *EthereumAnchorSession) RegisterAnchor(identifier [32]byte, merkleRoot [32]byte, anchorSchemaVersion *big.Int) (*types.Transaction, error) {
	return _EthereumAnchor.Contract.RegisterAnchor(&_EthereumAnchor.TransactOpts, identifier, merkleRoot, anchorSchemaVersion)
}

// RegisterAnchor is a paid mutator transaction binding the contract method 0x77bfeb8b.
//
// Solidity: function registerAnchor(identifier bytes32, merkleRoot bytes32, anchorSchemaVersion uint256) returns()
func (_EthereumAnchor *EthereumAnchorTransactorSession) RegisterAnchor(identifier [32]byte, merkleRoot [32]byte, anchorSchemaVersion *big.Int) (*types.Transaction, error) {
	return _EthereumAnchor.Contract.RegisterAnchor(&_EthereumAnchor.TransactOpts, identifier, merkleRoot, anchorSchemaVersion)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(newOwner address) returns()
func (_EthereumAnchor *EthereumAnchorTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _EthereumAnchor.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(newOwner address) returns()
func (_EthereumAnchor *EthereumAnchorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _EthereumAnchor.Contract.TransferOwnership(&_EthereumAnchor.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(newOwner address) returns()
func (_EthereumAnchor *EthereumAnchorTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _EthereumAnchor.Contract.TransferOwnership(&_EthereumAnchor.TransactOpts, newOwner)
}
