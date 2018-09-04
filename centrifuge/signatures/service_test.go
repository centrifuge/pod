// +build unit

package signatures

import (
	"fmt"
	"os"
	"testing"
	testing2 "github.com/CentrifugeInc/go-centrifuge/centrifuge/context/testing"
	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/coredocument"
	"golang.org/x/crypto/ed25519"
	"math/big"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/identity"
	"github.com/stretchr/testify/assert"
	"encoding/base64"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/tools"
	"github.com/go-errors/errors"
)

var (
	testSigningService                 SigningService
	testKeys                           []*identity.EthereumIdentityKey
	key1Pub, key2Pub, key3Pub, key4Pub ed25519.PublicKey
	key1, key2, key3, key4             ed25519.PrivateKey
	id1                                = []byte{1, 1, 1, 1, 1, 1}
	id2                                = []byte{2, 2, 2, 2, 2, 2}
	id3                                = []byte{3, 3, 3, 3, 3, 3}
	testCase													 = 0
)

// Mock IdentityService
type EthereumIdentityMocked struct {
	CentrifugeId []byte
}
func (id *EthereumIdentityMocked) String() string {
	return fmt.Sprintf("CentrifugeId [%s]", id.CentrifugeIdString())
}
func (id *EthereumIdentityMocked) GetCentrifugeId() []byte {
	return id.CentrifugeId
}
func (id *EthereumIdentityMocked) CentrifugeIdString() string {
	return base64.StdEncoding.EncodeToString(id.CentrifugeId)
}
func (id *EthereumIdentityMocked) CentrifugeIdBytes() [identity.CentIdByteLength]byte {
	var idBytes [identity.CentIdByteLength]byte
	copy(idBytes[:], id.CentrifugeId[:identity.CentIdByteLength])
	return idBytes
}
func (id *EthereumIdentityMocked) CentrifugeIdBigInt() *big.Int {
	bi := tools.ByteSliceToBigInt(id.CentrifugeId)
	return bi
}
func (id *EthereumIdentityMocked) SetCentrifugeId(b []byte) error {
	if len(b) != identity.CentIdByteLength {
		return errors.New("CentrifugeId has incorrect length")
	}
	if tools.IsEmptyByteSlice(b) {
		return errors.New("CentrifugeId can't be empty")
	}
	id.CentrifugeId = b
	return nil
}
func (id *EthereumIdentityMocked) GetCurrentP2PKey() (ret string, err error) {
	return "", nil
}
func (id *EthereumIdentityMocked) GetLastKeyForPurpose(keyPurpose int) (key []byte, err error) {
	return []byte{}, nil
}
func (id *EthereumIdentityMocked) AddKeyToIdentity(keyPurpose int, key []byte) (confirmations chan *identity.WatchIdentity, err error) {
	return nil, nil
}
func (id *EthereumIdentityMocked)CheckIdentityExists() (exists bool, err error) {
	return true, nil
}
func (id *EthereumIdentityMocked) FetchKey() (identity.IdentityKey, error) {
	return testKeys[testCase], nil
}

type EthereumIdentityMockedService struct {}
func (ids *EthereumIdentityMockedService) CheckIdentityExists(centrifugeId []byte) (exists bool, err error) {
	return true, nil
}
func (ids *EthereumIdentityMockedService) LookupIdentityForId(centrifugeId []byte) (id identity.Identity, err error) {
	return &EthereumIdentityMocked{
		CentrifugeId: centrifugeId,
	}, nil
}
func (ids *EthereumIdentityMockedService) CreateIdentity(centrifugeId []byte) (id identity.Identity, confirmations chan *identity.WatchIdentity, err error) {
	return
}
//

