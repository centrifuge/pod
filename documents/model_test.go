// +build unit

package documents

import (
	"crypto/sha256"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/p2p"
	"github.com/centrifuge/go-centrifuge/protobufs/gen/go/document"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/mock"

	"github.com/centrifuge/centrifuge-protobufs/documenttypes"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/precise-proofs/proofs"
	"github.com/golang/protobuf/ptypes/any"

	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common/hexutil"

	"github.com/centrifuge/go-centrifuge/anchors"

	"github.com/centrifuge/go-centrifuge/ethereum"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/storage/leveldb"

	"github.com/centrifuge/go-centrifuge/config/configstore"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/bootstrap/bootstrappers/testlogging"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/queue"
	"github.com/centrifuge/go-centrifuge/testingutils/commons"
	"github.com/centrifuge/go-centrifuge/transactions"
	"github.com/stretchr/testify/assert"
)

var ctx map[string]interface{}
var ConfigService config.Service
var cfg config.Configuration

func TestMain(m *testing.M) {
	ctx = make(map[string]interface{})
	ethClient := &testingcommons.MockEthClient{}
	ethClient.On("GetEthClient").Return(nil)
	ctx[ethereum.BootstrappedEthereumClient] = ethClient
	ibootstappers := []bootstrap.TestBootstrapper{
		&testlogging.TestLoggingBootstrapper{},
		&config.Bootstrapper{},
		&leveldb.Bootstrapper{},
		&configstore.Bootstrapper{},
		transactions.Bootstrapper{},
		&queue.Bootstrapper{},
		&anchors.Bootstrapper{},
		&Bootstrapper{},
	}
	ctx[identity.BootstrappedIDService] = &testingcommons.MockIDService{}
	bootstrap.RunTestBootstrappers(ibootstappers, ctx)
	ConfigService = ctx[config.BootstrappedConfigStorage].(config.Service)
	cfg = ctx[bootstrap.BootstrappedConfig].(config.Configuration)
	cfg.Set("keys.p2p.publicKey", "../build/resources/p2pKey.pub.pem")
	cfg.Set("keys.p2p.privateKey", "../build/resources/p2pKey.key.pem")
	cfg.Set("keys.signing.publicKey", "../build/resources/signingKey.pub.pem")
	cfg.Set("keys.signing.privateKey", "../build/resources/signingKey.key.pem")
	cfg.Set("keys.ethauth.publicKey", "../build/resources/ethauth.pub.pem")
	cfg.Set("keys.ethauth.privateKey", "../build/resources/ethauth.key.pem")
	result := m.Run()
	bootstrap.RunTestTeardown(ibootstappers)
	os.Exit(result)
}

func Test_fetchUniqueCollaborators(t *testing.T) {

	tests := []struct {
		old    [][]byte
		new    []string
		result []identity.CentID
		err    bool
	}{
		{
			new:    []string{"0x010203040506"},
			result: []identity.CentID{{1, 2, 3, 4, 5, 6}},
		},

		{
			old:    [][]byte{{1, 2, 3, 2, 3, 1}},
			new:    []string{"0x010203040506"},
			result: []identity.CentID{{1, 2, 3, 4, 5, 6}},
		},

		{
			old: [][]byte{{1, 2, 3, 2, 3, 1}, {1, 2, 3, 4, 5, 6}},
			new: []string{"0x010203040506"},
		},

		{
			old: [][]byte{{1, 2, 3, 2, 3, 1}, {1, 2, 3, 4, 5, 6}},
		},

		// new collaborator with wrong format
		{
			old: [][]byte{{1, 2, 3, 2, 3, 1}, {1, 2, 3, 4, 5, 6}},
			new: []string{"0x0102030405"},
			err: true,
		},
	}

	for _, c := range tests {
		uc, err := fetchUniqueCollaborators(c.old, c.new)
		if err != nil {
			if c.err {
				continue
			}

			t.Fatal(err)
		}

		assert.Equal(t, c.result, uc)
	}
}

