// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package witness

import (
	"math/big"
	"strings"
)

// EthereumWitnessABI is the input ABI used to generate the binding from.
const EthereumWitnessABI = "[{\"constant\":true,\"inputs\":[{\"name\":\"version\",\"type\":\"bytes32\"}],\"name\":\"getWitness\",\"outputs\":[{\"name\":\"\",\"type\":\"bytes32[2]\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"bytes32\"},{\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"witness_list\",\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"version\",\"type\":\"bytes32\"},{\"name\":\"signature\",\"type\":\"bytes32\"}],\"name\":\"witnessDocument\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"}],"

// EthereumWitness is an auto generated Go binding around an Ethereum contract.
type EthereumWitness struct {
	EthereumWitnessCaller     // Read-only binding to the contract
	EthereumWitnessTransactor // Write-only binding to the contract
}

// EthereumWitnessCaller is an auto generated read-only Go binding around an Ethereum contract.
type EthereumWitnessCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// EthereumWitnessTransactor is an auto generated write-only Go binding around an Ethereum contract.
type EthereumWitnessTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// EthereumWitnessSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type EthereumWitnessSession struct {
	Contract     *EthereumWitness  // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// EthereumWitnessCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type EthereumWitnessCallerSession struct {
	Contract *EthereumWitnessCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts          // Call options to use throughout this session
}

// EthereumWitnessTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type EthereumWitnessTransactorSession struct {
	Contract     *EthereumWitnessTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts          // Transaction auth options to use throughout this session
}

// EthereumWitnessRaw is an auto generated low-level Go binding around an Ethereum contract.
type EthereumWitnessRaw struct {
	Contract *EthereumWitness // Generic contract binding to access the raw methods on
}

// EthereumWitnessCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type EthereumWitnessCallerRaw struct {
	Contract *EthereumWitnessCaller // Generic read-only contract binding to access the raw methods on
}

// EthereumWitnessTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type EthereumWitnessTransactorRaw struct {
	Contract *EthereumWitnessTransactor // Generic write-only contract binding to access the raw methods on
}

// NewEthereumWitness creates a new instance of EthereumWitness, bound to a specific deployed contract.
func NewEthereumWitness(address common.Address, backend bind.ContractBackend) (*EthereumWitness, error) {
	contract, err := bindEthereumWitness(address, backend, backend)
	if err != nil {
		return nil, err
	}
	return &EthereumWitness{EthereumWitnessCaller: EthereumWitnessCaller{contract: contract}, EthereumWitnessTransactor: EthereumWitnessTransactor{contract: contract}}, nil
}

// NewEthereumWitnessCaller creates a new read-only instance of EthereumWitness, bound to a specific deployed contract.
func NewEthereumWitnessCaller(address common.Address, caller bind.ContractCaller) (*EthereumWitnessCaller, error) {
	contract, err := bindEthereumWitness(address, caller, nil)
	if err != nil {
		return nil, err
	}
	return &EthereumWitnessCaller{contract: contract}, nil
}

// NewEthereumWitnessTransactor creates a new write-only instance of EthereumWitness, bound to a specific deployed contract.
func NewEthereumWitnessTransactor(address common.Address, transactor bind.ContractTransactor) (*EthereumWitnessTransactor, error) {
	contract, err := bindEthereumWitness(address, nil, transactor)
	if err != nil {
		return nil, err
	}
	return &EthereumWitnessTransactor{contract: contract}, nil
}

