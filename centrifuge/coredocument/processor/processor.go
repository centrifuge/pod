package coredocumentprocessor

import (
	"context"
	"crypto/sha256"
	"fmt"

	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/coredocument"
	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/p2p"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/anchor"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/code"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/coredocument"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/errors"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/identity"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/p2p"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/signatures"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/tools"
	"github.com/centrifuge/precise-proofs/proofs"
	logging "github.com/ipfs/go-log"
)

var log = logging.Logger("coredocument")

// Processor identifies an implementation, which can do a bunch of things with a CoreDocument.
// E.g. send, anchor, etc.
type Processor interface {
	Send(coreDocument *coredocumentpb.CoreDocument, ctx context.Context, recipient []byte) (err error)
	Anchor(document *coredocumentpb.CoreDocument) (err error)
}

// defaultProcessor implements Processor interface
type defaultProcessor struct {
	IdentityService identity.IdentityService
}

// NewDefaultProcessor returns the default implementation of CoreDocument Processor
func NewDefaultProcessor() Processor {
	return &defaultProcessor{
		IdentityService: identity.NewEthereumIdentityService()}
}

// Send sends the given defaultProcessor to the given recipient on the P2P layer
func (dp *defaultProcessor) Send(coreDocument *coredocumentpb.CoreDocument, ctx context.Context, recipient []byte) (err error) {
	if coreDocument == nil {
		return errors.NilError(coreDocument)
	}

	id, err := dp.IdentityService.LookupIdentityForId(recipient)
	if err != nil {
		log.Errorf("error fetching receiver identity: %v\n", err)
		return err
	}

	lastB58Key, err := id.GetCurrentP2PKey()
	if err != nil {
		log.Errorf("error fetching p2p key: %v\n", err)
		return err
	}

	log.Infof("Sending Document to CentID [%v] with Key [%v]\n", recipient, lastB58Key)
	clientWithProtocol := fmt.Sprintf("/ipfs/%s", lastB58Key)
	client := p2p.OpenClient(clientWithProtocol)
	log.Infof("Done opening connection against [%s]\n", lastB58Key)

	hostInstance := p2p.GetHost()
	bSenderId, err := hostInstance.ID().ExtractPublicKey().Bytes()
	if err != nil {
		return fmt.Errorf("failed to extract pub key: %v", err)
	}

	_, err = client.Post(context.Background(), &p2ppb.P2PMessage{Document: coreDocument, SenderCentrifugeId: bSenderId})
	if err != nil {
		// this is p2pError, lets not format it
		return err
	}

	return nil
}

// Anchor anchors the given CoreDocument
func (dp *defaultProcessor) Anchor(document *coredocumentpb.CoreDocument) error {
	if document == nil {
		return errors.NilError(document)
	}

	log.Infof("Anchoring document with identifiers: [document: %#x, current: %#x, next: %#x], rootHash: %#x", document.DocumentIdentifier, document.CurrentIdentifier, document.NextIdentifier, document.DocumentRoot)
	log.Debugf("Anchoring document with details %v", document)

	id, err := tools.SliceToByte32(document.CurrentIdentifier)
	if err != nil {
		log.Error(err)
		return err
	}

	// TODO: we should replace this with using the DocumentRoot once signing has been properly implemented
	rootHash, err := tools.SliceToByte32(document.DataRoot)
	if err != nil {
		log.Error(err)
		return err
	}

	confirmations, err := anchor.RegisterAsAnchor(id, rootHash)
	if err != nil {
		log.Error(err)
		return err
	}

	anchorWatch := <-confirmations
	return anchorWatch.Error
}

func (dp *defaultProcessor) getDocumentTree(document *coredocumentpb.CoreDocument) (tree *proofs.DocumentTree, err error) {
	t := proofs.NewDocumentTree(proofs.TreeOptions{Hash: sha256.New()})
	tree = &t
	err = tree.AddLeavesFromDocument(document, document.CoredocumentSalts)
	if err != nil {
		return nil, err
	}
	err = tree.Generate()
	if err != nil {
		return nil, err
	}
	return tree, nil
}

func (dp *defaultProcessor) calculateSigningRoot(document *coredocumentpb.CoreDocument) error {
	valid, errMsg, errs := coredocument.Validate(document)
	if !valid {
		return errors.NewWithErrors(code.DocumentInvalid, errMsg, errs)
	}

	tree, err := dp.getDocumentTree(document)
	if err != nil {
		return err
	}
	document.SigningRoot = tree.RootHash()
	return nil
}

func (dp *defaultProcessor) Sign(document *coredocumentpb.CoreDocument) (err error) {
	// TODO: The signing root shouldn't be set in this method, instead we should split the entire flow into two separate parts: create/update document & sign document
	err = dp.calculateSigningRoot(document)
	if err != nil {
		return err
	}
	signingService := signatures.GetSigningService()
	err = signingService.Sign(document)
	if err != nil {
		return err
	}
	return nil
}
