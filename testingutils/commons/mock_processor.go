// +build unit integration

package testingcommons

import (
	"context"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/p2p"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/stretchr/testify/mock"
)

type MockRequestProcessor struct {
	mock.Mock
}

func (m *MockRequestProcessor) RequestDocumentWithAccessToken(ctx context.Context, granterDID identity.DID, tokenIdentifier,
	documentIdentifier, delegatingDocumentIdentifier []byte) (*p2ppb.GetDocumentResponse, error) {
	args := m.Called(granterDID, tokenIdentifier, documentIdentifier, delegatingDocumentIdentifier)
	resp, _ := args.Get(0).(*p2ppb.GetDocumentResponse)
	return resp, args.Error(1)
}
