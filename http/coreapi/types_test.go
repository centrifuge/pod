// +build unit

package coreapi

import (
	"math/big"
	"strings"
	"testing"

	coredocumentpb "github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	testingdocuments "github.com/centrifuge/go-centrifuge/testingutils/documents"
	testingidentity "github.com/centrifuge/go-centrifuge/testingutils/identity"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

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

func TestTypes_DeriveResponseHeader(t *testing.T) {
	model := new(testingdocuments.MockModel)
	model.On("GetCollaborators", mock.Anything).Return(documents.CollaboratorsAccess{}, errors.New("error fetching collaborators")).Once()
	_, err := DeriveResponseHeader(nil, model, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error fetching collaborators")
	model.AssertExpectations(t)

	id := utils.RandomSlice(32)
	did1 := testingidentity.GenerateRandomDID()
	did2 := testingidentity.GenerateRandomDID()
	ca := documents.CollaboratorsAccess{
		ReadCollaborators:      []identity.DID{did1},
		ReadWriteCollaborators: []identity.DID{did2},
	}
	model = new(testingdocuments.MockModel)
	model.On("GetCollaborators", mock.Anything).Return(ca, nil).Once()
	model.On("ID").Return(id).Once()
	model.On("CurrentVersion").Return(id).Once()
	model.On("Author").Return(nil, errors.New("somerror"))
	model.On("Timestamp").Return(nil, errors.New("somerror"))
	model.On("NFTs").Return(nil)
	model.On("CalculateTransitionRulesFingerprint").Return(utils.RandomSlice(32), nil)
	resp, err := DeriveResponseHeader(nil, model, "")
	assert.NoError(t, err)
	assert.Equal(t, hexutil.Encode(id), resp.DocumentID)
	assert.Equal(t, hexutil.Encode(id), resp.VersionID)
	assert.Len(t, resp.ReadAccess, 1)
	assert.Equal(t, resp.ReadAccess[0].String(), did1.String())
	assert.Len(t, resp.WriteAccess, 1)
	assert.Equal(t, resp.WriteAccess[0].String(), did2.String())
	model.AssertExpectations(t)
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
	regIDs := [][]byte{
		utils.RandomSlice(32),
		utils.RandomSlice(32),
	}
	tokIDs := [][]byte{
		utils.RandomSlice(32),
		utils.RandomSlice(32),
	}
	tokIDx := []*big.Int{
		big.NewInt(1),
		big.NewInt(2),
	}
	addrs := []common.Address{
		common.BytesToAddress(utils.RandomSlice(20)),
		common.BytesToAddress(utils.RandomSlice(20)),
	}
	tests := []struct {
		name         string
		TR           func() documents.TokenRegistry
		NFTs         func() []*coredocumentpb.NFT
		isErr        bool
		errLen       int
		errMsg       string
		nftLen       int
		expectedNFTs []NFT
	}{
		{
			name: "1 nft, no error",
			TR: func() documents.TokenRegistry {
				m := new(testingdocuments.MockRegistry)
				m.On("OwnerOf", mock.Anything, mock.Anything).Return(addrs[0], nil).Once()
				m.On("CurrentIndexOfToken", mock.Anything, mock.Anything).Return(tokIDx[0], nil).Once()
				return m
			},
			NFTs: func() []*coredocumentpb.NFT {
				return []*coredocumentpb.NFT{
					{
						RegistryId: regIDs[0],
						TokenId:    tokIDs[0],
					},
				}
			},
			isErr:  false,
			nftLen: 1,
			expectedNFTs: []NFT{
				{
					Registry:   hexutil.Encode(regIDs[0][:20]),
					Owner:      addrs[0].Hex(),
					TokenID:    hexutil.Encode(tokIDs[0]),
					TokenIndex: hexutil.Encode(tokIDx[0].Bytes()),
				},
			},
		},
		{
			name: "2 nft, no error",
			TR: func() documents.TokenRegistry {
				m := new(testingdocuments.MockRegistry)
				m.On("OwnerOf", mock.Anything, mock.Anything).Return(addrs[0], nil).Once()
				m.On("OwnerOf", mock.Anything, mock.Anything).Return(addrs[1], nil).Once()
				m.On("CurrentIndexOfToken", mock.Anything, mock.Anything).Return(tokIDx[0], nil).Once()
				m.On("CurrentIndexOfToken", mock.Anything, mock.Anything).Return(tokIDx[1], nil).Once()
				return m
			},
			NFTs: func() []*coredocumentpb.NFT {
				return []*coredocumentpb.NFT{
					{
						RegistryId: regIDs[0],
						TokenId:    tokIDs[0],
					},
					{
						RegistryId: regIDs[1],
						TokenId:    tokIDs[1],
					},
				}
			},
			isErr:  false,
			nftLen: 2,
			expectedNFTs: []NFT{
				{
					Registry:   hexutil.Encode(regIDs[0][:20]),
					Owner:      addrs[0].Hex(),
					TokenID:    hexutil.Encode(tokIDs[0]),
					TokenIndex: hexutil.Encode(tokIDx[0].Bytes()),
				},
				{
					Registry:   hexutil.Encode(regIDs[1][:20]),
					Owner:      addrs[1].Hex(),
					TokenID:    hexutil.Encode(tokIDs[1]),
					TokenIndex: hexutil.Encode(tokIDx[1].Bytes()),
				},
			},
		},
		{
			name: "2 nft, ownerOf error",
			TR: func() documents.TokenRegistry {
				m := new(testingdocuments.MockRegistry)
				m.On("OwnerOf", mock.Anything, mock.Anything).Return(addrs[0], errors.New("owner error")).Once()
				m.On("OwnerOf", mock.Anything, mock.Anything).Return(addrs[1], nil).Once()
				m.On("CurrentIndexOfToken", mock.Anything, mock.Anything).Return(tokIDx[0], nil).Once()
				m.On("CurrentIndexOfToken", mock.Anything, mock.Anything).Return(tokIDx[1], nil).Once()
				return m
			},
			NFTs: func() []*coredocumentpb.NFT {
				return []*coredocumentpb.NFT{
					{
						RegistryId: regIDs[0],
						TokenId:    tokIDs[0],
					},
					{
						RegistryId: regIDs[1],
						TokenId:    tokIDs[1],
					},
				}
			},
			isErr:  true,
			errLen: 1,
			errMsg: "owner",
			nftLen: 1,
			expectedNFTs: []NFT{
				{
					Registry:   hexutil.Encode(regIDs[1][:20]),
					Owner:      addrs[1].Hex(),
					TokenID:    hexutil.Encode(tokIDs[1]),
					TokenIndex: hexutil.Encode(tokIDx[1].Bytes()),
				},
			},
		},
		{
			name: "2 nft, CurrentIndexOfToken error",
			TR: func() documents.TokenRegistry {
				m := new(testingdocuments.MockRegistry)
				m.On("OwnerOf", mock.Anything, mock.Anything).Return(addrs[0], nil).Once()
				m.On("OwnerOf", mock.Anything, mock.Anything).Return(addrs[1], nil).Once()
				m.On("CurrentIndexOfToken", mock.Anything, mock.Anything).Return(tokIDx[0], errors.New("CurrentIndexOfToken error")).Once()
				m.On("CurrentIndexOfToken", mock.Anything, mock.Anything).Return(tokIDx[1], nil).Once()
				return m
			},
			NFTs: func() []*coredocumentpb.NFT {
				return []*coredocumentpb.NFT{
					{
						RegistryId: regIDs[0],
						TokenId:    tokIDs[0],
					},
					{
						RegistryId: regIDs[1],
						TokenId:    tokIDs[1],
					},
				}
			},
			isErr:  false,
			nftLen: 2,
			expectedNFTs: []NFT{
				{
					Registry:   hexutil.Encode(regIDs[0][:20]),
					Owner:      addrs[0].Hex(),
					TokenID:    hexutil.Encode(tokIDs[0]),
					TokenIndex: hexutil.Encode([]byte{}),
				},
				{
					Registry:   hexutil.Encode(regIDs[1][:20]),
					Owner:      addrs[1].Hex(),
					TokenID:    hexutil.Encode(tokIDs[1]),
					TokenIndex: hexutil.Encode(tokIDx[1].Bytes()),
				},
			},
		},
		{
			name: "2 nft, 2 CurrentIndexOfToken error",
			TR: func() documents.TokenRegistry {
				m := new(testingdocuments.MockRegistry)
				m.On("OwnerOf", mock.Anything, mock.Anything).Return(addrs[0], nil).Once()
				m.On("OwnerOf", mock.Anything, mock.Anything).Return(addrs[1], nil).Once()
				m.On("CurrentIndexOfToken", mock.Anything, mock.Anything).Return(tokIDx[0], errors.New("CurrentIndexOfToken error")).Once()
				m.On("CurrentIndexOfToken", mock.Anything, mock.Anything).Return(tokIDx[1], errors.New("CurrentIndexOfToken error")).Once()
				return m
			},
			NFTs: func() []*coredocumentpb.NFT {
				return []*coredocumentpb.NFT{
					{
						RegistryId: regIDs[0],
						TokenId:    tokIDs[0],
					},
					{
						RegistryId: regIDs[1],
						TokenId:    tokIDs[1],
					},
				}
			},
			isErr:  false,
			errMsg: "CurrentIndexOfToken",
			nftLen: 2,
			expectedNFTs: []NFT{
				{
					Registry:   hexutil.Encode(regIDs[0][:20]),
					Owner:      addrs[0].Hex(),
					TokenID:    hexutil.Encode(tokIDs[0]),
					TokenIndex: hexutil.Encode([]byte{}),
				},
				{
					Registry:   hexutil.Encode(regIDs[1][:20]),
					Owner:      addrs[1].Hex(),
					TokenID:    hexutil.Encode(tokIDs[1]),
					TokenIndex: hexutil.Encode([]byte{}),
				},
			},
		},
		{
			name: "2 nft, ownerOf and CurrentIndexOfToken error",
			TR: func() documents.TokenRegistry {
				m := new(testingdocuments.MockRegistry)
				m.On("OwnerOf", mock.Anything, mock.Anything).Return(addrs[0], errors.New("owner error")).Once()
				m.On("OwnerOf", mock.Anything, mock.Anything).Return(addrs[1], nil).Once()
				m.On("CurrentIndexOfToken", mock.Anything, mock.Anything).Return(tokIDx[0], nil).Once()
				m.On("CurrentIndexOfToken", mock.Anything, mock.Anything).Return(tokIDx[1], errors.New("CurrentIndexOfToken error")).Once()
				return m
			},
			NFTs: func() []*coredocumentpb.NFT {
				return []*coredocumentpb.NFT{
					{
						RegistryId: regIDs[0],
						TokenId:    tokIDs[0],
					},
					{
						RegistryId: regIDs[1],
						TokenId:    tokIDs[1],
					},
				}
			},
			isErr:  true,
			errLen: 1,
			errMsg: "owner",
			nftLen: 1,
			expectedNFTs: []NFT{
				{
					Registry:   hexutil.Encode(regIDs[1][:20]),
					Owner:      addrs[1].Hex(),
					TokenID:    hexutil.Encode(tokIDs[1]),
					TokenIndex: hexutil.Encode([]byte{}),
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			n, err := convertNFTs(test.TR(), test.NFTs())
			if test.isErr {
				assert.Error(t, err)
				assert.Equal(t, errors.Len(err), test.errLen)
				assert.Contains(t, err.Error(), test.errMsg)
			} else {
				assert.NoError(t, err)
			}
			assert.Len(t, n, test.nftLen)
			if test.nftLen > 0 {
				for i, nn := range n {
					assert.Equal(t, strings.ToLower(nn.Registry), strings.ToLower(test.expectedNFTs[i].Registry))
					assert.Equal(t, strings.ToLower(nn.TokenIndex), strings.ToLower(test.expectedNFTs[i].TokenIndex))
					assert.Equal(t, strings.ToLower(nn.TokenID), strings.ToLower(test.expectedNFTs[i].TokenID))
					assert.Equal(t, strings.ToLower(nn.Owner), strings.ToLower(test.expectedNFTs[i].Owner))
				}
			}
		})
	}
}
