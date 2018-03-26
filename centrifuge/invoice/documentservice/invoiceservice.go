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
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/identity"
	"github.com/go-errors/errors"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/keytools"
	"fmt"
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
		centrifugeId := string(element[:])
		peerId, err := identity.ResolveIdentityForKey(centrifugeId, 1)
		if err != nil {
			log.Printf("Error: %v\n", err)
			return nil, err
		}

		lastKey := len(peerId.Keys[1])-1
		if len(peerId.Keys[1]) == 0 {
			return nil, errors.Wrap("Identity doesn't have p2p key", 1)
		}
		// Default to last key of that type
		lastb58Key, err := keytools.PublicKeyToP2PKey(peerId.Keys[1][lastKey].Key)
		if err != nil {
			return nil, err
		}
		log.Printf("Sending Invoice to CentID [%v] with Key [%v]\n", centrifugeId, lastb58Key.Pretty())
		clientWithProtocol := fmt.Sprintf("/ipfs/%s", lastb58Key.Pretty())
		client := p2p.OpenClient(clientWithProtocol)
		log.Printf("Done opening connection against [%s]\n", peerId.Keys[1][lastKey].String())
		_, err = client.Transmit(context.Background(), &p2p.P2PMessage{&coreDoc})
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