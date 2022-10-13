//go:build unit

package v2

import (
	"context"
	"math/big"
	"testing"
	"time"

	keystoreType "github.com/centrifuge/chain-custom-types/pkg/keystore"
	proxyType "github.com/centrifuge/chain-custom-types/pkg/proxy"
	"github.com/centrifuge/go-centrifuge/centchain"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/crypto"
	"github.com/centrifuge/go-centrifuge/crypto/ed25519"
	protocolIDDispatcher "github.com/centrifuge/go-centrifuge/dispatcher"
	"github.com/centrifuge/go-centrifuge/errors"
	p2pcommon "github.com/centrifuge/go-centrifuge/p2p/common"
	"github.com/centrifuge/go-centrifuge/pallets/keystore"
	"github.com/centrifuge/go-centrifuge/pallets/proxy"
	"github.com/centrifuge/go-centrifuge/testingutils"
	testingcommons "github.com/centrifuge/go-centrifuge/testingutils/common"
	genericUtils "github.com/centrifuge/go-centrifuge/testingutils/generic"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types/codec"
	"github.com/libp2p/go-libp2p-core/protocol"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestService_CreateIdentity(t *testing.T) {
	service, mocks := getServiceWithMocks(t)

	ctx := context.Background()

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	req := &CreateIdentityRequest{
		Identity:         accountID,
		PrecommitEnabled: true,
	}

	meta, err := testingutils.GetTestMetadata()
	assert.NoError(t, err)

	genericUtils.GetMock[*centchain.APIMock](mocks).On("GetMetadataLatest").
		Return(meta, nil).
		Once()

	storageKey, err := types.CreateStorageKey(meta, "System", "Account", accountID.ToBytes())
	assert.NoError(t, err)

	genericUtils.GetMock[*centchain.APIMock](mocks).On("GetStorageLatest", storageKey, mock.Anything).
		Run(
			func(args mock.Arguments) {
				accountInfo, ok := args.Get(1).(*types.AccountInfo)
				assert.True(t, ok)

				accountInfo.Nonce = 1
				accountInfo.Consumers = 2
				accountInfo.Providers = 3
				accountInfo.Data.Free = types.NewU128(*big.NewInt(1111))
				accountInfo.Data.FreeFrozen = types.NewU128(*big.NewInt(2222))
				accountInfo.Data.Reserved = types.NewU128(*big.NewInt(3333))
				accountInfo.Data.MiscFrozen = types.NewU128(*big.NewInt(4444))
			},
		).
		Return(true, nil).Once()

	genericUtils.GetMock[*config.ServiceMock](mocks).On("CreateAccount", mock.Anything).
		Run(
			func(args mock.Arguments) {
				account, ok := args.Get(0).(config.Account)
				assert.True(t, ok)

				assert.Equal(t, req.Identity, account.GetIdentity())
				assert.Equal(t, req.WebhookURL, account.GetWebhookURL())
				assert.Equal(t, req.PrecommitEnabled, account.GetPrecommitEnabled())
				assert.NotNil(t, account.GetSigningPublicKey())
			},
		).Return(nil).Once()

	protocolID := p2pcommon.ProtocolForIdentity(accountID)

	genericUtils.GetMock[*protocolIDDispatcher.DispatcherMock[protocol.ID]](mocks).On("Dispatch", ctx, protocolID).
		Return(nil).Once()

	acc, err := service.CreateIdentity(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, acc)
}

func TestService_CreateIdentity_InvalidRequest(t *testing.T) {
	service, mocks := getServiceWithMocks(t)

	ctx := context.Background()

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	webhookURL := "invalidURL"

	req := &CreateIdentityRequest{
		Identity:         accountID,
		WebhookURL:       webhookURL,
		PrecommitEnabled: true,
	}

	meta, err := testingutils.GetTestMetadata()
	assert.NoError(t, err)

	genericUtils.GetMock[*centchain.APIMock](mocks).On("GetMetadataLatest").
		Return(meta, nil).
		Once()

	storageKey, err := types.CreateStorageKey(meta, "System", "Account", accountID.ToBytes())
	assert.NoError(t, err)

	genericUtils.GetMock[*centchain.APIMock](mocks).On("GetStorageLatest", storageKey, mock.Anything).
		Run(
			func(args mock.Arguments) {
				accountInfo, ok := args.Get(1).(*types.AccountInfo)
				assert.True(t, ok)

				accountInfo.Nonce = 1
				accountInfo.Consumers = 2
				accountInfo.Providers = 3
				accountInfo.Data.Free = types.NewU128(*big.NewInt(1111))
				accountInfo.Data.FreeFrozen = types.NewU128(*big.NewInt(2222))
				accountInfo.Data.Reserved = types.NewU128(*big.NewInt(3333))
				accountInfo.Data.MiscFrozen = types.NewU128(*big.NewInt(4444))
			},
		).
		Return(true, nil).Once()

	acc, err := service.CreateIdentity(ctx, req)
	assert.ErrorIs(t, err, ErrInvalidWebhookURL)
	assert.Nil(t, acc)
}

func TestService_CreateIdentity_AccountStorageError(t *testing.T) {
	service, mocks := getServiceWithMocks(t)

	ctx := context.Background()

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	webhookURL := "https://centrifuge.io"

	req := &CreateIdentityRequest{
		Identity:         accountID,
		WebhookURL:       webhookURL,
		PrecommitEnabled: true,
	}

	meta, err := testingutils.GetTestMetadata()
	assert.NoError(t, err)

	genericUtils.GetMock[*centchain.APIMock](mocks).On("GetMetadataLatest").
		Return(meta, nil).
		Once()

	storageKey, err := types.CreateStorageKey(meta, "System", "Account", accountID.ToBytes())
	assert.NoError(t, err)

	genericUtils.GetMock[*centchain.APIMock](mocks).On("GetStorageLatest", storageKey, mock.Anything).
		Run(
			func(args mock.Arguments) {
				accountInfo, ok := args.Get(1).(*types.AccountInfo)
				assert.True(t, ok)

				accountInfo.Nonce = 1
				accountInfo.Consumers = 2
				accountInfo.Providers = 3
				accountInfo.Data.Free = types.NewU128(*big.NewInt(1111))
				accountInfo.Data.FreeFrozen = types.NewU128(*big.NewInt(2222))
				accountInfo.Data.Reserved = types.NewU128(*big.NewInt(3333))
				accountInfo.Data.MiscFrozen = types.NewU128(*big.NewInt(4444))
			},
		).
		Return(true, nil).Once()

	genericUtils.GetMock[*config.ServiceMock](mocks).On("CreateAccount", mock.Anything).
		Return(errors.New("error")).Once()

	acc, err := service.CreateIdentity(ctx, req)
	assert.ErrorIs(t, err, ErrAccountStorage)
	assert.Nil(t, acc)
}

func TestService_CreateIdentity_ProtocolIDDispatcherError(t *testing.T) {
	service, mocks := getServiceWithMocks(t)

	ctx := context.Background()

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	webhookURL := "https://centrifuge.io"

	req := &CreateIdentityRequest{
		Identity:         accountID,
		WebhookURL:       webhookURL,
		PrecommitEnabled: true,
	}

	meta, err := testingutils.GetTestMetadata()
	assert.NoError(t, err)

	genericUtils.GetMock[*centchain.APIMock](mocks).On("GetMetadataLatest").
		Return(meta, nil).
		Once()

	storageKey, err := types.CreateStorageKey(meta, "System", "Account", accountID.ToBytes())
	assert.NoError(t, err)

	genericUtils.GetMock[*centchain.APIMock](mocks).On("GetStorageLatest", storageKey, mock.Anything).
		Run(
			func(args mock.Arguments) {
				accountInfo, ok := args.Get(1).(*types.AccountInfo)
				assert.True(t, ok)

				accountInfo.Nonce = 1
				accountInfo.Consumers = 2
				accountInfo.Providers = 3
				accountInfo.Data.Free = types.NewU128(*big.NewInt(1111))
				accountInfo.Data.FreeFrozen = types.NewU128(*big.NewInt(2222))
				accountInfo.Data.Reserved = types.NewU128(*big.NewInt(3333))
				accountInfo.Data.MiscFrozen = types.NewU128(*big.NewInt(4444))
			},
		).
		Return(true, nil).Once()

	genericUtils.GetMock[*config.ServiceMock](mocks).On("CreateAccount", mock.Anything).
		Run(
			func(args mock.Arguments) {
				account, ok := args.Get(0).(config.Account)
				assert.True(t, ok)

				assert.Equal(t, req.Identity, account.GetIdentity())
				assert.Equal(t, req.WebhookURL, account.GetWebhookURL())
				assert.Equal(t, req.PrecommitEnabled, account.GetPrecommitEnabled())
				assert.NotNil(t, account.GetSigningPublicKey())
			},
		).Return(nil).Once()

	protocolID := p2pcommon.ProtocolForIdentity(accountID)

	genericUtils.GetMock[*protocolIDDispatcher.DispatcherMock[protocol.ID]](mocks).On("Dispatch", ctx, protocolID).
		Return(errors.New("error")).Once()

	acc, err := service.CreateIdentity(ctx, req)
	assert.ErrorIs(t, err, ErrProtocolIDDispatch)
	assert.Nil(t, acc)
}

func TestService_ValidateKey(t *testing.T) {
	service, mocks := getServiceWithMocks(t)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)
	publicKey := utils.RandomSlice(32)
	keyPurpose := keystoreType.KeyPurposeP2PDocumentSigning

	keyID := &keystoreType.KeyID{
		Hash:       types.NewHash(publicKey),
		KeyPurpose: keyPurpose,
	}

	keystoreKey := &keystoreType.Key{}

	genericUtils.GetMock[*keystore.APIMock](mocks).On("GetKey", accountID, keyID).
		Return(keystoreKey, nil).
		Once()

	err = service.ValidateKey(accountID, publicKey, keyPurpose, time.Now())
	assert.NoError(t, err)
}

