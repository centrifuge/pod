package identity

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"github.com/centrifuge/go-centrifuge/utils"
		"math/big"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/mock"
	"github.com/ethereum/go-ethereum/common"
	"github.com/centrifuge/go-centrifuge/ethereum"
	"github.com/ethereum/go-ethereum/ethclient"
	"net/url"
	"context"
	"github.com/go-errors/errors"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/testingutils/commons"
	"fmt"
)

type MockConfig struct{}

func (MockConfig) GetEthereumDefaultAccountName() string {
	return "mockacc"
}

type MockIDFactory struct{
	mock.Mock
}

func (f MockIDFactory) CreateIdentity(opts *bind.TransactOpts, _centrifugeId *big.Int) (*types.Transaction, error) {
	args := f.Called(opts, _centrifugeId)
	id := args.Get(0).(*types.Transaction)
	return id, args.Error(1)
}

type MockIDRegistry struct{
	mock.Mock
}

func (r MockIDRegistry) GetIdentityByCentrifugeId(opts *bind.CallOpts, bigInt *big.Int) (common.Address, error) {
	args := r.Called(opts, bigInt)
	id := args.Get(0).(common.Address)
	return id, args.Error(1)
}

type MockGethClient struct{
	mock.Mock
}

func (gc MockGethClient) GetEthClient() *ethclient.Client {
	args := gc.Called()
	return args.Get(0).(*ethclient.Client)
}

func (gc MockGethClient) GetNodeURL() *url.URL {
	args := gc.Called()
	return args.Get(0).(*url.URL)
}

func (gc MockGethClient) GetTxOpts(accountName string) (*bind.TransactOpts, error) {
	args := gc.Called(accountName)
	return args.Get(0).(*bind.TransactOpts), args.Error(1)
}

func (gc MockGethClient) SubmitTransactionWithRetries(contractMethod interface{}, opts *bind.TransactOpts, params ...interface{}) (tx *types.Transaction, err error) {
	args := gc.Called(contractMethod, opts, params)
	return args.Get(0).(*types.Transaction), args.Error(1)
}

func (gc MockGethClient) GetGethCallOpts() (*bind.CallOpts, context.CancelFunc) {
	args := gc.Called()
	return args.Get(0).(*bind.CallOpts), args.Get(1).(func())
}

