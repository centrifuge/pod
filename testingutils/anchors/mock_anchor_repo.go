// +build unit integration

package testinganchors

import (
	"time"

	"github.com/centrifuge/go-centrifuge/anchors"
	"github.com/stretchr/testify/mock"
)

type MockAnchorService struct {
	mock.Mock
	anchors.Service
}

func (r *MockAnchorService) GetAnchorData(anchorID anchors.AnchorID) (docRoot anchors.DocumentRoot, anchoredTime time.Time, err error) {
	args := r.Called(anchorID)
	docRoot, _ = args.Get(0).(anchors.DocumentRoot)
	return docRoot, anchoredTime, args.Error(1)
}