func TestService_ValidateKey_ValidationErrors(t *testing.T) {
	service, _ := getServiceWithMocks(t)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)
	publicKey := utils.RandomSlice(32)
	keyPurpose := keystoreType.KeyPurposeP2PDocumentSigning

	// Nil account ID.
	err = service.ValidateKey(nil, publicKey, keyPurpose, time.Now())
	assert.NotNil(t, err)

	// Invalid public key length.
	err = service.ValidateKey(accountID, utils.RandomSlice(31), keyPurpose, time.Now())
	assert.NotNil(t, err)
}

func TestService_ValidateKey_KeyRetrievalError(t *testing.T) {
	service, mocks := getServiceWithMocks(t)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)
	publicKey := utils.RandomSlice(32)
	keyPurpose := keystoreType.KeyPurposeP2PDocumentSigning

	keyID := &keystoreType.KeyID{
		Hash:       types.NewHash(publicKey),
		KeyPurpose: keyPurpose,
	}

	genericUtils.GetMock[*keystore.APIMock](mocks).On("GetKey", accountID, keyID).
		Return(nil, errors.New("error")).
		Once()

	err = service.ValidateKey(accountID, publicKey, keyPurpose, time.Now())
	assert.ErrorIs(t, err, ErrKeyRetrieval)
}

