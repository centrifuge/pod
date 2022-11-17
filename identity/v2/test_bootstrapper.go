//go:build integration || testworld

package v2

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"time"

	gsrpc "github.com/centrifuge/go-substrate-rpc-client/v4"

	keystoreTypes "github.com/centrifuge/chain-custom-types/pkg/keystore"
	proxyType "github.com/centrifuge/chain-custom-types/pkg/proxy"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/crypto"
	"github.com/centrifuge/go-centrifuge/pallets/keystore"
	"github.com/centrifuge/go-centrifuge/pallets/proxy"
	genericUtils "github.com/centrifuge/go-centrifuge/testingutils/generic"
	"github.com/centrifuge/go-centrifuge/testingutils/keyrings"
	proxyUtils "github.com/centrifuge/go-centrifuge/testingutils/proxy"
	"github.com/centrifuge/go-substrate-rpc-client/v4/signature"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
)

func (b *Bootstrapper) TestBootstrap(serviceCtx map[string]any) error {
	return b.Bootstrap(serviceCtx)
}

func (b *Bootstrapper) TestTearDown() error {
	return nil
}

type AccountTestBootstrapper struct {
	Bootstrapper
}

func (b *AccountTestBootstrapper) TestBootstrap(serviceCtx map[string]any) error {
	if err := b.Bootstrap(serviceCtx); err != nil {
		return err
	}

	log.Info("Generating test account for Alice")

	_, err := BootstrapTestAccount(serviceCtx, keyrings.AliceKeyRingPair)

	if err != nil {
		return fmt.Errorf("couldn't bootstrap test account for Alice: %w", err)
	}

	return nil
}

func (b *AccountTestBootstrapper) TestTearDown() error {
	return nil
}

func BootstrapTestAccount(
	serviceCtx map[string]any,
	accountKeyringPair signature.KeyringPair,
) (config.Account, error) {
	accountID, err := types.NewAccountID(accountKeyringPair.PublicKey)

	if err != nil {
		return nil, fmt.Errorf("couldn't get account ID: %w", err)
	}

	log.Info("Creating identity")

	acc, err := CreateTestIdentity(serviceCtx, accountID, "")

	if err != nil {
		return nil, fmt.Errorf("couldn't create test account: %w", err)
	}

	cfgService := genericUtils.GetService[config.Service](serviceCtx)

	podOperator, err := cfgService.GetPodOperator()

	if err != nil {
		return nil, fmt.Errorf("couldn't get pod operator: %w", err)
	}

	log.Info("Adding funds to pod operator")

	if err = AddFundsToAccount(serviceCtx, accountKeyringPair, podOperator.GetAccountID().ToBytes()); err != nil {
		return nil, fmt.Errorf("couldn't add funds to pod operator: %w", err)
	}

	proxyPairs := ProxyPairs{
		{
			Delegate:  podOperator.GetAccountID(),
			ProxyType: proxyType.PodOperation,
		},
		{
			Delegate:  podOperator.GetAccountID(),
			ProxyType: proxyType.KeystoreManagement,
		},
	}

	log.Info("Adding pod operator proxies")

	if err := AddAndWaitForTestProxies(serviceCtx, accountKeyringPair, proxyPairs); err != nil {
		return nil, fmt.Errorf("couldn't create test proxies: %w", err)
	}

	log.Info("Adding keys to keystore")

	if err := AddAccountKeysToStore(serviceCtx, acc); err != nil {
		return nil, fmt.Errorf("couldn't add keys to keystore: %w", err)
	}

	return acc, nil
}

func CreateTestIdentity(
	serviceCtx map[string]any,
	accountID *types.AccountID,
	webhookURL string,
) (config.Account, error) {
	cfgService := genericUtils.GetService[config.Service](serviceCtx)
	identityService := genericUtils.GetService[Service](serviceCtx)

	if acc, err := cfgService.GetAccount(accountID.ToBytes()); err == nil {
		log.Info("Account already created for - ", accountID.ToHexString())

		return acc, nil
	}

	acc, err := identityService.CreateIdentity(context.Background(), &CreateIdentityRequest{
		Identity:         accountID,
		WebhookURL:       webhookURL,
		PrecommitEnabled: true,
	})

	if err != nil {
		return nil, fmt.Errorf("couldn't create identity: %w", err)
	}

	return acc, nil
}

type ProxyPair struct {
	Delegate  *types.AccountID
	ProxyType proxyType.CentrifugeProxyType
}

type ProxyPairs []ProxyPair

