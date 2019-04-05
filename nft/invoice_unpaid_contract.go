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

// InvoiceUnpaidContractABI is the input ABI used to generate the binding from.
const InvoiceUnpaidContractABI = "[{\"constant\":true,\"inputs\":[{\"name\":\"interfaceId\",\"type\":\"bytes4\"}],\"name\":\"supportsInterface\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"name\",\"outputs\":[{\"name\":\"\",\"type\":\"string\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"getApproved\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"to\",\"type\":\"address\"},{\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"approve\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"currentIndexOfToken\",\"outputs\":[{\"name\":\"index\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"totalSupply\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"from\",\"type\":\"address\"},{\"name\":\"to\",\"type\":\"address\"},{\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"transferFrom\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"owner\",\"type\":\"address\"},{\"name\":\"index\",\"type\":\"uint256\"}],\"name\":\"tokenOfOwnerByIndex\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"from\",\"type\":\"address\"},{\"name\":\"to\",\"type\":\"address\"},{\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"safeTransferFrom\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"index\",\"type\":\"uint256\"}],\"name\":\"tokenByIndex\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"ownerOf\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"balanceOf\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"getAnchorRegistry\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"symbol\",\"outputs\":[{\"name\":\"\",\"type\":\"string\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"to\",\"type\":\"address\"},{\"name\":\"approved\",\"type\":\"bool\"}],\"name\":\"setApprovalForAll\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"from\",\"type\":\"address\"},{\"name\":\"to\",\"type\":\"address\"},{\"name\":\"tokenId\",\"type\":\"uint256\"},{\"name\":\"_data\",\"type\":\"bytes\"}],\"name\":\"safeTransferFrom\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"getIdentityFactory\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"tokenURI\",\"outputs\":[{\"name\":\"\",\"type\":\"string\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"owner\",\"type\":\"address\"},{\"name\":\"operator\",\"type\":\"address\"}],\"name\":\"isApprovedForAll\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"tokenId\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"tokenIndex\",\"type\":\"uint256\"}],\"name\":\"InvoiceUnpaidMinted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"to\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"Transfer\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"approved\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"Approval\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"operator\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"approved\",\"type\":\"bool\"}],\"name\":\"ApprovalForAll\",\"type\":\"event\"},{\"constant\":true,\"inputs\":[{\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"getTokenDetails\",\"outputs\":[{\"name\":\"invoiceSender\",\"type\":\"address\"},{\"name\":\"grossAmount\",\"type\":\"bytes\"},{\"name\":\"currency\",\"type\":\"bytes\"},{\"name\":\"dueDate\",\"type\":\"bytes\"},{\"name\":\"anchorId\",\"type\":\"uint256\"},{\"name\":\"nextAnchorId\",\"type\":\"uint256\"},{\"name\":\"documentRoot\",\"type\":\"bytes32\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"isTokenLatestDocument\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"tokenUriBase\",\"type\":\"string\"},{\"name\":\"anchorRegistry\",\"type\":\"address\"},{\"name\":\"identityFactory\",\"type\":\"address\"}],\"name\":\"initialize\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"name\",\"type\":\"string\"},{\"name\":\"symbol\",\"type\":\"string\"}],\"name\":\"initialize\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"initialize\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"name\",\"type\":\"string\"},{\"name\":\"symbol\",\"type\":\"string\"},{\"name\":\"tokenUriBase\",\"type\":\"string\"},{\"name\":\"anchorRegistry\",\"type\":\"address\"},{\"name\":\"identityFactory\",\"type\":\"address\"}],\"name\":\"initialize\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"to\",\"type\":\"address\"},{\"name\":\"tokenId\",\"type\":\"uint256\"},{\"name\":\"anchorId\",\"type\":\"uint256\"},{\"name\":\"properties\",\"type\":\"bytes[]\"},{\"name\":\"values\",\"type\":\"bytes[]\"},{\"name\":\"salts\",\"type\":\"bytes32[]\"},{\"name\":\"proofs\",\"type\":\"bytes32[][]\"}],\"name\":\"mint\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]"

// InvoiceUnpaidContract is an auto generated Go binding around an Ethereum contract.
type InvoiceUnpaidContract struct {
	InvoiceUnpaidContractCaller     // Read-only binding to the contract
	InvoiceUnpaidContractTransactor // Write-only binding to the contract
	InvoiceUnpaidContractFilterer   // Log filterer for contract events
}

// InvoiceUnpaidContractCaller is an auto generated read-only Go binding around an Ethereum contract.
type InvoiceUnpaidContractCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// InvoiceUnpaidContractTransactor is an auto generated write-only Go binding around an Ethereum contract.
type InvoiceUnpaidContractTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// InvoiceUnpaidContractFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type InvoiceUnpaidContractFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// InvoiceUnpaidContractSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type InvoiceUnpaidContractSession struct {
	Contract     *InvoiceUnpaidContract // Generic contract binding to set the session for
	CallOpts     bind.CallOpts          // Call options to use throughout this session
	TransactOpts bind.TransactOpts      // Transaction auth options to use throughout this session
}

// InvoiceUnpaidContractCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type InvoiceUnpaidContractCallerSession struct {
	Contract *InvoiceUnpaidContractCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts                // Call options to use throughout this session
}

// InvoiceUnpaidContractTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type InvoiceUnpaidContractTransactorSession struct {
	Contract     *InvoiceUnpaidContractTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts                // Transaction auth options to use throughout this session
}

// InvoiceUnpaidContractRaw is an auto generated low-level Go binding around an Ethereum contract.
type InvoiceUnpaidContractRaw struct {
	Contract *InvoiceUnpaidContract // Generic contract binding to access the raw methods on
}

// InvoiceUnpaidContractCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type InvoiceUnpaidContractCallerRaw struct {
	Contract *InvoiceUnpaidContractCaller // Generic read-only contract binding to access the raw methods on
}

// InvoiceUnpaidContractTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type InvoiceUnpaidContractTransactorRaw struct {
	Contract *InvoiceUnpaidContractTransactor // Generic write-only contract binding to access the raw methods on
}