func TestMain(m *testing.M) {
	testing2.InitTestConfig()
	testSigningService = SigningService{IdentityService: &EthereumIdentityMockedService{}}

	// Generated with: key1Pub, key1, _ := ed25519.GenerateKey(rand.Reader)
	key1Pub = []byte{230, 49, 10, 12, 200, 149, 43, 184, 145, 87, 163, 252, 114, 31, 91, 163, 24, 237, 36, 51, 165, 8, 34, 104, 97, 49, 114, 85, 255, 15, 195, 199}
	key1 = []byte{102, 109, 71, 239, 130, 229, 128, 189, 37, 96, 223, 5, 189, 91, 210, 47, 89, 4, 165, 6, 188, 53, 49, 250, 109, 151, 234, 139, 57, 205, 231, 253, 230, 49, 10, 12, 200, 149, 43, 184, 145, 87, 163, 252, 114, 31, 91, 163, 24, 237, 36, 51, 165, 8, 34, 104, 97, 49, 114, 85, 255, 15, 195, 199}
	key2Pub = []byte{113, 94, 241, 90, 37, 21, 191, 41, 217, 134, 68, 34, 107, 207, 87, 203, 0, 204, 113, 213, 126, 0, 98, 57, 220, 45, 155, 242, 151, 76, 39, 231}
	key2 = []byte{250, 254, 250, 250, 96, 33, 238, 120, 185, 198, 83, 69, 160, 207, 138, 239, 51, 172, 90, 252, 23, 83, 19, 166, 136, 131, 142, 18, 119, 36, 176, 0, 113, 94, 241, 90, 37, 21, 191, 41, 217, 134, 68, 34, 107, 207, 87, 203, 0, 204, 113, 213, 126, 0, 98, 57, 220, 45, 155, 242, 151, 76, 39, 231}
	key3Pub = []byte{153, 38, 210, 185, 3, 120, 208, 57, 159, 178, 175, 184, 253, 124, 188, 72, 228, 245, 66, 98, 7, 1, 124, 244, 49, 92, 75, 59, 147, 186, 37, 145}
	key3 = []byte{169, 105, 156, 14, 78, 30, 37, 102, 84, 55, 7, 105, 163, 167, 111, 93, 125, 15, 81, 210, 84, 27, 148, 50, 240, 36, 50, 3, 105, 106, 251, 135, 153, 38, 210, 185, 3, 120, 208, 57, 159, 178, 175, 184, 253, 124, 188, 72, 228, 245, 66, 98, 7, 1, 124, 244, 49, 92, 75, 59, 147, 186, 37, 145}
	key4 = []byte{122, 14, 67, 51, 130, 66, 68, 193, 100, 229, 106, 56, 134, 179, 142, 245, 213, 199, 100, 108, 18, 156, 27, 10, 143, 156, 250, 85, 253, 123, 12, 255}
	key4Pub = []byte{142, 110, 173, 1, 3, 27, 149, 34, 232, 120, 203, 249, 23, 246, 46, 66, 9, 136, 96, 204, 31, 123, 166, 133, 36, 102, 58, 27, 26, 252, 239, 116, 122, 14, 67, 51, 130, 66, 68, 193, 100, 229, 106, 56, 134, 179, 142, 245, 213, 199, 100, 108, 18, 156, 27, 10, 143, 156, 250, 85, 253, 123, 12, 255}

	b32k1Pub ,_ := tools.SliceToByte32(key1Pub)
	b32k2Pub ,_ := tools.SliceToByte32(key2Pub)
	b32k3Pub ,_ := tools.SliceToByte32(key3Pub)

	// Valid key
	testKeys = []*identity.EthereumIdentityKey{
		{
			b32k1Pub,
			[]*big.Int{big.NewInt(2)},
			big.NewInt(0),
		},
		// Revoked key
		{
			b32k2Pub,
			[]*big.Int{big.NewInt(2)},
			big.NewInt(1000),
		},
		// Key with non-signing purpose
		{
			b32k3Pub,
			[]*big.Int{big.NewInt(1)},
			big.NewInt(0),
		},
	}

	os.Exit(m.Run())
}

func TestSignatureValidation(t *testing.T) {
	valid, err := testSigningService.ValidateKey(id1, key1Pub)
	assert.Nil(t, err)
	assert.True(t, valid)
}