func (p ProxyPairs) GetDelegateAccountIDs() []*types.AccountID {
	accountIDMap := make(map[string]struct{})

	var accountIDs []*types.AccountID

	for _, proxyPair := range p {
		if _, ok := accountIDMap[proxyPair.Delegate.ToHexString()]; ok {
			continue
		}

		accountIDMap[proxyPair.Delegate.ToHexString()] = struct{}{}

		accountIDs = append(accountIDs, proxyPair.Delegate)
	}

	return accountIDs
}

func AddAndWaitForTestProxies(
	serviceCtx map[string]any,
	delegatorKrp signature.KeyringPair,
	proxyPairs ProxyPairs,
) error {
	proxyAPI := genericUtils.GetService[proxy.API](serviceCtx)

	delegator, err := types.NewAccountID(delegatorKrp.PublicKey)

	if err != nil {
		return fmt.Errorf("couldn't create delegator account ID: %w", err)
	}

	ctx := context.Background()

	for _, proxyPair := range proxyPairs {
		if err := proxyAPI.AddProxy(ctx, proxyPair.Delegate, proxyPair.ProxyType, 0, delegatorKrp); err != nil {
			return fmt.Errorf("couldn't add proxy to %s: %w", delegator.ToHexString(), err)
		}
	}

	err = proxyUtils.WaitForProxiesToBeAdded(
		ctx,
		serviceCtx,
		delegator,
		proxyPairs.GetDelegateAccountIDs()...,
	)

	if err != nil {
		return fmt.Errorf("proxies were not added: %w", err)
	}

	return nil
}

func AddAccountKeysToStore(
	serviceCtx map[string]any,
	acc config.Account,
) error {
	unstoredAccountKeys, err := getUnstoredAccountKeys(serviceCtx, acc)
	if err != nil {
		return fmt.Errorf("couldn't get account keys: %w", err)
	}

	keystoreAPI := genericUtils.GetService[keystore.API](serviceCtx)

	var keys []*keystoreTypes.AddKey

	for _, unstoredAccountKey := range unstoredAccountKeys {
		keys = append(keys, &keystoreTypes.AddKey{
			Key:     unstoredAccountKey.Hash,
			Purpose: unstoredAccountKey.KeyPurpose,
			KeyType: keystoreTypes.KeyTypeECDSA,
		})
	}

	_, err = keystoreAPI.AddKeys(contextutil.WithAccount(context.Background(), acc), keys)
	if err != nil {
		return fmt.Errorf("couldn't store keys: %w", err)
	}

	return nil
}

const (
	defaultBalance         = "10000000000000000000000"
	balanceTransferTimeout = 15 * time.Minute
)

