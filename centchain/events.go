package centchain

import (
	centEvents "github.com/centrifuge/chain-custom-types/pkg/events"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
)

// Events holds the default events and custom events for centrifuge chain
type Events struct {
	types.EventRecords
	centEvents.Events
}
