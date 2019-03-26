package identity

import "strings"

// StringsToDIDs converts hex strings to DIDs.
func StringsToDIDs(strs ...string) ([]*DID, error) {
	dids := make([]*DID, len(strs), len(strs))

	for i, str := range strs {
		str = strings.TrimSpace(str)
		if str == "" {
			continue
		}

		did, err := NewDIDFromString(str)
		if err != nil {
			return nil, err
		}

		dids[i] = &did
	}

	return dids, nil
}

// DIDsToStrings converts DIDs to hex strings.
func DIDsToStrings(dids ...*DID) []string {
	strs := make([]string, len(dids), len(dids))
	for i, did := range dids {
		if did == nil {
			continue
		}

		strs[i] = did.String()
	}

	return strs
}

// DIDsToBytes converts DIDs to bytes.
func DIDsToBytes(dids ...*DID) [][]byte {
	bytes := make([][]byte, len(dids), len(dids))
	for i, did := range dids {
		if did == nil {
			continue
		}

		bytes[i] = did[:]
	}

	return bytes
}

// BytesToDIDs converts bytes to DIDs
func BytesToDIDs(bytes ...[]byte) ([]*DID, error) {
	dids := make([]*DID, len(bytes), len(bytes))
	for i, bs := range bytes {
		if len(bs) < 1 {
			continue
		}

		did, err := NewDIDFromBytes(bs)
		if err != nil {
			return nil, err
		}
		dids[i] = &did
	}

	return dids, nil
}
