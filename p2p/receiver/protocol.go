package receiver

import (
	"fmt"
	"strings"

	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/libp2p/go-libp2p-protocol"
)

// CentrifugeProtocol is the centrifuge wire protocol
const CentrifugeProtocol protocol.ID = "/centrifuge/0.0.1"

// ProtocolForCID creates the protocol string for the given CID
func ProtocolForCID(CID identity.CentID) protocol.ID {
	return protocol.ID(fmt.Sprintf("%s/%s", CentrifugeProtocol, CID.String()))
}

// ExtractCID extracts CID from a protocol string
func ExtractCID(id protocol.ID) (identity.CentID, error) {
	parts := strings.Split(string(id), "/")
	cidHexStr := parts[len(parts)-1]
	return identity.CentIDFromString(cidHexStr)
}
