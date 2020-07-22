// +build unit

package funding

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/centrifuge/go-centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/bootstrap/bootstrappers/testlogging"
	"github.com/centrifuge/go-centrifuge/centchain"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/config/configstore"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/documents/generic"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/ethereum"
	"github.com/centrifuge/go-centrifuge/extensions"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/identity/ideth"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/p2p"
	"github.com/centrifuge/go-centrifuge/queue"
	"github.com/centrifuge/go-centrifuge/storage/leveldb"
	"github.com/centrifuge/go-centrifuge/testingutils/config"
	"github.com/centrifuge/go-centrifuge/testingutils/documents"
	"github.com/centrifuge/go-centrifuge/testingutils/identity"
	"github.com/centrifuge/go-centrifuge/testingutils/testingjobs"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/go-centrifuge/utils/byteutils"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var ctx = map[string]interface{}{}
var cfg config.Configuration

var (
	did = testingidentity.GenerateRandomDID()
)

func TestMain(m *testing.M) {
	ethClient := new(ethereum.MockEthClient)
	ethClient.On("GetEthClient").Return(nil)
	ctx[ethereum.BootstrappedEthereumClient] = ethClient
	centChainClient := &centchain.MockAPI{}
	ctx[centchain.BootstrappedCentChainClient] = centChainClient
	jobMan := &testingjobs.MockJobManager{}
	ctx[jobs.BootstrappedService] = jobMan
	done := make(chan error)
	jobMan.On("ExecuteWithinJob", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(jobs.NilJobID(), done, nil)
	ctx[bootstrap.BootstrappedNFTService] = new(testingdocuments.MockRegistry)
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
		&queue.Starter{},
	}
	bootstrap.RunTestBootstrappers(ibootstrappers, ctx)
	cfg = ctx[bootstrap.BootstrappedConfig].(config.Configuration)
	cfg.Set("identityId", did.String())
	result := m.Run()
	bootstrap.RunTestTeardown(ibootstrappers)
	os.Exit(result)
}

func TestAttributesUtils(t *testing.T) {
	g, _ := generic.CreateGenericWithEmbedCD(t, nil, testingidentity.GenerateRandomDID(), nil)
	docSrv := new(testingdocuments.MockService)
	docSrv.On("GetCurrentVersion", mock.Anything, mock.Anything).Return(g, nil)
	data := CreateData()

	// Fill attributes list
	a, err := extensions.FillAttributeList(data, "0", fundingFieldKey)
	assert.NoError(t, err)
	// fill attribute list does not add the idx of attribute set as an attribute
	assert.Len(t, a, 12)

	// Creating an attributes list generates the correct attributes and adds an idx as an attribute
	attributes, err := extensions.CreateAttributesList(g, data, fundingFieldKey, AttrFundingLabel)
	assert.NoError(t, err)
	assert.Len(t, attributes, 13)

	for _, attribute := range attributes {
		if attribute.KeyLabel == "funding_agreement[0].currency" {
			assert.Equal(t, "eur", attribute.Value.Str)
			break
		}

		// apr was not set
		assert.NotEqual(t, "funding_agreement[0].apr", attribute.KeyLabel)
	}

	// add attributes to Document
	err = g.AddAttributes(documents.CollaboratorsAccess{}, true, attributes...)
	assert.NoError(t, err)

	var agreementID string
	for _, attribute := range attributes {
		if attribute.KeyLabel == "funding_agreement[0].agreement_id" {
			agreementID, err = attribute.Value.String()
			assert.NoError(t, err)
			break
		}
	}

	// wrong attributeSetID
	idx, err := extensions.FindAttributeSetIDX(g, "randomID", AttrFundingLabel, agreementIDLabel, fundingFieldKey)
	assert.Error(t, err)

	// correct
	idx, err = extensions.FindAttributeSetIDX(g, agreementID, AttrFundingLabel, agreementIDLabel, fundingFieldKey)
	assert.Equal(t, "0", idx)
	assert.NoError(t, err)

	// add second attributeSet
	data.AgreementID = extensions.NewAttributeSetID()
	a2, err := extensions.CreateAttributesList(g, data, fundingFieldKey, AttrFundingLabel)
	assert.NoError(t, err)

	var aID string
	for _, attribute := range a2 {
		if attribute.KeyLabel == "funding_agreement[1].agreement_id" {
			aID, err = attribute.Value.String()
			assert.NoError(t, err)
			//break
		}
	}

	err = g.AddAttributes(documents.CollaboratorsAccess{}, true, a2...)
	assert.NoError(t, err)

	// latest idx
	model, err := docSrv.GetCurrentVersion(context.Background(), g.Document.DocumentIdentifier)
	assert.NoError(t, err)

	lastIdx, err := extensions.GetArrayLatestIDX(model, AttrFundingLabel)
	assert.NoError(t, err)

	n, err := documents.NewInt256("1")
	assert.NoError(t, err)
	assert.Equal(t, lastIdx, n)

	// index should be 1
	idx, err = extensions.FindAttributeSetIDX(g, aID, AttrFundingLabel, agreementIDLabel, fundingFieldKey)
	assert.Equal(t, "1", idx)
	assert.NoError(t, err)

	// delete the first attribute set
	idx, err = extensions.FindAttributeSetIDX(g, agreementID, AttrFundingLabel, agreementIDLabel, fundingFieldKey)
	assert.NoError(t, err)

	model, err = extensions.DeleteAttributesSet(model, Data{}, idx, fundingFieldKey)
	assert.NoError(t, err)
	assert.Len(t, model.GetAttributes(), 13)

	// error when trying to delete non existing attribute set
	idx, err = extensions.FindAttributeSetIDX(g, agreementID, AttrFundingLabel, agreementIDLabel, fundingFieldKey)
	assert.Error(t, err)

	// check that latest idx is still 1 even though the first set of attributes have been deleted ?
	latest, err := extensions.GetArrayLatestIDX(model, AttrFundingLabel)
	assert.NoError(t, err)
	assert.Equal(t, latest, n)

	// non existent typeLabel for attribute set
	_, err = extensions.GetArrayLatestIDX(model, "randomLabel")
	assert.Error(t, err)

	// check that we can no longer find the attributes from the first set
	idx, err = extensions.FindAttributeSetIDX(g, agreementID, AttrFundingLabel, agreementIDLabel, fundingFieldKey)
	assert.Error(t, err)

	// test increment array attr idx
	n, err = documents.NewInt256("2")
	assert.NoError(t, err)

	newIdx, err := extensions.IncrementArrayAttrIDX(model, AttrFundingLabel)
	assert.NoError(t, err)

	v, err := newIdx.Value.String()
	assert.NoError(t, err)
	assert.Equal(t, "2", v)
	assert.Equal(t, AttrFundingLabel, newIdx.KeyLabel)
}

