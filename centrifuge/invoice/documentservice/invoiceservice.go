package documentservice

import (
	"log"
	"golang.org/x/net/context"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/p2p"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/invoice"
	cc "github.com/CentrifugeInc/go-centrifuge/centrifuge/context"
 	google_protobuf2 "github.com/golang/protobuf/ptypes/empty"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/anchor"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/tools"
	"github.com/spf13/viper"
)


type InvoiceDocumentService struct {}

func (s *InvoiceDocumentService) SendInvoiceDocument(ctx context.Context, sendInvoiceEnvelope *invoice.SendInvoiceEnvelope) (*invoice.InvoiceDocument, error) {
	err := cc.Node.GetInvoiceStorageService().PutDocument(sendInvoiceEnvelope.Document)
	if err != nil {
		return nil, err
	}

	coreDoc := invoice.ConvertToCoreDocument(sendInvoiceEnvelope.Document)
	// Sign document
	// Uncomment once fixed
	//signingService := cc.Node.GetSigningService()
	//signingService.Sign(&coreDoc)

	// Anchor Document if configure to do so - temp approach
	if (viper.GetBool("anchor.ethereum.enabled")) {
		confirmations := make(chan *anchor.Anchor, 1)
		id := tools.RandomString32()
		rootHash := tools.RandomString32()
		anchor.RegisterAsAnchor(id, rootHash, confirmations)
		_ = <-confirmations
	}

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

func (s *InvoiceDocumentService) GetReceivedInvoiceDocuments (ctx context.Context, empty *google_protobuf2.Empty) (*invoice.ReceivedInvoices, error) {
	doc, err := cc.Node.GetInvoiceStorageService().GetReceivedDocuments()
	return doc, err
}