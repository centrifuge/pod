//go:build testworld

package expect

import (
	"crypto/tls"
	"net/http"
	"testing"

	"github.com/gavv/httpexpect"
)

type httpLog struct {
	httpexpect.Logger
}

func CreateInsecureClientWithExpect(t *testing.T, baseURL string) *httpexpect.Expect {
	config := httpexpect.Config{
		BaseURL:  baseURL,
		Client:   createInsecureClient(),
		Reporter: httpexpect.NewAssertReporter(t),
		Printers: []httpexpect.Printer{
			httpexpect.NewCompactPrinter(&httpLog{t}),
		},
	}

	return httpexpect.WithConfig(config)
}

func createInsecureClient() *http.Client {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	return &http.Client{Transport: tr}
}
