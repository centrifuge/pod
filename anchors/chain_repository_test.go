// +build unit

package anchors

import (
	"context"
	"testing"
	"time"

	"github.com/centrifuge/go-centrifuge/centchain"
	"github.com/centrifuge/go-centrifuge/errors"
	testingconfig "github.com/centrifuge/go-centrifuge/testingutils/config"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/go-substrate-rpc-client/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRepository_PreCommit(t *testing.T) {
	api := new(centchain.MockAPI)
	repo := NewRepository(api)
	anchorID := AnchorID(utils.RandomByte32())
	signingRoot := DocumentRoot(utils.RandomByte32())

	// missing account
	_, _, _, err := repo.PreCommit(context.Background(), anchorID, signingRoot)
	assert.Error(t, err)

	// failed meta data latest
	ctx := testingconfig.CreateAccountContext(t, cfg)
	api.On("GetMetadataLatest").Return(nil, errors.New("failed to get metadata")).Once()
	_, _, _, err = repo.PreCommit(ctx, anchorID, signingRoot)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get metadata")

	// failed to create call
	api.On("GetMetadataLatest").Return(&types.Metadata{}, nil).Once()
	_, _, _, err = repo.PreCommit(ctx, anchorID, signingRoot)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported metadata version")

	// failed to submit extrinsic
	meta := centchain.MetaDataWithCall(PreCommit)
	api.On("GetMetadataLatest").Return(meta, nil)
	api.On("SubmitExtrinsic", meta, mock.Anything, mock.Anything).Return(nil, nil, nil, errors.New("failed to submit extrinsic")).Once()
	_, _, _, err = repo.PreCommit(ctx, anchorID, signingRoot)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to submit extrinsic")

	// success
	api.On("SubmitExtrinsic", meta, mock.Anything, mock.Anything).Return(nil, nil, nil, nil).Once()
	_, _, _, err = repo.PreCommit(ctx, anchorID, signingRoot)
	assert.NoError(t, err)
	api.AssertExpectations(t)
}

func TestRepository_Commit(t *testing.T) {
	api := new(centchain.MockAPI)
	repo := NewRepository(api)
	anchorID := AnchorID(utils.RandomByte32())
	documentRoot := DocumentRoot(utils.RandomByte32())
	proof := utils.RandomByte32()
	storedUntil := time.Now()

	// missing account
	_, _, _, err := repo.Commit(context.Background(), anchorID, documentRoot, proof, storedUntil)
	assert.Error(t, err)

	// failed meta data latest
	ctx := testingconfig.CreateAccountContext(t, cfg)
	api.On("GetMetadataLatest").Return(nil, errors.New("failed to get metadata")).Once()
	_, _, _, err = repo.Commit(ctx, anchorID, documentRoot, proof, storedUntil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get metadata")

	// failed to create call
	api.On("GetMetadataLatest").Return(&types.Metadata{}, nil).Once()
	_, _, _, err = repo.Commit(ctx, anchorID, documentRoot, proof, storedUntil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported metadata version")

	// failed to submit extrinsic
	meta := centchain.MetaDataWithCall(Commit)
	api.On("GetMetadataLatest").Return(meta, nil)
	api.On("SubmitExtrinsic", meta, mock.Anything, mock.Anything).Return(nil, nil, nil, errors.New("failed to submit extrinsic")).Once()
	_, _, _, err = repo.Commit(ctx, anchorID, documentRoot, proof, storedUntil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to submit extrinsic")

	// success
	api.On("SubmitExtrinsic", meta, mock.Anything, mock.Anything).Return(nil, nil, nil, nil).Once()
	_, _, _, err = repo.Commit(ctx, anchorID, documentRoot, proof, storedUntil)
	assert.NoError(t, err)
	api.AssertExpectations(t)
}
