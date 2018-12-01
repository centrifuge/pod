package testworld

import (
	"crypto/tls"
	"net/http"
	"testing"
	"time"

	"github.com/gavv/httpexpect"
)

func createInsecureClient(t *testing.T, baseURL string) *httpexpect.Expect {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	config := httpexpect.Config{
		BaseURL: baseURL,
		Client: &http.Client{
			Transport: transport,
			Timeout:   time.Second * 600,
		},
		Reporter: httpexpect.NewAssertReporter(t),
		Printers: []httpexpect.Printer{
			httpexpect.NewCompactPrinter(t),
		},
	}
	return httpexpect.WithConfig(config)
}

func getInvoiceAndCheck(e *httpexpect.Expect, docIdentifier string, currency string) *httpexpect.Value {
	objGet := e.GET("/invoice/"+docIdentifier).
		WithHeader("accept", "application/json").
		WithHeader("Content-Type", "application/json").
		Expect().Status(http.StatusOK).JSON().NotNull()
	objGet.Path("$.header.document_id").String().Equal(docIdentifier)
	objGet.Path("$.data.currency").String().Equal(currency)
	return objGet
}

func createInvoice(e *httpexpect.Expect, payload map[string]interface{}) *httpexpect.Object {
	obj := e.POST("/invoice").
		WithHeader("accept", "application/json").
		WithHeader("Content-Type", "application/json").
		WithJSON(payload).
		Expect().Status(http.StatusOK).JSON().Object()
	return obj
}
