//go:build integration

package pending_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/bootstrap/bootstrappers/integration_test"
	"github.com/centrifuge/go-centrifuge/bootstrap/bootstrappers/testlogging"
	"github.com/centrifuge/go-centrifuge/centchain"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/config/configstore"
	"github.com/centrifuge/go-centrifuge/contextutil"
	protocolIDDispatcher "github.com/centrifuge/go-centrifuge/dispatcher"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/documents/entity"
	"github.com/centrifuge/go-centrifuge/documents/entityrelationship"
	"github.com/centrifuge/go-centrifuge/documents/generic"
	v2 "github.com/centrifuge/go-centrifuge/identity/v2"
	"github.com/centrifuge/go-centrifuge/ipfs"
	"github.com/centrifuge/go-centrifuge/jobs"
	nftv3 "github.com/centrifuge/go-centrifuge/nft/v3"
	"github.com/centrifuge/go-centrifuge/p2p"
	"github.com/centrifuge/go-centrifuge/pallets"
	"github.com/centrifuge/go-centrifuge/pending"
	"github.com/centrifuge/go-centrifuge/storage"
	"github.com/centrifuge/go-centrifuge/storage/leveldb"
	genericUtils "github.com/centrifuge/go-centrifuge/testingutils/generic"
	jobsUtil "github.com/centrifuge/go-centrifuge/testingutils/jobs"
	"github.com/centrifuge/go-centrifuge/testingutils/keyrings"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/stretchr/testify/assert"
)

var integrationTestBootstrappers = []bootstrap.TestBootstrapper{
	&integration_test.Bootstrapper{},
	&testlogging.TestLoggingBootstrapper{},
	&config.Bootstrapper{},
	&leveldb.Bootstrapper{},
	&configstore.Bootstrapper{},
	&jobs.Bootstrapper{},
	centchain.Bootstrapper{},
	&pallets.Bootstrapper{},
	&protocolIDDispatcher.Bootstrapper{},
	&v2.AccountTestBootstrapper{},
	documents.Bootstrapper{},
	pending.Bootstrapper{},
	&ipfs.TestBootstrapper{},
	&nftv3.Bootstrapper{},
	&p2p.Bootstrapper{},
	documents.PostBootstrapper{},
	generic.Bootstrapper{},
	entityrelationship.Bootstrapper{},
	entity.Bootstrapper{},
}

var (
	cfgService      config.Service
	documentService documents.Service
	documentRepo    documents.Repository
	pendingRepo     pending.Repository
	pendingService  pending.Service
	dispatcher      jobs.Dispatcher
)

func TestMain(m *testing.M) {
	ctx := bootstrap.RunTestBootstrappers(integrationTestBootstrappers, nil)

	cfgService = genericUtils.GetService[config.Service](ctx)
	documentService = genericUtils.GetService[documents.Service](ctx)
	documentRepo = genericUtils.GetService[documents.Repository](ctx)
	pendingService = genericUtils.GetService[pending.Service](ctx)
	dispatcher = genericUtils.GetService[jobs.Dispatcher](ctx)
	pendingRepo = pending.NewRepository(ctx[storage.BootstrappedDB].(storage.Repository))

	registerTestDocuments()

	result := m.Run()

	bootstrap.RunTestTeardown(integrationTestBootstrappers)

	os.Exit(result)
}

func registerTestDocuments() {
	documentSchemes := []string{
		generic.Scheme,
		entity.Scheme,
		entityrelationship.Scheme,
	}

	for _, documentScheme := range documentSchemes {
		testDoc, err := getTestDoc(documentScheme)

		if err != nil {
			panic(fmt.Errorf("couldn't get test document for scheme %s", documentScheme))
		}

		documentRepo.Register(testDoc)
	}
}

func Test_Integration_Service_GetPendingDocument(t *testing.T) {
	accs, err := cfgService.GetAccounts()
	assert.NoError(t, err)
	assert.NotEmpty(t, accs)

	acc := accs[0]

	ctx := contextutil.WithAccount(context.Background(), acc)

	testDoc, err := getTestDoc(generic.Scheme)
	assert.NoError(t, err)

	res, err := pendingService.Get(ctx, testDoc.ID(), documents.Pending)
	assert.ErrorIs(t, err, documents.ErrDocumentNotFound)
	assert.Nil(t, res)

	err = pendingRepo.Create(acc.GetIdentity().ToBytes(), testDoc.ID(), testDoc)
	assert.NoError(t, err)

	res, err = pendingService.Get(ctx, testDoc.ID(), documents.Pending)
	assert.NoError(t, err)
	assert.Equal(t, testDoc, res)
}

