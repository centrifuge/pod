//go:build integration || testworld

package integration_test

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	pathUtils "github.com/centrifuge/go-centrifuge/testingutils/path"

	"github.com/centrifuge/go-centrifuge/testingutils"
	gsrpc "github.com/centrifuge/go-substrate-rpc-client/v4"
	logging "github.com/ipfs/go-log"
)

var (
	log = logging.Logger("integration_test_bootstrapper")
)

type Bootstrapper struct{}

func (b *Bootstrapper) TestBootstrap(_ map[string]any) error {
	if err := os.Chdir(pathUtils.ProjectRoot); err != nil {
		log.Errorf("Couldn't change path to project root: %s", err)

		return err
	}

	if testingutils.IsCentChainRunning() {
		log.Debug("Centrifuge chain is already running, skipping bootstrapper")

		return nil
	}

	if err := startCentChain(log); err != nil {
		return fmt.Errorf("couldn't start Centrifuge Chain: %w", err)
	}

	return b.waitForOnboarding()
}

func (b *Bootstrapper) TestTearDown() error {
	return nil
}

const (
	onboardingTimeout       = 3 * time.Minute
	onboardingCheckInterval = 5 * time.Second

	defaultCentchainURL = "ws://localhost:9946"
)

func (b *Bootstrapper) waitForOnboarding() error {
	log.Infof("Waiting for onboarding to finish with timeout - %s", onboardingTimeout)

	api, err := gsrpc.NewSubstrateAPI(defaultCentchainURL)

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
	centchainRunScriptPath = "build/scripts/test-dependencies/test-centchain/run.sh"
)

func startCentChain(log *logging.ZapEventLogger) error {
	log.Infof("Starting Centrifuge Chain")

	//scriptPath := path.AppendPathToProjectRoot(centchainRunScriptPath)

	cmd := exec.Command("bash", "-c", centchainRunScriptPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout

	err := cmd.Run()

	if err != nil {
		return fmt.Errorf("couldn't start cent chain: %s", err)
	}

	return nil
}
