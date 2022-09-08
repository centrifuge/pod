//go:build unit

package anchors

import (
	"context"
	"testing"
	"time"

	"github.com/centrifuge/go-substrate-rpc-client/v4/signature"

	centMocks "github.com/centrifuge/go-centrifuge/centchain"
	"github.com/centrifuge/go-centrifuge/config"
	configMocks "github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/crypto"
	"github.com/centrifuge/go-centrifuge/errors"
	proxyMocks "github.com/centrifuge/go-centrifuge/identity/v2/proxy"
	"github.com/centrifuge/go-centrifuge/testingutils"
	testingcommons "github.com/centrifuge/go-centrifuge/testingutils/commons"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/blake2b"
)

func TestService_GetAnchorData(t *testing.T) {
	centAPIMock := centMocks.NewAPIMock(t)
	proxyAPIMock := proxyMocks.NewProxyAPIMock(t)
	cfgServiceMock := config.NewServiceMock(t)

	anchorLifespan := 1 * time.Minute

	service := newService(anchorLifespan, cfgServiceMock, centAPIMock, proxyAPIMock)

	_, id, err := crypto.GenerateHashPair(32)
	assert.NoError(t, err)
	anchorID, err := ToAnchorID(id)
	assert.NoError(t, err)
	anchorIDHash := types.NewHash(anchorID[:])

	signingRoot := utils.RandomByte32()
	proof := utils.RandomByte32()
	b2bHash, err := blake2b.New256(nil)
	assert.NoError(t, err)
	_, err = b2bHash.Write(append(signingRoot[:], proof[:]...))
	assert.NoError(t, err)
	docRoot, err := ToDocumentRoot(b2bHash.Sum(nil))
	assert.NoError(t, err)
	docRootHash := types.NewHash(docRoot[:])

	centAPIMock.On("Call", mock.Anything, getByID, anchorIDHash).
		Run(func(args mock.Arguments) {
			ad := args.Get(0).(*AnchorData)
			ad.AnchorID = anchorIDHash
			ad.BlockNumber = 1
			ad.DocumentRoot = docRootHash
		}).
		Return(nil)

	docRootRes, anchoredTime, err := service.GetAnchorData(anchorID)
	assert.NoError(t, err)
	assert.Equal(t, docRoot, docRootRes)
	assert.Equal(t, time.Unix(0, 0), anchoredTime)
}

func TestService_GetAnchorData_APICallError(t *testing.T) {
	centAPIMock := centMocks.NewAPIMock(t)
	proxyAPIMock := proxyMocks.NewProxyAPIMock(t)
	cfgServiceMock := config.NewServiceMock(t)

	anchorLifespan := 1 * time.Minute

	service := newService(anchorLifespan, cfgServiceMock, centAPIMock, proxyAPIMock)

	_, id, err := crypto.GenerateHashPair(32)
	assert.NoError(t, err)
	anchorID, err := ToAnchorID(id)
	assert.NoError(t, err)
	anchorIDHash := types.NewHash(anchorID[:])

	var (
		docRoot      DocumentRoot
		anchoredTime time.Time
	)

	centAPIMock.On("Call", mock.Anything, getByID, anchorIDHash).
		Return(errors.New("api error"))

	docRootRes, anchoredTimeRes, err := service.GetAnchorData(anchorID)
	assert.ErrorIs(t, err, ErrAnchorRetrieval)
	assert.Equal(t, docRoot, docRootRes)
	assert.Equal(t, anchoredTime, anchoredTimeRes)
}

func TestService_GetAnchorData_EmptyDocumentRoot(t *testing.T) {
	centAPIMock := centMocks.NewAPIMock(t)
	proxyAPIMock := proxyMocks.NewProxyAPIMock(t)
	cfgServiceMock := config.NewServiceMock(t)

	anchorLifespan := 1 * time.Minute

	service := newService(anchorLifespan, cfgServiceMock, centAPIMock, proxyAPIMock)

	_, id, err := crypto.GenerateHashPair(32)
	assert.NoError(t, err)
	anchorID, err := ToAnchorID(id)
	assert.NoError(t, err)
	anchorIDHash := types.NewHash(anchorID[:])

	signingRoot := utils.RandomByte32()
	proof := utils.RandomByte32()
	b2bHash, err := blake2b.New256(nil)
	assert.NoError(t, err)
	_, err = b2bHash.Write(append(signingRoot[:], proof[:]...))
	assert.NoError(t, err)

	var (
		docRoot      DocumentRoot
		anchoredTime time.Time
	)

	centAPIMock.On("Call", mock.Anything, getByID, anchorIDHash).
		Run(func(args mock.Arguments) {
			ad := args.Get(0).(*AnchorData)
			ad.AnchorID = anchorIDHash
			ad.BlockNumber = 1
		}).
		Return(nil)

	docRootRes, anchoredTimeRes, err := service.GetAnchorData(anchorID)
	assert.ErrorIs(t, err, ErrEmptyDocumentRoot)
	assert.Equal(t, docRoot, docRootRes)
	assert.Equal(t, anchoredTime, anchoredTimeRes)
}

