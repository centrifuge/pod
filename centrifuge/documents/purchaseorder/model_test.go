package purchaseorder

import (
	"testing"

	"github.com/centrifuge/go-centrifuge/centrifuge/testingutils/documents"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/stretchr/testify/assert"
)

func TestPO_FromCoreDocuments_invalidParameter(t *testing.T) {
	poModel := &PurchaseOrderModel{}

	emptyCoreDocument := &coredocumentpb.CoreDocument{}
	err := poModel.UnpackCoreDocument(emptyCoreDocument)
	assert.Error(t, err, "it should not be possible to init a empty core document")

	err = poModel.UnpackCoreDocument(nil)
	assert.Error(t, err, "it should not be possible to init a empty core document")

	invalidEmbeddedData := &any.Any{TypeUrl: "invalid"}
	coreDocument := &coredocumentpb.CoreDocument{EmbeddedData: invalidEmbeddedData}
	err = poModel.UnpackCoreDocument(coreDocument)
	assert.Error(t, err, "it should not be possible to init invalid typeUrl")

}

func TestPO_InitCoreDocument_successful(t *testing.T) {
	poModel := &PurchaseOrderModel{}

	poData := testingdocuments.CreatePOData()

	coreDocument := testingdocuments.CreateCDWithEmbeddedPO(t, poData)
	err := poModel.UnpackCoreDocument(coreDocument)
	assert.Nil(t, err, "valid coredocument shouldn't produce an error")
}
