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

// EthereumAnchorRepositoryContractABI is the input ABI used to generate the binding from.
const EthereumAnchorRepositoryContractABI = "[{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"commits\",\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"preCommits\",\"outputs\":[{\"name\":\"signingRoot\",\"type\":\"bytes32\"},{\"name\":\"centrifugeId\",\"type\":\"uint48\"},{\"name\":\"expirationBlock\",\"type\":\"uint32\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"name\":\"_identityRegistry\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"anchorId\",\"type\":\"uint256\"},{\"indexed\":true,\"name\":\"centrifugeId\",\"type\":\"uint48\"},{\"indexed\":false,\"name\":\"documentRoot\",\"type\":\"bytes32\"},{\"indexed\":false,\"name\":\"blockHeight\",\"type\":\"uint32\"}],\"name\":\"AnchorCommitted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"anchorId\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"blockHeight\",\"type\":\"uint32\"}],\"name\":\"AnchorPreCommitted\",\"type\":\"event\"},{\"constant\":false,\"inputs\":[{\"name\":\"_anchorId\",\"type\":\"uint256\"},{\"name\":\"_signingRoot\",\"type\":\"bytes32\"},{\"name\":\"_centrifugeId\",\"type\":\"uint48\"},{\"name\":\"_signature\",\"type\":\"bytes\"},{\"name\":\"_expirationBlock\",\"type\":\"uint256\"}],\"name\":\"preCommit\",\"outputs\":[],\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_anchorId\",\"type\":\"uint256\"},{\"name\":\"_documentRoot\",\"type\":\"bytes32\"},{\"name\":\"_centrifugeId\",\"type\":\"uint48\"},{\"name\":\"_documentProofs\",\"type\":\"bytes32[]\"},{\"name\":\"_signature\",\"type\":\"bytes\"}],\"name\":\"commit\",\"outputs\":[],\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_anchorId\",\"type\":\"uint256\"}],\"name\":\"getAnchorById\",\"outputs\":[{\"name\":\"anchorId\",\"type\":\"uint256\"},{\"name\":\"documentRoot\",\"type\":\"bytes32\"},{\"name\":\"centrifugeId\",\"type\":\"uint48\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_anchorId\",\"type\":\"uint256\"}],\"name\":\"hasValidPreCommit\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"}]"

// EthereumAnchorRepositoryContract is an auto generated Go binding around an Ethereum contract.
type EthereumAnchorRepositoryContract struct {
	EthereumAnchorRepositoryContractCaller     // Read-only binding to the contract
	EthereumAnchorRepositoryContractTransactor // Write-only binding to the contract
	EthereumAnchorRepositoryContractFilterer   // Log filterer for contract events
}

// EthereumAnchorRepositoryContractCaller is an auto generated read-only Go binding around an Ethereum contract.
type EthereumAnchorRepositoryContractCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// EthereumAnchorRepositoryContractTransactor is an auto generated write-only Go binding around an Ethereum contract.
type EthereumAnchorRepositoryContractTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// EthereumAnchorRepositoryContractFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type EthereumAnchorRepositoryContractFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// EthereumAnchorRepositoryContractSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type EthereumAnchorRepositoryContractSession struct {
	Contract     *EthereumAnchorRepositoryContract // Generic contract binding to set the session for
	CallOpts     bind.CallOpts                     // Call options to use throughout this session
	TransactOpts bind.TransactOpts                 // Transaction auth options to use throughout this session
}

// EthereumAnchorRepositoryContractCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type EthereumAnchorRepositoryContractCallerSession struct {
	Contract *EthereumAnchorRepositoryContractCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts                           // Call options to use throughout this session
}

// EthereumAnchorRepositoryContractTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type EthereumAnchorRepositoryContractTransactorSession struct {
	Contract     *EthereumAnchorRepositoryContractTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts                           // Transaction auth options to use throughout this session
}

