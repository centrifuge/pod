package keystore

import (
	"context"

	"github.com/centrifuge/go-centrifuge/centchain"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	logging "github.com/ipfs/go-log"
)

const (
	ErrContextAccountRetrieval = errors.Error("couldn't retrieve account from context")
	ErrAccountProxyRetrieval   = errors.Error("couldn't retrieve account proxy")
	ErrKeyringPairRetrieval    = errors.Error("couldn't retrieve keyring pair")
	ErrMetadataRetrieval       = errors.Error("couldn't retrieve metadata")
	ErrCallCreation            = errors.Error("couldn't create call")
	ErrSubmitAndWatchExtrinsic = errors.Error("couldn't submit and watch extrinsic")
	ErrKeyIDEncoding           = errors.Error("couldn't encode key ID")
	ErrIdentityEncoding        = errors.Error("couldn't encode identity")
	ErrKeyPurposeEncoding      = errors.Error("couldn't encode key purpose")
	ErrStorageKeyCreation      = errors.Error("couldn't create storage key")
	ErrKeyStorageRetrieval     = errors.Error("couldn't retrieve key from storage")
)

const (
	PalletName = "Keystore"

	AddKeysCall    = PalletName + ".add_keys"
	RevokeKeysCall = PalletName + ".revoke_keys"

	KeysStorageName             = "Keys"
	LastKeyByPurposeStorageName = "LastKeyByPurpose"
)

type API interface {
	AddKeys(ctx context.Context, keys []*types.AddKey) (*centchain.ExtrinsicInfo, error)
	RevokeKeys(ctx context.Context, keys []*types.Hash, keyPurpose types.KeyPurpose) (*centchain.ExtrinsicInfo, error)
	GetKey(ctx context.Context, keyID *types.KeyID) (*types.Key, error)
	GetLastKeyByPurpose(ctx context.Context, keyPurpose types.KeyPurpose) (*types.Hash, error)
}

type api struct {
	api centchain.API
	log *logging.ZapEventLogger
}

func NewAPI(centAPI centchain.API) API {
	return &api{
		api: centAPI,
		log: logging.Logger("keystore_api"),
	}
}

func (a *api) AddKeys(ctx context.Context, keys []*types.AddKey) (*centchain.ExtrinsicInfo, error) {
	//TODO(cdamian): Add validation from the NFT branch

	acc, err := contextutil.Account(ctx)

	if err != nil {
		a.log.Errorf("Couldn't retrieve account from context: %s", err)

		return nil, ErrContextAccountRetrieval
	}

	accProxy, err := acc.GetAccountProxies().WithProxyType(types.KeystoreManagement)

	if err != nil {
		a.log.Errorf("Couldn't get account proxy: %s", err)

		return nil, ErrAccountProxyRetrieval
	}

	krp, err := accProxy.ToKeyringPair()

	if err != nil {
		a.log.Errorf("Couldn't retrieve key ring pair from account: %s", err)

		return nil, ErrKeyringPairRetrieval
	}

	meta, err := a.api.GetMetadataLatest()

	if err != nil {
		a.log.Errorf("Couldn't retrieve latest metadata: %s", err)

		return nil, ErrMetadataRetrieval
	}

	call, err := types.NewCall(meta, AddKeysCall, keys)

	if err != nil {
		a.log.Errorf("Couldn't create call: %s", err)

		return nil, ErrCallCreation
	}

	extInfo, err := a.api.SubmitAndWatch(ctx, meta, call, *krp)

	if err != nil {
		a.log.Errorf("Couldn't submit and watch extrinsic: %s", err)

		return nil, ErrSubmitAndWatchExtrinsic
	}

	return &extInfo, nil
}

