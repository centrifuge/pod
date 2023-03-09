//go:build unit

package jobs

import (
	"context"
	"encoding/gob"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/centrifuge/gocelery/v2"
	"github.com/centrifuge/pod/bootstrap"
	"github.com/centrifuge/pod/config"
	"github.com/centrifuge/pod/notification"
	"github.com/centrifuge/pod/storage/leveldb"
	testingcommons "github.com/centrifuge/pod/testingutils/common"
	"github.com/stretchr/testify/assert"
)

const (
	tempDirPattern = "dispatcher-test-*"
)

func TestNewDispatcher(t *testing.T) {
	randomStoragePath, err := testingcommons.GetRandomTestStoragePath(tempDirPattern)
	assert.NoError(t, err)

	defer func() {
		_ = os.RemoveAll(randomStoragePath)
	}()

	db, err := leveldb.NewLevelDBStorage(randomStoragePath)
	assert.NoError(t, err)

	dispatcher, err := NewDispatcher(db, 10, 1*time.Second)
	assert.NoError(t, err)
	assert.NotNil(t, dispatcher)
}

func TestDispatcher_Start(t *testing.T) {
	randomStoragePath, err := testingcommons.GetRandomTestStoragePath(tempDirPattern)
	assert.NoError(t, err)

	defer func() {
		_ = os.RemoveAll(randomStoragePath)
	}()

	db, err := leveldb.NewLevelDBStorage(randomStoragePath)
	assert.NoError(t, err)

	dispatcher, err := NewDispatcher(db, 10, 1*time.Second)
	assert.NoError(t, err)
	assert.NotNil(t, dispatcher)

	configServiceMock := config.NewServiceMock(t)
	serviceCtx := map[string]any{
		config.BootstrappedConfigStorage: configServiceMock,
	}

	var wg sync.WaitGroup
	startupErrChan := make(chan error, 1)

	ctx := context.WithValue(context.Background(), bootstrap.NodeObjRegistry, serviceCtx)

	ctx, cancel := context.WithCancel(ctx)

	wg.Add(1)

	go dispatcher.Start(ctx, &wg, startupErrChan)

	select {
	case err := <-startupErrChan:
		assert.Nil(t, err)
	case <-time.After(3 * time.Second):
	}

	cancel()

	wg.Wait()
}

