// +build integration unit

package testingcoredocument

import (
	"github.com/centrifuge/centrifuge-protobufs/documenttypes"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/invoice"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/golang/protobuf/proto"

	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/golang/protobuf/ptypes/any"
)

func GenerateCoreDocument() *coredocumentpb.CoreDocument {
	identifier := utils.RandomSlice(32)
	invData := &invoicepb.InvoiceData{}
	dataSalts, _ := documents.GenerateNewSalts(invData, "invoice")

	serializedInv, _ := proto.Marshal(invData)
	doc := &coredocumentpb.CoreDocument{
		DocumentIdentifier: identifier,
		CurrentVersion:     identifier,
		NextVersion:        utils.RandomSlice(32),
		EmbeddedData: &any.Any{
			TypeUrl: documenttypes.InvoiceDataTypeUrl,
			Value:   serializedInv,
		},
		EmbeddedDataSalts: documents.ConvertToProtoSalts(dataSalts),
	}
	cdSalts, _ := documents.GenerateNewSalts(doc, "")
	doc.CoredocumentSalts = documents.ConvertToProtoSalts(cdSalts)
	return doc
}