func Test_Integration_Service_GetNonPendingGenericDocument(t *testing.T) {
	accs, err := cfgService.GetAccounts()
	assert.NoError(t, err)
	assert.NotEmpty(t, accs)

	acc := accs[0]

	ctx := contextutil.WithAccount(context.Background(), acc)

	testDoc, err := getTestDoc(generic.Scheme)
	assert.NoError(t, err)

	err = testDoc.SetStatus(documents.Committed)
	assert.NoError(t, err)

	res, err := pendingService.Get(ctx, testDoc.ID(), documents.Committed)
	assert.NotNil(t, err)
	assert.Nil(t, res)

	err = documentRepo.Create(acc.GetIdentity().ToBytes(), testDoc.ID(), testDoc)
	assert.NoError(t, err)

	res, err = pendingService.Get(ctx, testDoc.ID(), documents.Committed)
	assert.NoError(t, err)
	assert.Equal(t, testDoc, res)
}

func Test_Integration_Service_GetNonPendingEntityDocument(t *testing.T) {
	accs, err := cfgService.GetAccounts()
	assert.NoError(t, err)
	assert.NotEmpty(t, accs)

	acc := accs[0]

	ctx := contextutil.WithAccount(context.Background(), acc)

	testDoc, err := getTestDoc(entity.Scheme)
	assert.NoError(t, err)

	err = testDoc.SetStatus(documents.Committed)
	assert.NoError(t, err)

	res, err := pendingService.Get(ctx, testDoc.ID(), documents.Committed)
	assert.NotNil(t, err)
	assert.Nil(t, res)

	err = documentRepo.Create(acc.GetIdentity().ToBytes(), testDoc.ID(), testDoc)
	assert.NoError(t, err)

	res, err = pendingService.Get(ctx, testDoc.ID(), documents.Committed)
	assert.NoError(t, err)
	assert.Equal(t, testDoc, res)
}

func Test_Integration_Service_GetVersion_FromDocumentRepository(t *testing.T) {
	accs, err := cfgService.GetAccounts()
	assert.NoError(t, err)
	assert.NotEmpty(t, accs)

	acc := accs[0]

	ctx := contextutil.WithAccount(context.Background(), acc)

	testDoc, err := getTestDoc(generic.Scheme)
	assert.NoError(t, err)

	err = testDoc.SetStatus(documents.Committed)
	assert.NoError(t, err)

	res, err := pendingService.GetVersion(ctx, testDoc.ID(), testDoc.CurrentVersion())
	assert.NotNil(t, err)
	assert.Nil(t, res)

	err = documentRepo.Create(acc.GetIdentity().ToBytes(), testDoc.CurrentVersion(), testDoc)
	assert.NoError(t, err)

	res, err = pendingService.GetVersion(ctx, testDoc.ID(), testDoc.CurrentVersion())
	assert.NoError(t, err)
	assert.Equal(t, testDoc, res)
}

func Test_Integration_Service_GetVersion_FromPendingRepository(t *testing.T) {
	accs, err := cfgService.GetAccounts()
	assert.NoError(t, err)
	assert.NotEmpty(t, accs)

	acc := accs[0]

	ctx := contextutil.WithAccount(context.Background(), acc)

	testDoc, err := getTestDoc(generic.Scheme)
	assert.NoError(t, err)

	err = testDoc.SetStatus(documents.Pending)
	assert.NoError(t, err)

	res, err := pendingService.GetVersion(ctx, testDoc.ID(), testDoc.CurrentVersion())
	assert.NotNil(t, err)
	assert.Nil(t, res)

	err = pendingRepo.Create(acc.GetIdentity().ToBytes(), testDoc.ID(), testDoc)
	assert.NoError(t, err)

	res, err = pendingService.GetVersion(ctx, testDoc.ID(), testDoc.CurrentVersion())
	assert.NoError(t, err)
	assert.Equal(t, testDoc, res)
}

