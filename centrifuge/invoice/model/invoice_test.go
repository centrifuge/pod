// +build unit

package model

import (
	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/stretchr/testify/assert"
	"testing"
)


func TestInitWithEmptyCoreDocument(t *testing.T) {

	invoice := &Invoice{}
	emptyCoreDocument := &coredocumentpb.CoreDocument{}

	err := invoice.InitWithCoreDocument(emptyCoreDocument)

	assert.Error(t,err,"it should not be possible to init a empty core document")


}


