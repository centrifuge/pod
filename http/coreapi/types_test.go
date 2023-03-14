//go:build unit

package coreapi

import (
	"math/big"
	"math/rand"
	"testing"
	"time"

	coredocumentpb "github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types/codec"
	"github.com/centrifuge/pod/documents"
	"github.com/centrifuge/pod/errors"
	testingcommons "github.com/centrifuge/pod/testingutils/common"
	"github.com/centrifuge/pod/utils"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestTypes_DeriveResponseHeader(t *testing.T) {
	documentMock := documents.NewDocumentMock(t)

	documentID := utils.RandomSlice(32)
	currentVersion := utils.RandomSlice(32)
	nextVersion := utils.RandomSlice(32)

	accountID1, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)
	accountID2, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	ca := documents.CollaboratorsAccess{
		ReadCollaborators:      []*types.AccountID{accountID1},
		ReadWriteCollaborators: []*types.AccountID{accountID2},
	}

	documentMock.On("GetCollaborators", mock.Anything).Return(ca, nil).Once()
	documentMock.On("ID").Return(documentID).Once()
	documentMock.On("CurrentVersion").Return(currentVersion).Once()
	documentMock.On("PreviousVersion").Return(nil).Once()
	documentMock.On("NextVersion").Return(nextVersion).Once()
	documentMock.On("Author").Return(nil, errors.New("error"))
	documentMock.On("Timestamp").Return(time.Now(), errors.New("error"))
	documentMock.On("NFTs").Return(nil)

	transitionRulesFingerprint := utils.RandomSlice(32)

	documentMock.On("CalculateTransitionRulesFingerprint").Return(transitionRulesFingerprint, nil)

	resp, err := DeriveResponseHeader(documentMock, "")
	assert.NoError(t, err)
	assert.Equal(t, hexutil.Encode(documentID), resp.DocumentID)
	assert.Equal(t, "0x", resp.PreviousVersionID)
	assert.Equal(t, hexutil.Encode(currentVersion), resp.VersionID)
	assert.Equal(t, "", resp.Author)
	assert.Equal(t, "", resp.CreatedAt)
	assert.Len(t, resp.ReadAccess, 1)
	assert.Equal(t, resp.ReadAccess[0].ToHexString(), accountID1.ToHexString())
	assert.Len(t, resp.WriteAccess, 1)
	assert.Equal(t, resp.WriteAccess[0].ToHexString(), accountID2.ToHexString())
	assert.Nil(t, resp.NFTs)
	assert.Equal(t, "", resp.JobID)
	assert.Equal(t, transitionRulesFingerprint, resp.Fingerprint.Bytes())
}

func TestTypes_DeriveResponseHeader_GetCollaboratorsError(t *testing.T) {
	documentMock := documents.NewDocumentMock(t)

	documentMock.On("GetCollaborators", mock.Anything).
		Return(documents.CollaboratorsAccess{}, errors.New("error")).
		Once()

	_, err := DeriveResponseHeader(documentMock, "")
	assert.NotNil(t, err)
}

func TestTypes_DeriveResponseHeader_FingerprintCalculationError(t *testing.T) {
	documentMock := documents.NewDocumentMock(t)

	accountID1, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)
	accountID2, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	ca := documents.CollaboratorsAccess{
		ReadCollaborators:      []*types.AccountID{accountID1},
		ReadWriteCollaborators: []*types.AccountID{accountID2},
	}

	documentMock.On("GetCollaborators", mock.Anything).Return(ca, nil).Once()
	documentMock.On("Author").Return(nil, errors.New("error"))
	documentMock.On("Timestamp").Return(time.Now(), errors.New("error"))

	fingerprintCalculationError := errors.New("error")

	documentMock.On("CalculateTransitionRulesFingerprint").
		Return(nil, fingerprintCalculationError)

	_, err = DeriveResponseHeader(documentMock, "")
	assert.NotNil(t, err)
}