func TestService_ValidateKey_ValidKey(t *testing.T) {
	service, mocks := getServiceWithMocks(t)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)
	publicKey := utils.RandomSlice(32)
	keyPurpose := keystoreType.KeyPurposeP2PDocumentSigning

	keyID := &keystoreType.KeyID{
		Hash:       types.NewHash(publicKey),
		KeyPurpose: keyPurpose,
	}

	blockNumber := types.U32(11)

	keystoreKey := &keystoreType.Key{
		RevokedAt: types.NewOption[types.U32](blockNumber),
	}

	genericUtils.GetMock[*keystore.APIMock](mocks).On("GetKey", accountID, keyID).
		Return(keystoreKey, nil).
		Once()

	blockHash := types.NewHash(utils.RandomSlice(32))

	genericUtils.GetMock[*centchain.APIMock](mocks).On("GetBlockHash", uint64(blockNumber)).
		Return(blockHash, nil).
		Once()

	validationTime := time.Now()

	blockTimestamp := types.NewUCompactFromUInt(uint64(validationTime.Add(1 * time.Hour).UnixMilli()))

	encodedTimestamp, err := codec.Encode(blockTimestamp)
	assert.NoError(t, err)

	block := &types.SignedBlock{
		Block: types.Block{
			Header: types.Header{
				Number: types.BlockNumber(blockNumber - 1),
			},
			Extrinsics: []types.Extrinsic{
				{
					Method: types.Call{
						CallIndex: setTimestampCallIndex,
						Args:      encodedTimestamp,
					},
				},
			},
		},
	}

	genericUtils.GetMock[*centchain.APIMock](mocks).On("GetBlock", blockHash).
		Return(block, nil).
		Once()

	meta, err := testingutils.GetTestMetadata()
	assert.NoError(t, err)

	genericUtils.GetMock[*centchain.APIMock](mocks).On("GetMetadataLatest").
		Return(meta, nil).
		Once()

	err = service.ValidateKey(accountID, publicKey, keyPurpose, validationTime)
	assert.NoError(t, err)
}

