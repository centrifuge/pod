//go:build testworld

package bootstrap

import (
	"fmt"

	identityv2 "github.com/centrifuge/go-centrifuge/identity/v2"
	"github.com/centrifuge/go-centrifuge/testingutils/keyrings"
	"github.com/centrifuge/go-centrifuge/testworld/park/factory"
	"github.com/centrifuge/go-centrifuge/testworld/park/host"
	"github.com/centrifuge/go-substrate-rpc-client/v4/signature"
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

		log.Infof("\n\nTransferring funds to pod operator for - %s\n", hostName)

		err = identityv2.AddFundsToAccount(
			hostControlUnit.GetServiceCtx(),
			hostAccount.GetKeyringPair(),
			hostAccount.GetPodOperatorAccountID().ToBytes(),
		)

		if err != nil {
			return nil, fmt.Errorf("couldn't transfer funds to pod operator: %w", err)
		}

		log.Infof("\n\nAdding test proxies for - %s\n", hostName)

		if err := factory.AddTestHostAccountProxies(hostControlUnit.GetServiceCtx(), hostAccount); err != nil {
			return nil, fmt.Errorf("couldn't add test host account proxies: %w", err)
		}

		log.Infof("\n\nStoring public keys for - %s\n", hostName)

		if err := identityv2.AddAccountKeysToStore(hostControlUnit.GetServiceCtx(), hostAccount.GetAccount()); err != nil {
			return nil, fmt.Errorf("couldn't add test account keys to store: %w", err)
		}

		log.Infof("\n\nCreating host for - %s\n", hostName)

		testHosts[hostName] = host.NewHost(hostAccount, hostControlUnit)

		bootstrapPeers = append(bootstrapPeers, hostControlUnit.GetP2PAddress())
	}

	return testHosts, nil
}
