// +build testworld

package testworld

func defaultInvoicePayload(collaborators []string) map[string]interface{} {

	return map[string]interface{}{
		"data": map[string]interface{}{
			"invoice_number": "12324",
			"due_date":       "2018-09-26T23:12:37.902198664Z",
			"gross_amount":   "40",
			"currency":       "USD",
			"net_amount":     "40",
		},
		"collaborators": collaborators,
	}

}

func defaultNFTPayload(collaborators []string) map[string]interface{} {

	return map[string]interface{}{
		"data": map[string]interface{}{
			"invoice_number": "12324",
			"due_date":       "2018-09-26T23:12:37.902198664Z",
			"gross_amount":   "40",
			"currency":       "USD",
			"net_amount":     "40",
			"document_type":  "invoice",
		},
		"collaborators": collaborators,
	}

}

func updatedInvoicePayload(collaborators []string) map[string]interface{} {
	return map[string]interface{}{
		"data": map[string]interface{}{
			"invoice_number": "12324",
			"due_date":       "2018-09-26T23:12:37.902198664Z",
			"gross_amount":   "40",
			"currency":       "EUR",
			"net_amount":     "42",
		},
		"collaborators": collaborators,
	}

}

func defaultProofPayload() map[string]interface{} {
	return map[string]interface{}{
		"type":   "http://github.com/centrifuge/centrifuge-protobufs/invoice/#invoice.InvoiceData",
		"fields": []string{"invoice.net_amount", "invoice.currency"},
	}
}
