// +build unit

package errors

import (
	"net/http"
	"testing"
)

func TestHTTPCode(t *testing.T) {
	tests := []struct {
		code Code
		want int
	}{
		{
			code: Ok,
			want: http.StatusOK,
		},

		{
			code: DocumentNotFound,
			want: http.StatusNotFound,
		},

		{
			code: Code(100),
			want: http.StatusInternalServerError,
		},
	}

	for _, c := range tests {
		if got := HTTPCode(c.code); got != c.want {
			t.Fatalf("HTTP code mismatch: %d != %d", got, c.want)
		}
	}
}

func Test_getCode(t *testing.T) {
	tests := []struct {
		code int32
		want Code
	}{
		{
			code: 0,
			want: Ok,
		},

		{
			code: 5,
			want: AuthenticationFailed,
		},

		{
			code: 10,
			want: Unknown,
		},
	}

	for _, c := range tests {
		if got := getCode(c.code); got != c.want {
			t.Fatalf("Error code mismatch: %d != %d", got, c.want)
		}
	}
}
