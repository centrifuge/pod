package documents

import (
	"context"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
)

// AnchorProcessor identifies an implementation, which can do a bunch of things with a CoreDocument.
// E.g. send, anchor, etc.
type AnchorProcessor interface {
	Send(ctx context.Context, cd coredocumentpb.CoreDocument, recipient identity.DID) (err error)
	PrepareForSignatureRequests(ctx context.Context, model Model) error
	RequestSignatures(ctx context.Context, model Model) error
	PrepareForAnchoring(model Model) error
	PreAnchorDocument(ctx context.Context, model Model) error
	AnchorDocument(ctx context.Context, model Model) error
	SendDocument(ctx context.Context, model Model) error
}

// updaterFunc is a wrapper that will be called to save the state of the model between processor steps
type updaterFunc func(id []byte, model Model) error

// AnchorDocument add signature, requests signatures, anchors document, and sends the anchored document
// to collaborators
func AnchorDocument(ctx context.Context, model Model, proc AnchorProcessor, updater updaterFunc, preAnchor bool) (Model, error) {
	id := model.CurrentVersion()
	err := proc.PrepareForSignatureRequests(ctx, model)
	if err != nil {
		return nil, errors.NewTypedError(ErrDocumentAnchoring, errors.New("failed to prepare document for signatures: %v", err))
	}

	err = updater(id, model)
	if err != nil {
		return nil, err
	}

	if preAnchor {
		err = proc.PreAnchorDocument(ctx, model)
		if err != nil {
			return nil, err
		}
	}

	err = proc.RequestSignatures(ctx, model)
	if err != nil {
		return nil, errors.NewTypedError(ErrDocumentAnchoring, errors.New("failed to collect signatures: %v", err))
	}

	err = updater(id, model)
	if err != nil {
		return nil, errors.NewTypedError(ErrDocumentAnchoring, err)
	}

	err = proc.PrepareForAnchoring(model)
	if err != nil {
		return nil, errors.NewTypedError(ErrDocumentAnchoring, errors.New("failed to prepare for anchoring: %v", err))
	}

	err = updater(id, model)
	if err != nil {
		return nil, err
	}

	// TODO [TXManager] this function creates a child task in the queue which should be removed and called from the TxManger function
	err = proc.AnchorDocument(ctx, model)
	if err != nil {
		return nil, errors.NewTypedError(ErrDocumentAnchoring, errors.New("failed to anchor document: %v", err))
	}

	// set the status to committed
	if err = model.SetStatus(Committed); err != nil {
		return nil, err
	}

	err = updater(id, model)
	if err != nil {
		return nil, errors.NewTypedError(ErrDocumentAnchoring, err)
	}

	err = proc.SendDocument(ctx, model)
	if err != nil {
		return nil, errors.NewTypedError(ErrDocumentAnchoring, errors.New("failed to send anchored document: %v", err))
	}

	err = updater(id, model)
	if err != nil {
		return nil, errors.NewTypedError(ErrDocumentAnchoring, err)
	}

	return model, nil
}
