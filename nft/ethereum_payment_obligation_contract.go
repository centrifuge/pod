// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package nft

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

// EthereumPaymentObligationContractABI is the input ABI used to generate the binding from.
const EthereumPaymentObligationContractABI = "[{\"constant\":true,\"inputs\":[{\"name\":\"_interfaceId\",\"type\":\"bytes4\"}],\"name\":\"supportsInterface\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"name\",\"outputs\":[{\"name\":\"\",\"type\":\"string\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_tokenId\",\"type\":\"uint256\"}],\"name\":\"getApproved\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_to\",\"type\":\"address\"},{\"name\":\"_tokenId\",\"type\":\"uint256\"}],\"name\":\"approve\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"totalSupply\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"InterfaceId_ERC165\",\"outputs\":[{\"name\":\"\",\"type\":\"bytes4\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_from\",\"type\":\"address\"},{\"name\":\"_to\",\"type\":\"address\"},{\"name\":\"_tokenId\",\"type\":\"uint256\"}],\"name\":\"transferFrom\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_owner\",\"type\":\"address\"},{\"name\":\"_index\",\"type\":\"uint256\"}],\"name\":\"tokenOfOwnerByIndex\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_from\",\"type\":\"address\"},{\"name\":\"_to\",\"type\":\"address\"},{\"name\":\"_tokenId\",\"type\":\"uint256\"}],\"name\":\"safeTransferFrom\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_tokenId\",\"type\":\"uint256\"}],\"name\":\"exists\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_index\",\"type\":\"uint256\"}],\"name\":\"tokenByIndex\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"anchorRegistry\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_tokenId\",\"type\":\"uint256\"}],\"name\":\"ownerOf\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_owner\",\"type\":\"address\"}],\"name\":\"balanceOf\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"symbol\",\"outputs\":[{\"name\":\"\",\"type\":\"string\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_to\",\"type\":\"address\"},{\"name\":\"_approved\",\"type\":\"bool\"}],\"name\":\"setApprovalForAll\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_from\",\"type\":\"address\"},{\"name\":\"_to\",\"type\":\"address\"},{\"name\":\"_tokenId\",\"type\":\"uint256\"},{\"name\":\"_data\",\"type\":\"bytes\"}],\"name\":\"safeTransferFrom\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_tokenId\",\"type\":\"uint256\"}],\"name\":\"tokenURI\",\"outputs\":[{\"name\":\"\",\"type\":\"string\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_owner\",\"type\":\"address\"},{\"name\":\"_operator\",\"type\":\"address\"}],\"name\":\"isApprovedForAll\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"name\":\"_name\",\"type\":\"string\"},{\"name\":\"_symbol\",\"type\":\"string\"},{\"name\":\"_anchorRegistry\",\"type\":\"address\"},{\"name\":\"_identityRegistry\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"tokenId\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"tokenURI\",\"type\":\"string\"}],\"name\":\"PaymentObligationMinted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"_from\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"_to\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"_tokenId\",\"type\":\"uint256\"}],\"name\":\"Transfer\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"_owner\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"_approved\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"_tokenId\",\"type\":\"uint256\"}],\"name\":\"Approval\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"_owner\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"_operator\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"_approved\",\"type\":\"bool\"}],\"name\":\"ApprovalForAll\",\"type\":\"event\"},{\"constant\":false,\"inputs\":[{\"name\":\"_to\",\"type\":\"address\"},{\"name\":\"_tokenId\",\"type\":\"uint256\"},{\"name\":\"_tokenURI\",\"type\":\"string\"},{\"name\":\"_anchorId\",\"type\":\"uint256\"},{\"name\":\"_merkleRoot\",\"type\":\"bytes32\"},{\"name\":\"_collaboratorField\",\"type\":\"string\"},{\"name\":\"_values\",\"type\":\"string[5]\"},{\"name\":\"_salts\",\"type\":\"bytes32[5]\"},{\"name\":\"_proofs\",\"type\":\"bytes32[][5]\"}],\"name\":\"mint\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_tokenId\",\"type\":\"uint256\"}],\"name\":\"getTokenDetails\",\"outputs\":[{\"name\":\"grossAmount\",\"type\":\"string\"},{\"name\":\"currency\",\"type\":\"string\"},{\"name\":\"dueDate\",\"type\":\"string\"},{\"name\":\"anchorId\",\"type\":\"uint256\"},{\"name\":\"documentRoot\",\"type\":\"bytes32\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"}]"

// EthereumPaymentObligationContract is an auto generated Go binding around an Ethereum contract.
type EthereumPaymentObligationContract struct {
	EthereumPaymentObligationContractCaller     // Read-only binding to the contract
	EthereumPaymentObligationContractTransactor // Write-only binding to the contract
	EthereumPaymentObligationContractFilterer   // Log filterer for contract events
}

// EthereumPaymentObligationContractCaller is an auto generated read-only Go binding around an Ethereum contract.
type EthereumPaymentObligationContractCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// EthereumPaymentObligationContractTransactor is an auto generated write-only Go binding around an Ethereum contract.
type EthereumPaymentObligationContractTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// EthereumPaymentObligationContractFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type EthereumPaymentObligationContractFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// EthereumPaymentObligationContractSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type EthereumPaymentObligationContractSession struct {
	Contract     *EthereumPaymentObligationContract // Generic contract binding to set the session for
	CallOpts     bind.CallOpts                      // Call options to use throughout this session
	TransactOpts bind.TransactOpts                  // Transaction auth options to use throughout this session
}

// EthereumPaymentObligationContractCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type EthereumPaymentObligationContractCallerSession struct {
	Contract *EthereumPaymentObligationContractCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts                            // Call options to use throughout this session
}

// EthereumPaymentObligationContractTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type EthereumPaymentObligationContractTransactorSession struct {
	Contract     *EthereumPaymentObligationContractTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts                            // Transaction auth options to use throughout this session
}

// EthereumPaymentObligationContractRaw is an auto generated low-level Go binding around an Ethereum contract.
type EthereumPaymentObligationContractRaw struct {
	Contract *EthereumPaymentObligationContract // Generic contract binding to access the raw methods on
}

// EthereumPaymentObligationContractCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type EthereumPaymentObligationContractCallerRaw struct {
	Contract *EthereumPaymentObligationContractCaller // Generic read-only contract binding to access the raw methods on
}

// EthereumPaymentObligationContractTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type EthereumPaymentObligationContractTransactorRaw struct {
	Contract *EthereumPaymentObligationContractTransactor // Generic write-only contract binding to access the raw methods on
}

// NewEthereumPaymentObligationContract creates a new instance of EthereumPaymentObligationContract, bound to a specific deployed contract.
func NewEthereumPaymentObligationContract(address common.Address, backend bind.ContractBackend) (*EthereumPaymentObligationContract, error) {
	contract, err := bindEthereumPaymentObligationContract(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &EthereumPaymentObligationContract{EthereumPaymentObligationContractCaller: EthereumPaymentObligationContractCaller{contract: contract}, EthereumPaymentObligationContractTransactor: EthereumPaymentObligationContractTransactor{contract: contract}, EthereumPaymentObligationContractFilterer: EthereumPaymentObligationContractFilterer{contract: contract}}, nil
}

// NewEthereumPaymentObligationContractCaller creates a new read-only instance of EthereumPaymentObligationContract, bound to a specific deployed contract.
func NewEthereumPaymentObligationContractCaller(address common.Address, caller bind.ContractCaller) (*EthereumPaymentObligationContractCaller, error) {
	contract, err := bindEthereumPaymentObligationContract(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &EthereumPaymentObligationContractCaller{contract: contract}, nil
}

// NewEthereumPaymentObligationContractTransactor creates a new write-only instance of EthereumPaymentObligationContract, bound to a specific deployed contract.
func NewEthereumPaymentObligationContractTransactor(address common.Address, transactor bind.ContractTransactor) (*EthereumPaymentObligationContractTransactor, error) {
	contract, err := bindEthereumPaymentObligationContract(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &EthereumPaymentObligationContractTransactor{contract: contract}, nil
}

// NewEthereumPaymentObligationContractFilterer creates a new log filterer instance of EthereumPaymentObligationContract, bound to a specific deployed contract.
func NewEthereumPaymentObligationContractFilterer(address common.Address, filterer bind.ContractFilterer) (*EthereumPaymentObligationContractFilterer, error) {
	contract, err := bindEthereumPaymentObligationContract(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &EthereumPaymentObligationContractFilterer{contract: contract}, nil
}

// bindEthereumPaymentObligationContract binds a generic wrapper to an already deployed contract.
func bindEthereumPaymentObligationContract(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(EthereumPaymentObligationContractABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_EthereumPaymentObligationContract *EthereumPaymentObligationContractRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _EthereumPaymentObligationContract.Contract.EthereumPaymentObligationContractCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_EthereumPaymentObligationContract *EthereumPaymentObligationContractRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _EthereumPaymentObligationContract.Contract.EthereumPaymentObligationContractTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_EthereumPaymentObligationContract *EthereumPaymentObligationContractRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _EthereumPaymentObligationContract.Contract.EthereumPaymentObligationContractTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_EthereumPaymentObligationContract *EthereumPaymentObligationContractCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _EthereumPaymentObligationContract.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_EthereumPaymentObligationContract *EthereumPaymentObligationContractTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _EthereumPaymentObligationContract.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_EthereumPaymentObligationContract *EthereumPaymentObligationContractTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _EthereumPaymentObligationContract.Contract.contract.Transact(opts, method, params...)
}

// InterfaceIdERC165 is a free data retrieval call binding the contract method 0x19fa8f50.
//
// Solidity: function InterfaceId_ERC165() constant returns(bytes4)
func (_EthereumPaymentObligationContract *EthereumPaymentObligationContractCaller) InterfaceIdERC165(opts *bind.CallOpts) ([4]byte, error) {
	var (
		ret0 = new([4]byte)
	)
	out := ret0
	err := _EthereumPaymentObligationContract.contract.Call(opts, out, "InterfaceId_ERC165")
	return *ret0, err
}

// InterfaceIdERC165 is a free data retrieval call binding the contract method 0x19fa8f50.
//
// Solidity: function InterfaceId_ERC165() constant returns(bytes4)
func (_EthereumPaymentObligationContract *EthereumPaymentObligationContractSession) InterfaceIdERC165() ([4]byte, error) {
	return _EthereumPaymentObligationContract.Contract.InterfaceIdERC165(&_EthereumPaymentObligationContract.CallOpts)
}

// InterfaceIdERC165 is a free data retrieval call binding the contract method 0x19fa8f50.
//
// Solidity: function InterfaceId_ERC165() constant returns(bytes4)
func (_EthereumPaymentObligationContract *EthereumPaymentObligationContractCallerSession) InterfaceIdERC165() ([4]byte, error) {
	return _EthereumPaymentObligationContract.Contract.InterfaceIdERC165(&_EthereumPaymentObligationContract.CallOpts)
}

// AnchorRegistry is a free data retrieval call binding the contract method 0x5a180c0a.
//
// Solidity: function anchorRegistry() constant returns(address)
func (_EthereumPaymentObligationContract *EthereumPaymentObligationContractCaller) AnchorRegistry(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _EthereumPaymentObligationContract.contract.Call(opts, out, "anchorRegistry")
	return *ret0, err
}

// AnchorRegistry is a free data retrieval call binding the contract method 0x5a180c0a.
//
// Solidity: function anchorRegistry() constant returns(address)
func (_EthereumPaymentObligationContract *EthereumPaymentObligationContractSession) AnchorRegistry() (common.Address, error) {
	return _EthereumPaymentObligationContract.Contract.AnchorRegistry(&_EthereumPaymentObligationContract.CallOpts)
}

// AnchorRegistry is a free data retrieval call binding the contract method 0x5a180c0a.
//
// Solidity: function anchorRegistry() constant returns(address)
func (_EthereumPaymentObligationContract *EthereumPaymentObligationContractCallerSession) AnchorRegistry() (common.Address, error) {
	return _EthereumPaymentObligationContract.Contract.AnchorRegistry(&_EthereumPaymentObligationContract.CallOpts)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(_owner address) constant returns(uint256)
func (_EthereumPaymentObligationContract *EthereumPaymentObligationContractCaller) BalanceOf(opts *bind.CallOpts, _owner common.Address) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _EthereumPaymentObligationContract.contract.Call(opts, out, "balanceOf", _owner)
	return *ret0, err
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(_owner address) constant returns(uint256)
func (_EthereumPaymentObligationContract *EthereumPaymentObligationContractSession) BalanceOf(_owner common.Address) (*big.Int, error) {
	return _EthereumPaymentObligationContract.Contract.BalanceOf(&_EthereumPaymentObligationContract.CallOpts, _owner)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(_owner address) constant returns(uint256)
func (_EthereumPaymentObligationContract *EthereumPaymentObligationContractCallerSession) BalanceOf(_owner common.Address) (*big.Int, error) {
	return _EthereumPaymentObligationContract.Contract.BalanceOf(&_EthereumPaymentObligationContract.CallOpts, _owner)
}

// Exists is a free data retrieval call binding the contract method 0x4f558e79.
//
// Solidity: function exists(_tokenId uint256) constant returns(bool)
func (_EthereumPaymentObligationContract *EthereumPaymentObligationContractCaller) Exists(opts *bind.CallOpts, _tokenId *big.Int) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _EthereumPaymentObligationContract.contract.Call(opts, out, "exists", _tokenId)
	return *ret0, err
}

// Exists is a free data retrieval call binding the contract method 0x4f558e79.
//
// Solidity: function exists(_tokenId uint256) constant returns(bool)
func (_EthereumPaymentObligationContract *EthereumPaymentObligationContractSession) Exists(_tokenId *big.Int) (bool, error) {
	return _EthereumPaymentObligationContract.Contract.Exists(&_EthereumPaymentObligationContract.CallOpts, _tokenId)
}

// Exists is a free data retrieval call binding the contract method 0x4f558e79.
//
// Solidity: function exists(_tokenId uint256) constant returns(bool)
func (_EthereumPaymentObligationContract *EthereumPaymentObligationContractCallerSession) Exists(_tokenId *big.Int) (bool, error) {
	return _EthereumPaymentObligationContract.Contract.Exists(&_EthereumPaymentObligationContract.CallOpts, _tokenId)
}

// GetApproved is a free data retrieval call binding the contract method 0x081812fc.
//
// Solidity: function getApproved(_tokenId uint256) constant returns(address)
func (_EthereumPaymentObligationContract *EthereumPaymentObligationContractCaller) GetApproved(opts *bind.CallOpts, _tokenId *big.Int) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _EthereumPaymentObligationContract.contract.Call(opts, out, "getApproved", _tokenId)
	return *ret0, err
}

// GetApproved is a free data retrieval call binding the contract method 0x081812fc.
//
// Solidity: function getApproved(_tokenId uint256) constant returns(address)
func (_EthereumPaymentObligationContract *EthereumPaymentObligationContractSession) GetApproved(_tokenId *big.Int) (common.Address, error) {
	return _EthereumPaymentObligationContract.Contract.GetApproved(&_EthereumPaymentObligationContract.CallOpts, _tokenId)
}

// GetApproved is a free data retrieval call binding the contract method 0x081812fc.
//
// Solidity: function getApproved(_tokenId uint256) constant returns(address)
func (_EthereumPaymentObligationContract *EthereumPaymentObligationContractCallerSession) GetApproved(_tokenId *big.Int) (common.Address, error) {
	return _EthereumPaymentObligationContract.Contract.GetApproved(&_EthereumPaymentObligationContract.CallOpts, _tokenId)
}

// GetTokenDetails is a free data retrieval call binding the contract method 0xc1e03728.
//
// Solidity: function getTokenDetails(_tokenId uint256) constant returns(grossAmount string, currency string, dueDate string, anchorId uint256, documentRoot bytes32)
func (_EthereumPaymentObligationContract *EthereumPaymentObligationContractCaller) GetTokenDetails(opts *bind.CallOpts, _tokenId *big.Int) (struct {
	GrossAmount  string
	Currency     string
	DueDate      string
	AnchorId     *big.Int
	DocumentRoot [32]byte
}, error) {
	ret := new(struct {
		GrossAmount  string
		Currency     string
		DueDate      string
		AnchorId     *big.Int
		DocumentRoot [32]byte
	})
	out := ret
	err := _EthereumPaymentObligationContract.contract.Call(opts, out, "getTokenDetails", _tokenId)
	return *ret, err
}

// GetTokenDetails is a free data retrieval call binding the contract method 0xc1e03728.
//
// Solidity: function getTokenDetails(_tokenId uint256) constant returns(grossAmount string, currency string, dueDate string, anchorId uint256, documentRoot bytes32)
func (_EthereumPaymentObligationContract *EthereumPaymentObligationContractSession) GetTokenDetails(_tokenId *big.Int) (struct {
	GrossAmount  string
	Currency     string
	DueDate      string
	AnchorId     *big.Int
	DocumentRoot [32]byte
}, error) {
	return _EthereumPaymentObligationContract.Contract.GetTokenDetails(&_EthereumPaymentObligationContract.CallOpts, _tokenId)
}

// GetTokenDetails is a free data retrieval call binding the contract method 0xc1e03728.
//
// Solidity: function getTokenDetails(_tokenId uint256) constant returns(grossAmount string, currency string, dueDate string, anchorId uint256, documentRoot bytes32)
func (_EthereumPaymentObligationContract *EthereumPaymentObligationContractCallerSession) GetTokenDetails(_tokenId *big.Int) (struct {
	GrossAmount  string
	Currency     string
	DueDate      string
	AnchorId     *big.Int
	DocumentRoot [32]byte
}, error) {
	return _EthereumPaymentObligationContract.Contract.GetTokenDetails(&_EthereumPaymentObligationContract.CallOpts, _tokenId)
}

// IsApprovedForAll is a free data retrieval call binding the contract method 0xe985e9c5.
//
// Solidity: function isApprovedForAll(_owner address, _operator address) constant returns(bool)
func (_EthereumPaymentObligationContract *EthereumPaymentObligationContractCaller) IsApprovedForAll(opts *bind.CallOpts, _owner common.Address, _operator common.Address) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _EthereumPaymentObligationContract.contract.Call(opts, out, "isApprovedForAll", _owner, _operator)
	return *ret0, err
}

// IsApprovedForAll is a free data retrieval call binding the contract method 0xe985e9c5.
//
// Solidity: function isApprovedForAll(_owner address, _operator address) constant returns(bool)
func (_EthereumPaymentObligationContract *EthereumPaymentObligationContractSession) IsApprovedForAll(_owner common.Address, _operator common.Address) (bool, error) {
	return _EthereumPaymentObligationContract.Contract.IsApprovedForAll(&_EthereumPaymentObligationContract.CallOpts, _owner, _operator)
}

// IsApprovedForAll is a free data retrieval call binding the contract method 0xe985e9c5.
//
// Solidity: function isApprovedForAll(_owner address, _operator address) constant returns(bool)
func (_EthereumPaymentObligationContract *EthereumPaymentObligationContractCallerSession) IsApprovedForAll(_owner common.Address, _operator common.Address) (bool, error) {
	return _EthereumPaymentObligationContract.Contract.IsApprovedForAll(&_EthereumPaymentObligationContract.CallOpts, _owner, _operator)
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() constant returns(string)
func (_EthereumPaymentObligationContract *EthereumPaymentObligationContractCaller) Name(opts *bind.CallOpts) (string, error) {
	var (
		ret0 = new(string)
	)
	out := ret0
	err := _EthereumPaymentObligationContract.contract.Call(opts, out, "name")
	return *ret0, err
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() constant returns(string)
func (_EthereumPaymentObligationContract *EthereumPaymentObligationContractSession) Name() (string, error) {
	return _EthereumPaymentObligationContract.Contract.Name(&_EthereumPaymentObligationContract.CallOpts)
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() constant returns(string)
func (_EthereumPaymentObligationContract *EthereumPaymentObligationContractCallerSession) Name() (string, error) {
	return _EthereumPaymentObligationContract.Contract.Name(&_EthereumPaymentObligationContract.CallOpts)
}

// OwnerOf is a free data retrieval call binding the contract method 0x6352211e.
//
// Solidity: function ownerOf(_tokenId uint256) constant returns(address)
func (_EthereumPaymentObligationContract *EthereumPaymentObligationContractCaller) OwnerOf(opts *bind.CallOpts, _tokenId *big.Int) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _EthereumPaymentObligationContract.contract.Call(opts, out, "ownerOf", _tokenId)
	return *ret0, err
}

// OwnerOf is a free data retrieval call binding the contract method 0x6352211e.
//
// Solidity: function ownerOf(_tokenId uint256) constant returns(address)
func (_EthereumPaymentObligationContract *EthereumPaymentObligationContractSession) OwnerOf(_tokenId *big.Int) (common.Address, error) {
	return _EthereumPaymentObligationContract.Contract.OwnerOf(&_EthereumPaymentObligationContract.CallOpts, _tokenId)
}

// OwnerOf is a free data retrieval call binding the contract method 0x6352211e.
//
// Solidity: function ownerOf(_tokenId uint256) constant returns(address)
func (_EthereumPaymentObligationContract *EthereumPaymentObligationContractCallerSession) OwnerOf(_tokenId *big.Int) (common.Address, error) {
	return _EthereumPaymentObligationContract.Contract.OwnerOf(&_EthereumPaymentObligationContract.CallOpts, _tokenId)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(_interfaceId bytes4) constant returns(bool)
func (_EthereumPaymentObligationContract *EthereumPaymentObligationContractCaller) SupportsInterface(opts *bind.CallOpts, _interfaceId [4]byte) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _EthereumPaymentObligationContract.contract.Call(opts, out, "supportsInterface", _interfaceId)
	return *ret0, err
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(_interfaceId bytes4) constant returns(bool)
func (_EthereumPaymentObligationContract *EthereumPaymentObligationContractSession) SupportsInterface(_interfaceId [4]byte) (bool, error) {
	return _EthereumPaymentObligationContract.Contract.SupportsInterface(&_EthereumPaymentObligationContract.CallOpts, _interfaceId)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(_interfaceId bytes4) constant returns(bool)
func (_EthereumPaymentObligationContract *EthereumPaymentObligationContractCallerSession) SupportsInterface(_interfaceId [4]byte) (bool, error) {
	return _EthereumPaymentObligationContract.Contract.SupportsInterface(&_EthereumPaymentObligationContract.CallOpts, _interfaceId)
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() constant returns(string)
func (_EthereumPaymentObligationContract *EthereumPaymentObligationContractCaller) Symbol(opts *bind.CallOpts) (string, error) {
	var (
		ret0 = new(string)
	)
	out := ret0
	err := _EthereumPaymentObligationContract.contract.Call(opts, out, "symbol")
	return *ret0, err
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() constant returns(string)
func (_EthereumPaymentObligationContract *EthereumPaymentObligationContractSession) Symbol() (string, error) {
	return _EthereumPaymentObligationContract.Contract.Symbol(&_EthereumPaymentObligationContract.CallOpts)
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() constant returns(string)
func (_EthereumPaymentObligationContract *EthereumPaymentObligationContractCallerSession) Symbol() (string, error) {
	return _EthereumPaymentObligationContract.Contract.Symbol(&_EthereumPaymentObligationContract.CallOpts)
}

// TokenByIndex is a free data retrieval call binding the contract method 0x4f6ccce7.
//
// Solidity: function tokenByIndex(_index uint256) constant returns(uint256)
func (_EthereumPaymentObligationContract *EthereumPaymentObligationContractCaller) TokenByIndex(opts *bind.CallOpts, _index *big.Int) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _EthereumPaymentObligationContract.contract.Call(opts, out, "tokenByIndex", _index)
	return *ret0, err
}

// TokenByIndex is a free data retrieval call binding the contract method 0x4f6ccce7.
//
// Solidity: function tokenByIndex(_index uint256) constant returns(uint256)
func (_EthereumPaymentObligationContract *EthereumPaymentObligationContractSession) TokenByIndex(_index *big.Int) (*big.Int, error) {
	return _EthereumPaymentObligationContract.Contract.TokenByIndex(&_EthereumPaymentObligationContract.CallOpts, _index)
}

// TokenByIndex is a free data retrieval call binding the contract method 0x4f6ccce7.
//
// Solidity: function tokenByIndex(_index uint256) constant returns(uint256)
func (_EthereumPaymentObligationContract *EthereumPaymentObligationContractCallerSession) TokenByIndex(_index *big.Int) (*big.Int, error) {
	return _EthereumPaymentObligationContract.Contract.TokenByIndex(&_EthereumPaymentObligationContract.CallOpts, _index)
}

// TokenOfOwnerByIndex is a free data retrieval call binding the contract method 0x2f745c59.
//
// Solidity: function tokenOfOwnerByIndex(_owner address, _index uint256) constant returns(uint256)
func (_EthereumPaymentObligationContract *EthereumPaymentObligationContractCaller) TokenOfOwnerByIndex(opts *bind.CallOpts, _owner common.Address, _index *big.Int) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _EthereumPaymentObligationContract.contract.Call(opts, out, "tokenOfOwnerByIndex", _owner, _index)
	return *ret0, err
}

// TokenOfOwnerByIndex is a free data retrieval call binding the contract method 0x2f745c59.
//
// Solidity: function tokenOfOwnerByIndex(_owner address, _index uint256) constant returns(uint256)
func (_EthereumPaymentObligationContract *EthereumPaymentObligationContractSession) TokenOfOwnerByIndex(_owner common.Address, _index *big.Int) (*big.Int, error) {
	return _EthereumPaymentObligationContract.Contract.TokenOfOwnerByIndex(&_EthereumPaymentObligationContract.CallOpts, _owner, _index)
}

// TokenOfOwnerByIndex is a free data retrieval call binding the contract method 0x2f745c59.
//
// Solidity: function tokenOfOwnerByIndex(_owner address, _index uint256) constant returns(uint256)
func (_EthereumPaymentObligationContract *EthereumPaymentObligationContractCallerSession) TokenOfOwnerByIndex(_owner common.Address, _index *big.Int) (*big.Int, error) {
	return _EthereumPaymentObligationContract.Contract.TokenOfOwnerByIndex(&_EthereumPaymentObligationContract.CallOpts, _owner, _index)
}

// TokenURI is a free data retrieval call binding the contract method 0xc87b56dd.
//
// Solidity: function tokenURI(_tokenId uint256) constant returns(string)
func (_EthereumPaymentObligationContract *EthereumPaymentObligationContractCaller) TokenURI(opts *bind.CallOpts, _tokenId *big.Int) (string, error) {
	var (
		ret0 = new(string)
	)
	out := ret0
	err := _EthereumPaymentObligationContract.contract.Call(opts, out, "tokenURI", _tokenId)
	return *ret0, err
}

// TokenURI is a free data retrieval call binding the contract method 0xc87b56dd.
//
// Solidity: function tokenURI(_tokenId uint256) constant returns(string)
func (_EthereumPaymentObligationContract *EthereumPaymentObligationContractSession) TokenURI(_tokenId *big.Int) (string, error) {
	return _EthereumPaymentObligationContract.Contract.TokenURI(&_EthereumPaymentObligationContract.CallOpts, _tokenId)
}

// TokenURI is a free data retrieval call binding the contract method 0xc87b56dd.
//
// Solidity: function tokenURI(_tokenId uint256) constant returns(string)
func (_EthereumPaymentObligationContract *EthereumPaymentObligationContractCallerSession) TokenURI(_tokenId *big.Int) (string, error) {
	return _EthereumPaymentObligationContract.Contract.TokenURI(&_EthereumPaymentObligationContract.CallOpts, _tokenId)
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() constant returns(uint256)
func (_EthereumPaymentObligationContract *EthereumPaymentObligationContractCaller) TotalSupply(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _EthereumPaymentObligationContract.contract.Call(opts, out, "totalSupply")
	return *ret0, err
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() constant returns(uint256)
func (_EthereumPaymentObligationContract *EthereumPaymentObligationContractSession) TotalSupply() (*big.Int, error) {
	return _EthereumPaymentObligationContract.Contract.TotalSupply(&_EthereumPaymentObligationContract.CallOpts)
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() constant returns(uint256)
func (_EthereumPaymentObligationContract *EthereumPaymentObligationContractCallerSession) TotalSupply() (*big.Int, error) {
	return _EthereumPaymentObligationContract.Contract.TotalSupply(&_EthereumPaymentObligationContract.CallOpts)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(_to address, _tokenId uint256) returns()
func (_EthereumPaymentObligationContract *EthereumPaymentObligationContractTransactor) Approve(opts *bind.TransactOpts, _to common.Address, _tokenId *big.Int) (*types.Transaction, error) {
	return _EthereumPaymentObligationContract.contract.Transact(opts, "approve", _to, _tokenId)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(_to address, _tokenId uint256) returns()
func (_EthereumPaymentObligationContract *EthereumPaymentObligationContractSession) Approve(_to common.Address, _tokenId *big.Int) (*types.Transaction, error) {
	return _EthereumPaymentObligationContract.Contract.Approve(&_EthereumPaymentObligationContract.TransactOpts, _to, _tokenId)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(_to address, _tokenId uint256) returns()
func (_EthereumPaymentObligationContract *EthereumPaymentObligationContractTransactorSession) Approve(_to common.Address, _tokenId *big.Int) (*types.Transaction, error) {
	return _EthereumPaymentObligationContract.Contract.Approve(&_EthereumPaymentObligationContract.TransactOpts, _to, _tokenId)
}

// Mint is a paid mutator transaction binding the contract method 0xcb40425f.
//
// Solidity: function mint(_to address, _tokenId uint256, _tokenURI string, _anchorId uint256, _merkleRoot bytes32, _collaboratorField string, _values string[5], _salts bytes32[5], _proofs bytes32[][5]) returns()
func (_EthereumPaymentObligationContract *EthereumPaymentObligationContractTransactor) Mint(opts *bind.TransactOpts, _to common.Address, _tokenId *big.Int, _tokenURI string, _anchorId *big.Int, _merkleRoot [32]byte, _collaboratorField string, _values [5]string, _salts [5][32]byte, _proofs [5][][32]byte) (*types.Transaction, error) {
	return _EthereumPaymentObligationContract.contract.Transact(opts, "mint", _to, _tokenId, _tokenURI, _anchorId, _merkleRoot, _collaboratorField, _values, _salts, _proofs)
}

// Mint is a paid mutator transaction binding the contract method 0xcb40425f.
//
// Solidity: function mint(_to address, _tokenId uint256, _tokenURI string, _anchorId uint256, _merkleRoot bytes32, _collaboratorField string, _values string[5], _salts bytes32[5], _proofs bytes32[][5]) returns()
func (_EthereumPaymentObligationContract *EthereumPaymentObligationContractSession) Mint(_to common.Address, _tokenId *big.Int, _tokenURI string, _anchorId *big.Int, _merkleRoot [32]byte, _collaboratorField string, _values [5]string, _salts [5][32]byte, _proofs [5][][32]byte) (*types.Transaction, error) {
	return _EthereumPaymentObligationContract.Contract.Mint(&_EthereumPaymentObligationContract.TransactOpts, _to, _tokenId, _tokenURI, _anchorId, _merkleRoot, _collaboratorField, _values, _salts, _proofs)
}

// Mint is a paid mutator transaction binding the contract method 0xcb40425f.
//
// Solidity: function mint(_to address, _tokenId uint256, _tokenURI string, _anchorId uint256, _merkleRoot bytes32, _collaboratorField string, _values string[5], _salts bytes32[5], _proofs bytes32[][5]) returns()
func (_EthereumPaymentObligationContract *EthereumPaymentObligationContractTransactorSession) Mint(_to common.Address, _tokenId *big.Int, _tokenURI string, _anchorId *big.Int, _merkleRoot [32]byte, _collaboratorField string, _values [5]string, _salts [5][32]byte, _proofs [5][][32]byte) (*types.Transaction, error) {
	return _EthereumPaymentObligationContract.Contract.Mint(&_EthereumPaymentObligationContract.TransactOpts, _to, _tokenId, _tokenURI, _anchorId, _merkleRoot, _collaboratorField, _values, _salts, _proofs)
}

// SafeTransferFrom is a paid mutator transaction binding the contract method 0xb88d4fde.
//
// Solidity: function safeTransferFrom(_from address, _to address, _tokenId uint256, _data bytes) returns()
func (_EthereumPaymentObligationContract *EthereumPaymentObligationContractTransactor) SafeTransferFrom(opts *bind.TransactOpts, _from common.Address, _to common.Address, _tokenId *big.Int, _data []byte) (*types.Transaction, error) {
	return _EthereumPaymentObligationContract.contract.Transact(opts, "safeTransferFrom", _from, _to, _tokenId, _data)
}

// SafeTransferFrom is a paid mutator transaction binding the contract method 0xb88d4fde.
//
// Solidity: function safeTransferFrom(_from address, _to address, _tokenId uint256, _data bytes) returns()
func (_EthereumPaymentObligationContract *EthereumPaymentObligationContractSession) SafeTransferFrom(_from common.Address, _to common.Address, _tokenId *big.Int, _data []byte) (*types.Transaction, error) {
	return _EthereumPaymentObligationContract.Contract.SafeTransferFrom(&_EthereumPaymentObligationContract.TransactOpts, _from, _to, _tokenId, _data)
}

// SafeTransferFrom is a paid mutator transaction binding the contract method 0xb88d4fde.
//
// Solidity: function safeTransferFrom(_from address, _to address, _tokenId uint256, _data bytes) returns()
func (_EthereumPaymentObligationContract *EthereumPaymentObligationContractTransactorSession) SafeTransferFrom(_from common.Address, _to common.Address, _tokenId *big.Int, _data []byte) (*types.Transaction, error) {
	return _EthereumPaymentObligationContract.Contract.SafeTransferFrom(&_EthereumPaymentObligationContract.TransactOpts, _from, _to, _tokenId, _data)
}

// SetApprovalForAll is a paid mutator transaction binding the contract method 0xa22cb465.
//
// Solidity: function setApprovalForAll(_to address, _approved bool) returns()
func (_EthereumPaymentObligationContract *EthereumPaymentObligationContractTransactor) SetApprovalForAll(opts *bind.TransactOpts, _to common.Address, _approved bool) (*types.Transaction, error) {
	return _EthereumPaymentObligationContract.contract.Transact(opts, "setApprovalForAll", _to, _approved)
}

// SetApprovalForAll is a paid mutator transaction binding the contract method 0xa22cb465.
//
// Solidity: function setApprovalForAll(_to address, _approved bool) returns()
func (_EthereumPaymentObligationContract *EthereumPaymentObligationContractSession) SetApprovalForAll(_to common.Address, _approved bool) (*types.Transaction, error) {
	return _EthereumPaymentObligationContract.Contract.SetApprovalForAll(&_EthereumPaymentObligationContract.TransactOpts, _to, _approved)
}

// SetApprovalForAll is a paid mutator transaction binding the contract method 0xa22cb465.
//
// Solidity: function setApprovalForAll(_to address, _approved bool) returns()
func (_EthereumPaymentObligationContract *EthereumPaymentObligationContractTransactorSession) SetApprovalForAll(_to common.Address, _approved bool) (*types.Transaction, error) {
	return _EthereumPaymentObligationContract.Contract.SetApprovalForAll(&_EthereumPaymentObligationContract.TransactOpts, _to, _approved)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(_from address, _to address, _tokenId uint256) returns()
func (_EthereumPaymentObligationContract *EthereumPaymentObligationContractTransactor) TransferFrom(opts *bind.TransactOpts, _from common.Address, _to common.Address, _tokenId *big.Int) (*types.Transaction, error) {
	return _EthereumPaymentObligationContract.contract.Transact(opts, "transferFrom", _from, _to, _tokenId)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(_from address, _to address, _tokenId uint256) returns()
func (_EthereumPaymentObligationContract *EthereumPaymentObligationContractSession) TransferFrom(_from common.Address, _to common.Address, _tokenId *big.Int) (*types.Transaction, error) {
	return _EthereumPaymentObligationContract.Contract.TransferFrom(&_EthereumPaymentObligationContract.TransactOpts, _from, _to, _tokenId)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(_from address, _to address, _tokenId uint256) returns()
func (_EthereumPaymentObligationContract *EthereumPaymentObligationContractTransactorSession) TransferFrom(_from common.Address, _to common.Address, _tokenId *big.Int) (*types.Transaction, error) {
	return _EthereumPaymentObligationContract.Contract.TransferFrom(&_EthereumPaymentObligationContract.TransactOpts, _from, _to, _tokenId)
}

// EthereumPaymentObligationContractApprovalIterator is returned from FilterApproval and is used to iterate over the raw logs and unpacked data for Approval events raised by the EthereumPaymentObligationContract contract.
type EthereumPaymentObligationContractApprovalIterator struct {
	Event *EthereumPaymentObligationContractApproval // Event containing the contract specifics and raw log

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
func (it *EthereumPaymentObligationContractApprovalIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(EthereumPaymentObligationContractApproval)
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
		it.Event = new(EthereumPaymentObligationContractApproval)
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
func (it *EthereumPaymentObligationContractApprovalIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *EthereumPaymentObligationContractApprovalIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// EthereumPaymentObligationContractApproval represents a Approval event raised by the EthereumPaymentObligationContract contract.
type EthereumPaymentObligationContractApproval struct {
	Owner    common.Address
	Approved common.Address
	TokenId  *big.Int
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterApproval is a free log retrieval operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: e Approval(_owner indexed address, _approved indexed address, _tokenId indexed uint256)
func (_EthereumPaymentObligationContract *EthereumPaymentObligationContractFilterer) FilterApproval(opts *bind.FilterOpts, _owner []common.Address, _approved []common.Address, _tokenId []*big.Int) (*EthereumPaymentObligationContractApprovalIterator, error) {

	var _ownerRule []interface{}
	for _, _ownerItem := range _owner {
		_ownerRule = append(_ownerRule, _ownerItem)
	}
	var _approvedRule []interface{}
	for _, _approvedItem := range _approved {
		_approvedRule = append(_approvedRule, _approvedItem)
	}
	var _tokenIdRule []interface{}
	for _, _tokenIdItem := range _tokenId {
		_tokenIdRule = append(_tokenIdRule, _tokenIdItem)
	}

	logs, sub, err := _EthereumPaymentObligationContract.contract.FilterLogs(opts, "Approval", _ownerRule, _approvedRule, _tokenIdRule)
	if err != nil {
		return nil, err
	}
	return &EthereumPaymentObligationContractApprovalIterator{contract: _EthereumPaymentObligationContract.contract, event: "Approval", logs: logs, sub: sub}, nil
}

// WatchApproval is a free log subscription operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: e Approval(_owner indexed address, _approved indexed address, _tokenId indexed uint256)
func (_EthereumPaymentObligationContract *EthereumPaymentObligationContractFilterer) WatchApproval(opts *bind.WatchOpts, sink chan<- *EthereumPaymentObligationContractApproval, _owner []common.Address, _approved []common.Address, _tokenId []*big.Int) (event.Subscription, error) {

	var _ownerRule []interface{}
	for _, _ownerItem := range _owner {
		_ownerRule = append(_ownerRule, _ownerItem)
	}
	var _approvedRule []interface{}
	for _, _approvedItem := range _approved {
		_approvedRule = append(_approvedRule, _approvedItem)
	}
	var _tokenIdRule []interface{}
	for _, _tokenIdItem := range _tokenId {
		_tokenIdRule = append(_tokenIdRule, _tokenIdItem)
	}

	logs, sub, err := _EthereumPaymentObligationContract.contract.WatchLogs(opts, "Approval", _ownerRule, _approvedRule, _tokenIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(EthereumPaymentObligationContractApproval)
				if err := _EthereumPaymentObligationContract.contract.UnpackLog(event, "Approval", log); err != nil {
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

// EthereumPaymentObligationContractApprovalForAllIterator is returned from FilterApprovalForAll and is used to iterate over the raw logs and unpacked data for ApprovalForAll events raised by the EthereumPaymentObligationContract contract.
type EthereumPaymentObligationContractApprovalForAllIterator struct {
	Event *EthereumPaymentObligationContractApprovalForAll // Event containing the contract specifics and raw log

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
func (it *EthereumPaymentObligationContractApprovalForAllIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(EthereumPaymentObligationContractApprovalForAll)
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
		it.Event = new(EthereumPaymentObligationContractApprovalForAll)
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
func (it *EthereumPaymentObligationContractApprovalForAllIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *EthereumPaymentObligationContractApprovalForAllIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// EthereumPaymentObligationContractApprovalForAll represents a ApprovalForAll event raised by the EthereumPaymentObligationContract contract.
type EthereumPaymentObligationContractApprovalForAll struct {
	Owner    common.Address
	Operator common.Address
	Approved bool
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterApprovalForAll is a free log retrieval operation binding the contract event 0x17307eab39ab6107e8899845ad3d59bd9653f200f220920489ca2b5937696c31.
//
// Solidity: e ApprovalForAll(_owner indexed address, _operator indexed address, _approved bool)
func (_EthereumPaymentObligationContract *EthereumPaymentObligationContractFilterer) FilterApprovalForAll(opts *bind.FilterOpts, _owner []common.Address, _operator []common.Address) (*EthereumPaymentObligationContractApprovalForAllIterator, error) {

	var _ownerRule []interface{}
	for _, _ownerItem := range _owner {
		_ownerRule = append(_ownerRule, _ownerItem)
	}
	var _operatorRule []interface{}
	for _, _operatorItem := range _operator {
		_operatorRule = append(_operatorRule, _operatorItem)
	}

	logs, sub, err := _EthereumPaymentObligationContract.contract.FilterLogs(opts, "ApprovalForAll", _ownerRule, _operatorRule)
	if err != nil {
		return nil, err
	}
	return &EthereumPaymentObligationContractApprovalForAllIterator{contract: _EthereumPaymentObligationContract.contract, event: "ApprovalForAll", logs: logs, sub: sub}, nil
}

// WatchApprovalForAll is a free log subscription operation binding the contract event 0x17307eab39ab6107e8899845ad3d59bd9653f200f220920489ca2b5937696c31.
//
// Solidity: e ApprovalForAll(_owner indexed address, _operator indexed address, _approved bool)
func (_EthereumPaymentObligationContract *EthereumPaymentObligationContractFilterer) WatchApprovalForAll(opts *bind.WatchOpts, sink chan<- *EthereumPaymentObligationContractApprovalForAll, _owner []common.Address, _operator []common.Address) (event.Subscription, error) {

	var _ownerRule []interface{}
	for _, _ownerItem := range _owner {
		_ownerRule = append(_ownerRule, _ownerItem)
	}
	var _operatorRule []interface{}
	for _, _operatorItem := range _operator {
		_operatorRule = append(_operatorRule, _operatorItem)
	}

	logs, sub, err := _EthereumPaymentObligationContract.contract.WatchLogs(opts, "ApprovalForAll", _ownerRule, _operatorRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(EthereumPaymentObligationContractApprovalForAll)
				if err := _EthereumPaymentObligationContract.contract.UnpackLog(event, "ApprovalForAll", log); err != nil {
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

// EthereumPaymentObligationContractPaymentObligationMintedIterator is returned from FilterPaymentObligationMinted and is used to iterate over the raw logs and unpacked data for PaymentObligationMinted events raised by the EthereumPaymentObligationContract contract.
type EthereumPaymentObligationContractPaymentObligationMintedIterator struct {
	Event *EthereumPaymentObligationContractPaymentObligationMinted // Event containing the contract specifics and raw log

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
func (it *EthereumPaymentObligationContractPaymentObligationMintedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(EthereumPaymentObligationContractPaymentObligationMinted)
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
		it.Event = new(EthereumPaymentObligationContractPaymentObligationMinted)
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
func (it *EthereumPaymentObligationContractPaymentObligationMintedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *EthereumPaymentObligationContractPaymentObligationMintedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// EthereumPaymentObligationContractPaymentObligationMinted represents a PaymentObligationMinted event raised by the EthereumPaymentObligationContract contract.
type EthereumPaymentObligationContractPaymentObligationMinted struct {
	To       common.Address
	TokenId  *big.Int
	TokenURI string
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterPaymentObligationMinted is a free log retrieval operation binding the contract event 0xe2e4e975c4de5fbb1416db2c5ff8e2f4108bbfcdfd27a1a1eb03935cbfa3b8f9.
//
// Solidity: e PaymentObligationMinted(to address, tokenId uint256, tokenURI string)
func (_EthereumPaymentObligationContract *EthereumPaymentObligationContractFilterer) FilterPaymentObligationMinted(opts *bind.FilterOpts) (*EthereumPaymentObligationContractPaymentObligationMintedIterator, error) {

	logs, sub, err := _EthereumPaymentObligationContract.contract.FilterLogs(opts, "PaymentObligationMinted")
	if err != nil {
		return nil, err
	}
	return &EthereumPaymentObligationContractPaymentObligationMintedIterator{contract: _EthereumPaymentObligationContract.contract, event: "PaymentObligationMinted", logs: logs, sub: sub}, nil
}

// WatchPaymentObligationMinted is a free log subscription operation binding the contract event 0xe2e4e975c4de5fbb1416db2c5ff8e2f4108bbfcdfd27a1a1eb03935cbfa3b8f9.
//
// Solidity: e PaymentObligationMinted(to address, tokenId uint256, tokenURI string)
func (_EthereumPaymentObligationContract *EthereumPaymentObligationContractFilterer) WatchPaymentObligationMinted(opts *bind.WatchOpts, sink chan<- *EthereumPaymentObligationContractPaymentObligationMinted) (event.Subscription, error) {

	logs, sub, err := _EthereumPaymentObligationContract.contract.WatchLogs(opts, "PaymentObligationMinted")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(EthereumPaymentObligationContractPaymentObligationMinted)
				if err := _EthereumPaymentObligationContract.contract.UnpackLog(event, "PaymentObligationMinted", log); err != nil {
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

// EthereumPaymentObligationContractTransferIterator is returned from FilterTransfer and is used to iterate over the raw logs and unpacked data for Transfer events raised by the EthereumPaymentObligationContract contract.
type EthereumPaymentObligationContractTransferIterator struct {
	Event *EthereumPaymentObligationContractTransfer // Event containing the contract specifics and raw log

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
func (it *EthereumPaymentObligationContractTransferIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(EthereumPaymentObligationContractTransfer)
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
		it.Event = new(EthereumPaymentObligationContractTransfer)
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
func (it *EthereumPaymentObligationContractTransferIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *EthereumPaymentObligationContractTransferIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// EthereumPaymentObligationContractTransfer represents a Transfer event raised by the EthereumPaymentObligationContract contract.
type EthereumPaymentObligationContractTransfer struct {
	From    common.Address
	To      common.Address
	TokenId *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterTransfer is a free log retrieval operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: e Transfer(_from indexed address, _to indexed address, _tokenId indexed uint256)
func (_EthereumPaymentObligationContract *EthereumPaymentObligationContractFilterer) FilterTransfer(opts *bind.FilterOpts, _from []common.Address, _to []common.Address, _tokenId []*big.Int) (*EthereumPaymentObligationContractTransferIterator, error) {

	var _fromRule []interface{}
	for _, _fromItem := range _from {
		_fromRule = append(_fromRule, _fromItem)
	}
	var _toRule []interface{}
	for _, _toItem := range _to {
		_toRule = append(_toRule, _toItem)
	}
	var _tokenIdRule []interface{}
	for _, _tokenIdItem := range _tokenId {
		_tokenIdRule = append(_tokenIdRule, _tokenIdItem)
	}

	logs, sub, err := _EthereumPaymentObligationContract.contract.FilterLogs(opts, "Transfer", _fromRule, _toRule, _tokenIdRule)
	if err != nil {
		return nil, err
	}
	return &EthereumPaymentObligationContractTransferIterator{contract: _EthereumPaymentObligationContract.contract, event: "Transfer", logs: logs, sub: sub}, nil
}

// WatchTransfer is a free log subscription operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: e Transfer(_from indexed address, _to indexed address, _tokenId indexed uint256)
func (_EthereumPaymentObligationContract *EthereumPaymentObligationContractFilterer) WatchTransfer(opts *bind.WatchOpts, sink chan<- *EthereumPaymentObligationContractTransfer, _from []common.Address, _to []common.Address, _tokenId []*big.Int) (event.Subscription, error) {

	var _fromRule []interface{}
	for _, _fromItem := range _from {
		_fromRule = append(_fromRule, _fromItem)
	}
	var _toRule []interface{}
	for _, _toItem := range _to {
		_toRule = append(_toRule, _toItem)
	}
	var _tokenIdRule []interface{}
	for _, _tokenIdItem := range _tokenId {
		_tokenIdRule = append(_tokenIdRule, _tokenIdItem)
	}

	logs, sub, err := _EthereumPaymentObligationContract.contract.WatchLogs(opts, "Transfer", _fromRule, _toRule, _tokenIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(EthereumPaymentObligationContractTransfer)
				if err := _EthereumPaymentObligationContract.contract.UnpackLog(event, "Transfer", log); err != nil {
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