// NewInvoiceUnpaidContract creates a new instance of InvoiceUnpaidContract, bound to a specific deployed contract.
func NewInvoiceUnpaidContract(address common.Address, backend bind.ContractBackend) (*InvoiceUnpaidContract, error) {
	contract, err := bindInvoiceUnpaidContract(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &InvoiceUnpaidContract{InvoiceUnpaidContractCaller: InvoiceUnpaidContractCaller{contract: contract}, InvoiceUnpaidContractTransactor: InvoiceUnpaidContractTransactor{contract: contract}, InvoiceUnpaidContractFilterer: InvoiceUnpaidContractFilterer{contract: contract}}, nil
}

// NewInvoiceUnpaidContractCaller creates a new read-only instance of InvoiceUnpaidContract, bound to a specific deployed contract.
func NewInvoiceUnpaidContractCaller(address common.Address, caller bind.ContractCaller) (*InvoiceUnpaidContractCaller, error) {
	contract, err := bindInvoiceUnpaidContract(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &InvoiceUnpaidContractCaller{contract: contract}, nil
}

// NewInvoiceUnpaidContractTransactor creates a new write-only instance of InvoiceUnpaidContract, bound to a specific deployed contract.
func NewInvoiceUnpaidContractTransactor(address common.Address, transactor bind.ContractTransactor) (*InvoiceUnpaidContractTransactor, error) {
	contract, err := bindInvoiceUnpaidContract(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &InvoiceUnpaidContractTransactor{contract: contract}, nil
}

// NewInvoiceUnpaidContractFilterer creates a new log filterer instance of InvoiceUnpaidContract, bound to a specific deployed contract.
func NewInvoiceUnpaidContractFilterer(address common.Address, filterer bind.ContractFilterer) (*InvoiceUnpaidContractFilterer, error) {
	contract, err := bindInvoiceUnpaidContract(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &InvoiceUnpaidContractFilterer{contract: contract}, nil
}

// bindInvoiceUnpaidContract binds a generic wrapper to an already deployed contract.
func bindInvoiceUnpaidContract(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(InvoiceUnpaidContractABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_InvoiceUnpaidContract *InvoiceUnpaidContractRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _InvoiceUnpaidContract.Contract.InvoiceUnpaidContractCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_InvoiceUnpaidContract *InvoiceUnpaidContractRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _InvoiceUnpaidContract.Contract.InvoiceUnpaidContractTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_InvoiceUnpaidContract *InvoiceUnpaidContractRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _InvoiceUnpaidContract.Contract.InvoiceUnpaidContractTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_InvoiceUnpaidContract *InvoiceUnpaidContractCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _InvoiceUnpaidContract.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_InvoiceUnpaidContract *InvoiceUnpaidContractTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _InvoiceUnpaidContract.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_InvoiceUnpaidContract *InvoiceUnpaidContractTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _InvoiceUnpaidContract.Contract.contract.Transact(opts, method, params...)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address owner) constant returns(uint256)
func (_InvoiceUnpaidContract *InvoiceUnpaidContractCaller) BalanceOf(opts *bind.CallOpts, owner common.Address) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _InvoiceUnpaidContract.contract.Call(opts, out, "balanceOf", owner)
	return *ret0, err
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address owner) constant returns(uint256)
func (_InvoiceUnpaidContract *InvoiceUnpaidContractSession) BalanceOf(owner common.Address) (*big.Int, error) {
	return _InvoiceUnpaidContract.Contract.BalanceOf(&_InvoiceUnpaidContract.CallOpts, owner)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address owner) constant returns(uint256)
func (_InvoiceUnpaidContract *InvoiceUnpaidContractCallerSession) BalanceOf(owner common.Address) (*big.Int, error) {
	return _InvoiceUnpaidContract.Contract.BalanceOf(&_InvoiceUnpaidContract.CallOpts, owner)
}

// CurrentIndexOfToken is a free data retrieval call binding the contract method 0x0ee77b1a.
//
// Solidity: function currentIndexOfToken(uint256 tokenId) constant returns(uint256 index)
func (_InvoiceUnpaidContract *InvoiceUnpaidContractCaller) CurrentIndexOfToken(opts *bind.CallOpts, tokenId *big.Int) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _InvoiceUnpaidContract.contract.Call(opts, out, "currentIndexOfToken", tokenId)
	return *ret0, err
}

// CurrentIndexOfToken is a free data retrieval call binding the contract method 0x0ee77b1a.
//
// Solidity: function currentIndexOfToken(uint256 tokenId) constant returns(uint256 index)
func (_InvoiceUnpaidContract *InvoiceUnpaidContractSession) CurrentIndexOfToken(tokenId *big.Int) (*big.Int, error) {
	return _InvoiceUnpaidContract.Contract.CurrentIndexOfToken(&_InvoiceUnpaidContract.CallOpts, tokenId)
}

// CurrentIndexOfToken is a free data retrieval call binding the contract method 0x0ee77b1a.
//
// Solidity: function currentIndexOfToken(uint256 tokenId) constant returns(uint256 index)
func (_InvoiceUnpaidContract *InvoiceUnpaidContractCallerSession) CurrentIndexOfToken(tokenId *big.Int) (*big.Int, error) {
	return _InvoiceUnpaidContract.Contract.CurrentIndexOfToken(&_InvoiceUnpaidContract.CallOpts, tokenId)
}

// GetAnchorRegistry is a free data retrieval call binding the contract method 0x95d506fc.
//
// Solidity: function getAnchorRegistry() constant returns(address)
func (_InvoiceUnpaidContract *InvoiceUnpaidContractCaller) GetAnchorRegistry(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _InvoiceUnpaidContract.contract.Call(opts, out, "getAnchorRegistry")
	return *ret0, err
}

// GetAnchorRegistry is a free data retrieval call binding the contract method 0x95d506fc.
//
// Solidity: function getAnchorRegistry() constant returns(address)
func (_InvoiceUnpaidContract *InvoiceUnpaidContractSession) GetAnchorRegistry() (common.Address, error) {
	return _InvoiceUnpaidContract.Contract.GetAnchorRegistry(&_InvoiceUnpaidContract.CallOpts)
}

// GetAnchorRegistry is a free data retrieval call binding the contract method 0x95d506fc.
//
// Solidity: function getAnchorRegistry() constant returns(address)
func (_InvoiceUnpaidContract *InvoiceUnpaidContractCallerSession) GetAnchorRegistry() (common.Address, error) {
	return _InvoiceUnpaidContract.Contract.GetAnchorRegistry(&_InvoiceUnpaidContract.CallOpts)
}

// GetApproved is a free data retrieval call binding the contract method 0x081812fc.
//
// Solidity: function getApproved(uint256 tokenId) constant returns(address)
func (_InvoiceUnpaidContract *InvoiceUnpaidContractCaller) GetApproved(opts *bind.CallOpts, tokenId *big.Int) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _InvoiceUnpaidContract.contract.Call(opts, out, "getApproved", tokenId)
	return *ret0, err
}

// GetApproved is a free data retrieval call binding the contract method 0x081812fc.
//
// Solidity: function getApproved(uint256 tokenId) constant returns(address)
func (_InvoiceUnpaidContract *InvoiceUnpaidContractSession) GetApproved(tokenId *big.Int) (common.Address, error) {
	return _InvoiceUnpaidContract.Contract.GetApproved(&_InvoiceUnpaidContract.CallOpts, tokenId)
}

// GetApproved is a free data retrieval call binding the contract method 0x081812fc.
//
// Solidity: function getApproved(uint256 tokenId) constant returns(address)
func (_InvoiceUnpaidContract *InvoiceUnpaidContractCallerSession) GetApproved(tokenId *big.Int) (common.Address, error) {
	return _InvoiceUnpaidContract.Contract.GetApproved(&_InvoiceUnpaidContract.CallOpts, tokenId)
}

// GetIdentityFactory is a free data retrieval call binding the contract method 0xba207d9b.
//
// Solidity: function getIdentityFactory() constant returns(address)
func (_InvoiceUnpaidContract *InvoiceUnpaidContractCaller) GetIdentityFactory(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _InvoiceUnpaidContract.contract.Call(opts, out, "getIdentityFactory")
	return *ret0, err
}

// GetIdentityFactory is a free data retrieval call binding the contract method 0xba207d9b.
//
// Solidity: function getIdentityFactory() constant returns(address)
func (_InvoiceUnpaidContract *InvoiceUnpaidContractSession) GetIdentityFactory() (common.Address, error) {
	return _InvoiceUnpaidContract.Contract.GetIdentityFactory(&_InvoiceUnpaidContract.CallOpts)
}

// GetIdentityFactory is a free data retrieval call binding the contract method 0xba207d9b.
//
// Solidity: function getIdentityFactory() constant returns(address)
func (_InvoiceUnpaidContract *InvoiceUnpaidContractCallerSession) GetIdentityFactory() (common.Address, error) {
	return _InvoiceUnpaidContract.Contract.GetIdentityFactory(&_InvoiceUnpaidContract.CallOpts)
}

// GetTokenDetails is a free data retrieval call binding the contract method 0xc1e03728.
//
// Solidity: function getTokenDetails(uint256 tokenId) constant returns(address invoiceSender, bytes grossAmount, bytes currency, bytes dueDate, uint256 anchorId, uint256 nextAnchorId, bytes32 documentRoot)
func (_InvoiceUnpaidContract *InvoiceUnpaidContractCaller) GetTokenDetails(opts *bind.CallOpts, tokenId *big.Int) (struct {
	InvoiceSender common.Address
	GrossAmount   []byte
	Currency      []byte
	DueDate       []byte
	AnchorId      *big.Int
	NextAnchorId  *big.Int
	DocumentRoot  [32]byte
}, error) {
	ret := new(struct {
		InvoiceSender common.Address
		GrossAmount   []byte
		Currency      []byte
		DueDate       []byte
		AnchorId      *big.Int
		NextAnchorId  *big.Int
		DocumentRoot  [32]byte
	})
	out := ret
	err := _InvoiceUnpaidContract.contract.Call(opts, out, "getTokenDetails", tokenId)
	return *ret, err
}

// GetTokenDetails is a free data retrieval call binding the contract method 0xc1e03728.
//
// Solidity: function getTokenDetails(uint256 tokenId) constant returns(address invoiceSender, bytes grossAmount, bytes currency, bytes dueDate, uint256 anchorId, uint256 nextAnchorId, bytes32 documentRoot)
func (_InvoiceUnpaidContract *InvoiceUnpaidContractSession) GetTokenDetails(tokenId *big.Int) (struct {
	InvoiceSender common.Address
	GrossAmount   []byte
	Currency      []byte
	DueDate       []byte
	AnchorId      *big.Int
	NextAnchorId  *big.Int
	DocumentRoot  [32]byte
}, error) {
	return _InvoiceUnpaidContract.Contract.GetTokenDetails(&_InvoiceUnpaidContract.CallOpts, tokenId)
}

// GetTokenDetails is a free data retrieval call binding the contract method 0xc1e03728.
//
// Solidity: function getTokenDetails(uint256 tokenId) constant returns(address invoiceSender, bytes grossAmount, bytes currency, bytes dueDate, uint256 anchorId, uint256 nextAnchorId, bytes32 documentRoot)
func (_InvoiceUnpaidContract *InvoiceUnpaidContractCallerSession) GetTokenDetails(tokenId *big.Int) (struct {
	InvoiceSender common.Address
	GrossAmount   []byte
	Currency      []byte
	DueDate       []byte
	AnchorId      *big.Int
	NextAnchorId  *big.Int
	DocumentRoot  [32]byte
}, error) {
	return _InvoiceUnpaidContract.Contract.GetTokenDetails(&_InvoiceUnpaidContract.CallOpts, tokenId)
}

// IsApprovedForAll is a free data retrieval call binding the contract method 0xe985e9c5.
//
// Solidity: function isApprovedForAll(address owner, address operator) constant returns(bool)
func (_InvoiceUnpaidContract *InvoiceUnpaidContractCaller) IsApprovedForAll(opts *bind.CallOpts, owner common.Address, operator common.Address) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _InvoiceUnpaidContract.contract.Call(opts, out, "isApprovedForAll", owner, operator)
	return *ret0, err
}

// IsApprovedForAll is a free data retrieval call binding the contract method 0xe985e9c5.
//
// Solidity: function isApprovedForAll(address owner, address operator) constant returns(bool)
func (_InvoiceUnpaidContract *InvoiceUnpaidContractSession) IsApprovedForAll(owner common.Address, operator common.Address) (bool, error) {
	return _InvoiceUnpaidContract.Contract.IsApprovedForAll(&_InvoiceUnpaidContract.CallOpts, owner, operator)
}

// IsApprovedForAll is a free data retrieval call binding the contract method 0xe985e9c5.
//
// Solidity: function isApprovedForAll(address owner, address operator) constant returns(bool)
func (_InvoiceUnpaidContract *InvoiceUnpaidContractCallerSession) IsApprovedForAll(owner common.Address, operator common.Address) (bool, error) {
	return _InvoiceUnpaidContract.Contract.IsApprovedForAll(&_InvoiceUnpaidContract.CallOpts, owner, operator)
}

// IsTokenLatestDocument is a free data retrieval call binding the contract method 0x8c504b67.
//
// Solidity: function isTokenLatestDocument(uint256 tokenId) constant returns(bool)
func (_InvoiceUnpaidContract *InvoiceUnpaidContractCaller) IsTokenLatestDocument(opts *bind.CallOpts, tokenId *big.Int) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _InvoiceUnpaidContract.contract.Call(opts, out, "isTokenLatestDocument", tokenId)
	return *ret0, err
}

