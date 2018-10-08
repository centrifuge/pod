// +build unit

package documents

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewError(t *testing.T) {

	err := NewError("test err msg")
	assert.Error(t, err, "New should return an error")
}

func TestAppendError(t *testing.T) {

	err := fmt.Errorf("test error")

	err = AppendError(nil, err)
	assert.Equal(t, 1, LenError(err), "err should only include one error")

	err = AppendError(err, fmt.Errorf("second error"))
	assert.Equal(t, 2, LenError(err), "err should include two errors")
	assert.Equal(t, 2, len(Errors(err)), "returned error array should include two errors")

	err = AppendError(fmt.Errorf("err 1"), fmt.Errorf("err 2"))
	assert.Equal(t, 2, LenError(err), "err should include two errors")

}

func TestLenError(t *testing.T) {

	err := fmt.Errorf("test error")

	assert.Equal(t, 0, LenError(nil), "nil should return len 0")
	assert.Equal(t, 1, LenError(err), "normal error should have length 1")

}

func TestErrors(t *testing.T) {

	err := AppendError(fmt.Errorf("err 1"), fmt.Errorf("err 2"))
	errArray := Errors(err)
	assert.Equal(t, 2, len(errArray), "array should contain two errors")

	errArray = Errors(fmt.Errorf("err 1"))
	assert.Equal(t, 1, len(errArray), "array should contain one errors")

	errArray = Errors(nil)
	assert.Equal(t, 0, len(errArray), "array should be nil")

}