func TestService_ValidateKey_InvalidKey(t *testing.T) {
	service, mocks := getServiceWithMocks(t)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)
	publicKey := utils.RandomSlice(32)
	keyPurpose := keystoreType.KeyPurposeP2PDocumentSigning

	keyID := &keystoreType.KeyID{
		Hash:       types.NewHash(publicKey),
		KeyPurpose: keyPurpose,
	}

	blockNumber := types.U32(11)

	keystoreKey := &keystoreType.Key{
		RevokedAt: types.NewOption[types.U32](blockNumber),
	}

	genericUtils.GetMock[*keystore.APIMock](mocks).On("GetKey", accountID, keyID).
		Return(keystoreKey, nil).
		Once()

	blockHash := types.NewHash(utils.RandomSlice(32))

	genericUtils.GetMock[*centchain.APIMock](mocks).On("GetBlockHash", uint64(blockNumber)).
		Return(blockHash, nil).
		Once()

	validationTime := time.Now()

	blockTimestamp := types.NewUCompactFromUInt(uint64(validationTime.Add(-1 * time.Hour).Unix()))

	encodedTimestamp, err := codec.Encode(blockTimestamp)
	assert.NoError(t, err)

	block := &types.SignedBlock{
		Block: types.Block{
			Header: types.Header{
				Number: types.BlockNumber(blockNumber - 1),
			},
			Extrinsics: []types.Extrinsic{
				{
					Method: types.Call{
						CallIndex: setTimestampCallIndex,
						Args:      encodedTimestamp,
					},
				},
			},
		},
	}

	genericUtils.GetMock[*centchain.APIMock](mocks).On("GetBlock", blockHash).
		Return(block, nil).
		Once()

	meta, err := testingutils.GetTestMetadata()
	assert.NoError(t, err)

	genericUtils.GetMock[*centchain.APIMock](mocks).On("GetMetadataLatest").
		Return(meta, nil).
		Once()

	err = service.ValidateKey(accountID, publicKey, keyPurpose, validationTime)
	assert.ErrorIs(t, err, ErrKeyRevoked)
}

