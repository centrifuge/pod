//go:build integration || testworld

package v2

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/centrifuge/pod/pallets/utility"

	proxyType "github.com/centrifuge/chain-custom-types/pkg/proxy"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/centrifuge/pod/centchain"
	genericUtils "github.com/centrifuge/pod/testingutils/generic"

	"github.com/centrifuge/pod/config"

	"github.com/centrifuge/go-substrate-rpc-client/v4/signature"
	"github.com/centrifuge/pod/pallets"
	"github.com/centrifuge/pod/testingutils/keyrings"
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

const (
	accountBootstrapTimeout = 10 * time.Minute
)

func (b *AccountTestBootstrapper) TestBootstrap(serviceCtx map[string]any) error {
	if err := b.Bootstrap(serviceCtx); err != nil {
		return err
	}

	log.Info("Generating test account")

	ctx, cancel := context.WithTimeout(context.Background(), accountBootstrapTimeout)
	defer cancel()

	if _, err := BootstrapTestAccount(ctx, serviceCtx, getRandomTestKeyring()); err != nil {
		return fmt.Errorf("couldn't bootstrap test account: %w", err)
	}

	return nil
}

func (b *AccountTestBootstrapper) TestTearDown() error {
	return nil
}

var (
	testKeyrings = []signature.KeyringPair{
		keyrings.AliceKeyRingPair,
		keyrings.BobKeyRingPair,
		keyrings.CharlieKeyRingPair,
		keyrings.DaveKeyRingPair,
		keyrings.EveKeyRingPair,
		keyrings.FerdieKeyRingPair,
	}
)

func getRandomTestKeyring() signature.KeyringPair {
	rand.Seed(time.Now().Unix())

	return testKeyrings[rand.Intn(len(testKeyrings))]
}

func BootstrapTestAccount(
	ctx context.Context,
	serviceCtx map[string]any,
	originKrp signature.KeyringPair,
) (config.Account, error) {
	anonymousProxyAccountID, err := pallets.CreateAnonymousProxy(serviceCtx, originKrp)

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

	if err = pallets.ExecuteWithTestClient(ctx, serviceCtx, originKrp, utility.BatchCalls(calls...)); err != nil {
		return nil, fmt.Errorf("couldn't execute post account bootstrap: %w", err)
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
		pallets.GetBalanceTransferCallCreationFn(defaultBalance, acc.GetIdentity().ToBytes()),
		pallets.GetBalanceTransferCallCreationFn(defaultBalance, podOperator.GetAccountID().ToBytes()),
	}

	postBootstrapFns = append(
		postBootstrapFns,
		pallets.GetAddProxyCallCreationFns(
			acc.GetIdentity(),
			pallets.ProxyPairs{
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

	addKeysCall, err := pallets.GetAddKeysCall(serviceCtx, acc)

	if err != nil {
		return nil, fmt.Errorf("couldn't get AddKeys call: %w", err)
	}

	postBootstrapFns = append(postBootstrapFns, addKeysCall)

	return postBootstrapFns, nil
}
