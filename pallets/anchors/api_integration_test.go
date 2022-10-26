//go:build integration

package anchors_test

import (
	"context"
	"os"
	"testing"

	"github.com/centrifuge/go-centrifuge/config/configstore"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/bootstrap/bootstrappers/integration_test"
	"github.com/centrifuge/go-centrifuge/bootstrap/bootstrappers/testlogging"
	"github.com/centrifuge/go-centrifuge/centchain"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/crypto"
	"github.com/centrifuge/go-centrifuge/dispatcher"
	v2 "github.com/centrifuge/go-centrifuge/identity/v2"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/pallets"
	"github.com/centrifuge/go-centrifuge/pallets/anchors"
	"github.com/centrifuge/go-centrifuge/storage/leveldb"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/blake2b"
)

var integrationTestBootstrappers = []bootstrap.TestBootstrapper{
	&testlogging.TestLoggingBootstrapper{},
	&config.Bootstrapper{},
	&leveldb.Bootstrapper{},
	&configstore.Bootstrapper{},
	&jobs.Bootstrapper{},
	&integration_test.Bootstrapper{},
	centchain.Bootstrapper{},
	&pallets.Bootstrapper{},
	&dispatcher.Bootstrapper{},
	&v2.Bootstrapper{},
}

var (
	configSrv config.Service
	anchorSrv anchors.API
)

func TestMain(m *testing.M) {
	ctx := bootstrap.RunTestBootstrappers(integrationTestBootstrappers, nil)
	configSrv = ctx[config.BootstrappedConfigStorage].(config.Service)
	anchorSrv = ctx[pallets.BootstrappedAnchorService].(anchors.API)

	result := m.Run()

	bootstrap.RunTestTeardown(integrationTestBootstrappers)

	os.Exit(result)
}

func TestCommitAnchor(t *testing.T) {
	pre, id, err := crypto.GenerateHashPair(32)
	assert.NoError(t, err)
	anchorID, err := anchors.ToAnchorID(id)
	assert.NoError(t, err)

	signingRoot := utils.RandomByte32()
	proof := utils.RandomByte32()
	b2bHash, err := blake2b.New256(nil)
	assert.NoError(t, err)
	_, err = b2bHash.Write(append(signingRoot[:], proof[:]...))
	assert.NoError(t, err)
	docRoot, err := anchors.ToDocumentRoot(b2bHash.Sum(nil))
	assert.NoError(t, err)

	accounts, err := configSrv.GetAccounts()
	assert.NoError(t, err)
	assert.True(t, len(accounts) > 0)

	acc := accounts[0]

	ctx := contextutil.WithAccount(context.Background(), acc)

	_, _, err = anchorSrv.GetAnchorData(anchorID)
	assert.Error(t, err)

	// precommit document
	err = anchorSrv.PreCommitAnchor(ctx, anchorID, signingRoot)
	assert.NoError(t, err)

	// commit document
	preImage, err := anchors.ToAnchorID(pre)
	assert.NoError(t, err)
	err = anchorSrv.CommitAnchor(ctx, preImage, docRoot, proof)
	assert.NoError(t, err)

	// get committed doc root
	gDocRoot, _, err := anchorSrv.GetAnchorData(anchorID)
	assert.NoError(t, err)
	assert.Equal(t, docRoot, gDocRoot)
}
