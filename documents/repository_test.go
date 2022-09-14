//go:build unit

package documents

import (
	"testing"
	"time"

	"github.com/centrifuge/go-centrifuge/errors"

	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common/hexutil"

	"github.com/centrifuge/go-centrifuge/storage"
	"github.com/stretchr/testify/assert"
)

func TestNewDBRepository(t *testing.T) {
	storageRepoMock := storage.NewRepositoryMock(t)

	storageRepoMock.On("Register", new(latestVersion)).
		Once()

	res := NewDBRepository(storageRepoMock)
	assert.NotNil(t, res)
}

func TestRepo_Register(t *testing.T) {
	storageRepoMock := storage.NewRepositoryMock(t)

	storageRepoMock.On("Register", new(latestVersion)).
		Once()

	repo := NewDBRepository(storageRepoMock)
	assert.NotNil(t, repo)

	documentMock := NewDocumentMock(t)

	storageRepoMock.On("Register", documentMock).
		Once()

	repo.Register(documentMock)
}

func TestRepo_Exists(t *testing.T) {
	storageRepoMock := storage.NewRepositoryMock(t)

	storageRepoMock.On("Register", new(latestVersion)).
		Once()

	repo := NewDBRepository(storageRepoMock)
	assert.NotNil(t, repo)

	accountID := utils.RandomSlice(32)
	documentID := utils.RandomSlice(32)

	key := getKey(accountID, documentID)

	storageRepoMock.On("Exists", key).
		Once().
		Return(true)

	res := repo.Exists(accountID, documentID)
	assert.True(t, res)

	storageRepoMock.On("Exists", key).
		Once().
		Return(false)

	res = repo.Exists(accountID, documentID)
	assert.False(t, res)
}

func TestRepo_Get(t *testing.T) {
	storageRepoMock := storage.NewRepositoryMock(t)

	storageRepoMock.On("Register", new(latestVersion)).
		Once()

	repo := NewDBRepository(storageRepoMock)
	assert.NotNil(t, repo)

	accountID := utils.RandomSlice(32)
	documentID := utils.RandomSlice(32)

	key := getKey(accountID, documentID)

	documentMock := NewDocumentMock(t)

	storageRepoMock.On("Get", key).
		Once().
		Return(documentMock, nil)

	res, err := repo.Get(accountID, documentID)
	assert.NoError(t, err)
	assert.Equal(t, documentMock, res)
}

func TestRepo_Get_RepoError(t *testing.T) {
	storageRepoMock := storage.NewRepositoryMock(t)

	storageRepoMock.On("Register", new(latestVersion)).
		Once()

	repo := NewDBRepository(storageRepoMock)
	assert.NotNil(t, repo)

	accountID := utils.RandomSlice(32)
	documentID := utils.RandomSlice(32)

	key := getKey(accountID, documentID)

	repoErr := errors.New("error")

	storageRepoMock.On("Get", key).
		Once().
		Return(nil, repoErr)

	res, err := repo.Get(accountID, documentID)
	assert.ErrorIs(t, err, repoErr)
	assert.Nil(t, res)
}