// bindEthereumWitness binds a generic wrapper to an already deployed contract.
func bindEthereumWitness(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(EthereumWitnessABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_EthereumWitness *EthereumWitnessRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _EthereumWitness.Contract.EthereumWitnessCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_EthereumWitness *EthereumWitnessRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _EthereumWitness.Contract.EthereumWitnessTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_EthereumWitness *EthereumWitnessRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _EthereumWitness.Contract.EthereumWitnessTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_EthereumWitness *EthereumWitnessCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _EthereumWitness.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_EthereumWitness *EthereumWitnessTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _EthereumWitness.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_EthereumWitness *EthereumWitnessTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _EthereumWitness.Contract.contract.Transact(opts, method, params...)
}

// GetWitness is a free data retrieval call binding the contract method 0x47e72246.
//
// Solidity: function getWitness(version bytes32) constant returns(bytes32[2])
func (_EthereumWitness *EthereumWitnessCaller) GetWitness(opts *bind.CallOpts, version [32]byte) ([2][32]byte, error) {
	var (
		ret0 = new([2][32]byte)
	)
	out := ret0
	err := _EthereumWitness.contract.Call(opts, out, "getWitness", version)
	return *ret0, err
}

// GetWitness is a free data retrieval call binding the contract method 0x47e72246.
//
// Solidity: function getWitness(version bytes32) constant returns(bytes32[2])
func (_EthereumWitness *EthereumWitnessSession) GetWitness(version [32]byte) ([2][32]byte, error) {
	return _EthereumWitness.Contract.GetWitness(&_EthereumWitness.CallOpts, version)
}

// GetWitness is a free data retrieval call binding the contract method 0x47e72246.
//
// Solidity: function getWitness(version bytes32) constant returns(bytes32[2])
func (_EthereumWitness *EthereumWitnessCallerSession) GetWitness(version [32]byte) ([2][32]byte, error) {
	return _EthereumWitness.Contract.GetWitness(&_EthereumWitness.CallOpts, version)
}

// Witness_list is a free data retrieval call binding the contract method 0xd8133496.
//
// Solidity: function witness_list( bytes32,  uint256) constant returns(bytes32)
func (_EthereumWitness *EthereumWitnessCaller) Witness_list(opts *bind.CallOpts, arg0 [32]byte, arg1 *big.Int) ([32]byte, error) {
	var (
		ret0 = new([32]byte)
	)
	out := ret0
	err := _EthereumWitness.contract.Call(opts, out, "witness_list", arg0, arg1)
	return *ret0, err
}

// Witness_list is a free data retrieval call binding the contract method 0xd8133496.
//
// Solidity: function witness_list( bytes32,  uint256) constant returns(bytes32)
func (_EthereumWitness *EthereumWitnessSession) Witness_list(arg0 [32]byte, arg1 *big.Int) ([32]byte, error) {
	return _EthereumWitness.Contract.Witness_list(&_EthereumWitness.CallOpts, arg0, arg1)
}

// Witness_list is a free data retrieval call binding the contract method 0xd8133496.
//
// Solidity: function witness_list( bytes32,  uint256) constant returns(bytes32)
func (_EthereumWitness *EthereumWitnessCallerSession) Witness_list(arg0 [32]byte, arg1 *big.Int) ([32]byte, error) {
	return _EthereumWitness.Contract.Witness_list(&_EthereumWitness.CallOpts, arg0, arg1)
}

// WitnessDocument is a paid mutator transaction binding the contract method 0xe617f632.
//
// Solidity: function witnessDocument(version bytes32, signature bytes32) returns()
func (_EthereumWitness *EthereumWitnessTransactor) WitnessDocument(opts *bind.TransactOpts, version [32]byte, signature [32]byte) (*types.Transaction, error) {
	return _EthereumWitness.contract.Transact(opts, "witnessDocument", version, signature)
}

// WitnessDocument is a paid mutator transaction binding the contract method 0xe617f632.
//
// Solidity: function witnessDocument(version bytes32, signature bytes32) returns()
func (_EthereumWitness *EthereumWitnessSession) WitnessDocument(version [32]byte, signature [32]byte) (*types.Transaction, error) {
	return _EthereumWitness.Contract.WitnessDocument(&_EthereumWitness.TransactOpts, version, signature)
}

// WitnessDocument is a paid mutator transaction binding the contract method 0xe617f632.
//
// Solidity: function witnessDocument(version bytes32, signature bytes32) returns()
func (_EthereumWitness *EthereumWitnessTransactorSession) WitnessDocument(version [32]byte, signature [32]byte) (*types.Transaction, error) {
	return _EthereumWitness.Contract.WitnessDocument(&_EthereumWitness.TransactOpts, version, signature)
}
