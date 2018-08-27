package keytools

import (
	logging "github.com/ipfs/go-log"
)

var log = logging.Logger("keytools")

const (
	PublicKey  = "PUBLIC KEY"
	PrivateKey = "PRIVATE KEY"
)

const (
	CurveEd25519   string = "ed25519"
	CurveSecp256K1 string = "secp256k1"
)

const MaxMsgLen = 32
