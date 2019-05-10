// +build unit

package funding

import (
	"context"
	"fmt"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

func TestService_SignVerify(t *testing.T) {
	fundingAmount := 5
	srv, model, fundingID := setupFundingsForTesting(t, fundingAmount)

	// add signature
	acc := &mockAccount{}
	acc.On("GetIdentityID").Return(utils.RandomSlice(20), nil)
	// success
	signature := utils.RandomSlice(32)
	acc.On("SignMsg", mock.Anything).Return(&coredocumentpb.Signature{Signature: signature}, nil)
	ctx, err := contextutil.New(context.Background(), acc)
	assert.NoError(t, err)

	model, err = srv.Sign(ctx, fundingID, utils.RandomSlice(32))
	assert.NoError(t, err)

	response, err := srv.DeriveFundingResponse(ctx,model,fundingID)
	assert.NoError(t, err)
	fmt.Println(response.Data.Signatures)

}