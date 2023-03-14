//go:build integration

package anchors_test

import (
	"context"
	"os"
	"testing"

	genericUtils "github.com/centrifuge/pod/testingutils/generic"

	"github.com/centrifuge/pod/config/configstore"

	"github.com/centrifuge/pod/bootstrap"
	"github.com/centrifuge/pod/bootstrap/bootstrappers/integration_test"
	"github.com/centrifuge/pod/bootstrap/bootstrappers/testlogging"
	"github.com/centrifuge/pod/centchain"
	"github.com/centrifuge/pod/config"
	"github.com/centrifuge/pod/contextutil"
	"github.com/centrifuge/pod/crypto"
	"github.com/centrifuge/pod/dispatcher"
	v2 "github.com/centrifuge/pod/identity/v2"
	"github.com/centrifuge/pod/jobs"
	"github.com/centrifuge/pod/pallets"
	"github.com/centrifuge/pod/pallets/anchors"
	"github.com/centrifuge/pod/storage/leveldb"
	"github.com/centrifuge/pod/utils"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/blake2b"
)

var integrationTestBootstrappers = []bootstrap.TestBootstrapper{
	&integration_test.Bootstrapper{},
	&testlogging.TestLoggingBootstrapper{},
	&config.Bootstrapper{},
	&leveldb.Bootstrapper{},
	&configstore.Bootstrapper{},
	&jobs.Bootstrapper{},
	centchain.Bootstrapper{},
	&pallets.Bootstrapper{},
	&dispatcher.Bootstrapper{},
	&v2.AccountTestBootstrapper{},
}

var (
	configSrv config.Service
	anchorSrv anchors.API
)

func TestMain(m *testing.M) {
	ctx := bootstrap.RunTestBootstrappers(integrationTestBootstrappers, nil)
	configSrv = genericUtils.GetService[config.Service](ctx)
	anchorSrv = genericUtils.GetService[anchors.API](ctx)

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
