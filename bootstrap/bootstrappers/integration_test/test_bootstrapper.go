package integration_test

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/testingutils"
	gsrpc "github.com/centrifuge/go-substrate-rpc-client/v4"
	logging "github.com/ipfs/go-log"
)

var (
	log = logging.Logger("integration_test_bootstrapper")
)

type Bootstrapper struct{}

func (b *Bootstrapper) TestBootstrap(args map[string]any) error {
	if testingutils.IsCentChainRunning() {
		log.Debug("Centrifuge chain is already running, skipping bootstrapper")

		return nil
	}

	configFile, ok := args[config.BootstrappedConfigFile].(string)

	if !ok {
		return errors.New("config file not present")
	}

	cfg := config.LoadConfiguration(configFile)

	if err := startCentChain(log); err != nil {
		return fmt.Errorf("couldn't start Centrifuge Chain: %w", err)
	}

	return waitForOnboarding(log, cfg)
}

func (b *Bootstrapper) TestTearDown() error {
	return nil
}

const (
	onboardingTimeout       = 3 * time.Minute
	onboardingCheckInterval = 5 * time.Second
)

func waitForOnboarding(log *logging.ZapEventLogger, cfg config.Configuration) error {
	log.Infof("Waiting for onboarding to finish with timeout - %s", onboardingTimeout)

	api, err := gsrpc.NewSubstrateAPI(cfg.GetCentChainNodeURL())

	if err != nil {
		return fmt.Errorf("couldn't create substrate API: %w", err)
	}

	ctx, canc := context.WithTimeout(context.Background(), onboardingTimeout)
	defer canc()

	ticker := time.NewTicker(onboardingCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("context done while waiting for onboarding: %w", ctx.Err())
		case <-ticker.C:
			latestBlock, err := api.RPC.Chain.GetBlockLatest()

			if err != nil {
				return fmt.Errorf("couldn't retrieve latest block: %w", err)
			}

			if latestBlock.Block.Header.Number > 0 {
				log.Info("Onboarding finished")

				return nil
			}
		}
	}
}

const (
	centchainRunScript = "build/scripts/test-dependencies/test-centchain/run.sh"
)

func startCentChain(log *logging.ZapEventLogger) error {
	log.Infof("Starting Centrifuge Chain")

	cmd := exec.Command("bash", "-c", centchainRunScript)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout

	err := cmd.Run()

	if err != nil {
		return fmt.Errorf("couldn't start cent chain: %s", err)
	}

	return nil
}
