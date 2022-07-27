//go:build unit

package contextutil

import (
	"context"
	"testing"

	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"
)

func TestDIDFromContext(t *testing.T) {
	// missing header
	_, err := DIDFromContext(context.Background())
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(ErrDIDMissingFromContext, err))

	// invalid did
	_, err = DIDFromContext(context.WithValue(context.Background(), config.AccountHeaderKey, "some value"))
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(hexutil.ErrMissingPrefix, err))

	// success
	did, err := identity.NewDIDFromString("0xF72855759A39FB75fC7341139f5d7A3974d4DA08")
	assert.NoError(t, err)
	ddid, err := DIDFromContext(context.WithValue(context.Background(), config.AccountHeaderKey, did.String()))
	assert.NoError(t, err)
	assert.Equal(t, did, ddid)
}