func TestCoreDocumentModel_PrepareNewVersion(t *testing.T) {
	dm := NewCoreDocModel()
	cd := dm.Document
	assert.NotNil(t, cd)

	//collaborators need to be hex string
	collabs := []string{"some ID"}
	newDocModel, err := dm.PrepareNewVersion(collabs)
	assert.Error(t, err)
	assert.Nil(t, newDocModel)

	// missing DocumentRoot
	c1 := utils.RandomSlice(6)
	c2 := utils.RandomSlice(6)
	c := []string{hexutil.Encode(c1), hexutil.Encode(c2)}
	ndm, err := dm.PrepareNewVersion(c)
	assert.NotNil(t, err)
	assert.Nil(t, ndm)

	// successful preparation of new version upon addition of DocumentRoot
	cd.DocumentRoot = utils.RandomSlice(32)
	ndm, err = dm.PrepareNewVersion(c)
	assert.Nil(t, err)
	assert.NotNil(t, ndm)

	// successful updating of version in new Document
	ncd := ndm.Document
	ocd := dm.Document
	assert.Equal(t, ncd.PreviousVersion, ocd.CurrentVersion)
	assert.Equal(t, ncd.CurrentVersion, ocd.NextVersion)

	// DocumentIdentifier has not changed
	assert.Equal(t, ncd.DocumentIdentifier, ocd.DocumentIdentifier)

	// DocumentRoot was updated
	assert.Equal(t, ncd.PreviousRoot, ocd.DocumentRoot)

	// TokenRegistry was copied over
	assert.Equal(t, ndm.TokenRegistry, dm.TokenRegistry)
}

func TestReadACLs_initReadRules(t *testing.T) {
	dm := NewCoreDocModel()
	cd := dm.Document
	err := dm.initReadRules(nil)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(ErrZeroCollaborators, err))

	cs := []identity.CentID{identity.RandomCentID()}
	err = dm.initReadRules(cs)
	assert.NoError(t, err)
	assert.Len(t, cd.ReadRules, 1)
	assert.Len(t, cd.Roles, 1)

	err = dm.initReadRules(cs)
	assert.NoError(t, err)
	assert.Len(t, cd.ReadRules, 1)
	assert.Len(t, cd.Roles, 1)
}

func TestReadAccessValidator_AccountCanRead(t *testing.T) {
	dm := NewCoreDocModel()
	account, err := identity.CentIDFromString("0x010203040506")
	assert.NoError(t, err)

	dm.Document.DocumentRoot = utils.RandomSlice(32)
	ndm, err := dm.PrepareNewVersion([]string{account.String()})
	cd := ndm.Document
	assert.NoError(t, err)
	assert.NotNil(t, cd.ReadRules)
	assert.NotNil(t, cd.Roles)

	// account who cant access
	rcid := identity.RandomCentID()
	assert.False(t, ndm.AccountCanRead(rcid))

	// account can access
	assert.True(t, ndm.AccountCanRead(account))
}

func TestGetSigningProofHashes(t *testing.T) {
	docAny := &any.Any{
		TypeUrl: documenttypes.InvoiceDataTypeUrl,
		Value:   []byte{},
	}
	dm := NewCoreDocModel()
	cd := dm.Document
	cd.EmbeddedData = docAny
	cd.DataRoot = utils.RandomSlice(32)
	err := dm.setCoreDocumentSalts()
	assert.NoError(t, err)

	err = dm.CalculateSigningRoot(cd.DataRoot)
	assert.Nil(t, err)

	err = dm.CalculateDocumentRoot()
	assert.Nil(t, err)

	hashes, err := dm.getSigningRootProofHashes()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(hashes))

	valid, err := proofs.ValidateProofSortedHashes(cd.SigningRoot, hashes, cd.DocumentRoot, sha256.New())
	assert.True(t, valid)
	assert.Nil(t, err)
}

func TestGetDataProofHashes(t *testing.T) {
	docAny := &any.Any{
		TypeUrl: documenttypes.InvoiceDataTypeUrl,
		Value:   []byte{},
	}
	dm := NewCoreDocModel()
	cd := dm.Document
	cd.EmbeddedData = docAny
	cd.DataRoot = utils.RandomSlice(32)
	err := dm.setCoreDocumentSalts()
	assert.NoError(t, err)

	err = dm.CalculateSigningRoot(cd.DataRoot)
	assert.Nil(t, err)

	err = dm.CalculateDocumentRoot()
	assert.Nil(t, err)

	hashes, err := dm.getDataProofHashes(cd.DataRoot)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(hashes))

	valid, err := proofs.ValidateProofSortedHashes(cd.DataRoot, hashes, cd.DocumentRoot, sha256.New())
	assert.True(t, valid)
	assert.Nil(t, err)
}

