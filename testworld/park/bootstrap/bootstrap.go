//go:build testworld

package bootstrap

import (
	"fmt"

	"github.com/centrifuge/go-centrifuge/testingutils/keyrings"
	"github.com/centrifuge/go-centrifuge/testworld/park/host"
	"github.com/centrifuge/go-substrate-rpc-client/v4/signature"
	logging "github.com/ipfs/go-log"
)

var (
	log = logging.Logger("testworld-bootstrap")
)

var (
	testKeyringPairs = map[host.Name]signature.KeyringPair{
		host.Alice:   keyrings.AliceKeyRingPair,
		host.Bob:     keyrings.BobKeyRingPair,
		host.Charlie: keyrings.CharlieKeyRingPair,
		host.Dave:    keyrings.DaveKeyRingPair,
	}
)

func CreateTestHosts(webhookURL string) (map[host.Name]*host.Host, error) {
	var bootstrapPeers []string

	testHosts := make(map[host.Name]*host.Host)

	for hostName, keyringPair := range testKeyringPairs {
		log.Infof("\n\nBootstrapping host control unit for - %s\n", hostName)

		hostControlUnit, err := bootstrapHostControlUnit(&bootstrapPeers)

		if err != nil {
			return nil, fmt.Errorf("couldn't bootstrap test host services: %w", err)
		}

		if err := hostControlUnit.Start(); err != nil {
			return nil, fmt.Errorf("couldn't start control unit: %w", err)
		}

		log.Infof("\n\nBootstrapping host account for - %s\n", hostName)

		hostAccount, podAuthProxy, err := bootstrapHostAccount(hostControlUnit.GetServiceCtx(), keyringPair, webhookURL)

		if err != nil {
			return nil, fmt.Errorf("couldn't bootstrap test host account: %w", err)
		}

		log.Infof("\n\nCreating host for - %s\n", hostName)

		testHost, err := createHost(hostControlUnit, hostAccount, podAuthProxy)

		if err != nil {
			return nil, fmt.Errorf("create test host: %w", err)
		}

		testHosts[hostName] = testHost
	}

	return testHosts, nil
}