// IsTokenLatestDocument is a free data retrieval call binding the contract method 0x8c504b67.
//
// Solidity: function isTokenLatestDocument(uint256 tokenId) constant returns(bool)
func (_InvoiceUnpaidContract *InvoiceUnpaidContractSession) IsTokenLatestDocument(tokenId *big.Int) (bool, error) {
	return _InvoiceUnpaidContract.Contract.IsTokenLatestDocument(&_InvoiceUnpaidContract.CallOpts, tokenId)
}

// IsTokenLatestDocument is a free data retrieval call binding the contract method 0x8c504b67.
//
// Solidity: function isTokenLatestDocument(uint256 tokenId) constant returns(bool)
func (_InvoiceUnpaidContract *InvoiceUnpaidContractCallerSession) IsTokenLatestDocument(tokenId *big.Int) (bool, error) {
	return _InvoiceUnpaidContract.Contract.IsTokenLatestDocument(&_InvoiceUnpaidContract.CallOpts, tokenId)
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() constant returns(string)
func (_InvoiceUnpaidContract *InvoiceUnpaidContractCaller) Name(opts *bind.CallOpts) (string, error) {
	var (
		ret0 = new(string)
	)
	out := ret0
	err := _InvoiceUnpaidContract.contract.Call(opts, out, "name")
	return *ret0, err
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() constant returns(string)
func (_InvoiceUnpaidContract *InvoiceUnpaidContractSession) Name() (string, error) {
	return _InvoiceUnpaidContract.Contract.Name(&_InvoiceUnpaidContract.CallOpts)
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() constant returns(string)
func (_InvoiceUnpaidContract *InvoiceUnpaidContractCallerSession) Name() (string, error) {
	return _InvoiceUnpaidContract.Contract.Name(&_InvoiceUnpaidContract.CallOpts)
}

// OwnerOf is a free data retrieval call binding the contract method 0x6352211e.
//
// Solidity: function ownerOf(uint256 tokenId) constant returns(address)
func (_InvoiceUnpaidContract *InvoiceUnpaidContractCaller) OwnerOf(opts *bind.CallOpts, tokenId *big.Int) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _InvoiceUnpaidContract.contract.Call(opts, out, "ownerOf", tokenId)
	return *ret0, err
}

// OwnerOf is a free data retrieval call binding the contract method 0x6352211e.
//
// Solidity: function ownerOf(uint256 tokenId) constant returns(address)
func (_InvoiceUnpaidContract *InvoiceUnpaidContractSession) OwnerOf(tokenId *big.Int) (common.Address, error) {
	return _InvoiceUnpaidContract.Contract.OwnerOf(&_InvoiceUnpaidContract.CallOpts, tokenId)
}

// OwnerOf is a free data retrieval call binding the contract method 0x6352211e.
//
// Solidity: function ownerOf(uint256 tokenId) constant returns(address)
func (_InvoiceUnpaidContract *InvoiceUnpaidContractCallerSession) OwnerOf(tokenId *big.Int) (common.Address, error) {
	return _InvoiceUnpaidContract.Contract.OwnerOf(&_InvoiceUnpaidContract.CallOpts, tokenId)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) constant returns(bool)
func (_InvoiceUnpaidContract *InvoiceUnpaidContractCaller) SupportsInterface(opts *bind.CallOpts, interfaceId [4]byte) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _InvoiceUnpaidContract.contract.Call(opts, out, "supportsInterface", interfaceId)
	return *ret0, err
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) constant returns(bool)
func (_InvoiceUnpaidContract *InvoiceUnpaidContractSession) SupportsInterface(interfaceId [4]byte) (bool, error) {
	return _InvoiceUnpaidContract.Contract.SupportsInterface(&_InvoiceUnpaidContract.CallOpts, interfaceId)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) constant returns(bool)
func (_InvoiceUnpaidContract *InvoiceUnpaidContractCallerSession) SupportsInterface(interfaceId [4]byte) (bool, error) {
	return _InvoiceUnpaidContract.Contract.SupportsInterface(&_InvoiceUnpaidContract.CallOpts, interfaceId)
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() constant returns(string)
func (_InvoiceUnpaidContract *InvoiceUnpaidContractCaller) Symbol(opts *bind.CallOpts) (string, error) {
	var (
		ret0 = new(string)
	)
	out := ret0
	err := _InvoiceUnpaidContract.contract.Call(opts, out, "symbol")
	return *ret0, err
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() constant returns(string)
func (_InvoiceUnpaidContract *InvoiceUnpaidContractSession) Symbol() (string, error) {
	return _InvoiceUnpaidContract.Contract.Symbol(&_InvoiceUnpaidContract.CallOpts)
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() constant returns(string)
func (_InvoiceUnpaidContract *InvoiceUnpaidContractCallerSession) Symbol() (string, error) {
	return _InvoiceUnpaidContract.Contract.Symbol(&_InvoiceUnpaidContract.CallOpts)
}

// TokenByIndex is a free data retrieval call binding the contract method 0x4f6ccce7.
//
// Solidity: function tokenByIndex(uint256 index) constant returns(uint256)
func (_InvoiceUnpaidContract *InvoiceUnpaidContractCaller) TokenByIndex(opts *bind.CallOpts, index *big.Int) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _InvoiceUnpaidContract.contract.Call(opts, out, "tokenByIndex", index)
	return *ret0, err
}

// TokenByIndex is a free data retrieval call binding the contract method 0x4f6ccce7.
//
// Solidity: function tokenByIndex(uint256 index) constant returns(uint256)
func (_InvoiceUnpaidContract *InvoiceUnpaidContractSession) TokenByIndex(index *big.Int) (*big.Int, error) {
	return _InvoiceUnpaidContract.Contract.TokenByIndex(&_InvoiceUnpaidContract.CallOpts, index)
}

// TokenByIndex is a free data retrieval call binding the contract method 0x4f6ccce7.
//
// Solidity: function tokenByIndex(uint256 index) constant returns(uint256)
func (_InvoiceUnpaidContract *InvoiceUnpaidContractCallerSession) TokenByIndex(index *big.Int) (*big.Int, error) {
	return _InvoiceUnpaidContract.Contract.TokenByIndex(&_InvoiceUnpaidContract.CallOpts, index)
}

// TokenOfOwnerByIndex is a free data retrieval call binding the contract method 0x2f745c59.
//
// Solidity: function tokenOfOwnerByIndex(address owner, uint256 index) constant returns(uint256)
func (_InvoiceUnpaidContract *InvoiceUnpaidContractCaller) TokenOfOwnerByIndex(opts *bind.CallOpts, owner common.Address, index *big.Int) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _InvoiceUnpaidContract.contract.Call(opts, out, "tokenOfOwnerByIndex", owner, index)
	return *ret0, err
}

// TokenOfOwnerByIndex is a free data retrieval call binding the contract method 0x2f745c59.
//
// Solidity: function tokenOfOwnerByIndex(address owner, uint256 index) constant returns(uint256)
func (_InvoiceUnpaidContract *InvoiceUnpaidContractSession) TokenOfOwnerByIndex(owner common.Address, index *big.Int) (*big.Int, error) {
	return _InvoiceUnpaidContract.Contract.TokenOfOwnerByIndex(&_InvoiceUnpaidContract.CallOpts, owner, index)
}

// TokenOfOwnerByIndex is a free data retrieval call binding the contract method 0x2f745c59.
//
// Solidity: function tokenOfOwnerByIndex(address owner, uint256 index) constant returns(uint256)
func (_InvoiceUnpaidContract *InvoiceUnpaidContractCallerSession) TokenOfOwnerByIndex(owner common.Address, index *big.Int) (*big.Int, error) {
	return _InvoiceUnpaidContract.Contract.TokenOfOwnerByIndex(&_InvoiceUnpaidContract.CallOpts, owner, index)
}

// TokenURI is a free data retrieval call binding the contract method 0xc87b56dd.
//
// Solidity: function tokenURI(uint256 tokenId) constant returns(string)
func (_InvoiceUnpaidContract *InvoiceUnpaidContractCaller) TokenURI(opts *bind.CallOpts, tokenId *big.Int) (string, error) {
	var (
		ret0 = new(string)
	)
	out := ret0
	err := _InvoiceUnpaidContract.contract.Call(opts, out, "tokenURI", tokenId)
	return *ret0, err
}

// TokenURI is a free data retrieval call binding the contract method 0xc87b56dd.
//
// Solidity: function tokenURI(uint256 tokenId) constant returns(string)
func (_InvoiceUnpaidContract *InvoiceUnpaidContractSession) TokenURI(tokenId *big.Int) (string, error) {
	return _InvoiceUnpaidContract.Contract.TokenURI(&_InvoiceUnpaidContract.CallOpts, tokenId)
}

// TokenURI is a free data retrieval call binding the contract method 0xc87b56dd.
//
// Solidity: function tokenURI(uint256 tokenId) constant returns(string)
func (_InvoiceUnpaidContract *InvoiceUnpaidContractCallerSession) TokenURI(tokenId *big.Int) (string, error) {
	return _InvoiceUnpaidContract.Contract.TokenURI(&_InvoiceUnpaidContract.CallOpts, tokenId)
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() constant returns(uint256)
func (_InvoiceUnpaidContract *InvoiceUnpaidContractCaller) TotalSupply(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _InvoiceUnpaidContract.contract.Call(opts, out, "totalSupply")
	return *ret0, err
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() constant returns(uint256)
func (_InvoiceUnpaidContract *InvoiceUnpaidContractSession) TotalSupply() (*big.Int, error) {
	return _InvoiceUnpaidContract.Contract.TotalSupply(&_InvoiceUnpaidContract.CallOpts)
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() constant returns(uint256)
func (_InvoiceUnpaidContract *InvoiceUnpaidContractCallerSession) TotalSupply() (*big.Int, error) {
	return _InvoiceUnpaidContract.Contract.TotalSupply(&_InvoiceUnpaidContract.CallOpts)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address to, uint256 tokenId) returns()
func (_InvoiceUnpaidContract *InvoiceUnpaidContractTransactor) Approve(opts *bind.TransactOpts, to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _InvoiceUnpaidContract.contract.Transact(opts, "approve", to, tokenId)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address to, uint256 tokenId) returns()
func (_InvoiceUnpaidContract *InvoiceUnpaidContractSession) Approve(to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _InvoiceUnpaidContract.Contract.Approve(&_InvoiceUnpaidContract.TransactOpts, to, tokenId)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address to, uint256 tokenId) returns()
func (_InvoiceUnpaidContract *InvoiceUnpaidContractTransactorSession) Approve(to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _InvoiceUnpaidContract.Contract.Approve(&_InvoiceUnpaidContract.TransactOpts, to, tokenId)
}

// Initialize is a paid mutator transaction binding the contract method 0xd6d0faee.
//
// Solidity: function initialize(string name, string symbol, string tokenUriBase, address anchorRegistry, address identityFactory) returns()
func (_InvoiceUnpaidContract *InvoiceUnpaidContractTransactor) Initialize(opts *bind.TransactOpts, name string, symbol string, tokenUriBase string, anchorRegistry common.Address, identityFactory common.Address) (*types.Transaction, error) {
	return _InvoiceUnpaidContract.contract.Transact(opts, "initialize", name, symbol, tokenUriBase, anchorRegistry, identityFactory)
}

// Initialize is a paid mutator transaction binding the contract method 0xd6d0faee.
//
// Solidity: function initialize(string name, string symbol, string tokenUriBase, address anchorRegistry, address identityFactory) returns()
func (_InvoiceUnpaidContract *InvoiceUnpaidContractSession) Initialize(name string, symbol string, tokenUriBase string, anchorRegistry common.Address, identityFactory common.Address) (*types.Transaction, error) {
	return _InvoiceUnpaidContract.Contract.Initialize(&_InvoiceUnpaidContract.TransactOpts, name, symbol, tokenUriBase, anchorRegistry, identityFactory)
}

// Initialize is a paid mutator transaction binding the contract method 0xd6d0faee.
//
// Solidity: function initialize(string name, string symbol, string tokenUriBase, address anchorRegistry, address identityFactory) returns()
func (_InvoiceUnpaidContract *InvoiceUnpaidContractTransactorSession) Initialize(name string, symbol string, tokenUriBase string, anchorRegistry common.Address, identityFactory common.Address) (*types.Transaction, error) {
	return _InvoiceUnpaidContract.Contract.Initialize(&_InvoiceUnpaidContract.TransactOpts, name, symbol, tokenUriBase, anchorRegistry, identityFactory)
}

// Mint is a paid mutator transaction binding the contract method 0xc15778db.
//
// Solidity: function mint(address to, uint256 tokenId, uint256 anchorId, bytes[] properties, bytes[] values, bytes32[] salts, bytes32[][] proofs) returns()
func (_InvoiceUnpaidContract *InvoiceUnpaidContractTransactor) Mint(opts *bind.TransactOpts, to common.Address, tokenId *big.Int, anchorId *big.Int, properties [][]byte, values [][]byte, salts [][32]byte, proofs [][][32]byte) (*types.Transaction, error) {
	return _InvoiceUnpaidContract.contract.Transact(opts, "mint", to, tokenId, anchorId, properties, values, salts, proofs)
}

// Mint is a paid mutator transaction binding the contract method 0xc15778db.
//
// Solidity: function mint(address to, uint256 tokenId, uint256 anchorId, bytes[] properties, bytes[] values, bytes32[] salts, bytes32[][] proofs) returns()
func (_InvoiceUnpaidContract *InvoiceUnpaidContractSession) Mint(to common.Address, tokenId *big.Int, anchorId *big.Int, properties [][]byte, values [][]byte, salts [][32]byte, proofs [][][32]byte) (*types.Transaction, error) {
	return _InvoiceUnpaidContract.Contract.Mint(&_InvoiceUnpaidContract.TransactOpts, to, tokenId, anchorId, properties, values, salts, proofs)
}

// Mint is a paid mutator transaction binding the contract method 0xc15778db.
//
// Solidity: function mint(address to, uint256 tokenId, uint256 anchorId, bytes[] properties, bytes[] values, bytes32[] salts, bytes32[][] proofs) returns()
func (_InvoiceUnpaidContract *InvoiceUnpaidContractTransactorSession) Mint(to common.Address, tokenId *big.Int, anchorId *big.Int, properties [][]byte, values [][]byte, salts [][32]byte, proofs [][][32]byte) (*types.Transaction, error) {
	return _InvoiceUnpaidContract.Contract.Mint(&_InvoiceUnpaidContract.TransactOpts, to, tokenId, anchorId, properties, values, salts, proofs)
}

// SafeTransferFrom is a paid mutator transaction binding the contract method 0xb88d4fde.
//
// Solidity: function safeTransferFrom(address from, address to, uint256 tokenId, bytes _data) returns()
func (_InvoiceUnpaidContract *InvoiceUnpaidContractTransactor) SafeTransferFrom(opts *bind.TransactOpts, from common.Address, to common.Address, tokenId *big.Int, _data []byte) (*types.Transaction, error) {
	return _InvoiceUnpaidContract.contract.Transact(opts, "safeTransferFrom", from, to, tokenId, _data)
}

// SafeTransferFrom is a paid mutator transaction binding the contract method 0xb88d4fde.
//
// Solidity: function safeTransferFrom(address from, address to, uint256 tokenId, bytes _data) returns()
func (_InvoiceUnpaidContract *InvoiceUnpaidContractSession) SafeTransferFrom(from common.Address, to common.Address, tokenId *big.Int, _data []byte) (*types.Transaction, error) {
	return _InvoiceUnpaidContract.Contract.SafeTransferFrom(&_InvoiceUnpaidContract.TransactOpts, from, to, tokenId, _data)
}

// SafeTransferFrom is a paid mutator transaction binding the contract method 0xb88d4fde.
//
// Solidity: function safeTransferFrom(address from, address to, uint256 tokenId, bytes _data) returns()
func (_InvoiceUnpaidContract *InvoiceUnpaidContractTransactorSession) SafeTransferFrom(from common.Address, to common.Address, tokenId *big.Int, _data []byte) (*types.Transaction, error) {
	return _InvoiceUnpaidContract.Contract.SafeTransferFrom(&_InvoiceUnpaidContract.TransactOpts, from, to, tokenId, _data)
}

// SetApprovalForAll is a paid mutator transaction binding the contract method 0xa22cb465.
//
// Solidity: function setApprovalForAll(address to, bool approved) returns()
func (_InvoiceUnpaidContract *InvoiceUnpaidContractTransactor) SetApprovalForAll(opts *bind.TransactOpts, to common.Address, approved bool) (*types.Transaction, error) {
	return _InvoiceUnpaidContract.contract.Transact(opts, "setApprovalForAll", to, approved)
}

// SetApprovalForAll is a paid mutator transaction binding the contract method 0xa22cb465.
//
// Solidity: function setApprovalForAll(address to, bool approved) returns()
func (_InvoiceUnpaidContract *InvoiceUnpaidContractSession) SetApprovalForAll(to common.Address, approved bool) (*types.Transaction, error) {
	return _InvoiceUnpaidContract.Contract.SetApprovalForAll(&_InvoiceUnpaidContract.TransactOpts, to, approved)
}

// SetApprovalForAll is a paid mutator transaction binding the contract method 0xa22cb465.
//
// Solidity: function setApprovalForAll(address to, bool approved) returns()
func (_InvoiceUnpaidContract *InvoiceUnpaidContractTransactorSession) SetApprovalForAll(to common.Address, approved bool) (*types.Transaction, error) {
	return _InvoiceUnpaidContract.Contract.SetApprovalForAll(&_InvoiceUnpaidContract.TransactOpts, to, approved)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address from, address to, uint256 tokenId) returns()
func (_InvoiceUnpaidContract *InvoiceUnpaidContractTransactor) TransferFrom(opts *bind.TransactOpts, from common.Address, to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _InvoiceUnpaidContract.contract.Transact(opts, "transferFrom", from, to, tokenId)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address from, address to, uint256 tokenId) returns()
func (_InvoiceUnpaidContract *InvoiceUnpaidContractSession) TransferFrom(from common.Address, to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _InvoiceUnpaidContract.Contract.TransferFrom(&_InvoiceUnpaidContract.TransactOpts, from, to, tokenId)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address from, address to, uint256 tokenId) returns()
func (_InvoiceUnpaidContract *InvoiceUnpaidContractTransactorSession) TransferFrom(from common.Address, to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _InvoiceUnpaidContract.Contract.TransferFrom(&_InvoiceUnpaidContract.TransactOpts, from, to, tokenId)
}

// InvoiceUnpaidContractApprovalIterator is returned from FilterApproval and is used to iterate over the raw logs and unpacked data for Approval events raised by the InvoiceUnpaidContract contract.
type InvoiceUnpaidContractApprovalIterator struct {
	Event *InvoiceUnpaidContractApproval // Event containing the contract specifics and raw log

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
func (it *InvoiceUnpaidContractApprovalIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(InvoiceUnpaidContractApproval)
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
		it.Event = new(InvoiceUnpaidContractApproval)
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
func (it *InvoiceUnpaidContractApprovalIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *InvoiceUnpaidContractApprovalIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// InvoiceUnpaidContractApproval represents a Approval event raised by the InvoiceUnpaidContract contract.
type InvoiceUnpaidContractApproval struct {
	Owner    common.Address
	Approved common.Address
	TokenId  *big.Int
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterApproval is a free log retrieval operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed owner, address indexed approved, uint256 indexed tokenId)
func (_InvoiceUnpaidContract *InvoiceUnpaidContractFilterer) FilterApproval(opts *bind.FilterOpts, owner []common.Address, approved []common.Address, tokenId []*big.Int) (*InvoiceUnpaidContractApprovalIterator, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var approvedRule []interface{}
	for _, approvedItem := range approved {
		approvedRule = append(approvedRule, approvedItem)
	}
	var tokenIdRule []interface{}
	for _, tokenIdItem := range tokenId {
		tokenIdRule = append(tokenIdRule, tokenIdItem)
	}

	logs, sub, err := _InvoiceUnpaidContract.contract.FilterLogs(opts, "Approval", ownerRule, approvedRule, tokenIdRule)
	if err != nil {
		return nil, err
	}
	return &InvoiceUnpaidContractApprovalIterator{contract: _InvoiceUnpaidContract.contract, event: "Approval", logs: logs, sub: sub}, nil
}

// WatchApproval is a free log subscription operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed owner, address indexed approved, uint256 indexed tokenId)
func (_InvoiceUnpaidContract *InvoiceUnpaidContractFilterer) WatchApproval(opts *bind.WatchOpts, sink chan<- *InvoiceUnpaidContractApproval, owner []common.Address, approved []common.Address, tokenId []*big.Int) (event.Subscription, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var approvedRule []interface{}
	for _, approvedItem := range approved {
		approvedRule = append(approvedRule, approvedItem)
	}
	var tokenIdRule []interface{}
	for _, tokenIdItem := range tokenId {
		tokenIdRule = append(tokenIdRule, tokenIdItem)
	}

	logs, sub, err := _InvoiceUnpaidContract.contract.WatchLogs(opts, "Approval", ownerRule, approvedRule, tokenIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(InvoiceUnpaidContractApproval)
				if err := _InvoiceUnpaidContract.contract.UnpackLog(event, "Approval", log); err != nil {
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

// InvoiceUnpaidContractApprovalForAllIterator is returned from FilterApprovalForAll and is used to iterate over the raw logs and unpacked data for ApprovalForAll events raised by the InvoiceUnpaidContract contract.
type InvoiceUnpaidContractApprovalForAllIterator struct {
	Event *InvoiceUnpaidContractApprovalForAll // Event containing the contract specifics and raw log

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
func (it *InvoiceUnpaidContractApprovalForAllIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(InvoiceUnpaidContractApprovalForAll)
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
		it.Event = new(InvoiceUnpaidContractApprovalForAll)
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
func (it *InvoiceUnpaidContractApprovalForAllIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *InvoiceUnpaidContractApprovalForAllIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// InvoiceUnpaidContractApprovalForAll represents a ApprovalForAll event raised by the InvoiceUnpaidContract contract.
type InvoiceUnpaidContractApprovalForAll struct {
	Owner    common.Address
	Operator common.Address
	Approved bool
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterApprovalForAll is a free log retrieval operation binding the contract event 0x17307eab39ab6107e8899845ad3d59bd9653f200f220920489ca2b5937696c31.
//
// Solidity: event ApprovalForAll(address indexed owner, address indexed operator, bool approved)
func (_InvoiceUnpaidContract *InvoiceUnpaidContractFilterer) FilterApprovalForAll(opts *bind.FilterOpts, owner []common.Address, operator []common.Address) (*InvoiceUnpaidContractApprovalForAllIterator, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}

	logs, sub, err := _InvoiceUnpaidContract.contract.FilterLogs(opts, "ApprovalForAll", ownerRule, operatorRule)
	if err != nil {
		return nil, err
	}
	return &InvoiceUnpaidContractApprovalForAllIterator{contract: _InvoiceUnpaidContract.contract, event: "ApprovalForAll", logs: logs, sub: sub}, nil
}

// WatchApprovalForAll is a free log subscription operation binding the contract event 0x17307eab39ab6107e8899845ad3d59bd9653f200f220920489ca2b5937696c31.
//
// Solidity: event ApprovalForAll(address indexed owner, address indexed operator, bool approved)
func (_InvoiceUnpaidContract *InvoiceUnpaidContractFilterer) WatchApprovalForAll(opts *bind.WatchOpts, sink chan<- *InvoiceUnpaidContractApprovalForAll, owner []common.Address, operator []common.Address) (event.Subscription, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}

	logs, sub, err := _InvoiceUnpaidContract.contract.WatchLogs(opts, "ApprovalForAll", ownerRule, operatorRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(InvoiceUnpaidContractApprovalForAll)
				if err := _InvoiceUnpaidContract.contract.UnpackLog(event, "ApprovalForAll", log); err != nil {
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

// InvoiceUnpaidContractInvoiceUnpaidMintedIterator is returned from FilterInvoiceUnpaidMinted and is used to iterate over the raw logs and unpacked data for InvoiceUnpaidMinted events raised by the InvoiceUnpaidContract contract.
type InvoiceUnpaidContractInvoiceUnpaidMintedIterator struct {
	Event *InvoiceUnpaidContractInvoiceUnpaidMinted // Event containing the contract specifics and raw log

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
func (it *InvoiceUnpaidContractInvoiceUnpaidMintedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(InvoiceUnpaidContractInvoiceUnpaidMinted)
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
		it.Event = new(InvoiceUnpaidContractInvoiceUnpaidMinted)
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
func (it *InvoiceUnpaidContractInvoiceUnpaidMintedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *InvoiceUnpaidContractInvoiceUnpaidMintedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// InvoiceUnpaidContractInvoiceUnpaidMinted represents a InvoiceUnpaidMinted event raised by the InvoiceUnpaidContract contract.
type InvoiceUnpaidContractInvoiceUnpaidMinted struct {
	To         common.Address
	TokenId    *big.Int
	TokenIndex *big.Int
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterInvoiceUnpaidMinted is a free log retrieval operation binding the contract event 0x5a336d012a393ced99cb7c5ae8ca7d664290e1673401a403e6c2363823eb5bdc.
//
// Solidity: event InvoiceUnpaidMinted(address to, uint256 tokenId, uint256 tokenIndex)
func (_InvoiceUnpaidContract *InvoiceUnpaidContractFilterer) FilterInvoiceUnpaidMinted(opts *bind.FilterOpts) (*InvoiceUnpaidContractInvoiceUnpaidMintedIterator, error) {

	logs, sub, err := _InvoiceUnpaidContract.contract.FilterLogs(opts, "InvoiceUnpaidMinted")
	if err != nil {
		return nil, err
	}
	return &InvoiceUnpaidContractInvoiceUnpaidMintedIterator{contract: _InvoiceUnpaidContract.contract, event: "InvoiceUnpaidMinted", logs: logs, sub: sub}, nil
}

// WatchInvoiceUnpaidMinted is a free log subscription operation binding the contract event 0x5a336d012a393ced99cb7c5ae8ca7d664290e1673401a403e6c2363823eb5bdc.
//
// Solidity: event InvoiceUnpaidMinted(address to, uint256 tokenId, uint256 tokenIndex)
func (_InvoiceUnpaidContract *InvoiceUnpaidContractFilterer) WatchInvoiceUnpaidMinted(opts *bind.WatchOpts, sink chan<- *InvoiceUnpaidContractInvoiceUnpaidMinted) (event.Subscription, error) {

	logs, sub, err := _InvoiceUnpaidContract.contract.WatchLogs(opts, "InvoiceUnpaidMinted")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(InvoiceUnpaidContractInvoiceUnpaidMinted)
				if err := _InvoiceUnpaidContract.contract.UnpackLog(event, "InvoiceUnpaidMinted", log); err != nil {
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

// InvoiceUnpaidContractTransferIterator is returned from FilterTransfer and is used to iterate over the raw logs and unpacked data for Transfer events raised by the InvoiceUnpaidContract contract.
type InvoiceUnpaidContractTransferIterator struct {
	Event *InvoiceUnpaidContractTransfer // Event containing the contract specifics and raw log

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
func (it *InvoiceUnpaidContractTransferIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(InvoiceUnpaidContractTransfer)
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
		it.Event = new(InvoiceUnpaidContractTransfer)
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
func (it *InvoiceUnpaidContractTransferIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *InvoiceUnpaidContractTransferIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// InvoiceUnpaidContractTransfer represents a Transfer event raised by the InvoiceUnpaidContract contract.
type InvoiceUnpaidContractTransfer struct {
	From    common.Address
	To      common.Address
	TokenId *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterTransfer is a free log retrieval operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed from, address indexed to, uint256 indexed tokenId)
func (_InvoiceUnpaidContract *InvoiceUnpaidContractFilterer) FilterTransfer(opts *bind.FilterOpts, from []common.Address, to []common.Address, tokenId []*big.Int) (*InvoiceUnpaidContractTransferIterator, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}
	var tokenIdRule []interface{}
	for _, tokenIdItem := range tokenId {
		tokenIdRule = append(tokenIdRule, tokenIdItem)
	}

	logs, sub, err := _InvoiceUnpaidContract.contract.FilterLogs(opts, "Transfer", fromRule, toRule, tokenIdRule)
	if err != nil {
		return nil, err
	}
	return &InvoiceUnpaidContractTransferIterator{contract: _InvoiceUnpaidContract.contract, event: "Transfer", logs: logs, sub: sub}, nil
}

// WatchTransfer is a free log subscription operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed from, address indexed to, uint256 indexed tokenId)
func (_InvoiceUnpaidContract *InvoiceUnpaidContractFilterer) WatchTransfer(opts *bind.WatchOpts, sink chan<- *InvoiceUnpaidContractTransfer, from []common.Address, to []common.Address, tokenId []*big.Int) (event.Subscription, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}
	var tokenIdRule []interface{}
	for _, tokenIdItem := range tokenId {
		tokenIdRule = append(tokenIdRule, tokenIdItem)
	}

	logs, sub, err := _InvoiceUnpaidContract.contract.WatchLogs(opts, "Transfer", fromRule, toRule, tokenIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(InvoiceUnpaidContractTransfer)
				if err := _InvoiceUnpaidContract.contract.UnpackLog(event, "Transfer", log); err != nil {
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
