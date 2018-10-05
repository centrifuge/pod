// +build unit

package centerrors

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/centrifuge/go-centrifuge/centrifuge/code"
)

func TestError(t *testing.T) {
	tests := []struct {
		code   code.Code
		msg    string
		errors map[string]string
	}{
		{
			code: code.AuthenticationFailed,
			msg:  "Node authentication failed",
		},

		{
			code: code.DocumentNotFound,
			msg:  "Invalid document",
			errors: map[string]string{
				"document_root":   "root empty",
				"next_identifier": "invalid identifier",
			},
		},

		{
			code: code.Code(100),
			msg:  "Unknown error",
		},
	}

	for _, c := range tests {
		err := NewWithErrors(c.code, c.msg, c.errors)
		p2perr, ok := FromError(err)
		if !ok {
			t.Fatalf("unexpected conversion error: %T", err)
		}

		if got := p2perr.Message(); got != c.msg {
			t.Fatalf("message mismatch: %s != %s", got, c.msg)
		}

		if got := p2perr.Errors(); !reflect.DeepEqual(got, c.errors) {
			t.Fatalf("errors mismatch: %v != %v", got, c.errors)
		}

		want := code.To(int32(c.code))

		if got := p2perr.Code(); got != want {
			t.Fatalf("code mismatch: %v != %v", got, want)
		}
	}
}

func TestWrap(t *testing.T) {

	simpleErr := fmt.Errorf("simple-error 1")
	simpleErr2 := fmt.Errorf("simple-error 2")

	//case: error & error
	errors := Wrap(simpleErr, simpleErr2)
	centError, ok := FromError(errors)
	assert.False(t, ok, "error is not a cent error msg")

	assert.True(t, len(centError.Message()) >= len(simpleErr.Error())+len(simpleErr2.Error()), "error msg should contain both error msg's")
	assert.Equal(t, 0, len(centError.Errors()), "error map should be empty")
	assert.Equal(t, code.Unknown, centError.Code(), "code should be from type unkown")

	//case: error & centerror
	error3 := New(code.DocumentInvalid, "test document invalid")
	errors = Wrap(errors, error3)
	centError, ok = FromError(errors)
	assert.True(t, ok, "transformation to Error should work")
	assert.Equal(t, centError.Code(), code.DocumentInvalid, "code should be copied from srcError ")

	//case: centerror & error
	errors = Wrap(errors, fmt.Errorf("simple-error 4"))
	centError, ok = FromError(errors)
	assert.True(t, ok, "transformation to Error should work")
	assert.Equal(t, centError.Code(), code.DocumentInvalid, "code should be copied from srcError ")

	//case: centerror (no map) & centerror (no map)
	error5 := New(code.AuthenticationFailed, "test auth failed")
	errors = Wrap(errors, error5)
	centError, ok = FromError(errors)
	assert.True(t, ok, "transformation to Error should work")
	assert.Equal(t, centError.Code(), code.DocumentInvalid, "code should not be changed ")

}

func TestWrapErrorWithMap(t *testing.T) {

	//case: no map & map
	errorDst := New(code.DocumentInvalid, "test document invalid")
	errorMap := make(map[string]string)
	errorMap["error1"] = "first error"
	errorMap["error2"] = "second error"

	errorSrc := NewWithErrors(code.DocumentInvalid, "test msg", errorMap)
	errorDst = Wrap(errorDst, errorSrc)
	centError, ok := FromError(errorDst)
	assert.True(t, ok, "transformation to Error should work")
	assert.Equal(t, 2, len(centError.Errors()), "map should contain 2 entries")

	//case: map & no map
	error2 := New(code.AuthenticationFailed, "test auth failed")
	errorDst = Wrap(errorDst, error2)
	centError, ok = FromError(errorDst)
	assert.True(t, ok, "transformation to Error should work")
	assert.Equal(t, 2, len(centError.Errors()), "map should contain 2 entries")

	//case: map & map
	errorMap2 := make(map[string]string)
	errorMap2["error1"] = "third error" //same id
	errorMap2["error4"] = "fourth error"
	errorSrc = NewWithErrors(code.DocumentInvalid, "test msg", errorMap2)
	errorDst = Wrap(errorDst, errorSrc)
	centError, ok = FromError(errorDst)
	assert.True(t, ok, "transformation to Error should work")
	assert.Equal(t, 4, len(centError.Errors()), "map should contain 4 entries")

}

func TestWrapErrorWithMsg(t *testing.T) {

	errorMsg1 := "test msg error 1"
	errorMsg2 := "test msg error 2"

	err := fmt.Errorf(errorMsg1)
	err = Wrap(err, errorMsg2)

	assert.Error(t, err, "wrap should return an error")
	assert.True(t, len(err.Error()) >= (len(errorMsg1)+len(errorMsg2)), "wrap should append the src error msg")

	err = Wrap(nil, errorMsg1)
	assert.Error(t, err, "should return an error")
	assert.Equal(t, errorMsg1, err.Error(), "error should include error msg")

}
