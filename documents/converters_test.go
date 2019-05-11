// +build unit

package documents

import (
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/protobufs/gen/go/document"
	"github.com/centrifuge/go-centrifuge/testingutils/identity"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestBinaryAttachments(t *testing.T) {
	atts := []*BinaryAttachment{
		{
			Name:     "some name",
			FileType: "pdf",
			Size:     1024,
			Data:     utils.RandomSlice(32),
			Checksum: utils.RandomSlice(32),
		},

		{
			Name:     "some name 1",
			FileType: "jpeg",
			Size:     4096,
			Data:     utils.RandomSlice(32),
			Checksum: utils.RandomSlice(32),
		},
	}

	catts := ToClientAttachments(atts)
	patts := ToProtocolAttachments(atts)

	fcatts, err := FromClientAttachments(catts)
	assert.NoError(t, err)
	assert.Equal(t, atts, fcatts)

	fpatts := FromProtocolAttachments(patts)
	assert.Equal(t, atts, fpatts)

	catts[0].Checksum = "some checksum"
	_, err = FromClientAttachments(catts)
	assert.Error(t, err)

	catts[0].Data = "some data"
	_, err = FromClientAttachments(catts)
	assert.Error(t, err)
}

func TestPaymentDetails(t *testing.T) {
	did := testingidentity.GenerateRandomDID()
	dec := new(Decimal)
	err := dec.SetString("0.99")
	assert.NoError(t, err)
	details := []*PaymentDetails{
		{
			ID:     "some id",
			Payee:  &did,
			Amount: dec,
		},
	}

	cdetails := ToClientPaymentDetails(details)
	pdetails, err := ToProtocolPaymentDetails(details)
	assert.NoError(t, err)

	fcdetails, err := FromClientPaymentDetails(cdetails)
	assert.NoError(t, err)
	fpdetails, err := FromProtocolPaymentDetails(pdetails)
	assert.NoError(t, err)

	assert.Equal(t, details, fcdetails)
	assert.Equal(t, details, fpdetails)

	cdetails[0].Payee = "some did"
	_, err = FromClientPaymentDetails(cdetails)
	assert.Error(t, err)

	cdetails[0].Amount = "0.1.1"
	_, err = FromClientPaymentDetails(cdetails)
	assert.Error(t, err)

	pdetails[0].Amount = utils.RandomSlice(40)
	_, err = FromProtocolPaymentDetails(pdetails)
	assert.Error(t, err)
}

func TestCollaboratorAccess(t *testing.T) {
	id1 := testingidentity.GenerateRandomDID()
	id2 := testingidentity.GenerateRandomDID()
	id3 := testingidentity.GenerateRandomDID()
	rcs := &documentpb.ReadAccess{
		Collaborators: []string{id1.String(), id2.String()},
	}

	wcs := &documentpb.WriteAccess{
		Collaborators: []string{id2.String(), id3.String()},
	}

	ca, err := FromClientCollaboratorAccess(rcs, wcs)
	assert.NoError(t, err)
	assert.Len(t, ca.ReadCollaborators, 1)
	assert.Len(t, ca.ReadWriteCollaborators, 2)

	grcs, gwcs := ToClientCollaboratorAccess(ca)
	assert.Len(t, grcs.Collaborators, 1)
	assert.Equal(t, grcs.Collaborators[0], id1.String())
	assert.Len(t, gwcs.Collaborators, 2)
	assert.Contains(t, gwcs.Collaborators, id2.String(), id3.String())

	wcs.Collaborators[0] = "wrg id"
	_, err = FromClientCollaboratorAccess(rcs, wcs)
	assert.Error(t, err)

	rcs.Collaborators[0] = "wrg id"
	_, err = FromClientCollaboratorAccess(rcs, wcs)
	assert.Error(t, err)
}

func TestDeriveResponseHeader(t *testing.T) {
	model := new(mockModel)
	model.On("GetCollaborators", mock.Anything).Return(CollaboratorsAccess{}, errors.New("error fetching collaborators")).Once()
	_, err := DeriveResponseHeader(nil, model)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error fetching collaborators")
	model.AssertExpectations(t)

	id := utils.RandomSlice(32)
	did1 := testingidentity.GenerateRandomDID()
	did2 := testingidentity.GenerateRandomDID()
	ca := CollaboratorsAccess{
		ReadCollaborators:      []identity.DID{did1},
		ReadWriteCollaborators: []identity.DID{did2},
	}
	model = new(mockModel)
	model.On("GetCollaborators", mock.Anything).Return(ca, nil).Once()
	model.On("ID").Return(id).Once()
	model.On("CurrentVersion").Return(id).Once()
	model.On("Author").Return(nil, errors.New("somerror"))
	model.On("Timestamp").Return(nil, errors.New("somerror"))
	model.On("NFTs").Return(nil)
	resp, err := DeriveResponseHeader(nil, model)
	assert.NoError(t, err)
	assert.Equal(t, hexutil.Encode(id), resp.DocumentId)
	assert.Equal(t, hexutil.Encode(id), resp.Version)
	assert.Len(t, resp.ReadAccess.Collaborators, 1)
	assert.Equal(t, resp.ReadAccess.Collaborators[0], did1.String())
	assert.Len(t, resp.WriteAccess.Collaborators, 1)
	assert.Equal(t, resp.WriteAccess.Collaborators[0], did2.String())
	model.AssertExpectations(t)
}

func TestConvertNFTs(t *testing.T) {
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
		TR           func() TokenRegistry
		NFTs         func() []*coredocumentpb.NFT
		isErr        bool
		errLen       int
		errMsg       string
		nftLen       int
		expectedNFTs []*documentpb.NFT
	}{
		{
			name: "1 nft, no error",
			TR: func() TokenRegistry {
				m := new(mockRegistry)
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
			expectedNFTs: []*documentpb.NFT{
				{
					Registry:   hexutil.Encode(regIDs[0][:20]),
					Owner:      addrs[0].Hex(),
					TokenId:    hexutil.Encode(tokIDs[0]),
					TokenIndex: hexutil.Encode(tokIDx[0].Bytes()),
				},
			},
		},
		{
			name: "2 nft, no error",
			TR: func() TokenRegistry {
				m := new(mockRegistry)
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
			expectedNFTs: []*documentpb.NFT{
				{
					Registry:   hexutil.Encode(regIDs[0][:20]),
					Owner:      addrs[0].Hex(),
					TokenId:    hexutil.Encode(tokIDs[0]),
					TokenIndex: hexutil.Encode(tokIDx[0].Bytes()),
				},
				{
					Registry:   hexutil.Encode(regIDs[1][:20]),
					Owner:      addrs[1].Hex(),
					TokenId:    hexutil.Encode(tokIDs[1]),
					TokenIndex: hexutil.Encode(tokIDx[1].Bytes()),
				},
			},
		},
		{
			name: "2 nft, ownerOf error",
			TR: func() TokenRegistry {
				m := new(mockRegistry)
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
			expectedNFTs: []*documentpb.NFT{
				{
					Registry:   hexutil.Encode(regIDs[1][:20]),
					Owner:      addrs[1].Hex(),
					TokenId:    hexutil.Encode(tokIDs[1]),
					TokenIndex: hexutil.Encode(tokIDx[1].Bytes()),
				},
			},
		},
		{
			name: "2 nft, CurrentIndexOfToken error",
			TR: func() TokenRegistry {
				m := new(mockRegistry)
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
			isErr:  true,
			errLen: 1,
			errMsg: "CurrentIndexOfToken",
			nftLen: 1,
			expectedNFTs: []*documentpb.NFT{
				{
					Registry:   hexutil.Encode(regIDs[1][:20]),
					Owner:      addrs[1].Hex(),
					TokenId:    hexutil.Encode(tokIDs[1]),
					TokenIndex: hexutil.Encode(tokIDx[1].Bytes()),
				},
			},
		},
		{
			name: "2 nft, 2 CurrentIndexOfToken error",
			TR: func() TokenRegistry {
				m := new(mockRegistry)
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
			isErr:  true,
			errLen: 2,
			errMsg: "CurrentIndexOfToken",
			nftLen: 0,
		},
		{
			name: "2 nft, ownerOf and CurrentIndexOfToken error",
			TR: func() TokenRegistry {
				m := new(mockRegistry)
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
			errLen: 2,
			errMsg: "owner",
			nftLen: 0,
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
					assert.Equal(t, strings.ToLower(nn.TokenId), strings.ToLower(test.expectedNFTs[i].TokenId))
					assert.Equal(t, strings.ToLower(nn.Owner), strings.ToLower(test.expectedNFTs[i].Owner))
				}
			}
		})
	}
}

func TestAttributes(t *testing.T) {
	attrs := map[string]*documentpb.Attribute{
		"time_test": {
			Type:  AttrTimestamp.String(),
			Value: time.Now().UTC().Format(time.RFC3339),
		},

		"string_test": {
			Type:  AttrString.String(),
			Value: "some string",
		},

		"bytes_test": {
			Type:  AttrBytes.String(),
			Value: hexutil.Encode([]byte("some bytes data")),
		},

		"int256_test": {
			Type:  AttrInt256.String(),
			Value: "1000000001",
		},

		"decimal_test": {
			Type:  AttrDecimal.String(),
			Value: "1000.000001",
		},
	}

	attrMap, err := FromClientAttributes(attrs)
	assert.NoError(t, err)

	cattrs, err := ToClientAttributes(attrMap)
	assert.NoError(t, err)
	for k := range cattrs {
		cattrs[k].Key = ""
	}
	assert.Equal(t, attrs, cattrs)

	// failed ToClient
	var key AttrKey
	var attr Attribute
	for k := range attrMap {
		key = k
		attr = attrMap[k]
		break
	}

	attr.Value.Type = AttributeType("some type")
	attrMap[key] = attr
	cattrs, err = ToClientAttributes(attrMap)
	assert.Error(t, err)

	// failed FromClient
	attrs[""] = &documentpb.Attribute{}
	attrMap, err = FromClientAttributes(attrs)
	assert.Error(t, err)
}

func TestP2PAttributes(t *testing.T) {
	cattrs := map[string]*documentpb.Attribute{
		"time_test": {
			Type:  AttrTimestamp.String(),
			Value: time.Now().UTC().Format(time.RFC3339),
		},

		"string_test": {
			Type:  AttrString.String(),
			Value: "some string",
		},

		"bytes_test": {
			Type:  AttrBytes.String(),
			Value: hexutil.Encode([]byte("some bytes data")),
		},

		"int256_test": {
			Type:  AttrInt256.String(),
			Value: "1000000001",
		},

		"decimal_test": {
			Type:  AttrDecimal.String(),
			Value: "1000.000001",
		},
	}

	attrs, err := FromClientAttributes(cattrs)
	assert.NoError(t, err)

	pattrs, err := toProtocolAttributes(attrs)
	assert.NoError(t, err)

	attrs1, err := fromProtocolAttributes(pattrs)
	assert.NoError(t, err)
	assert.Equal(t, attrs, attrs1)

	pattrs1, err := toProtocolAttributes(attrs1)
	assert.NoError(t, err)
	assert.Equal(t, pattrs, pattrs1)

	attrKey, err := AttrKeyFromBytes(pattrs1[0].Key)
	assert.NoError(t, err)
	val := attrs1[attrKey]
	val.Value.Type = AttributeType("some type")
	attrs1[attrKey] = val
	_, err = toProtocolAttributes(attrs1)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(ErrNotValidAttrType, err))

	pattrs[0].Type = coredocumentpb.AttributeType(0)
	_, err = fromProtocolAttributes(pattrs)
	assert.Error(t, err)

	pattrs[0].Key = utils.RandomSlice(31)
	_, err = fromProtocolAttributes(pattrs)
	assert.Error(t, err)
}

func TestAttributes_signed(t *testing.T) {
	cattrs := map[string]*documentpb.Attribute{
		"time_test": {
			Type:  AttrTimestamp.String(),
			Value: time.Now().UTC().Format(time.RFC3339),
		},

		"string_test": {
			Type:  AttrString.String(),
			Value: "some string",
		},

		"bytes_test": {
			Type:  AttrBytes.String(),
			Value: hexutil.Encode([]byte("some bytes data")),
		},

		"int256_test": {
			Type:  AttrInt256.String(),
			Value: "1000000001",
		},

		"decimal_test": {
			Type:  AttrDecimal.String(),
			Value: "1000.000001",
		},
	}

	attrs, err := FromClientAttributes(cattrs)
	assert.NoError(t, err)

	label := "signed_label"
	did := testingidentity.GenerateRandomDID()
	id := utils.RandomSlice(32)
	version := utils.RandomSlice(32)
	value := utils.RandomSlice(50)

	var epayload []byte
	epayload = append(epayload, did[:]...)
	epayload = append(epayload, id...)
	epayload = append(epayload, version...)
	epayload = append(epayload, value...)

	signature := utils.RandomSlice(32)
	acc := new(mockAccount)
	acc.On("SignMsg", epayload).Return(&coredocumentpb.Signature{Signature: signature}, nil).Once()
	model := new(mockModel)
	model.On("ID").Return(id).Once()
	model.On("NextVersion").Return(version).Twice()
	attr, err := NewSignedAttribute(label, did, acc, model, value)
	assert.NoError(t, err)
	acc.AssertExpectations(t)
	model.AssertExpectations(t)
	attrs[attr.Key] = attr

	pattrs, err := toProtocolAttributes(attrs)
	assert.NoError(t, err)

	gattrs, err := fromProtocolAttributes(pattrs)
	assert.NoError(t, err)
	assert.Equal(t, attrs, gattrs)

	// wrong id
	signed := pattrs[len(pattrs)-1].GetSignedVal()
	signed.Identity = nil
	_, err = fromProtocolAttributes(pattrs)
	assert.Error(t, err)
}