func TestDocumentSignatures(t *testing.T) {
	// Any 32byte value for these identifiers is ok (such as a ed25519 public key)
	doc := &coredocumentpb.CoreDocument{
		SigningRoot: key1Pub,
	}

	sig := testSigningService.MakeSignature(doc, id1, key1, key1Pub)
	assert.Equal(t, sig.Signature, []byte{0x4e, 0x3d, 0x90, 0x5f, 0x25, 0xc7, 0x90, 0x63, 0x7e, 0x6c, 0xd0, 0xe6, 0xc7, 0xbd, 0xe6,
	0x81, 0x3b, 0xd0, 0x5b, 0x94, 0x76, 0x86, 0x4e, 0xcb, 0xb9, 0x36, 0x48, 0x44, 0x4b, 0x98, 0xd2, 0x4b, 0x6a, 0x65, 0x22, 0x92,
	0x1c, 0x8a, 0xdb, 0xfe, 0xb7, 0x6f, 0xfe, 0x34, 0x52, 0xa3, 0x49, 0xe4, 0xda, 0xdc, 0x5d, 0x1b, 0x0, 0x79, 0x54, 0x60, 0x29,
	0x22, 0x94, 0xb, 0x3c, 0x90, 0x3c, 0x3})

	valid, err := testSigningService.ValidateSignature(sig, doc.SigningRoot)
	assert.Nil(t, err)
	assert.True(t, valid)

	testCase++

	sig = testSigningService.MakeSignature(doc, id1, key2, key2Pub)
	assert.Equal(t, sig.Signature, []byte{0x20, 0x66, 0xb0, 0x94, 0xf8, 0xce, 0x6, 0x85, 0xe0, 0xa9, 0x21, 0xe1, 0xd6, 0xde,
	0xd2, 0xd9, 0x3d, 0x42, 0x3e, 0x70, 0x92, 0x7, 0xa6, 0xa9, 0xdf, 0x68, 0xa, 0x9f, 0x7e, 0x6d, 0xf0, 0x3d, 0x1d, 0x1a, 0x3,
	0xc, 0xe4, 0x32, 0x8c, 0x9b, 0xbe, 0x31, 0xba, 0x67, 0xca, 0xa1, 0xba, 0xc9, 0x9b, 0x29, 0x68, 0xdb, 0x5e, 0x56, 0xa6, 0xd3,
	0x6e, 0x35, 0xfa, 0xe1, 0x19, 0xc2, 0x45, 0x4})

	valid, err = testSigningService.ValidateSignature(sig, doc.SigningRoot)
	assert.NotNil(t, err)
	assert.EqualError(t, err, "[Key: 715ef15a2515bf29d98644226bcf57cb00cc71d57e006239dc2d9bf2974c27e7] Key is currently revoked since block [1000]")

	testCase++

	sig = testSigningService.MakeSignature(doc, id1, key3, key3Pub)
	assert.Equal(t, sig.Signature, []byte{0x1d, 0x1a, 0xb9, 0xf2, 0xf2, 0x25, 0x89, 0x66, 0xb, 0xf8, 0x79, 0xc8, 0x31, 0x93,
	0xc3, 0x57, 0x8, 0xaa, 0x2a, 0x53, 0x42, 0x2, 0xe5, 0x51, 0xcc, 0xb6, 0x5e, 0xf, 0x12, 0xb7, 0xe6, 0x18, 0x7f, 0x3e, 0xfb,
	0x23, 0xdb, 0x13, 0xc8, 0xd7, 0x73, 0x8f, 0x65, 0x71, 0x87, 0xe7, 0x4e, 0x10, 0x50, 0x80, 0xf1, 0x17, 0x55, 0xb7, 0xd4, 0xeb,
	0x0, 0x3d, 0x90, 0xd7, 0x9c, 0x1b, 0xfc, 0x5})

	valid, err = testSigningService.ValidateSignature(sig, doc.SigningRoot)
	assert.NotNil(t, err)
	assert.EqualError(t, err, "[Key: 9926d2b90378d0399fb2afb8fd7cbc48e4f5426207017cf4315c4b3b93ba2591] Key doesn't have purpose [2]")

}

//func TestDocumentSigning(t *testing.T) {
//	dataRoot := testingutils.Rand32Bytes()
//	documentIdentifier := testingutils.Rand32Bytes()
//	nextIdentifier := testingutils.Rand32Bytes()
//
//	doc := &coredocumentpb.CoreDocument{
//		DataRoot:           dataRoot,
//		DocumentIdentifier: documentIdentifier,
//		NextIdentifier:     nextIdentifier,
//	}
//
//	testSigningService.Sign(doc)
//	valid, err := testSigningService.ValidateSignaturesOnDocument(doc)
//	if !valid || err != nil {
//		// t.Fatal("Signature validation failed")
//		fmt.Println("TEST IGNORED") // TODO
//	}
//}
