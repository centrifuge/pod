//go:build integration || testworld

package v2

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/centrifuge/go-centrifuge/pallets/utility"

	keystoreTypes "github.com/centrifuge/chain-custom-types/pkg/keystore"
	proxyType "github.com/centrifuge/chain-custom-types/pkg/proxy"
	"github.com/centrifuge/go-centrifuge/centchain"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/crypto"
	"github.com/centrifuge/go-centrifuge/pallets/keystore"
	"github.com/centrifuge/go-centrifuge/pallets/proxy"
	genericUtils "github.com/centrifuge/go-centrifuge/testingutils/generic"
	"github.com/centrifuge/go-substrate-rpc-client/v4/signature"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
)

func BootstrapTestAccount(
	ctx context.Context,
	serviceCtx map[string]any,
	originKrp signature.KeyringPair,
) (config.Account, error) {
	anonymousProxyAccountID, err := CreateAnonymousProxy(serviceCtx, originKrp)

	if err != nil {
		return nil, fmt.Errorf("couldn't get create anonymous proxy: %w", err)
	}

	log.Info("Creating identity")

	acc, err := CreateTestIdentity(serviceCtx, anonymousProxyAccountID, "")

	if err != nil {
		return nil, fmt.Errorf("couldn't create test account: %w", err)
	}

	calls, err := getPostAccountBootstrapCalls(serviceCtx, acc)

	if err != nil {
		return nil, fmt.Errorf("couldn't get post account bootstrap calls: %w", err)
	}

	if err = ExecutePostAccountBootstrap(ctx, serviceCtx, originKrp, calls...); err != nil {
		return nil, fmt.Errorf("couldn't execute post account bootstrap: %w", err)
	}

	return acc, nil
}

const (
	createAnonymousProxyTimeout = 10 * time.Minute
)

func CreateAnonymousProxy(
	serviceCtx map[string]any,
	originKrp signature.KeyringPair,
) (*types.AccountID, error) {
	cfg := genericUtils.GetService[config.Configuration](serviceCtx)

	testClient, err := centchain.NewTestClient(cfg.GetCentChainNodeURL())

	if err != nil {
		return nil, fmt.Errorf("couldn't create funds client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), createAnonymousProxyTimeout)
	defer cancel()

	// TODO(cdamian): The following retry logic is required because, in some cases, there are no "AnonymousCreated"
	// events in the block where the extrinsic was created.
	anonymousProxyCreateFn := func() (*types.AccountID, error) {
		fn := getCreateAnonymousProxyCallCreationFn(proxyType.Any, 0, 0)

		blockHash, err := testClient.SubmitAndWait(ctx, originKrp, fn)

		if err != nil {
			return nil, fmt.Errorf("couldn't create anonymous proxy: %w", err)
		}

		events, err := testClient.GetEvents(*blockHash)

		if err != nil {
			return nil, fmt.Errorf("couldn't get events: %w", err)
		}

		return getAnonymousProxyCreatedByAccount(originKrp, events)
	}

	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("context expired while creating proxy: %w", ctx.Err())
		default:
			anonymousProxyAccountID, err := anonymousProxyCreateFn()

			if err == nil {
				return anonymousProxyAccountID, nil
			}

			log.Errorf("Couldn't create anonymous proxy, retrying. Error: %s", err)
		}
	}
}

