package keystore

import (
	"context"

	"github.com/centrifuge/go-centrifuge/pallets/proxy"

	"github.com/centrifuge/chain-custom-types/pkg/keystore"
	proxyType "github.com/centrifuge/chain-custom-types/pkg/proxy"
	"github.com/centrifuge/go-centrifuge/centchain"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types/codec"
	logging "github.com/ipfs/go-log"
)

var (
	log = logging.Logger("keystore_api")
)

const (
	ErrContextAccountRetrieval  = errors.Error("couldn't retrieve account from context")
	ErrMetadataRetrieval        = errors.Error("couldn't retrieve metadata")
	ErrCallCreation             = errors.Error("couldn't create call")
	ErrSubmitAndWatchExtrinsic  = errors.Error("couldn't submit and watch extrinsic")
	ErrKeyIDEncoding            = errors.Error("couldn't encode key ID")
	ErrAccountIDEncoding        = errors.Error("couldn't encode identity")
	ErrKeyPurposeEncoding       = errors.Error("couldn't encode key purpose")
	ErrStorageKeyCreation       = errors.Error("couldn't create storage key")
	ErrKeyStorageRetrieval      = errors.Error("couldn't retrieve key from storage")
	ErrKeyNotFound              = errors.Error("key not found")
	ErrLastKeyByPurposeNotFound = errors.Error("last key by purpose not found")
)

const (
	PalletName = "Keystore"

	AddKeysCall    = PalletName + ".add_keys"
	RevokeKeysCall = PalletName + ".revoke_keys"

	KeysStorageName             = "Keys"
	LastKeyByPurposeStorageName = "LastKeyByPurpose"
)

//go:generate mockery --name API --structname APIMock --filename api_mock.go --inpackage

type API interface {
	AddKeys(ctx context.Context, keys []*keystore.AddKey) (*centchain.ExtrinsicInfo, error)
	RevokeKeys(ctx context.Context, keys []*types.Hash, keyPurpose keystore.KeyPurpose) (*centchain.ExtrinsicInfo, error)
	GetKey(accountID *types.AccountID, keyID *keystore.KeyID) (*keystore.Key, error)
	GetLastKeyByPurpose(accountID *types.AccountID, keyPurpose keystore.KeyPurpose) (*types.Hash, error)
}

type api struct {
	cfgService config.Service
	api        centchain.API
	proxyAPI   proxy.API
}

func NewAPI(cfgService config.Service, centAPI centchain.API, proxyAPI proxy.API) API {
	return &api{
		cfgService: cfgService,
		api:        centAPI,
		proxyAPI:   proxyAPI,
	}
}

func (a *api) AddKeys(ctx context.Context, keys []*keystore.AddKey) (*centchain.ExtrinsicInfo, error) {
	//TODO(cdamian): Add validation from the NFT branch

	acc, err := contextutil.Account(ctx)

	if err != nil {
		log.Errorf("Couldn't retrieve account from context: %s", err)

		return nil, ErrContextAccountRetrieval
	}

	meta, err := a.api.GetMetadataLatest()

	if err != nil {
		log.Errorf("Couldn't retrieve latest metadata: %s", err)

		return nil, ErrMetadataRetrieval
	}

	call, err := types.NewCall(meta, AddKeysCall, keys)

	if err != nil {
		log.Errorf("Couldn't create call: %s", err)

		return nil, ErrCallCreation
	}

	podOperator, err := a.cfgService.GetPodOperator()

	if err != nil {
		log.Errorf("Couldn't retrieve pod operator: %s", err)

		return nil, errors.ErrPodOperatorRetrieval
	}

	extInfo, err := a.proxyAPI.ProxyCall(
		ctx,
		acc.GetIdentity(),
		podOperator.ToKeyringPair(),
		types.NewOption(proxyType.KeystoreManagement),
		call,
	)

	if err != nil {
		log.Errorf("Couldn't perform proxy call: %s", err)

		return nil, errors.ErrProxyCall
	}

	return extInfo, nil
}

