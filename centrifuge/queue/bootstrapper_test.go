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
	context := map[string]interface{}{}
	err := InstallQueuedTask(context, MockQueuedTask{})
	assert.Nil(t, err, "Installation of tasks should be successful")
	assert.Equal(t, 1, len(context[BootstrappedQueuedTasks].([]QueuedTask)))
}

func TestInstallQueuedTaskappendingToListOfTasks(t *testing.T) {
	queuedTasks := []QueuedTask{MockQueuedTask{}}
	context := map[string]interface{}{BootstrappedQueuedTasks: queuedTasks}
	err := InstallQueuedTask(context, MockQueuedTask{})
	assert.Nil(t, err, "Installation of tasks should be successful")
	assert.Equal(t, 2, len(context[BootstrappedQueuedTasks].([]QueuedTask)))
}

func TestInstallQueuedTaskQueuedTasksListHasInvalidType(t *testing.T) {
	err := InstallQueuedTask(map[string]interface{}{BootstrappedQueuedTasks: 1}, MockQueuedTask{})
	assert.NotNil(t, err, "Installation of tasks should NOT be successful")
}
