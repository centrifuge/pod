package documents

import (
	"fmt"

	"github.com/centrifuge/go-centrifuge/header"
)

// anchorProcessor has same methods to coredoc processor
// this is to avoid import cycles
// this will disappear once we have queueing logic in place
type anchorProcessor interface {
	PrepareForSignatureRequests(ctx *header.ContextHeader, model Model) error
	RequestSignatures(ctx *header.ContextHeader, model Model) error
	PrepareForAnchoring(model Model) error
	AnchorDocument(ctx *header.ContextHeader, model Model) error
	SendDocument(ctx *header.ContextHeader, model Model) error
}

// updaterFunc is a wrapper that will be called to save the state of the model between processor steps
type updaterFunc func(id []byte, model Model) error

// AnchorDocument add signature, requests signatures, anchors document, and sends the anchored document
// to collaborators
func AnchorDocument(ctx *header.ContextHeader, model Model, proc anchorProcessor, updater updaterFunc) (Model, error) {
	cd, err := model.PackCoreDocument()
	if err != nil {
		return nil, err
	}

	id := cd.CurrentVersion
	err = proc.PrepareForSignatureRequests(ctx, model)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare document for signatures: %v", err)
	}

	err = updater(id, model)
	if err != nil {
		return nil, err
	}

	err = proc.RequestSignatures(ctx, model)
	if err != nil {
		return nil, fmt.Errorf("failed to collect signatures: %v", err)
	}

	err = updater(id, model)
	if err != nil {
		return nil, err
	}

	err = proc.PrepareForAnchoring(model)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare for anchoring: %v", err)
	}

	err = updater(id, model)
	if err != nil {
		return nil, err
	}

	err = proc.AnchorDocument(ctx, model)
	if err != nil {
		return nil, fmt.Errorf("failed to anchor document: %v", err)
	}

	err = updater(id, model)
	if err != nil {
		return nil, err
	}

	err = proc.SendDocument(ctx, model)
	if err != nil {
		return nil, fmt.Errorf("failed to send anchored document: %v", err)
	}

	err = updater(id, model)
	if err != nil {
		return nil, err
	}

	return model, nil
}
