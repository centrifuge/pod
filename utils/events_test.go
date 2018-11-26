// +build unit

package utils

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockIterator struct {
	next bool
	err  error
}

func (m *mockIterator) Next() bool {
	return m.next
}

func (m *mockIterator) Error() error {
	return m.err
}

func (m *mockIterator) Close() error {
	return nil
}

func TestLookForEvent_iterator_error(t *testing.T) {
	iter := &mockIterator{next: false, err: fmt.Errorf("failed iterator")}
	err := LookForEvent(iter)
	assert.NotNil(t, err, "error should be non nil")
	assert.Contains(t, err.Error(), "failed iterator")
}

func TestLookForEvent_event_not_found(t *testing.T) {
	iter := &mockIterator{}
	err := LookForEvent(iter)
	assert.NotNil(t, err, "error should be non nil")
	assert.Equal(t, err, ErrEventNotFound)
}

func TestLookForEvent_success(t *testing.T) {
	iter := &mockIterator{next: true}
	err := LookForEvent(iter)
	assert.Nil(t, err, "error should be nil")
}