func TestService_ValidateKey_BlockHashRetrievalError(t *testing.T) {
	service, mocks := getServiceWithMocks(t)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)
	publicKey := utils.RandomSlice(32)
	keyPurpose := keystoreType.KeyPurposeP2PDocumentSigning

	keyID := &keystoreType.KeyID{
		Hash:       types.NewHash(publicKey),
		KeyPurpose: keyPurpose,
	}

	blockNumber := types.U32(11)

	keystoreKey := &keystoreType.Key{
		RevokedAt: types.NewOption[types.U32](blockNumber),
	}

	genericUtils.GetMock[*keystore.APIMock](mocks).On("GetKey", accountID, keyID).
		Return(keystoreKey, nil).
		Once()

	genericUtils.GetMock[*centchain.APIMock](mocks).On("GetBlockHash", uint64(blockNumber)).
		Return(nil, errors.New("error")).
		Once()

	err = service.ValidateKey(accountID, publicKey, keyPurpose, time.Now())
	assert.ErrorIs(t, err, ErrBlockHashRetrieval)
}

func TestService_ValidateKey_BlockRetrievalError(t *testing.T) {
	service, mocks := getServiceWithMocks(t)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)
	publicKey := utils.RandomSlice(32)
	keyPurpose := keystoreType.KeyPurposeP2PDocumentSigning

	keyID := &keystoreType.KeyID{
		Hash:       types.NewHash(publicKey),
		KeyPurpose: keyPurpose,
	}

	blockNumber := types.U32(11)

	keystoreKey := &keystoreType.Key{
		RevokedAt: types.NewOption[types.U32](blockNumber),
	}

	genericUtils.GetMock[*keystore.APIMock](mocks).On("GetKey", accountID, keyID).
		Return(keystoreKey, nil).
		Once()

	blockHash := types.NewHash(utils.RandomSlice(32))

	genericUtils.GetMock[*centchain.APIMock](mocks).On("GetBlockHash", uint64(blockNumber)).
		Return(blockHash, nil).
		Once()

	genericUtils.GetMock[*centchain.APIMock](mocks).On("GetBlock", blockHash).
		Return(nil, errors.New("error")).
		Once()

	err = service.ValidateKey(accountID, publicKey, keyPurpose, time.Now())
	assert.ErrorIs(t, err, ErrBlockRetrieval)
}

func TestService_ValidateKey_MetadataRetrievalError(t *testing.T) {
	service, mocks := getServiceWithMocks(t)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)
	publicKey := utils.RandomSlice(32)
	keyPurpose := keystoreType.KeyPurposeP2PDocumentSigning

	keyID := &keystoreType.KeyID{
		Hash:       types.NewHash(publicKey),
		KeyPurpose: keyPurpose,
	}

	blockNumber := types.U32(11)

	keystoreKey := &keystoreType.Key{
		RevokedAt: types.NewOption[types.U32](blockNumber),
	}

	genericUtils.GetMock[*keystore.APIMock](mocks).On("GetKey", accountID, keyID).
		Return(keystoreKey, nil).
		Once()

	blockHash := types.NewHash(utils.RandomSlice(32))

	genericUtils.GetMock[*centchain.APIMock](mocks).On("GetBlockHash", uint64(blockNumber)).
		Return(blockHash, nil).
		Once()

	validationTime := time.Now()

	blockTimestamp := types.NewUCompactFromUInt(uint64(validationTime.Add(1 * time.Hour).Unix()))

	encodedTimestamp, err := codec.Encode(blockTimestamp)
	assert.NoError(t, err)

	block := &types.SignedBlock{
		Block: types.Block{
			Header: types.Header{
				Number: types.BlockNumber(blockNumber - 1),
			},
			Extrinsics: []types.Extrinsic{
				{
					Method: types.Call{
						CallIndex: setTimestampCallIndex,
						Args:      encodedTimestamp,
					},
				},
			},
		},
	}

	genericUtils.GetMock[*centchain.APIMock](mocks).On("GetBlock", blockHash).
		Return(block, nil).
		Once()

	genericUtils.GetMock[*centchain.APIMock](mocks).On("GetMetadataLatest").
		Return(nil, errors.New("error")).
		Once()

	err = service.ValidateKey(accountID, publicKey, keyPurpose, validationTime)
	assert.ErrorIs(t, err, errors.ErrMetadataRetrieval)
}

