// +build unit integration

package testinganchors

import (
	"github.com/centrifuge/go-centrifuge/anchors"
	"github.com/stretchr/testify/mock"
)

type MockAnchorRepo struct {
	mock.Mock
	anchors.AnchorRepository
}

func (r *MockAnchorRepo) GetDocumentRootOf(anchorID anchors.AnchorID) (anchors.DocumentRoot, error) {
	args := r.Called(anchorID)
	docRoot, _ := args.Get(0).(anchors.DocumentRoot)
	return docRoot, args.Error(1)
}
