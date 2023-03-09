//go:build integration || testworld

package v2

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/centrifuge/go-substrate-rpc-client/v4/signature"
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
