//go:build integration
// +build integration

package centchain_test

import (
	"context"
	"os"
	"sync"
	"testing"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	cc "github.com/centrifuge/go-centrifuge/bootstrap/bootstrappers/testingbootstrap"
	"github.com/centrifuge/go-centrifuge/centchain"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/stretchr/testify/assert"
)

var api centchain.API
var cfg config.Configuration
var cfgSrv config.Service

func TestMain(m *testing.M) {
	ctx := cc.TestFunctionalEthereumBootstrap()
	api = ctx[centchain.BootstrappedCentChainClient].(centchain.API)
	dispatcher := ctx[jobs.BootstrappedDispatcher].(jobs.Dispatcher)
	cfg = ctx[bootstrap.BootstrappedConfig].(config.Configuration)
	cfgSrv = ctx[config.BootstrappedConfigStorage].(config.Service)
	ctxh, canc := context.WithCancel(context.Background())
	wg := new(sync.WaitGroup)
	wg.Add(1)
	go dispatcher.Start(ctxh, wg, nil)
	result := m.Run()
	cc.TestFunctionalEthereumTearDown()
	canc()
	wg.Wait()
	os.Exit(result)
}

func TestApi_SubmitAndWatch(t *testing.T) {
	meta, err := api.GetMetadataLatest()
	assert.NoError(t, err)

	call, err := types.NewCall(meta, "System.remark", []byte{})
	assert.NoError(t, err)

	id, err := cfg.GetIdentityID()
	assert.NoError(t, err)

	acc, err := cfgSrv.GetAccount(id)
	assert.NoError(t, err)

	caa := acc.GetCentChainAccount()
	kr, err := caa.KeyRingPair()
	assert.NoError(t, err)

	info, err := api.SubmitAndWatch(contextutil.WithAccount(context.Background(), acc), meta, call, kr)
	assert.NoError(t, err)

	for _, event := range info.Events {
		if event.Name == centchain.ExtrinsicFailedEventName {
			t.Fatalf("Extrinsic failed")
		}
	}
}