func TestService_PreCommitAnchor(t *testing.T) {
	centAPIMock := centMocks.NewAPIMock(t)
	proxyAPIMock := proxyMocks.NewProxyAPIMock(t)
	cfgServiceMock := config.NewServiceMock(t)

	anchorLifespan := 1 * time.Minute

	service := newService(anchorLifespan, cfgServiceMock, centAPIMock, proxyAPIMock)

	mockAccount := configMocks.NewAccountMock(t)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	mockAccount.On("GetIdentity").Return(accountID)

	ctx := contextutil.WithAccount(context.Background(), mockAccount)

	meta, err := testingutils.GetTestMetadata()
	assert.NoError(t, err)

	centAPIMock.On("GetMetadataLatest").
		Return(meta, nil)

	podOperatorMock := config.NewPodOperatorMock(t)

	cfgServiceMock.On("GetPodOperator").
		Return(podOperatorMock, nil)

	var krp signature.KeyringPair

	podOperatorMock.On("ToKeyringPair").
		Return(krp)

	_, id, err := crypto.GenerateHashPair(32)
	assert.NoError(t, err)
	anchorID, err := ToAnchorID(id)
	assert.NoError(t, err)

	signingRoot := utils.RandomByte32()

	call, err := types.NewCall(meta, preCommit, types.NewHash(anchorID[:]), types.NewHash(signingRoot[:]))
	assert.NoError(t, err)

	proxyAPIMock.On("ProxyCall", ctx, accountID, krp, types.NewOption(types.PodOperation), call).
		Return(nil, nil)

	err = service.PreCommitAnchor(ctx, anchorID, signingRoot)
	assert.NoError(t, err)
}

func TestService_PreCommitAnchor_AccountContextError(t *testing.T) {
	centAPIMock := centMocks.NewAPIMock(t)
	proxyAPIMock := proxyMocks.NewProxyAPIMock(t)
	cfgServiceMock := config.NewServiceMock(t)

	anchorLifespan := 1 * time.Minute

	service := newService(anchorLifespan, cfgServiceMock, centAPIMock, proxyAPIMock)

	_, id, err := crypto.GenerateHashPair(32)
	assert.NoError(t, err)
	anchorID, err := ToAnchorID(id)
	assert.NoError(t, err)

	signingRoot := utils.RandomByte32()

	err = service.PreCommitAnchor(context.Background(), anchorID, signingRoot)
	assert.ErrorIs(t, err, errors.ErrContextAccountRetrieval)
}

func TestService_PreCommitAnchor_MetadataError(t *testing.T) {
	centAPIMock := centMocks.NewAPIMock(t)
	proxyAPIMock := proxyMocks.NewProxyAPIMock(t)
	cfgServiceMock := config.NewServiceMock(t)

	anchorLifespan := 1 * time.Minute

	service := newService(anchorLifespan, cfgServiceMock, centAPIMock, proxyAPIMock)

	mockAccount := configMocks.NewAccountMock(t)

	ctx := contextutil.WithAccount(context.Background(), mockAccount)

	centAPIMock.On("GetMetadataLatest").
		Return(nil, errors.New("metadata error"))

	_, id, err := crypto.GenerateHashPair(32)
	assert.NoError(t, err)
	anchorID, err := ToAnchorID(id)
	assert.NoError(t, err)

	signingRoot := utils.RandomByte32()

	err = service.PreCommitAnchor(ctx, anchorID, signingRoot)
	assert.ErrorIs(t, err, errors.ErrMetadataRetrieval)
}

