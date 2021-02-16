// +build integration unit testworld

package anchors

import (
	"context"
	"time"

	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/stretchr/testify/mock"

	"github.com/centrifuge/go-centrifuge/bootstrap"
)

func (b Bootstrapper) TestBootstrap(context map[string]interface{}) error {
	if _, ok := context[bootstrap.BootstrappedConfig]; !ok {
		return errors.New("config hasn't been initialized")
	}

	return b.Bootstrap(context)
}

func (b Bootstrapper) TestTearDown() error {
	return nil
}

type MockAnchorService struct {
	mock.Mock
	Service
}

func (m *MockAnchorService) CommitAnchor(ctx context.Context, anchorID AnchorID, documentRoot DocumentRoot, documentProof [32]byte) (err error) {
	args := m.Called(anchorID, documentRoot, documentProof)
	return args.Error(0)
}

func (m *MockAnchorService) GetAnchorData(anchorID AnchorID) (docRoot DocumentRoot, anchoredTime time.Time, err error) {
	args := m.Called(anchorID)
	docRoot, _ = args.Get(0).(DocumentRoot)
	return docRoot, anchoredTime, args.Error(1)
}

func (m *MockAnchorService) PreCommitAnchor(ctx context.Context, anchorID AnchorID, signingRoot DocumentRoot) (err error) {
	args := m.Called(anchorID, signingRoot)
	return args.Error(0)
}

// RandomDocumentRoot returns a randomly generated DocumentRoot
func RandomDocumentRoot() DocumentRoot {
	root, _ := ToDocumentRoot(utils.RandomSlice(DocumentRootLength))
	return root
}
