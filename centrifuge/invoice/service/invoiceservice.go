package invoiceservice

import (
	"golang.org/x/net/context"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/invoice"
	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/invoice"
 	google_protobuf2 "github.com/golang/protobuf/ptypes/empty"
	"github.com/spf13/viper"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/invoice/repository"
	"github.com/go-errors/errors"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/coredocument/repository"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/coredocument"
)

// Struct needed as it is used to register the grpc services attached to the grpc server
type InvoiceDocumentService struct {}

func (s *InvoiceDocumentService) HandleSendInvoiceDocument(ctx context.Context, sendInvoiceEnvelope *invoicepb.SendInvoiceEnvelope) (*invoicepb.InvoiceDocument, error) {
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

	errs := []error{}
	for _, element := range sendInvoiceEnvelope.Recipients {
		err1 := coreDoc.Send(ctx, string(element[:]))
		if err1 != nil {
			errs = append(errs, err1)
		}
	}

	if len(errs) != 0 {
		return nil, errors.Errorf("%v", errs)
	}
	return sendInvoiceEnvelope.Document, nil
}

func (s *InvoiceDocumentService) HandleGetInvoiceDocument(ctx context.Context, getInvoiceDocumentEnvelope *invoicepb.GetInvoiceDocumentEnvelope) (*invoicepb.InvoiceDocument, error) {
	doc, err := invoicerepository.GetInvoiceRepository().FindById(getInvoiceDocumentEnvelope.DocumentIdentifier)
	if err != nil {
		doc1, err1 := coredocumentrepository.GetCoreDocumentRepository().FindById(getInvoiceDocumentEnvelope.DocumentIdentifier)
		if err1 == nil {
			doc = invoice.NewInvoiceFromCoreDocument(&coredocument.CoreDocument{doc1}).Document
			err = err1
		}
	}
	return doc, err
}

func (s *InvoiceDocumentService) HandleGetReceivedInvoiceDocuments (ctx context.Context, empty *google_protobuf2.Empty) (*invoicepb.ReceivedInvoices, error) {
	return nil, nil
}

/*
This function will be more complex in the future, to check if the document should be anchored or not.
*/
func (s *InvoiceDocumentService) IsAnchoringRequired() (bool) {
	return viper.GetBool("anchor.ethereum.enabled")
}