func TestService_PreCommitAnchor_CallCreationError(t *testing.T) {
	centAPIMock := centMocks.NewAPIMock(t)
	proxyAPIMock := proxyMocks.NewProxyAPIMock(t)
	cfgServiceMock := config.NewServiceMock(t)

	anchorLifespan := 1 * time.Minute

	service := newService(anchorLifespan, cfgServiceMock, centAPIMock, proxyAPIMock)

	mockAccount := configMocks.NewAccountMock(t)

	ctx := contextutil.WithAccount(context.Background(), mockAccount)

	var meta types.Metadata

	// NOTE - types.MetadataV14Data does not have info on the Anchor pallet,
	// causing types.NewCall to fail.
	err := types.DecodeFromHex(types.MetadataV14Data, &meta)
	assert.NoError(t, err)

	centAPIMock.On("GetMetadataLatest").
		Return(&meta, nil)

	_, id, err := crypto.GenerateHashPair(32)
	assert.NoError(t, err)
	anchorID, err := ToAnchorID(id)
	assert.NoError(t, err)

	signingRoot := utils.RandomByte32()

	err = service.PreCommitAnchor(ctx, anchorID, signingRoot)
	assert.ErrorIs(t, err, errors.ErrCallCreation)
}

func TestService_PreCommitAnchor_PodOperatorError(t *testing.T) {
	centAPIMock := centMocks.NewAPIMock(t)
	proxyAPIMock := proxyMocks.NewProxyAPIMock(t)
	cfgServiceMock := config.NewServiceMock(t)

	anchorLifespan := 1 * time.Minute

	service := newService(anchorLifespan, cfgServiceMock, centAPIMock, proxyAPIMock)

	mockAccount := configMocks.NewAccountMock(t)

	ctx := contextutil.WithAccount(context.Background(), mockAccount)

	meta, err := testingutils.GetTestMetadata()
	assert.NoError(t, err)

	centAPIMock.On("GetMetadataLatest").
		Return(meta, nil)

	cfgServiceMock.On("GetPodOperator").
		Return(nil, errors.New("error"))

	_, id, err := crypto.GenerateHashPair(32)
	assert.NoError(t, err)
	anchorID, err := ToAnchorID(id)
	assert.NoError(t, err)

	signingRoot := utils.RandomByte32()

	err = service.PreCommitAnchor(ctx, anchorID, signingRoot)
	assert.ErrorIs(t, err, errors.ErrPodOperatorRetrieval)
}

func TestService_PreCommitAnchor_ProxyCallError(t *testing.T) {
	centAPIMock := centMocks.NewAPIMock(t)
	proxyAPIMock := proxyMocks.NewProxyAPIMock(t)
	cfgServiceMock := config.NewServiceMock(t)

	anchorLifespan := 1 * time.Minute

	service := newService(anchorLifespan, cfgServiceMock, centAPIMock, proxyAPIMock)

	mockAccount := configMocks.NewAccountMock(t)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	mockAccount.On("GetIdentity").Return(accountID)

	ctx := contextutil.WithAccount(context.Background(), mockAccount)

	meta, err := testingutils.GetTestMetadata()
	assert.NoError(t, err)

	centAPIMock.On("GetMetadataLatest").
		Return(meta, nil)

	podOperatorMock := config.NewPodOperatorMock(t)

	cfgServiceMock.On("GetPodOperator").
		Return(podOperatorMock, nil)

	var krp signature.KeyringPair

	podOperatorMock.On("ToKeyringPair").
		Return(krp)

	_, id, err := crypto.GenerateHashPair(32)
	assert.NoError(t, err)
	anchorID, err := ToAnchorID(id)
	assert.NoError(t, err)

	signingRoot := utils.RandomByte32()

	call, err := types.NewCall(meta, preCommit, types.NewHash(anchorID[:]), types.NewHash(signingRoot[:]))
	assert.NoError(t, err)

	proxyAPIMock.On("ProxyCall", ctx, accountID, krp, types.NewOption(types.PodOperation), call).
		Return(nil, errors.New("proxy call error"))

	err = service.PreCommitAnchor(ctx, anchorID, signingRoot)
	assert.ErrorIs(t, err, errors.ErrProxyCall)
}

