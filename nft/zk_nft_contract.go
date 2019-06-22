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

// ZkNFTContractABI is the input ABI used to generate the binding from.
const ZkNFTContractABI = "[{\"constant\":true,\"inputs\":[{\"name\":\"interfaceId\",\"type\":\"bytes4\"}],\"name\":\"supportsInterface\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"name\",\"outputs\":[{\"name\":\"\",\"type\":\"string\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"getApproved\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"uri_prefix\",\"outputs\":[{\"name\":\"\",\"type\":\"string\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"to\",\"type\":\"address\"},{\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"approve\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"what\",\"type\":\"bytes32\"},{\"name\":\"data_\",\"type\":\"string\"}],\"name\":\"file\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"totalSupply\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"from\",\"type\":\"address\"},{\"name\":\"to\",\"type\":\"address\"},{\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"transferFrom\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"owner\",\"type\":\"address\"},{\"name\":\"index\",\"type\":\"uint256\"}],\"name\":\"tokenOfOwnerByIndex\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"ratings\",\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"from\",\"type\":\"address\"},{\"name\":\"to\",\"type\":\"address\"},{\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"safeTransferFrom\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"a\",\"type\":\"uint256[2]\"},{\"name\":\"b\",\"type\":\"uint256[2][2]\"},{\"name\":\"c\",\"type\":\"uint256[2]\"},{\"name\":\"input\",\"type\":\"uint256[7]\"}],\"name\":\"verifyTx\",\"outputs\":[{\"name\":\"r\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"index\",\"type\":\"uint256\"}],\"name\":\"tokenByIndex\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"ownerOf\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"balanceOf\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"x\",\"type\":\"bytes32\"}],\"name\":\"unpack\",\"outputs\":[{\"name\":\"y\",\"type\":\"uint256\"},{\"name\":\"z\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"symbol\",\"outputs\":[{\"name\":\"\",\"type\":\"string\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"anchors\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"to\",\"type\":\"address\"},{\"name\":\"approved\",\"type\":\"bool\"}],\"name\":\"setApprovalForAll\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"from\",\"type\":\"address\"},{\"name\":\"to\",\"type\":\"address\"},{\"name\":\"tokenId\",\"type\":\"uint256\"},{\"name\":\"_data\",\"type\":\"bytes\"}],\"name\":\"safeTransferFrom\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"tokenURI\",\"outputs\":[{\"name\":\"\",\"type\":\"string\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"usr\",\"type\":\"address\"},{\"name\":\"tkn\",\"type\":\"uint256\"},{\"name\":\"anchor\",\"type\":\"uint256\"},{\"name\":\"data_root\",\"type\":\"bytes32\"},{\"name\":\"signatures_root\",\"type\":\"bytes32\"},{\"name\":\"amount\",\"type\":\"uint256\"},{\"name\":\"rating\",\"type\":\"uint256\"},{\"name\":\"points\",\"type\":\"uint256[8]\"}],\"name\":\"mint\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"data_root\",\"type\":\"bytes32\"},{\"name\":\"nft_amount\",\"type\":\"uint256\"},{\"name\":\"rating\",\"type\":\"uint256\"},{\"name\":\"points\",\"type\":\"uint256[8]\"}],\"name\":\"verify\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"owner\",\"type\":\"address\"},{\"name\":\"operator\",\"type\":\"address\"}],\"name\":\"isApprovedForAll\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"what\",\"type\":\"bytes32\"},{\"name\":\"data_\",\"type\":\"bytes32\"}],\"name\":\"file\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"uri\",\"outputs\":[{\"name\":\"\",\"type\":\"string\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"data\",\"outputs\":[{\"name\":\"amount\",\"type\":\"uint256\"},{\"name\":\"anchor\",\"type\":\"uint256\"},{\"name\":\"rating\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"anchor\",\"type\":\"uint256\"},{\"name\":\"droot\",\"type\":\"bytes32\"},{\"name\":\"sigs\",\"type\":\"bytes32\"}],\"name\":\"checkAnchor\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"name\":\"name\",\"type\":\"string\"},{\"name\":\"symbol\",\"type\":\"string\"},{\"name\":\"anchors_\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"to\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"Transfer\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"approved\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"Approval\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"operator\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"approved\",\"type\":\"bool\"}],\"name\":\"ApprovalForAll\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"s\",\"type\":\"string\"}],\"name\":\"Verified\",\"type\":\"event\"}]"

// ZkNFTContract is an auto generated Go binding around an Ethereum contract.
type ZkNFTContract struct {
	ZkNFTContractCaller     // Read-only binding to the contract
	ZkNFTContractTransactor // Write-only binding to the contract
	ZkNFTContractFilterer   // Log filterer for contract events
}

// ZkNFTContractCaller is an auto generated read-only Go binding around an Ethereum contract.
type ZkNFTContractCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ZkNFTContractTransactor is an auto generated write-only Go binding around an Ethereum contract.
type ZkNFTContractTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ZkNFTContractFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type ZkNFTContractFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ZkNFTContractSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ZkNFTContractSession struct {
	Contract     *ZkNFTContract    // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// ZkNFTContractCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ZkNFTContractCallerSession struct {
	Contract *ZkNFTContractCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts        // Call options to use throughout this session
}

// ZkNFTContractTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ZkNFTContractTransactorSession struct {
	Contract     *ZkNFTContractTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts        // Transaction auth options to use throughout this session
}

// ZkNFTContractRaw is an auto generated low-level Go binding around an Ethereum contract.
type ZkNFTContractRaw struct {
	Contract *ZkNFTContract // Generic contract binding to access the raw methods on
}

// ZkNFTContractCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ZkNFTContractCallerRaw struct {
	Contract *ZkNFTContractCaller // Generic read-only contract binding to access the raw methods on
}

// ZkNFTContractTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ZkNFTContractTransactorRaw struct {
	Contract *ZkNFTContractTransactor // Generic write-only contract binding to access the raw methods on
}

