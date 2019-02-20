package documents

import (
	"encoding/json"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/utils"
)

// CoreDocument holds the protobuf coredocument.
type CoreDocument struct {
	document coredocumentpb.CoreDocument
}

// jsonCD is an intermediate document type used for marshalling and un-marshaling CoreDocument to/from json.
type jsonCD struct {
	CoreDocument *coredocumentpb.CoreDocument `json:"core_document"`
}

// MarshalJSON returns a JSON formatted representation of the CoreDocument
func (cd *CoreDocument) MarshalJSON() ([]byte, error) {
	return json.Marshal(&jsonCD{CoreDocument: &cd.document})
}

// UnmarshalJSON loads the json formatted CoreDocument.
func (cd *CoreDocument) UnmarshalJSON(data []byte) error {
	jcd := new(jsonCD)
	err := json.Unmarshal(data, jcd)
	if err != nil {
		return err
	}

	cd.document = *jcd.CoreDocument
	return nil
}

// NewCoreDocument returns a new CoreDocument.
func NewCoreDocument() *CoreDocument {
	id := utils.RandomSlice(32)
	cd := coredocumentpb.CoreDocument{
		DocumentIdentifier: id,
		CurrentVersion:     id,
		NextVersion:        utils.RandomSlice(32),
	}

	return &CoreDocument{cd}
}

// setSalts generate salts for core document.
// This is no-op if the salts are already generated.
// TODO remove this
func (cd *CoreDocument) setSalts() error {
	if cd.document.CoredocumentSalts != nil {
		return nil
	}

	pSalts, err := GenerateNewSalts(&cd.document, "", nil)
	if err != nil {
		return err
	}

	cd.document.CoredocumentSalts = ConvertToProtoSalts(pSalts)
	return nil
}
