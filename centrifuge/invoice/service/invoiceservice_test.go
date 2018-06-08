// +build unit

package invoiceservice

import (
	"context"
	"crypto/sha256"
	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/coredocument"
	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/invoice"
	cc "github.com/CentrifugeInc/go-centrifuge/centrifuge/context"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/invoice"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/invoice/repository"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/testingutils"
	"github.com/centrifuge/precise-proofs/proofs"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	cc.TestBootstrap()
	result := m.Run()
	cc.TestTearDown()
	os.Exit(result)
}

func TestInvoiceDocumentService_SendReceive(t *testing.T) {
	s := InvoiceDocumentService{}
	identifier := testingutils.Rand32Bytes()
	identifierIncorrect := testingutils.Rand32Bytes()
	doc := invoice.NewEmptyInvoice()
	doc.Document.CoreDocument = &coredocumentpb.CoreDocument{
		DocumentIdentifier: identifier,
		CurrentIdentifier:  identifier,
		NextIdentifier:     testingutils.Rand32Bytes(),
		DataMerkleRoot:     testingutils.Rand32Bytes(),
	}

	sentDoc, err := s.HandleSendInvoiceDocument(context.Background(), &invoicepb.SendInvoiceEnvelope{Recipients: [][]byte{}, Document: doc.Document})
	assert.Nil(t, err, "Error in RPC Call")

	assert.Equal(t, sentDoc.CoreDocument.DocumentIdentifier, identifier,
		"DocumentIdentifier doesn't match")

	receivedDoc, err := s.HandleGetInvoiceDocument(context.Background(),
		&invoicepb.GetInvoiceDocumentEnvelope{DocumentIdentifier: identifier})
	assert.Nil(t, err, "Error in RPC Call")
	assert.Equal(t, receivedDoc.CoreDocument.DocumentIdentifier, identifier,
		"DocumentIdentifier doesn't match")

	_, err = s.HandleGetInvoiceDocument(context.Background(),
		&invoicepb.GetInvoiceDocumentEnvelope{DocumentIdentifier: identifierIncorrect})
	assert.NotNil(t, err,
		"RPC call should have raised exception")

}

func TestInvoiceDocumentService_HandleCreateInvoiceProof(t *testing.T) {
	s := InvoiceDocumentService{}

	identifier := testingutils.Rand32Bytes()
	inv := invoice.NewEmptyInvoice()
	inv.Document.CoreDocument = &coredocumentpb.CoreDocument{
		DocumentIdentifier: identifier,
		CurrentIdentifier:  identifier,
		NextIdentifier:     testingutils.Rand32Bytes(),
		// TODO: below should be actual merkle root
		DataMerkleRoot: testingutils.Rand32Bytes(),
	}
	inv.CalculateMerkleRoot()
	err := invoicerepository.GetInvoiceRepository().Store(inv.Document)
	assert.Nil(t, err)

	proofRequest := &invoicepb.CreateInvoiceProofEnvelope{
		DocumentIdentifier: identifier,
		Fields:             []string{"currency", "country", "amount"},
	}

	invoiceProof, err := s.HandleCreateInvoiceProof(context.Background(), proofRequest)
	assert.Nil(t, err)
	assert.Equal(t, len(proofRequest.Fields), len(invoiceProof.FieldProofs))
	assert.Equal(t, proofRequest.Fields[0], invoiceProof.FieldProofs[0].Property)
	sha256Hash := sha256.New()
	valid, err := proofs.ValidateProof(invoiceProof.FieldProofs[0], inv.Document.CoreDocument.DocumentRoot, sha256Hash)
	assert.True(t, valid)
	assert.Nil(t, err)
}
