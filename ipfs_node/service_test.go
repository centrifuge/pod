//go:build unit

package ipfs_node

import (
	"bytes"
	"context"
	"testing"

	"github.com/ipfs/go-ipfs/core/coreapi"

	"github.com/centrifuge/go-centrifuge/errors"

	"github.com/ipfs/interface-go-ipfs-core/path"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	coreAPIMock := NewCoreAPIMock(t)

	assert.NotNil(t, New(coreAPIMock))
}

func TestService_GetBlock(t *testing.T) {
	coreAPIMock := NewCoreAPIMock(t)
	blockAPIMock := NewBlockAPIMock(t)

	s := New(coreAPIMock)

	ctx := context.Background()

	blockPath := "testPath"

	coreAPIMock.On("Block").
		Return(blockAPIMock)

	byteRes := []byte("test-bytes")
	blockAPIMock.On("Get", ctx, path.New(blockPath)).
		Return(bytes.NewReader(byteRes), nil)

	b, err := s.GetBlock(ctx, blockPath)

	assert.NoError(t, err)
	assert.Equal(t, byteRes, b)
}

func TestService_GetBlock_RetrievalError(t *testing.T) {
	coreAPIMock := NewCoreAPIMock(t)
	blockAPIMock := NewBlockAPIMock(t)

	s := New(coreAPIMock)

	ctx := context.Background()

	blockPath := "testPath"

	coreAPIMock.On("Block").
		Return(blockAPIMock)

	blockAPIMock.On("Get", ctx, path.New(blockPath)).
		Return(nil, errors.New("error"))

	b, err := s.GetBlock(ctx, blockPath)

	assert.ErrorIs(t, err, ErrBlockRetrieval)
	assert.Nil(t, b)
}

type errReader struct{}

func (e *errReader) Read(_ []byte) (int, error) {
	return 0, errors.New("read err")
}

func TestService_GetBlock_ReadErr(t *testing.T) {
	coreAPIMock := NewCoreAPIMock(t)
	blockAPIMock := NewBlockAPIMock(t)

	s := New(coreAPIMock)

	ctx := context.Background()

	blockPath := "testPath"

	coreAPIMock.On("Block").
		Return(blockAPIMock)

	blockAPIMock.On("Get", ctx, path.New(blockPath)).
		Return(&errReader{}, nil)

	b, err := s.GetBlock(ctx, blockPath)

	assert.ErrorIs(t, err, ErrBlockRead)
	assert.Nil(t, b)
}

func TestService_StatBlock(t *testing.T) {
	coreAPIMock := NewCoreAPIMock(t)
	blockAPIMock := NewBlockAPIMock(t)

	s := New(coreAPIMock)

	ctx := context.Background()

	blockPath := "testPath"

	coreAPIMock.On("Block").
		Return(blockAPIMock)

	blockStat := &coreapi.BlockStat{}

	blockAPIMock.On("Stat", ctx, path.New(blockPath)).
		Return(blockStat, nil)

	r, err := s.StatBlock(ctx, blockPath)

	assert.NoError(t, err)
	assert.Equal(t, blockStat, r)
}

func TestService_StatBlock_StatError(t *testing.T) {
	coreAPIMock := NewCoreAPIMock(t)
	blockAPIMock := NewBlockAPIMock(t)

	s := New(coreAPIMock)

	ctx := context.Background()

	blockPath := "testPath"

	coreAPIMock.On("Block").
		Return(blockAPIMock)

	blockAPIMock.On("Stat", ctx, path.New(blockPath)).
		Return(nil, errors.New("error"))

	r, err := s.StatBlock(ctx, blockPath)

	assert.ErrorIs(t, err, ErrBlockStat)
	assert.Nil(t, r)
}

func TestService_PutBlock(t *testing.T) {
	coreAPIMock := NewCoreAPIMock(t)
	blockAPIMock := NewBlockAPIMock(t)

	s := New(coreAPIMock)

	ctx := context.Background()

	coreAPIMock.On("Block").
		Return(blockAPIMock)

	blockStat := &coreapi.BlockStat{}

	data := []byte("test")

	blockAPIMock.On("Put", ctx, bytes.NewReader(data)).
		Return(blockStat, nil)

	r, err := s.PutBlock(ctx, data)

	assert.NoError(t, err)
	assert.Equal(t, blockStat, r)
}

func TestService_PutBlock_PutError(t *testing.T) {
	coreAPIMock := NewCoreAPIMock(t)
	blockAPIMock := NewBlockAPIMock(t)

	s := New(coreAPIMock)

	ctx := context.Background()

	coreAPIMock.On("Block").
		Return(blockAPIMock)

	data := []byte("test")

	blockAPIMock.On("Put", ctx, bytes.NewReader(data)).
		Return(nil, errors.New("error"))

	r, err := s.PutBlock(ctx, data)

	assert.ErrorIs(t, err, ErrBlockPut)
	assert.Nil(t, r)
}
