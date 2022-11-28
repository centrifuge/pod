//go:build unit

package pending

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/storage"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/stretchr/testify/assert"
)

func TestRepository_Get(t *testing.T) {
	storageRepositoryMock := storage.NewRepositoryMock(t)

	repository := repo{storageRepositoryMock}

	accountID := utils.RandomSlice(32)
	documentID := utils.RandomSlice(32)

	documentMock := documents.NewDocumentMock(t)

	storageRepositoryMock.On("Get", repository.getKey(accountID, documentID)).
		Return(documentMock, nil).
		Once()

	res, err := repository.Get(accountID, documentID)
	assert.NoError(t, err)
	assert.Equal(t, documentMock, res)
}

func TestRepository_Get_StorageRepoError(t *testing.T) {
	storageRepositoryMock := storage.NewRepositoryMock(t)

	repository := repo{storageRepositoryMock}

	accountID := utils.RandomSlice(32)
	documentID := utils.RandomSlice(32)

	repoErr := errors.New("error")

	storageRepositoryMock.On("Get", repository.getKey(accountID, documentID)).
		Return(nil, repoErr).
		Once()

	res, err := repository.Get(accountID, documentID)
	assert.ErrorIs(t, err, repoErr)
	assert.Nil(t, res)
}

func TestRepository_Get_InvalidModel(t *testing.T) {
	storageRepositoryMock := storage.NewRepositoryMock(t)

	repository := repo{storageRepositoryMock}

	accountID := utils.RandomSlice(32)
	documentID := utils.RandomSlice(32)

	model := &unknownDoc{}

	storageRepositoryMock.On("Get", repository.getKey(accountID, documentID)).
		Return(model, nil).
		Once()

	res, err := repository.Get(accountID, documentID)
	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestRepository_Create(t *testing.T) {
	storageRepositoryMock := storage.NewRepositoryMock(t)

	repository := repo{storageRepositoryMock}

	accountID := utils.RandomSlice(32)
	documentID := utils.RandomSlice(32)

	documentMock := documents.NewDocumentMock(t)

	storageRepositoryMock.On("Create", repository.getKey(accountID, documentID), documentMock).
		Return(nil).
		Once()

	err := repository.Create(accountID, documentID, documentMock)
	assert.NoError(t, err)
}

func TestRepository_Create_StorageRepoError(t *testing.T) {
	storageRepositoryMock := storage.NewRepositoryMock(t)

	repository := repo{storageRepositoryMock}

	accountID := utils.RandomSlice(32)
	documentID := utils.RandomSlice(32)

	documentMock := documents.NewDocumentMock(t)

	repoErr := errors.New("error")

	storageRepositoryMock.On("Create", repository.getKey(accountID, documentID), documentMock).
		Return(repoErr).
		Once()

	err := repository.Create(accountID, documentID, documentMock)
	assert.ErrorIs(t, err, repoErr)
}

func TestRepository_Update(t *testing.T) {
	storageRepositoryMock := storage.NewRepositoryMock(t)

	repository := repo{storageRepositoryMock}

	accountID := utils.RandomSlice(32)
	documentID := utils.RandomSlice(32)

	documentMock := documents.NewDocumentMock(t)

	storageRepositoryMock.On("Update", repository.getKey(accountID, documentID), documentMock).
		Return(nil).
		Once()

	err := repository.Update(accountID, documentID, documentMock)
	assert.NoError(t, err)
}

func TestRepository_Update_StorageRepoError(t *testing.T) {
	storageRepositoryMock := storage.NewRepositoryMock(t)

	repository := repo{storageRepositoryMock}

	accountID := utils.RandomSlice(32)
	documentID := utils.RandomSlice(32)

	documentMock := documents.NewDocumentMock(t)

	repoErr := errors.New("error")

	storageRepositoryMock.On("Update", repository.getKey(accountID, documentID), documentMock).
		Return(repoErr).
		Once()

	err := repository.Update(accountID, documentID, documentMock)
	assert.ErrorIs(t, err, repoErr)
}

func TestRepository_Delete(t *testing.T) {
	storageRepositoryMock := storage.NewRepositoryMock(t)

	repository := repo{storageRepositoryMock}

	accountID := utils.RandomSlice(32)
	documentID := utils.RandomSlice(32)

	storageRepositoryMock.On("Delete", repository.getKey(accountID, documentID)).
		Return(nil).
		Once()

	err := repository.Delete(accountID, documentID)
	assert.NoError(t, err)
}

func TestRepository_Delete_StorageRepoError(t *testing.T) {
	storageRepositoryMock := storage.NewRepositoryMock(t)

	repository := repo{storageRepositoryMock}

	accountID := utils.RandomSlice(32)
	documentID := utils.RandomSlice(32)

	repoErr := errors.New("error")

	storageRepositoryMock.On("Delete", repository.getKey(accountID, documentID)).
		Return(repoErr).
		Once()

	err := repository.Delete(accountID, documentID)
	assert.ErrorIs(t, err, repoErr)
}

type unknownDoc struct {
	SomeString string `json:"some_string"`
}

func (unknownDoc) Type() reflect.Type {
	return reflect.TypeOf(unknownDoc{})
}

func (u *unknownDoc) JSON() ([]byte, error) {
	return json.Marshal(u)
}

func (u *unknownDoc) FromJSON(j []byte) error {
	return json.Unmarshal(j, u)
}