type MockIDContract struct{
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
	centID, _ := ToCentID(utils.RandomSlice(CentIDLength))
	c, f, r, g, i := &MockConfig{}, &MockIDFactory{}, &MockIDRegistry{}, &MockGethClient{}, &MockIDContract{}
	g.On("GetGethCallOpts").Return(&bind.CallOpts{}, func() {})
	g.On("GetEthClient").Return(&ethclient.Client{})
	r.On("GetIdentityByCentrifugeId", mock.Anything, centID.BigInt()).Return(common.BytesToAddress(utils.RandomSlice(20)), nil)
	i.On("GetKeysByPurpose", mock.Anything, mock.Anything).Return([][32]byte{{1}}, nil)
	srv := NewEthereumIdentityService(c, f, r, func() ethereum.Client {
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
	centID, _ := ToCentID(utils.RandomSlice(CentIDLength))
	c, f, r, g, i := &MockConfig{}, &MockIDFactory{}, &MockIDRegistry{}, &MockGethClient{}, &MockIDContract{}
	g.On("GetGethCallOpts").Return(&bind.CallOpts{}, func() {})
	r.On("GetIdentityByCentrifugeId", mock.Anything, centID.BigInt()).Return(common.BytesToAddress(utils.RandomSlice(20)), errors.New("ID lookup failed"))
	i.On("GetKeysByPurpose", mock.Anything, mock.Anything).Return([][32]byte{{1}}, nil)
	srv := NewEthereumIdentityService(c, f, r, func() ethereum.Client {
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
	centID, _ := ToCentID(utils.RandomSlice(CentIDLength))
	c, f, r, g, i := &MockConfig{}, &MockIDFactory{}, &MockIDRegistry{}, &MockGethClient{}, &MockIDContract{}
	g.On("GetGethCallOpts").Return(&bind.CallOpts{}, func() {})
	g.On("GetEthClient").Return(&ethclient.Client{})
	r.On("GetIdentityByCentrifugeId", mock.Anything, centID.BigInt()).Return(common.BytesToAddress(utils.RandomSlice(20)), nil)
	i.On("GetKeysByPurpose", mock.Anything, mock.Anything).Return([][32]byte{{1}}, errors.New("p2p key error"))
	srv := NewEthereumIdentityService(c, f, r, func() ethereum.Client {
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

//
//func TestGetClientsP2PURLs_fail(t *testing.T) {
//	centID, _ := ToCentID(utils.RandomSlice(CentIDLength))
//	centIDs := []CentID{centID}
//	srv := &mockIDService{}
//	id := &mockID{}
//	id.On("CurrentP2PKey").Return("", fmt.Errorf("error identity")).Once()
//	srv.On("LookupIdentityForID", centIDs[0]).Return(id, nil).Once()
//	IDService = srv
//	p2pURLs, err := GetClientsP2PURLs(centIDs)
//	srv.AssertExpectations(t)
//	id.AssertExpectations(t)
//	assert.Empty(t, p2pURLs, "p2p is not empty")
//	assert.Errorf(t, err, "error should not be nil")
//	assert.Contains(t, err.Error(), "error identity")
//}
//
//func TestGetClientsP2PURLs_success(t *testing.T) {
//	centID, _ := ToCentID(utils.RandomSlice(CentIDLength))
//	centIDs := []CentID{centID}
//	id := &mockID{}
//	id.On("CurrentP2PKey").Return("target", nil).Once()
//	srv := &mockIDService{}
//	srv.On("LookupIdentityForID", centIDs[0]).Return(id, nil).Once()
//	IDService = srv
//	p2pURLs, err := GetClientsP2PURLs(centIDs)
//	srv.AssertExpectations(t)
//	id.AssertExpectations(t)
//	assert.Nil(t, err, "should be nil")
//	assert.NotEmpty(t, p2pURLs, "should not be empty")
//	assert.Equal(t, p2pURLs[0], "/ipfs/target")
//}
//
//func TestGetIdentityKey_fail_lookup(t *testing.T) {
//	centID, _ := ToCentID(utils.RandomSlice(CentIDLength))
//	srv := &mockIDService{}
//	srv.On("LookupIdentityForID", centID).Return(nil, fmt.Errorf("lookup failed")).Once()
//	IDService = srv
//	id, err := GetIdentityKey(centID, utils.RandomSlice(32))
//	srv.AssertExpectations(t)
//	assert.Error(t, err, "must be not nil")
//	assert.Contains(t, err.Error(), "lookup failed")
//	assert.Nil(t, id, "must be nil")
//}
//
//func TestGetIdentityKey_fail_FetchKey(t *testing.T) {
//	centID, _ := ToCentID(utils.RandomSlice(CentIDLength))
//	pubKey := utils.RandomSlice(32)
//	id := &mockID{}
//	srv := &mockIDService{}
//	srv.On("LookupIdentityForID", centID).Return(id, nil).Once()
//	id.On("FetchKey", pubKey).Return(nil, fmt.Errorf("fetch key failed")).Once()
//	IDService = srv
//	key, err := GetIdentityKey(centID, pubKey)
//	srv.AssertExpectations(t)
//	id.AssertExpectations(t)
//	assert.Error(t, err, "must be not nil")
//	assert.Contains(t, err.Error(), "fetch key failed")
//	assert.Nil(t, key, "must be nil")
//}
//
//func TestGetIdentityKey_fail_empty(t *testing.T) {
//	centID, _ := ToCentID(utils.RandomSlice(CentIDLength))
//	pubKey := utils.RandomSlice(32)
//	var emptyKey [32]byte
//	idkey := &EthereumIdentityKey{Key: emptyKey}
//	id := &mockID{}
//	srv := &mockIDService{}
//	srv.On("LookupIdentityForID", centID).Return(id, nil).Once()
//	id.On("FetchKey", pubKey).Return(idkey, nil).Once()
//	IDService = srv
//	key, err := GetIdentityKey(centID, pubKey)
//	srv.AssertExpectations(t)
//	id.AssertExpectations(t)
//	assert.Error(t, err, "must be not nil")
//	assert.Contains(t, err.Error(), "key not found for identity")
//	assert.Nil(t, key, "must be nil")
//}
//
//func TestGetIdentityKey_Success(t *testing.T) {
//	centID, _ := ToCentID(utils.RandomSlice(CentIDLength))
//	pubKey := utils.RandomSlice(32)
//	pkey := utils.RandomByte32()
//	idkey := &EthereumIdentityKey{Key: pkey}
//	id := &mockID{}
//	srv := &mockIDService{}
//	srv.On("LookupIdentityForID", centID).Return(id, nil).Once()
//	id.On("FetchKey", pubKey).Return(idkey, nil).Once()
//	IDService = srv
//	key, err := GetIdentityKey(centID, pubKey)
//	srv.AssertExpectations(t)
//	id.AssertExpectations(t)
//	assert.Nil(t, err, "error must be nil")
//	assert.NotNil(t, key, "must not be nil")
//	assert.Equal(t, key, idkey)
//}
//
//func TestValidateKey_fail_getId(t *testing.T) {
//	centID, _ := ToCentID(utils.RandomSlice(CentIDLength))
//	pubKey := utils.RandomSlice(32)
//	srv := &mockIDService{}
//	srv.On("LookupIdentityForID", centID).Return(nil, fmt.Errorf("failed at GetIdentity")).Once()
//	IDService = srv
//	err := ValidateKey(centID, pubKey, KeyPurposeSigning)
//	srv.AssertExpectations(t)
//	assert.Error(t, err, "must be not nil")
//	assert.Contains(t, err.Error(), "failed at GetIdentity")
//}
//
//func TestValidateKey_fail_mismatch_key(t *testing.T) {
//	centID, _ := ToCentID(utils.RandomSlice(CentIDLength))
//	pubKey := utils.RandomSlice(32)
//	idkey := &EthereumIdentityKey{Key: utils.RandomByte32()}
//	id := &mockID{}
//	srv := &mockIDService{}
//	srv.On("LookupIdentityForID", centID).Return(id, nil).Once()
//	id.On("FetchKey", pubKey).Return(idkey, nil).Once()
//	IDService = srv
//	err := ValidateKey(centID, pubKey, KeyPurposeSigning)
//	srv.AssertExpectations(t)
//	id.AssertExpectations(t)
//	assert.Error(t, err, "must be not nil")
//	assert.Contains(t, err.Error(), " Key doesn't match")
//}
//
//func TestValidateKey_fail_missing_purpose(t *testing.T) {
//	centID, _ := ToCentID(utils.RandomSlice(CentIDLength))
//	pubKey := utils.RandomByte32()
//	idkey := &EthereumIdentityKey{Key: pubKey}
//	id := &mockID{}
//	srv := &mockIDService{}
//	srv.On("LookupIdentityForID", centID).Return(id, nil).Once()
//	id.On("FetchKey", pubKey[:]).Return(idkey, nil).Once()
//	IDService = srv
//	err := ValidateKey(centID, pubKey[:], KeyPurposeSigning)
//	srv.AssertExpectations(t)
//	id.AssertExpectations(t)
//	assert.Error(t, err, "must be not nil")
//	assert.Contains(t, err.Error(), "Key doesn't have purpose")
//}
//
//func TestValidateKey_fail_wrong_purpose(t *testing.T) {
//	centID, _ := ToCentID(utils.RandomSlice(CentIDLength))
//	pubKey := utils.RandomByte32()
//	idkey := &EthereumIdentityKey{
//		Key:      pubKey,
//		Purposes: []*big.Int{big.NewInt(KeyPurposeEthMsgAuth)},
//	}
//	id := &mockID{}
//	srv := &mockIDService{}
//	srv.On("LookupIdentityForID", centID).Return(id, nil).Once()
//	id.On("FetchKey", pubKey[:]).Return(idkey, nil).Once()
//	IDService = srv
//	err := ValidateKey(centID, pubKey[:], KeyPurposeSigning)
//	srv.AssertExpectations(t)
//	id.AssertExpectations(t)
//	assert.Error(t, err, "must be not nil")
//	assert.Contains(t, err.Error(), "Key doesn't have purpose")
//}
//
//func TestValidateKey_fail_revocation(t *testing.T) {
//	centID, _ := ToCentID(utils.RandomSlice(CentIDLength))
//	pubKey := utils.RandomByte32()
//	idkey := &EthereumIdentityKey{
//		Key:       pubKey,
//		Purposes:  []*big.Int{big.NewInt(KeyPurposeSigning)},
//		RevokedAt: big.NewInt(1),
//	}
//	id := &mockID{}
//	srv := &mockIDService{}
//	srv.On("LookupIdentityForID", centID).Return(id, nil).Once()
//	id.On("FetchKey", pubKey[:]).Return(idkey, nil).Once()
//	IDService = srv
//	err := ValidateKey(centID, pubKey[:], KeyPurposeSigning)
//	srv.AssertExpectations(t)
//	id.AssertExpectations(t)
//	assert.Error(t, err, "must be not nil")
//	assert.Contains(t, err.Error(), "Key is currently revoked since block")
//}
//
//func TestValidateKey_success(t *testing.T) {
//	centID, _ := ToCentID(utils.RandomSlice(CentIDLength))
//	pubKey := utils.RandomByte32()
//	idkey := &EthereumIdentityKey{
//		Key:       pubKey,
//		Purposes:  []*big.Int{big.NewInt(KeyPurposeSigning)},
//		RevokedAt: big.NewInt(0),
//	}
//	id := &mockID{}
//	srv := &mockIDService{}
//	srv.On("LookupIdentityForID", centID).Return(id, nil).Once()
//	id.On("FetchKey", pubKey[:]).Return(idkey, nil).Once()
//	IDService = srv
//	err := ValidateKey(centID, pubKey[:], KeyPurposeSigning)
//	srv.AssertExpectations(t)
//	id.AssertExpectations(t)
//	assert.Nil(t, err, "must be nil")
//}


func TestValidateSignature_invalid_key(t *testing.T) {
	sig := &coredocumentpb.Signature{EntityId: utils.RandomSlice(CentIDLength)}
	srv := NewEthereumIdentityService(c, f, r, func() ethereum.Client {
		return g
	}, func(address common.Address, backend bind.ContractBackend) (contract, error) {
		return i, nil
	})
	centId, _ := ToCentID(sig.EntityId)
	srv.On("LookupIdentityForID", centId).Return(nil, fmt.Errorf("failed GetIdentity")).Once()
	identity.IDService = srv
	err := ValidateSignature(sig, key1Pub)
	srv.AssertExpectations(t)
	assert.NotNil(t, err, "must be not nil")
	assert.Contains(t, err.Error(), "failed GetIdentity")
}