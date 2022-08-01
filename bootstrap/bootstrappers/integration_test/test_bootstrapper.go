package integration_test

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/testingutils"
	gsrpc "github.com/centrifuge/go-substrate-rpc-client/v4"
	logging "github.com/ipfs/go-log"
)

type Bootstrapper struct{}

var (
	once sync.Once
)

func (b *Bootstrapper) TestBootstrap(args map[string]interface{}) error {
	var err error

	once.Do(func() {
		if testingutils.IsCentChainRunning() {
			return
		}

		log := logging.Logger("integration_test_bootstrapper")

		configSrv, ok := args[config.BootstrappedConfigStorage].(config.Service)

		if !ok {
			err = errors.New("config service not initialised")
			return
		}

		if err := startCentChain(log); err != nil {
			err = fmt.Errorf("couldn't start Centrifuge Chain: %w", err)
			return
		}

		err = waitForOnboarding(log, configSrv)
	})

	return err
}

func (b *Bootstrapper) TestTearDown() error {
	return nil
}

const (
	onboardingTimeout  = 3 * time.Minute
	onboardingInterval = 5 * time.Second
)

func waitForOnboarding(log *logging.ZapEventLogger, configSrv config.Service) error {
	log.Infof("Waiting for onboarding to finish with timeout - %s", onboardingTimeout)

	cfg, err := configSrv.GetConfig()

	if err != nil {
		return fmt.Errorf("couldn't retrieve config: %w", err)
	}

	api, err := gsrpc.NewSubstrateAPI(cfg.GetCentChainNodeURL())

	if err != nil {
		return fmt.Errorf("couldn't create substrate API: %w", err)
	}

	ctx, canc := context.WithTimeout(context.Background(), onboardingTimeout)
	defer canc()

	ticker := time.NewTicker(onboardingInterval)
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