func TestService_CommitAnchor(t *testing.T) {
	centAPIMock := centMocks.NewAPIMock(t)
	proxyAPIMock := proxyMocks.NewProxyAPIMock(t)
	cfgServiceMock := config.NewServiceMock(t)

	anchorLifespan := 1 * time.Minute

	service := newService(anchorLifespan, cfgServiceMock, centAPIMock, proxyAPIMock)

	mockAccount := configMocks.NewAccountMock(t)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	mockAccount.On("GetIdentity").Return(accountID)

	ctx := contextutil.WithAccount(context.Background(), mockAccount)

	meta, err := testingutils.GetTestMetadata()
	assert.NoError(t, err)

	centAPIMock.On("GetMetadataLatest").
		Return(meta, nil)

	podOperatorMock := config.NewPodOperatorMock(t)

	cfgServiceMock.On("GetPodOperator").
		Return(podOperatorMock, nil)

	var krp signature.KeyringPair

	podOperatorMock.On("ToKeyringPair").
		Return(krp)

	_, id, err := crypto.GenerateHashPair(32)
	assert.NoError(t, err)
	anchorID, err := ToAnchorID(id)
	assert.NoError(t, err)

	signingRoot := utils.RandomByte32()
	proof := utils.RandomByte32()
	b2bHash, err := blake2b.New256(nil)
	assert.NoError(t, err)
	_, err = b2bHash.Write(append(signingRoot[:], proof[:]...))
	assert.NoError(t, err)
	docRoot, err := ToDocumentRoot(b2bHash.Sum(nil))
	assert.NoError(t, err)

	call, err := types.NewCall(
		meta,
		commit,
		types.NewHash(anchorID[:]),
		types.NewHash(docRoot[:]),
		types.NewHash(proof[:]),
		types.NewMoment(time.Now().UTC().Add(anchorLifespan)),
	)
	assert.NoError(t, err)

	proxyAPIMock.On("ProxyCall", ctx, accountID, krp, types.NewOption(types.PodOperation), call).
		Return(nil, nil)

	err = service.CommitAnchor(ctx, anchorID, docRoot, proof)
	assert.NoError(t, err)
}

func TestService_CommitAnchor_AccountContextError(t *testing.T) {
	centAPIMock := centMocks.NewAPIMock(t)
	proxyAPIMock := proxyMocks.NewProxyAPIMock(t)
	cfgServiceMock := config.NewServiceMock(t)

	anchorLifespan := 1 * time.Minute

	service := newService(anchorLifespan, cfgServiceMock, centAPIMock, proxyAPIMock)

	_, id, err := crypto.GenerateHashPair(32)
	assert.NoError(t, err)
	anchorID, err := ToAnchorID(id)
	assert.NoError(t, err)

	signingRoot := utils.RandomByte32()
	proof := utils.RandomByte32()
	b2bHash, err := blake2b.New256(nil)
	assert.NoError(t, err)
	_, err = b2bHash.Write(append(signingRoot[:], proof[:]...))
	assert.NoError(t, err)
	docRoot, err := ToDocumentRoot(b2bHash.Sum(nil))
	assert.NoError(t, err)

	err = service.CommitAnchor(context.Background(), anchorID, docRoot, proof)
	assert.ErrorIs(t, err, errors.ErrContextAccountRetrieval)
}

func TestService_CommitAnchor_MetadataError(t *testing.T) {
	centAPIMock := centMocks.NewAPIMock(t)
	proxyAPIMock := proxyMocks.NewProxyAPIMock(t)
	cfgServiceMock := config.NewServiceMock(t)

	anchorLifespan := 1 * time.Minute

	service := newService(anchorLifespan, cfgServiceMock, centAPIMock, proxyAPIMock)

	mockAccount := configMocks.NewAccountMock(t)

	ctx := contextutil.WithAccount(context.Background(), mockAccount)

	centAPIMock.On("GetMetadataLatest").
		Return(nil, errors.New("metadata error"))

	_, id, err := crypto.GenerateHashPair(32)
	assert.NoError(t, err)
	anchorID, err := ToAnchorID(id)
	assert.NoError(t, err)

	signingRoot := utils.RandomByte32()
	proof := utils.RandomByte32()

	b2bHash, err := blake2b.New256(nil)
	assert.NoError(t, err)

	_, err = b2bHash.Write(append(signingRoot[:], proof[:]...))
	assert.NoError(t, err)

	docRoot, err := ToDocumentRoot(b2bHash.Sum(nil))
	assert.NoError(t, err)

	err = service.CommitAnchor(ctx, anchorID, docRoot, proof)
	assert.ErrorIs(t, err, errors.ErrMetadataRetrieval)
}

