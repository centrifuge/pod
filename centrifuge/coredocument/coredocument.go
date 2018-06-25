package coredocument

import (
	"context"
	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/coredocument"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/identity"
	"log"
	"fmt"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/p2p"
	"github.com/go-errors/errors"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/anchor"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/tools"
	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/p2p"
)

type CoreDocument struct {
	Document *coredocumentpb.CoreDocument
}

// Sender identifies someone who can send a given CoreDocument via the Context to the recipient
type Sender interface {
	Send(cd *CoreDocument, ctx context.Context, recipient string) (err error)
}

// SendP2P is a sender specifcially for the peer to peer layer
type SendP2P struct{}


func NewCoreDocument(document *coredocumentpb.CoreDocument) (*CoreDocument) {
	return &CoreDocument{Document: document}
}

// Send sends the given CoreDocument to the given recipient on the P2P layer
func (sendP2P SendP2P) Send(cd *CoreDocument, ctx context.Context, recipient string) (err error) {
	peerId, err := identity.ResolveP2PEthereumIdentityForId(recipient)
	if err != nil {
		log.Printf("Error: %v\n", err)
		return err
	}

	if len(peerId.Keys[1]) == 0 {
		return errors.Wrap("Identity doesn't have p2p key", 1)
	}

	// Default to last key of that type
	lastb58Key, err := peerId.GetLastB58KeyForType(1)
	if err != nil {
		return err
	}
	log.Printf("Sending Document to CentID [%v] with Key [%v]\n", recipient, lastb58Key)
	clientWithProtocol := fmt.Sprintf("/ipfs/%s", lastb58Key)
	client := p2p.OpenClient(clientWithProtocol)
	log.Printf("Done opening connection against [%s]\n", lastb58Key)
	_, err = client.Post(context.Background(), &p2ppb.P2PMessage{Document: cd.Document})
	if err != nil {
		return err
	}
	return
}

func (cd *CoreDocument) Anchor() (err error) {
	//Remove this as soon as signing is fixed, we will read from the CoreDocument signature fields
	id := tools.RandomString32()
	rootHash := tools.RandomString32()
	//
	confirmations := make(chan *anchor.WatchAnchor, 1)
	err = anchor.RegisterAsAnchor(id, rootHash, confirmations)
	if err != nil {
		return err
	}
	anchorWatch := <-confirmations
	err = anchorWatch.Error
	return
}

func (cd *CoreDocument) Sign() {
	//signingService := cc.Node.GetSigningService()
	//signingService.Sign(cd.Document)
	return
}
