// +build unit

package identity

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/signatures"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var ctx = map[string]interface{}{}
var cfg *config.Configuration

func TestMain(m *testing.M) {
	ibootstappers := []bootstrap.TestBootstrapper{
		&config.Bootstrapper{},
	}
	bootstrap.RunTestBootstrappers(ibootstappers, ctx)
	cfg = ctx[config.BootstrappedConfig].(*config.Configuration)
	cfg.Set("keys.signing.publicKey", "../build/resources/signingKey.pub.pem")
	cfg.Set("keys.signing.privateKey", "../build/resources/signingKey.key.pem")
	cfg.Set("keys.ethauth.publicKey", "../build/resources/ethauth.pub.pem")
	cfg.Set("keys.ethauth.privateKey", "../build/resources/ethauth.key.pem")
	result := m.Run()
	bootstrap.RunTestTeardown(ibootstappers)
	os.Exit(result)
}

// mockID implements Identity
type mockID struct {
	mock.Mock
}

func (i *mockID) String() string {
	args := i.Called()
	return args.String(0)
}

func (i *mockID) CentID() CentID {
	args := i.Called()
	return args.Get(0).(CentID)
}

func (i *mockID) SetCentrifugeID(centId CentID) {
	i.Called(centId)
}

func (i *mockID) CurrentP2PKey() (ret string, err error) {
	args := i.Called()
	return args.String(0), args.Error(1)
}

func (i *mockID) LastKeyForPurpose(keyPurpose int) (key []byte, err error) {
	args := i.Called(keyPurpose)
	return args.Get(0).([]byte), args.Error(1)
}

func (i *mockID) AddKeyToIdentity(ctx context.Context, keyPurpose int, key []byte) (confirmations chan *WatchIdentity, err error) {
	args := i.Called(ctx, keyPurpose, key)
	return args.Get(0).(chan *WatchIdentity), args.Error(1)
}

func (i *mockID) FetchKey(key []byte) (Key, error) {
	args := i.Called(key)
	idKey := args.Get(0)
	if idKey != nil {
		if k, ok := idKey.(Key); ok {
			return k, args.Error(1)
		}
	}
	return nil, args.Error(1)
}

// mockIDService implements Service
type mockIDService struct {
	mock.Mock
}

func (srv *mockIDService) GetIdentityAddress(centID CentID) (common.Address, error) {
	args := srv.Called(centID)
	id := args.Get(0).(common.Address)
	return id, args.Error(1)
}

func (srv *mockIDService) LookupIdentityForID(centID CentID) (Identity, error) {
	args := srv.Called(centID)
	id := args.Get(0)
	if id != nil {
		return id.(Identity), args.Error(1)
	}
	return nil, args.Error(1)

}

func (srv *mockIDService) CreateIdentity(centID CentID) (Identity, chan *WatchIdentity, error) {
	args := srv.Called(centID)
	id := args.Get(0).(Identity)
	return id, args.Get(1).(chan *WatchIdentity), args.Error(2)
}

func (srv *mockIDService) CheckIdentityExists(centID CentID) (exists bool, err error) {
	args := srv.Called(centID)
	return args.Bool(0), args.Error(1)
}

func TestGetIdentityConfig_Success(t *testing.T) {
	idConfig, err := GetIdentityConfig(cfg)
	assert.Nil(t, err)
	assert.NotNil(t, idConfig)
	configId, err := cfg.GetIdentityID()
	assert.Nil(t, err)
	idBytes := idConfig.ID[:]
	assert.Equal(t, idBytes, configId)
	assert.Equal(t, 3, len(idConfig.Keys))
}

func TestGetIdentityConfig_Error(t *testing.T) {
	//Wrong Hex
	currentId := cfg.GetString("identityId")
	cfg.Set("identityId", "ABCD")
	idConfig, err := GetIdentityConfig(cfg)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "hex string without 0x prefix")
	assert.Nil(t, idConfig)
	cfg.Set("identityId", currentId)

	//Wrong length
	currentId = cfg.GetString("identityId")
	cfg.Set("identityId", "0x0101010101")
	idConfig, err = GetIdentityConfig(cfg)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "invalid length byte slice provided for centID")
	assert.Nil(t, idConfig)
	cfg.Set("identityId", currentId)

	//Wrong public signing key path
	currentKeyPath, _ := cfg.GetSigningKeyPair()
	cfg.Set("keys.signing.publicKey", "./build/resources/signingKey.pub.pem")
	idConfig, err = GetIdentityConfig(cfg)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "no such file or directory")
	assert.Nil(t, idConfig)
	cfg.Set("keys.signing.publicKey", currentKeyPath)

	//Wrong public ethauth key path
	currentKeyPath, _ = cfg.GetEthAuthKeyPair()
	cfg.Set("keys.ethauth.publicKey", "./build/resources/ethauth.pub.pem")
	idConfig, err = GetIdentityConfig(cfg)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "no such file or directory")
	assert.Nil(t, idConfig)
	cfg.Set("keys.ethauth.publicKey", currentKeyPath)
}

