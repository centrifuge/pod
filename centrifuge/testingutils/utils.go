package testingutils

import "crypto/rand"

func Rand32Bytes () []byte {
	randbytes := make([]byte, 32)
	rand.Read(randbytes)
	return randbytes
}