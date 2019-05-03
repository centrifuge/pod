// +build unit

package funding

import (
	"testing"

	"github.com/stretchr/testify/assert"
)
import clientfundingpb "github.com/centrifuge/go-centrifuge/protobufs/gen/go/funding"

func TestInitFundingFromData(t *testing.T) {
	fdc := &clientfundingpb.FundingData{Currency: "eur"}
	fd := &Data{}
	fd.initFundingFromData(fdc)
	assert.Equal(t, fdc.Currency, fd.Currency)

}