func TestDispatcher_Start_CanceledContext(t *testing.T) {
	randomStoragePath, err := testingcommons.GetRandomTestStoragePath(tempDirPattern)
	assert.NoError(t, err)

	defer func() {
		_ = os.RemoveAll(randomStoragePath)
	}()

	db, err := leveldb.NewLevelDBStorage(randomStoragePath)
	assert.NoError(t, err)

	dispatcher, err := NewDispatcher(db, 10, 1*time.Second)
	assert.NoError(t, err)
	assert.NotNil(t, dispatcher)

	configServiceMock := config.NewServiceMock(t)
	serviceCtx := map[string]any{
		config.BootstrappedConfigStorage: configServiceMock,
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	ctx = context.WithValue(ctx, bootstrap.NodeObjRegistry, serviceCtx)

	var wg sync.WaitGroup
	startupErrChan := make(chan error, 1)

	wg.Add(1)

	go dispatcher.Start(ctx, &wg, startupErrChan)

	select {
	case err := <-startupErrChan:
		assert.NotNil(t, err)
	case <-time.After(3 * time.Second):
		assert.Fail(t, "Expected start error")
	}

	wg.Wait()
}

func TestDispatcher_Start_MissingNodeObjRegistry(t *testing.T) {
	randomStoragePath, err := testingcommons.GetRandomTestStoragePath(tempDirPattern)
	assert.NoError(t, err)

	defer func() {
		_ = os.RemoveAll(randomStoragePath)
	}()

	db, err := leveldb.NewLevelDBStorage(randomStoragePath)
	assert.NoError(t, err)

	dispatcher, err := NewDispatcher(db, 10, 1*time.Second)
	assert.NoError(t, err)
	assert.NotNil(t, dispatcher)

	var wg sync.WaitGroup
	startupErrChan := make(chan error, 1)

	ctx, cancel := context.WithCancel(context.Background())

	wg.Add(1)

	go dispatcher.Start(ctx, &wg, startupErrChan)

	select {
	case err := <-startupErrChan:
		assert.NotNil(t, err)
	case <-time.After(3 * time.Second):
		assert.Fail(t, "Expected start error")
	}

	cancel()

	wg.Wait()
}

func TestDispatcher_Start_MissingConfigService(t *testing.T) {
	randomStoragePath, err := testingcommons.GetRandomTestStoragePath(tempDirPattern)
	assert.NoError(t, err)

	defer func() {
		_ = os.RemoveAll(randomStoragePath)
	}()

	db, err := leveldb.NewLevelDBStorage(randomStoragePath)
	assert.NoError(t, err)

	dispatcher, err := NewDispatcher(db, 10, 1*time.Second)
	assert.NoError(t, err)
	assert.NotNil(t, dispatcher)

	var serviceCtx map[string]any

	var wg sync.WaitGroup
	startupErrChan := make(chan error, 1)

	ctx := context.WithValue(context.Background(), bootstrap.NodeObjRegistry, serviceCtx)

	ctx, cancel := context.WithCancel(ctx)

	wg.Add(1)

	go dispatcher.Start(ctx, &wg, startupErrChan)

	select {
	case err := <-startupErrChan:
		assert.NotNil(t, err)
	case <-time.After(3 * time.Second):
	}

	cancel()

	wg.Wait()
}

func TestDispatcher_Dispatch_WithRunner(t *testing.T) {
	randomStoragePath, err := testingcommons.GetRandomTestStoragePath(tempDirPattern)
	assert.NoError(t, err)

	defer func() {
		_ = os.RemoveAll(randomStoragePath)
	}()

	db, err := leveldb.NewLevelDBStorage(randomStoragePath)
	assert.NoError(t, err)

	dispatcher, err := NewDispatcher(db, 10, 1*time.Second)
	assert.NoError(t, err)
	assert.NotNil(t, dispatcher)

	// Start a test server that should receive the job notification message.
	notificationReceivedChan := make(chan struct{})

	testServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		var resp notification.Message
		defer request.Body.Close()
		defer close(notificationReceivedChan)

		data, err := ioutil.ReadAll(request.Body)
		assert.NoError(t, err)

		err = json.Unmarshal(data, &resp)
		assert.NoError(t, err)

		writer.WriteHeader(http.StatusOK)
	}))
	defer testServer.Close()

	// Create the account and config service mocks that will be used for retrieving the webhook URL.
	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetWebhookURL").
		Return(testServer.URL)

	configServiceMock := config.NewServiceMock(t)
	configServiceMock.On("GetAccount", accountID.ToBytes()).
		Return(accountMock, nil)

	serviceCtx := map[string]any{
		config.BootstrappedConfigStorage: configServiceMock,
	}

	// Start the dispatcher
	var wg sync.WaitGroup
	startupErrChan := make(chan error, 1)

	ctx := context.WithValue(context.Background(), bootstrap.NodeObjRegistry, serviceCtx)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	wg.Add(1)

	go dispatcher.Start(ctx, &wg, startupErrChan)

	select {
	case err := <-startupErrChan:
		assert.Nil(t, err)
	case <-time.After(3 * time.Second):
	}

	// Register a runner with a set of test tasks.
	tasksChan := make(chan bool)

	taskResult := struct {
		Result bool
	}{
		Result: true,
	}

	gob.Register(taskResult)

	testArgs := []any{
		"first_arg",
		2,
	}

	loadTasksFn := func() map[string]Task {
		return map[string]Task{
			"first_task": {
				RunnerFunc: func(args []interface{}, overrides map[string]interface{}) (result interface{}, err error) {
					assert.Equal(t, testArgs, args)

					assert.Equal(t, "initial_override", overrides["initial_override"])

					overrides["first_task"] = "some_override_1"

					tasksChan <- true

					return nil, nil
				},
				Next: "second_task",
			},
			"second_task": {
				RunnerFunc: func(args []interface{}, overrides map[string]interface{}) (result interface{}, err error) {
					assert.Equal(t, testArgs, args)

					assert.Equal(t, "some_override_1", overrides["first_task"])

					overrides["second_task"] = "some_override_2"

					tasksChan <- true

					return nil, nil
				},
				Next: "third_task",
			},
			"third_task": {
				RunnerFunc: func(args []interface{}, overrides map[string]interface{}) (result interface{}, err error) {
					assert.Equal(t, testArgs, args)

					assert.Equal(t, "some_override_2", overrides["second_task"])

					tasksChan <- true

					return taskResult, nil
				},
			},
		}
	}

	jobName := "test-job"

	registerRes := dispatcher.RegisterRunner(jobName, &testJob{loadTasksFn: loadTasksFn})
	assert.True(t, registerRes)

	// Dispatch the job.
	job := gocelery.NewRunnerJob(
		"Test description",
		jobName,
		"first_task",
		testArgs,
		map[string]interface{}{"initial_override": "initial_override"},
		time.Time{},
	)

	res, err := dispatcher.Dispatch(accountID, job)
	assert.NoError(t, err)

	// Ensure that the job Result waits until the job finishes.
	awaitCtx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	awaitChan := make(chan struct{})

	go func() {
		defer close(awaitChan)

		awaitRes, err := res.Await(awaitCtx)
		assert.NoError(t, err)
		assert.Equal(t, taskResult, awaitRes)
	}()

	// Check that each task was executed before the job finishes.
	taskResCount := 0
	notificationReceived := false

