// +build ethereum

package coredocumentprocessor

import (
	"testing"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/identity"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/testingcommons"
	"context"
)

func TestDefaultProcessor_Anchor(t *testing.T) {
	dp := DefaultProcessor(identity.NewEthereumIdentityService(), &testingcommons.MockP2PWrapperClient{})
	dp.Anchor(context.Background(), nil, []identity.CentID{})
}