// +build unit

package funding

import (
	"testing"

	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/protobufs/gen/go/document"
	clientfunpb "github.com/centrifuge/go-centrifuge/protobufs/gen/go/funding"
	"github.com/centrifuge/go-centrifuge/testingutils/config"
	"github.com/centrifuge/go-centrifuge/testingutils/documents"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var configService config.Service

func TestGRPCHandler_GetList(t *testing.T) {
	srv := &MockService{}
	h := &grpcHandler{service: srv, config: configService}

	srv.On("GetVersion", mock.Anything, mock.Anything, mock.Anything).Return(&testingdocuments.MockModel{}, nil)
	srv.On("DeriveFundingListResponse", mock.Anything, mock.Anything).Return(&clientfunpb.FundingListResponse{Header: new(documentpb.ResponseHeader)}, nil).Once()

	response, err := h.GetListVersion(testingconfig.HandlerContext(configService), &clientfunpb.GetListVersionRequest{DocumentId: hexutil.Encode(utils.RandomSlice(32)), VersionId: hexutil.Encode(utils.RandomSlice(32))})
	assert.NoError(t, err)
	assert.NotNil(t, response)
}
