package migrationutils

import (
	"crypto/rand"
	"os"
	"path/filepath"
)

// RandomSlice returns a randomly filled byte array with length of given size
func RandomSlice(size int) (out []byte) {
	r := make([]byte, size)
	_, err := rand.Read(r)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		panic(err)
	}
	return r
}

// RandomByte32 returns a randomly filled byte array with length of 32
func RandomByte32() (out [32]byte) {
	r := RandomSlice(32)
	copy(out[:], r[:32])
	return
}

// CleanupDBFiles deletes all files in the provided path
func CleanupDBFiles(prefix string) {
	files, err := filepath.Glob(prefix + "*")
	if err != nil {
		panic(err)
	}
	for _, f := range files {
		if err := os.RemoveAll(f); err != nil {
			panic(err)
		}
	}
}
