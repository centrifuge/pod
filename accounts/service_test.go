// +build unit

package accounts

import (
	"testing"

	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/stretchr/testify/assert"
)

var srv Service

func TestService(t *testing.T) {
	acc := NewAccount(utils.RandomSlice(32), utils.RandomSlice(32), "")
	cacc, err := srv.CreateAccount(acc)
	assert.NoError(t, err)
	assert.Equal(t, acc, cacc)

	gacc, err := srv.GetAccount(acc.AccountID())
	assert.NoError(t, err)
	assert.Equal(t, acc, gacc)

	acc.(*account).Secret = utils.RandomSlice(32)
	uacc, err := srv.UpdateAccount(acc)
	assert.NoError(t, err)
	assert.Equal(t, acc, uacc)

	assert.NoError(t, srv.DeleteAccount(acc.AccountID()))
}
