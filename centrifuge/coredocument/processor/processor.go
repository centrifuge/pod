package coredocumentprocessor

import (
	"context"
	"fmt"

	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/coredocument"
	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/p2p"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/centerrors"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/coredocument"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/identity"
	centED25519 "github.com/CentrifugeInc/go-centrifuge/centrifuge/keytools/ed25519"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/p2p"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/signatures"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/tools"
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
	IdentityService identity.Service
}

// DefaultProcessor returns the default implementation of CoreDocument Processor
// TODO(ved): I don't think we need the processor since IdentityService is available globally
func DefaultProcessor() Processor {
	return &defaultProcessor{
		IdentityService: identity.NewEthereumIdentityService()}
}

// Send sends the given defaultProcessor to the given recipient on the P2P layer
func (dp *defaultProcessor) Send(coreDocument *coredocumentpb.CoreDocument, ctx context.Context, recipient []byte) (err error) {
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
	client, err := p2p.OpenClient(clientWithProtocol)
	if err != nil {
		return fmt.Errorf("failed to open client: %v", err)
	}

	log.Infof("Done opening connection against [%s]\n", lastB58Key)
	hostInstance := p2p.GetHost()
	bSenderId, err := hostInstance.ID().ExtractPublicKey().Bytes()
	if err != nil {
		err = centerrors.Wrap(err, "failed to extract pub key")
		log.Error(err)
		return err
	}

	_, err = client.Post(context.Background(), &p2ppb.P2PMessage{Document: coreDocument, SenderCentrifugeId: bSenderId})
	if err != nil {
		err = centerrors.Wrap(err, "failed to post to the node")
		log.Error(err)
		return err
	}

	return nil
}

// Anchor anchors the given CoreDocument
func (dp *defaultProcessor) Anchor(document *coredocumentpb.CoreDocument) error {
	if document == nil {
		return centerrors.NilError(document)
	}

	_, err := tools.SliceToByte32(document.CurrentIdentifier)
	if err != nil {
		log.Error(err)
		return err
	}

	log.Infof("Anchoring document with identifiers: [document: %#x, current: %#x, next: %#x], rootHash: %#x", document.DocumentIdentifier, document.CurrentIdentifier, document.NextIdentifier, document.DocumentRoot)
	log.Debugf("Anchoring document with details %v", document)

	err = coredocument.CalculateSigningRoot(document)
	if err != nil {
		log.Error(err)
		return err
	}

	// TODO anchoring
	//rootHash, err := tools.SliceToByte32(document.DocumentRoot) //TODO: CHANGE
	//if err != nil {
	//	log.Error(err)
	//	return err
	//}
	//
	//idConfig, err := centED25519.GetIDConfig()
	//if err != nil {
	//	log.Error(err)
	//	return err
	//}
	//
	//var centId [identity.CentIdByteLength]byte
	//copy(centId[:], idConfig.ID[:identity.CentIdByteLength])
	//
	//signature, err := secp256k1.SignEthereum(anchoring.GenerateCommitHash(id, centId, rootHash), idConfig.PrivateKey)
	//if err != nil {
	//	log.Error(err)
	//	return err
	//}
	//
	//// TODO documentProofs has to be included when we develop precommit flow
	//confirmations, err := anchoring.CommitAnchor(id, rootHash, centId, [][anchoring.DocumentProofLength]byte{tools.RandomByte32()}, signature)
	//if err != nil {
	//	log.Error(err)
	//	return err
	//}
	//
	//anchorWatch := <-confirmations
	return nil
}

func (dp *defaultProcessor) Sign(document *coredocumentpb.CoreDocument) (err error) {
	// TODO: The signing root shouldn't be set in this method, instead we should split the entire flow into two separate parts: create/update document & sign document
	err = coredocument.CalculateSigningRoot(document)
	if err != nil {
		return err
	}

	idConfig, err := centED25519.GetIDConfig()
	if err != nil {
		return err
	}

	sig := signatures.Sign(idConfig, document)
	document.Signatures = append(document.Signatures, sig)
	return nil
}