func getAnonymousProxyCreatedByAccount(
	originKrp signature.KeyringPair,
	events *centchain.Events,
) (*types.AccountID, error) {
	if len(events.Proxy_PureCreated) == 0 {
		return nil, errors.New("no 'AnonymousCreated' events")
	}

	for _, event := range events.Proxy_PureCreated {
		if bytes.Equal(event.Who.ToBytes(), originKrp.PublicKey) {
			return &event.Pure, nil
		}
	}

	return nil, errors.New("anonymous proxy not found")
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

const (
	defaultBalance = "10000000000000000000000"
)

func getPostAccountBootstrapCalls(serviceCtx map[string]any, acc config.Account) ([]centchain.CallProviderFn, error) {
	cfgService := genericUtils.GetService[config.Service](serviceCtx)

	podOperator, err := cfgService.GetPodOperator()

	if err != nil {
		return nil, fmt.Errorf("couldn't get pod operator: %w", err)
	}

	postBootstrapFns := []centchain.CallProviderFn{
		GetBalanceTransferCallCreationFn(defaultBalance, acc.GetIdentity().ToBytes()),
		GetBalanceTransferCallCreationFn(defaultBalance, podOperator.GetAccountID().ToBytes()),
	}

	postBootstrapFns = append(
		postBootstrapFns,
		GetAddProxyCallCreationFns(
			acc.GetIdentity(),
			ProxyPairs{
				{
					Delegate:  podOperator.GetAccountID(),
					ProxyType: proxyType.PodOperation,
				},
				{
					Delegate:  podOperator.GetAccountID(),
					ProxyType: proxyType.KeystoreManagement,
				},
			},
		)...,
	)

	addKeysCall, err := GetAddKeysCall(serviceCtx, acc)

	if err != nil {
		return nil, fmt.Errorf("couldn't get AddKeys call: %w", err)
	}

	postBootstrapFns = append(postBootstrapFns, addKeysCall)

	return postBootstrapFns, nil
}

func ExecutePostAccountBootstrap(
	ctx context.Context,
	serviceCtx map[string]any,
	originKrp signature.KeyringPair,
	callCreationFns ...centchain.CallProviderFn,
) error {
	cfg := genericUtils.GetService[config.Configuration](serviceCtx)

	testClient, err := centchain.NewTestClient(cfg.GetCentChainNodeURL())

	if err != nil {
		return fmt.Errorf("couldn't create funds client: %w", err)
	}

	defer testClient.Close()

	if _, err = testClient.SubmitAndWait(ctx, originKrp, utility.BatchCalls(callCreationFns...)); err != nil {
		return fmt.Errorf("couldn't submit post account bootstrap batch call: %w", err)
	}

	return nil
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

func GetAddProxyCallCreationFns(anonymousProxyAccountID *types.AccountID, proxyPairs ProxyPairs) []centchain.CallProviderFn {
	var callCreationFns []centchain.CallProviderFn

	for _, proxyPair := range proxyPairs {
		callCreationFn := getAddProxyToAnonymousProxyCall(anonymousProxyAccountID, proxyPair.Delegate, proxyPair.ProxyType)

		callCreationFns = append(callCreationFns, callCreationFn)
	}

	return callCreationFns
}

func getCreateAnonymousProxyCallCreationFn(
	pt proxyType.CentrifugeProxyType,
	delay types.U32,
	index types.U16,
) centchain.CallProviderFn {
	return func(meta *types.Metadata) (*types.Call, error) {
		call, err := types.NewCall(
			meta,
			proxy.ProxyCreatePure,
			pt,
			delay,
			index,
		)

		if err != nil {
			return nil, err
		}

		return &call, nil
	}
}

func getAddProxyToAnonymousProxyCall(
	anonymousProxyID *types.AccountID,
	delegate *types.AccountID,
	pt proxyType.CentrifugeProxyType,
) centchain.CallProviderFn {
	return func(meta *types.Metadata) (*types.Call, error) {
		delegateMultiAddress, err := types.NewMultiAddressFromAccountID(delegate.ToBytes())

		if err != nil {
			return nil, err
		}

		addProxyCall, err := types.NewCall(
			meta,
			proxy.ProxyAdd,
			delegateMultiAddress,
			pt,
			types.U32(0),
		)

		if err != nil {
			return nil, err
		}

		delegatorMultiAddress, err := types.NewMultiAddressFromAccountID(anonymousProxyID.ToBytes())

		if err != nil {
			return nil, err
		}

		proxyCall, err := types.NewCall(
			meta,
			proxy.ProxyCall,
			delegatorMultiAddress,
			types.NewOption(proxyType.Any),
			addProxyCall,
		)

		if err != nil {
			return nil, err
		}

		return &proxyCall, nil
	}
}

func GetBalanceTransferCallCreationFn(balance string, receiverAccountID []byte) centchain.CallProviderFn {
	return func(meta *types.Metadata) (*types.Call, error) {
		dest, err := types.NewMultiAddressFromAccountID(receiverAccountID)

		if err != nil {
			return nil, err
		}

		b, ok := big.NewInt(0).SetString(balance, 10)

		if !ok {
			return nil, errors.New("couldn't create balance int")
		}

		call, err := types.NewCall(meta, "Balances.transfer", dest, types.NewUCompact(b))

		if err != nil {
			return nil, err
		}

		return &call, nil
	}
}

func getAddKeysArgsForAccount(
	serviceCtx map[string]any,
	acc config.Account,
) ([]*keystoreTypes.AddKey, error) {
	unstoredAccountKeys, err := getUnstoredAccountKeys(serviceCtx, acc)
	if err != nil {
		return nil, fmt.Errorf("couldn't get account keys: %w", err)
	}

	var keys []*keystoreTypes.AddKey

	for _, unstoredAccountKey := range unstoredAccountKeys {
		keys = append(keys, &keystoreTypes.AddKey{
			Key:     unstoredAccountKey.Hash,
			Purpose: unstoredAccountKey.KeyPurpose,
			KeyType: keystoreTypes.KeyTypeECDSA,
		})
	}

	return keys, nil
}

func GetAddKeysCall(
	serviceCtx map[string]any,
	acc config.Account,
) (centchain.CallProviderFn, error) {
	keys, err := getAddKeysArgsForAccount(serviceCtx, acc)

	if err != nil {
		return nil, err
	}

	return func(meta *types.Metadata) (*types.Call, error) {
		addKeysCall, err := types.NewCall(meta, keystore.AddKeysCall, keys)

		if err != nil {
			return nil, err
		}

		delegatorMultiAddress, err := types.NewMultiAddressFromAccountID(acc.GetIdentity().ToBytes())

		if err != nil {
			return nil, err
		}

		proxyCall, err := types.NewCall(
			meta,
			proxy.ProxyCall,
			delegatorMultiAddress,
			types.NewOption(proxyType.Any),
			addKeysCall,
		)

		if err != nil {
			return nil, err
		}

		return &proxyCall, nil
	}, nil
}
