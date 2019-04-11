// +build unit

package api

import (
	"context"
	"flag"
	"os"
	"sync"
	"testing"

	"github.com/centrifuge/go-centrifuge/documents/entity"

	"github.com/centrifuge/go-centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/bootstrap/bootstrappers/testlogging"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/config/configstore"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/documents/entityrelationship"
	"github.com/centrifuge/go-centrifuge/documents/invoice"
	"github.com/centrifuge/go-centrifuge/documents/purchaseorder"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/ethereum"
	"github.com/centrifuge/go-centrifuge/identity/ideth"
	"github.com/centrifuge/go-centrifuge/nft"
	"github.com/centrifuge/go-centrifuge/p2p"
	"github.com/centrifuge/go-centrifuge/queue"
	"github.com/centrifuge/go-centrifuge/storage/leveldb"
	"github.com/centrifuge/go-centrifuge/testingutils/commons"
	"github.com/centrifuge/go-centrifuge/transactions/txv1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var ctx = map[string]interface{}{}
var cfg config.Configuration

func TestMain(m *testing.M) {
	ethClient := &testingcommons.MockEthClient{}
	ethClient.On("GetEthClient").Return(nil)
	ctx[ethereum.BootstrappedEthereumClient] = ethClient

	ibootstappers := []bootstrap.TestBootstrapper{
		&testlogging.TestLoggingBootstrapper{},
		&config.Bootstrapper{},
		&leveldb.Bootstrapper{},
		txv1.Bootstrapper{},
		&queue.Bootstrapper{},
		&ideth.Bootstrapper{},
		&configstore.Bootstrapper{},
		anchors.Bootstrapper{},
		documents.Bootstrapper{},
		&invoice.Bootstrapper{},
		&entityrelationship.Bootstrapper{},
		&purchaseorder.Bootstrapper{},
		&ethereum.Bootstrapper{},
		&nft.Bootstrapper{},
		&queue.Starter{},
		p2p.Bootstrapper{},
		documents.PostBootstrapper{},
		&entity.Bootstrapper{},
	}
	bootstrap.RunTestBootstrappers(ibootstappers, ctx)

	cfg = ctx[bootstrap.BootstrappedConfig].(config.Configuration)
	flag.Parse()
	result := m.Run()
	bootstrap.RunTestTeardown(ibootstappers)
	os.Exit(result)
}

func TestCentAPIServer_StartContextCancel(t *testing.T) {
	cfg.Set("nodeHostname", "0.0.0.0")
	cfg.Set("nodePort", 9000)
	cfg.Set("centrifugeNetwork", "")
	capi := apiServer{config: cfg}
	ctx, canc := context.WithCancel(context.WithValue(context.Background(), bootstrap.NodeObjRegistry, ctx))
	startErr := make(chan error)
	var wg sync.WaitGroup
	wg.Add(1)
	go capi.Start(ctx, &wg, startErr)
	// cancel the context to shutdown the server
	canc()
	wg.Wait()
}

func TestCentAPIServer_StartListenError(t *testing.T) {
	// cause an error by using an invalid port
	cfg.Set("nodeHostname", "0.0.0.0")
	cfg.Set("nodePort", 100000000)
	cfg.Set("centrifugeNetwork", "")
	ctx, cancel := context.WithCancel(context.WithValue(context.Background(), bootstrap.NodeObjRegistry, ctx))
	defer cancel()
	capi := apiServer{config: cfg}
	startErr := make(chan error)
	var wg sync.WaitGroup
	wg.Add(1)
	go capi.Start(ctx, &wg, startErr)
	err := <-startErr
	wg.Wait()
	assert.NotNil(t, err, "Error should be not nil")
	assert.Equal(t, "listen tcp: address 100000000: invalid port", err.Error())
}

func TestCentAPIServer_FailedToGetRegistry(t *testing.T) {
	// cause an error by using an invalid port
	cfg.Set("nodeHostname", "0.0.0.0")
	cfg.Set("nodePort", 100000000)
	cfg.Set("centrifugeNetwork", "")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	capi := apiServer{config: cfg}
	startErr := make(chan error)
	var wg sync.WaitGroup
	wg.Add(1)
	go capi.Start(ctx, &wg, startErr)
	err := <-startErr
	wg.Wait()
	assert.NotNil(t, err, "Error should be not nil")
	assert.Equal(t, "failed to get NodeObjRegistry", err.Error())
}

func Test_auth(t *testing.T) {
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return ctx.Value(config.AccountHeaderKey), nil
	}

	// send ping path
	resp, err := auth(
		context.Background(),
		nil,
		&grpc.UnaryServerInfo{FullMethod: noAuthPaths[0]},
		handler,
	)
	assert.Nil(t, resp)
	assert.Nil(t, err)

	// send no auth
	resp, err = auth(
		context.Background(),
		nil,
		&grpc.UnaryServerInfo{FullMethod: "some method"},
		handler,
	)

	assert.Nil(t, resp)
	assert.True(t, errors.IsOfType(ErrNoAuthHeader, err))

	// send Auth
	ctx := metadata.NewIncomingContext(
		context.Background(),
		map[string][]string{"authorization": {"1234567890"}})

	resp, err = auth(
		ctx,
		nil,
		&grpc.UnaryServerInfo{FullMethod: "some method"},
		handler,
	)

	assert.Nil(t, err)
	assert.Equal(t, "1234567890", resp)
}