func TestService_ValidateKey_ExtrinsicTimestampRetrievalError(t *testing.T) {
	service, mocks := getServiceWithMocks(t)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)
	publicKey := utils.RandomSlice(32)
	keyPurpose := keystoreType.KeyPurposeP2PDocumentSigning

	keyID := &keystoreType.KeyID{
		Hash:       types.NewHash(publicKey),
		KeyPurpose: keyPurpose,
	}

	blockNumber := types.U32(11)

	keystoreKey := &keystoreType.Key{
		RevokedAt: types.NewOption[types.U32](blockNumber),
	}

	genericUtils.GetMock[*keystore.APIMock](mocks).On("GetKey", accountID, keyID).
		Return(keystoreKey, nil).
		Once()

	blockHash := types.NewHash(utils.RandomSlice(32))

	genericUtils.GetMock[*centchain.APIMock](mocks).On("GetBlockHash", uint64(blockNumber)).
		Return(blockHash, nil).
		Once()

	validationTime := time.Now()

	blockTimestamp := types.NewUCompactFromUInt(uint64(validationTime.Add(1 * time.Hour).Unix()))

	encodedTimestamp, err := codec.Encode(blockTimestamp)
	assert.NoError(t, err)

	block := &types.SignedBlock{
		Block: types.Block{
			Header: types.Header{
				Number: types.BlockNumber(blockNumber - 1),
			},
			Extrinsics: []types.Extrinsic{
				{
					Method: types.Call{
						// Invalid call index for "set.timestamp"
						CallIndex: types.CallIndex{
							SectionIndex: 11,
							MethodIndex:  22,
						},
						Args: encodedTimestamp,
					},
				},
			},
		},
	}

	genericUtils.GetMock[*centchain.APIMock](mocks).On("GetBlock", blockHash).
		Return(block, nil).
		Once()

	meta, err := testingutils.GetTestMetadata()
	assert.NoError(t, err)

	genericUtils.GetMock[*centchain.APIMock](mocks).On("GetMetadataLatest").
		Return(meta, nil).
		Once()

	err = service.ValidateKey(accountID, publicKey, keyPurpose, validationTime)
	assert.ErrorIs(t, err, ErrBlockTimestampRetrieval)
}

func TestService_ValidateSignature(t *testing.T) {
	service, mocks := getServiceWithMocks(t)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	message := utils.RandomSlice(32)

	pubKey, privateKey, err := ed25519.GenerateSigningKeyPair()
	assert.NoError(t, err)

	signature, err := crypto.SignMessage(privateKey, message, crypto.CurveEd25519)
	assert.NoError(t, err)

	keyPurpose := keystoreType.KeyPurposeP2PDocumentSigning

	keyID := &keystoreType.KeyID{
		Hash:       types.NewHash(pubKey),
		KeyPurpose: keyPurpose,
	}

	keystoreKey := &keystoreType.Key{}

	genericUtils.GetMock[*keystore.APIMock](mocks).On("GetKey", accountID, keyID).
		Return(keystoreKey, nil).
		Once()

	err = service.ValidateSignature(accountID, pubKey, message, signature, time.Now())
	assert.NoError(t, err)
}

