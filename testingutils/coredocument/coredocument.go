// +build integration unit

package testingcoredocument

import (
	"github.com/centrifuge/centrifuge-protobufs/documenttypes"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/invoice"
	"github.com/golang/protobuf/proto"

	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/precise-proofs/proofs"
	"github.com/golang/protobuf/ptypes/any"
)

func GenerateCoreDocument() *coredocumentpb.CoreDocument {
	identifier := utils.RandomSlice(32)
	dataSalts := &invoicepb.InvoiceDataSalts{}
	invData := &invoicepb.InvoiceData{}
	proofs.FillSalts(invData, dataSalts)

	serializedInv, _ := proto.Marshal(invData)
	serializedInvSalts, _ := proto.Marshal(dataSalts)
	salts := &coredocumentpb.CoreDocumentSalts{}
	doc := &coredocumentpb.CoreDocument{
		DocumentIdentifier: identifier,
		CurrentVersion:     identifier,
		NextVersion:        utils.RandomSlice(32),
		CoredocumentSalts:  salts,
		EmbeddedData: &any.Any{
			TypeUrl: documenttypes.InvoiceDataTypeUrl,
			Value:   serializedInv,
		},
		EmbeddedDataSalts: &any.Any{
			TypeUrl: documenttypes.InvoiceSaltsTypeUrl,
			Value:   serializedInvSalts,
		},
	}
	proofs.FillSalts(doc, salts)
	return doc
}
