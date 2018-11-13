package documents

import (
	"fmt"
	"reflect"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/config"
	"context"
)

// Model is an interface to abstract away model specificness like invoice or purchaseOrder
// The interface can cast into the type specified by the model if required
type Model interface {

	// Get the ID of the document represented by this model
	ID() ([]byte, error)

	// PackCoreDocument packs the implementing document into a core document
	// should create the identifiers for the core document if not present
	PackCoreDocument() (*coredocumentpb.CoreDocument, error)

	// UnpackCoreDocument must return the document.Model
	// assumes that core document has valid identifiers set
	UnpackCoreDocument(cd *coredocumentpb.CoreDocument) error

	//Returns the underlying type of the Model
	Type() reflect.Type

	// JSON return the json representation of the model
	JSON() ([]byte, error)

	// FromJSON initialize the model with a json
	FromJSON(json []byte) error
}

// Placeholder to pass custom request objects down the pipeline
type ContextHeader struct {
	context context.Context
	self *identity.IdentityConfig
}

// NewContextHeader creates new instance of the request headers needed
func NewContextHeader(context context.Context, config *config.Configuration) (*ContextHeader, error) {
	idConfig, err := identity.GetIdentityConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to get id config: %v", err)
	}

	return &ContextHeader{self: idConfig, context: context}, nil
}

// Self returns Self CentID
func (h *ContextHeader) Self() *identity.IdentityConfig {
	return h.self
}

// Context returns context.Context of the request
func (h *ContextHeader) Context() context.Context {
	return h.context
}
