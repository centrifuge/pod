package invoiceservice

import (
	"golang.org/x/net/context"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/invoice"
	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/invoice"
 	google_protobuf2 "github.com/golang/protobuf/ptypes/empty"
	"github.com/spf13/viper"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/invoice/repository"
)

// Struct needed as it is used to register the grpc services attached to the grpc server
type InvoiceDocumentService struct {}

func (s *InvoiceDocumentService) SendInvoiceDocument(ctx context.Context, sendInvoiceEnvelope *invoicepb.SendInvoiceEnvelope) (*invoicepb.InvoiceDocument, error) {
	err := invoicerepository.GetInvoiceRepository().Store(sendInvoiceEnvelope.Document)
	if err != nil {
		return nil, err
	}

	inv := invoice.NewInvoice(sendInvoiceEnvelope.Document)
	inv.CalculateMerkleRoot()
	coreDoc := inv.ConvertToCoreDocument()
	// Sign document
	// Uncomment once fixed
	//coreDoc.Sign()

	if (s.IsAnchoringRequired()) {
		coreDoc.Anchor()
	}

	for _, element := range sendInvoiceEnvelope.Recipients {
		coreDoc.Send(ctx, string(element[:]))
	}
	return sendInvoiceEnvelope.Document, nil
}

func (s *InvoiceDocumentService) GetInvoiceDocument(ctx context.Context, getInvoiceDocumentEnvelope *invoicepb.GetInvoiceDocumentEnvelope) (*invoicepb.InvoiceDocument, error) {
	doc, err := invoicerepository.GetInvoiceRepository().FindById(getInvoiceDocumentEnvelope.DocumentIdentifier)
	return doc, err
}

func (s *InvoiceDocumentService) GetReceivedInvoiceDocuments (ctx context.Context, empty *google_protobuf2.Empty) (*invoicepb.ReceivedInvoices, error) {
	return nil, nil
}

/*
This function will be more complex in the future, to check if the document should be anchored or not.
*/
func (s *InvoiceDocumentService) IsAnchoringRequired() (bool) {
	return viper.GetBool("anchor.ethereum.enabled")
}