func TestService_ValidateSignature_KeyValidationError(t *testing.T) {
	service, mocks := getServiceWithMocks(t)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	message := utils.RandomSlice(32)

	pubKey, privateKey, err := ed25519.GenerateSigningKeyPair()
	assert.NoError(t, err)

	signature, err := crypto.SignMessage(privateKey, message, crypto.CurveEd25519)
	assert.NoError(t, err)

	keyPurpose := keystoreType.KeyPurposeP2PDocumentSigning

	keyID := &keystoreType.KeyID{
		Hash:       types.NewHash(pubKey),
		KeyPurpose: keyPurpose,
	}

	genericUtils.GetMock[*keystore.APIMock](mocks).On("GetKey", accountID, keyID).
		Return(nil, errors.New("error")).
		Once()

	err = service.ValidateSignature(accountID, pubKey, message, signature, time.Now())
	assert.ErrorIs(t, err, ErrKeyRetrieval)
}

func TestService_ValidateSignature_InvalidSignature(t *testing.T) {
	service, mocks := getServiceWithMocks(t)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	message := utils.RandomSlice(32)

	pubKey, _, err := ed25519.GenerateSigningKeyPair()
	assert.NoError(t, err)

	// Invalid signature
	signature := utils.RandomSlice(32)

	keyPurpose := keystoreType.KeyPurposeP2PDocumentSigning

	keyID := &keystoreType.KeyID{
		Hash:       types.NewHash(pubKey),
		KeyPurpose: keyPurpose,
	}

	keystoreKey := &keystoreType.Key{}

	genericUtils.GetMock[*keystore.APIMock](mocks).On("GetKey", accountID, keyID).
		Return(keystoreKey, nil).
		Once()

	err = service.ValidateSignature(accountID, pubKey, message, signature, time.Now())
	assert.ErrorIs(t, err, ErrInvalidSignature)
}

func TestService_ValidateAccount(t *testing.T) {
	service, mocks := getServiceWithMocks(t)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	meta, err := testingutils.GetTestMetadata()
	assert.NoError(t, err)

	genericUtils.GetMock[*centchain.APIMock](mocks).On("GetMetadataLatest").
		Return(meta, nil).
		Once()

	storageKey, err := types.CreateStorageKey(meta, "System", "Account", accountID.ToBytes())
	assert.NoError(t, err)

	genericUtils.GetMock[*centchain.APIMock](mocks).On("GetStorageLatest", storageKey, mock.IsType(&types.AccountInfo{})).
		Return(true, nil).
		Once()

	err = service.ValidateAccount(accountID)
	assert.NoError(t, err)
}

func TestService_ValidateAccount_ValidationError(t *testing.T) {
	service, _ := getServiceWithMocks(t)

	err := service.ValidateAccount(nil)
	assert.ErrorIs(t, err, ErrInvalidAccountID)
}

func TestService_ValidateAccount_MetadataRetrievalError(t *testing.T) {
	service, mocks := getServiceWithMocks(t)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	genericUtils.GetMock[*centchain.APIMock](mocks).On("GetMetadataLatest").
		Return(nil, errors.New("error")).
		Once()

	err = service.ValidateAccount(accountID)
	assert.ErrorIs(t, err, ErrMetadataRetrieval)
}

func TestService_ValidateAccount_StorageKeyCreationError(t *testing.T) {
	service, mocks := getServiceWithMocks(t)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	meta, err := testingutils.GetTestMetadata()
	assert.NoError(t, err)

	// Replace metadata pallet info to ensure an error when creating the storage key.
	meta.AsMetadataV14.Pallets = nil

	genericUtils.GetMock[*centchain.APIMock](mocks).On("GetMetadataLatest").
		Return(meta, nil).
		Once()

	err = service.ValidateAccount(accountID)
	assert.ErrorIs(t, err, ErrAccountStorageKeyCreation)
}

