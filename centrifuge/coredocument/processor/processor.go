package coredocumentprocessor

import (
	"context"
	"fmt"

	"github.com/CentrifugeInc/go-centrifuge/centrifuge/anchors"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/centerrors"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/config"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/coredocument"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/coredocument/repository"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/identity"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/keytools/ed25519"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/keytools/secp256k1"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/p2p"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/signatures"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/tools"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/version"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/p2p"
	logging "github.com/ipfs/go-log"
)

var log = logging.Logger("coredocument")

// Processor identifies an implementation, which can do a bunch of things with a CoreDocument.
// E.g. send, anchor, etc.
type Processor interface {
	Send(ctx context.Context, coreDocument *coredocumentpb.CoreDocument, recipient identity.CentID) (err error)
	Anchor(ctx context.Context, document *coredocumentpb.CoreDocument,
		/* TODO remove collaborators once we have them in the document it self */
		collaborators []identity.CentID) (err error)
}

// defaultProcessor implements Processor interface
type defaultProcessor struct {
	IdentityService identity.Service
	P2PClient       p2p.Client
}

// DefaultProcessor returns the default implementation of CoreDocument Processor
func DefaultProcessor(idService identity.Service, p2pClient p2p.Client) Processor {
	return &defaultProcessor{
		IdentityService: idService,
		P2PClient:       p2pClient,
	}
}

// Send sends the given defaultProcessor to the given recipient on the P2P layer
func (dp *defaultProcessor) Send(ctx context.Context, coreDocument *coredocumentpb.CoreDocument, recipient identity.CentID) (err error) {
	if coreDocument == nil {
		return centerrors.NilError(coreDocument)
	}

	id, err := dp.IdentityService.LookupIdentityForID(recipient)
	if err != nil {
		err = centerrors.Wrap(err, "error fetching receiver identity")
		log.Error(err)
		return err
	}

	lastB58Key, err := id.GetCurrentP2PKey()
	if err != nil {
		err = centerrors.Wrap(err, "error fetching p2p key")
		log.Error(err)
		return err
	}

	log.Infof("Sending Document to CentID [%v] with Key [%v]\n", recipient, lastB58Key)
	clientWithProtocol := fmt.Sprintf("/ipfs/%s", lastB58Key)
	client, err := dp.P2PClient.OpenClient(clientWithProtocol)
	if err != nil {
		return fmt.Errorf("failed to open client: %v", err)
	}

	log.Infof("Done opening connection against [%s]\n", lastB58Key)
	hostInstance := p2p.GetHost()
	pubKey, err := hostInstance.ID().ExtractPublicKey()
	if err != nil {
		err = centerrors.Wrap(err, "failed to get pub key")
		log.Error(err)
		return err
	}

	bSenderId, err := pubKey.Bytes()
	if err != nil {
		err = centerrors.Wrap(err, "failed to extract bytes")
		log.Error(err)
		return err
	}

	header := &p2ppb.CentrifugeHeader{
		SenderCentrifugeId: bSenderId,
		CentNodeVersion:    version.GetVersion().String(),
		NetworkIdentifier:  config.Config.GetNetworkID(),
	}
	_, err = client.SendAnchoredDocument(context.Background(), &p2ppb.AnchDocumentRequest{Document: coreDocument, Header: header})
	if err != nil {
		err = centerrors.Wrap(err, "failed to post to the node")
		log.Error(err)
		return err
	}

	return nil
}

// Anchor anchors the given CoreDocument
// This method should:
// - calculate the signing root
// - sign document with own key
// - collect signatures (incl. validate)
// - store signatures on coredocument
// - calculate DocumentRoot
// - store doc in db
// - anchor the document
// - send anchored document to collaborators [NOT NEEDED since we do this in the current flow already because HandleSend****Document does it after anchoring]
func (dp *defaultProcessor) Anchor(ctx context.Context, document *coredocumentpb.CoreDocument, collaborators []identity.CentID) error {
	if document == nil {
		return centerrors.NilError(document)
	}

	anchorID, err := anchors.NewAnchorID(document.CurrentIdentifier)
	if err != nil {
		log.Error(err)
		return centerrors.Wrap(err, "anchoring error")
	}

	// calculate the signing root
	err = coredocument.CalculateSigningRoot(document)
	if err != nil {
		log.Error(err)
		return centerrors.Wrap(err, "anchoring error")
	}

	// sign document with own key and append it to signatures
	idConfig, err := ed25519.GetIDConfig()
	if err != nil {
		log.Error(err)
		return centerrors.Wrap(err, "anchoring error")
	}
	sig := signatures.Sign(idConfig, document.SigningRoot)
	document.Signatures = append(document.Signatures, sig)

	// collect signatures (incl. validate)
	// store signatures on coredocument
	err = dp.P2PClient.GetSignaturesForDocument(ctx, document, collaborators)
	if err != nil {
		log.Error(err)
		return centerrors.Wrap(err, "anchoring error")
	}

	// calculate DocumentRoot
	err = coredocument.CalculateDocumentRoot(document)
	if err != nil {
		log.Error(err)
		return centerrors.Wrap(err, "anchoring error")
	}

	// store doc in db
	err = coredocumentrepository.GetRepository().Create(document.CurrentIdentifier, document)
	if err != nil {
		log.Error(err)
		return centerrors.Wrap(err, "anchoring error")
	}

	rootHash, err := anchors.NewDocRoot(document.DocumentRoot)
	if err != nil {
		log.Error(err)
		return centerrors.Wrap(err, "anchoring error")
	}

	myCentID, err := identity.NewCentID(idConfig.ID)
	if err != nil {
		log.Error(err)
		return centerrors.Wrap(err, "anchoring error")
	}

	// generate message authentication code for the anchor call
	secpIDConfig, err := secp256k1.GetIDConfig()
	mac, err := secp256k1.SignEthereum(anchors.GenerateCommitHash(anchorID, myCentID, rootHash), secpIDConfig.PrivateKey)
	if err != nil {
		log.Error(err)
		return centerrors.Wrap(err, "anchoring error")
	}

	log.Infof("Anchoring document with identifiers: [document: %#x, current: %#x, next: %#x], rootHash: %#x", document.DocumentIdentifier, document.CurrentIdentifier, document.NextIdentifier, document.DocumentRoot)
	log.Debugf("Anchoring document with details %v", document)
	// TODO documentProofs has to be included when we develop precommit flow
	confirmations, err := anchors.CommitAnchor(anchorID, rootHash, myCentID, [][anchors.DocumentProofLength]byte{tools.RandomByte32()}, mac)
	if err != nil {
		log.Error(err)
		return centerrors.Wrap(err, "anchoring error")
	}
	<-confirmations
	return nil
}