func Test_Integration_Service_Create_WithoutDocumentID(t *testing.T) {
	accs, err := cfgService.GetAccounts()
	assert.NoError(t, err)
	assert.NotEmpty(t, accs)

	acc := accs[0]

	ctx := contextutil.WithAccount(context.Background(), acc)

	documentSchemes := []string{
		generic.Scheme,
		entity.Scheme,
		entityrelationship.Scheme,
	}

	for _, documentScheme := range documentSchemes {
		testName := fmt.Sprintf("create-without-id-for-%s", documentScheme)

		t.Run(testName, func(t *testing.T) {
			documentData, err := getDataForDocumentType(documentScheme)
			assert.NoError(t, err)

			payload := documents.UpdatePayload{
				CreatePayload: documents.CreatePayload{
					Scheme: documentScheme,
					Data:   documentData,
				},
			}

			doc, err := pendingService.Create(ctx, payload)
			assert.NoError(t, err)
			assert.NotNil(t, doc)

			res, err := pendingService.Get(ctx, doc.ID(), documents.Pending)
			assert.NoError(t, err)
			assert.NotNil(t, doc, res)

			payload.DocumentID = doc.ID()

			doc, err = pendingService.Create(ctx, payload)
			assert.ErrorIs(t, err, pending.ErrPendingDocumentExists)
			assert.Nil(t, doc)

		})
	}
}

func Test_Integration_Service_Create_WithDocumentID(t *testing.T) {
	accs, err := cfgService.GetAccounts()
	assert.NoError(t, err)
	assert.NotEmpty(t, accs)

	acc := accs[0]

	ctx := contextutil.WithAccount(context.Background(), acc)

	documentSchemes := []string{
		generic.Scheme,
		entity.Scheme,
		entityrelationship.Scheme,
	}

	for _, documentScheme := range documentSchemes {
		testName := fmt.Sprintf("create-with-id-for-%s", documentScheme)

		t.Run(testName, func(t *testing.T) {
			documentData, err := getDataForDocumentType(documentScheme)
			assert.NoError(t, err)

			payload := documents.UpdatePayload{
				CreatePayload: documents.CreatePayload{
					Scheme: documentScheme,
					Data:   documentData,
				},
			}

			testDoc, err := documentService.Derive(ctx, payload)
			assert.NoError(t, err)
			assert.NotNil(t, testDoc)

			err = testDoc.SetStatus(documents.Committed)
			assert.NoError(t, err)

			err = documentRepo.Create(acc.GetIdentity().ToBytes(), testDoc.ID(), testDoc)
			assert.NoError(t, err)

			payload.DocumentID = testDoc.ID()

			doc, err := pendingService.Create(ctx, payload)
			assert.NoError(t, err)
			assert.NotNil(t, doc)
			assert.Equal(t, testDoc.ID(), doc.ID())

			res, err := pendingService.Get(ctx, doc.ID(), documents.Pending)
			assert.NoError(t, err)
			assert.NotNil(t, doc, res)
		})
	}
}

func Test_Integration_Service_Clone(t *testing.T) {
	accs, err := cfgService.GetAccounts()
	assert.NoError(t, err)
	assert.NotEmpty(t, accs)

	acc := accs[0]

	ctx := contextutil.WithAccount(context.Background(), acc)

	documentSchemes := []string{
		generic.Scheme,
		entity.Scheme,
		entityrelationship.Scheme,
	}

	for _, documentScheme := range documentSchemes {
		testName := fmt.Sprintf("clone-%s", documentScheme)

		t.Run(testName, func(t *testing.T) {
			documentData, err := getDataForDocumentType(documentScheme)
			assert.NoError(t, err)

			updatePayload := documents.UpdatePayload{
				CreatePayload: documents.CreatePayload{
					Scheme: documentScheme,
					Data:   documentData,
				},
			}

			testDoc, err := documentService.Derive(ctx, updatePayload)
			assert.NoError(t, err)
			assert.NotNil(t, testDoc)

			err = testDoc.SetStatus(documents.Committed)
			assert.NoError(t, err)

			err = documentRepo.Create(acc.GetIdentity().ToBytes(), testDoc.ID(), testDoc)
			assert.NoError(t, err)

			clonePayload := documents.ClonePayload{
				Scheme:     documentScheme,
				TemplateID: testDoc.ID(),
			}

			doc, err := pendingService.Clone(ctx, clonePayload)
			assert.NoError(t, err)
			assert.NotNil(t, doc)
			assert.NotEqual(t, testDoc.ID(), doc.ID())

			res, err := pendingService.Get(ctx, doc.ID(), documents.Pending)
			assert.NoError(t, err)
			assert.NotNil(t, doc, res)

			clonePayload.TemplateID = doc.ID()

			doc, err = pendingService.Clone(ctx, clonePayload)
			assert.ErrorIs(t, err, pending.ErrPendingDocumentExists)
			assert.Nil(t, doc)
		})
	}
}

