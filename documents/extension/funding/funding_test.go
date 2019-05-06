// +build unit

package funding

import (
	"testing"

	clientfundingpb "github.com/centrifuge/go-centrifuge/protobufs/gen/go/funding"
	"github.com/stretchr/testify/assert"
)

func TestInitFundingFromData(t *testing.T) {
	fdc := &clientfundingpb.FundingData{Currency: "eur"}
	fd := &Data{}
	fd.initFundingFromData(fdc)
	assert.Equal(t, fdc.Currency, fd.Currency)

}