func TestTypes_DeriveResponseHeader_InvalidNFTs(t *testing.T) {
	documentMock := documents.NewDocumentMock(t)

	documentID := utils.RandomSlice(32)
	currentVersion := utils.RandomSlice(32)
	nextVersion := utils.RandomSlice(32)

	accountID1, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)
	accountID2, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	ca := documents.CollaboratorsAccess{
		ReadCollaborators:      []*types.AccountID{accountID1},
		ReadWriteCollaborators: []*types.AccountID{accountID2},
	}

	documentMock.On("GetCollaborators", mock.Anything).Return(ca, nil).Once()
	documentMock.On("ID").Return(documentID).Once()
	documentMock.On("CurrentVersion").Return(currentVersion).Once()
	documentMock.On("PreviousVersion").Return(nil).Once()
	documentMock.On("NextVersion").Return(nextVersion).Once()
	documentMock.On("Author").Return(nil, errors.New("error"))
	documentMock.On("Timestamp").Return(time.Now(), errors.New("error"))
	documentMock.On("NFTs").
		Return([]*coredocumentpb.NFT{
			{
				CollectionId: nil,
				ItemId:       nil,
			},
		},
		)

	transitionRulesFingerprint := utils.RandomSlice(32)

	documentMock.On("CalculateTransitionRulesFingerprint").Return(transitionRulesFingerprint, nil)

	resp, err := DeriveResponseHeader(documentMock, "")
	assert.NoError(t, err)
	assert.Equal(t, hexutil.Encode(documentID), resp.DocumentID)
	assert.Equal(t, "0x", resp.PreviousVersionID)
	assert.Equal(t, hexutil.Encode(currentVersion), resp.VersionID)
	assert.Equal(t, "", resp.Author)
	assert.Equal(t, "", resp.CreatedAt)
	assert.Len(t, resp.ReadAccess, 1)
	assert.Equal(t, resp.ReadAccess[0].ToHexString(), accountID1.ToHexString())
	assert.Len(t, resp.WriteAccess, 1)
	assert.Equal(t, resp.WriteAccess[0].ToHexString(), accountID2.ToHexString())
	assert.Nil(t, resp.NFTs)
	assert.Equal(t, "", resp.JobID)
	assert.Equal(t, transitionRulesFingerprint, resp.Fingerprint.Bytes())
}

func TestTypes_toDocumentCreatePayload(t *testing.T) {
	request := CreateDocumentRequest{Scheme: "invoice"}
	request.Data = invoiceData()

	// success
	payload, err := ToDocumentsCreatePayload(request)
	assert.NoError(t, err)
	assert.Equal(t, payload.Scheme, "invoice")
	assert.NotNil(t, payload.Data)

	// failure
	request.Attributes = map[string]AttributeRequest{
		"invalid": {Type: "unknown", Value: "some value"},
	}

	_, err = ToDocumentsCreatePayload(request)
	assert.Error(t, err)
}

func TestTypes_convertNFTs(t *testing.T) {
	collectionID1 := types.U64(rand.Uint64())
	collectionID2 := types.U64(rand.Uint64())
	itemID1 := types.NewU128(*big.NewInt(rand.Int63()))
	itemID2 := types.NewU128(*big.NewInt(rand.Int63()))

	encodedCollectionID1, err := codec.Encode(collectionID1)
	assert.NoError(t, err)
	encodedCollectionID2, err := codec.Encode(collectionID2)
	assert.NoError(t, err)
	encodedItemID1, err := codec.Encode(itemID1)
	assert.NoError(t, err)
	encodedItemID2, err := codec.Encode(itemID2)
	assert.NoError(t, err)

	tests := []struct {
		name          string
		NFTs          []*coredocumentpb.NFT
		expectedNFTs  []*NFT
		expectedError bool
	}{
		{
			name: "valid NFTs",
			NFTs: []*coredocumentpb.NFT{
				{
					CollectionId: encodedCollectionID1,
					ItemId:       encodedItemID1,
				},
				{
					CollectionId: encodedCollectionID2,
					ItemId:       encodedItemID2,
				},
			},
			expectedNFTs: []*NFT{
				{
					CollectionID: collectionID1,
					ItemID:       itemID1.String(),
				},
				{
					CollectionID: collectionID2,
					ItemID:       itemID2.String(),
				},
			},
			expectedError: false,
		},
		{
			name: "invalid collection IDs 1",
			NFTs: []*coredocumentpb.NFT{
				{
					CollectionId: nil,
					ItemId:       encodedItemID1,
				},
				{
					CollectionId: encodedCollectionID2,
					ItemId:       encodedItemID2,
				},
			},
			expectedError: true,
		},
		{
			name: "invalid collection IDs 2",
			NFTs: []*coredocumentpb.NFT{
				{
					CollectionId: encodedCollectionID1,
					ItemId:       encodedItemID1,
				},
				{
					CollectionId: nil,
					ItemId:       encodedItemID2,
				},
			},
			expectedError: true,
		},
		{
			name: "invalid item IDs 1",
			NFTs: []*coredocumentpb.NFT{
				{
					CollectionId: encodedCollectionID1,
					ItemId:       nil,
				},
				{
					CollectionId: encodedCollectionID2,
					ItemId:       encodedItemID2,
				},
			},
			expectedError: true,
		},
		{
			name: "invalid item IDs 2",
			NFTs: []*coredocumentpb.NFT{
				{
					CollectionId: encodedCollectionID1,
					ItemId:       encodedItemID2,
				},
				{
					CollectionId: encodedCollectionID2,
					ItemId:       nil,
				},
			},
			expectedError: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res, err := convertNFTs(test.NFTs)

			if test.expectedError {
				assert.NotNil(t, err)
				assert.Nil(t, res)
				return
			}

			assert.NoError(t, err)

			for _, expectedNFT := range test.expectedNFTs {
				assert.Contains(t, res, expectedNFT)
			}
		})
	}
}

