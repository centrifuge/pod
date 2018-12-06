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

func getInvoiceAndCheck(e *httpexpect.Expect, params map[string]interface{}) *httpexpect.Value {
	docIdentifier := params["document_id"].(string)

	objGet := e.GET("/invoice/"+docIdentifier).
		WithHeader("accept", "application/json").
		WithHeader("Content-Type", "application/json").
		Expect().Status(http.StatusOK).JSON().NotNull()
	objGet.Path("$.header.document_id").String().Equal(docIdentifier)
	objGet.Path("$.data.currency").String().Equal(params["currency"].(string))

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

func updateInvoice(e *httpexpect.Expect, docIdentifier string, payload map[string]interface{}) *httpexpect.Object {
	obj := e.PUT("/invoice/"+docIdentifier).
		WithHeader("accept", "application/json").
		WithHeader("Content-Type", "application/json").
		WithJSON(payload).
		Expect().Status(http.StatusOK).JSON().Object()
	return obj
}

func getDocumentIdentifier(t *testing.T, response *httpexpect.Object) string {
	docIdentifier := response.Value("header").Path("$.document_id").String().NotEmpty().Raw()
	if docIdentifier == "" {
		t.Error("docIdentifier empty")
	}
	return docIdentifier
}