func invalidData() Data {
	return Data{
		Currency:              "eur",
		Days:                  "90",
		Amount:                "1000",
		RepaymentAmount:       "1200.12",
		Fee:                   "10",
		BorrowerID:            "",
		FunderID:              testingidentity.GenerateRandomDID().String(),
		NFTAddress:            hexutil.Encode(utils.RandomSlice(32)),
		RepaymentDueDate:      time.Now().UTC().Format(time.RFC3339),
		RepaymentOccurredDate: time.Now().UTC().Format(time.RFC3339),
		PaymentDetailsID:      hexutil.Encode(utils.RandomSlice(32)),
	}
}

func TestService_CreateFundingAgreement(t *testing.T) {
	// missing document.
	docSrv := new(testingdocuments.MockService)
	docSrv.On("GetCurrentVersion", mock.Anything).Return(nil, errors.New("failed to get document")).Once()
	srv := DefaultService(docSrv, nil)
	docID := utils.RandomSlice(32)
	ctx := context.Background()
	_, _, err := srv.CreateFundingAgreement(ctx, docID, new(Data))
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrDocumentNotFound, err))

	// failed to create attribute
	m := new(testingdocuments.MockModel)
	docSrv.On("GetCurrentVersion", mock.Anything).Return(m, nil)
	m.On("AttributeExists", mock.Anything).Return(true).Once()
	m.On("GetAttribute", mock.Anything).Return(documents.Attribute{}, errors.New("attribute not found")).Once()
	_, _, err = srv.CreateFundingAgreement(ctx, docID, new(Data))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "attribute not found")

	// invalid dids
	data := invalidData()
	m.On("AttributeExists", mock.Anything).Return(false)
	_, _, err = srv.CreateFundingAgreement(ctx, docID, &data)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(identity.ErrMalformedAddress, err))

	// failed to add attributes
	data = CreateData()
	m.On("AddAttributes", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("failed to add attrs")).Once()
	m.On("AttributeExists", mock.Anything).Return(false)
	_, _, err = srv.CreateFundingAgreement(ctx, docID, &data)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to add attrs")

	// failed to update document
	m.On("AddAttributes", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	docSrv.On("Update", ctx, m).Return(nil, jobs.NilJobID(), errors.New("failed to update")).Once()
	_, _, err = srv.CreateFundingAgreement(ctx, docID, &data)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update")

	// success
	docSrv.On("Update", ctx, m).Return(m, jobs.NewJobID(), nil)
	d, _, err := srv.CreateFundingAgreement(ctx, docID, &data)
	assert.NoError(t, err)
	assert.Equal(t, d, m)
	docSrv.AssertExpectations(t)
	m.AssertExpectations(t)
}

