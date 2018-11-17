package identity

import "errors"

const (
	centIDParam string = "CentID"
)

func getBytes32(key interface{}) ([32]byte, error) {
	var fixed [32]byte
	b, ok := key.([]interface{})
	if !ok {
		return fixed, errors.New("could not parse interface to []byte")
	}
	// convert and copy b byte values
	for i, v := range b {
		fv := v.(float64)
		fixed[i] = byte(fv)
	}
	return fixed, nil
}

func getCentID(key interface{}) (CentID, error) {
	var fixed [CentIDLength]byte
	b, ok := key.([]interface{})
	if !ok {
		return fixed, errors.New("could not parse interface to []byte")
	}
	// convert and copy b byte values
	for i, v := range b {
		fv := v.(float64)
		fixed[i] = byte(fv)
	}
	return fixed, nil
}
