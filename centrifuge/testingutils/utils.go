package testingutils

import (
	"crypto/rand"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/config"
	"github.com/stretchr/testify/mock"
	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/coredocument"
	"context"
	"fmt"
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

func GenerateP2PRecipients(quantity int) ([][]byte) {
	recipients := make([][]byte, quantity)

	for i := 0; i < quantity; i++ {
		recipients[0] = []byte(fmt.Sprintf("RecipientNo[%d]", quantity))
	}
	return recipients
}

func GenerateCoreDocument()(*coredocumentpb.CoreDocument){
	identifier := Rand32Bytes()
	return &coredocumentpb.CoreDocument{
		DocumentIdentifier: identifier,
		CurrentIdentifier:  identifier,
		NextIdentifier:     Rand32Bytes(),
		DataMerkleRoot:     Rand32Bytes(),
	}
}

type MockCoreDocumentProcessor struct {
	mock.Mock
}

func (m *MockCoreDocumentProcessor) Send(coreDocument *coredocumentpb.CoreDocument, ctx context.Context, recipient string) (err error) {
	args := m.Called(coreDocument, ctx, recipient)
	return args.Error(0)
}

func (m *MockCoreDocumentProcessor) Anchor(coreDocument *coredocumentpb.CoreDocument) (err error) {
	args := m.Called(coreDocument)
	return args.Error(0)
}