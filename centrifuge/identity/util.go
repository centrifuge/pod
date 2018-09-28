package identity

import "errors"

const (
	CentIdParam string = "CentID"
	BlockHeight string = "BlockHeight"
)

func getBytes32(key interface{}) ([32]byte, error) {
	var fixed [32]byte
	b, ok := key.([]interface{})
	if !ok {
		return fixed, errors.New("Could not parse interface to []byte")
	}
	// convert and copy b byte values
	for i, v := range b {
		fv := v.(float64)
		fixed[i] = byte(fv)
	}
	return fixed, nil
}

func getCentID(key interface{}) (CentID, error) {
	var fixed [CentIDByteLength]byte
	b, ok := key.([]interface{})
	if !ok {
		return fixed, errors.New("Could not parse interface to []byte")
	}
	// convert and copy b byte values
	for i, v := range b {
		fv := v.(float64)
		fixed[i] = byte(fv)
	}
	return fixed, nil
}

func parseBlockHeight(valMap map[string]interface{}) (uint64, error) {
	if bhi, ok := valMap[BlockHeight]; ok {
		bhf, ok := bhi.(float64)
		if ok {
			return uint64(bhf), nil
		} else {
			return 0, errors.New("value can not be parsed")
		}
	}
	return 0, errors.New("value can not be parsed")
}
