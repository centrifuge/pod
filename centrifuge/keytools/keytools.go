package keytools

import (
	logging "github.com/ipfs/go-log"
)

var log = logging.Logger("keytools")

const (
	PUBLIC_KEY  = "PUBLIC KEY"
	PRIVATE_KEY = "PRIVATE KEY"
)

const (
	CURVE_ED25519   string = "ed25519"
	CURVE_SECP256K1 string = "secp256k1"
)

const MAX_MSG_LEN = 32


