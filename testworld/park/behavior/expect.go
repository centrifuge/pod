//go:build testworld

package behavior

import (
	"crypto/tls"
	"net/http"
	"os"
	"testing"

	"github.com/gavv/httpexpect"
)

// TODO(cdamian) is this still relevant?
var isRunningOnCI = len(os.Getenv("TRAVIS")) != 0

type httpLog struct {
	logger httpexpect.Logger
}

func (h *httpLog) Logf(fm string, args ...interface{}) {
	if !isRunningOnCI {
		h.logger.Logf(fm, args...)
	}
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