func TestToCentId(t *testing.T) {
	tests := []struct {
		name  string
		slice []byte
		err   string
	}{
		{
			"smallerSlice",
			utils.RandomSlice(CentIDLength - 1),
			"invalid length byte slice provided for centID",
		},
		{
			"largerSlice",
			utils.RandomSlice(CentIDLength + 1),
			"invalid length byte slice provided for centID",
		},
		{
			"nilSlice",
			nil,
			"empty bytes provided",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := ToCentID(test.slice)
			assert.Equal(t, test.err, err.Error())
		})
	}
}

func TestNewCentIdEqual(t *testing.T) {
	randomBytes := utils.RandomSlice(CentIDLength)
	centrifugeIdA, err := ToCentID(randomBytes)
	assert.Nil(t, err, "centrifugeId not initialized correctly ")

	centrifugeIdB, err := ToCentID(randomBytes)
	assert.Nil(t, err, "centrifugeId not initialized correctly ")
	assert.True(t, centrifugeIdA.Equal(centrifugeIdB), "centrifuge Id's should be the equal")

	randomBytes = utils.RandomSlice(CentIDLength)
	centrifugeIdC, _ := ToCentID(randomBytes)
	assert.False(t, centrifugeIdA.Equal(centrifugeIdC), "centrifuge Id's should not be equal")
}

func TestCentIDFromString(t *testing.T) {
	tests := []struct {
		id     string
		result CentID
		err    error
	}{
		{
			id:     "0x010203040506",
			result: [CentIDLength]byte{1, 2, 3, 4, 5, 6},
		},

		{
			id:  "0x01020304050607",
			err: fmt.Errorf("invalid length byte slice provided for centID"),
		},

		{
			id:  "0xsome random",
			err: fmt.Errorf("failed to decode id"),
		},

		{
			id:  "some random",
			err: fmt.Errorf("hex string without 0x"),
		},
	}

	for _, c := range tests {
		id, err := CentIDFromString(c.id)
		if c.err == nil {
			assert.Nil(t, err, "must be nil")
			assert.Equal(t, c.result, id, "id must match")
			continue
		}

		assert.Error(t, err, "must be a non nil error")
		assert.Contains(t, err.Error(), c.err.Error())
	}
}

func TestCentIDsFromStrings(t *testing.T) {
	// fail due to error
	ids := []string{"0x010203040506", "some id"}
	cids, err := CentIDsFromStrings(ids)
	assert.Error(t, err)
	assert.Nil(t, cids)

	ids = []string{"0x010203040506", "0x020301020304"}
	cids, err = CentIDsFromStrings(ids)
	assert.Nil(t, err)
	assert.NotNil(t, cids)
	assert.Equal(t, cids, []CentID{{1, 2, 3, 4, 5, 6}, {2, 3, 1, 2, 3, 4}})
}

func TestValidateCentrifugeIDBytes(t *testing.T) {
	c := RandomCentID()
	assert.True(t, ValidateCentrifugeIDBytes(c[:], c) == nil)

	err := ValidateCentrifugeIDBytes(utils.RandomSlice(20), c)
	if assert.Error(t, err) {
		assert.Equal(t, "invalid length byte slice provided for centID", err.Error())
	}

	err = ValidateCentrifugeIDBytes(utils.RandomSlice(6), c)
	if assert.Error(t, err) {
		assert.Equal(t, "provided bytes doesn't match centID", err.Error())
	}
}

func TestSign(t *testing.T) {
	key1Pub := []byte{230, 49, 10, 12, 200, 149, 43, 184, 145, 87, 163, 252, 114, 31, 91, 163, 24, 237, 36, 51, 165, 8, 34, 104, 97, 49, 114, 85, 255, 15, 195, 199}
	key1 := []byte{102, 109, 71, 239, 130, 229, 128, 189, 37, 96, 223, 5, 189, 91, 210, 47, 89, 4, 165, 6, 188, 53, 49, 250, 109, 151, 234, 139, 57, 205, 231, 253, 230, 49, 10, 12, 200, 149, 43, 184, 145, 87, 163, 252, 114, 31, 91, 163, 24, 237, 36, 51, 165, 8, 34, 104, 97, 49, 114, 85, 255, 15, 195, 199}
	c := RandomCentID()
	msg := utils.RandomSlice(100)
	sig := Sign(&IDConfig{c, map[int]IDKey{KeyPurposeSigning: {PrivateKey: key1, PublicKey: key1Pub}}}, KeyPurposeSigning, msg)

	err := signatures.VerifySignature(key1Pub, msg, sig.Signature)
	assert.True(t, err == nil)
}
