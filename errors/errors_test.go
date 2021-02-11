// +build unit

package errors

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestNew(t *testing.T) {
	tests := []struct {
		msg    string
		args   []interface{}
		result string
	}{
		// empty message
		{},

		// simple message
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
	_, ok := lerr.(listError)
	assert.True(t, ok)
	assert.Equal(t, len, Len(lerr))
	assert.Equal(t, result, lerr.Error())
}

func TestAppendError(t *testing.T) {
	// both nil
	assert.Nil(t, AppendError(nil, nil))

	// errn nil, and err not nil but simple error
	serr := New("some error")
	lerr := AppendError(serr, nil)
	checkListError(t, lerr, 1, "[some error]")

	// errn nil, and err not nil and a list error
	lerr = AppendError(lerr, nil)
	checkListError(t, lerr, 1, "[some error]")

	// err nil and errn not nil and simple error
	lerr = AppendError(nil, serr)
	checkListError(t, lerr, 1, "[some error]")

	// err nil and errn not nil and list error
	lerr = AppendError(nil, lerr)
	checkListError(t, lerr, 1, "[some error]")

	// both simple errors
	lerr = AppendError(serr, serr)
	checkListError(t, lerr, 2, "[some error; some error]")

	// err simple and errn list
	lerr = AppendError(serr, lerr)
	checkListError(t, lerr, 3, "[some error; some error; some error]")

	// err list error and errn simple error
	lerr = AppendError(lerr, serr)
	checkListError(t, lerr, 4, "[some error; some error; some error; some error]")

	// both list errors
	lerr = AppendError(AppendError(serr, nil), AppendError(serr, nil))
	checkListError(t, lerr, 2, "[some error; some error]")
}

func TestIsOfType(t *testing.T) {
	// single type error
	const errBadErr = Error("bad error")
	serr := New("some error")
	terr := NewTypedError(ErrUnknown, serr)
	assert.True(t, IsOfType(ErrUnknown, terr))
	assert.False(t, IsOfType(errBadErr, terr))
	assert.Equal(t, "unknown error: some error", terr.Error())

	// recursive error
	terr = NewTypedError(errBadErr, terr)
	assert.True(t, IsOfType(ErrUnknown, terr))
	assert.True(t, IsOfType(errBadErr, terr))
	assert.Equal(t, "bad error: unknown error: some error", terr.Error())

	// simple error
	assert.False(t, IsOfType(ErrUnknown, serr))
	assert.False(t, IsOfType(errBadErr, serr))
	assert.True(t, IsOfType(serr, serr))

	// list error
	lerr := AppendError(serr, nil)
	assert.False(t, IsOfType(ErrUnknown, lerr))
	assert.False(t, IsOfType(errBadErr, lerr))
	terr = NewTypedError(errBadErr, lerr)
	assert.True(t, IsOfType(errBadErr, terr))

	// status err
	serr = status.Error(codes.Unknown, errBadErr.Error())
	assert.True(t, IsOfType(errBadErr, serr))
}

func TestMask(t *testing.T) {
	errBadErr := Error("bad error")
	terr := NewTypedError(ErrUnknown, errBadErr)
	s := Mask(terr)
	assert.Equal(t, "error has been masked", s.Error())

	s = Mask(errBadErr)
	assert.Equal(t, "error has been masked", s.Error())
}

func TestGetHTTPCode(t *testing.T) {
	err := New("some error")
	code, msg := GetHTTPDetails(err)
	assert.Equal(t, http.StatusInternalServerError, code)
	assert.Equal(t, "some error", msg)

	err = NewHTTPError(http.StatusBadRequest, err)
	code, msg = GetHTTPDetails(err)
	assert.Equal(t, http.StatusBadRequest, code)
	assert.Equal(t, "some error", msg)

	err = NewHTTPError(int(codes.AlreadyExists), New("some error"))
	code, msg = GetHTTPDetails(err)
	assert.Equal(t, http.StatusConflict, code)
	assert.Equal(t, "some error", msg)
}

func checkStackTrace(t *testing.T, err error, trace string) {
	sterr, ok := err.(*withStack)
	assert.True(t, ok)
	werr := detachStackTrace(err)
	assert.Equal(t, sterr.err.Error(), werr.Error())
	st := StackTrace(err)
	assert.NotEmpty(t, st)
	assert.Contains(t, st, trace)
	st = StackTrace(werr)
	assert.Empty(t, st)
}

func TestNewWithStackTrace(t *testing.T) {
	err := New("some error")
	checkStackTrace(t, err, "github.com/centrifuge/go-centrifuge/errors.TestNewWithStackTrace")
}

func TestWithStackTrace(t *testing.T) {
	err := WithStackTrace(New("some error"))
	checkStackTrace(t, err, "github.com/centrifuge/go-centrifuge/errors.TestWithStackTrace")
}
