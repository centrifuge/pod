// nolint
package userapi

import (
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/extensions/transferdetails"
	"github.com/centrifuge/go-centrifuge/httpapi/coreapi"
)

// TransferDetailData is the default transfer details extension schema
//type TransferDetailData struct {
//	TransferId          string `json:"transfer_id,omitempty" attr:"bytes"`
//	SenderId            string `json:"sender_id,omitempty" attr:"bytes"`
//	RecipientId         string `json:"recipient_id,omitempty" attr:"bytes"`
//	ScheduledDate       string `json:"scheduled_date,omitempty" attr:"timestamp"`
//	SettlementDate      string `json:"settlement_date,omitempty" attr:"timestamp"`
//	SettlementReference string `json:"settlement_reference,omitempty" attr:"timestamp"`
//	Amount              string `json:"amount,omitempty" attr:"decimal"`
//	// the currency and amount will be combined once we have standardised multiformats
//	Currency     string `json:"currency,omitempty" attr:"string"`
//	Status       string `json:"status,omitempty" attr:"string"`
//	TransferType string `json:"transfer_type,omitempty" attr:"string"`
//	Data         string `json:"data,omitempty" attr:"string"`
//}

// TODO: think: generic custom attribute set creation?

//CreateTransferDetailRequest
type CreateTransferDetailRequest struct {
	DocumentID string                              `json:"document_id" swaggertype:"primitive,string"`
	Data       *transferdetails.TransferDetailData `json:"data"`
}

// UpdateTransferDetailRequest
type UpdateTransferDetailRequest struct {
	DocumentID string                              `json:"document_id" swaggertype:"primitive,string"`
	TransferID string                              `json:"transfer_id" swaggertype:"primitive,string"`
	Data       *transferdetails.TransferDetailData `json:"data"`
}

// TransferDetailResponse
type TransferDetailResponse struct {
	Header *coreapi.ResponseHeader             `json:"header"`
	Data   *transferdetails.TransferDetailData `json:"data"`
}

// TransferDetailListResponse
type TransferDetailListResponse struct {
	Header *coreapi.ResponseHeader             `json:"header"`
	Data   *transferdetails.TransferDetailData `json:"data"`
}

func toTransferDetailCreatePayload(request CreateTransferDetailRequest) (*transferdetails.CreateTransferDetailRequest, error) {
	payload := new(transferdetails.CreateTransferDetailRequest)
	payload.Data = request.Data
	payload.DocumentID = request.DocumentID

	return payload, nil
}

func toClientAttributes(attrs []documents.Attribute) (map[documents.AttrKey]documents.Attribute, error) {
	if len(attrs) < 1 {
		return nil, nil
	}

	m := make(map[documents.AttrKey]documents.Attribute)
	for _, v := range attrs {
		m[v.Key] = documents.Attribute{
			KeyLabel: v.Key.String(),
			Key:      v.Key,
			Value:    v.Value,
		}
	}

	return m, nil
}