func TestTypes_toAttributeMapResponse(t *testing.T) {
	dec, err := documents.NewDecimal("100001.002")
	assert.NoError(t, err)

	attrs := AttributeMapRequest{
		"string_test": {
			Type:  "string",
			Value: "hello, world!",
		},

		"decimal_test": {
			Type:  "decimal",
			Value: "100001.001",
		},

		"monetary_test": {
			Type: "monetary",
			MonetaryValue: &MonetaryValue{
				ID:      "USD",
				Value:   dec,
				ChainID: []byte{1},
			},
		},
	}

	atts, err := ToDocumentAttributes(attrs)
	assert.NoError(t, err)
	assert.Len(t, atts, 3)

	var attrList []documents.Attribute
	for _, v := range atts {
		attrList = append(attrList, v)
	}
	cattrs, err := toAttributeMapResponse(attrList)
	assert.NoError(t, err)
	assert.Len(t, cattrs, len(attrs))
	assert.Equal(t, cattrs["string_test"].Value, attrs["string_test"].Value)
	assert.Equal(t, cattrs["decimal_test"].Value, attrs["decimal_test"].Value)
	assert.Equal(t, cattrs["monetary_test"].MonetaryValue, attrs["monetary_test"].MonetaryValue)
	assert.NotEqual(t, cattrs["string_test"].Key.String(), cattrs["decimal_test"].Key.String())

	attrs["monetary_test_empty"] = AttributeRequest{Type: "monetary"}
	_, err = ToDocumentAttributes(attrs)
	assert.Error(t, err)
	delete(attrs, "monetary_test_empty")

	attrs["monetary_test_dec_empty"] = AttributeRequest{Type: "monetary", MonetaryValue: &MonetaryValue{ID: "USD", ChainID: []byte{1}}}
	_, err = ToDocumentAttributes(attrs)
	assert.Error(t, err)
	delete(attrs, "monetary_test_dec_empty")

	attrs["invalid"] = AttributeRequest{Type: "unknown", Value: "some value"}
	_, err = ToDocumentAttributes(attrs)
	assert.Error(t, err)

	attrList = append(attrList, documents.Attribute{Value: documents.AttrVal{Type: "invalid"}})
	_, err = toAttributeMapResponse(attrList)
	assert.Error(t, err)
}

func invoiceData() map[string]interface{} {
	return map[string]interface{}{
		"number":       "12345",
		"status":       "unpaid",
		"gross_amount": "12.345",
		"recipient":    "0xBAEb33a61f05e6F269f1c4b4CFF91A901B54DaF7",
		"date_due":     "2019-05-24T14:48:44.308854Z", // rfc3339nano
		"date_paid":    "2019-05-24T14:48:44Z",        // rfc3339
		"currency":     "EUR",
		"attachments": []map[string]interface{}{
			{
				"name":      "test",
				"file_type": "pdf",
				"size":      1000202,
				"data":      "0xBAEb33a61f05e6F269f1c4b4CFF91A901B54DaF7",
				"checksum":  "0xBAEb33a61f05e6F269f1c4b4CFF91A901B54DaF3",
			},
		},
	}
}
