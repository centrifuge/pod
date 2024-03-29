//go:build unit

package contextutil

import (
	"context"
	"testing"

	"github.com/centrifuge/pod/config"
	testingcommons "github.com/centrifuge/pod/testingutils/common"
	"github.com/stretchr/testify/assert"
)

func TestContextActions(t *testing.T) {
	ctx := context.Background()

	ctxAccount, err := Account(ctx)
	assert.ErrorIs(t, err, ErrSelfNotFound)
	assert.Nil(t, ctxAccount)

	identity, err := Identity(ctx)
	assert.ErrorIs(t, err, ErrSelfNotFound)
	assert.Nil(t, identity)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").Return(accountID)

	ctx = WithAccount(ctx, accountMock)

	ctxAccount, err = Account(ctx)
	assert.NoError(t, err)
	assert.Equal(t, accountMock, ctxAccount)

	identity, err = Identity(ctx)
	assert.NoError(t, err)
	assert.Equal(t, accountID, identity)
}
