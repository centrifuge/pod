package testworld

import (
	"crypto/tls"
	"net/http"
	"testing"

	"github.com/gavv/httpexpect"
)

const typeInvoice string = "invoice"
const typePO string = "purchaseorder"
const poPrefix string = "po"

func createInsecureClientWithExpect(t *testing.T, baseURL string) *httpexpect.Expect {
	config := httpexpect.Config{
		BaseURL:  baseURL,
		Client:   createInsecureClient(),
		Reporter: httpexpect.NewAssertReporter(t),
		Printers: []httpexpect.Printer{
			httpexpect.NewCompactPrinter(t),
		},
	}
	return httpexpect.WithConfig(config)
}

func getDocumentAndCheck(e *httpexpect.Expect, documentType string, params map[string]interface{}) *httpexpect.Value {
	docIdentifier := params["document_id"].(string)

	objGet := e.GET("/"+documentType+"/"+docIdentifier).
		WithHeader("accept", "application/json").
		WithHeader("Content-Type", "application/json").
		Expect().Status(http.StatusOK).JSON().NotNull()
	objGet.Path("$.header.document_id").String().Equal(docIdentifier)
	objGet.Path("$.data.currency").String().Equal(params["currency"].(string))

	return objGet
}

func createDocument(e *httpexpect.Expect, documentType string, status int, payload map[string]interface{}) *httpexpect.Object {
	obj := e.POST("/"+documentType).
		WithHeader("accept", "application/json").
		WithHeader("Content-Type", "application/json").
		WithJSON(payload).
		Expect().Status(status).JSON().Object()
	return obj
}

func updateDocument(e *httpexpect.Expect, documentType string, status int, docIdentifier string, payload map[string]interface{}) *httpexpect.Object {
	obj := e.PUT("/"+documentType+"/"+docIdentifier).
		WithHeader("accept", "application/json").
		WithHeader("Content-Type", "application/json").
		WithJSON(payload).
		Expect().Status(status).JSON().Object()
	return obj
}

func getDocumentIdentifier(t *testing.T, response *httpexpect.Object) string {
	docIdentifier := response.Value("header").Path("$.document_id").String().NotEmpty().Raw()
	if docIdentifier == "" {
		t.Error("docIdentifier empty")
	}
	return docIdentifier
}

func mintNFT(e *httpexpect.Expect, httpStatus int, payload map[string]interface{}) *httpexpect.Object {
	resp := e.POST("/token/mint").
		WithHeader("accept", "application/json").
		WithHeader("Content-Type", "application/json").
		WithJSON(payload).
		Expect().Status(httpStatus)

	httpObj := resp.JSON().Object()
	return httpObj
}

func getProof(e *httpexpect.Expect, httpStatus int, documentID string, payload map[string]interface{}) *httpexpect.Object {
	resp := e.POST("/document/"+documentID+"/proof").
		WithHeader("accept", "application/json").
		WithHeader("Content-Type", "application/json").
		WithJSON(payload).
		Expect().Status(httpStatus)
	return resp.JSON().Object()
}

func getNodeConfig(e *httpexpect.Expect, httpStatus int) *httpexpect.Object {
	resp := e.GET("/config/node").
		WithHeader("accept", "application/json").
		WithHeader("Content-Type", "application/json").
		Expect().Status(httpStatus)
	return resp.JSON().Object()
}

func getTenantConfig(e *httpexpect.Expect, httpStatus int, identifier string) *httpexpect.Object {
	resp := e.GET("/config/tenant/"+identifier).
		WithHeader("accept", "application/json").
		WithHeader("Content-Type", "application/json").
		Expect().Status(httpStatus)
	return resp.JSON().Object()
}

func getAllTenantConfigs(e *httpexpect.Expect, httpStatus int) *httpexpect.Object {
	resp := e.GET("/config/tenant").
		WithHeader("accept", "application/json").
		WithHeader("Content-Type", "application/json").
		Expect().Status(httpStatus)
	return resp.JSON().Object()
}

// TODO add rest of the endpoints for config

func createInsecureClient() *http.Client {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	return &http.Client{Transport: tr}
}