func (a *api) RevokeKeys(
	ctx context.Context,
	keys []*types.Hash,
	keyPurpose keystore.KeyPurpose,
) (*centchain.ExtrinsicInfo, error) {
	//TODO(cdamian): Add validation from the NFT branch

	acc, err := contextutil.Account(ctx)

	if err != nil {
		log.Errorf("Couldn't retrieve account from context: %s", err)

		return nil, ErrContextAccountRetrieval
	}

	meta, err := a.api.GetMetadataLatest()

	if err != nil {
		log.Errorf("Couldn't retrieve latest metadata: %s", err)

		return nil, ErrMetadataRetrieval
	}

	call, err := types.NewCall(meta, RevokeKeysCall, keys, keyPurpose)

	if err != nil {
		log.Errorf("Couldn't create call: %s", err)

		return nil, ErrCallCreation
	}

	podOperator, err := a.cfgService.GetPodOperator()

	if err != nil {
		log.Errorf("Couldn't retrieve pod operator: %s", err)

		return nil, errors.ErrPodOperatorRetrieval
	}

	extInfo, err := a.proxyAPI.ProxyCall(
		ctx,
		acc.GetIdentity(),
		podOperator.ToKeyringPair(),
		types.NewOption(proxyType.KeystoreManagement),
		call,
	)

	if err != nil {
		log.Errorf("Couldn't submit and watch extrinsic: %s", err)

		return nil, ErrSubmitAndWatchExtrinsic
	}

	return extInfo, nil
}

func (a *api) GetKey(accountID *types.AccountID, keyID *keystore.KeyID) (*keystore.Key, error) {
	meta, err := a.api.GetMetadataLatest()

	if err != nil {
		log.Errorf("Couldn't retrieve latest metadata: %s", err)

		return nil, ErrMetadataRetrieval
	}

	encodedKeyID, err := codec.Encode(keyID)

	if err != nil {
		log.Errorf("Couldn't encode key ID: %s", err)

		return nil, ErrKeyIDEncoding
	}

	encodedAccountID, err := codec.Encode(accountID)

	if err != nil {
		log.Errorf("Couldn't encode account ID: %s", err)

		return nil, ErrAccountIDEncoding
	}

	storageKey, err := types.CreateStorageKey(meta, PalletName, KeysStorageName, encodedAccountID, encodedKeyID)

	if err != nil {
		log.Errorf("Couldn't create storage key: %s", err)

		return nil, ErrStorageKeyCreation
	}

	var key keystore.Key

	ok, err := a.api.GetStorageLatest(storageKey, &key)

	if err != nil {
		log.Errorf("Couldn't retrieve key from storage: %s", err)

		return nil, ErrKeyStorageRetrieval
	}

	if !ok {
		log.Error("Key not found")

		return nil, ErrKeyNotFound
	}

	return &key, nil
}

func (a *api) GetLastKeyByPurpose(accountID *types.AccountID, keyPurpose keystore.KeyPurpose) (*types.Hash, error) {
	meta, err := a.api.GetMetadataLatest()

	if err != nil {
		log.Errorf("Couldn't retrieve latest metadata: %s", err)

		return nil, ErrMetadataRetrieval
	}

	encodedKeyPurpose, err := codec.Encode(keyPurpose)

	if err != nil {
		log.Errorf("Couldn't encode key purpose: %s", err)

		return nil, ErrKeyPurposeEncoding
	}

	encodedAccountID, err := codec.Encode(accountID)

	if err != nil {
		log.Errorf("Couldn't encode account ID: %s", err)

		return nil, ErrAccountIDEncoding
	}

	storageKey, err := types.CreateStorageKey(
		meta,
		PalletName,
		LastKeyByPurposeStorageName,
		encodedAccountID,
		encodedKeyPurpose,
	)

	if err != nil {
		log.Errorf("Couldn't create storage key: %s", err)

		return nil, ErrStorageKeyCreation
	}

	var key types.Hash

	ok, err := a.api.GetStorageLatest(storageKey, &key)

	if err != nil {
		log.Errorf("Couldn't retrieve key from storage: %s", err)

		return nil, ErrKeyStorageRetrieval

	}

	if !ok {
		log.Error("Last key by purpose not found")

		return nil, ErrLastKeyByPurposeNotFound
	}

	return &key, nil
}
