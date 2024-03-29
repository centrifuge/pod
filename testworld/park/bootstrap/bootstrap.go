//go:build testworld

package bootstrap

import (
	"context"
	"fmt"
	"time"

	"github.com/centrifuge/pod/pallets/utility"

	"github.com/centrifuge/pod/pallets"

	proxyType "github.com/centrifuge/chain-custom-types/pkg/proxy"
	"github.com/centrifuge/go-substrate-rpc-client/v4/signature"
	"github.com/centrifuge/pod/centchain"
	"github.com/centrifuge/pod/testingutils/keyrings"
	"github.com/centrifuge/pod/testworld/park/factory"
	"github.com/centrifuge/pod/testworld/park/host"
	logging "github.com/ipfs/go-log"
)

var (
	log = logging.Logger("testworld-bootstrap")
)

var (
	hostsMap = map[host.Name]signature.KeyringPair{
		host.Alice:   keyrings.AliceKeyRingPair,
		host.Bob:     keyrings.BobKeyRingPair,
		host.Charlie: keyrings.CharlieKeyRingPair,
	}
)

const (
	postAccountBootstrapTimeout = 10 * time.Minute
)

func CreateTestHosts(webhookURL string) (map[host.Name]*host.Host, error) {
	var bootstrapPeers []string

	testHosts := make(map[host.Name]*host.Host)

	for hostName, testHostKrp := range hostsMap {
		log.Infof("\n\nCreating host control unit for - %s\n", hostName)

		hostControlUnit, err := factory.CreateHostControlUnit(bootstrapPeers)

		if err != nil {
			return nil, fmt.Errorf("couldn't create test host services: %w", err)
		}

		if err := hostControlUnit.Start(); err != nil {
			return nil, fmt.Errorf("couldn't start control unit: %w", err)
		}

		log.Infof("\n\nCreating host account for - %s\n", hostName)

		hostAccount, err := factory.CreateTestHostAccount(hostControlUnit.GetServiceCtx(), testHostKrp, webhookURL)

		if err != nil {
			return nil, fmt.Errorf("couldn't create test host account: %w", err)
		}

		log.Infof("\n\nExecuting post account bootstrap for - %s\n", hostName)

		postAccountBootstrapCalls, err := getPostAccountBootstrapCalls(hostControlUnit.GetServiceCtx(), hostAccount)

		if err != nil {
			return nil, fmt.Errorf("couldn't get post account bootstrap calls: %w", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), postAccountBootstrapTimeout)
		defer cancel()

		err = pallets.ExecuteWithTestClient(
			ctx,
			hostControlUnit.GetServiceCtx(),
			testHostKrp,
			utility.BatchCalls(postAccountBootstrapCalls...),
		)

		if err != nil {
			return nil, fmt.Errorf("couldn't execute post account bootstrap calls: %w", err)
		}

		log.Infof("\n\nCreating host for - %s\n", hostName)

		testHosts[hostName] = host.NewHost(hostAccount, hostControlUnit, testHostKrp)

		bootstrapPeers = append(bootstrapPeers, hostControlUnit.GetP2PAddress())
	}

	return testHosts, nil
}

const (
	defaultBalance = "10000000000000000000000"
)

func GetPostAccountCreationCalls(serviceCtx map[string]any, hostAccount *host.Account) ([]centchain.CallProviderFn, error) {
	postCreationCalls := []centchain.CallProviderFn{
		pallets.GetBalanceTransferCallCreationFn(defaultBalance, hostAccount.GetAccountID().ToBytes()),
	}

	postCreationCalls = append(
		postCreationCalls,
		pallets.GetAddProxyCallCreationFns(
			hostAccount.GetAccountID(),
			pallets.ProxyPairs{
				{
					Delegate:  hostAccount.GetPodOperatorAccountID(),
					ProxyType: proxyType.PodOperation,
				},
				{
					Delegate:  hostAccount.GetPodAuthProxyAccountID(),
					ProxyType: proxyType.PodAuth,
				},
			})...,
	)

	addKeysCall, err := pallets.GetAddKeysCall(serviceCtx, hostAccount.GetAccount())

	if err != nil {
		return nil, fmt.Errorf("couldn't get AddKeys call: %w", err)
	}

	postCreationCalls = append(postCreationCalls, addKeysCall)

	return postCreationCalls, nil
}

func getPostAccountBootstrapCalls(serviceCtx map[string]any, hostAccount *host.Account) ([]centchain.CallProviderFn, error) {
	postBootstrapCalls := []centchain.CallProviderFn{
		pallets.GetBalanceTransferCallCreationFn(defaultBalance, hostAccount.GetPodOperatorAccountID().ToBytes()),
	}

	postCreationCalls, err := GetPostAccountCreationCalls(serviceCtx, hostAccount)

	if err != nil {
		return nil, err
	}

	return append(postBootstrapCalls, postCreationCalls...), nil
}
