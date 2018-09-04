package testingutils

import (
	"context"
	"crypto/rand"
	"fmt"

	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/coredocument"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/config"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/identity"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/tools"
	"github.com/centrifuge/precise-proofs/proofs"
	"github.com/stretchr/testify/mock"
)

func MockConfigOption(key string, value interface{}) func() {
	mockedValue := config.Config.V.Get(key)
	config.Config.V.Set(key, value)
	return func() {
		config.Config.V.Set(key, mockedValue)
	}
}

func Rand32Bytes() []byte {
	randbytes := make([]byte, 32)
	rand.Read(randbytes)
	return randbytes
}

func GenerateP2PRecipients(quantity int) [][]byte {
	recipients := make([][]byte, quantity)

	for i := 0; i < quantity; i++ {
		recipients[i] = []byte(fmt.Sprintf("RecipientNo[%d]", i))
	}
	return recipients
}

func GenerateCoreDocument() *coredocumentpb.CoreDocument {
	identifier := Rand32Bytes()
	salts := &coredocumentpb.CoreDocumentSalts{}
	proofs.FillSalts(salts)
	return &coredocumentpb.CoreDocument{
		DataRoot:           tools.RandomSlice(32),
		DocumentIdentifier: identifier,
		CurrentIdentifier:  identifier,
		NextIdentifier:     Rand32Bytes(),
		CoredocumentSalts:  salts,
	}
}

type MockCoreDocumentProcessor struct {
	mock.Mock
}

func (m *MockCoreDocumentProcessor) Send(coreDocument *coredocumentpb.CoreDocument, ctx context.Context, recipient []byte) (err error) {
	args := m.Called(coreDocument, ctx, recipient)
	return args.Error(0)
}

func (m *MockCoreDocumentProcessor) Anchor(coreDocument *coredocumentpb.CoreDocument) (err error) {
	args := m.Called(coreDocument)
	return args.Error(0)
}

// MockIDService implements Service
type MockIDService struct {
	mock.Mock
}

// LookUpIdentityForID returns a
func (srv *MockIDService) LookupIdentityForId(centID []byte) (identity.Identity, error) {
	args := srv.Called(centID)
	id, _ := args.Get(0).(identity.Identity)
	return id, args.Error(1)
}

func (srv *MockIDService) CreateIdentity(centID []byte) (identity.Identity, chan *identity.WatchIdentity, error) {
	args := srv.Called(centID)
	id, _ := args.Get(0).(identity.Identity)
	return id, args.Get(1).(chan *identity.WatchIdentity), args.Error(2)
}

func (srv *MockIDService) CheckIdentityExists(centID []byte) (exists bool, err error) {
	args := srv.Called(centID)
	return args.Bool(0), args.Error(1)
}

type MockSubscription struct {
	ErrChan chan error
}

func (m *MockSubscription) Err() <-chan error {
	return m.ErrChan
}

func (*MockSubscription) Unsubscribe() {}