func TestService_CommitAnchor_CallCreationError(t *testing.T) {
	centAPIMock := centMocks.NewAPIMock(t)
	proxyAPIMock := proxyMocks.NewProxyAPIMock(t)
	cfgServiceMock := config.NewServiceMock(t)

	anchorLifespan := 1 * time.Minute

	service := newService(anchorLifespan, cfgServiceMock, centAPIMock, proxyAPIMock)

	mockAccount := configMocks.NewAccountMock(t)

	ctx := contextutil.WithAccount(context.Background(), mockAccount)

	var meta types.Metadata

	// NOTE - types.MetadataV14Data does not have info on the Anchor pallet,
	// causing types.NewCall to fail.
	err := types.DecodeFromHex(types.MetadataV14Data, &meta)
	assert.NoError(t, err)

	centAPIMock.On("GetMetadataLatest").
		Return(&meta, nil)

	_, id, err := crypto.GenerateHashPair(32)
	assert.NoError(t, err)
	anchorID, err := ToAnchorID(id)
	assert.NoError(t, err)

	signingRoot := utils.RandomByte32()
	proof := utils.RandomByte32()
	b2bHash, err := blake2b.New256(nil)
	assert.NoError(t, err)
	_, err = b2bHash.Write(append(signingRoot[:], proof[:]...))
	assert.NoError(t, err)
	docRoot, err := ToDocumentRoot(b2bHash.Sum(nil))
	assert.NoError(t, err)

	err = service.CommitAnchor(ctx, anchorID, docRoot, proof)
	assert.ErrorIs(t, err, errors.ErrCallCreation)
}

func TestService_CommitAnchor_PodOperatorError(t *testing.T) {
	centAPIMock := centMocks.NewAPIMock(t)
	proxyAPIMock := proxyMocks.NewProxyAPIMock(t)
	cfgServiceMock := config.NewServiceMock(t)

	anchorLifespan := 1 * time.Minute

	service := newService(anchorLifespan, cfgServiceMock, centAPIMock, proxyAPIMock)

	mockAccount := configMocks.NewAccountMock(t)

	ctx := contextutil.WithAccount(context.Background(), mockAccount)

	meta, err := testingutils.GetTestMetadata()
	assert.NoError(t, err)

	centAPIMock.On("GetMetadataLatest").
		Return(meta, nil)

	cfgServiceMock.On("GetPodOperator").
		Return(nil, errors.New("error"))

	_, id, err := crypto.GenerateHashPair(32)
	assert.NoError(t, err)
	anchorID, err := ToAnchorID(id)
	assert.NoError(t, err)

	signingRoot := utils.RandomByte32()
	proof := utils.RandomByte32()

	b2bHash, err := blake2b.New256(nil)
	assert.NoError(t, err)

	_, err = b2bHash.Write(append(signingRoot[:], proof[:]...))
	assert.NoError(t, err)

	docRoot, err := ToDocumentRoot(b2bHash.Sum(nil))
	assert.NoError(t, err)

	err = service.CommitAnchor(ctx, anchorID, docRoot, proof)
	assert.ErrorIs(t, err, errors.ErrPodOperatorRetrieval)
}

func TestService_CommitAnchor_ProxyCallError(t *testing.T) {
	centAPIMock := centMocks.NewAPIMock(t)
	proxyAPIMock := proxyMocks.NewProxyAPIMock(t)
	cfgServiceMock := config.NewServiceMock(t)

	anchorLifespan := 1 * time.Minute

	service := newService(anchorLifespan, cfgServiceMock, centAPIMock, proxyAPIMock)

	mockAccount := configMocks.NewAccountMock(t)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	mockAccount.On("GetIdentity").Return(accountID)

	ctx := contextutil.WithAccount(context.Background(), mockAccount)

	meta, err := testingutils.GetTestMetadata()
	assert.NoError(t, err)

	centAPIMock.On("GetMetadataLatest").
		Return(meta, nil)

	podOperatorMock := config.NewPodOperatorMock(t)

	cfgServiceMock.On("GetPodOperator").
		Return(podOperatorMock, nil)

	var krp signature.KeyringPair

	podOperatorMock.On("ToKeyringPair").
		Return(krp)

	_, id, err := crypto.GenerateHashPair(32)
	assert.NoError(t, err)
	anchorID, err := ToAnchorID(id)
	assert.NoError(t, err)

	signingRoot := utils.RandomByte32()
	proof := utils.RandomByte32()
	b2bHash, err := blake2b.New256(nil)
	assert.NoError(t, err)
	_, err = b2bHash.Write(append(signingRoot[:], proof[:]...))
	assert.NoError(t, err)
	docRoot, err := ToDocumentRoot(b2bHash.Sum(nil))
	assert.NoError(t, err)

	call, err := types.NewCall(
		meta,
		commit,
		types.NewHash(anchorID[:]),
		types.NewHash(docRoot[:]),
		types.NewHash(proof[:]),
		types.NewMoment(time.Now().UTC().Add(anchorLifespan)),
	)
	assert.NoError(t, err)

	proxyAPIMock.On("ProxyCall", ctx, accountID, krp, types.NewOption(types.PodOperation), call).
		Return(nil, errors.New("proxy call error"))

	err = service.CommitAnchor(ctx, anchorID, docRoot, proof)
	assert.ErrorIs(t, err, errors.ErrProxyCall)
}
