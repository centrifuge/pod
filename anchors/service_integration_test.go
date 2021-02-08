// +build integration

package anchors

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/bootstrap/bootstrappers/testlogging"
	"github.com/centrifuge/go-centrifuge/centchain"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/config/configstore"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/crypto"
	"github.com/centrifuge/go-centrifuge/ethereum"
	"github.com/centrifuge/go-centrifuge/identity/ideth"
	"github.com/centrifuge/go-centrifuge/jobs/jobsv1"
	"github.com/centrifuge/go-centrifuge/jobs/jobsv2"
	"github.com/centrifuge/go-centrifuge/queue"
	"github.com/centrifuge/go-centrifuge/storage/leveldb"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common/hexutil"
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
		jobsv1.Bootstrapper{},
		&queue.Bootstrapper{},
		jobsv2.Bootstrapper{},
		centchain.Bootstrapper{},
		ethereum.Bootstrapper{},
		&ideth.Bootstrapper{},
		&configstore.Bootstrapper{},
		Bootstrapper{},
	}

	ctx := make(map[string]interface{})
	bootstrap.RunTestBootstrappers(bootstappers, ctx)
	configSrv = ctx[config.BootstrappedConfigStorage].(config.Service)
	anchorSrv = ctx[BootstrappedAnchorService].(Service)
	dispatcher := ctx[jobsv2.BootstrappedDispatcher].(jobsv2.Dispatcher)
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

	accs, err := configSrv.GetAccounts()
	assert.NoError(t, err)
	assert.True(t, len(accs) > 0)
	ctx, err := contextutil.New(context.Background(), accs[0])
	assert.NoError(t, err)
	fmt.Println(hexutil.Encode(pre), anchorID.String(), hexutil.Encode(docRoot[:]),
		hexutil.Encode(signingRoot[:]))

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