func TestGetDocumentSigningTree(t *testing.T) {
	docAny := &any.Any{
		TypeUrl: documenttypes.InvoiceDataTypeUrl,
		Value:   []byte{},
	}
	dm := NewCoreDocModel()
	cd := dm.Document
	cd.EmbeddedData = docAny
	err := dm.setCoreDocumentSalts()
	assert.NoError(t, err)
	tree, err := dm.GetDocumentSigningTree(cd.DataRoot)
	assert.Nil(t, err)
	assert.NotNil(t, tree)

	_, leaf := tree.GetLeafByProperty("data_root")
	assert.NotNil(t, leaf)

	_, leaf = tree.GetLeafByProperty("cd_root")
	assert.NotNil(t, leaf)
}

func TestGetDocumentSigningTree_EmptyEmbeddedData(t *testing.T) {
	dm := NewCoreDocModel()
	cd := dm.Document
	err := dm.setCoreDocumentSalts()
	assert.NoError(t, err)
	tree, err := dm.GetDocumentSigningTree(cd.DataRoot)
	assert.NotNil(t, err)
	assert.Nil(t, tree)
}

// TestGetDocumentRootTree tests that the documentroottree is properly calculated
func TestGetDocumentRootTree(t *testing.T) {
	dm := NewCoreDocModel()
	cd := &coredocumentpb.CoreDocument{SigningRoot: []byte{0x72, 0xee, 0xb8, 0x88, 0x92, 0xf7, 0x6, 0x19, 0x82, 0x76, 0xe9, 0xe7, 0xfe, 0xcc, 0x33, 0xa, 0x66, 0x78, 0xd4, 0xa6, 0x5f, 0xf6, 0xa, 0xca, 0x2b, 0xe4, 0x17, 0xa9, 0xf6, 0x15, 0x67, 0xa1}}
	dm.Document = cd
	tree, err := dm.GetDocumentRootTree()

	// Manually constructing the two node tree:
	signaturesLengthLeaf := sha256.Sum256(append(append(compactProperties[SignaturesField], []byte{48}...), make([]byte, 32)...))
	expectedRootHash := sha256.Sum256(append(dm.Document.SigningRoot, signaturesLengthLeaf[:]...))
	assert.Nil(t, err)
	assert.Equal(t, expectedRootHash[:], tree.RootHash())
}

func TestCreateProofs(t *testing.T) {
	h := sha256.New()
	testTree := NewDefaultTree(nil)
	props := []proofs.Property{NewLeafProperty("sample_field", []byte{0, 0, 0, 200}), NewLeafProperty("sample_field2", []byte{0, 0, 0, 202})}
	compactProps := [][]byte{props[0].Compact, props[1].Compact}
	err := testTree.AddLeaf(proofs.LeafNode{Hash: utils.RandomSlice(32), Hashed: true, Property: props[0]})
	assert.NoError(t, err)
	err = testTree.AddLeaf(proofs.LeafNode{Hash: utils.RandomSlice(32), Hashed: true, Property: props[1]})
	assert.NoError(t, err)
	err = testTree.Generate()
	assert.NoError(t, err)
	docAny := &any.Any{
		TypeUrl: documenttypes.InvoiceDataTypeUrl,
		Value:   []byte{},
	}
	dm := NewCoreDocModel()
	cd := dm.Document
	cd.EmbeddedData = docAny
	cd.Collaborators = [][]byte{utils.RandomSlice(32), utils.RandomSlice(32)}
	err = dm.setCoreDocumentSalts()
	assert.NoError(t, err)
	err = dm.CalculateSigningRoot(testTree.RootHash())
	assert.NoError(t, err)
	err = dm.CalculateDocumentRoot()
	assert.NoError(t, err)
	cdTree, err := dm.GetDocumentTree()
	assert.NoError(t, err)
	tests := []struct {
		fieldName   string
		fromCoreDoc bool
		proofLength int
	}{
		{
			"sample_field",
			false,
			3,
		},
		{
			"document_identifier",
			true,
			6,
		},
		{
			"sample_field2",
			false,
			3,
		},
		{
			"collaborators[0]",
			true,
			6,
		},
	}
	for _, test := range tests {
		t.Run(test.fieldName, func(t *testing.T) {
			p, err := dm.CreateProofs(testTree, []string{test.fieldName})
			assert.NoError(t, err)
			assert.Equal(t, test.proofLength, len(p[0].SortedHashes))
			var l *proofs.LeafNode
			if test.fromCoreDoc {
				_, l = cdTree.GetLeafByProperty(test.fieldName)
			} else {
				_, l = testTree.GetLeafByProperty(test.fieldName)
				assert.Contains(t, compactProps, l.Property.CompactName())
			}
			valid, err := proofs.ValidateProofSortedHashes(l.Hash, p[0].SortedHashes, cd.DocumentRoot, h)
			assert.NoError(t, err)
			assert.True(t, valid)
		})
	}
}

