package entity

import "github.com/centrifuge/go-centrifuge/errors"

const (
	ErrMultiplePaymentMethodsSet = errors.Error("multiple payment methods are set")
	ErrNoPaymentMethodSet        = errors.Error("no payment method is set")
	ErrEntityInvalidData         = errors.Error("invalid entity data")
	ErrP2PDocumentRequest        = errors.Error("couldn't request document via the p2p layer")
	ErrDocumentDerive            = errors.Error("couldn't derive document")
	ErrIdentityNotACollaborator  = errors.Error("identity not a collaborator")
	ErrEntityDataNoIdentity      = errors.Error("no identity in entity data")
)