checkLoop:
	for {
		select {
		case taskRes := <-tasksChan:
			assert.True(t, taskRes)
			taskResCount++
		case <-awaitChan:
			break checkLoop
		case <-awaitCtx.Done():
			assert.Fail(t, "Await context done")
			break checkLoop
		}
	}

	assert.Equal(t, 3, taskResCount)

	// Check that the notification was sent to the test server.
	notificationWaitCtx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	select {
	case <-notificationWaitCtx.Done():
		assert.Fail(t, "Notification wait context done")
	case <-notificationReceivedChan:
		notificationReceived = true
	}

	assert.True(t, notificationReceived)

	resJob, err := dispatcher.Job(accountID, job.ID)
	assert.NoError(t, err)
	assert.Equal(t, job.ID, resJob.ID)
	assert.Equal(t, job.Runner, resJob.Runner)
	assert.Equal(t, job.Desc, resJob.Desc)
}

func TestDispatcher_Dispatch_WithRunnerFunc(t *testing.T) {
	randomStoragePath, err := testingcommons.GetRandomTestStoragePath(tempDirPattern)
	assert.NoError(t, err)

	defer func() {
		_ = os.RemoveAll(randomStoragePath)
	}()

	db, err := leveldb.NewLevelDBStorage(randomStoragePath)
	assert.NoError(t, err)

	dispatcher, err := NewDispatcher(db, 10, 1*time.Second)
	assert.NoError(t, err)
	assert.NotNil(t, dispatcher)

	// Start a test server that should receive the job notification message.
	notificationReceivedChan := make(chan struct{})

	testServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		var resp notification.Message
		defer request.Body.Close()
		defer close(notificationReceivedChan)

		data, err := ioutil.ReadAll(request.Body)
		assert.NoError(t, err)

		err = json.Unmarshal(data, &resp)
		assert.NoError(t, err)

		writer.WriteHeader(http.StatusOK)
	}))
	defer testServer.Close()

	// Create the account and config service mocks that will be used for retrieving the webhook URL.
	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetWebhookURL").
		Return(testServer.URL)

	configServiceMock := config.NewServiceMock(t)
	configServiceMock.On("GetAccount", accountID.ToBytes()).
		Return(accountMock, nil)

	serviceCtx := map[string]any{
		config.BootstrappedConfigStorage: configServiceMock,
	}

	// Start the dispatcher
	var wg sync.WaitGroup
	startupErrChan := make(chan error, 1)

	ctx := context.WithValue(context.Background(), bootstrap.NodeObjRegistry, serviceCtx)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	wg.Add(1)

	go dispatcher.Start(ctx, &wg, startupErrChan)

	select {
	case err := <-startupErrChan:
		assert.Nil(t, err)
	case <-time.After(3 * time.Second):
	}

	// Register a runner with a set of test tasks.
	tasksChan := make(chan bool)

	taskResult := struct {
		Result bool
	}{
		Result: true,
	}

	gob.Register(taskResult)

	testArgs := []any{
		"first_arg",
		2,
	}

	jobName := "test-job"

	registerRes := dispatcher.RegisterRunnerFunc(
		jobName,
		func(args []interface{}, overrides map[string]interface{}) (result interface{}, err error) {
			assert.Equal(t, testArgs, args)

			assert.Equal(t, "initial_override", overrides["initial_override"])

			tasksChan <- true
			return taskResult, nil
		},
	)
	assert.True(t, registerRes)

	// Dispatch the job.
	job := gocelery.NewRunnerFuncJob(
		"Test description",
		jobName,
		testArgs,
		map[string]interface{}{"initial_override": "initial_override"},
		time.Time{},
	)

	res, err := dispatcher.Dispatch(accountID, job)
	assert.NoError(t, err)

	// Ensure that the job Result waits until the job finishes.
	awaitCtx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	awaitChan := make(chan struct{})

	go func() {
		defer close(awaitChan)

		awaitRes, err := res.Await(awaitCtx)
		assert.NoError(t, err)
		assert.Equal(t, taskResult, awaitRes)
	}()

	// Check that each task was executed before the job finishes.
	taskResCount := 0
	notificationReceived := false

checkLoop:
	for {
		select {
		case taskRes := <-tasksChan:
			assert.True(t, taskRes)
			taskResCount++
		case <-awaitChan:
			break checkLoop
		case <-awaitCtx.Done():
			assert.Fail(t, "Await context done")
			break checkLoop
		}
	}

	assert.Equal(t, 1, taskResCount)

	// Check that the notification was sent to the test server.
	notificationWaitCtx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	select {
	case <-notificationWaitCtx.Done():
		assert.Fail(t, "Notification wait context done")
	case <-notificationReceivedChan:
		notificationReceived = true
	}

	assert.True(t, notificationReceived)

	resJob, err := dispatcher.Job(accountID, job.ID)
	assert.NoError(t, err)
	assert.Equal(t, job.ID, resJob.ID)
	assert.Equal(t, job.Runner, resJob.Runner)
	assert.Equal(t, job.Desc, resJob.Desc)
}

type testJob struct {
	loadTasksFn func() map[string]Task
	Base
}

func (t *testJob) New() gocelery.Runner {
	tasks := t.loadTasksFn()

	t.Base = NewBase(tasks)

	return t
}