func TestService_UpdateFundingAgreement(t *testing.T) {
	// missing document.
	docSrv := new(testingdocuments.MockService)
	docSrv.On("GetCurrentVersion", mock.Anything).Return(nil, errors.New("failed to get document")).Once()
	srv := DefaultService(docSrv, nil)
	docID := utils.RandomSlice(32)
	fundingID := utils.RandomSlice(32)
	ctx := testingconfig.CreateAccountContext(t, cfg)
	_, _, err := srv.UpdateFundingAgreement(ctx, docID, fundingID, new(Data))
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrDocumentNotFound, err))

	// missing attribute
	g, _ := generic.CreateGenericWithEmbedCD(t, ctx, did, nil)
	docID = g.ID()
	docSrv.On("GetCurrentVersion", mock.Anything).Return(g, nil)
	_, _, err = srv.UpdateFundingAgreement(ctx, docID, fundingID, new(Data))
	assert.Error(t, err)

	// invalid identities
	data := CreateData()
	fundingID, err = hexutil.Decode(data.AgreementID)
	assert.NoError(t, err)
	attrs, err := extensions.CreateAttributesList(g, data, fundingFieldKey, AttrFundingLabel)
	assert.NoError(t, err)
	err = g.AddAttributes(documents.CollaboratorsAccess{}, false, attrs...)
	assert.NoError(t, err)
	_, _, err = srv.UpdateFundingAgreement(ctx, docID, fundingID, new(Data))
	assert.Error(t, err)

	// update fails
	err = g.AddAttributes(documents.CollaboratorsAccess{}, false, attrs...)
	assert.NoError(t, err)
	docSrv.On("Update", mock.Anything, mock.Anything).Return(nil, jobs.NilJobID(), errors.New("update failed")).Once()
	_, _, err = srv.UpdateFundingAgreement(ctx, docID, fundingID, &data)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "update failed")

	// success
	err = g.AddAttributes(documents.CollaboratorsAccess{}, false, attrs...)
	assert.NoError(t, err)
	docSrv.On("Update", mock.Anything, mock.Anything).Return(g, jobs.NewJobID(), nil).Once()
	m, _, err := srv.UpdateFundingAgreement(ctx, docID, fundingID, &data)
	assert.NoError(t, err)
	assert.Equal(t, m, g)
	docSrv.AssertExpectations(t)
}

func TestService_SignFundingAgreement(t *testing.T) {
	// missing agreement
	ctx := testingconfig.CreateAccountContext(t, cfg)
	g, _ := generic.CreateGenericWithEmbedCD(t, ctx, did, nil)
	docSrv := new(testingdocuments.MockService)
	s := DefaultService(docSrv, nil)
	docID := g.ID()
	docSrv.On("GetCurrentVersion", docID).Return(g, nil)
	_, _, err := s.SignFundingAgreement(ctx, docID, utils.RandomSlice(32))
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(extensions.ErrAttributeSetNotFound, err))

	// failed update
	data := CreateData()
	attrs, err := extensions.CreateAttributesList(g, data, fundingFieldKey, AttrFundingLabel)
	assert.NoError(t, err)
	err = g.AddAttributes(documents.CollaboratorsAccess{}, false, attrs...)
	assert.NoError(t, err)
	fundingID, err := hexutil.Decode(data.AgreementID)
	assert.NoError(t, err)
	docSrv.On("Update", ctx, g).Return(nil, nil, errors.New("failed to update")).Once()
	_, _, err = s.SignFundingAgreement(ctx, docID, fundingID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update")

	// success
	docSrv.On("Update", ctx, g).Return(g, jobs.NewJobID(), nil)
	m, _, err := s.SignFundingAgreement(ctx, docID, fundingID)
	assert.NoError(t, err)
	assert.Equal(t, g, m)
	docSrv.AssertExpectations(t)
}

func TestService_GetDataAndSignatures(t *testing.T) {
	ctx := testingconfig.CreateAccountContext(t, cfg)
	g, _ := generic.CreateGenericWithEmbedCD(t, ctx, did, nil)
	docSrv := new(testingdocuments.MockService)
	srv := DefaultService(docSrv, nil)

	// missing funding id
	fundingID := byteutils.HexBytes(utils.RandomSlice(32)).String()
	_, _, err := srv.GetDataAndSignatures(ctx, g, fundingID, "")
	assert.Error(t, err)

	// success
	data := CreateData()
	attrs, err := extensions.CreateAttributesList(g, data, fundingFieldKey, AttrFundingLabel)
	assert.NoError(t, err)
	err = g.AddAttributes(documents.CollaboratorsAccess{}, false, attrs...)
	assert.NoError(t, err)
	data1, sigs, err := srv.GetDataAndSignatures(ctx, g, data.AgreementID, "")
	assert.NoError(t, err)
	assert.Equal(t, data, data1)
	assert.Len(t, sigs, 0)
}