func TestRepo_Get_InvalidModel(t *testing.T) {
	storageRepoMock := storage.NewRepositoryMock(t)

	storageRepoMock.On("Register", new(latestVersion)).
		Once()

	repo := NewDBRepository(storageRepoMock)
	assert.NotNil(t, repo)

	accountID := utils.RandomSlice(32)
	documentID := utils.RandomSlice(32)

	key := getKey(accountID, documentID)

	modelMock := storage.NewModelMock(t)

	storageRepoMock.On("Get", key).
		Once().
		Return(modelMock, nil)

	res, err := repo.Get(accountID, documentID)
	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestRepo_Create(t *testing.T) {
	storageRepoMock := storage.NewRepositoryMock(t)

	repo := &repo{db: storageRepoMock}

	accountID := utils.RandomSlice(32)
	documentID := utils.RandomSlice(32)
	documentMock := NewDocumentMock(t)

	key := getKey(accountID, documentID)

	storageRepoMock.On("Create", key, documentMock).
		Return(nil)

	for _, test := range indexUpdateTests {
		t.Run(test.name, func(t *testing.T) {
			test.expectationsFn(documentMock, storageRepoMock, accountID, documentID)

			err := repo.Create(accountID, documentID, documentMock)

			if test.expectedError {
				assert.NotNil(t, err)
				return
			}

			assert.Nil(t, err)
		})
	}
}

func TestRepo_Create_StorageRepoCreateError(t *testing.T) {
	storageRepoMock := storage.NewRepositoryMock(t)

	repo := &repo{db: storageRepoMock}

	accountID := utils.RandomSlice(32)
	documentID := utils.RandomSlice(32)
	documentMock := NewDocumentMock(t)

	key := getKey(accountID, documentID)

	repoErr := errors.New("error")

	storageRepoMock.On("Create", key, documentMock).
		Once().
		Return(repoErr)

	err := repo.Create(accountID, documentID, documentMock)
	assert.ErrorIs(t, err, repoErr)
}

func TestRepo_Create_UncommitedDocument(t *testing.T) {
	storageRepoMock := storage.NewRepositoryMock(t)

	repo := &repo{db: storageRepoMock}

	accountID := utils.RandomSlice(32)
	documentID := utils.RandomSlice(32)
	documentMock := NewDocumentMock(t)

	key := getKey(accountID, documentID)

	storageRepoMock.On("Create", key, documentMock).
		Once().
		Return(nil)

	documentMock.On("GetStatus").
		Once().
		Return(Pending)

	err := repo.Create(accountID, documentID, documentMock)
	assert.NoError(t, err)
}

func TestRepo_Update(t *testing.T) {
	storageRepoMock := storage.NewRepositoryMock(t)

	repo := &repo{db: storageRepoMock}

	accountID := utils.RandomSlice(32)
	documentID := utils.RandomSlice(32)
	documentMock := NewDocumentMock(t)

	key := getKey(accountID, documentID)

	storageRepoMock.On("Update", key, documentMock).
		Return(nil)

	for _, test := range indexUpdateTests {
		t.Run(test.name, func(t *testing.T) {
			test.expectationsFn(documentMock, storageRepoMock, accountID, documentID)

			err := repo.Update(accountID, documentID, documentMock)

			if test.expectedError {
				assert.NotNil(t, err)
				return
			}

			assert.Nil(t, err)
		})
	}
}

func TestRepo_Update_StorageRepoUpdateError(t *testing.T) {
	storageRepoMock := storage.NewRepositoryMock(t)

	repo := &repo{db: storageRepoMock}

	accountID := utils.RandomSlice(32)
	documentID := utils.RandomSlice(32)
	documentMock := NewDocumentMock(t)

	key := getKey(accountID, documentID)

	repoErr := errors.New("error")

	storageRepoMock.On("Update", key, documentMock).
		Once().
		Return(repoErr)

	err := repo.Update(accountID, documentID, documentMock)
	assert.ErrorIs(t, err, repoErr)
}

func TestRepo_Update_UncommitedDocument(t *testing.T) {
	storageRepoMock := storage.NewRepositoryMock(t)

	repo := &repo{db: storageRepoMock}

	accountID := utils.RandomSlice(32)
	documentID := utils.RandomSlice(32)
	documentMock := NewDocumentMock(t)

	key := getKey(accountID, documentID)

	storageRepoMock.On("Update", key, documentMock).
		Once().
		Return(nil)

	documentMock.On("GetStatus").
		Once().
		Return(Pending)

	err := repo.Update(accountID, documentID, documentMock)
	assert.NoError(t, err)
}

func TestRepo_GetLatest(t *testing.T) {
	storageRepoMock := storage.NewRepositoryMock(t)

	repo := &repo{db: storageRepoMock}

	accountID := utils.RandomSlice(32)
	documentID := utils.RandomSlice(32)
	currentVersion := utils.RandomSlice(32)

	lv := &latestVersion{
		CurrentVersion: currentVersion,
	}

	latestKey := getLatestKey(accountID, documentID)

	storageRepoMock.On("Get", latestKey).
		Once().
		Return(lv, nil)

	key := getKey(accountID, currentVersion)

	documentMock := NewDocumentMock(t)

	storageRepoMock.On("Get", key).
		Once().
		Return(documentMock, nil)

	res, err := repo.GetLatest(accountID, documentID)
	assert.NoError(t, err)
	assert.Equal(t, documentMock, res)
}

func TestRepo_GetLatest_GetLatestVersionError(t *testing.T) {
	storageRepoMock := storage.NewRepositoryMock(t)

	repo := &repo{db: storageRepoMock}

	accountID := utils.RandomSlice(32)
	documentID := utils.RandomSlice(32)

	latestKey := getLatestKey(accountID, documentID)

	repoErr := errors.New("error")

	storageRepoMock.On("Get", latestKey).
		Once().
		Return(nil, repoErr)

	res, err := repo.GetLatest(accountID, documentID)
	assert.ErrorIs(t, err, repoErr)
	assert.Nil(t, res)
}

func TestRepo_GetLatest_GetDocumentError(t *testing.T) {
	storageRepoMock := storage.NewRepositoryMock(t)

	repo := &repo{db: storageRepoMock}

	accountID := utils.RandomSlice(32)
	documentID := utils.RandomSlice(32)
	currentVersion := utils.RandomSlice(32)

	lv := &latestVersion{
		CurrentVersion: currentVersion,
	}

	latestKey := getLatestKey(accountID, documentID)

	storageRepoMock.On("Get", latestKey).
		Once().
		Return(lv, nil)

	key := getKey(accountID, currentVersion)

	repoErr := errors.New("error")

	storageRepoMock.On("Get", key).
		Once().
		Return(nil, repoErr)

	res, err := repo.GetLatest(accountID, documentID)
	assert.ErrorIs(t, err, repoErr)
	assert.Nil(t, res)
}

func TestRepo_StoreLatestIndex(t *testing.T) {
	storageRepoMock := storage.NewRepositoryMock(t)

	repo := &repo{db: storageRepoMock}

	key := utils.RandomSlice(32)

	currentVersion := utils.RandomSlice(32)
	nextVersion := utils.RandomSlice(32)
	timestamp := time.Now()

	documentMock := NewDocumentMock(t)

	documentMock.On("CurrentVersion").
		Return(currentVersion)

	documentMock.On("NextVersion").
		Return(nextVersion)

	documentMock.On("Timestamp").
		Return(timestamp, nil)

	lv := &latestVersion{
		CurrentVersion: currentVersion,
		NextVersion:    nextVersion,
		Timestamp:      timestamp,
	}

	storageRepoMock.On("Create", key, lv).
		Once().
		Return(nil)

	err := repo.storeLatestIndex(key, documentMock, false)
	assert.NoError(t, err)

	storageRepoMock.On("Update", key, lv).
		Once().
		Return(nil)

	err = repo.storeLatestIndex(key, documentMock, true)
	assert.NoError(t, err)
}

func TestRepo_StoreLatestIndex_RepoError(t *testing.T) {
	storageRepoMock := storage.NewRepositoryMock(t)

	repo := &repo{db: storageRepoMock}

	key := utils.RandomSlice(32)

	currentVersion := utils.RandomSlice(32)
	nextVersion := utils.RandomSlice(32)
	timestamp := time.Now()

	documentMock := NewDocumentMock(t)

	documentMock.On("CurrentVersion").
		Return(currentVersion)

	documentMock.On("NextVersion").
		Return(nextVersion)

	documentMock.On("Timestamp").
		Return(timestamp, nil)

	lv := &latestVersion{
		CurrentVersion: currentVersion,
		NextVersion:    nextVersion,
		Timestamp:      timestamp,
	}

	repoErr := errors.New("error")

	storageRepoMock.On("Create", key, lv).
		Once().
		Return(repoErr)

	err := repo.storeLatestIndex(key, documentMock, false)
	assert.ErrorIs(t, err, repoErr)

	storageRepoMock.On("Update", key, lv).
		Once().
		Return(repoErr)

	err = repo.storeLatestIndex(key, documentMock, true)
	assert.ErrorIs(t, err, repoErr)
}

func TestRepo_UpdateLatestIndex(t *testing.T) {
	storageRepoMock := storage.NewRepositoryMock(t)

	repo := &repo{db: storageRepoMock}

	accountID := utils.RandomSlice(32)
	documentID := utils.RandomSlice(32)
	documentMock := NewDocumentMock(t)

	for _, test := range indexUpdateTests {
		t.Run(test.name, func(t *testing.T) {
			test.expectationsFn(documentMock, storageRepoMock, accountID, documentID)

			err := repo.updateLatestIndex(accountID, documentMock)

			if test.expectedError {
				assert.NotNil(t, err)
				return
			}

			assert.Nil(t, err)
		})
	}
}

func TestGetLatestVersion(t *testing.T) {
	storageRepoMock := storage.NewRepositoryMock(t)

	repo := &repo{db: storageRepoMock}

	key := utils.RandomSlice(32)

	lv := &latestVersion{}

	storageRepoMock.On("Get", key).
		Once().
		Return(lv, nil)

	res, err := repo.getLatestVersion(key)
	assert.NoError(t, err)
	assert.Equal(t, lv, res)
}

func TestGetLatestVersion_RepoError(t *testing.T) {
	storageRepoMock := storage.NewRepositoryMock(t)

	repo := &repo{db: storageRepoMock}

	key := utils.RandomSlice(32)

	repoErr := errors.New("error")

	storageRepoMock.On("Get", key).
		Once().
		Return(nil, repoErr)

	res, err := repo.getLatestVersion(key)
	assert.ErrorIs(t, err, repoErr)
	assert.Nil(t, res)
}

func TestGetLatestVersion_TypeMismatch(t *testing.T) {
	storageRepoMock := storage.NewRepositoryMock(t)

	repo := &repo{db: storageRepoMock}

	key := utils.RandomSlice(32)

	storageRepoMock.On("Get", key).
		Once().
		Return(nil, nil)

	storageRepoMock.On("Delete", key).
		Once().
		Return(nil)

	res, err := repo.getLatestVersion(key)
	assert.ErrorIs(t, err, ErrDocumentNotFound)
	assert.Nil(t, res)
}

func TestGetLatestVersion_TypeMismatch_DeleteError(t *testing.T) {
	storageRepoMock := storage.NewRepositoryMock(t)

	repo := &repo{db: storageRepoMock}

	key := utils.RandomSlice(32)

	storageRepoMock.On("Get", key).
		Once().
		Return(nil, nil)

	repoErr := errors.New("error")

	storageRepoMock.On("Delete", key).
		Once().
		Return(repoErr)

	res, err := repo.getLatestVersion(key)
	assert.ErrorIs(t, err, repoErr)
	assert.Nil(t, res)
}

func TestGetKey(t *testing.T) {
	accountID := utils.RandomSlice(32)
	documentID := utils.RandomSlice(32)

	hexKey := hexutil.Encode(append(accountID, documentID...))

	res := getKey(accountID, documentID)
	assert.Equal(t, append([]byte(DocPrefix), []byte(hexKey)...), res)
}

func TestGetLatestKey(t *testing.T) {
	accountID := utils.RandomSlice(32)
	documentID := utils.RandomSlice(32)

	hexKey := hexutil.Encode(append(accountID, documentID...))

	res := getLatestKey(accountID, documentID)
	assert.Equal(t, append([]byte(LatestPrefix), []byte(hexKey)...), res)
}

var (
	indexUpdateTests = []indexUpdateTest{
		{
			name:           "create latest index",
			expectationsFn: expectCreateLatestIndex,
			expectedError:  false,
		},
		{
			name:           "create latest index with repo error",
			expectationsFn: expectCreateLatestIndexWithRepoError,
			expectedError:  true,
		},
		{
			name:           "update latest index due to version",
			expectationsFn: expectUpdateLatestIndexDueToVersion,
			expectedError:  false,
		},
		{
			name:           "update latest index due to version with repo error",
			expectationsFn: expectUpdateLatestIndexDueToVersionWithRepoError,
			expectedError:  true,
		},
		{
			name:           "update latest index due to timestamp",
			expectationsFn: expectUpdateLatestIndexDueToTimestamp,
			expectedError:  false,
		},
		{
			name:           "update latest index due to timestamp with repo error",
			expectationsFn: expectUpdateLatestIndexDueToTimestampWithRepoError,
			expectedError:  true,
		},
	}
)

type indexUpdateTest struct {
	name           string
	expectationsFn indexUpdateExpectationsFn
	expectedError  bool
}

type indexUpdateExpectationsFn func(
	documentMock *DocumentMock,
	storageRepoMock *storage.RepositoryMock,
	accountID []byte,
	documentID []byte,
)

func expectCreateLatestIndex(
	documentMock *DocumentMock,
	storageRepoMock *storage.RepositoryMock,
	accountID []byte,
	documentID []byte,
) {
	documentMock.On("GetStatus").
		Once().
		Return(Committed)

	documentMock.On("ID").
		Once().
		Return(documentID)

	latestKey := getLatestKey(accountID, documentID)

	storageRepoMock.On("Get", latestKey).
		Once().
		Return(nil, errors.New("error"))

	currentVersion := utils.RandomSlice(32)
	nextVersion := utils.RandomSlice(32)
	timestamp := time.Now()

	documentMock.On("CurrentVersion").
		Once().
		Return(currentVersion)

	documentMock.On("NextVersion").
		Once().
		Return(nextVersion)

	documentMock.On("Timestamp").
		Once().
		Return(timestamp, nil)

	latestVersion := &latestVersion{
		CurrentVersion: currentVersion,
		NextVersion:    nextVersion,
		Timestamp:      timestamp,
	}

	storageRepoMock.On("Create", latestKey, latestVersion).
		Once().
		Return(nil)
}

func expectCreateLatestIndexWithRepoError(
	documentMock *DocumentMock,
	storageRepoMock *storage.RepositoryMock,
	accountID []byte,
	documentID []byte,
) {
	documentMock.On("GetStatus").
		Once().
		Return(Committed)

	documentMock.On("ID").
		Once().
		Return(documentID)

	latestKey := getLatestKey(accountID, documentID)

	storageRepoMock.On("Get", latestKey).
		Once().
		Return(nil, errors.New("error"))

	currentVersion := utils.RandomSlice(32)
	nextVersion := utils.RandomSlice(32)
	timestamp := time.Now()

	documentMock.On("CurrentVersion").
		Once().
		Return(currentVersion)

	documentMock.On("NextVersion").
		Once().
		Return(nextVersion)

	documentMock.On("Timestamp").
		Once().
		Return(timestamp, nil)

	latestVersion := &latestVersion{
		CurrentVersion: currentVersion,
		NextVersion:    nextVersion,
		Timestamp:      timestamp,
	}

	repoErr := errors.New("error")

	storageRepoMock.On("Create", latestKey, latestVersion).
		Once().
		Return(repoErr)
}

func expectUpdateLatestIndexDueToVersion(
	documentMock *DocumentMock,
	storageRepoMock *storage.RepositoryMock,
	accountID []byte,
	documentID []byte,
) {
	documentMock.On("GetStatus").
		Once().
		Return(Committed)

	documentMock.On("ID").
		Once().
		Return(documentID)

	latestKey := getLatestKey(accountID, documentID)

	currentVersion := utils.RandomSlice(32)
	nextVersion := utils.RandomSlice(32)
	timestamp := time.Now()

	lv := &latestVersion{
		CurrentVersion: currentVersion,
		NextVersion:    nextVersion,
		Timestamp:      timestamp,
	}

	storageRepoMock.On("Get", latestKey).
		Once().
		Return(lv, nil)

	documentCurrentVersion := nextVersion
	documentNextVersion := utils.RandomSlice(32)

	documentMock.On("CurrentVersion").
		Times(2).
		Return(documentCurrentVersion)

	documentMock.On("NextVersion").
		Once().
		Return(documentNextVersion)

	documentMock.On("Timestamp").
		Once().
		Return(timestamp, nil)

	lv = &latestVersion{
		CurrentVersion: documentCurrentVersion,
		NextVersion:    documentNextVersion,
		Timestamp:      timestamp,
	}

	storageRepoMock.On("Update", latestKey, lv).
		Once().
		Return(nil)
}

func expectUpdateLatestIndexDueToVersionWithRepoError(
	documentMock *DocumentMock,
	storageRepoMock *storage.RepositoryMock,
	accountID []byte,
	documentID []byte,
) {
	documentMock.On("GetStatus").
		Once().
		Return(Committed)

	documentMock.On("ID").
		Once().
		Return(documentID)

	latestKey := getLatestKey(accountID, documentID)

	currentVersion := utils.RandomSlice(32)
	nextVersion := utils.RandomSlice(32)
	timestamp := time.Now()

	lv := &latestVersion{
		CurrentVersion: currentVersion,
		NextVersion:    nextVersion,
		Timestamp:      timestamp,
	}

	storageRepoMock.On("Get", latestKey).
		Once().
		Return(lv, nil)

	documentCurrentVersion := nextVersion
	documentNextVersion := utils.RandomSlice(32)

	documentMock.On("CurrentVersion").
		Times(2).
		Return(documentCurrentVersion)

	documentMock.On("NextVersion").
		Once().
		Return(documentNextVersion)

	documentMock.On("Timestamp").
		Once().
		Return(timestamp, nil)

	lv = &latestVersion{
		CurrentVersion: documentCurrentVersion,
		NextVersion:    documentNextVersion,
		Timestamp:      timestamp,
	}

	repoErr := errors.New("error")

	storageRepoMock.On("Update", latestKey, lv).
		Once().
		Return(repoErr)
}

func expectUpdateLatestIndexDueToTimestamp(
	documentMock *DocumentMock,
	storageRepoMock *storage.RepositoryMock,
	accountID []byte,
	documentID []byte,
) {
	documentMock.On("GetStatus").
		Once().
		Return(Committed)

	documentMock.On("ID").
		Once().
		Return(documentID)

	latestKey := getLatestKey(accountID, documentID)

	currentVersion := utils.RandomSlice(32)
	nextVersion := utils.RandomSlice(32)
	timestamp := time.Now()
	documentTimestamp := timestamp.Add(1 * time.Hour)

	lv := &latestVersion{
		CurrentVersion: currentVersion,
		NextVersion:    nextVersion,
		Timestamp:      timestamp,
	}

	storageRepoMock.On("Get", latestKey).
		Once().
		Return(lv, nil)

	documentMock.On("CurrentVersion").
		Times(2).
		Return(currentVersion)

	documentMock.On("Timestamp").
		Times(2).
		Return(documentTimestamp, nil)

	documentMock.On("NextVersion").
		Once().
		Return(nextVersion)

	lv = &latestVersion{
		CurrentVersion: currentVersion,
		NextVersion:    nextVersion,
		Timestamp:      documentTimestamp,
	}

	storageRepoMock.On("Update", latestKey, lv).
		Once().
		Return(nil)
}

func expectUpdateLatestIndexDueToTimestampWithRepoError(
	documentMock *DocumentMock,
	storageRepoMock *storage.RepositoryMock,
	accountID []byte,
	documentID []byte,
) {
	documentMock.On("GetStatus").
		Once().
		Return(Committed)

	documentMock.On("ID").
		Once().
		Return(documentID)

	latestKey := getLatestKey(accountID, documentID)

	currentVersion := utils.RandomSlice(32)
	nextVersion := utils.RandomSlice(32)
	timestamp := time.Now()
	documentTimestamp := timestamp.Add(1 * time.Hour)

	lv := &latestVersion{
		CurrentVersion: currentVersion,
		NextVersion:    nextVersion,
		Timestamp:      timestamp,
	}

	storageRepoMock.On("Get", latestKey).
		Once().
		Return(lv, nil)

	documentMock.On("CurrentVersion").
		Times(2).
		Return(currentVersion)

	documentMock.On("Timestamp").
		Times(2).
		Return(documentTimestamp, nil)

	documentMock.On("NextVersion").
		Once().
		Return(nextVersion)

	lv = &latestVersion{
		CurrentVersion: currentVersion,
		NextVersion:    nextVersion,
		Timestamp:      documentTimestamp,
	}

	repoErr := errors.New("error")

	storageRepoMock.On("Update", latestKey, lv).
		Once().
		Return(repoErr)
}
