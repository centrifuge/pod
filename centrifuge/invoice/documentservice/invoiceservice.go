package documentservice

import (
	"log"
	"golang.org/x/net/context"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/p2p"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/invoice"
	cc "github.com/CentrifugeInc/go-centrifuge/centrifuge/context"
)

type InvoiceDocumentService struct {}

func (s *InvoiceDocumentService) SendInvoiceDocument(ctx context.Context, sendInvoiceEnvelope *invoice.SendInvoiceEnvelope) (*invoice.InvoiceDocument, error) {
	err := cc.Node.GetInvoiceStorageService().PutDocument(sendInvoiceEnvelope.Document)
	if err != nil {
		return nil, err
	}

	coreDoc := invoice.ConvertToCoreDocument(sendInvoiceEnvelope.Document)
	// Sign document
	signingService := cc.Node.GetSigningService()
	signingService.Sign(&coreDoc)

	for _, element := range sendInvoiceEnvelope.Recipients {
		addr := string(element[:])
		client := p2p.OpenClient(addr)
		log.Print("Done opening connection")
		_, err := client.Transmit(context.Background(), &p2p.P2PMessage{&coreDoc})
		if err != nil {
			return nil, err
		}
	}
	return sendInvoiceEnvelope.Document, nil
}

func (s *InvoiceDocumentService) GetInvoiceDocument(ctx context.Context, getInvoiceDocumentEnvelope *invoice.GetInvoiceDocumentEnvelope) (*invoice.InvoiceDocument, error) {
	doc, err := cc.Node.GetInvoiceStorageService().GetDocument(getInvoiceDocumentEnvelope.DocumentIdentifier)
	return doc, err
}

func (s *InvoiceDocumentService) GetReceivedInvoiceDocuments (ctx context.Context, empty *invoice.Empty) (*invoice.ReceivedInvoices, error) {
	doc, err := cc.Node.GetInvoiceStorageService().GetReceivedDocuments()
	return doc, err
}