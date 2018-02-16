package tools


import (
	"crypto/rand"
)


func RandomString32() (ret string){
	b := RandomByte32()
	ret = string(b[:32])
	return
}


func RandomByte32() (out [32]byte) {
	r := make([]byte, 32)
	_, err := rand.Read(r)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		panic(err)
	}
	copy(out[:], r[:32])
	return
}