func Test_Integration_Service_Update(t *testing.T) {
	accs, err := cfgService.GetAccounts()
	assert.NoError(t, err)
	assert.NotEmpty(t, accs)

	acc := accs[0]

	ctx := contextutil.WithAccount(context.Background(), acc)

	documentSchemes := []string{
		generic.Scheme,
		entity.Scheme,
		entityrelationship.Scheme,
	}

	for _, documentScheme := range documentSchemes {
		testName := fmt.Sprintf("update-%s", documentScheme)

		t.Run(testName, func(t *testing.T) {
			documentData, err := getDataForDocumentType(documentScheme)
			assert.NoError(t, err)

			payload := documents.UpdatePayload{
				CreatePayload: documents.CreatePayload{
					Scheme: documentScheme,
					Data:   documentData,
				},
			}

			doc, err := pendingService.Create(ctx, payload)
			assert.NoError(t, err)
			assert.NotNil(t, doc)

			payload.DocumentID = doc.ID()

			doc, err = pendingService.Update(ctx, payload)
			assert.NoError(t, err)
			assert.NotNil(t, doc)
		})
	}
}

func Test_Integration_Service_Commit(t *testing.T) {
	accs, err := cfgService.GetAccounts()
	assert.NoError(t, err)
	assert.NotEmpty(t, accs)

	acc := accs[0]

	ctx := contextutil.WithAccount(context.Background(), acc)

	documentSchemes := []string{
		generic.Scheme,
		entity.Scheme,
		entityrelationship.Scheme,
	}

	for _, documentScheme := range documentSchemes {
		testName := fmt.Sprintf("commit-%s", documentScheme)

		t.Run(testName, func(t *testing.T) {
			documentData, err := getDataForDocumentType(documentScheme)
			assert.NoError(t, err)

			payload := documents.UpdatePayload{
				CreatePayload: documents.CreatePayload{
					Scheme: documentScheme,
					Data:   documentData,
				},
			}

			doc, err := pendingService.Create(ctx, payload)
			assert.NoError(t, err)
			assert.NotNil(t, doc)

			res, jobID, err := pendingService.Commit(ctx, doc.ID())
			assert.NoError(t, err)
			assert.NotEmpty(t, jobID)
			assert.NotNil(t, res)

			err = jobsUtil.WaitForJobToFinish(ctx, dispatcher, acc.GetIdentity(), jobID)
			assert.NoError(t, err)
		})
	}
}

func Test_Integration_Service_AddSignedAttribute(t *testing.T) {
	accs, err := cfgService.GetAccounts()
	assert.NoError(t, err)
	assert.NotEmpty(t, accs)

	acc := accs[0]

	ctx := contextutil.WithAccount(context.Background(), acc)

	documentSchemes := []string{
		generic.Scheme,
		entity.Scheme,
		entityrelationship.Scheme,
	}

	for _, documentScheme := range documentSchemes {
		testName := fmt.Sprintf("add-signed-attribute-%s", documentScheme)

		t.Run(testName, func(t *testing.T) {
			documentData, err := getDataForDocumentType(documentScheme)
			assert.NoError(t, err)

			payload := documents.UpdatePayload{
				CreatePayload: documents.CreatePayload{
					Scheme: documentScheme,
					Data:   documentData,
				},
			}

			doc, err := pendingService.Create(ctx, payload)
			assert.NoError(t, err)
			assert.NotNil(t, doc)

			label := "test-label"
			value := utils.RandomSlice(32)
			attrType := documents.AttrBytes

			res, err := pendingService.AddSignedAttribute(ctx, doc.ID(), label, value, attrType)
			assert.NoError(t, err)
			assert.NotNil(t, res)

			doc, err = pendingService.Get(ctx, doc.ID(), documents.Pending)
			assert.NoError(t, err)

			attrKey, err := documents.AttrKeyFromLabel(label)
			assert.NoError(t, err)

			assert.True(t, doc.AttributeExists(attrKey))
		})
	}
}

