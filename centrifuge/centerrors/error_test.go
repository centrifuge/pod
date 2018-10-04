// +build unit

package centerrors

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/centrifuge/go-centrifuge/centrifuge/code"
)

func TestP2PError(t *testing.T) {
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

func TestWrapErrors(t *testing.T) {

	simpleErr := fmt.Errorf("simple-error 1")
	simpleErr2 := fmt.Errorf("simple-error 2")

	errors := WrapErrors(simpleErr, simpleErr2)
	centError, ok := FromError(errors)
	assert.True(t, ok, "transformation to Error should work")
	assert.Equal(t, 2, len(centError.Errors()), "error map should contain two errors")
	assert.Equal(t, code.Unknown, centError.Code(), "code should be from type unkown")

	simpleErr3 := fmt.Errorf("simple-error 3")
	errors = WrapErrors(errors, simpleErr3)
	centError, ok = FromError(errors)
	assert.True(t, ok, "transformation to Error should work")
	assert.Equal(t, 3, len(centError.Errors()), "error map should contain two errors")

	errorMap := make(map[string]string)
	errorMap["invalidInvoice"] = "test invalid invoice"
	errorMap["test"] = "another test error"

	errors = WrapErrors(errors, NewWithErrors(code.DocumentInvalid, "invalid document", errorMap))
	centError, ok = FromError(errors)

	assert.True(t, ok, "transformation to Error should work")
	assert.Equal(t, 5, len(centError.Errors()), "error map should contain two errors")
	assert.Equal(t, code.DocumentInvalid, centError.Code(), "code should be from type unkown")

}
