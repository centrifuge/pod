// +build unit

package ethid

import (
	"math/big"
	"testing"

	"github.com/centrifuge/go-centrifuge/testingutils/commons"

	"github.com/centrifuge/go-centrifuge/identity"

	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/ethereum"
	"github.com/centrifuge/go-centrifuge/testingutils/config"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockIDFactory struct {
	mock.Mock
}

func (f MockIDFactory) CreateIdentity(opts *bind.TransactOpts, _centrifugeId *big.Int) (*types.Transaction, error) {
	args := f.Called(opts, _centrifugeId)
	id := args.Get(0).(*types.Transaction)
	return id, args.Error(1)
}

type MockIDRegistry struct {
	mock.Mock
}

func (r MockIDRegistry) GetIdentityByCentrifugeId(opts *bind.CallOpts, bigInt *big.Int) (common.Address, error) {
	args := r.Called(opts, bigInt)
	id := args.Get(0).(common.Address)
	return id, args.Error(1)
}

type MockIDContract struct {
	mock.Mock
}

func (mic MockIDContract) AddKey(opts *bind.TransactOpts, _key [32]byte, _kPurpose *big.Int) (*types.Transaction, error) {
	args := mic.Called(opts, _key, _kPurpose)
	return args.Get(0).(*types.Transaction), args.Error(1)
}

func (mic MockIDContract) GetKeysByPurpose(opts *bind.CallOpts, _purpose *big.Int) ([][32]byte, error) {
	args := mic.Called(opts, _purpose)
	return args.Get(0).([][32]byte), args.Error(1)
}

func (mic MockIDContract) GetKey(opts *bind.CallOpts, _key [32]byte) (struct {
	Key       [32]byte
	Purposes  []*big.Int
	RevokedAt *big.Int
}, error) {
	args := mic.Called(opts, _key)
	return args.Get(0).(struct {
		Key       [32]byte
		Purposes  []*big.Int
		RevokedAt *big.Int
	}), args.Error(1)
}

func (mic MockIDContract) FilterKeyAdded(opts *bind.FilterOpts, key [][32]byte, purpose []*big.Int) (*EthereumIdentityContractKeyAddedIterator, error) {
	args := mic.Called(opts, key, purpose)
	return args.Get(0).(*EthereumIdentityContractKeyAddedIterator), args.Error(1)
}

func TestGetClientP2PURL_happy(t *testing.T) {
	centID, _ := identity.ToCentID(utils.RandomSlice(identity.CentIDLength))
	c, f, r, g, i := &testingconfig.MockConfig{}, &MockIDFactory{}, &MockIDRegistry{}, &testingcommons.MockEthClient{}, &MockIDContract{}
	g.On("GetGethCallOpts").Return(&bind.CallOpts{}, func() {})
	g.On("GetEthClient").Return(&ethclient.Client{})
	r.On("GetIdentityByCentrifugeId", mock.Anything, centID.BigInt()).Return(common.BytesToAddress(utils.RandomSlice(20)), nil)
	i.On("GetKeysByPurpose", mock.Anything, mock.Anything).Return([][32]byte{{1}}, nil)
	srv := NewEthereumIdentityService(c, f, r, nil, func() ethereum.Client {
		return g
	}, func(address common.Address, backend bind.ContractBackend) (contract, error) {
		return i, nil
	})
	p2p, err := srv.GetClientP2PURL(centID)
	f.AssertExpectations(t)
	r.AssertExpectations(t)
	g.AssertExpectations(t)
	assert.NotEmpty(t, p2p, "p2p url is empty")
	assert.Nil(t, err, "error should be nil")
}

func TestGetClientP2PURL_fail_identity_lookup(t *testing.T) {
	centID, _ := identity.ToCentID(utils.RandomSlice(identity.CentIDLength))
	c, f, r, g, i := &testingconfig.MockConfig{}, &MockIDFactory{}, &MockIDRegistry{}, &testingcommons.MockEthClient{}, &MockIDContract{}
	g.On("GetGethCallOpts").Return(&bind.CallOpts{}, func() {})
	r.On("GetIdentityByCentrifugeId", mock.Anything, centID.BigInt()).Return(common.BytesToAddress(utils.RandomSlice(20)), errors.New("ID lookup failed"))
	i.On("GetKeysByPurpose", mock.Anything, mock.Anything).Return([][32]byte{{1}}, nil)
	srv := NewEthereumIdentityService(c, f, r, nil, func() ethereum.Client {
		return g
	}, func(address common.Address, backend bind.ContractBackend) (contract, error) {
		return i, nil
	})
	p2p, err := srv.GetClientP2PURL(centID)
	f.AssertExpectations(t)
	r.AssertExpectations(t)
	g.AssertExpectations(t)
	assert.Empty(t, p2p, "p2p is not empty")
	assert.Errorf(t, err, "error should not be nil")
	assert.Contains(t, err.Error(), "ID lookup failed")
}

func TestGetClientP2PURL_fail_p2pkey_error(t *testing.T) {
	centID, _ := identity.ToCentID(utils.RandomSlice(identity.CentIDLength))
	c, f, r, g, i := &testingconfig.MockConfig{}, &MockIDFactory{}, &MockIDRegistry{}, &testingcommons.MockEthClient{}, &MockIDContract{}
	g.On("GetGethCallOpts").Return(&bind.CallOpts{}, func() {})
	g.On("GetEthClient").Return(&ethclient.Client{})
	r.On("GetIdentityByCentrifugeId", mock.Anything, centID.BigInt()).Return(common.BytesToAddress(utils.RandomSlice(20)), nil)
	i.On("GetKeysByPurpose", mock.Anything, mock.Anything).Return([][32]byte{{1}}, errors.New("p2p key error"))
	srv := NewEthereumIdentityService(c, f, r, nil, func() ethereum.Client {
		return g
	}, func(address common.Address, backend bind.ContractBackend) (contract, error) {
		return i, nil
	})
	p2p, err := srv.GetClientP2PURL(centID)
	f.AssertExpectations(t)
	r.AssertExpectations(t)
	g.AssertExpectations(t)
	assert.Empty(t, p2p, "p2p is not empty")
	assert.Errorf(t, err, "error should not be nil")
	assert.Contains(t, err.Error(), "p2p key error")
}

func TestGetIdentityKey_fail_lookup(t *testing.T) {
	centID, _ := identity.ToCentID(utils.RandomSlice(identity.CentIDLength))
	pubKey := utils.RandomSlice(32)
	c, f, r, g, i := &testingconfig.MockConfig{}, &MockIDFactory{}, &MockIDRegistry{}, &testingcommons.MockEthClient{}, &MockIDContract{}
	g.On("GetGethCallOpts").Return(&bind.CallOpts{}, func() {})
	r.On("GetIdentityByCentrifugeId", mock.Anything, centID.BigInt()).Return(common.BytesToAddress(utils.RandomSlice(20)), errors.New("ID lookup failed"))
	srv := NewEthereumIdentityService(c, f, r, nil, func() ethereum.Client {
		return g
	}, func(address common.Address, backend bind.ContractBackend) (contract, error) {
		return i, nil
	})
	p2p, err := srv.GetIdentityKey(centID, pubKey)
	f.AssertExpectations(t)
	r.AssertExpectations(t)
	g.AssertExpectations(t)
	assert.Empty(t, p2p, "p2p is not empty")
	assert.Errorf(t, err, "error should not be nil")
	assert.Contains(t, err.Error(), "ID lookup failed")
}

func TestGetIdentityKey_fail_FetchKey(t *testing.T) {
	centID, _ := identity.ToCentID(utils.RandomSlice(identity.CentIDLength))
	pubKey := utils.RandomSlice(32)
	c, f, r, g, i := &testingconfig.MockConfig{}, &MockIDFactory{}, &MockIDRegistry{}, &testingcommons.MockEthClient{}, &MockIDContract{}
	g.On("GetGethCallOpts").Return(&bind.CallOpts{}, func() {})
	g.On("GetEthClient").Return(&ethclient.Client{})
	r.On("GetIdentityByCentrifugeId", mock.Anything, centID.BigInt()).Return(common.BytesToAddress(utils.RandomSlice(20)), nil)
	i.On("GetKey", mock.Anything, mock.Anything).Return(struct {
		Key       [32]byte
		Purposes  []*big.Int
		RevokedAt *big.Int
	}{
		[32]byte{1},
		[]*big.Int{big.NewInt(identity.KeyPurposeP2P)},
		big.NewInt(1),
	}, errors.New("p2p key error"))
	srv := NewEthereumIdentityService(c, f, r, nil, func() ethereum.Client {
		return g
	}, func(address common.Address, backend bind.ContractBackend) (contract, error) {
		return i, nil
	})
	p2p, err := srv.GetIdentityKey(centID, pubKey)
	f.AssertExpectations(t)
	r.AssertExpectations(t)
	g.AssertExpectations(t)
	assert.Empty(t, p2p, "p2p is not empty")
	assert.Errorf(t, err, "error should not be nil")
	assert.Contains(t, err.Error(), "p2p key error")
}

func TestGetIdentityKey_fail_empty(t *testing.T) {
	centID, _ := identity.ToCentID(utils.RandomSlice(identity.CentIDLength))
	pubKey := utils.RandomSlice(32)
	c, f, r, g, i := &testingconfig.MockConfig{}, &MockIDFactory{}, &MockIDRegistry{}, &testingcommons.MockEthClient{}, &MockIDContract{}
	g.On("GetGethCallOpts").Return(&bind.CallOpts{}, func() {})
	g.On("GetEthClient").Return(&ethclient.Client{})
	r.On("GetIdentityByCentrifugeId", mock.Anything, centID.BigInt()).Return(common.BytesToAddress(utils.RandomSlice(20)), nil)
	i.On("GetKey", mock.Anything, mock.Anything).Return(struct {
		Key       [32]byte
		Purposes  []*big.Int
		RevokedAt *big.Int
	}{
		[32]byte{},
		[]*big.Int{big.NewInt(identity.KeyPurposeP2P)},
		big.NewInt(1),
	}, nil)
	srv := NewEthereumIdentityService(c, f, r, nil, func() ethereum.Client {
		return g
	}, func(address common.Address, backend bind.ContractBackend) (contract, error) {
		return i, nil
	})
	p2p, err := srv.GetIdentityKey(centID, pubKey)
	f.AssertExpectations(t)
	r.AssertExpectations(t)
	g.AssertExpectations(t)
	assert.Empty(t, p2p, "p2p is not empty")
	assert.Errorf(t, err, "error should not be nil")
	assert.Contains(t, err.Error(), "key not found for identity")
}

func TestGetIdentityKey_Success(t *testing.T) {
	centID, _ := identity.ToCentID(utils.RandomSlice(identity.CentIDLength))
	pubKey := utils.RandomSlice(32)
	c, f, r, g, i := &testingconfig.MockConfig{}, &MockIDFactory{}, &MockIDRegistry{}, &testingcommons.MockEthClient{}, &MockIDContract{}
	g.On("GetGethCallOpts").Return(&bind.CallOpts{}, func() {})
	g.On("GetEthClient").Return(&ethclient.Client{})
	r.On("GetIdentityByCentrifugeId", mock.Anything, centID.BigInt()).Return(common.BytesToAddress(utils.RandomSlice(20)), nil)
	i.On("GetKey", mock.Anything, mock.Anything).Return(struct {
		Key       [32]byte
		Purposes  []*big.Int
		RevokedAt *big.Int
	}{
		[32]byte{1},
		[]*big.Int{big.NewInt(identity.KeyPurposeP2P)},
		big.NewInt(1),
	}, nil)
	srv := NewEthereumIdentityService(c, f, r, nil, func() ethereum.Client {
		return g
	}, func(address common.Address, backend bind.ContractBackend) (contract, error) {
		return i, nil
	})
	p2p, err := srv.GetIdentityKey(centID, pubKey)
	f.AssertExpectations(t)
	r.AssertExpectations(t)
	g.AssertExpectations(t)
	assert.NotEmpty(t, p2p, "p2p is empty")
	assert.Nil(t, err, "error must be nil")
}

func TestValidateKey_success(t *testing.T) {
	centID, _ := identity.ToCentID(utils.RandomSlice(identity.CentIDLength))
	pubKey := utils.RandomSlice(32)
	var key [32]byte
	copy(key[:], pubKey)
	c, f, r, g, i := &testingconfig.MockConfig{}, &MockIDFactory{}, &MockIDRegistry{}, &testingcommons.MockEthClient{}, &MockIDContract{}
	g.On("GetGethCallOpts").Return(&bind.CallOpts{}, func() {})
	g.On("GetEthClient").Return(&ethclient.Client{})
	r.On("GetIdentityByCentrifugeId", mock.Anything, centID.BigInt()).Return(common.BytesToAddress(utils.RandomSlice(20)), nil)
	i.On("GetKey", mock.Anything, mock.Anything).Return(struct {
		Key       [32]byte
		Purposes  []*big.Int
		RevokedAt *big.Int
	}{
		key,
		[]*big.Int{big.NewInt(identity.KeyPurposeSigning)},
		big.NewInt(0),
	}, nil)
	srv := NewEthereumIdentityService(c, f, r, nil, func() ethereum.Client {
		return g
	}, func(address common.Address, backend bind.ContractBackend) (contract, error) {
		return i, nil
	})
	err := srv.ValidateKey(centID, pubKey, identity.KeyPurposeSigning)
	f.AssertExpectations(t)
	r.AssertExpectations(t)
	g.AssertExpectations(t)
	assert.Nil(t, err, "error must be nil")
}

func TestValidateKey_fail_getId(t *testing.T) {
	centID, _ := identity.ToCentID(utils.RandomSlice(identity.CentIDLength))
	pubKey := utils.RandomSlice(32)
	var key [32]byte
	copy(key[:], pubKey)
	c, f, r, g, i := &testingconfig.MockConfig{}, &MockIDFactory{}, &MockIDRegistry{}, &testingcommons.MockEthClient{}, &MockIDContract{}
	g.On("GetGethCallOpts").Return(&bind.CallOpts{}, func() {})
	g.On("GetEthClient").Return(&ethclient.Client{})
	r.On("GetIdentityByCentrifugeId", mock.Anything, centID.BigInt()).Return(common.BytesToAddress(utils.RandomSlice(20)), nil)
	i.On("GetKey", mock.Anything, mock.Anything).Return(struct {
		Key       [32]byte
		Purposes  []*big.Int
		RevokedAt *big.Int
	}{
		key,
		[]*big.Int{big.NewInt(identity.KeyPurposeSigning)},
		big.NewInt(0),
	}, errors.New("Key error"))
	srv := NewEthereumIdentityService(c, f, r, nil, func() ethereum.Client {
		return g
	}, func(address common.Address, backend bind.ContractBackend) (contract, error) {
		return i, nil
	})
	err := srv.ValidateKey(centID, pubKey, identity.KeyPurposeSigning)
	f.AssertExpectations(t)
	r.AssertExpectations(t)
	g.AssertExpectations(t)
	assert.Contains(t, err.Error(), "Key error")
}

func TestValidateKey_fail_mismatch_key(t *testing.T) {
	centID, _ := identity.ToCentID(utils.RandomSlice(identity.CentIDLength))
	pubKey := utils.RandomSlice(32)
	c, f, r, g, i := &testingconfig.MockConfig{}, &MockIDFactory{}, &MockIDRegistry{}, &testingcommons.MockEthClient{}, &MockIDContract{}
	g.On("GetGethCallOpts").Return(&bind.CallOpts{}, func() {})
	g.On("GetEthClient").Return(&ethclient.Client{})
	r.On("GetIdentityByCentrifugeId", mock.Anything, centID.BigInt()).Return(common.BytesToAddress(utils.RandomSlice(20)), nil)
	i.On("GetKey", mock.Anything, mock.Anything).Return(struct {
		Key       [32]byte
		Purposes  []*big.Int
		RevokedAt *big.Int
	}{
		[32]byte{1},
		[]*big.Int{big.NewInt(identity.KeyPurposeSigning)},
		big.NewInt(0),
	}, nil)
	srv := NewEthereumIdentityService(c, f, r, nil, func() ethereum.Client {
		return g
	}, func(address common.Address, backend bind.ContractBackend) (contract, error) {
		return i, nil
	})
	err := srv.ValidateKey(centID, pubKey, identity.KeyPurposeSigning)
	f.AssertExpectations(t)
	r.AssertExpectations(t)
	g.AssertExpectations(t)
	assert.Contains(t, err.Error(), "Key doesn't match")
}

func TestValidateKey_fail_missing_purpose(t *testing.T) {
	centID, _ := identity.ToCentID(utils.RandomSlice(identity.CentIDLength))
	pubKey := utils.RandomSlice(32)
	var key [32]byte
	copy(key[:], pubKey)
	c, f, r, g, i := &testingconfig.MockConfig{}, &MockIDFactory{}, &MockIDRegistry{}, &testingcommons.MockEthClient{}, &MockIDContract{}
	g.On("GetGethCallOpts").Return(&bind.CallOpts{}, func() {})
	g.On("GetEthClient").Return(&ethclient.Client{})
	r.On("GetIdentityByCentrifugeId", mock.Anything, centID.BigInt()).Return(common.BytesToAddress(utils.RandomSlice(20)), nil)
	i.On("GetKey", mock.Anything, mock.Anything).Return(struct {
		Key       [32]byte
		Purposes  []*big.Int
		RevokedAt *big.Int
	}{
		key,
		nil,
		big.NewInt(0),
	}, nil)
	srv := NewEthereumIdentityService(c, f, r, nil, func() ethereum.Client {
		return g
	}, func(address common.Address, backend bind.ContractBackend) (contract, error) {
		return i, nil
	})
	err := srv.ValidateKey(centID, pubKey, identity.KeyPurposeSigning)
	f.AssertExpectations(t)
	r.AssertExpectations(t)
	g.AssertExpectations(t)
	assert.Contains(t, err.Error(), "Key doesn't have purpose")
}

func TestValidateKey_fail_wrong_purpose(t *testing.T) {
	centID, _ := identity.ToCentID(utils.RandomSlice(identity.CentIDLength))
	pubKey := utils.RandomSlice(32)
	var key [32]byte
	copy(key[:], pubKey)
	c, f, r, g, i := &testingconfig.MockConfig{}, &MockIDFactory{}, &MockIDRegistry{}, &testingcommons.MockEthClient{}, &MockIDContract{}
	g.On("GetGethCallOpts").Return(&bind.CallOpts{}, func() {})
	g.On("GetEthClient").Return(&ethclient.Client{})
	r.On("GetIdentityByCentrifugeId", mock.Anything, centID.BigInt()).Return(common.BytesToAddress(utils.RandomSlice(20)), nil)
	i.On("GetKey", mock.Anything, mock.Anything).Return(struct {
		Key       [32]byte
		Purposes  []*big.Int
		RevokedAt *big.Int
	}{
		key,
		[]*big.Int{big.NewInt(identity.KeyPurposeP2P)},
		big.NewInt(0),
	}, nil)
	srv := NewEthereumIdentityService(c, f, r, nil, func() ethereum.Client {
		return g
	}, func(address common.Address, backend bind.ContractBackend) (contract, error) {
		return i, nil
	})
	err := srv.ValidateKey(centID, pubKey, identity.KeyPurposeSigning)
	f.AssertExpectations(t)
	r.AssertExpectations(t)
	g.AssertExpectations(t)
	assert.Contains(t, err.Error(), "Key doesn't have purpose")
}

func TestValidateKey_fail_revocation(t *testing.T) {
	centID, _ := identity.ToCentID(utils.RandomSlice(identity.CentIDLength))
	pubKey := utils.RandomSlice(32)
	var key [32]byte
	copy(key[:], pubKey)
	c, f, r, g, i := &testingconfig.MockConfig{}, &MockIDFactory{}, &MockIDRegistry{}, &testingcommons.MockEthClient{}, &MockIDContract{}
	g.On("GetGethCallOpts").Return(&bind.CallOpts{}, func() {})
	g.On("GetEthClient").Return(&ethclient.Client{})
	r.On("GetIdentityByCentrifugeId", mock.Anything, centID.BigInt()).Return(common.BytesToAddress(utils.RandomSlice(20)), nil)
	i.On("GetKey", mock.Anything, mock.Anything).Return(struct {
		Key       [32]byte
		Purposes  []*big.Int
		RevokedAt *big.Int
	}{
		key,
		[]*big.Int{big.NewInt(identity.KeyPurposeSigning)},
		big.NewInt(1),
	}, nil)
	srv := NewEthereumIdentityService(c, f, r, nil, func() ethereum.Client {
		return g
	}, func(address common.Address, backend bind.ContractBackend) (contract, error) {
		return i, nil
	})
	err := srv.ValidateKey(centID, pubKey, identity.KeyPurposeSigning)
	f.AssertExpectations(t)
	r.AssertExpectations(t)
	g.AssertExpectations(t)
	assert.Contains(t, err.Error(), "Key is currently revoked since block")
}
