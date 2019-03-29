package identity

import (
	"strings"

	"github.com/centrifuge/go-centrifuge/utils"
)

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

// DIDsPointers returns the pointers to DIDs
func DIDsPointers(dids ...DID) []*DID {
	var pdids []*DID
	for _, did := range dids {
		did := did
		if utils.IsEmptyAddress(did.ToAddress()) {
			pdids = append(pdids, nil)
			continue
		}

		pdids = append(pdids, &did)
	}

	return pdids
}

// FromPointerDIDs return pointer DIDs to value DIDs
func FromPointerDIDs(pdids ...*DID) []DID {
	dids := make([]DID, len(pdids), len(pdids))
	for i, pdid := range pdids {
		pdid := pdid
		if pdid == nil {
			dids[i] = DID{}
			continue
		}

		dids[i] = *pdid
	}

	return dids
}

// RemoveDuplicateDIDs removes duplicate DIDs
func RemoveDuplicateDIDs(dids []DID) []DID {
	m := make(map[string]struct{})
	var res []DID
	for _, did := range dids {
		ls := strings.ToLower(did.String())
		if _, ok := m[ls]; ok {
			continue
		}

		res = append(res, did)
		m[ls] = struct{}{}
	}

	return res
}
