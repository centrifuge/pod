package coredocument

import (
	"context"
	"crypto/sha256"
	"fmt"
	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/coredocument"
	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/p2p"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/anchor"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/errors"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/identity"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/p2p"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/signatures"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/tools"
	"github.com/centrifuge/precise-proofs/proofs"
	logging "github.com/ipfs/go-log"
)

var log = logging.Logger("coredocument")

// ----- ERROR -----
type ErrInconsistentState struct {
	message string
}

func NewErrInconsistentState(message string) *ErrInconsistentState {
	padded := ""
	if len(message) > 0 {
		padded = fmt.Sprintf(": %s", message)
	}
	return &ErrInconsistentState{
		message: fmt.Sprintf("Inconsistent CoreDocument state%s", padded),
	}
}
func (e *ErrInconsistentState) Error() string {
	return e.message
}

// ----- END ERROR -----

// CoreDocumentProcessor is the processor that can deal with CoreDocuments and performs actions on them such as
// anchoring, sending on the p2p level, or signing.
type CoreDocumentProcessor struct {
	IdentityService identity.IdentityService
}

// CoreDocumentProcessorInterface identifies an implementation, which can do a bunch of things with a CoreDocument.
// E.g. send, anchor, etc.
type CoreDocumentProcessorInterface interface {
	Send(coreDocument *coredocumentpb.CoreDocument, ctx context.Context, recipient []byte) (err error)
	Anchor(document *coredocumentpb.CoreDocument) (err error)
}

func GetDefaultCoreDocumentProcessor() CoreDocumentProcessorInterface {
	return &CoreDocumentProcessor{IdentityService: identity.NewEthereumIdentityService()}
}

// Send sends the given CoreDocumentProcessor to the given recipient on the P2P layer
func (cdp *CoreDocumentProcessor) Send(coreDocument *coredocumentpb.CoreDocument, ctx context.Context, recipient []byte) (err error) {
	if coreDocument == nil {
		return errors.GenerateNilParameterError(coreDocument)
	}

	id, err := cdp.IdentityService.LookupIdentityForId(recipient)
	if err != nil {
		log.Errorf("Error: %v\n", err)
		return err
	}

	lastb58Key, err := id.GetCurrentP2PKey()
	if err != nil {
		log.Errorf("Error: %v\n", err)
		return err
	}

	log.Infof("Sending Document to CentID [%v] with Key [%v]\n", recipient, lastb58Key)
	clientWithProtocol := fmt.Sprintf("/ipfs/%s", lastb58Key)
	client := p2p.OpenClient(clientWithProtocol)
	log.Infof("Done opening connection against [%s]\n", lastb58Key)

	hostInstance := p2p.GetHost()
	bSenderId, err := hostInstance.ID().ExtractPublicKey().Bytes()
	if err != nil {
		return err
	}
	_, err = client.Post(context.Background(), &p2ppb.P2PMessage{Document: coreDocument, SenderCentrifugeId: bSenderId})
	if err != nil {
		return err
	}
	return
}

// Anchor anchors the given CoreDocument
func (cd *CoreDocumentProcessor) Anchor(document *coredocumentpb.CoreDocument) error {
	if document == nil {
		return errors.GenerateNilParameterError(document)
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

	confirmations := make(chan *anchor.WatchAnchor, 1)
	err = anchor.RegisterAsAnchor(id, rootHash, confirmations)
	if err != nil {
		log.Error(err)
		return err
	}
	anchorWatch := <-confirmations
	err = anchorWatch.Error
	return err
}

// ValidateCoreDocument checks that all required fields are set before doing any processing with it
func (cd *CoreDocumentProcessor) ValidateCoreDocument(document *coredocumentpb.CoreDocument) (valid bool, err error) {
	if !tools.CheckMultiple32BytesFilled(document.DocumentIdentifier, document.NextIdentifier, document.CurrentIdentifier, document.DataRoot) {
		return false, errors.New("Found empty value in CoreDocument")
	}

	if document.CoredocumentSalts == nil {
		return false, errors.New("CoreDocumentSalts is not set")
	}

	// Spot checking that DocumentIdentifier salt is filled. Perhaps it would be better to validate all salts in the future.
	if tools.IsEmptyByteSlice(document.CoredocumentSalts.DocumentIdentifier) || len(document.CoredocumentSalts.DocumentIdentifier) != 32 {
		return false, errors.New("CoreDocumentSalts not filled")
	}

	return true, nil
}

func (cd *CoreDocumentProcessor) getDocumentTree(document *coredocumentpb.CoreDocument) (tree *proofs.DocumentTree, err error) {
	t := proofs.NewDocumentTree()
	tree = &t
	sha256Hash := sha256.New()
	tree.SetHashFunc(sha256Hash)
	err = tree.FillTree(document, document.CoredocumentSalts)
	if err != nil {
		return nil, err
	}
	return tree, nil
}

func (cd *CoreDocumentProcessor) CalculateSigningRoot(document *coredocumentpb.CoreDocument) error {
	valid, err := cd.ValidateCoreDocument(document)
	if !valid {
		return err
	}
	tree, err := cd.getDocumentTree(document)
	document.SigningRoot = tree.RootHash()
	return nil
}

func (cd *CoreDocumentProcessor) Sign(document *coredocumentpb.CoreDocument) (err error) {
	// TODO: The signing root shouldn't be set in this method, instead we should split the entire flow into two separate parts: create/update document & sign document
	err = cd.CalculateSigningRoot(document)
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
