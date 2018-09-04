// +build unit

package queue

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type MockQueuedTask struct{}

func (MockQueuedTask) Name() string {
	return "MockTask"
}

func (MockQueuedTask) Init() error {
	return nil
}

func TestInstallQueuedTaskCreatingNewListOfTasks(t *testing.T) {
	called := false
	err := InstallQueuedTask(map[string]interface{}{}, func() QueuedTask {
		called = true
		return nil
	})
	assert.Nil(t, err, "Installation of tasks should be successful")
	assert.True(t, called, "task creation callback should have been executed")
}

func TestInstallQueuedTaskappendingToListOfTasks(t *testing.T) {
	called := false
	queuedTasks := []QueuedTask{MockQueuedTask{}}
	context := map[string]interface{}{BootstrappedQueuedTasks: queuedTasks}
	err := InstallQueuedTask(context, func() QueuedTask {
		called = true
		return nil
	})
	assert.Nil(t, err, "Installation of tasks should be successful")
	assert.True(t, called, "task creation callback should have been executed")
	assert.Equal(t, 2, len(context[BootstrappedQueuedTasks].([]QueuedTask)))
}

func TestInstallQueuedTaskQueuedTasksListHasInvalidType(t *testing.T) {
	called := false
	err := InstallQueuedTask(map[string]interface{}{BootstrappedQueuedTasks: 1}, func() QueuedTask {
		called = true
		return nil
	})
	assert.NotNil(t, err, "Installation of tasks should NOT be successful")
}
