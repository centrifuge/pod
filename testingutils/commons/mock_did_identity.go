package testingcommons

import (
	"context"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/mock"

	"math/big"
)

// MockIdentityService implements Service
type MockIdentityService struct {
	mock.Mock
}

// AddKey adds a key to identity contract
func (i *MockIdentityService) AddKey(ctx context.Context, key identity.KeyDID) error {
	args := i.Called(ctx, key)
	return args.Error(0)
}

// AddKeysForAccount adds key from configuration
func (i *MockIdentityService) AddKeysForAccount(acc config.Account) error {
	args := i.Called(acc)
	return args.Error(0)
}

// GetKey return a key from the identity contract
func (i *MockIdentityService) GetKey(did identity.DID, key [32]byte) (*identity.KeyResponse, error) {
	args := i.Called(did, key)
	return args.Get(0).(*identity.KeyResponse), args.Error(1)

}

// RawExecute calls the execute method on the identity contract
func (i *MockIdentityService) RawExecute(ctx context.Context, to common.Address, data []byte) error {
	args := i.Called(ctx, to, data)
	return args.Error(0)
}

// Execute creates the abi encoding an calls the execute method on the identity contract
func (i *MockIdentityService) Execute(ctx context.Context, to common.Address, contractAbi, methodName string, args ...interface{}) error {
	a := i.Called(ctx, to, contractAbi, methodName, args)
	return a.Error(0)
}

// IsSignedWithPurpose verifies if a message is signed with one of the identities specific purpose keys
func (i *MockIdentityService) IsSignedWithPurpose(did identity.DID, message [32]byte, signature []byte, purpose *big.Int) (bool, error) {
	args := i.Called(did, message, signature, purpose)
	return args.Get(0).(bool), args.Error(1)
}

// AddMultiPurposeKey adds a key with multiple purposes
func (i *MockIdentityService) AddMultiPurposeKey(ctx context.Context, key [32]byte, purposes []*big.Int, keyType *big.Int) error {
	args := i.Called(ctx, key, purposes, keyType)
	return args.Error(0)
}

// RevokeKey revokes an existing key in the smart contract
func (i *MockIdentityService) RevokeKey(ctx context.Context, key [32]byte) error {
	args := i.Called(ctx, key)
	return args.Error(0)

}

// GetClientP2PURL returns the p2p url associated with the did
func (i *MockIdentityService) GetClientP2PURL(did identity.DID) (string, error) {
	args := i.Called(did)
	return args.Get(0).(string), args.Error(1)

}

//Exists checks if an identity contract exists
func (i *MockIdentityService) Exists(ctx context.Context, did identity.DID) error {
	args := i.Called(ctx, did)
	return args.Error(0)
}

// ValidateKey checks if a given key is valid for the given centrifugeID.
func (i *MockIdentityService) ValidateKey(ctx context.Context, did identity.DID, key []byte, purpose int64) error {
	args := i.Called(ctx, did, key, purpose)
	return args.Error(0)
}

// ValidateSignature checks if signature is valid for given identity
func (i *MockIdentityService) ValidateSignature(signature *coredocumentpb.Signature, message []byte) error {
	args := i.Called(signature, message)
	return args.Error(0)
}

// CurrentP2PKey retrieves the last P2P key stored in the identity
func (i *MockIdentityService) CurrentP2PKey(did identity.DID) (ret string, err error) {
	args := i.Called(did)
	return args.Get(0).(string), args.Error(1)
}

// GetClientsP2PURLs returns p2p urls associated with each centIDs
// will error out at first failure
func (i *MockIdentityService) GetClientsP2PURLs(dids []*identity.DID) ([]string, error) {
	args := i.Called(dids)
	return args.Get(0).([]string), args.Error(1)

}

// GetKeysByPurpose returns keys grouped by purpose from the identity contract.
func (i *MockIdentityService) GetKeysByPurpose(did identity.DID, purpose *big.Int) ([][32]byte, error) {
	args := i.Called(did, purpose)
	return args.Get(0).([][32]byte), args.Error(1)
}

// MockIdentityFactory implements Service
type MockIdentityFactory struct {
	mock.Mock
}

func (s *MockIdentityFactory) CreateIdentity(ctx context.Context) (did *identity.DID, err error) {
	args := s.Called(ctx)
	return args.Get(0).(*identity.DID), args.Error(1)
}

func (s *MockIdentityFactory) CalculateIdentityAddress(ctx context.Context) (*common.Address, error) {
	args := s.Called(ctx)
	return args.Get(0).(*common.Address), args.Error(1)
}
