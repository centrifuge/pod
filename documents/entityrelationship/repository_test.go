//go:build unit

package entityrelationship

import (
	"testing"

	"github.com/centrifuge/pod/documents"
	"github.com/centrifuge/pod/errors"
	"github.com/centrifuge/pod/storage"
	testingcommons "github.com/centrifuge/pod/testingutils/common"
	"github.com/centrifuge/pod/utils"
	"github.com/stretchr/testify/assert"
)

func TestRepository_FindEntityRelationshipIdentifier(t *testing.T) {
	documentsRepositoryMock := documents.NewRepositoryMock(t)
	storageRepositoryMock := storage.NewRepositoryMock(t)

	repo := newDBRepository(storageRepositoryMock, documentsRepositoryMock)

	entityIdentifier := utils.RandomSlice(32)

	ownerAccountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)
	targetAccountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	entityRelationships := getTestEntityRelationships(t, 3)

	storageRepositoryMock.On(
		"GetAllByPrefix",
		documents.DocPrefix+ownerAccountID.ToHexString(),
	).Return(entityRelationships, nil).Once()

	entityRelationships[1].(*EntityRelationship).Data.EntityIdentifier = entityIdentifier
	entityRelationships[1].(*EntityRelationship).Data.TargetIdentity = targetAccountID

	res, err := repo.FindEntityRelationshipIdentifier(entityIdentifier, ownerAccountID, targetAccountID)
	assert.NoError(t, err)
	assert.Equal(t, entityRelationships[1].(*EntityRelationship).ID(), res)
}

func TestRepository_FindEntityRelationshipIdentifier_StorageError(t *testing.T) {
	documentsRepositoryMock := documents.NewRepositoryMock(t)
	storageRepositoryMock := storage.NewRepositoryMock(t)

	repo := newDBRepository(storageRepositoryMock, documentsRepositoryMock)

	entityIdentifier := utils.RandomSlice(32)

	ownerAccountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)
	targetAccountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	storageRepositoryMock.On(
		"GetAllByPrefix",
		documents.DocPrefix+ownerAccountID.ToHexString(),
	).Return(nil, errors.New("error")).Once()

	res, err := repo.FindEntityRelationshipIdentifier(entityIdentifier, ownerAccountID, targetAccountID)
	assert.True(t, errors.IsOfType(ErrDocumentsStorageRetrieval, err))
	assert.Nil(t, res)
}

func TestRepository_FindEntityRelationshipIdentifier_StorageNoResults(t *testing.T) {
	documentsRepositoryMock := documents.NewRepositoryMock(t)
	storageRepositoryMock := storage.NewRepositoryMock(t)

	repo := newDBRepository(storageRepositoryMock, documentsRepositoryMock)

	entityIdentifier := utils.RandomSlice(32)

	ownerAccountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)
	targetAccountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	storageRepositoryMock.On(
		"GetAllByPrefix",
		documents.DocPrefix+ownerAccountID.ToHexString(),
	).Return(nil, nil).Once()

	res, err := repo.FindEntityRelationshipIdentifier(entityIdentifier, ownerAccountID, targetAccountID)
	assert.ErrorIs(t, err, documents.ErrDocumentNotFound)
	assert.Nil(t, res)
}

func TestRepository_FindEntityRelationshipIdentifier_DocumentNotFound(t *testing.T) {
	documentsRepositoryMock := documents.NewRepositoryMock(t)
	storageRepositoryMock := storage.NewRepositoryMock(t)

	repo := newDBRepository(storageRepositoryMock, documentsRepositoryMock)

	entityIdentifier := utils.RandomSlice(32)

	ownerAccountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)
	targetAccountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	entityRelationships := getTestEntityRelationships(t, 3)

	storageRepositoryMock.On(
		"GetAllByPrefix",
		documents.DocPrefix+ownerAccountID.ToHexString(),
	).Return(entityRelationships, nil).Once()

	res, err := repo.FindEntityRelationshipIdentifier(entityIdentifier, ownerAccountID, targetAccountID)
	assert.ErrorIs(t, err, documents.ErrDocumentNotFound)
	assert.Nil(t, res)
}

func TestRepo_ListAllRelationships(t *testing.T) {
	documentsRepositoryMock := documents.NewRepositoryMock(t)
	storageRepositoryMock := storage.NewRepositoryMock(t)

	repo := newDBRepository(storageRepositoryMock, documentsRepositoryMock)

	entityIdentifier := utils.RandomSlice(32)

	ownerAccountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	entityRelationships := getTestEntityRelationships(t, 5)

	storageRepositoryMock.On(
		"GetAllByPrefix",
		documents.DocPrefix+ownerAccountID.ToHexString(),
	).Return(entityRelationships, nil).Once()

	entityRelationships[1].(*EntityRelationship).Data.EntityIdentifier = entityIdentifier
	entityRelationships[3].(*EntityRelationship).Data.EntityIdentifier = entityIdentifier

	res, err := repo.ListAllRelationships(entityIdentifier, ownerAccountID)
	assert.NoError(t, err)
	assert.Len(t, res, 2)
}

func TestRepo_ListAllRelationships_StorageError(t *testing.T) {
	documentsRepositoryMock := documents.NewRepositoryMock(t)
	storageRepositoryMock := storage.NewRepositoryMock(t)

	repo := newDBRepository(storageRepositoryMock, documentsRepositoryMock)

	entityIdentifier := utils.RandomSlice(32)

	ownerAccountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	storageRepositoryMock.On(
		"GetAllByPrefix",
		documents.DocPrefix+ownerAccountID.ToHexString(),
	).Return(nil, errors.New("error")).Once()

	res, err := repo.ListAllRelationships(entityIdentifier, ownerAccountID)
	assert.True(t, errors.IsOfType(ErrDocumentsStorageRetrieval, err))
	assert.Nil(t, res)
}

func TestRepo_ListAllRelationships_StorageNoResults(t *testing.T) {
	documentsRepositoryMock := documents.NewRepositoryMock(t)
	storageRepositoryMock := storage.NewRepositoryMock(t)

	repo := newDBRepository(storageRepositoryMock, documentsRepositoryMock)

	entityIdentifier := utils.RandomSlice(32)

	ownerAccountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	storageRepositoryMock.On(
		"GetAllByPrefix",
		documents.DocPrefix+ownerAccountID.ToHexString(),
	).Return(nil, nil).Once()

	res, err := repo.ListAllRelationships(entityIdentifier, ownerAccountID)
	assert.NoError(t, err)
	assert.Nil(t, res)
}

func TestRepo_ListAllRelationships_NoResults(t *testing.T) {
	documentsRepositoryMock := documents.NewRepositoryMock(t)
	storageRepositoryMock := storage.NewRepositoryMock(t)

	repo := newDBRepository(storageRepositoryMock, documentsRepositoryMock)

	entityIdentifier := utils.RandomSlice(32)

	ownerAccountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	entityRelationships := getTestEntityRelationships(t, 5)

	storageRepositoryMock.On(
		"GetAllByPrefix",
		documents.DocPrefix+ownerAccountID.ToHexString(),
	).Return(entityRelationships, nil).Once()

	res, err := repo.ListAllRelationships(entityIdentifier, ownerAccountID)
	assert.NoError(t, err)
	assert.Nil(t, res)
}

func getTestEntityRelationships(t *testing.T, count int) []storage.Model {
	var res []storage.Model

	for i := 0; i < count; i++ {
		res = append(res, getTestEntityRelationship(t, documents.CollaboratorsAccess{}, nil))
	}

	return res
}