// EthereumAnchorRepositoryContractRaw is an auto generated low-level Go binding around an Ethereum contract.
type EthereumAnchorRepositoryContractRaw struct {
	Contract *EthereumAnchorRepositoryContract // Generic contract binding to access the raw methods on
}

// EthereumAnchorRepositoryContractCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type EthereumAnchorRepositoryContractCallerRaw struct {
	Contract *EthereumAnchorRepositoryContractCaller // Generic read-only contract binding to access the raw methods on
}

// EthereumAnchorRepositoryContractTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type EthereumAnchorRepositoryContractTransactorRaw struct {
	Contract *EthereumAnchorRepositoryContractTransactor // Generic write-only contract binding to access the raw methods on
}

// NewEthereumAnchorRepositoryContract creates a new instance of EthereumAnchorRepositoryContract, bound to a specific deployed contract.
func NewEthereumAnchorRepositoryContract(address common.Address, backend bind.ContractBackend) (*EthereumAnchorRepositoryContract, error) {
	contract, err := bindEthereumAnchorRepositoryContract(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &EthereumAnchorRepositoryContract{EthereumAnchorRepositoryContractCaller: EthereumAnchorRepositoryContractCaller{contract: contract}, EthereumAnchorRepositoryContractTransactor: EthereumAnchorRepositoryContractTransactor{contract: contract}, EthereumAnchorRepositoryContractFilterer: EthereumAnchorRepositoryContractFilterer{contract: contract}}, nil
}

// NewEthereumAnchorRepositoryContractCaller creates a new read-only instance of EthereumAnchorRepositoryContract, bound to a specific deployed contract.
func NewEthereumAnchorRepositoryContractCaller(address common.Address, caller bind.ContractCaller) (*EthereumAnchorRepositoryContractCaller, error) {
	contract, err := bindEthereumAnchorRepositoryContract(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &EthereumAnchorRepositoryContractCaller{contract: contract}, nil
}

// NewEthereumAnchorRepositoryContractTransactor creates a new write-only instance of EthereumAnchorRepositoryContract, bound to a specific deployed contract.
func NewEthereumAnchorRepositoryContractTransactor(address common.Address, transactor bind.ContractTransactor) (*EthereumAnchorRepositoryContractTransactor, error) {
	contract, err := bindEthereumAnchorRepositoryContract(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &EthereumAnchorRepositoryContractTransactor{contract: contract}, nil
}

// NewEthereumAnchorRepositoryContractFilterer creates a new log filterer instance of EthereumAnchorRepositoryContract, bound to a specific deployed contract.
func NewEthereumAnchorRepositoryContractFilterer(address common.Address, filterer bind.ContractFilterer) (*EthereumAnchorRepositoryContractFilterer, error) {
	contract, err := bindEthereumAnchorRepositoryContract(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &EthereumAnchorRepositoryContractFilterer{contract: contract}, nil
}

// bindEthereumAnchorRepositoryContract binds a generic wrapper to an already deployed contract.
func bindEthereumAnchorRepositoryContract(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(EthereumAnchorRepositoryContractABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_EthereumAnchorRepositoryContract *EthereumAnchorRepositoryContractRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _EthereumAnchorRepositoryContract.Contract.EthereumAnchorRepositoryContractCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_EthereumAnchorRepositoryContract *EthereumAnchorRepositoryContractRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _EthereumAnchorRepositoryContract.Contract.EthereumAnchorRepositoryContractTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_EthereumAnchorRepositoryContract *EthereumAnchorRepositoryContractRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _EthereumAnchorRepositoryContract.Contract.EthereumAnchorRepositoryContractTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_EthereumAnchorRepositoryContract *EthereumAnchorRepositoryContractCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _EthereumAnchorRepositoryContract.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_EthereumAnchorRepositoryContract *EthereumAnchorRepositoryContractTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _EthereumAnchorRepositoryContract.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_EthereumAnchorRepositoryContract *EthereumAnchorRepositoryContractTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _EthereumAnchorRepositoryContract.Contract.contract.Transact(opts, method, params...)
}

// Commits is a free data retrieval call binding the contract method 0xc7c4a615.
//
// Solidity: function commits( uint256) constant returns(bytes32)
func (_EthereumAnchorRepositoryContract *EthereumAnchorRepositoryContractCaller) Commits(opts *bind.CallOpts, arg0 *big.Int) ([32]byte, error) {
	var (
		ret0 = new([32]byte)
	)
	out := ret0
	err := _EthereumAnchorRepositoryContract.contract.Call(opts, out, "commits", arg0)
	return *ret0, err
}

// Commits is a free data retrieval call binding the contract method 0xc7c4a615.
//
// Solidity: function commits( uint256) constant returns(bytes32)
func (_EthereumAnchorRepositoryContract *EthereumAnchorRepositoryContractSession) Commits(arg0 *big.Int) ([32]byte, error) {
	return _EthereumAnchorRepositoryContract.Contract.Commits(&_EthereumAnchorRepositoryContract.CallOpts, arg0)
}

// Commits is a free data retrieval call binding the contract method 0xc7c4a615.
//
// Solidity: function commits( uint256) constant returns(bytes32)
func (_EthereumAnchorRepositoryContract *EthereumAnchorRepositoryContractCallerSession) Commits(arg0 *big.Int) ([32]byte, error) {
	return _EthereumAnchorRepositoryContract.Contract.Commits(&_EthereumAnchorRepositoryContract.CallOpts, arg0)
}

// GetAnchorById is a free data retrieval call binding the contract method 0x32bf361b.
//
// Solidity: function getAnchorById(_anchorId uint256) constant returns(anchorId uint256, documentRoot bytes32, centrifugeId uint48)
func (_EthereumAnchorRepositoryContract *EthereumAnchorRepositoryContractCaller) GetAnchorById(opts *bind.CallOpts, _anchorId *big.Int) (struct {
	AnchorId     *big.Int
	DocumentRoot [32]byte
	CentrifugeId *big.Int
}, error) {
	ret := new(struct {
		AnchorId     *big.Int
		DocumentRoot [32]byte
		CentrifugeId *big.Int
	})
	out := ret
	err := _EthereumAnchorRepositoryContract.contract.Call(opts, out, "getAnchorById", _anchorId)
	return *ret, err
}

// GetAnchorById is a free data retrieval call binding the contract method 0x32bf361b.
//
// Solidity: function getAnchorById(_anchorId uint256) constant returns(anchorId uint256, documentRoot bytes32, centrifugeId uint48)
func (_EthereumAnchorRepositoryContract *EthereumAnchorRepositoryContractSession) GetAnchorById(_anchorId *big.Int) (struct {
	AnchorId     *big.Int
	DocumentRoot [32]byte
	CentrifugeId *big.Int
}, error) {
	return _EthereumAnchorRepositoryContract.Contract.GetAnchorById(&_EthereumAnchorRepositoryContract.CallOpts, _anchorId)
}

// GetAnchorById is a free data retrieval call binding the contract method 0x32bf361b.
//
// Solidity: function getAnchorById(_anchorId uint256) constant returns(anchorId uint256, documentRoot bytes32, centrifugeId uint48)
func (_EthereumAnchorRepositoryContract *EthereumAnchorRepositoryContractCallerSession) GetAnchorById(_anchorId *big.Int) (struct {
	AnchorId     *big.Int
	DocumentRoot [32]byte
	CentrifugeId *big.Int
}, error) {
	return _EthereumAnchorRepositoryContract.Contract.GetAnchorById(&_EthereumAnchorRepositoryContract.CallOpts, _anchorId)
}

// HasValidPreCommit is a free data retrieval call binding the contract method 0xb5c7d034.
//
// Solidity: function hasValidPreCommit(_anchorId uint256) constant returns(bool)
func (_EthereumAnchorRepositoryContract *EthereumAnchorRepositoryContractCaller) HasValidPreCommit(opts *bind.CallOpts, _anchorId *big.Int) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _EthereumAnchorRepositoryContract.contract.Call(opts, out, "hasValidPreCommit", _anchorId)
	return *ret0, err
}

// HasValidPreCommit is a free data retrieval call binding the contract method 0xb5c7d034.
//
// Solidity: function hasValidPreCommit(_anchorId uint256) constant returns(bool)
func (_EthereumAnchorRepositoryContract *EthereumAnchorRepositoryContractSession) HasValidPreCommit(_anchorId *big.Int) (bool, error) {
	return _EthereumAnchorRepositoryContract.Contract.HasValidPreCommit(&_EthereumAnchorRepositoryContract.CallOpts, _anchorId)
}

// HasValidPreCommit is a free data retrieval call binding the contract method 0xb5c7d034.
//
// Solidity: function hasValidPreCommit(_anchorId uint256) constant returns(bool)
func (_EthereumAnchorRepositoryContract *EthereumAnchorRepositoryContractCallerSession) HasValidPreCommit(_anchorId *big.Int) (bool, error) {
	return _EthereumAnchorRepositoryContract.Contract.HasValidPreCommit(&_EthereumAnchorRepositoryContract.CallOpts, _anchorId)
}

// PreCommits is a free data retrieval call binding the contract method 0xd04cc3da.
//
// Solidity: function preCommits( uint256) constant returns(signingRoot bytes32, centrifugeId uint48, expirationBlock uint32)
func (_EthereumAnchorRepositoryContract *EthereumAnchorRepositoryContractCaller) PreCommits(opts *bind.CallOpts, arg0 *big.Int) (struct {
	SigningRoot     [32]byte
	CentrifugeId    *big.Int
	ExpirationBlock uint32
}, error) {
	ret := new(struct {
		SigningRoot     [32]byte
		CentrifugeId    *big.Int
		ExpirationBlock uint32
	})
	out := ret
	err := _EthereumAnchorRepositoryContract.contract.Call(opts, out, "preCommits", arg0)
	return *ret, err
}

// PreCommits is a free data retrieval call binding the contract method 0xd04cc3da.
//
// Solidity: function preCommits( uint256) constant returns(signingRoot bytes32, centrifugeId uint48, expirationBlock uint32)
func (_EthereumAnchorRepositoryContract *EthereumAnchorRepositoryContractSession) PreCommits(arg0 *big.Int) (struct {
	SigningRoot     [32]byte
	CentrifugeId    *big.Int
	ExpirationBlock uint32
}, error) {
	return _EthereumAnchorRepositoryContract.Contract.PreCommits(&_EthereumAnchorRepositoryContract.CallOpts, arg0)
}

// PreCommits is a free data retrieval call binding the contract method 0xd04cc3da.
//
// Solidity: function preCommits( uint256) constant returns(signingRoot bytes32, centrifugeId uint48, expirationBlock uint32)
func (_EthereumAnchorRepositoryContract *EthereumAnchorRepositoryContractCallerSession) PreCommits(arg0 *big.Int) (struct {
	SigningRoot     [32]byte
	CentrifugeId    *big.Int
	ExpirationBlock uint32
}, error) {
	return _EthereumAnchorRepositoryContract.Contract.PreCommits(&_EthereumAnchorRepositoryContract.CallOpts, arg0)
}

// Commit is a paid mutator transaction binding the contract method 0xad90a76b.
//
// Solidity: function commit(_anchorId uint256, _documentRoot bytes32, _centrifugeId uint48, _documentProofs bytes32[], _signature bytes) returns()
func (_EthereumAnchorRepositoryContract *EthereumAnchorRepositoryContractTransactor) Commit(opts *bind.TransactOpts, _anchorId *big.Int, _documentRoot [32]byte, _centrifugeId *big.Int, _documentProofs [][32]byte, _signature []byte) (*types.Transaction, error) {
	return _EthereumAnchorRepositoryContract.contract.Transact(opts, "commit", _anchorId, _documentRoot, _centrifugeId, _documentProofs, _signature)
}

// Commit is a paid mutator transaction binding the contract method 0xad90a76b.
//
// Solidity: function commit(_anchorId uint256, _documentRoot bytes32, _centrifugeId uint48, _documentProofs bytes32[], _signature bytes) returns()
func (_EthereumAnchorRepositoryContract *EthereumAnchorRepositoryContractSession) Commit(_anchorId *big.Int, _documentRoot [32]byte, _centrifugeId *big.Int, _documentProofs [][32]byte, _signature []byte) (*types.Transaction, error) {
	return _EthereumAnchorRepositoryContract.Contract.Commit(&_EthereumAnchorRepositoryContract.TransactOpts, _anchorId, _documentRoot, _centrifugeId, _documentProofs, _signature)
}

// Commit is a paid mutator transaction binding the contract method 0xad90a76b.
//
// Solidity: function commit(_anchorId uint256, _documentRoot bytes32, _centrifugeId uint48, _documentProofs bytes32[], _signature bytes) returns()
func (_EthereumAnchorRepositoryContract *EthereumAnchorRepositoryContractTransactorSession) Commit(_anchorId *big.Int, _documentRoot [32]byte, _centrifugeId *big.Int, _documentProofs [][32]byte, _signature []byte) (*types.Transaction, error) {
	return _EthereumAnchorRepositoryContract.Contract.Commit(&_EthereumAnchorRepositoryContract.TransactOpts, _anchorId, _documentRoot, _centrifugeId, _documentProofs, _signature)
}

// PreCommit is a paid mutator transaction binding the contract method 0xf098d34c.
//
// Solidity: function preCommit(_anchorId uint256, _signingRoot bytes32, _centrifugeId uint48, _signature bytes, _expirationBlock uint256) returns()
func (_EthereumAnchorRepositoryContract *EthereumAnchorRepositoryContractTransactor) PreCommit(opts *bind.TransactOpts, _anchorId *big.Int, _signingRoot [32]byte, _centrifugeId *big.Int, _signature []byte, _expirationBlock *big.Int) (*types.Transaction, error) {
	return _EthereumAnchorRepositoryContract.contract.Transact(opts, "preCommit", _anchorId, _signingRoot, _centrifugeId, _signature, _expirationBlock)
}

// PreCommit is a paid mutator transaction binding the contract method 0xf098d34c.
//
// Solidity: function preCommit(_anchorId uint256, _signingRoot bytes32, _centrifugeId uint48, _signature bytes, _expirationBlock uint256) returns()
func (_EthereumAnchorRepositoryContract *EthereumAnchorRepositoryContractSession) PreCommit(_anchorId *big.Int, _signingRoot [32]byte, _centrifugeId *big.Int, _signature []byte, _expirationBlock *big.Int) (*types.Transaction, error) {
	return _EthereumAnchorRepositoryContract.Contract.PreCommit(&_EthereumAnchorRepositoryContract.TransactOpts, _anchorId, _signingRoot, _centrifugeId, _signature, _expirationBlock)
}

// PreCommit is a paid mutator transaction binding the contract method 0xf098d34c.
//
// Solidity: function preCommit(_anchorId uint256, _signingRoot bytes32, _centrifugeId uint48, _signature bytes, _expirationBlock uint256) returns()
func (_EthereumAnchorRepositoryContract *EthereumAnchorRepositoryContractTransactorSession) PreCommit(_anchorId *big.Int, _signingRoot [32]byte, _centrifugeId *big.Int, _signature []byte, _expirationBlock *big.Int) (*types.Transaction, error) {
	return _EthereumAnchorRepositoryContract.Contract.PreCommit(&_EthereumAnchorRepositoryContract.TransactOpts, _anchorId, _signingRoot, _centrifugeId, _signature, _expirationBlock)
}

// EthereumAnchorRepositoryContractAnchorCommittedIterator is returned from FilterAnchorCommitted and is used to iterate over the raw logs and unpacked data for AnchorCommitted events raised by the EthereumAnchorRepositoryContract contract.
type EthereumAnchorRepositoryContractAnchorCommittedIterator struct {
	Event *EthereumAnchorRepositoryContractAnchorCommitted // Event containing the contract specifics and raw log

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
func (it *EthereumAnchorRepositoryContractAnchorCommittedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(EthereumAnchorRepositoryContractAnchorCommitted)
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
		it.Event = new(EthereumAnchorRepositoryContractAnchorCommitted)
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
func (it *EthereumAnchorRepositoryContractAnchorCommittedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *EthereumAnchorRepositoryContractAnchorCommittedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// EthereumAnchorRepositoryContractAnchorCommitted represents a AnchorCommitted event raised by the EthereumAnchorRepositoryContract contract.
type EthereumAnchorRepositoryContractAnchorCommitted struct {
	From         common.Address
	AnchorId     *big.Int
	CentrifugeId *big.Int
	DocumentRoot [32]byte
	BlockHeight  uint32
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterAnchorCommitted is a free log retrieval operation binding the contract event 0x307682cd6852f8e18285627aa89f76c08e16e4978ab1c80dedbc5e1d43ddba66.
//
// Solidity: e AnchorCommitted(from indexed address, anchorId indexed uint256, centrifugeId indexed uint48, documentRoot bytes32, blockHeight uint32)
func (_EthereumAnchorRepositoryContract *EthereumAnchorRepositoryContractFilterer) FilterAnchorCommitted(opts *bind.FilterOpts, from []common.Address, anchorId []*big.Int, centrifugeId []*big.Int) (*EthereumAnchorRepositoryContractAnchorCommittedIterator, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var anchorIdRule []interface{}
	for _, anchorIdItem := range anchorId {
		anchorIdRule = append(anchorIdRule, anchorIdItem)
	}
	var centrifugeIdRule []interface{}
	for _, centrifugeIdItem := range centrifugeId {
		centrifugeIdRule = append(centrifugeIdRule, centrifugeIdItem)
	}

	logs, sub, err := _EthereumAnchorRepositoryContract.contract.FilterLogs(opts, "AnchorCommitted", fromRule, anchorIdRule, centrifugeIdRule)
	if err != nil {
		return nil, err
	}
	return &EthereumAnchorRepositoryContractAnchorCommittedIterator{contract: _EthereumAnchorRepositoryContract.contract, event: "AnchorCommitted", logs: logs, sub: sub}, nil
}

// WatchAnchorCommitted is a free log subscription operation binding the contract event 0x307682cd6852f8e18285627aa89f76c08e16e4978ab1c80dedbc5e1d43ddba66.
//
// Solidity: e AnchorCommitted(from indexed address, anchorId indexed uint256, centrifugeId indexed uint48, documentRoot bytes32, blockHeight uint32)
func (_EthereumAnchorRepositoryContract *EthereumAnchorRepositoryContractFilterer) WatchAnchorCommitted(opts *bind.WatchOpts, sink chan<- *EthereumAnchorRepositoryContractAnchorCommitted, from []common.Address, anchorId []*big.Int, centrifugeId []*big.Int) (event.Subscription, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var anchorIdRule []interface{}
	for _, anchorIdItem := range anchorId {
		anchorIdRule = append(anchorIdRule, anchorIdItem)
	}
	var centrifugeIdRule []interface{}
	for _, centrifugeIdItem := range centrifugeId {
		centrifugeIdRule = append(centrifugeIdRule, centrifugeIdItem)
	}

	logs, sub, err := _EthereumAnchorRepositoryContract.contract.WatchLogs(opts, "AnchorCommitted", fromRule, anchorIdRule, centrifugeIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(EthereumAnchorRepositoryContractAnchorCommitted)
				if err := _EthereumAnchorRepositoryContract.contract.UnpackLog(event, "AnchorCommitted", log); err != nil {
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

// EthereumAnchorRepositoryContractAnchorPreCommittedIterator is returned from FilterAnchorPreCommitted and is used to iterate over the raw logs and unpacked data for AnchorPreCommitted events raised by the EthereumAnchorRepositoryContract contract.
type EthereumAnchorRepositoryContractAnchorPreCommittedIterator struct {
	Event *EthereumAnchorRepositoryContractAnchorPreCommitted // Event containing the contract specifics and raw log

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
func (it *EthereumAnchorRepositoryContractAnchorPreCommittedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(EthereumAnchorRepositoryContractAnchorPreCommitted)
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
		it.Event = new(EthereumAnchorRepositoryContractAnchorPreCommitted)
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
func (it *EthereumAnchorRepositoryContractAnchorPreCommittedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *EthereumAnchorRepositoryContractAnchorPreCommittedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// EthereumAnchorRepositoryContractAnchorPreCommitted represents a AnchorPreCommitted event raised by the EthereumAnchorRepositoryContract contract.
type EthereumAnchorRepositoryContractAnchorPreCommitted struct {
	From        common.Address
	AnchorId    *big.Int
	BlockHeight uint32
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterAnchorPreCommitted is a free log retrieval operation binding the contract event 0xaa2928be4e330731bc1f0289edebfc72ccb9979ffc703a3de4edd8ea760462da.
//
// Solidity: e AnchorPreCommitted(from indexed address, anchorId indexed uint256, blockHeight uint32)
func (_EthereumAnchorRepositoryContract *EthereumAnchorRepositoryContractFilterer) FilterAnchorPreCommitted(opts *bind.FilterOpts, from []common.Address, anchorId []*big.Int) (*EthereumAnchorRepositoryContractAnchorPreCommittedIterator, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var anchorIdRule []interface{}
	for _, anchorIdItem := range anchorId {
		anchorIdRule = append(anchorIdRule, anchorIdItem)
	}

	logs, sub, err := _EthereumAnchorRepositoryContract.contract.FilterLogs(opts, "AnchorPreCommitted", fromRule, anchorIdRule)
	if err != nil {
		return nil, err
	}
	return &EthereumAnchorRepositoryContractAnchorPreCommittedIterator{contract: _EthereumAnchorRepositoryContract.contract, event: "AnchorPreCommitted", logs: logs, sub: sub}, nil
}

// WatchAnchorPreCommitted is a free log subscription operation binding the contract event 0xaa2928be4e330731bc1f0289edebfc72ccb9979ffc703a3de4edd8ea760462da.
//
// Solidity: e AnchorPreCommitted(from indexed address, anchorId indexed uint256, blockHeight uint32)
func (_EthereumAnchorRepositoryContract *EthereumAnchorRepositoryContractFilterer) WatchAnchorPreCommitted(opts *bind.WatchOpts, sink chan<- *EthereumAnchorRepositoryContractAnchorPreCommitted, from []common.Address, anchorId []*big.Int) (event.Subscription, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var anchorIdRule []interface{}
	for _, anchorIdItem := range anchorId {
		anchorIdRule = append(anchorIdRule, anchorIdItem)
	}

	logs, sub, err := _EthereumAnchorRepositoryContract.contract.WatchLogs(opts, "AnchorPreCommitted", fromRule, anchorIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(EthereumAnchorRepositoryContractAnchorPreCommitted)
				if err := _EthereumAnchorRepositoryContract.contract.UnpackLog(event, "AnchorPreCommitted", log); err != nil {
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
