//go:build unit

package generic

import (
	"context"
	"testing"

	coredocumentpb "github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/anypb"
)

func TestService_DeriveFromCoreDocument(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)

	srv := NewService(documentServiceMock)

	generic := getTestGeneric(t, documents.CollaboratorsAccess{}, nil)

	embedData := &anypb.Any{
		TypeUrl: generic.DocumentType(),
		Value:   utils.RandomSlice(32),
	}

	cd := &coredocumentpb.CoreDocument{
		EmbeddedData: embedData,
	}

	res, err := srv.DeriveFromCoreDocument(cd)
	assert.NoError(t, err)
	assert.IsType(t, &Generic{}, res)

	// Invalid core document

	cd = &coredocumentpb.CoreDocument{
		EmbeddedData: embedData,
		Attributes: []*coredocumentpb.Attribute{
			{
				// Invalid key.
				Key: utils.RandomSlice(31),
			},
		},
	}

	res, err = srv.DeriveFromCoreDocument(cd)
	assert.True(t, errors.IsOfType(documents.ErrDocumentUnPackingCoreDocument, err))
	assert.Nil(t, res)
}

func TestService_New(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)

	srv := NewService(documentServiceMock)

	res, err := srv.New("")
	assert.NoError(t, err)
	assert.Equal(t, new(Generic), res)

	res, err = srv.New("test")
	assert.NoError(t, err)
	assert.Equal(t, new(Generic), res)
}

func TestService_Validate(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)

	srv := NewService(documentServiceMock)

	err := srv.Validate(nil, nil, nil)
	assert.NoError(t, err)

	err = srv.Validate(context.Background(), nil, nil)
	assert.NoError(t, err)

	err = srv.Validate(context.Background(), documents.NewDocumentMock(t), nil)
	assert.NoError(t, err)

	err = srv.Validate(context.Background(), documents.NewDocumentMock(t), documents.NewDocumentMock(t))
	assert.NoError(t, err)
}
