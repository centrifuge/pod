//go:build unit
// +build unit

package jobs

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/config"

	"github.com/centrifuge/go-centrifuge/notification"
	"github.com/centrifuge/go-centrifuge/storage/leveldb"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/gocelery/v2"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"
)

func TestDispatch(t *testing.T) {
	t.Run("With webhook", func(t *testing.T) {
		t.Parallel()
		dispatch(t, true)
	})

	t.Run("Without webhook", func(t *testing.T) {
		t.Parallel()
		dispatch(t, false)
	})
}

func dispatch(t *testing.T, webhook bool) {
	ctx, s, resChan, did, d, mockAssert := setup(t, webhook)
	go s.ListenAndServe()
	defer s.Close()

	ctx, cancel := context.WithCancel(ctx)
	t.Cleanup(cancel)
	wg := new(sync.WaitGroup)
	wg.Add(1)
	go d.Start(ctx, wg, nil)
	name := hexutil.Encode(utils.RandomSlice(32))
	assert.True(t, d.RegisterRunnerFunc(name, func(args []interface{}, overrides map[string]interface{}) (interface{},
		error) {
		return args[0].(int) + args[1].(int), nil
	}))

	job := gocelery.NewRunnerFuncJob("Test", name, []interface{}{1, 2}, nil, time.Now())
	_, err := d.Job(did, job.ID)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, gocelery.ErrNotFound))

	res, err := d.Dispatch(did, job)
	assert.NoError(t, err)
	owner, err := d.(*dispatcher).jobOwner(job.ID)
	assert.NoError(t, err)
	assert.Equal(t, did, owner)
	r, err := res.Await(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 3, r)

	j, err := d.Job(did, job.ID)
	assert.NoError(t, err)
	assert.True(t, j.HasCompleted())
	assert.True(t, j.IsSuccessful())
	assert.Len(t, j.Tasks, 1)
	assert.True(t, j.LastTask().IsSuccessful())
	assert.Equal(t, j.LastTask().Tries, uint(1))
	assert.Equal(t, j.LastTask().Result, 3)

	nr, err := d.Result(did, job.ID)
	assert.NoError(t, err)
	r, err = nr.Await(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 3, r)

	if webhook {
		assert.True(t, bytes.Equal(job.ID[:], <-resChan))
	} else {
		assert.Len(t, resChan, 0)
	}
	mockAssert()
}

func setup(t *testing.T, webhook bool) (context.Context, *http.Server, chan []byte, identity.DID, Dispatcher, func()) {
	did := identity.NewDID(common.BytesToAddress(utils.RandomSlice(20)))
	db, err := leveldb.NewLevelDBStorage(leveldb.GetRandomTestStoragePath())
	assert.NoError(t, err)
	d, err := NewDispatcher(db, 10, 2*time.Minute)
	assert.NoError(t, err)
	resChan := make(chan []byte)
	s := prepareServer(t, resChan)
	url := ""
	if webhook {
		url = fmt.Sprintf("http://%s/webhook", s.Addr)
	}
	ctx, assert := getContext(t, did, url)
	return ctx, s, resChan, did, d, assert
}

func prepareServer(t *testing.T, resChan chan<- []byte) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/webhook", func(writer http.ResponseWriter, request *http.Request) {
		var resp notification.Message
		defer request.Body.Close()
		data, err := ioutil.ReadAll(request.Body)
		assert.NoError(t, err)

		err = json.Unmarshal(data, &resp)
		assert.NoError(t, err)
		writer.Write([]byte("success"))
		resChan <- resp.Job.ID
	})

	addr, _, err := utils.GetFreeAddrPort()
	assert.NoError(t, err)
	server := &http.Server{Addr: addr, Handler: mux}
	return server
}

func getContext(t *testing.T, did identity.DID, url string) (context.Context, func()) {
	cfgSrv := new(config.MockService)
	acc := new(config.MockAccount)
	acc.On("GetReceiveEventNotificationEndpoint").Return(url).Once()
	cfgSrv.On("GetAccount", did[:]).Return(acc, nil).Once()
	ctx := context.WithValue(
		context.Background(),
		bootstrap.NodeObjRegistry, map[string]interface{}{config.BootstrappedConfigStorage: cfgSrv})
	assert := func() {
		cfgSrv.AssertExpectations(t)
		acc.AssertExpectations(t)
	}

	return ctx, assert
}