type mockRegistry struct {
	mock.Mock
}

func (m mockRegistry) OwnerOf(registry common.Address, tokenID []byte) (common.Address, error) {
	args := m.Called(registry, tokenID)
	addr, _ := args.Get(0).(common.Address)
	return addr, args.Error(1)
}

func Test_addNFTToReadRules(t *testing.T) {
	dm := NewCoreDocModel()
	// wrong registry or token format
	registry := common.HexToAddress("0xf72855759a39fb75fc7341139f5d7a3974d4da08")
	tokenID := utils.RandomSlice(34)
	_, err := dm.AddNFTToReadRules(registry, tokenID)
	assert.Error(t, err)

	dm.Document.DocumentRoot = utils.RandomSlice(32)
	dm, err = dm.PrepareNewVersion([]string{"0x010203040506"})
	cd := dm.Document
	assert.NoError(t, err)
	assert.Len(t, cd.ReadRules, 1)
	assert.Equal(t, cd.ReadRules[0].Action, coredocumentpb.Action_ACTION_READ_SIGN)
	assert.Len(t, cd.Roles, 1)

	tokenID = utils.RandomSlice(32)
	_, err = dm.AddNFTToReadRules(registry, tokenID)
	assert.NoError(t, err)
	assert.Len(t, cd.ReadRules, 2)
	assert.Equal(t, cd.ReadRules[1].Action, coredocumentpb.Action_ACTION_READ)
	assert.Len(t, cd.Roles, 2)
}

func TestReadAccessValidator_NFTOwnerCanRead(t *testing.T) {
	dm := NewCoreDocModel()
	dm.Document.DocumentRoot = utils.RandomSlice(32)
	account, err := identity.CentIDFromString("0x010203040506")
	assert.NoError(t, err)

	dm, err = dm.PrepareNewVersion([]string{account.String()})
	assert.NoError(t, err)

	registry := common.HexToAddress("0xf72855759a39fb75fc7341139f5d7a3974d4da08")

	// account can read
	err = dm.NFTOwnerCanRead(registry, nil, account)
	assert.NoError(t, err)

	// account not in read rules and nft missing
	account, err = identity.CentIDFromString("0x010203040505")
	assert.NoError(t, err)
	tokenID := utils.RandomSlice(32)
	err = dm.NFTOwnerCanRead(registry, tokenID, account)
	assert.Error(t, err)

	tr := mockRegistry{}
	tr.On("OwnerOf", registry, tokenID).Return(nil, errors.New("failed to get owner of")).Once()
	dm.TokenRegistry = tr
	dm.AddNFTToReadRules(registry, tokenID)
	err = dm.NFTOwnerCanRead(registry, tokenID, account)
	assert.Error(t, err)
	assert.Contains(t, err, "failed to get owner of")
	tr.AssertExpectations(t)

	// not the same owner
	owner := common.BytesToAddress(utils.RandomSlice(20))
	tr.On("OwnerOf", registry, tokenID).Return(owner, nil).Once()
	dm.TokenRegistry = tr
	err = dm.NFTOwnerCanRead(registry, tokenID, account)
	assert.Error(t, err)
	tr.AssertExpectations(t)
}

