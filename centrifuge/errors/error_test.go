// +build unit

package errors

import (
	"reflect"
	"testing"

	"fmt"

	"github.com/CentrifugeInc/go-centrifuge/centrifuge/code"
	"github.com/magiconair/properties/assert"
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

func TestWrap(t *testing.T) {
	// simple error
	err := fmt.Errorf("simple-error")
	err = Wrap(err, "wrapped error")
	assert.Equal(t, err.Error(), "wrapped error: simple-error")

	// p2p error
	err = New(code.Unknown, "p2p-error")
	err = Wrap(err, "wrapped error")
	assert.Equal(t, err.Error(), "[1]wrapped error: p2p-error")
}