func (a *api) RevokeKeys(
	ctx context.Context,
	keys []*types.Hash,
	keyPurpose types.KeyPurpose,
) (*centchain.ExtrinsicInfo, error) {
	//TODO(cdamian): Add validation from the NFT branch

	acc, err := contextutil.Account(ctx)

	if err != nil {
		a.log.Errorf("Couldn't retrieve account from context: %s", err)

		return nil, ErrContextAccountRetrieval
	}

	accProxy, err := acc.GetAccountProxies().WithProxyType(types.KeystoreManagement)

	if err != nil {
		a.log.Errorf("Couldn't get account proxy: %s", err)

		return nil, ErrAccountProxyRetrieval
	}

	krp, err := accProxy.ToKeyringPair()

	if err != nil {
		a.log.Errorf("Couldn't retrieve key ring pair from account: %s", err)

		return nil, ErrKeyringPairRetrieval
	}

	meta, err := a.api.GetMetadataLatest()

	if err != nil {
		a.log.Errorf("Couldn't retrieve latest metadata: %s", err)

		return nil, ErrMetadataRetrieval
	}

	call, err := types.NewCall(meta, RevokeKeysCall, keys, keyPurpose)

	if err != nil {
		a.log.Errorf("Couldn't create call: %s", err)

		return nil, ErrCallCreation
	}

	extInfo, err := a.api.SubmitAndWatch(ctx, meta, call, *krp)

	if err != nil {
		a.log.Errorf("Couldn't submit and watch extrinsic: %s", err)

		return nil, ErrSubmitAndWatchExtrinsic
	}

	return &extInfo, nil
}

func (a *api) GetKey(ctx context.Context, keyID *types.KeyID) (*types.Key, error) {
	//TODO(cdamian): Add validation from the NFT branch

	acc, err := contextutil.Account(ctx)

	if err != nil {
		a.log.Errorf("Couldn't retrieve account from context: %s", err)

		return nil, ErrContextAccountRetrieval
	}

	meta, err := a.api.GetMetadataLatest()

	if err != nil {
		a.log.Errorf("Couldn't retrieve latest metadata: %s", err)

		return nil, ErrMetadataRetrieval
	}

	encodedKeyID, err := types.Encode(keyID)

	if err != nil {
		a.log.Errorf("Couldn't encode key ID: %s", err)

		return nil, ErrKeyIDEncoding
	}

	encodedIdentity, err := types.Encode(acc.GetIdentity())

	if err != nil {
		a.log.Errorf("Couldn't encode identity: %s", err)

		return nil, ErrIdentityEncoding
	}

	storageKey, err := types.CreateStorageKey(meta, PalletName, KeysStorageName, encodedIdentity, encodedKeyID)

	if err != nil {
		a.log.Errorf("Couldn't create storage key: %s", err)

		return nil, ErrStorageKeyCreation
	}

	var key types.Key

	// TODO(cdamian): Use the OK from the NFT branch.
	if err := a.api.GetStorageLatest(storageKey, &key); err != nil {
		a.log.Errorf("Couldn't retrieve key from storage: %s", err)

		return nil, ErrKeyStorageRetrieval
	}

	return &key, nil
}

func (a *api) GetLastKeyByPurpose(ctx context.Context, keyPurpose types.KeyPurpose) (*types.Hash, error) {
	//TODO(cdamian): Add validation from the NFT branch

	acc, err := contextutil.Account(ctx)

	if err != nil {
		a.log.Errorf("Couldn't retrieve account from context: %s", err)

		return nil, ErrContextAccountRetrieval
	}

	meta, err := a.api.GetMetadataLatest()

	if err != nil {
		a.log.Errorf("Couldn't retrieve latest metadata: %s", err)

		return nil, ErrMetadataRetrieval
	}

	encodedKeyPurpose, err := types.Encode(keyPurpose)

	if err != nil {
		a.log.Errorf("Couldn't encode key purpose: %s", err)

		return nil, ErrKeyPurposeEncoding
	}

	encodedIdentity, err := types.Encode(acc.GetIdentity())

	if err != nil {
		a.log.Errorf("Couldn't encode identity: %s", err)

		return nil, ErrIdentityEncoding
	}

	storageKey, err := types.CreateStorageKey(
		meta,
		PalletName,
		LastKeyByPurposeStorageName,
		encodedIdentity,
		encodedKeyPurpose,
	)

	if err != nil {
		a.log.Errorf("Couldn't create storage key: %s", err)

		return nil, ErrStorageKeyCreation
	}

	var key types.Hash

	// TODO(cdamian): Use the OK from the NFT branch.
	if err := a.api.GetStorageLatest(storageKey, &key); err != nil {
		a.log.Errorf("Couldn't retrieve key from storage: %s", err)

		return nil, ErrKeyStorageRetrieval

	}

	return &key, nil
}