func Test_Integration_Service_RemoveCollaborators(t *testing.T) {
	accs, err := cfgService.GetAccounts()
	assert.NoError(t, err)
	assert.NotEmpty(t, accs)

	acc := accs[0]

	ctx := contextutil.WithAccount(context.Background(), acc)

	documentSchemes := []string{
		generic.Scheme,
		entity.Scheme,
		entityrelationship.Scheme,
	}

	for _, documentScheme := range documentSchemes {
		testName := fmt.Sprintf("remove-collaborators-%s", documentScheme)

		t.Run(testName, func(t *testing.T) {
			documentData, err := getDataForDocumentType(documentScheme)
			assert.NoError(t, err)

			readCollab1, err := types.NewAccountID(keyrings.BobKeyRingPair.PublicKey)
			assert.NoError(t, err)
			writeCollab1, err := types.NewAccountID(keyrings.CharlieKeyRingPair.PublicKey)
			assert.NoError(t, err)

			payload := documents.UpdatePayload{
				CreatePayload: documents.CreatePayload{
					Scheme: documentScheme,
					Data:   documentData,
					Collaborators: documents.CollaboratorsAccess{
						ReadCollaborators: []*types.AccountID{
							readCollab1,
						},
						ReadWriteCollaborators: []*types.AccountID{
							writeCollab1,
						},
					},
				},
			}

			doc, err := pendingService.Create(ctx, payload)
			assert.NoError(t, err)
			assert.NotNil(t, doc)

			res, err := pendingService.RemoveCollaborators(ctx, doc.ID(), []*types.AccountID{readCollab1, writeCollab1})
			assert.NoError(t, err)
			assert.NotNil(t, res)

			collaborators, err := res.GetCollaborators()
			assert.NoError(t, err)

			assert.NotContains(t, collaborators.ReadCollaborators, readCollab1)
			assert.NotContains(t, collaborators.ReadWriteCollaborators, writeCollab1)
		})
	}
}

func Test_Integration_Service_Role_AddGetUpdate(t *testing.T) {
	accs, err := cfgService.GetAccounts()
	assert.NoError(t, err)
	assert.NotEmpty(t, accs)

	acc := accs[0]

	ctx := contextutil.WithAccount(context.Background(), acc)

	documentSchemes := []string{
		generic.Scheme,
		entity.Scheme,
		entityrelationship.Scheme,
	}

	for _, documentScheme := range documentSchemes {
		testName := fmt.Sprintf("role-actions-%s", documentScheme)

		t.Run(testName, func(t *testing.T) {
			documentData, err := getDataForDocumentType(documentScheme)
			assert.NoError(t, err)

			collab1, err := types.NewAccountID(keyrings.BobKeyRingPair.PublicKey)
			assert.NoError(t, err)
			collab2, err := types.NewAccountID(keyrings.CharlieKeyRingPair.PublicKey)
			assert.NoError(t, err)

			payload := documents.UpdatePayload{
				CreatePayload: documents.CreatePayload{
					Scheme: documentScheme,
					Data:   documentData,
				},
			}

			doc, err := pendingService.Create(ctx, payload)
			assert.NoError(t, err)
			assert.NotNil(t, doc)

			roleKey := "role-key"

			role, err := pendingService.AddRole(ctx, doc.ID(), roleKey, []*types.AccountID{collab1})
			assert.NoError(t, err)
			assert.NotNil(t, role)

			assert.Contains(t, role.GetCollaborators(), collab1.ToBytes())

			role, err = pendingService.GetRole(ctx, doc.ID(), role.GetRoleKey())
			assert.NoError(t, err)
			assert.NotNil(t, role)

			assert.Contains(t, role.GetCollaborators(), collab1.ToBytes())

			role, err = pendingService.UpdateRole(ctx, doc.ID(), role.GetRoleKey(), []*types.AccountID{collab1, collab2})
			assert.NoError(t, err)
			assert.NotNil(t, role)

			assert.Contains(t, role.GetCollaborators(), collab1.ToBytes())
			assert.Contains(t, role.GetCollaborators(), collab2.ToBytes())
		})
	}
}

