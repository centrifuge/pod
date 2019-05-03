// +build unit

package funding

import (
	"github.com/stretchr/testify/assert"
	"testing"
)
import clientfundingpb "github.com/centrifuge/go-centrifuge/protobufs/gen/go/funding"

func TestInitFundingFromData(t *testing.T) {
	fdc := &clientfundingpb.FundingData{Currency: "eur"}
	fd := &FundingData{}
	fd.initFundingFromData(fdc)
	assert.Equal(t,fdc.Currency, fd.Currency)

}
