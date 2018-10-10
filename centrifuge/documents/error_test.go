// +build unit

package documents

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewError(t *testing.T) {

	err := NewError("test_error", "error msg")
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

func TestConvertToMap(t *testing.T) {

	err := NewError("test1", "error msg")

	err = AppendError(err, NewError("test2", "error msg2"))

	errMap := ConvertToMap(err)
	assert.Equal(t, 2, len(errMap), "map should have two entries")

	err = AppendError(err, fmt.Errorf("standard error"))
	err = AppendError(err, fmt.Errorf("standard error2"))
	errMap = ConvertToMap(err)
	assert.Equal(t, 4, len(errMap), "map should have 4 entries")

	assert.NotEqual(t, "", errMap["error_1"], "first standard error should have id 'error_1'")
	assert.Equal(t, "", errMap["error_3"], "no standard error with id 'error_3'")

	errMap = ConvertToMap(nil)
	assert.Equal(t, 0, len(errMap), "map should be empty")

	fmt.Print(errMap)

}
