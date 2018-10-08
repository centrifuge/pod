// +build unit

package documenterror

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {

	err := New("test err msg")
	assert.Error(t, err, "New should return an error")
}

func TestAppend(t *testing.T) {

	err := fmt.Errorf("test error")

	err = Append(nil, err)
	assert.Equal(t, 1, Len(err), "err should only include one error")

	err = Append(err, fmt.Errorf("second error"))
	assert.Equal(t, 2, Len(err), "err should include two errors")
	assert.Equal(t, 2, len(Errors(err)), "returned error array should include two errors")

	err = Append(fmt.Errorf("err 1"), fmt.Errorf("err 2"))
	assert.Equal(t, 2, Len(err), "err should include two errors")

}

func TestLen(t *testing.T) {

	err := fmt.Errorf("test error")

	assert.Equal(t, 0, Len(nil), "nil should return len 0")
	assert.Equal(t, 1, Len(err), "normal error should have length 1")

}

func TestErrors(t *testing.T) {

	err := Append(fmt.Errorf("err 1"), fmt.Errorf("err 2"))
	errArray := Errors(err)
	assert.Equal(t, 2, len(errArray), "array should contain two errors")

	errArray = Errors(fmt.Errorf("err 1"))
	assert.Equal(t, 1, len(errArray), "array should contain one errors")

	errArray = Errors(nil)
	assert.Equal(t, 0, len(errArray), "array should be nil")

}
