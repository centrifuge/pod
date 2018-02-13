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

// EthereumAnchorRegistryContractABI is the input ABI used to generate the binding from.
const EthereumAnchorRegistryContractABI = "[{\"constant\":true,\"inputs\":[{\"name\":\"identifier\",\"type\":\"bytes32\"}],\"name\":\"getAnchorById\",\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\"},{\"name\":\"\",\"type\":\"bytes32\"},{\"name\":\"\",\"type\":\"bytes32\"},{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"identifier\",\"type\":\"bytes32\"},{\"name\":\"merkleRoot\",\"type\":\"bytes32\"},{\"name\":\"anchorSchemaVersion\",\"type\":\"uint256\"}],\"name\":\"registerAnchor\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"anchors\",\"outputs\":[{\"name\":\"identifier\",\"type\":\"bytes32\"},{\"name\":\"merkleRoot\",\"type\":\"bytes32\"},{\"name\":\"timestamp\",\"type\":\"bytes32\"},{\"name\":\"schemaVersion\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"identifier\",\"type\":\"bytes32\"},{\"indexed\":false,\"name\":\"rootHash\",\"type\":\"bytes32\"},{\"indexed\":false,\"name\":\"timestamp\",\"type\":\"bytes32\"},{\"indexed\":false,\"name\":\"anchorSchemaVersion\",\"type\":\"uint256\"}],\"name\":\"AnchorRegistered\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"}]"

// EthereumAnchorRegistryContract is an auto generated Go binding around an Ethereum contract.
type EthereumAnchorRegistryContract struct {
	EthereumAnchorRegistryContractCaller     // Read-only binding to the contract
	EthereumAnchorRegistryContractTransactor // Write-only binding to the contract
}

// EthereumAnchorRegistryContractCaller is an auto generated read-only Go binding around an Ethereum contract.
type EthereumAnchorRegistryContractCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// EthereumAnchorRegistryContractTransactor is an auto generated write-only Go binding around an Ethereum contract.
type EthereumAnchorRegistryContractTransactor struct {
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
	contract, err := bindEthereumAnchorRegistryContract(address, backend, backend)
	if err != nil {
		return nil, err
	}
	return &EthereumAnchorRegistryContract{EthereumAnchorRegistryContractCaller: EthereumAnchorRegistryContractCaller{contract: contract}, EthereumAnchorRegistryContractTransactor: EthereumAnchorRegistryContractTransactor{contract: contract}}, nil
}

// NewEthereumAnchorRegistryContractCaller creates a new read-only instance of EthereumAnchorRegistryContract, bound to a specific deployed contract.
func NewEthereumAnchorRegistryContractCaller(address common.Address, caller bind.ContractCaller) (*EthereumAnchorRegistryContractCaller, error) {
	contract, err := bindEthereumAnchorRegistryContract(address, caller, nil)
	if err != nil {
		return nil, err
	}
	return &EthereumAnchorRegistryContractCaller{contract: contract}, nil
}

// NewEthereumAnchorRegistryContractTransactor creates a new write-only instance of EthereumAnchorRegistryContract, bound to a specific deployed contract.
func NewEthereumAnchorRegistryContractTransactor(address common.Address, transactor bind.ContractTransactor) (*EthereumAnchorRegistryContractTransactor, error) {
	contract, err := bindEthereumAnchorRegistryContract(address, nil, transactor)
	if err != nil {
		return nil, err
	}
	return &EthereumAnchorRegistryContractTransactor{contract: contract}, nil
}

// bindEthereumAnchorRegistryContract binds a generic wrapper to an already deployed contract.
func bindEthereumAnchorRegistryContract(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(EthereumAnchorRegistryContractABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor), nil
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