func TestCoreDocumentModel_AddAccessTokenToReadRules(t *testing.T) {
	m := NewCoreDocModel()
	m.Document.DocumentRoot = utils.RandomSlice(32)
	idConfig, err := identity.GetIdentityConfig(cfg)
	assert.NoError(t, err)

	cd := m.Document
	assert.Len(t, cd.ReadRules, 0)
	assert.Len(t, cd.Roles, 0)

	// invalid access token payload
	payload := documentpb.AccessTokenParams{
		// invalid grantee format
		Grantee: "randomCentID",
		DocumentIdentifier: "randomDocID",
	}
	_, err = m.AddAccessTokenToReadRules(*idConfig, payload)
    assert.Error(t, err, "	failed to construct AT: hex string without 0x prefix")

	// valid access token payload
	id := idConfig.ID.String()
		payload = documentpb.AccessTokenParams{
		Grantee: id,
		DocumentIdentifier: string(m.Document.DocumentIdentifier),
	}
	dm, err := m.AddAccessTokenToReadRules(*idConfig, payload)
	assert.NoError(t, err)
	assert.Len(t, dm.Document.ReadRules, 1)
	assert.Equal(t, dm.Document.ReadRules[0].Action, coredocumentpb.Action_ACTION_READ)
	assert.Len(t, dm.Document.Roles, 1)
}

func TestCoreDocumentModel_ATOwnerCanRead(t *testing.T) {
m := NewCoreDocModel()
	m.Document.DocumentRoot = utils.RandomSlice(32)
	idConfig, err := identity.GetIdentityConfig(cfg)
	assert.NoError(t, err)

	assert.NoError(t, err)
	granteeID := idConfig.ID.String()
	payload := documentpb.AccessTokenParams{
		Grantee: granteeID,
		DocumentIdentifier: string(m.Document.DocumentIdentifier),
	}
	dm, err := m.AddAccessTokenToReadRules(*idConfig, payload)
	assert.NoError(t, err)
	dm.Document.DocumentRoot = utils.RandomSlice(32)
	docRoles := dm.Document.GetRoles()
	at := docRoles[0].AccessTokens[0]
	// wrong token identifier
	tr := &p2ppb.AccessTokenRequest{
		DelegatingDocumentIdentifier: dm.Document.DocumentIdentifier,
		AccessTokenId: []byte("randomtokenID"),
	}
	dr := &p2ppb.GetDocumentRequest{
		DocumentIdentifier: m.Document.DocumentIdentifier,
		AccessType: p2ppb.AccessType_ACCESS_TYPE_ACCESS_TOKEN_VERIFICATION,
		AccessTokenRequest: tr,
	}
	err = dm.ATOwnerCanRead(dr, idConfig.ID)
	assert.Error(t, err, "access token not found")
	// valid access token
	tr = &p2ppb.AccessTokenRequest{
		DelegatingDocumentIdentifier: dm.Document.DocumentIdentifier,
		AccessTokenId: at.Identifier,
	}
	dr.AccessTokenRequest = tr
	err = dm.ATOwnerCanRead(dr, idConfig.ID)
	assert.NoError(t, err)
}

func TestCoreDocumentModel_ValidateDocumentAccess(t *testing.T) {

}

func TestGetCoreDocumentSalts(t *testing.T) {
	dm := NewCoreDocModel()
	// From empty
	err := dm.setCoreDocumentSalts()
	assert.NoError(t, err)
	assert.NotNil(t, dm.Document.CoredocumentSalts)
	salts := dm.Document.CoredocumentSalts

	// Return existing
	err = dm.setCoreDocumentSalts()
	assert.NoError(t, err)
	assert.NotNil(t, dm.Document.CoredocumentSalts)
	assert.Equal(t, salts, dm.Document.CoredocumentSalts)
}

func TestGenerateNewSalts(t *testing.T) {
	dm := NewCoreDocModel()
	salts, err := GenerateNewSalts(dm.Document, "", nil)
	assert.NoError(t, err)
	assert.NotNil(t, salts)
}

func TestConvertToProofAndProtoSalts(t *testing.T) {
	dm := NewCoreDocModel()
	salts, err := GenerateNewSalts(dm.Document, "", nil)
	assert.NoError(t, err)
	assert.NotNil(t, salts)

	nilProto := ConvertToProtoSalts(nil)
	assert.Nil(t, nilProto)

	nilProof := ConvertToProofSalts(nil)
	assert.Nil(t, nilProof)

	protoSalts := ConvertToProtoSalts(salts)
	assert.NotNil(t, protoSalts)
	assert.Len(t, protoSalts, len(*salts))
	assert.Equal(t, protoSalts[0].Value, (*salts)[0].Value)

	cSalts := ConvertToProofSalts(protoSalts)
	assert.NotNil(t, cSalts)
	assert.Len(t, *cSalts, len(*salts))
	assert.Equal(t, (*cSalts)[0].Value, (*salts)[0].Value)
}
