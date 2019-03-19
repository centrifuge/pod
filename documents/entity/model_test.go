// +build unit

package entity

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/centrifuge/centrifuge-protobufs/documenttypes"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/bootstrap/bootstrappers/testlogging"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/config/configstore"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/ethereum"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/identity/ideth"
	"github.com/centrifuge/go-centrifuge/nft"
	"github.com/centrifuge/go-centrifuge/p2p"
	cliententitypb "github.com/centrifuge/go-centrifuge/protobufs/gen/go/entity"
	"github.com/centrifuge/go-centrifuge/queue"
	"github.com/centrifuge/go-centrifuge/storage/leveldb"
	"github.com/centrifuge/go-centrifuge/testingutils/commons"
	"github.com/centrifuge/go-centrifuge/testingutils/config"
	"github.com/centrifuge/go-centrifuge/testingutils/documents"
	"github.com/centrifuge/go-centrifuge/testingutils/identity"
	"github.com/centrifuge/go-centrifuge/testingutils/testingtx"
	"github.com/centrifuge/go-centrifuge/transactions"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var ctx = map[string]interface{}{}
var cfg config.Configuration
var configService config.Service
var defaultDID = testingidentity.GenerateRandomDID()

func TestMain(m *testing.M) {
	ethClient := &testingcommons.MockEthClient{}
	ethClient.On("GetEthClient").Return(nil)
	ctx[ethereum.BootstrappedEthereumClient] = ethClient
	txMan := &testingtx.MockTxManager{}
	ctx[transactions.BootstrappedService] = txMan
	done := make(chan bool)
	txMan.On("ExecuteWithinTX", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(transactions.NilTxID(), done, nil)
	ctx[nft.BootstrappedPayObService] = new(testingdocuments.MockRegistry)
	ibootstrappers := []bootstrap.TestBootstrapper{
		&testlogging.TestLoggingBootstrapper{},
		&config.Bootstrapper{},
		&leveldb.Bootstrapper{},
		&queue.Bootstrapper{},
		&ideth.Bootstrapper{},
		&configstore.Bootstrapper{},
		anchors.Bootstrapper{},
		documents.Bootstrapper{},
		p2p.Bootstrapper{},
		documents.PostBootstrapper{},
		&Bootstrapper{},
		&queue.Starter{},
	}
	bootstrap.RunTestBootstrappers(ibootstrappers, ctx)
	cfg = ctx[bootstrap.BootstrappedConfig].(config.Configuration)
	cfg.Set("identityId", cid.String())
	configService = ctx[config.BootstrappedConfigStorage].(config.Service)
	result := m.Run()
	bootstrap.RunTestTeardown(ibootstrappers)
	os.Exit(result)
}

func TestEntity_PackCoreDocument(t *testing.T) {
	ctx := testingconfig.CreateAccountContext(t, cfg)
	did, err := contextutil.AccountDID(ctx)
	assert.NoError(t, err)

	inv := new(Entity)
	assert.NoError(t, inv.InitEntityInput(testingdocuments.CreateEntityPayload(), did.String()))

	cd, err := inv.PackCoreDocument()
	assert.NoError(t, err)
	assert.NotNil(t, cd.EmbeddedData)
	assert.NotNil(t, cd.EmbeddedDataSalts)
}

func TestEntity_JSON(t *testing.T) {
	inv := new(Entity)
	ctx := testingconfig.CreateAccountContext(t, cfg)
	did, err := contextutil.AccountDID(ctx)
	assert.NoError(t, err)
	assert.NoError(t, inv.InitEntityInput(testingdocuments.CreateEntityPayload(), did.String()))

	cd, err := inv.PackCoreDocument()
	assert.NoError(t, err)
	jsonBytes, err := inv.JSON()
	assert.Nil(t, err, "marshal to json didn't work correctly")
	assert.True(t, json.Valid(jsonBytes), "json format not correct")

	inv = new(Entity)
	err = inv.FromJSON(jsonBytes)
	assert.Nil(t, err, "unmarshal JSON didn't work correctly")

	ncd, err := inv.PackCoreDocument()
	assert.Nil(t, err, "JSON unmarshal damaged entity variables")
	assert.Equal(t, cd, ncd)
}

func TestEntityModel_UnpackCoreDocument(t *testing.T) {
	var model = new(Entity)
	var err error

	// embed data missing
	err = model.UnpackCoreDocument(coredocumentpb.CoreDocument{})
	assert.Error(t, err)

	// embed data type is wrong
	err = model.UnpackCoreDocument(coredocumentpb.CoreDocument{EmbeddedData: new(any.Any)})
	assert.Error(t, err, "unpack must fail due to missing embed data")

	// embed data is wrong
	err = model.UnpackCoreDocument(coredocumentpb.CoreDocument{
		EmbeddedData: &any.Any{
			Value:   utils.RandomSlice(32),
			TypeUrl: documenttypes.EntityDataTypeUrl,
		},
	})
	assert.Error(t, err)

	// successful
	inv, cd := createCDWithEmbeddedEntity(t)
	err = model.UnpackCoreDocument(cd)
	assert.NoError(t, err)
	assert.Equal(t, model.getClientData(), inv.(*Entity).getClientData())
	assert.Equal(t, model.ID(), inv.ID())
	assert.Equal(t, model.CurrentVersion(), inv.CurrentVersion())
	assert.Equal(t, model.PreviousVersion(), inv.PreviousVersion())
}

func TestEntityModel_getClientData(t *testing.T) {
	invData := testingdocuments.CreateEntityData()
	inv := new(Entity)
	inv.loadFromP2PProtobuf(&invData)

	data := inv.getClientData()
	assert.NotNil(t, data, "entity data should not be nil")
	assert.Equal(t, data.GrossAmount, data.GrossAmount, "gross amount must match")
	assert.Equal(t, data.Recipient, hexutil.Encode(inv.Recipient[:]), "recipient should match")
	assert.Equal(t, data.Sender, hexutil.Encode(inv.Sender[:]), "sender should match")
	assert.Equal(t, data.Payee, hexutil.Encode(inv.Payee[:]), "payee should match")
}

func TestEntityModel_InitEntityInput(t *testing.T) {
	ctx := testingconfig.CreateAccountContext(t, cfg)
	did, err := contextutil.AccountDID(ctx)
	assert.NoError(t, err)

	// fail recipient
	data := &cliententitypb.EntityData{
		Sender:    "some number",
		Payee:     "some payee",
		Recipient: "some recipient",
		ExtraData: "some data",
	}
	inv := new(Entity)
	err = inv.InitEntityInput(&cliententitypb.EntityCreatePayload{Data: data}, did.String())
	assert.Error(t, err, "must return err")
	assert.Contains(t, err.Error(), "failed to decode extra data")
	assert.Nil(t, inv.Recipient)
	assert.Nil(t, inv.Sender)
	assert.Nil(t, inv.Payee)
	assert.Nil(t, inv.ExtraData)

	data.ExtraData = "0x010203020301"
	recipientDID := testingidentity.GenerateRandomDID()
	data.Recipient = recipientDID.String()
	err = inv.InitEntityInput(&cliententitypb.EntityCreatePayload{Data: data}, did.String())
	assert.Nil(t, err)
	assert.NotNil(t, inv.ExtraData)
	assert.NotNil(t, inv.Recipient)
	assert.Nil(t, inv.Sender)
	assert.Nil(t, inv.Payee)

	senderDID := testingidentity.GenerateRandomDID()
	data.Sender = senderDID.String()
	err = inv.InitEntityInput(&cliententitypb.EntityCreatePayload{Data: data}, did.String())
	assert.Nil(t, err)
	assert.NotNil(t, inv.ExtraData)
	assert.NotNil(t, inv.Recipient)
	assert.NotNil(t, inv.Sender)
	assert.Nil(t, inv.Payee)

	payeeDID := testingidentity.GenerateRandomDID()
	data.Payee = payeeDID.String()
	err = inv.InitEntityInput(&cliententitypb.EntityCreatePayload{Data: data}, did.String())
	assert.Nil(t, err)
	assert.NotNil(t, inv.ExtraData)
	assert.NotNil(t, inv.Recipient)
	assert.NotNil(t, inv.Sender)
	assert.NotNil(t, inv.Payee)

	data.ExtraData = "0x010203020301"
	collabs := []string{"0x010102040506", "some id"}
	err = inv.InitEntityInput(&cliententitypb.EntityCreatePayload{Data: data, Collaborators: collabs}, did.String())
	assert.Contains(t, err.Error(), "failed to decode collaborator")

	collab1, err := identity.NewDIDFromString("0xBAEb33a61f05e6F269f1c4b4CFF91A901B54DaF7")
	assert.NoError(t, err)
	collab2, err := identity.NewDIDFromString("0xBAEb33a61f05e6F269f1c4b4CFF91A901B54DaF3")
	assert.NoError(t, err)
	collabs = []string{collab1.String(), collab2.String()}
	err = inv.InitEntityInput(&cliententitypb.EntityCreatePayload{Data: data, Collaborators: collabs}, did.String())
	assert.Nil(t, err, "must be nil")
	assert.Equal(t, inv.Sender[:], senderDID[:])
	assert.Equal(t, inv.Payee[:], payeeDID[:])
	assert.Equal(t, inv.Recipient[:], recipientDID[:])
	assert.Equal(t, inv.ExtraData[:], []byte{1, 2, 3, 2, 3, 1})
}

func TestEntityModel_calculateDataRoot(t *testing.T) {
	ctx := testingconfig.CreateAccountContext(t, cfg)
	did, err := contextutil.AccountDID(ctx)
	assert.NoError(t, err)
	m := new(Entity)
	err = m.InitEntityInput(testingdocuments.CreateEntityPayload(), did.String())
	assert.Nil(t, err, "Init must pass")
	assert.Nil(t, m.EntitySalts, "salts must be nil")

	dr, err := m.CalculateDataRoot()
	assert.Nil(t, err, "calculate must pass")
	assert.False(t, utils.IsEmptyByteSlice(dr))
	assert.NotNil(t, m.EntitySalts, "salts must be created")
}

func TestEntity_CreateProofs(t *testing.T) {
	i := createEntity(t)
	rk := i.Document.Roles[0].RoleKey
	pf := fmt.Sprintf(documents.CDTreePrefix+".roles[%s].collaborators[0]", hexutil.Encode(rk))
	proof, err := i.CreateProofs([]string{"entity.entity_number", pf, documents.CDTreePrefix + ".document_type"})
	assert.Nil(t, err)
	assert.NotNil(t, proof)
	tree, err := i.CoreDocument.DocumentRootTree()
	assert.NoError(t, err)

	// Validate entity_number
	valid, err := tree.ValidateProof(proof[0])
	assert.Nil(t, err)
	assert.True(t, valid)

	// Validate roles
	valid, err = tree.ValidateProof(proof[1])
	assert.Nil(t, err)
	assert.True(t, valid)

	// Validate []byte value
	acc := identity.NewDIDFromBytes(proof[1].Value)
	assert.True(t, i.AccountCanRead(acc))

	// Validate document_type
	valid, err = tree.ValidateProof(proof[2])
	assert.Nil(t, err)
	assert.True(t, valid)
}

func TestEntityModel_createProofsFieldDoesNotExist(t *testing.T) {
	i := createEntity(t)
	_, err := i.CreateProofs([]string{"nonexisting"})
	assert.NotNil(t, err)
}

func TestEntityModel_GetDocumentID(t *testing.T) {
	i := createEntity(t)
	assert.Equal(t, i.CoreDocument.ID(), i.ID())
}

func TestEntityModel_getDocumentDataTree(t *testing.T) {
	i := Entity{EntityNumber: "3213121", NetAmount: 2, GrossAmount: 2}
	tree, err := i.getDocumentDataTree()
	assert.Nil(t, err, "tree should be generated without error")
	_, leaf := tree.GetLeafByProperty("entity.entity_number")
	assert.NotNil(t, leaf)
	assert.Equal(t, "entity.entity_number", leaf.Property.ReadableName())
}

func createEntity(t *testing.T) *Entity {
	i := new(Entity)
	err := i.InitEntityInput(testingdocuments.CreateEntityPayload(), defaultDID.String())
	assert.NoError(t, err)
	_, err = i.CalculateDataRoot()
	assert.NoError(t, err)
	_, err = i.CalculateSigningRoot()
	assert.NoError(t, err)
	_, err = i.CalculateDocumentRoot()
	assert.NoError(t, err)
	return i
}

func TestEntity_CollaboratorCanUpdate(t *testing.T) {
	inv := createEntity(t)
	id1 := defaultDID
	id2 := testingidentity.GenerateRandomDID()
	id3 := testingidentity.GenerateRandomDID()

	// wrong type
	err := inv.CollaboratorCanUpdate(new(mockModel), id1)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrDocumentInvalidType, err))
	assert.NoError(t, testRepo().Create(id1[:], inv.CurrentVersion(), inv))

	// update the document
	model, err := testRepo().Get(id1[:], inv.CurrentVersion())
	assert.NoError(t, err)
	oldInv := model.(*Entity)
	data := oldInv.getClientData()
	data.GrossAmount = 50
	err = inv.PrepareNewVersion(inv, data, []string{id3.String()})
	assert.NoError(t, err)

	// id1 should have permission
	assert.NoError(t, oldInv.CollaboratorCanUpdate(inv, id1))

	// id2 should fail since it doesn't have the permission to update
	assert.Error(t, oldInv.CollaboratorCanUpdate(inv, id2))

	// update the id3 rules to update only gross amount
	inv.CoreDocument.Document.TransitionRules[3].MatchType = coredocumentpb.FieldMatchType_FIELD_MATCH_TYPE_EXACT
	inv.CoreDocument.Document.TransitionRules[3].Field = append(compactPrefix(), 0, 0, 0, 14)
	inv.CoreDocument.Document.DocumentRoot = utils.RandomSlice(32)
	assert.NoError(t, testRepo().Create(id1[:], inv.CurrentVersion(), inv))

	// fetch the document
	model, err = testRepo().Get(id1[:], inv.CurrentVersion())
	assert.NoError(t, err)
	oldInv = model.(*Entity)
	data = oldInv.getClientData()
	data.GrossAmount = 55
	data.Currency = "INR"
	err = inv.PrepareNewVersion(inv, data, nil)
	assert.NoError(t, err)

	// id1 should have permission
	assert.NoError(t, oldInv.CollaboratorCanUpdate(inv, id1))

	// id2 should fail since it doesn't have the permission to update
	assert.Error(t, oldInv.CollaboratorCanUpdate(inv, id2))

	// id3 should fail with just one error since changing Currency is not allowed
	err = oldInv.CollaboratorCanUpdate(inv, id3)
	assert.Error(t, err)
	assert.Equal(t, 1, errors.Len(err))
	assert.Contains(t, err.Error(), "entity.currency")
}