// NewZkNFTContract creates a new instance of ZkNFTContract, bound to a specific deployed contract.
func NewZkNFTContract(address common.Address, backend bind.ContractBackend) (*ZkNFTContract, error) {
	contract, err := bindZkNFTContract(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &ZkNFTContract{ZkNFTContractCaller: ZkNFTContractCaller{contract: contract}, ZkNFTContractTransactor: ZkNFTContractTransactor{contract: contract}, ZkNFTContractFilterer: ZkNFTContractFilterer{contract: contract}}, nil
}

// NewZkNFTContractCaller creates a new read-only instance of ZkNFTContract, bound to a specific deployed contract.
func NewZkNFTContractCaller(address common.Address, caller bind.ContractCaller) (*ZkNFTContractCaller, error) {
	contract, err := bindZkNFTContract(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ZkNFTContractCaller{contract: contract}, nil
}

// NewZkNFTContractTransactor creates a new write-only instance of ZkNFTContract, bound to a specific deployed contract.
func NewZkNFTContractTransactor(address common.Address, transactor bind.ContractTransactor) (*ZkNFTContractTransactor, error) {
	contract, err := bindZkNFTContract(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ZkNFTContractTransactor{contract: contract}, nil
}

// NewZkNFTContractFilterer creates a new log filterer instance of ZkNFTContract, bound to a specific deployed contract.
func NewZkNFTContractFilterer(address common.Address, filterer bind.ContractFilterer) (*ZkNFTContractFilterer, error) {
	contract, err := bindZkNFTContract(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ZkNFTContractFilterer{contract: contract}, nil
}

// bindZkNFTContract binds a generic wrapper to an already deployed contract.
func bindZkNFTContract(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(ZkNFTContractABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ZkNFTContract *ZkNFTContractRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _ZkNFTContract.Contract.ZkNFTContractCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ZkNFTContract *ZkNFTContractRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ZkNFTContract.Contract.ZkNFTContractTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ZkNFTContract *ZkNFTContractRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ZkNFTContract.Contract.ZkNFTContractTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ZkNFTContract *ZkNFTContractCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _ZkNFTContract.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ZkNFTContract *ZkNFTContractTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ZkNFTContract.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ZkNFTContract *ZkNFTContractTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ZkNFTContract.Contract.contract.Transact(opts, method, params...)
}

// Anchors is a free data retrieval call binding the contract method 0x98d35f20.
//
// Solidity: function anchors() constant returns(address)
func (_ZkNFTContract *ZkNFTContractCaller) Anchors(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _ZkNFTContract.contract.Call(opts, out, "anchors")
	return *ret0, err
}

// Anchors is a free data retrieval call binding the contract method 0x98d35f20.
//
// Solidity: function anchors() constant returns(address)
func (_ZkNFTContract *ZkNFTContractSession) Anchors() (common.Address, error) {
	return _ZkNFTContract.Contract.Anchors(&_ZkNFTContract.CallOpts)
}

// Anchors is a free data retrieval call binding the contract method 0x98d35f20.
//
// Solidity: function anchors() constant returns(address)
func (_ZkNFTContract *ZkNFTContractCallerSession) Anchors() (common.Address, error) {
	return _ZkNFTContract.Contract.Anchors(&_ZkNFTContract.CallOpts)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(owner address) constant returns(uint256)
func (_ZkNFTContract *ZkNFTContractCaller) BalanceOf(opts *bind.CallOpts, owner common.Address) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _ZkNFTContract.contract.Call(opts, out, "balanceOf", owner)
	return *ret0, err
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(owner address) constant returns(uint256)
func (_ZkNFTContract *ZkNFTContractSession) BalanceOf(owner common.Address) (*big.Int, error) {
	return _ZkNFTContract.Contract.BalanceOf(&_ZkNFTContract.CallOpts, owner)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(owner address) constant returns(uint256)
func (_ZkNFTContract *ZkNFTContractCallerSession) BalanceOf(owner common.Address) (*big.Int, error) {
	return _ZkNFTContract.Contract.BalanceOf(&_ZkNFTContract.CallOpts, owner)
}

// Data is a free data retrieval call binding the contract method 0xf0ba8440.
//
// Solidity: function data( uint256) constant returns(amount uint256, anchor uint256, rating uint256)
func (_ZkNFTContract *ZkNFTContractCaller) Data(opts *bind.CallOpts, arg0 *big.Int) (struct {
	Amount *big.Int
	Anchor *big.Int
	Rating *big.Int
}, error) {
	ret := new(struct {
		Amount *big.Int
		Anchor *big.Int
		Rating *big.Int
	})
	out := ret
	err := _ZkNFTContract.contract.Call(opts, out, "data", arg0)
	return *ret, err
}

// Data is a free data retrieval call binding the contract method 0xf0ba8440.
//
// Solidity: function data( uint256) constant returns(amount uint256, anchor uint256, rating uint256)
func (_ZkNFTContract *ZkNFTContractSession) Data(arg0 *big.Int) (struct {
	Amount *big.Int
	Anchor *big.Int
	Rating *big.Int
}, error) {
	return _ZkNFTContract.Contract.Data(&_ZkNFTContract.CallOpts, arg0)
}

// Data is a free data retrieval call binding the contract method 0xf0ba8440.
//
// Solidity: function data( uint256) constant returns(amount uint256, anchor uint256, rating uint256)
func (_ZkNFTContract *ZkNFTContractCallerSession) Data(arg0 *big.Int) (struct {
	Amount *big.Int
	Anchor *big.Int
	Rating *big.Int
}, error) {
	return _ZkNFTContract.Contract.Data(&_ZkNFTContract.CallOpts, arg0)
}

// GetApproved is a free data retrieval call binding the contract method 0x081812fc.
//
// Solidity: function getApproved(tokenId uint256) constant returns(address)
func (_ZkNFTContract *ZkNFTContractCaller) GetApproved(opts *bind.CallOpts, tokenId *big.Int) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _ZkNFTContract.contract.Call(opts, out, "getApproved", tokenId)
	return *ret0, err
}

// GetApproved is a free data retrieval call binding the contract method 0x081812fc.
//
// Solidity: function getApproved(tokenId uint256) constant returns(address)
func (_ZkNFTContract *ZkNFTContractSession) GetApproved(tokenId *big.Int) (common.Address, error) {
	return _ZkNFTContract.Contract.GetApproved(&_ZkNFTContract.CallOpts, tokenId)
}

// GetApproved is a free data retrieval call binding the contract method 0x081812fc.
//
// Solidity: function getApproved(tokenId uint256) constant returns(address)
func (_ZkNFTContract *ZkNFTContractCallerSession) GetApproved(tokenId *big.Int) (common.Address, error) {
	return _ZkNFTContract.Contract.GetApproved(&_ZkNFTContract.CallOpts, tokenId)
}

// IsApprovedForAll is a free data retrieval call binding the contract method 0xe985e9c5.
//
// Solidity: function isApprovedForAll(owner address, operator address) constant returns(bool)
func (_ZkNFTContract *ZkNFTContractCaller) IsApprovedForAll(opts *bind.CallOpts, owner common.Address, operator common.Address) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _ZkNFTContract.contract.Call(opts, out, "isApprovedForAll", owner, operator)
	return *ret0, err
}

// IsApprovedForAll is a free data retrieval call binding the contract method 0xe985e9c5.
//
// Solidity: function isApprovedForAll(owner address, operator address) constant returns(bool)
func (_ZkNFTContract *ZkNFTContractSession) IsApprovedForAll(owner common.Address, operator common.Address) (bool, error) {
	return _ZkNFTContract.Contract.IsApprovedForAll(&_ZkNFTContract.CallOpts, owner, operator)
}

// IsApprovedForAll is a free data retrieval call binding the contract method 0xe985e9c5.
//
// Solidity: function isApprovedForAll(owner address, operator address) constant returns(bool)
func (_ZkNFTContract *ZkNFTContractCallerSession) IsApprovedForAll(owner common.Address, operator common.Address) (bool, error) {
	return _ZkNFTContract.Contract.IsApprovedForAll(&_ZkNFTContract.CallOpts, owner, operator)
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() constant returns(string)
func (_ZkNFTContract *ZkNFTContractCaller) Name(opts *bind.CallOpts) (string, error) {
	var (
		ret0 = new(string)
	)
	out := ret0
	err := _ZkNFTContract.contract.Call(opts, out, "name")
	return *ret0, err
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() constant returns(string)
func (_ZkNFTContract *ZkNFTContractSession) Name() (string, error) {
	return _ZkNFTContract.Contract.Name(&_ZkNFTContract.CallOpts)
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() constant returns(string)
func (_ZkNFTContract *ZkNFTContractCallerSession) Name() (string, error) {
	return _ZkNFTContract.Contract.Name(&_ZkNFTContract.CallOpts)
}

// OwnerOf is a free data retrieval call binding the contract method 0x6352211e.
//
// Solidity: function ownerOf(tokenId uint256) constant returns(address)
func (_ZkNFTContract *ZkNFTContractCaller) OwnerOf(opts *bind.CallOpts, tokenId *big.Int) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _ZkNFTContract.contract.Call(opts, out, "ownerOf", tokenId)
	return *ret0, err
}

// OwnerOf is a free data retrieval call binding the contract method 0x6352211e.
//
// Solidity: function ownerOf(tokenId uint256) constant returns(address)
func (_ZkNFTContract *ZkNFTContractSession) OwnerOf(tokenId *big.Int) (common.Address, error) {
	return _ZkNFTContract.Contract.OwnerOf(&_ZkNFTContract.CallOpts, tokenId)
}

// OwnerOf is a free data retrieval call binding the contract method 0x6352211e.
//
// Solidity: function ownerOf(tokenId uint256) constant returns(address)
func (_ZkNFTContract *ZkNFTContractCallerSession) OwnerOf(tokenId *big.Int) (common.Address, error) {
	return _ZkNFTContract.Contract.OwnerOf(&_ZkNFTContract.CallOpts, tokenId)
}

// Ratings is a free data retrieval call binding the contract method 0x36a9fc12.
//
// Solidity: function ratings() constant returns(bytes32)
func (_ZkNFTContract *ZkNFTContractCaller) Ratings(opts *bind.CallOpts) ([32]byte, error) {
	var (
		ret0 = new([32]byte)
	)
	out := ret0
	err := _ZkNFTContract.contract.Call(opts, out, "ratings")
	return *ret0, err
}

// Ratings is a free data retrieval call binding the contract method 0x36a9fc12.
//
// Solidity: function ratings() constant returns(bytes32)
func (_ZkNFTContract *ZkNFTContractSession) Ratings() ([32]byte, error) {
	return _ZkNFTContract.Contract.Ratings(&_ZkNFTContract.CallOpts)
}

// Ratings is a free data retrieval call binding the contract method 0x36a9fc12.
//
// Solidity: function ratings() constant returns(bytes32)
func (_ZkNFTContract *ZkNFTContractCallerSession) Ratings() ([32]byte, error) {
	return _ZkNFTContract.Contract.Ratings(&_ZkNFTContract.CallOpts)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(interfaceId bytes4) constant returns(bool)
func (_ZkNFTContract *ZkNFTContractCaller) SupportsInterface(opts *bind.CallOpts, interfaceId [4]byte) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _ZkNFTContract.contract.Call(opts, out, "supportsInterface", interfaceId)
	return *ret0, err
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(interfaceId bytes4) constant returns(bool)
func (_ZkNFTContract *ZkNFTContractSession) SupportsInterface(interfaceId [4]byte) (bool, error) {
	return _ZkNFTContract.Contract.SupportsInterface(&_ZkNFTContract.CallOpts, interfaceId)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(interfaceId bytes4) constant returns(bool)
func (_ZkNFTContract *ZkNFTContractCallerSession) SupportsInterface(interfaceId [4]byte) (bool, error) {
	return _ZkNFTContract.Contract.SupportsInterface(&_ZkNFTContract.CallOpts, interfaceId)
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() constant returns(string)
func (_ZkNFTContract *ZkNFTContractCaller) Symbol(opts *bind.CallOpts) (string, error) {
	var (
		ret0 = new(string)
	)
	out := ret0
	err := _ZkNFTContract.contract.Call(opts, out, "symbol")
	return *ret0, err
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() constant returns(string)
func (_ZkNFTContract *ZkNFTContractSession) Symbol() (string, error) {
	return _ZkNFTContract.Contract.Symbol(&_ZkNFTContract.CallOpts)
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() constant returns(string)
func (_ZkNFTContract *ZkNFTContractCallerSession) Symbol() (string, error) {
	return _ZkNFTContract.Contract.Symbol(&_ZkNFTContract.CallOpts)
}

// TokenByIndex is a free data retrieval call binding the contract method 0x4f6ccce7.
//
// Solidity: function tokenByIndex(index uint256) constant returns(uint256)
func (_ZkNFTContract *ZkNFTContractCaller) TokenByIndex(opts *bind.CallOpts, index *big.Int) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _ZkNFTContract.contract.Call(opts, out, "tokenByIndex", index)
	return *ret0, err
}

// TokenByIndex is a free data retrieval call binding the contract method 0x4f6ccce7.
//
// Solidity: function tokenByIndex(index uint256) constant returns(uint256)
func (_ZkNFTContract *ZkNFTContractSession) TokenByIndex(index *big.Int) (*big.Int, error) {
	return _ZkNFTContract.Contract.TokenByIndex(&_ZkNFTContract.CallOpts, index)
}

// TokenByIndex is a free data retrieval call binding the contract method 0x4f6ccce7.
//
// Solidity: function tokenByIndex(index uint256) constant returns(uint256)
func (_ZkNFTContract *ZkNFTContractCallerSession) TokenByIndex(index *big.Int) (*big.Int, error) {
	return _ZkNFTContract.Contract.TokenByIndex(&_ZkNFTContract.CallOpts, index)
}

// TokenOfOwnerByIndex is a free data retrieval call binding the contract method 0x2f745c59.
//
// Solidity: function tokenOfOwnerByIndex(owner address, index uint256) constant returns(uint256)
func (_ZkNFTContract *ZkNFTContractCaller) TokenOfOwnerByIndex(opts *bind.CallOpts, owner common.Address, index *big.Int) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _ZkNFTContract.contract.Call(opts, out, "tokenOfOwnerByIndex", owner, index)
	return *ret0, err
}

// TokenOfOwnerByIndex is a free data retrieval call binding the contract method 0x2f745c59.
//
// Solidity: function tokenOfOwnerByIndex(owner address, index uint256) constant returns(uint256)
func (_ZkNFTContract *ZkNFTContractSession) TokenOfOwnerByIndex(owner common.Address, index *big.Int) (*big.Int, error) {
	return _ZkNFTContract.Contract.TokenOfOwnerByIndex(&_ZkNFTContract.CallOpts, owner, index)
}

// TokenOfOwnerByIndex is a free data retrieval call binding the contract method 0x2f745c59.
//
// Solidity: function tokenOfOwnerByIndex(owner address, index uint256) constant returns(uint256)
func (_ZkNFTContract *ZkNFTContractCallerSession) TokenOfOwnerByIndex(owner common.Address, index *big.Int) (*big.Int, error) {
	return _ZkNFTContract.Contract.TokenOfOwnerByIndex(&_ZkNFTContract.CallOpts, owner, index)
}

// TokenURI is a free data retrieval call binding the contract method 0xc87b56dd.
//
// Solidity: function tokenURI(tokenId uint256) constant returns(string)
func (_ZkNFTContract *ZkNFTContractCaller) TokenURI(opts *bind.CallOpts, tokenId *big.Int) (string, error) {
	var (
		ret0 = new(string)
	)
	out := ret0
	err := _ZkNFTContract.contract.Call(opts, out, "tokenURI", tokenId)
	return *ret0, err
}

// TokenURI is a free data retrieval call binding the contract method 0xc87b56dd.
//
// Solidity: function tokenURI(tokenId uint256) constant returns(string)
func (_ZkNFTContract *ZkNFTContractSession) TokenURI(tokenId *big.Int) (string, error) {
	return _ZkNFTContract.Contract.TokenURI(&_ZkNFTContract.CallOpts, tokenId)
}

// TokenURI is a free data retrieval call binding the contract method 0xc87b56dd.
//
// Solidity: function tokenURI(tokenId uint256) constant returns(string)
func (_ZkNFTContract *ZkNFTContractCallerSession) TokenURI(tokenId *big.Int) (string, error) {
	return _ZkNFTContract.Contract.TokenURI(&_ZkNFTContract.CallOpts, tokenId)
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() constant returns(uint256)
func (_ZkNFTContract *ZkNFTContractCaller) TotalSupply(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _ZkNFTContract.contract.Call(opts, out, "totalSupply")
	return *ret0, err
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() constant returns(uint256)
func (_ZkNFTContract *ZkNFTContractSession) TotalSupply() (*big.Int, error) {
	return _ZkNFTContract.Contract.TotalSupply(&_ZkNFTContract.CallOpts)
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() constant returns(uint256)
func (_ZkNFTContract *ZkNFTContractCallerSession) TotalSupply() (*big.Int, error) {
	return _ZkNFTContract.Contract.TotalSupply(&_ZkNFTContract.CallOpts)
}

// Uri is a free data retrieval call binding the contract method 0xeac989f8.
//
// Solidity: function uri() constant returns(string)
func (_ZkNFTContract *ZkNFTContractCaller) Uri(opts *bind.CallOpts) (string, error) {
	var (
		ret0 = new(string)
	)
	out := ret0
	err := _ZkNFTContract.contract.Call(opts, out, "uri")
	return *ret0, err
}

// Uri is a free data retrieval call binding the contract method 0xeac989f8.
//
// Solidity: function uri() constant returns(string)
func (_ZkNFTContract *ZkNFTContractSession) Uri() (string, error) {
	return _ZkNFTContract.Contract.Uri(&_ZkNFTContract.CallOpts)
}

// Uri is a free data retrieval call binding the contract method 0xeac989f8.
//
// Solidity: function uri() constant returns(string)
func (_ZkNFTContract *ZkNFTContractCallerSession) Uri() (string, error) {
	return _ZkNFTContract.Contract.Uri(&_ZkNFTContract.CallOpts)
}

// UriPrefix is a free data retrieval call binding the contract method 0x08f2fc29.
//
// Solidity: function uri_prefix() constant returns(string)
func (_ZkNFTContract *ZkNFTContractCaller) UriPrefix(opts *bind.CallOpts) (string, error) {
	var (
		ret0 = new(string)
	)
	out := ret0
	err := _ZkNFTContract.contract.Call(opts, out, "uri_prefix")
	return *ret0, err
}

// UriPrefix is a free data retrieval call binding the contract method 0x08f2fc29.
//
// Solidity: function uri_prefix() constant returns(string)
func (_ZkNFTContract *ZkNFTContractSession) UriPrefix() (string, error) {
	return _ZkNFTContract.Contract.UriPrefix(&_ZkNFTContract.CallOpts)
}

// UriPrefix is a free data retrieval call binding the contract method 0x08f2fc29.
//
// Solidity: function uri_prefix() constant returns(string)
func (_ZkNFTContract *ZkNFTContractCallerSession) UriPrefix() (string, error) {
	return _ZkNFTContract.Contract.UriPrefix(&_ZkNFTContract.CallOpts)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(to address, tokenId uint256) returns()
func (_ZkNFTContract *ZkNFTContractTransactor) Approve(opts *bind.TransactOpts, to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _ZkNFTContract.contract.Transact(opts, "approve", to, tokenId)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(to address, tokenId uint256) returns()
func (_ZkNFTContract *ZkNFTContractSession) Approve(to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _ZkNFTContract.Contract.Approve(&_ZkNFTContract.TransactOpts, to, tokenId)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(to address, tokenId uint256) returns()
func (_ZkNFTContract *ZkNFTContractTransactorSession) Approve(to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _ZkNFTContract.Contract.Approve(&_ZkNFTContract.TransactOpts, to, tokenId)
}

// CheckAnchor is a paid mutator transaction binding the contract method 0xf8615b0d.
//
// Solidity: function checkAnchor(anchor uint256, droot bytes32, sigs bytes32) returns(bool)
func (_ZkNFTContract *ZkNFTContractTransactor) CheckAnchor(opts *bind.TransactOpts, anchor *big.Int, droot [32]byte, sigs [32]byte) (*types.Transaction, error) {
	return _ZkNFTContract.contract.Transact(opts, "checkAnchor", anchor, droot, sigs)
}

// CheckAnchor is a paid mutator transaction binding the contract method 0xf8615b0d.
//
// Solidity: function checkAnchor(anchor uint256, droot bytes32, sigs bytes32) returns(bool)
func (_ZkNFTContract *ZkNFTContractSession) CheckAnchor(anchor *big.Int, droot [32]byte, sigs [32]byte) (*types.Transaction, error) {
	return _ZkNFTContract.Contract.CheckAnchor(&_ZkNFTContract.TransactOpts, anchor, droot, sigs)
}

// CheckAnchor is a paid mutator transaction binding the contract method 0xf8615b0d.
//
// Solidity: function checkAnchor(anchor uint256, droot bytes32, sigs bytes32) returns(bool)
func (_ZkNFTContract *ZkNFTContractTransactorSession) CheckAnchor(anchor *big.Int, droot [32]byte, sigs [32]byte) (*types.Transaction, error) {
	return _ZkNFTContract.Contract.CheckAnchor(&_ZkNFTContract.TransactOpts, anchor, droot, sigs)
}

// File is a paid mutator transaction binding the contract method 0xe9b674b9.
//
// Solidity: function file(what bytes32, data_ bytes32) returns()
func (_ZkNFTContract *ZkNFTContractTransactor) File(opts *bind.TransactOpts, what [32]byte, data_ [32]byte) (*types.Transaction, error) {
	return _ZkNFTContract.contract.Transact(opts, "file", what, data_)
}

// File is a paid mutator transaction binding the contract method 0xe9b674b9.
//
// Solidity: function file(what bytes32, data_ bytes32) returns()
func (_ZkNFTContract *ZkNFTContractSession) File(what [32]byte, data_ [32]byte) (*types.Transaction, error) {
	return _ZkNFTContract.Contract.File(&_ZkNFTContract.TransactOpts, what, data_)
}

// File is a paid mutator transaction binding the contract method 0xe9b674b9.
//
// Solidity: function file(what bytes32, data_ bytes32) returns()
func (_ZkNFTContract *ZkNFTContractTransactorSession) File(what [32]byte, data_ [32]byte) (*types.Transaction, error) {
	return _ZkNFTContract.Contract.File(&_ZkNFTContract.TransactOpts, what, data_)
}

// Mint is a paid mutator transaction binding the contract method 0xce4ea0c7.
//
// Solidity: function mint(usr address, tkn uint256, anchor uint256, data_root bytes32, signatures_root bytes32, amount uint256, rating uint256, points uint256[8]) returns(uint256)
func (_ZkNFTContract *ZkNFTContractTransactor) Mint(opts *bind.TransactOpts, usr common.Address, tkn *big.Int, anchor *big.Int, data_root [32]byte, signatures_root [32]byte, amount *big.Int, rating *big.Int, points [8]*big.Int) (*types.Transaction, error) {
	return _ZkNFTContract.contract.Transact(opts, "mint", usr, tkn, anchor, data_root, signatures_root, amount, rating, points)
}

// Mint is a paid mutator transaction binding the contract method 0xce4ea0c7.
//
// Solidity: function mint(usr address, tkn uint256, anchor uint256, data_root bytes32, signatures_root bytes32, amount uint256, rating uint256, points uint256[8]) returns(uint256)
func (_ZkNFTContract *ZkNFTContractSession) Mint(usr common.Address, tkn *big.Int, anchor *big.Int, data_root [32]byte, signatures_root [32]byte, amount *big.Int, rating *big.Int, points [8]*big.Int) (*types.Transaction, error) {
	return _ZkNFTContract.Contract.Mint(&_ZkNFTContract.TransactOpts, usr, tkn, anchor, data_root, signatures_root, amount, rating, points)
}

// Mint is a paid mutator transaction binding the contract method 0xce4ea0c7.
//
// Solidity: function mint(usr address, tkn uint256, anchor uint256, data_root bytes32, signatures_root bytes32, amount uint256, rating uint256, points uint256[8]) returns(uint256)
func (_ZkNFTContract *ZkNFTContractTransactorSession) Mint(usr common.Address, tkn *big.Int, anchor *big.Int, data_root [32]byte, signatures_root [32]byte, amount *big.Int, rating *big.Int, points [8]*big.Int) (*types.Transaction, error) {
	return _ZkNFTContract.Contract.Mint(&_ZkNFTContract.TransactOpts, usr, tkn, anchor, data_root, signatures_root, amount, rating, points)
}

// SafeTransferFrom is a paid mutator transaction binding the contract method 0xb88d4fde.
//
// Solidity: function safeTransferFrom(from address, to address, tokenId uint256, _data bytes) returns()
func (_ZkNFTContract *ZkNFTContractTransactor) SafeTransferFrom(opts *bind.TransactOpts, from common.Address, to common.Address, tokenId *big.Int, _data []byte) (*types.Transaction, error) {
	return _ZkNFTContract.contract.Transact(opts, "safeTransferFrom", from, to, tokenId, _data)
}

// SafeTransferFrom is a paid mutator transaction binding the contract method 0xb88d4fde.
//
// Solidity: function safeTransferFrom(from address, to address, tokenId uint256, _data bytes) returns()
func (_ZkNFTContract *ZkNFTContractSession) SafeTransferFrom(from common.Address, to common.Address, tokenId *big.Int, _data []byte) (*types.Transaction, error) {
	return _ZkNFTContract.Contract.SafeTransferFrom(&_ZkNFTContract.TransactOpts, from, to, tokenId, _data)
}

// SafeTransferFrom is a paid mutator transaction binding the contract method 0xb88d4fde.
//
// Solidity: function safeTransferFrom(from address, to address, tokenId uint256, _data bytes) returns()
func (_ZkNFTContract *ZkNFTContractTransactorSession) SafeTransferFrom(from common.Address, to common.Address, tokenId *big.Int, _data []byte) (*types.Transaction, error) {
	return _ZkNFTContract.Contract.SafeTransferFrom(&_ZkNFTContract.TransactOpts, from, to, tokenId, _data)
}

// SetApprovalForAll is a paid mutator transaction binding the contract method 0xa22cb465.
//
// Solidity: function setApprovalForAll(to address, approved bool) returns()
func (_ZkNFTContract *ZkNFTContractTransactor) SetApprovalForAll(opts *bind.TransactOpts, to common.Address, approved bool) (*types.Transaction, error) {
	return _ZkNFTContract.contract.Transact(opts, "setApprovalForAll", to, approved)
}

// SetApprovalForAll is a paid mutator transaction binding the contract method 0xa22cb465.
//
// Solidity: function setApprovalForAll(to address, approved bool) returns()
func (_ZkNFTContract *ZkNFTContractSession) SetApprovalForAll(to common.Address, approved bool) (*types.Transaction, error) {
	return _ZkNFTContract.Contract.SetApprovalForAll(&_ZkNFTContract.TransactOpts, to, approved)
}

// SetApprovalForAll is a paid mutator transaction binding the contract method 0xa22cb465.
//
// Solidity: function setApprovalForAll(to address, approved bool) returns()
func (_ZkNFTContract *ZkNFTContractTransactorSession) SetApprovalForAll(to common.Address, approved bool) (*types.Transaction, error) {
	return _ZkNFTContract.Contract.SetApprovalForAll(&_ZkNFTContract.TransactOpts, to, approved)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(from address, to address, tokenId uint256) returns()
func (_ZkNFTContract *ZkNFTContractTransactor) TransferFrom(opts *bind.TransactOpts, from common.Address, to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _ZkNFTContract.contract.Transact(opts, "transferFrom", from, to, tokenId)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(from address, to address, tokenId uint256) returns()
func (_ZkNFTContract *ZkNFTContractSession) TransferFrom(from common.Address, to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _ZkNFTContract.Contract.TransferFrom(&_ZkNFTContract.TransactOpts, from, to, tokenId)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(from address, to address, tokenId uint256) returns()
func (_ZkNFTContract *ZkNFTContractTransactorSession) TransferFrom(from common.Address, to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _ZkNFTContract.Contract.TransferFrom(&_ZkNFTContract.TransactOpts, from, to, tokenId)
}

// Unpack is a paid mutator transaction binding the contract method 0x71516dd9.
//
// Solidity: function unpack(x bytes32) returns(y uint256, z uint256)
func (_ZkNFTContract *ZkNFTContractTransactor) Unpack(opts *bind.TransactOpts, x [32]byte) (*types.Transaction, error) {
	return _ZkNFTContract.contract.Transact(opts, "unpack", x)
}

// Unpack is a paid mutator transaction binding the contract method 0x71516dd9.
//
// Solidity: function unpack(x bytes32) returns(y uint256, z uint256)
func (_ZkNFTContract *ZkNFTContractSession) Unpack(x [32]byte) (*types.Transaction, error) {
	return _ZkNFTContract.Contract.Unpack(&_ZkNFTContract.TransactOpts, x)
}

// Unpack is a paid mutator transaction binding the contract method 0x71516dd9.
//
// Solidity: function unpack(x bytes32) returns(y uint256, z uint256)
func (_ZkNFTContract *ZkNFTContractTransactorSession) Unpack(x [32]byte) (*types.Transaction, error) {
	return _ZkNFTContract.Contract.Unpack(&_ZkNFTContract.TransactOpts, x)
}

// Verify is a paid mutator transaction binding the contract method 0xe19e62ce.
//
// Solidity: function verify(data_root bytes32, nft_amount uint256, rating uint256, points uint256[8]) returns()
func (_ZkNFTContract *ZkNFTContractTransactor) Verify(opts *bind.TransactOpts, data_root [32]byte, nft_amount *big.Int, rating *big.Int, points [8]*big.Int) (*types.Transaction, error) {
	return _ZkNFTContract.contract.Transact(opts, "verify", data_root, nft_amount, rating, points)
}

// Verify is a paid mutator transaction binding the contract method 0xe19e62ce.
//
// Solidity: function verify(data_root bytes32, nft_amount uint256, rating uint256, points uint256[8]) returns()
func (_ZkNFTContract *ZkNFTContractSession) Verify(data_root [32]byte, nft_amount *big.Int, rating *big.Int, points [8]*big.Int) (*types.Transaction, error) {
	return _ZkNFTContract.Contract.Verify(&_ZkNFTContract.TransactOpts, data_root, nft_amount, rating, points)
}

// Verify is a paid mutator transaction binding the contract method 0xe19e62ce.
//
// Solidity: function verify(data_root bytes32, nft_amount uint256, rating uint256, points uint256[8]) returns()
func (_ZkNFTContract *ZkNFTContractTransactorSession) Verify(data_root [32]byte, nft_amount *big.Int, rating *big.Int, points [8]*big.Int) (*types.Transaction, error) {
	return _ZkNFTContract.Contract.Verify(&_ZkNFTContract.TransactOpts, data_root, nft_amount, rating, points)
}

// VerifyTx is a paid mutator transaction binding the contract method 0x46d8653f.
//
// Solidity: function verifyTx(a uint256[2], b uint256[2][2], c uint256[2], input uint256[7]) returns(r bool)
func (_ZkNFTContract *ZkNFTContractTransactor) VerifyTx(opts *bind.TransactOpts, a [2]*big.Int, b [2][2]*big.Int, c [2]*big.Int, input [7]*big.Int) (*types.Transaction, error) {
	return _ZkNFTContract.contract.Transact(opts, "verifyTx", a, b, c, input)
}

// VerifyTx is a paid mutator transaction binding the contract method 0x46d8653f.
//
// Solidity: function verifyTx(a uint256[2], b uint256[2][2], c uint256[2], input uint256[7]) returns(r bool)
func (_ZkNFTContract *ZkNFTContractSession) VerifyTx(a [2]*big.Int, b [2][2]*big.Int, c [2]*big.Int, input [7]*big.Int) (*types.Transaction, error) {
	return _ZkNFTContract.Contract.VerifyTx(&_ZkNFTContract.TransactOpts, a, b, c, input)
}

// VerifyTx is a paid mutator transaction binding the contract method 0x46d8653f.
//
// Solidity: function verifyTx(a uint256[2], b uint256[2][2], c uint256[2], input uint256[7]) returns(r bool)
func (_ZkNFTContract *ZkNFTContractTransactorSession) VerifyTx(a [2]*big.Int, b [2][2]*big.Int, c [2]*big.Int, input [7]*big.Int) (*types.Transaction, error) {
	return _ZkNFTContract.Contract.VerifyTx(&_ZkNFTContract.TransactOpts, a, b, c, input)
}

// ZkNFTContractApprovalIterator is returned from FilterApproval and is used to iterate over the raw logs and unpacked data for Approval events raised by the ZkNFTContract contract.
type ZkNFTContractApprovalIterator struct {
	Event *ZkNFTContractApproval // Event containing the contract specifics and raw log

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
func (it *ZkNFTContractApprovalIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ZkNFTContractApproval)
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
		it.Event = new(ZkNFTContractApproval)
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
func (it *ZkNFTContractApprovalIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ZkNFTContractApprovalIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ZkNFTContractApproval represents a Approval event raised by the ZkNFTContract contract.
type ZkNFTContractApproval struct {
	Owner    common.Address
	Approved common.Address
	TokenId  *big.Int
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterApproval is a free log retrieval operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: e Approval(owner indexed address, approved indexed address, tokenId indexed uint256)
func (_ZkNFTContract *ZkNFTContractFilterer) FilterApproval(opts *bind.FilterOpts, owner []common.Address, approved []common.Address, tokenId []*big.Int) (*ZkNFTContractApprovalIterator, error) {

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

	logs, sub, err := _ZkNFTContract.contract.FilterLogs(opts, "Approval", ownerRule, approvedRule, tokenIdRule)
	if err != nil {
		return nil, err
	}
	return &ZkNFTContractApprovalIterator{contract: _ZkNFTContract.contract, event: "Approval", logs: logs, sub: sub}, nil
}

// WatchApproval is a free log subscription operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: e Approval(owner indexed address, approved indexed address, tokenId indexed uint256)
func (_ZkNFTContract *ZkNFTContractFilterer) WatchApproval(opts *bind.WatchOpts, sink chan<- *ZkNFTContractApproval, owner []common.Address, approved []common.Address, tokenId []*big.Int) (event.Subscription, error) {

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

	logs, sub, err := _ZkNFTContract.contract.WatchLogs(opts, "Approval", ownerRule, approvedRule, tokenIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ZkNFTContractApproval)
				if err := _ZkNFTContract.contract.UnpackLog(event, "Approval", log); err != nil {
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

// ZkNFTContractApprovalForAllIterator is returned from FilterApprovalForAll and is used to iterate over the raw logs and unpacked data for ApprovalForAll events raised by the ZkNFTContract contract.
type ZkNFTContractApprovalForAllIterator struct {
	Event *ZkNFTContractApprovalForAll // Event containing the contract specifics and raw log

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
func (it *ZkNFTContractApprovalForAllIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ZkNFTContractApprovalForAll)
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
		it.Event = new(ZkNFTContractApprovalForAll)
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
func (it *ZkNFTContractApprovalForAllIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ZkNFTContractApprovalForAllIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ZkNFTContractApprovalForAll represents a ApprovalForAll event raised by the ZkNFTContract contract.
type ZkNFTContractApprovalForAll struct {
	Owner    common.Address
	Operator common.Address
	Approved bool
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterApprovalForAll is a free log retrieval operation binding the contract event 0x17307eab39ab6107e8899845ad3d59bd9653f200f220920489ca2b5937696c31.
//
// Solidity: e ApprovalForAll(owner indexed address, operator indexed address, approved bool)
func (_ZkNFTContract *ZkNFTContractFilterer) FilterApprovalForAll(opts *bind.FilterOpts, owner []common.Address, operator []common.Address) (*ZkNFTContractApprovalForAllIterator, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}

	logs, sub, err := _ZkNFTContract.contract.FilterLogs(opts, "ApprovalForAll", ownerRule, operatorRule)
	if err != nil {
		return nil, err
	}
	return &ZkNFTContractApprovalForAllIterator{contract: _ZkNFTContract.contract, event: "ApprovalForAll", logs: logs, sub: sub}, nil
}

// WatchApprovalForAll is a free log subscription operation binding the contract event 0x17307eab39ab6107e8899845ad3d59bd9653f200f220920489ca2b5937696c31.
//
// Solidity: e ApprovalForAll(owner indexed address, operator indexed address, approved bool)
func (_ZkNFTContract *ZkNFTContractFilterer) WatchApprovalForAll(opts *bind.WatchOpts, sink chan<- *ZkNFTContractApprovalForAll, owner []common.Address, operator []common.Address) (event.Subscription, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}

	logs, sub, err := _ZkNFTContract.contract.WatchLogs(opts, "ApprovalForAll", ownerRule, operatorRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ZkNFTContractApprovalForAll)
				if err := _ZkNFTContract.contract.UnpackLog(event, "ApprovalForAll", log); err != nil {
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

// ZkNFTContractTransferIterator is returned from FilterTransfer and is used to iterate over the raw logs and unpacked data for Transfer events raised by the ZkNFTContract contract.
type ZkNFTContractTransferIterator struct {
	Event *ZkNFTContractTransfer // Event containing the contract specifics and raw log

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
func (it *ZkNFTContractTransferIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ZkNFTContractTransfer)
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
		it.Event = new(ZkNFTContractTransfer)
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
func (it *ZkNFTContractTransferIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ZkNFTContractTransferIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ZkNFTContractTransfer represents a Transfer event raised by the ZkNFTContract contract.
type ZkNFTContractTransfer struct {
	From    common.Address
	To      common.Address
	TokenId *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterTransfer is a free log retrieval operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: e Transfer(from indexed address, to indexed address, tokenId indexed uint256)
func (_ZkNFTContract *ZkNFTContractFilterer) FilterTransfer(opts *bind.FilterOpts, from []common.Address, to []common.Address, tokenId []*big.Int) (*ZkNFTContractTransferIterator, error) {

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

	logs, sub, err := _ZkNFTContract.contract.FilterLogs(opts, "Transfer", fromRule, toRule, tokenIdRule)
	if err != nil {
		return nil, err
	}
	return &ZkNFTContractTransferIterator{contract: _ZkNFTContract.contract, event: "Transfer", logs: logs, sub: sub}, nil
}

// WatchTransfer is a free log subscription operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: e Transfer(from indexed address, to indexed address, tokenId indexed uint256)
func (_ZkNFTContract *ZkNFTContractFilterer) WatchTransfer(opts *bind.WatchOpts, sink chan<- *ZkNFTContractTransfer, from []common.Address, to []common.Address, tokenId []*big.Int) (event.Subscription, error) {

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

	logs, sub, err := _ZkNFTContract.contract.WatchLogs(opts, "Transfer", fromRule, toRule, tokenIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ZkNFTContractTransfer)
				if err := _ZkNFTContract.contract.UnpackLog(event, "Transfer", log); err != nil {
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

// ZkNFTContractVerifiedIterator is returned from FilterVerified and is used to iterate over the raw logs and unpacked data for Verified events raised by the ZkNFTContract contract.
type ZkNFTContractVerifiedIterator struct {
	Event *ZkNFTContractVerified // Event containing the contract specifics and raw log

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
func (it *ZkNFTContractVerifiedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ZkNFTContractVerified)
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
		it.Event = new(ZkNFTContractVerified)
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
func (it *ZkNFTContractVerifiedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ZkNFTContractVerifiedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ZkNFTContractVerified represents a Verified event raised by the ZkNFTContract contract.
type ZkNFTContractVerified struct {
	S   string
	Raw types.Log // Blockchain specific contextual infos
}

// FilterVerified is a free log retrieval operation binding the contract event 0x3f3cfdb26fb5f9f1786ab4f1a1f9cd4c0b5e726cbdfc26e495261731aad44e39.
//
// Solidity: e Verified(s string)
func (_ZkNFTContract *ZkNFTContractFilterer) FilterVerified(opts *bind.FilterOpts) (*ZkNFTContractVerifiedIterator, error) {

	logs, sub, err := _ZkNFTContract.contract.FilterLogs(opts, "Verified")
	if err != nil {
		return nil, err
	}
	return &ZkNFTContractVerifiedIterator{contract: _ZkNFTContract.contract, event: "Verified", logs: logs, sub: sub}, nil
}

// WatchVerified is a free log subscription operation binding the contract event 0x3f3cfdb26fb5f9f1786ab4f1a1f9cd4c0b5e726cbdfc26e495261731aad44e39.
//
// Solidity: e Verified(s string)
func (_ZkNFTContract *ZkNFTContractFilterer) WatchVerified(opts *bind.WatchOpts, sink chan<- *ZkNFTContractVerified) (event.Subscription, error) {

	logs, sub, err := _ZkNFTContract.contract.WatchLogs(opts, "Verified")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ZkNFTContractVerified)
				if err := _ZkNFTContract.contract.UnpackLog(event, "Verified", log); err != nil {
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
