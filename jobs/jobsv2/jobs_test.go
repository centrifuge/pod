// +build unit

package jobsv2

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/storage/leveldb"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/gocelery/v2"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

func TestDispatcher(t *testing.T) {
	acc := identity.NewDID(common.BytesToAddress(utils.RandomSlice(20)))
	job := gocelery.NewRunnerFuncJob("Test", "add", []interface{}{1, 2}, nil, time.Now())
	db, err := leveldb.NewLevelDBStorage(leveldb.GetRandomTestStoragePath())
	assert.NoError(t, err)
	d, err := NewDispatcher(db, 10, 2*time.Minute)
	assert.NoError(t, err)
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	go d.Start(ctx)
	assert.True(t, d.RegisterRunnerFunc("add", func(args []interface{}, overrides map[string]interface{}) (interface{},
		error) {
		return args[0].(int) + args[1].(int), nil
	}))

	_, err = d.Job(acc, job.ID)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, gocelery.ErrNotFound))

	res, err := d.Dispatch(acc, job)
	assert.NoError(t, err)
	r, err := res.Await(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 3, r)

	j, err := d.Job(acc, job.ID)
	assert.NoError(t, err)
	assert.True(t, j.HasCompleted())
	assert.True(t, j.IsSuccessful())
	assert.Len(t, j.Tasks, 1)
	assert.True(t, j.LastTask().IsSuccessful())
	assert.Equal(t, j.LastTask().Tries, uint(1))
	assert.Equal(t, j.LastTask().Result, 3)

	nr, err := d.Result(acc, job.ID)
	assert.NoError(t, err)
	r, err = nr.Await(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 3, r)
}