func Test_Integration_Service_TransitionRule_AddGetDelete(t *testing.T) {
	accs, err := cfgService.GetAccounts()
	assert.NoError(t, err)
	assert.NotEmpty(t, accs)

	acc := accs[0]

	ctx := contextutil.WithAccount(context.Background(), acc)

	documentSchemes := []string{
		generic.Scheme,
		entity.Scheme,
		entityrelationship.Scheme,
	}

	for _, documentScheme := range documentSchemes {
		testName := fmt.Sprintf("transition-rules-actions-%s", documentScheme)

		t.Run(testName, func(t *testing.T) {
			documentData, err := getDataForDocumentType(documentScheme)
			assert.NoError(t, err)

			payload := documents.UpdatePayload{
				CreatePayload: documents.CreatePayload{
					Scheme: documentScheme,
					Data:   documentData,
				},
			}

			doc, err := pendingService.Create(ctx, payload)
			assert.NoError(t, err)
			assert.NotNil(t, doc)

			collab1, err := types.NewAccountID(keyrings.BobKeyRingPair.PublicKey)
			assert.NoError(t, err)

			roleKey := "role-key"

			role, err := pendingService.AddRole(ctx, doc.ID(), roleKey, []*types.AccountID{collab1})
			assert.NoError(t, err)
			assert.NotNil(t, role)

			attributeRule := pending.AttributeRule{
				KeyLabel: "key-label-1",
				RoleID:   role.GetRoleKey(),
			}

			wasm, err := ioutil.ReadFile("testingutils/compute_fields/simple_average.wasm")
			assert.NoError(t, err)

			computeFieldsRule := pending.ComputeFieldsRule{
				WASM: wasm,
				AttributeLabels: []string{
					"attribute-label-1",
				},
				TargetAttributeLabel: "target-attribute-label-1",
			}

			addTransitionRules := pending.AddTransitionRules{
				AttributeRules: []pending.AttributeRule{
					attributeRule,
				},
				ComputeFieldsRules: []pending.ComputeFieldsRule{
					computeFieldsRule,
				},
			}

			rules, err := pendingService.AddTransitionRules(ctx, doc.ID(), addTransitionRules)
			assert.NoError(t, err)
			assert.Len(t, rules, 2)

			res, err := pendingService.GetTransitionRule(ctx, doc.ID(), rules[0].GetRuleKey())
			assert.NoError(t, err)
			assert.NotNil(t, res)

			res, err = pendingService.GetTransitionRule(ctx, doc.ID(), rules[1].GetRuleKey())
			assert.NoError(t, err)
			assert.NotNil(t, res)

			err = pendingService.DeleteTransitionRule(ctx, doc.ID(), rules[0].GetRuleKey())
			assert.NoError(t, err)

			err = pendingService.DeleteTransitionRule(ctx, doc.ID(), rules[1].GetRuleKey())
			assert.NoError(t, err)

			res, err = pendingService.GetTransitionRule(ctx, doc.ID(), rules[0].GetRuleKey())
			assert.NotNil(t, err)
			assert.Nil(t, res)

			res, err = pendingService.GetTransitionRule(ctx, doc.ID(), rules[1].GetRuleKey())
			assert.NotNil(t, err)
			assert.Nil(t, res)
		})
	}
}

