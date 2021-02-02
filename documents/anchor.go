package documents

import (
	"context"

	coredocumentpb "github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
)

// AnchorProcessor identifies an implementation, which can do a bunch of things with a CoreDocument.
// E.g. send, anchor, etc.
type AnchorProcessor interface {
	Send(ctx context.Context, cd coredocumentpb.CoreDocument, recipient identity.DID) (err error)
	PrepareForSignatureRequests(ctx context.Context, doc Document) error
	RequestSignatures(ctx context.Context, doc Document) error
	PrepareForAnchoring(ctx context.Context, doc Document) error
	PreAnchorDocument(ctx context.Context, doc Document) error
	AnchorDocument(ctx context.Context, doc Document) error
	SendDocument(ctx context.Context, doc Document) error
}

// updaterFunc is a wrapper that will be called to save the state of the doc between processor steps
type updaterFunc func(id []byte, doc Document) error

// AnchorDocument add signature, requests signatures, anchors document, and sends the anchored document
// to collaborators
func AnchorDocument(ctx context.Context, doc Document, proc AnchorProcessor, updater updaterFunc, preAnchor bool) (Document, error) {
	id := doc.CurrentVersion()
	err := proc.PrepareForSignatureRequests(ctx, doc)
	if err != nil {
		return nil, errors.NewTypedError(ErrDocumentAnchoring, errors.New("failed to prepare document for signatures: %v", err))
	}

	err = updater(id, doc)
	if err != nil {
		return nil, err
	}

	if preAnchor {
		err = proc.PreAnchorDocument(ctx, doc)
		if err != nil {
			return nil, err
		}
	}

	err = proc.RequestSignatures(ctx, doc)
	if err != nil {
		return nil, errors.NewTypedError(ErrDocumentAnchoring, errors.New("failed to collect signatures: %v", err))
	}

	err = updater(id, doc)
	if err != nil {
		return nil, errors.NewTypedError(ErrDocumentAnchoring, err)
	}

	err = proc.PrepareForAnchoring(ctx, doc)
	if err != nil {
		return nil, errors.NewTypedError(ErrDocumentAnchoring, errors.New("failed to prepare for anchoring: %v", err))
	}

	err = updater(id, doc)
	if err != nil {
		return nil, err
	}

	// TODO [TXManager] this function creates a child task in the queue which should be removed and called from the TxManger function
	err = proc.AnchorDocument(ctx, doc)
	if err != nil {
		return nil, errors.NewTypedError(ErrDocumentAnchoring, errors.New("failed to anchor document: %v", err))
	}

	// set the status to committed
	if err = doc.SetStatus(Committed); err != nil {
		return nil, err
	}

	err = updater(id, doc)
	if err != nil {
		return nil, errors.NewTypedError(ErrDocumentAnchoring, err)
	}

	err = proc.SendDocument(ctx, doc)
	if err != nil {
		return nil, errors.NewTypedError(ErrDocumentAnchoring, errors.New("failed to send anchored document: %v", err))
	}

	err = updater(id, doc)
	if err != nil {
		return nil, errors.NewTypedError(ErrDocumentAnchoring, err)
	}

	return doc, nil
}
