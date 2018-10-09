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

// PaymentObligationABI is the input ABI used to generate the binding from.
const PaymentObligationABI = "[{\"constant\":true,\"inputs\":[{\"name\":\"_interfaceId\",\"type\":\"bytes4\"}],\"name\":\"supportsInterface\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"name\",\"outputs\":[{\"name\":\"\",\"type\":\"string\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_tokenId\",\"type\":\"uint256\"}],\"name\":\"getApproved\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_to\",\"type\":\"address\"},{\"name\":\"_tokenId\",\"type\":\"uint256\"}],\"name\":\"approve\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"totalSupply\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"InterfaceId_ERC165\",\"outputs\":[{\"name\":\"\",\"type\":\"bytes4\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_from\",\"type\":\"address\"},{\"name\":\"_to\",\"type\":\"address\"},{\"name\":\"_tokenId\",\"type\":\"uint256\"}],\"name\":\"transferFrom\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_owner\",\"type\":\"address\"},{\"name\":\"_index\",\"type\":\"uint256\"}],\"name\":\"tokenOfOwnerByIndex\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_from\",\"type\":\"address\"},{\"name\":\"_to\",\"type\":\"address\"},{\"name\":\"_tokenId\",\"type\":\"uint256\"}],\"name\":\"safeTransferFrom\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_tokenId\",\"type\":\"uint256\"}],\"name\":\"exists\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_index\",\"type\":\"uint256\"}],\"name\":\"tokenByIndex\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"anchorRegistry\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_tokenId\",\"type\":\"uint256\"}],\"name\":\"ownerOf\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_owner\",\"type\":\"address\"}],\"name\":\"balanceOf\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"symbol\",\"outputs\":[{\"name\":\"\",\"type\":\"string\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_to\",\"type\":\"address\"},{\"name\":\"_approved\",\"type\":\"bool\"}],\"name\":\"setApprovalForAll\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_from\",\"type\":\"address\"},{\"name\":\"_to\",\"type\":\"address\"},{\"name\":\"_tokenId\",\"type\":\"uint256\"},{\"name\":\"_data\",\"type\":\"bytes\"}],\"name\":\"safeTransferFrom\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_tokenId\",\"type\":\"uint256\"}],\"name\":\"tokenURI\",\"outputs\":[{\"name\":\"\",\"type\":\"string\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_owner\",\"type\":\"address\"},{\"name\":\"_operator\",\"type\":\"address\"}],\"name\":\"isApprovedForAll\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"name\":\"_name\",\"type\":\"string\"},{\"name\":\"_symbol\",\"type\":\"string\"},{\"name\":\"_anchorRegistry\",\"type\":\"address\"},{\"name\":\"_identityRegistry\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"tokenId\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"tokenURI\",\"type\":\"string\"}],\"name\":\"PaymentObligationMinted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"_from\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"_to\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"_tokenId\",\"type\":\"uint256\"}],\"name\":\"Transfer\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"_owner\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"_approved\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"_tokenId\",\"type\":\"uint256\"}],\"name\":\"Approval\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"_owner\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"_operator\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"_approved\",\"type\":\"bool\"}],\"name\":\"ApprovalForAll\",\"type\":\"event\"},{\"constant\":false,\"inputs\":[{\"name\":\"_to\",\"type\":\"address\"},{\"name\":\"_tokenId\",\"type\":\"uint256\"},{\"name\":\"_tokenURI\",\"type\":\"string\"},{\"name\":\"_anchorId\",\"type\":\"uint256\"},{\"name\":\"_merkleRoot\",\"type\":\"bytes32\"},{\"name\":\"_values\",\"type\":\"string[3]\"},{\"name\":\"_salts\",\"type\":\"bytes32[3]\"},{\"name\":\"_proofs\",\"type\":\"bytes32[][3]\"}],\"name\":\"mint\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_tokenId\",\"type\":\"uint256\"}],\"name\":\"getTokenDetails\",\"outputs\":[{\"name\":\"grossAmount\",\"type\":\"string\"},{\"name\":\"currency\",\"type\":\"string\"},{\"name\":\"dueDate\",\"type\":\"string\"},{\"name\":\"anchorId\",\"type\":\"uint256\"},{\"name\":\"documentRoot\",\"type\":\"bytes32\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"}]"

// PaymentObligation is an auto generated Go binding around an Ethereum contract.
type PaymentObligation struct {
	PaymentObligationCaller     // Read-only binding to the contract
	PaymentObligationTransactor // Write-only binding to the contract
	PaymentObligationFilterer   // Log filterer for contract events
}

// PaymentObligationCaller is an auto generated read-only Go binding around an Ethereum contract.
type PaymentObligationCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// PaymentObligationTransactor is an auto generated write-only Go binding around an Ethereum contract.
type PaymentObligationTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// PaymentObligationFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type PaymentObligationFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// PaymentObligationSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type PaymentObligationSession struct {
	Contract     *PaymentObligation // Generic contract binding to set the session for
	CallOpts     bind.CallOpts      // Call options to use throughout this session
	TransactOpts bind.TransactOpts  // Transaction auth options to use throughout this session
}

// PaymentObligationCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type PaymentObligationCallerSession struct {
	Contract *PaymentObligationCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts            // Call options to use throughout this session
}

// PaymentObligationTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type PaymentObligationTransactorSession struct {
	Contract     *PaymentObligationTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts            // Transaction auth options to use throughout this session
}

// PaymentObligationRaw is an auto generated low-level Go binding around an Ethereum contract.
type PaymentObligationRaw struct {
	Contract *PaymentObligation // Generic contract binding to access the raw methods on
}

// PaymentObligationCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type PaymentObligationCallerRaw struct {
	Contract *PaymentObligationCaller // Generic read-only contract binding to access the raw methods on
}

// PaymentObligationTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type PaymentObligationTransactorRaw struct {
	Contract *PaymentObligationTransactor // Generic write-only contract binding to access the raw methods on
}

// NewPaymentObligation creates a new instance of PaymentObligation, bound to a specific deployed contract.
func NewPaymentObligation(address common.Address, backend bind.ContractBackend) (*PaymentObligation, error) {
	contract, err := bindPaymentObligation(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &PaymentObligation{PaymentObligationCaller: PaymentObligationCaller{contract: contract}, PaymentObligationTransactor: PaymentObligationTransactor{contract: contract}, PaymentObligationFilterer: PaymentObligationFilterer{contract: contract}}, nil
}

// NewPaymentObligationCaller creates a new read-only instance of PaymentObligation, bound to a specific deployed contract.
func NewPaymentObligationCaller(address common.Address, caller bind.ContractCaller) (*PaymentObligationCaller, error) {
	contract, err := bindPaymentObligation(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &PaymentObligationCaller{contract: contract}, nil
}

// NewPaymentObligationTransactor creates a new write-only instance of PaymentObligation, bound to a specific deployed contract.
func NewPaymentObligationTransactor(address common.Address, transactor bind.ContractTransactor) (*PaymentObligationTransactor, error) {
	contract, err := bindPaymentObligation(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &PaymentObligationTransactor{contract: contract}, nil
}

// NewPaymentObligationFilterer creates a new log filterer instance of PaymentObligation, bound to a specific deployed contract.
func NewPaymentObligationFilterer(address common.Address, filterer bind.ContractFilterer) (*PaymentObligationFilterer, error) {
	contract, err := bindPaymentObligation(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &PaymentObligationFilterer{contract: contract}, nil
}

// bindPaymentObligation binds a generic wrapper to an already deployed contract.
func bindPaymentObligation(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(PaymentObligationABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_PaymentObligation *PaymentObligationRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _PaymentObligation.Contract.PaymentObligationCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_PaymentObligation *PaymentObligationRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _PaymentObligation.Contract.PaymentObligationTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_PaymentObligation *PaymentObligationRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _PaymentObligation.Contract.PaymentObligationTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_PaymentObligation *PaymentObligationCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _PaymentObligation.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_PaymentObligation *PaymentObligationTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _PaymentObligation.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_PaymentObligation *PaymentObligationTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _PaymentObligation.Contract.contract.Transact(opts, method, params...)
}

// InterfaceIdERC165 is a free data retrieval call binding the contract method 0x19fa8f50.
//
// Solidity: function InterfaceId_ERC165() constant returns(bytes4)
func (_PaymentObligation *PaymentObligationCaller) InterfaceIdERC165(opts *bind.CallOpts) ([4]byte, error) {
	var (
		ret0 = new([4]byte)
	)
	out := ret0
	err := _PaymentObligation.contract.Call(opts, out, "InterfaceId_ERC165")
	return *ret0, err
}

// InterfaceIdERC165 is a free data retrieval call binding the contract method 0x19fa8f50.
//
// Solidity: function InterfaceId_ERC165() constant returns(bytes4)
func (_PaymentObligation *PaymentObligationSession) InterfaceIdERC165() ([4]byte, error) {
	return _PaymentObligation.Contract.InterfaceIdERC165(&_PaymentObligation.CallOpts)
}

// InterfaceIdERC165 is a free data retrieval call binding the contract method 0x19fa8f50.
//
// Solidity: function InterfaceId_ERC165() constant returns(bytes4)
func (_PaymentObligation *PaymentObligationCallerSession) InterfaceIdERC165() ([4]byte, error) {
	return _PaymentObligation.Contract.InterfaceIdERC165(&_PaymentObligation.CallOpts)
}

// AnchorRegistry is a free data retrieval call binding the contract method 0x5a180c0a.
//
// Solidity: function anchorRegistry() constant returns(address)
func (_PaymentObligation *PaymentObligationCaller) AnchorRegistry(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _PaymentObligation.contract.Call(opts, out, "anchorRegistry")
	return *ret0, err
}

// AnchorRegistry is a free data retrieval call binding the contract method 0x5a180c0a.
//
// Solidity: function anchorRegistry() constant returns(address)
func (_PaymentObligation *PaymentObligationSession) AnchorRegistry() (common.Address, error) {
	return _PaymentObligation.Contract.AnchorRegistry(&_PaymentObligation.CallOpts)
}

// AnchorRegistry is a free data retrieval call binding the contract method 0x5a180c0a.
//
// Solidity: function anchorRegistry() constant returns(address)
func (_PaymentObligation *PaymentObligationCallerSession) AnchorRegistry() (common.Address, error) {
	return _PaymentObligation.Contract.AnchorRegistry(&_PaymentObligation.CallOpts)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(_owner address) constant returns(uint256)
func (_PaymentObligation *PaymentObligationCaller) BalanceOf(opts *bind.CallOpts, _owner common.Address) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _PaymentObligation.contract.Call(opts, out, "balanceOf", _owner)
	return *ret0, err
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(_owner address) constant returns(uint256)
func (_PaymentObligation *PaymentObligationSession) BalanceOf(_owner common.Address) (*big.Int, error) {
	return _PaymentObligation.Contract.BalanceOf(&_PaymentObligation.CallOpts, _owner)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(_owner address) constant returns(uint256)
func (_PaymentObligation *PaymentObligationCallerSession) BalanceOf(_owner common.Address) (*big.Int, error) {
	return _PaymentObligation.Contract.BalanceOf(&_PaymentObligation.CallOpts, _owner)
}

// Exists is a free data retrieval call binding the contract method 0x4f558e79.
//
// Solidity: function exists(_tokenId uint256) constant returns(bool)
func (_PaymentObligation *PaymentObligationCaller) Exists(opts *bind.CallOpts, _tokenId *big.Int) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _PaymentObligation.contract.Call(opts, out, "exists", _tokenId)
	return *ret0, err
}

// Exists is a free data retrieval call binding the contract method 0x4f558e79.
//
// Solidity: function exists(_tokenId uint256) constant returns(bool)
func (_PaymentObligation *PaymentObligationSession) Exists(_tokenId *big.Int) (bool, error) {
	return _PaymentObligation.Contract.Exists(&_PaymentObligation.CallOpts, _tokenId)
}

// Exists is a free data retrieval call binding the contract method 0x4f558e79.
//
// Solidity: function exists(_tokenId uint256) constant returns(bool)
func (_PaymentObligation *PaymentObligationCallerSession) Exists(_tokenId *big.Int) (bool, error) {
	return _PaymentObligation.Contract.Exists(&_PaymentObligation.CallOpts, _tokenId)
}

// GetApproved is a free data retrieval call binding the contract method 0x081812fc.
//
// Solidity: function getApproved(_tokenId uint256) constant returns(address)
func (_PaymentObligation *PaymentObligationCaller) GetApproved(opts *bind.CallOpts, _tokenId *big.Int) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _PaymentObligation.contract.Call(opts, out, "getApproved", _tokenId)
	return *ret0, err
}

// GetApproved is a free data retrieval call binding the contract method 0x081812fc.
//
// Solidity: function getApproved(_tokenId uint256) constant returns(address)
func (_PaymentObligation *PaymentObligationSession) GetApproved(_tokenId *big.Int) (common.Address, error) {
	return _PaymentObligation.Contract.GetApproved(&_PaymentObligation.CallOpts, _tokenId)
}

// GetApproved is a free data retrieval call binding the contract method 0x081812fc.
//
// Solidity: function getApproved(_tokenId uint256) constant returns(address)
func (_PaymentObligation *PaymentObligationCallerSession) GetApproved(_tokenId *big.Int) (common.Address, error) {
	return _PaymentObligation.Contract.GetApproved(&_PaymentObligation.CallOpts, _tokenId)
}

// GetTokenDetails is a free data retrieval call binding the contract method 0xc1e03728.
//
// Solidity: function getTokenDetails(_tokenId uint256) constant returns(grossAmount string, currency string, dueDate string, anchorId uint256, documentRoot bytes32)
func (_PaymentObligation *PaymentObligationCaller) GetTokenDetails(opts *bind.CallOpts, _tokenId *big.Int) (struct {
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
	err := _PaymentObligation.contract.Call(opts, out, "getTokenDetails", _tokenId)
	return *ret, err
}

// GetTokenDetails is a free data retrieval call binding the contract method 0xc1e03728.
//
// Solidity: function getTokenDetails(_tokenId uint256) constant returns(grossAmount string, currency string, dueDate string, anchorId uint256, documentRoot bytes32)
func (_PaymentObligation *PaymentObligationSession) GetTokenDetails(_tokenId *big.Int) (struct {
	GrossAmount  string
	Currency     string
	DueDate      string
	AnchorId     *big.Int
	DocumentRoot [32]byte
}, error) {
	return _PaymentObligation.Contract.GetTokenDetails(&_PaymentObligation.CallOpts, _tokenId)
}

// GetTokenDetails is a free data retrieval call binding the contract method 0xc1e03728.
//
// Solidity: function getTokenDetails(_tokenId uint256) constant returns(grossAmount string, currency string, dueDate string, anchorId uint256, documentRoot bytes32)
func (_PaymentObligation *PaymentObligationCallerSession) GetTokenDetails(_tokenId *big.Int) (struct {
	GrossAmount  string
	Currency     string
	DueDate      string
	AnchorId     *big.Int
	DocumentRoot [32]byte
}, error) {
	return _PaymentObligation.Contract.GetTokenDetails(&_PaymentObligation.CallOpts, _tokenId)
}

// IsApprovedForAll is a free data retrieval call binding the contract method 0xe985e9c5.
//
// Solidity: function isApprovedForAll(_owner address, _operator address) constant returns(bool)
func (_PaymentObligation *PaymentObligationCaller) IsApprovedForAll(opts *bind.CallOpts, _owner common.Address, _operator common.Address) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _PaymentObligation.contract.Call(opts, out, "isApprovedForAll", _owner, _operator)
	return *ret0, err
}

// IsApprovedForAll is a free data retrieval call binding the contract method 0xe985e9c5.
//
// Solidity: function isApprovedForAll(_owner address, _operator address) constant returns(bool)
func (_PaymentObligation *PaymentObligationSession) IsApprovedForAll(_owner common.Address, _operator common.Address) (bool, error) {
	return _PaymentObligation.Contract.IsApprovedForAll(&_PaymentObligation.CallOpts, _owner, _operator)
}

// IsApprovedForAll is a free data retrieval call binding the contract method 0xe985e9c5.
//
// Solidity: function isApprovedForAll(_owner address, _operator address) constant returns(bool)
func (_PaymentObligation *PaymentObligationCallerSession) IsApprovedForAll(_owner common.Address, _operator common.Address) (bool, error) {
	return _PaymentObligation.Contract.IsApprovedForAll(&_PaymentObligation.CallOpts, _owner, _operator)
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() constant returns(string)
func (_PaymentObligation *PaymentObligationCaller) Name(opts *bind.CallOpts) (string, error) {
	var (
		ret0 = new(string)
	)
	out := ret0
	err := _PaymentObligation.contract.Call(opts, out, "name")
	return *ret0, err
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() constant returns(string)
func (_PaymentObligation *PaymentObligationSession) Name() (string, error) {
	return _PaymentObligation.Contract.Name(&_PaymentObligation.CallOpts)
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() constant returns(string)
func (_PaymentObligation *PaymentObligationCallerSession) Name() (string, error) {
	return _PaymentObligation.Contract.Name(&_PaymentObligation.CallOpts)
}

// OwnerOf is a free data retrieval call binding the contract method 0x6352211e.
//
// Solidity: function ownerOf(_tokenId uint256) constant returns(address)
func (_PaymentObligation *PaymentObligationCaller) OwnerOf(opts *bind.CallOpts, _tokenId *big.Int) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _PaymentObligation.contract.Call(opts, out, "ownerOf", _tokenId)
	return *ret0, err
}

// OwnerOf is a free data retrieval call binding the contract method 0x6352211e.
//
// Solidity: function ownerOf(_tokenId uint256) constant returns(address)
func (_PaymentObligation *PaymentObligationSession) OwnerOf(_tokenId *big.Int) (common.Address, error) {
	return _PaymentObligation.Contract.OwnerOf(&_PaymentObligation.CallOpts, _tokenId)
}

// OwnerOf is a free data retrieval call binding the contract method 0x6352211e.
//
// Solidity: function ownerOf(_tokenId uint256) constant returns(address)
func (_PaymentObligation *PaymentObligationCallerSession) OwnerOf(_tokenId *big.Int) (common.Address, error) {
	return _PaymentObligation.Contract.OwnerOf(&_PaymentObligation.CallOpts, _tokenId)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(_interfaceId bytes4) constant returns(bool)
func (_PaymentObligation *PaymentObligationCaller) SupportsInterface(opts *bind.CallOpts, _interfaceId [4]byte) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _PaymentObligation.contract.Call(opts, out, "supportsInterface", _interfaceId)
	return *ret0, err
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(_interfaceId bytes4) constant returns(bool)
func (_PaymentObligation *PaymentObligationSession) SupportsInterface(_interfaceId [4]byte) (bool, error) {
	return _PaymentObligation.Contract.SupportsInterface(&_PaymentObligation.CallOpts, _interfaceId)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(_interfaceId bytes4) constant returns(bool)
func (_PaymentObligation *PaymentObligationCallerSession) SupportsInterface(_interfaceId [4]byte) (bool, error) {
	return _PaymentObligation.Contract.SupportsInterface(&_PaymentObligation.CallOpts, _interfaceId)
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() constant returns(string)
func (_PaymentObligation *PaymentObligationCaller) Symbol(opts *bind.CallOpts) (string, error) {
	var (
		ret0 = new(string)
	)
	out := ret0
	err := _PaymentObligation.contract.Call(opts, out, "symbol")
	return *ret0, err
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() constant returns(string)
func (_PaymentObligation *PaymentObligationSession) Symbol() (string, error) {
	return _PaymentObligation.Contract.Symbol(&_PaymentObligation.CallOpts)
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() constant returns(string)
func (_PaymentObligation *PaymentObligationCallerSession) Symbol() (string, error) {
	return _PaymentObligation.Contract.Symbol(&_PaymentObligation.CallOpts)
}

// TokenByIndex is a free data retrieval call binding the contract method 0x4f6ccce7.
//
// Solidity: function tokenByIndex(_index uint256) constant returns(uint256)
func (_PaymentObligation *PaymentObligationCaller) TokenByIndex(opts *bind.CallOpts, _index *big.Int) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _PaymentObligation.contract.Call(opts, out, "tokenByIndex", _index)
	return *ret0, err
}

// TokenByIndex is a free data retrieval call binding the contract method 0x4f6ccce7.
//
// Solidity: function tokenByIndex(_index uint256) constant returns(uint256)
func (_PaymentObligation *PaymentObligationSession) TokenByIndex(_index *big.Int) (*big.Int, error) {
	return _PaymentObligation.Contract.TokenByIndex(&_PaymentObligation.CallOpts, _index)
}

// TokenByIndex is a free data retrieval call binding the contract method 0x4f6ccce7.
//
// Solidity: function tokenByIndex(_index uint256) constant returns(uint256)
func (_PaymentObligation *PaymentObligationCallerSession) TokenByIndex(_index *big.Int) (*big.Int, error) {
	return _PaymentObligation.Contract.TokenByIndex(&_PaymentObligation.CallOpts, _index)
}

// TokenOfOwnerByIndex is a free data retrieval call binding the contract method 0x2f745c59.
//
// Solidity: function tokenOfOwnerByIndex(_owner address, _index uint256) constant returns(uint256)
func (_PaymentObligation *PaymentObligationCaller) TokenOfOwnerByIndex(opts *bind.CallOpts, _owner common.Address, _index *big.Int) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _PaymentObligation.contract.Call(opts, out, "tokenOfOwnerByIndex", _owner, _index)
	return *ret0, err
}

// TokenOfOwnerByIndex is a free data retrieval call binding the contract method 0x2f745c59.
//
// Solidity: function tokenOfOwnerByIndex(_owner address, _index uint256) constant returns(uint256)
func (_PaymentObligation *PaymentObligationSession) TokenOfOwnerByIndex(_owner common.Address, _index *big.Int) (*big.Int, error) {
	return _PaymentObligation.Contract.TokenOfOwnerByIndex(&_PaymentObligation.CallOpts, _owner, _index)
}

// TokenOfOwnerByIndex is a free data retrieval call binding the contract method 0x2f745c59.
//
// Solidity: function tokenOfOwnerByIndex(_owner address, _index uint256) constant returns(uint256)
func (_PaymentObligation *PaymentObligationCallerSession) TokenOfOwnerByIndex(_owner common.Address, _index *big.Int) (*big.Int, error) {
	return _PaymentObligation.Contract.TokenOfOwnerByIndex(&_PaymentObligation.CallOpts, _owner, _index)
}

// TokenURI is a free data retrieval call binding the contract method 0xc87b56dd.
//
// Solidity: function tokenURI(_tokenId uint256) constant returns(string)
func (_PaymentObligation *PaymentObligationCaller) TokenURI(opts *bind.CallOpts, _tokenId *big.Int) (string, error) {
	var (
		ret0 = new(string)
	)
	out := ret0
	err := _PaymentObligation.contract.Call(opts, out, "tokenURI", _tokenId)
	return *ret0, err
}

// TokenURI is a free data retrieval call binding the contract method 0xc87b56dd.
//
// Solidity: function tokenURI(_tokenId uint256) constant returns(string)
func (_PaymentObligation *PaymentObligationSession) TokenURI(_tokenId *big.Int) (string, error) {
	return _PaymentObligation.Contract.TokenURI(&_PaymentObligation.CallOpts, _tokenId)
}

// TokenURI is a free data retrieval call binding the contract method 0xc87b56dd.
//
// Solidity: function tokenURI(_tokenId uint256) constant returns(string)
func (_PaymentObligation *PaymentObligationCallerSession) TokenURI(_tokenId *big.Int) (string, error) {
	return _PaymentObligation.Contract.TokenURI(&_PaymentObligation.CallOpts, _tokenId)
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() constant returns(uint256)
func (_PaymentObligation *PaymentObligationCaller) TotalSupply(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _PaymentObligation.contract.Call(opts, out, "totalSupply")
	return *ret0, err
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() constant returns(uint256)
func (_PaymentObligation *PaymentObligationSession) TotalSupply() (*big.Int, error) {
	return _PaymentObligation.Contract.TotalSupply(&_PaymentObligation.CallOpts)
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() constant returns(uint256)
func (_PaymentObligation *PaymentObligationCallerSession) TotalSupply() (*big.Int, error) {
	return _PaymentObligation.Contract.TotalSupply(&_PaymentObligation.CallOpts)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(_to address, _tokenId uint256) returns()
func (_PaymentObligation *PaymentObligationTransactor) Approve(opts *bind.TransactOpts, _to common.Address, _tokenId *big.Int) (*types.Transaction, error) {
	return _PaymentObligation.contract.Transact(opts, "approve", _to, _tokenId)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(_to address, _tokenId uint256) returns()
func (_PaymentObligation *PaymentObligationSession) Approve(_to common.Address, _tokenId *big.Int) (*types.Transaction, error) {
	return _PaymentObligation.Contract.Approve(&_PaymentObligation.TransactOpts, _to, _tokenId)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(_to address, _tokenId uint256) returns()
func (_PaymentObligation *PaymentObligationTransactorSession) Approve(_to common.Address, _tokenId *big.Int) (*types.Transaction, error) {
	return _PaymentObligation.Contract.Approve(&_PaymentObligation.TransactOpts, _to, _tokenId)
}

// Mint is a paid mutator transaction binding the contract method 0xb279f8f1.
//
// Solidity: function mint(_to address, _tokenId uint256, _tokenURI string, _anchorId uint256, _merkleRoot bytes32, _values string[3], _salts bytes32[3], _proofs bytes32[][3]) returns()
func (_PaymentObligation *PaymentObligationTransactor) Mint(opts *bind.TransactOpts, _to common.Address, _tokenId *big.Int, _tokenURI string, _anchorId *big.Int, _merkleRoot [32]byte, _values [3]string, _salts [3][32]byte, _proofs [3][][32]byte) (*types.Transaction, error) {
	return _PaymentObligation.contract.Transact(opts, "mint", _to, _tokenId, _tokenURI, _anchorId, _merkleRoot, _values, _salts, _proofs)
}

// Mint is a paid mutator transaction binding the contract method 0xb279f8f1.
//
// Solidity: function mint(_to address, _tokenId uint256, _tokenURI string, _anchorId uint256, _merkleRoot bytes32, _values string[3], _salts bytes32[3], _proofs bytes32[][3]) returns()
func (_PaymentObligation *PaymentObligationSession) Mint(_to common.Address, _tokenId *big.Int, _tokenURI string, _anchorId *big.Int, _merkleRoot [32]byte, _values [3]string, _salts [3][32]byte, _proofs [3][][32]byte) (*types.Transaction, error) {
	return _PaymentObligation.Contract.Mint(&_PaymentObligation.TransactOpts, _to, _tokenId, _tokenURI, _anchorId, _merkleRoot, _values, _salts, _proofs)
}

// Mint is a paid mutator transaction binding the contract method 0xb279f8f1.
//
// Solidity: function mint(_to address, _tokenId uint256, _tokenURI string, _anchorId uint256, _merkleRoot bytes32, _values string[3], _salts bytes32[3], _proofs bytes32[][3]) returns()
func (_PaymentObligation *PaymentObligationTransactorSession) Mint(_to common.Address, _tokenId *big.Int, _tokenURI string, _anchorId *big.Int, _merkleRoot [32]byte, _values [3]string, _salts [3][32]byte, _proofs [3][][32]byte) (*types.Transaction, error) {
	return _PaymentObligation.Contract.Mint(&_PaymentObligation.TransactOpts, _to, _tokenId, _tokenURI, _anchorId, _merkleRoot, _values, _salts, _proofs)
}

// SafeTransferFrom is a paid mutator transaction binding the contract method 0xb88d4fde.
//
// Solidity: function safeTransferFrom(_from address, _to address, _tokenId uint256, _data bytes) returns()
func (_PaymentObligation *PaymentObligationTransactor) SafeTransferFrom(opts *bind.TransactOpts, _from common.Address, _to common.Address, _tokenId *big.Int, _data []byte) (*types.Transaction, error) {
	return _PaymentObligation.contract.Transact(opts, "safeTransferFrom", _from, _to, _tokenId, _data)
}

// SafeTransferFrom is a paid mutator transaction binding the contract method 0xb88d4fde.
//
// Solidity: function safeTransferFrom(_from address, _to address, _tokenId uint256, _data bytes) returns()
func (_PaymentObligation *PaymentObligationSession) SafeTransferFrom(_from common.Address, _to common.Address, _tokenId *big.Int, _data []byte) (*types.Transaction, error) {
	return _PaymentObligation.Contract.SafeTransferFrom(&_PaymentObligation.TransactOpts, _from, _to, _tokenId, _data)
}

// SafeTransferFrom is a paid mutator transaction binding the contract method 0xb88d4fde.
//
// Solidity: function safeTransferFrom(_from address, _to address, _tokenId uint256, _data bytes) returns()
func (_PaymentObligation *PaymentObligationTransactorSession) SafeTransferFrom(_from common.Address, _to common.Address, _tokenId *big.Int, _data []byte) (*types.Transaction, error) {
	return _PaymentObligation.Contract.SafeTransferFrom(&_PaymentObligation.TransactOpts, _from, _to, _tokenId, _data)
}

// SetApprovalForAll is a paid mutator transaction binding the contract method 0xa22cb465.
//
// Solidity: function setApprovalForAll(_to address, _approved bool) returns()
func (_PaymentObligation *PaymentObligationTransactor) SetApprovalForAll(opts *bind.TransactOpts, _to common.Address, _approved bool) (*types.Transaction, error) {
	return _PaymentObligation.contract.Transact(opts, "setApprovalForAll", _to, _approved)
}

// SetApprovalForAll is a paid mutator transaction binding the contract method 0xa22cb465.
//
// Solidity: function setApprovalForAll(_to address, _approved bool) returns()
func (_PaymentObligation *PaymentObligationSession) SetApprovalForAll(_to common.Address, _approved bool) (*types.Transaction, error) {
	return _PaymentObligation.Contract.SetApprovalForAll(&_PaymentObligation.TransactOpts, _to, _approved)
}

// SetApprovalForAll is a paid mutator transaction binding the contract method 0xa22cb465.
//
// Solidity: function setApprovalForAll(_to address, _approved bool) returns()
func (_PaymentObligation *PaymentObligationTransactorSession) SetApprovalForAll(_to common.Address, _approved bool) (*types.Transaction, error) {
	return _PaymentObligation.Contract.SetApprovalForAll(&_PaymentObligation.TransactOpts, _to, _approved)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(_from address, _to address, _tokenId uint256) returns()
func (_PaymentObligation *PaymentObligationTransactor) TransferFrom(opts *bind.TransactOpts, _from common.Address, _to common.Address, _tokenId *big.Int) (*types.Transaction, error) {
	return _PaymentObligation.contract.Transact(opts, "transferFrom", _from, _to, _tokenId)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(_from address, _to address, _tokenId uint256) returns()
func (_PaymentObligation *PaymentObligationSession) TransferFrom(_from common.Address, _to common.Address, _tokenId *big.Int) (*types.Transaction, error) {
	return _PaymentObligation.Contract.TransferFrom(&_PaymentObligation.TransactOpts, _from, _to, _tokenId)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(_from address, _to address, _tokenId uint256) returns()
func (_PaymentObligation *PaymentObligationTransactorSession) TransferFrom(_from common.Address, _to common.Address, _tokenId *big.Int) (*types.Transaction, error) {
	return _PaymentObligation.Contract.TransferFrom(&_PaymentObligation.TransactOpts, _from, _to, _tokenId)
}

// PaymentObligationApprovalIterator is returned from FilterApproval and is used to iterate over the raw logs and unpacked data for Approval events raised by the PaymentObligation contract.
type PaymentObligationApprovalIterator struct {
	Event *PaymentObligationApproval // Event containing the contract specifics and raw log

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
func (it *PaymentObligationApprovalIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PaymentObligationApproval)
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
		it.Event = new(PaymentObligationApproval)
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
func (it *PaymentObligationApprovalIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PaymentObligationApprovalIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PaymentObligationApproval represents a Approval event raised by the PaymentObligation contract.
type PaymentObligationApproval struct {
	Owner    common.Address
	Approved common.Address
	TokenId  *big.Int
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterApproval is a free log retrieval operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: e Approval(_owner indexed address, _approved indexed address, _tokenId indexed uint256)
func (_PaymentObligation *PaymentObligationFilterer) FilterApproval(opts *bind.FilterOpts, _owner []common.Address, _approved []common.Address, _tokenId []*big.Int) (*PaymentObligationApprovalIterator, error) {

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

	logs, sub, err := _PaymentObligation.contract.FilterLogs(opts, "Approval", _ownerRule, _approvedRule, _tokenIdRule)
	if err != nil {
		return nil, err
	}
	return &PaymentObligationApprovalIterator{contract: _PaymentObligation.contract, event: "Approval", logs: logs, sub: sub}, nil
}

// WatchApproval is a free log subscription operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: e Approval(_owner indexed address, _approved indexed address, _tokenId indexed uint256)
func (_PaymentObligation *PaymentObligationFilterer) WatchApproval(opts *bind.WatchOpts, sink chan<- *PaymentObligationApproval, _owner []common.Address, _approved []common.Address, _tokenId []*big.Int) (event.Subscription, error) {

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

	logs, sub, err := _PaymentObligation.contract.WatchLogs(opts, "Approval", _ownerRule, _approvedRule, _tokenIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PaymentObligationApproval)
				if err := _PaymentObligation.contract.UnpackLog(event, "Approval", log); err != nil {
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

// PaymentObligationApprovalForAllIterator is returned from FilterApprovalForAll and is used to iterate over the raw logs and unpacked data for ApprovalForAll events raised by the PaymentObligation contract.
type PaymentObligationApprovalForAllIterator struct {
	Event *PaymentObligationApprovalForAll // Event containing the contract specifics and raw log

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
func (it *PaymentObligationApprovalForAllIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PaymentObligationApprovalForAll)
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
		it.Event = new(PaymentObligationApprovalForAll)
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
func (it *PaymentObligationApprovalForAllIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PaymentObligationApprovalForAllIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PaymentObligationApprovalForAll represents a ApprovalForAll event raised by the PaymentObligation contract.
type PaymentObligationApprovalForAll struct {
	Owner    common.Address
	Operator common.Address
	Approved bool
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterApprovalForAll is a free log retrieval operation binding the contract event 0x17307eab39ab6107e8899845ad3d59bd9653f200f220920489ca2b5937696c31.
//
// Solidity: e ApprovalForAll(_owner indexed address, _operator indexed address, _approved bool)
func (_PaymentObligation *PaymentObligationFilterer) FilterApprovalForAll(opts *bind.FilterOpts, _owner []common.Address, _operator []common.Address) (*PaymentObligationApprovalForAllIterator, error) {

	var _ownerRule []interface{}
	for _, _ownerItem := range _owner {
		_ownerRule = append(_ownerRule, _ownerItem)
	}
	var _operatorRule []interface{}
	for _, _operatorItem := range _operator {
		_operatorRule = append(_operatorRule, _operatorItem)
	}

	logs, sub, err := _PaymentObligation.contract.FilterLogs(opts, "ApprovalForAll", _ownerRule, _operatorRule)
	if err != nil {
		return nil, err
	}
	return &PaymentObligationApprovalForAllIterator{contract: _PaymentObligation.contract, event: "ApprovalForAll", logs: logs, sub: sub}, nil
}

// WatchApprovalForAll is a free log subscription operation binding the contract event 0x17307eab39ab6107e8899845ad3d59bd9653f200f220920489ca2b5937696c31.
//
// Solidity: e ApprovalForAll(_owner indexed address, _operator indexed address, _approved bool)
func (_PaymentObligation *PaymentObligationFilterer) WatchApprovalForAll(opts *bind.WatchOpts, sink chan<- *PaymentObligationApprovalForAll, _owner []common.Address, _operator []common.Address) (event.Subscription, error) {

	var _ownerRule []interface{}
	for _, _ownerItem := range _owner {
		_ownerRule = append(_ownerRule, _ownerItem)
	}
	var _operatorRule []interface{}
	for _, _operatorItem := range _operator {
		_operatorRule = append(_operatorRule, _operatorItem)
	}

	logs, sub, err := _PaymentObligation.contract.WatchLogs(opts, "ApprovalForAll", _ownerRule, _operatorRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PaymentObligationApprovalForAll)
				if err := _PaymentObligation.contract.UnpackLog(event, "ApprovalForAll", log); err != nil {
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

// PaymentObligationPaymentObligationMintedIterator is returned from FilterPaymentObligationMinted and is used to iterate over the raw logs and unpacked data for PaymentObligationMinted events raised by the PaymentObligation contract.
type PaymentObligationPaymentObligationMintedIterator struct {
	Event *PaymentObligationPaymentObligationMinted // Event containing the contract specifics and raw log

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
func (it *PaymentObligationPaymentObligationMintedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PaymentObligationPaymentObligationMinted)
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
		it.Event = new(PaymentObligationPaymentObligationMinted)
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
func (it *PaymentObligationPaymentObligationMintedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PaymentObligationPaymentObligationMintedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PaymentObligationPaymentObligationMinted represents a PaymentObligationMinted event raised by the PaymentObligation contract.
type PaymentObligationPaymentObligationMinted struct {
	To       common.Address
	TokenId  *big.Int
	TokenURI string
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterPaymentObligationMinted is a free log retrieval operation binding the contract event 0xe2e4e975c4de5fbb1416db2c5ff8e2f4108bbfcdfd27a1a1eb03935cbfa3b8f9.
//
// Solidity: e PaymentObligationMinted(to address, tokenId uint256, tokenURI string)
func (_PaymentObligation *PaymentObligationFilterer) FilterPaymentObligationMinted(opts *bind.FilterOpts) (*PaymentObligationPaymentObligationMintedIterator, error) {

	logs, sub, err := _PaymentObligation.contract.FilterLogs(opts, "PaymentObligationMinted")
	if err != nil {
		return nil, err
	}
	return &PaymentObligationPaymentObligationMintedIterator{contract: _PaymentObligation.contract, event: "PaymentObligationMinted", logs: logs, sub: sub}, nil
}

// WatchPaymentObligationMinted is a free log subscription operation binding the contract event 0xe2e4e975c4de5fbb1416db2c5ff8e2f4108bbfcdfd27a1a1eb03935cbfa3b8f9.
//
// Solidity: e PaymentObligationMinted(to address, tokenId uint256, tokenURI string)
func (_PaymentObligation *PaymentObligationFilterer) WatchPaymentObligationMinted(opts *bind.WatchOpts, sink chan<- *PaymentObligationPaymentObligationMinted) (event.Subscription, error) {

	logs, sub, err := _PaymentObligation.contract.WatchLogs(opts, "PaymentObligationMinted")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PaymentObligationPaymentObligationMinted)
				if err := _PaymentObligation.contract.UnpackLog(event, "PaymentObligationMinted", log); err != nil {
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

// PaymentObligationTransferIterator is returned from FilterTransfer and is used to iterate over the raw logs and unpacked data for Transfer events raised by the PaymentObligation contract.
type PaymentObligationTransferIterator struct {
	Event *PaymentObligationTransfer // Event containing the contract specifics and raw log

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
func (it *PaymentObligationTransferIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PaymentObligationTransfer)
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
		it.Event = new(PaymentObligationTransfer)
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
func (it *PaymentObligationTransferIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PaymentObligationTransferIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PaymentObligationTransfer represents a Transfer event raised by the PaymentObligation contract.
type PaymentObligationTransfer struct {
	From    common.Address
	To      common.Address
	TokenId *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterTransfer is a free log retrieval operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: e Transfer(_from indexed address, _to indexed address, _tokenId indexed uint256)
func (_PaymentObligation *PaymentObligationFilterer) FilterTransfer(opts *bind.FilterOpts, _from []common.Address, _to []common.Address, _tokenId []*big.Int) (*PaymentObligationTransferIterator, error) {

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

	logs, sub, err := _PaymentObligation.contract.FilterLogs(opts, "Transfer", _fromRule, _toRule, _tokenIdRule)
	if err != nil {
		return nil, err
	}
	return &PaymentObligationTransferIterator{contract: _PaymentObligation.contract, event: "Transfer", logs: logs, sub: sub}, nil
}

// WatchTransfer is a free log subscription operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: e Transfer(_from indexed address, _to indexed address, _tokenId indexed uint256)
func (_PaymentObligation *PaymentObligationFilterer) WatchTransfer(opts *bind.WatchOpts, sink chan<- *PaymentObligationTransfer, _from []common.Address, _to []common.Address, _tokenId []*big.Int) (event.Subscription, error) {

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

	logs, sub, err := _PaymentObligation.contract.WatchLogs(opts, "Transfer", _fromRule, _toRule, _tokenIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PaymentObligationTransfer)
				if err := _PaymentObligation.contract.UnpackLog(event, "Transfer", log); err != nil {
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
