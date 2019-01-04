package ethid

import (
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
)

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

func getCentID(key interface{}) (identity.CentID, error) {
	var fixed [identity.CentIDLength]byte
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