func TestService_ValidateAccount_AccountStorageRetrievalError(t *testing.T) {
	service, mocks := getServiceWithMocks(t)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	meta, err := testingutils.GetTestMetadata()
	assert.NoError(t, err)

	genericUtils.GetMock[*centchain.APIMock](mocks).On("GetMetadataLatest").
		Return(meta, nil).
		Once()

	storageKey, err := types.CreateStorageKey(meta, "System", "Account", accountID.ToBytes())
	assert.NoError(t, err)

	genericUtils.GetMock[*centchain.APIMock](mocks).On("GetStorageLatest", storageKey, mock.IsType(&types.AccountInfo{})).
		Return(false, errors.New("error")).
		Once()

	err = service.ValidateAccount(accountID)
	assert.ErrorIs(t, err, ErrAccountStorageRetrieval)
}

func TestService_ValidateAccount_AccountNotFound_ProxyExists(t *testing.T) {
	service, mocks := getServiceWithMocks(t)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	meta, err := testingutils.GetTestMetadata()
	assert.NoError(t, err)

	genericUtils.GetMock[*centchain.APIMock](mocks).On("GetMetadataLatest").
		Return(meta, nil).
		Once()

	storageKey, err := types.CreateStorageKey(meta, "System", "Account", accountID.ToBytes())
	assert.NoError(t, err)

	genericUtils.GetMock[*centchain.APIMock](mocks).On("GetStorageLatest", storageKey, mock.IsType(&types.AccountInfo{})).
		Return(false, nil).
		Once()

	proxyRes := &types.ProxyStorageEntry{
		ProxyDefinitions: []types.ProxyDefinition{
			{
				ProxyType: types.U8(proxyType.Any),
			},
		},
	}

	genericUtils.GetMock[*proxy.APIMock](mocks).On("GetProxies", accountID).
		Return(proxyRes, nil).
		Once()

	err = service.ValidateAccount(accountID)
	assert.NoError(t, err)
}

func TestService_ValidateAccount_AccountNotFound_ProxyRetrievalError(t *testing.T) {
	service, mocks := getServiceWithMocks(t)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	meta, err := testingutils.GetTestMetadata()
	assert.NoError(t, err)

	genericUtils.GetMock[*centchain.APIMock](mocks).On("GetMetadataLatest").
		Return(meta, nil)

	storageKey, err := types.CreateStorageKey(meta, "System", "Account", accountID.ToBytes())
	assert.NoError(t, err)

	genericUtils.GetMock[*centchain.APIMock](mocks).On("GetStorageLatest", storageKey, mock.IsType(&types.AccountInfo{})).
		Return(false, nil)

	genericUtils.GetMock[*proxy.APIMock](mocks).On("GetProxies", accountID).
		Return(nil, errors.New("error")).
		Once()

	err = service.ValidateAccount(accountID)
	assert.ErrorIs(t, err, ErrAccountProxiesRetrieval)

	genericUtils.GetMock[*proxy.APIMock](mocks).On("GetProxies", accountID).
		Return(nil, proxy.ErrProxiesNotFound).
		Once()

	err = service.ValidateAccount(accountID)
	assert.ErrorIs(t, err, ErrInvalidAccount)
}

var setTimestampCallIndex = types.CallIndex{
	SectionIndex: 3,
	MethodIndex:  0,
}

func getServiceWithMocks(t *testing.T) (Service, []any) {
	configServiceMock := config.NewServiceMock(t)
	centAPIMock := centchain.NewAPIMock(t)
	keystoreAPIMock := keystore.NewAPIMock(t)
	proxyAPIMock := proxy.NewAPIMock(t)
	protocolIDDispatcherMock := protocolIDDispatcher.NewDispatcherMock[protocol.ID](t)

	service := NewService(
		configServiceMock,
		centAPIMock,
		keystoreAPIMock,
		proxyAPIMock,
		protocolIDDispatcherMock,
	)

	return service, []any{
		configServiceMock,
		centAPIMock,
		keystoreAPIMock,
		proxyAPIMock,
		protocolIDDispatcherMock,
	}
}