func AddFundsToAccount(
	serviceCtx map[string]any,
	senderKrp signature.KeyringPair,
	receiverPublicKey []byte,
) error {
	cfg := genericUtils.GetService[config.Configuration](serviceCtx)

	fundsClient, err := newFundsClient(cfg.GetCentChainNodeURL())

	if err != nil {
		return fmt.Errorf("couldn't create funds client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), balanceTransferTimeout)
	defer cancel()

	balance, ok := big.NewInt(0).SetString(defaultBalance, 10)

	if !ok {
		return errors.New("couldn't create balance amount")
	}

	if err := fundsClient.transfer(ctx, senderKrp, receiverPublicKey, types.NewUCompact(balance)); err != nil {
		return fmt.Errorf("couldn't transfer funds: %w", err)
	}

	return nil
}

func getUnstoredAccountKeys(
	serviceCtx map[string]any,
	acc config.Account,
) ([]*keystoreTypes.KeyID, error) {
	cfgService := genericUtils.GetService[config.Service](serviceCtx)
	cfg, err := cfgService.GetConfig()

	if err != nil {
		return nil, fmt.Errorf("couldn't get config: %w", err)
	}

	_, p2pPublicKey, err := crypto.ObtainP2PKeypair(cfg.GetP2PKeyPair())

	if err != nil {
		return nil, fmt.Errorf("couldn't obtain P2P key pair: %w", err)
	}

	p2pPublicKeyRaw, err := p2pPublicKey.Raw()

	if err != nil {
		return nil, fmt.Errorf("couldn't get raw P2P public key: %w", err)
	}

	keys := []*keystoreTypes.KeyID{
		{
			Hash:       types.NewHash(p2pPublicKeyRaw),
			KeyPurpose: keystoreTypes.KeyPurposeP2PDiscovery,
		},
		{
			Hash:       types.NewHash(acc.GetSigningPublicKey()),
			KeyPurpose: keystoreTypes.KeyPurposeP2PDocumentSigning,
		},
	}

	return filterUnstoredAccountKeys(serviceCtx, acc.GetIdentity(), keys)
}

func filterUnstoredAccountKeys(serviceCtx map[string]any, accountID *types.AccountID, keys []*keystoreTypes.KeyID) ([]*keystoreTypes.KeyID, error) {
	keystoreAPI := genericUtils.GetService[keystore.API](serviceCtx)

	return genericUtils.FilterSlice(keys, func(key *keystoreTypes.KeyID) (bool, error) {
		_, err := keystoreAPI.GetKey(accountID, key)

		if err != nil {
			if errors.Is(err, keystore.ErrKeyNotFound) {
				return true, nil
			}

			return false, err
		}

		return false, nil
	})
}

type fundsClient struct {
	api         *gsrpc.SubstrateAPI
	meta        *types.Metadata
	rv          *types.RuntimeVersion
	genesisHash types.Hash
}

func newFundsClient(url string) (*fundsClient, error) {
	api, err := gsrpc.NewSubstrateAPI(url)

	if err != nil {
		return nil, fmt.Errorf("couldn't get substrate API: %w", err)
	}

	meta, err := api.RPC.State.GetMetadataLatest()

	if err != nil {
		return nil, fmt.Errorf("couldn't get latest metadata: %w", err)
	}

	rv, err := api.RPC.State.GetRuntimeVersionLatest()
	if err != nil {
		return nil, fmt.Errorf("couldn't get latest runtime version: %w", err)
	}

	genesisHash, err := api.RPC.Chain.GetBlockHash(0)
	if err != nil {
		return nil, fmt.Errorf("couldn't get genesis hash: %w", err)
	}

	return &fundsClient{
		api,
		meta,
		rv,
		genesisHash,
	}, nil
}

const (
	submitTransferInterval = 1 * time.Second
)

func (f *fundsClient) transfer(ctx context.Context, from signature.KeyringPair, to []byte, balance types.UCompact) error {
	ticker := time.NewTicker(submitTransferInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("context done while submitting transfer: %w", ctx.Err())
		case <-ticker.C:
			if err := f.submitTransfer(ctx, from, to, balance); err == nil {
				return nil
			}
		}
	}
}

func (f *fundsClient) submitTransfer(ctx context.Context, from signature.KeyringPair, to []byte, balance types.UCompact) error {
	accountInfo, err := f.getAccountInfo(from.PublicKey)

	if err != nil {
		return err
	}

	dest, err := types.NewMultiAddressFromAccountID(to)

	if err != nil {
		return err
	}

	call, err := types.NewCall(f.meta, "Balances.transfer", dest, balance)

	if err != nil {
		return err
	}

	ext := types.NewExtrinsic(call)

	signOpts := types.SignatureOptions{
		BlockHash:          f.genesisHash, // using genesis since we're using immortal era
		Era:                types.ExtrinsicEra{IsMortalEra: false},
		GenesisHash:        f.genesisHash,
		Nonce:              types.NewUCompactFromUInt(uint64(accountInfo.Nonce)),
		SpecVersion:        f.rv.SpecVersion,
		Tip:                types.NewUCompactFromUInt(0),
		TransactionVersion: f.rv.TransactionVersion,
	}

	if err := ext.Sign(from, signOpts); err != nil {
		return err
	}

	sub, err := f.api.RPC.Author.SubmitAndWatchExtrinsic(ext)

	if err != nil {
		return err
	}

	defer sub.Unsubscribe()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("context done while waiting for transfer to be in block: %w", ctx.Err())
		case st := <-sub.Chan():
			ms, _ := st.MarshalJSON()

			log.Info("Got transfer status - ", string(ms))

			switch {
			case st.IsInBlock:
				return nil
			case st.IsUsurped:
				return errors.New("transfer did not go through")
			}
		}
	}
}

func (f *fundsClient) getAccountInfo(accountID []byte) (*types.AccountInfo, error) {
	storageKey, err := types.CreateStorageKey(f.meta, "System", "Account", accountID)

	if err != nil {
		return nil, err
	}

	var accountInfo types.AccountInfo

	ok, err := f.api.RPC.State.GetStorageLatest(storageKey, &accountInfo)

	if err != nil || !ok {
		return nil, errors.New("couldn't retrieve account info")
	}

	return &accountInfo, nil
}
