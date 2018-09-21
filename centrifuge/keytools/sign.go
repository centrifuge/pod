package keytools

import (
	"fmt"
	"strings"

	"github.com/centrifuge/go-centrifuge/centrifuge/keytools/secp256k1"
)

func SignMessage(privateKey, message []byte, curveType string, ethereumSign bool) ([]byte, error) {

	curveType = strings.ToLower(curveType)

	switch curveType {

	case CurveSecp256K1:
		msg := make([]byte, MaxMsgLen)
		copy(msg, message)
		if ethereumSign {
			return secp256k1.SignEthereum(msg, privateKey)
		}
		return secp256k1.Sign(msg, privateKey)

	case CurveEd25519:
		return []byte(""), fmt.Errorf("curve ed25519 not supported yet")

	default:
		return []byte(""), fmt.Errorf("curve %s not supported", curveType)
	}

}
