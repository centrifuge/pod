//go:build integration

package anchors

import (
	"context"
	"os"
	"sync"
	"testing"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/bootstrap/bootstrappers/integration_test"
	"github.com/centrifuge/go-centrifuge/bootstrap/bootstrappers/testlogging"
	"github.com/centrifuge/go-centrifuge/centchain"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/config/configstore"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/crypto"
	v2 "github.com/centrifuge/go-centrifuge/identity/v2"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/storage/leveldb"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/blake2b"
)

var (
	configSrv config.Service
	anchorSrv Service
)

func TestMain(m *testing.M) {
	var bootstappers = []bootstrap.TestBootstrapper{
		&testlogging.TestLoggingBootstrapper{},
		&config.Bootstrapper{},
		&leveldb.Bootstrapper{},
		jobs.Bootstrapper{},
		&configstore.Bootstrapper{},
		&integration_test.IntegrationTestBootstrapper{},
		centchain.Bootstrapper{},
		&v2.Bootstrapper{},
		Bootstrapper{},
	}

	ctx := make(map[string]interface{})
	bootstrap.RunTestBootstrappers(bootstappers, ctx)
	configSrv = ctx[config.BootstrappedConfigStorage].(config.Service)
	anchorSrv = ctx[BootstrappedAnchorService].(Service)
	dispatcher := ctx[jobs.BootstrappedDispatcher].(jobs.Dispatcher)
	ctxh, canc := context.WithCancel(context.Background())
	wg := new(sync.WaitGroup)

	wg.Add(1)
	go dispatcher.Start(ctxh, wg, nil)

	result := m.Run()

	bootstrap.RunTestTeardown(bootstappers)
	canc()
	wg.Wait()
	os.Exit(result)
}

func TestCommitAnchor(t *testing.T) {
	pre, id, err := crypto.GenerateHashPair(32)
	assert.NoError(t, err)
	anchorID, err := ToAnchorID(id)
	assert.NoError(t, err)

	signingRoot := utils.RandomByte32()
	proof := utils.RandomByte32()
	b2bHash, err := blake2b.New256(nil)
	assert.NoError(t, err)
	_, err = b2bHash.Write(append(signingRoot[:], proof[:]...))
	assert.NoError(t, err)
	docRoot, err := ToDocumentRoot(b2bHash.Sum(nil))
	assert.NoError(t, err)

	accounts, err := configSrv.GetAccounts()
	assert.NoError(t, err)
	assert.True(t, len(accounts) > 0)

	acc := accounts[0]

	ctx := contextutil.WithAccount(context.Background(), acc)
	ctx = context.WithValue(ctx, config.AccountHeaderKey, acc.GetIdentity())

	_, _, err = anchorSrv.GetAnchorData(anchorID)
	assert.Error(t, err)

	// precommit document
	err = anchorSrv.PreCommitAnchor(ctx, anchorID, signingRoot)
	assert.NoError(t, err)

	// commit document
	preImage, err := ToAnchorID(pre)
	assert.NoError(t, err)
	err = anchorSrv.CommitAnchor(ctx, preImage, docRoot, proof)
	assert.NoError(t, err)

	// get committed doc root
	gDocRoot, _, err := anchorSrv.GetAnchorData(anchorID)
	assert.NoError(t, err)
	assert.Equal(t, docRoot, gDocRoot)
}
