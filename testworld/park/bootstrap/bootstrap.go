//go:build testworld

package bootstrap

import (
	"fmt"

	"github.com/centrifuge/go-centrifuge/testingutils/keyrings"
	"github.com/centrifuge/go-centrifuge/testworld/park/factory"
	"github.com/centrifuge/go-centrifuge/testworld/park/host"
	"github.com/centrifuge/go-substrate-rpc-client/v4/signature"
	logging "github.com/ipfs/go-log"
)

var (
	log = logging.Logger("testworld-bootstrap")
)

type accountInfo struct {
	keyringPair           signature.KeyringPair
	podOperatorSecretSeed string
}

var (
	testKeyringPairs = map[host.Name]accountInfo{
		host.Alice: {
			keyringPair:           keyrings.AliceKeyRingPair,
			podOperatorSecretSeed: keyrings.FerdieKeyRingPair.URI,
		},
		host.Bob: {
			keyringPair:           keyrings.BobKeyRingPair,
			podOperatorSecretSeed: keyrings.EveKeyRingPair.URI,
		},
		host.Charlie: {
			keyringPair:           keyrings.CharlieKeyRingPair,
			podOperatorSecretSeed: keyrings.DaveKeyRingPair.URI,
		},
	}
)

func CreateTestHosts(webhookURL string) (map[host.Name]*host.Host, error) {
	var bootstrapPeers []string

	testHosts := make(map[host.Name]*host.Host)

	for hostName, accountInfo := range testKeyringPairs {
		log.Infof("\n\nCreating host control unit for - %s\n", hostName)

		hostControlUnit, err := factory.CreateHostControlUnit(bootstrapPeers, accountInfo.podOperatorSecretSeed)

		if err != nil {
			return nil, fmt.Errorf("couldn't create test host services: %w", err)
		}

		if err := hostControlUnit.Start(); err != nil {
			return nil, fmt.Errorf("couldn't start control unit: %w", err)
		}

		log.Infof("\n\nCreating host account for - %s\n", hostName)

		hostAccount, err := factory.CreateTestHostAccount(hostControlUnit.GetServiceCtx(), accountInfo.keyringPair, webhookURL)

		if err != nil {
			return nil, fmt.Errorf("couldn't create test host account: %w", err)
		}

		log.Infof("\n\nCreating host for - %s\n", hostName)

		testHosts[hostName] = host.NewHost(hostAccount, hostControlUnit)

		bootstrapPeers = append(bootstrapPeers, hostControlUnit.GetP2PAddress())
	}

	return testHosts, nil
}
