//go:build integration || testworld

package integration_test

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"time"

	gsrpc "github.com/centrifuge/go-substrate-rpc-client/v4"
	"github.com/centrifuge/pod/testingutils/path"
	logging "github.com/ipfs/go-log"
)

var (
	log = logging.Logger("integration_test_bootstrapper")
)

type Bootstrapper struct{}

func (b *Bootstrapper) TestBootstrap(_ map[string]any) error {
	rand.Seed(time.Now().Unix())

	if err := os.Chdir(path.ProjectRoot); err != nil {
		log.Errorf("Couldn't change path to project root: %s", err)

		return err
	}

	if centChainStartInProgress() {
		log.Debug("Centrifuge chain start is in progress")

		return b.waitForOnboarding()
	}

	if err := startCentChain(); err != nil {
		return fmt.Errorf("couldn't start Centrifuge Chain: %w", err)
	}

	return b.waitForOnboarding()
}

func (b *Bootstrapper) TestTearDown() error {
	return nil
}

const (
	onboardingTimeout              = 5 * time.Minute
	onboardingCheckInterval        = 30 * time.Second
	onboardingCheckInitialInterval = 0 * time.Second

	defaultCentchainURL = "ws://localhost:9946"
)

func (b *Bootstrapper) waitForOnboarding() error {
	log.Infof("Waiting for onboarding to finish with timeout - %s", onboardingTimeout)

	ctx, canc := context.WithTimeout(context.Background(), onboardingTimeout)
	defer canc()

	checkInterval := onboardingCheckInitialInterval

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("context done while waiting for onboarding: %w", ctx.Err())
		case <-time.After(checkInterval):
			checkInterval = onboardingCheckInterval

			api, err := gsrpc.NewSubstrateAPI(defaultCentchainURL)

			if err != nil {
				continue
			}

			latestBlock, err := api.RPC.Chain.GetBlockLatest()

			api.Client.Close()

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
	centchainRunScriptPath = "build/scripts/run_centrifuge_chain.sh"
)

func startCentChain() error {
	log.Infof("Starting Centrifuge Chain")

	cmd := exec.Command("bash", "-c", centchainRunScriptPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout

	err := cmd.Run()

	if err != nil {
		return fmt.Errorf("couldn't start cent chain: %s", err)
	}

	return nil
}

var (
	centChainContainers = []string{
		"alice",
		"bob",
		"cc-alice",
	}
)

func centChainStartInProgress() bool {
	for _, containerName := range centChainContainers {
		if containerRunning(containerName) {
			return true
		}
	}

	return false
}

const (
	containerCheckCmdFormat = `docker ps -a --filter "name=%s" --filter "status=running" --quiet`
)

func containerRunning(containerName string) bool {
	cmd := fmt.Sprintf(containerCheckCmdFormat, containerName)
	o, err := exec.Command("/bin/sh", "-c", cmd).Output()

	if err != nil {
		panic(err)
	}

	return len(o) != 0
}
