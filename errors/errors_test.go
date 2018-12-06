// +build unit

package errors

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	tests := []struct {
		msg    string
		args   []interface{}
		result string
	}{
		// empty message
		{},

		//// simple message
		{
			msg:    "some error",
			result: "some error",
		},

		// format error
		{
			msg:    "some error: %v: %v",
			args:   []interface{}{"error1", "error2"},
			result: "some error: error1: error2",
		},
	}

	for _, c := range tests {
		err := New(c.msg, c.args...)
		assert.Equal(t, c.result, err.Error(), "must match")
	}
}

func checkListError(t *testing.T, lerr error, len int, result string) {
	assert.NotNil(t, lerr)
	_, ok := lerr.(*listError)
	assert.True(t, ok)
	assert.Equal(t, len, Len(lerr))
	assert.Equal(t, result, lerr.Error())
}

func TestNewListError(t *testing.T) {
	// both nil
	assert.Nil(t, NewListError(nil, nil))

	// errn nil, and err not nil but simple error
	serr := errors.New("some error")
	lerr := NewListError(serr, nil)
	checkListError(t, lerr, 1, "[some error]")

	// errn nil, and err not nil and a list error
	lerr = NewListError(lerr, nil)
	checkListError(t, lerr, 1, "[some error]")

	// err nil and errn not nil and simple error
	lerr = NewListError(nil, serr)
	checkListError(t, lerr, 1, "[some error]")

	// err nil and errn not nil and list error
	lerr = NewListError(nil, lerr)
	checkListError(t, lerr, 1, "[some error]")

	// both simple errors
	lerr = NewListError(serr, serr)
	checkListError(t, lerr, 2, "[some error; some error]")

	// err simple and errn list
	lerr = NewListError(serr, lerr)
	checkListError(t, lerr, 3, "[some error; some error; some error]")

	// err list error and errn simple error
	lerr = NewListError(lerr, serr)
	checkListError(t, lerr, 4, "[some error; some error; some error; some error]")

	// both list errors
	lerr = NewListError(NewListError(serr, nil), NewListError(serr, nil))
	checkListError(t, lerr, 2, "[some error; some error]")

}