func Test_Integration_Service_Attributes_AddDelete(t *testing.T) {
	accs, err := cfgService.GetAccounts()
	assert.NoError(t, err)
	assert.NotEmpty(t, accs)

	acc := accs[0]

	ctx := contextutil.WithAccount(context.Background(), acc)

	documentSchemes := []string{
		generic.Scheme,
		entity.Scheme,
		entityrelationship.Scheme,
	}

	for _, documentScheme := range documentSchemes {
		testName := fmt.Sprintf("attributes-actions-%s", documentScheme)

		t.Run(testName, func(t *testing.T) {
			documentData, err := getDataForDocumentType(documentScheme)
			assert.NoError(t, err)

			payload := documents.UpdatePayload{
				CreatePayload: documents.CreatePayload{
					Scheme: documentScheme,
					Data:   documentData,
				},
			}

			doc, err := pendingService.Create(ctx, payload)
			assert.NoError(t, err)
			assert.NotNil(t, doc)

			attribute1 := documents.Attribute{
				KeyLabel: "key-label-1",
				Key:      documents.AttrKey(utils.RandomByte32()),
				Value: documents.AttrVal{
					Type: documents.AttrString,
					Str:  "test-string-1",
				},
			}
			attribute2 := documents.Attribute{
				KeyLabel: "key-label-2",
				Key:      documents.AttrKey(utils.RandomByte32()),
				Value: documents.AttrVal{
					Type:  documents.AttrBytes,
					Bytes: utils.RandomSlice(32),
				},
			}

			attributes := []documents.Attribute{
				attribute1,
				attribute2,
			}

			res, err := pendingService.AddAttributes(ctx, doc.ID(), attributes)
			assert.NoError(t, err)
			assert.NotNil(t, res)

			attr, err := res.GetAttribute(attribute1.Key)
			assert.NoError(t, err)
			assert.Equal(t, attribute1, attr)

			attr, err = res.GetAttribute(attribute2.Key)
			assert.NoError(t, err)
			assert.Equal(t, attribute2, attr)

			res, err = pendingService.DeleteAttribute(ctx, doc.ID(), attribute1.Key)
			assert.NoError(t, err)

			res, err = pendingService.DeleteAttribute(ctx, doc.ID(), attribute2.Key)
			assert.NoError(t, err)

			assert.Len(t, res.GetAttributes(), 0)
		})
	}
}

func getDataForDocumentType(scheme string) ([]byte, error) {
	var data any

	switch scheme {
	case entity.Scheme:
		accountID, err := types.NewAccountID(keyrings.BobKeyRingPair.PublicKey)
		if err != nil {
			return nil, err
		}

		data = entity.Data{
			Identity:  accountID,
			LegalName: "legal-name",
			Addresses: []entity.Address{
				{
					IsMain:        true,
					Label:         "label",
					Zip:           "zip",
					State:         "state",
					Country:       "country",
					AddressLine1:  "address-line-1",
					AddressLine2:  "address-line-2",
					ContactPerson: "contactPerson",
				},
			},
			PaymentDetails: []entity.PaymentDetail{
				{
					Predefined: true,
					OtherPaymentMethod: &entity.OtherPaymentMethod{
						Identifier:        utils.RandomSlice(32),
						Type:              "type",
						PayTo:             "pay-to",
						SupportedCurrency: "support-currency",
					},
				},
			},
			Contacts: []entity.Contact{
				{
					Name:  "name",
					Title: "title",
					Email: "email",
					Phone: "phone",
					Fax:   "fax",
				},
			},
		}
	case entityrelationship.Scheme:
		ownerAccountID, err := types.NewAccountID(keyrings.CharlieKeyRingPair.PublicKey)
		if err != nil {
			return nil, err
		}
		targetAccountID, err := types.NewAccountID(keyrings.DaveKeyRingPair.PublicKey)
		if err != nil {
			return nil, err
		}

		data = entityrelationship.Data{
			OwnerIdentity:    ownerAccountID,
			EntityIdentifier: utils.RandomSlice(32),
			TargetIdentity:   targetAccountID,
		}
	default: // Data remains nil
	}

	return json.Marshal(data)
}

func getTestDoc(documentScheme string) (documents.Document, error) {
	cd, err := documents.NewCoreDocument(utils.RandomSlice(32), documents.CollaboratorsAccess{}, nil)

	if err != nil {
		return nil, err
	}

	switch documentScheme {
	case entity.Scheme:
		return &entity.Entity{CoreDocument: cd}, nil
	case entityrelationship.Scheme:
		return &entityrelationship.EntityRelationship{CoreDocument: cd}, nil
	case generic.Scheme:
		return &generic.Generic{CoreDocument: cd}, nil
	}

	return nil, fmt.Errorf("scheme %s not supported", documentScheme)
}
