// +build pedersenunit

package pedersen

import (
	"encoding/hex"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestHashPayload(t *testing.T) {
	ph := New()

	//_, err := ph.Write(utils.RandomSlice(64))
	//assert.NoError(t, err)
	//
	//s := ph.Sum(nil)
	//assert.Len(t, s, 32)
	//
	//ph.Reset()
	//wrong length
	//_, err = ph.Write(utils.RandomSlice(32))
	//assert.Error(t, err)


	pw := "046d1ba196fd01e2ce7d784e85c401fa32c2ae8b6f57515d910fd87c69fde7044599c07a7da741240381f401c3506027f9808f24e0095b546f0003fd885d6d9b"
	pwb, err := hex.DecodeString(pw)
	assert.NoError(t, err)

	st := time.Now()
	_, err = ph.Write(pwb)

	assert.NoError(t, err)
	res := ph.Sum(nil)
	assert.Equal(t, "1329fd5020f30ef30d2ceec1f0db5c771d7f226719e8593dcf661523da1edecd", hex.EncodeToString(res))
	ph.Reset()

	for i := 0; i < 128; i++ {
		//st = time.Now()
		_, err = ph.Write(pwb)
		//fmt.Println(time.Now().Sub(st))
		//assert.NoError(t, err)
		_ = ph.Sum(nil)
		//assert.Equal(t, "1329fd5020f30ef30d2ceec1f0db5c771d7f226719e8593dcf661523da1edecd", hex.EncodeToString(res))
		ph.Reset()
	}
	et := time.Now().Sub(st)
	fmt.Println(et)
	//
	//_, err = ph.Write(pwb)
	//assert.NoError(t, err)
	//res = ph.Sum(nil)
	//assert.Equal(t, "cdde1eda231566cf3d59e81967227f1d775cdbf0c1ee2c0df30ef32050fd2913", hex.EncodeToString(res))
	//ph.Reset()
	//_, err = ph.Write(pwb)
	//assert.NoError(t, err)
	//res = ph.Sum(nil)
	//assert.Equal(t, "cdde1eda231566cf3d59e81967227f1d775cdbf0c1ee2c0df30ef32050fd2913", hex.EncodeToString(res))
	//ph.Reset()
	//_, err = ph.Write(pwb)
	//assert.NoError(t, err)
	//res = ph.Sum(nil)
	//assert.Equal(t, "cdde1eda231566cf3d59e81967227f1d775cdbf0c1ee2c0df30ef32050fd2913", hex.EncodeToString(res))
	//ph.Reset()
	//_, err = ph.Write(pwb)
	//assert.NoError(t, err)
	//res = ph.Sum(nil)
	//assert.Equal(t, "cdde1eda231566cf3d59e81967227f1d775cdbf0c1ee2c0df30ef32050fd2913", hex.EncodeToString(res))
	//ph.Reset()
	//_, err = ph.Write(pwb)
	//assert.NoError(t, err)
	//res = ph.Sum(nil)
	//assert.Equal(t, "cdde1eda231566cf3d59e81967227f1d775cdbf0c1ee2c0df30ef32050fd2913", hex.EncodeToString(res))
	//ph.Reset()
	//_, err = ph.Write(pwb)
	//assert.NoError(t, err)
	//res = ph.Sum(nil)
	//assert.Equal(t, "cdde1eda231566cf3d59e81967227f1d775cdbf0c1ee2c0df30ef32050fd2913", hex.EncodeToString(res))
	//ph.Reset()
	//_, err = ph.Write(pwb)
	//assert.NoError(t, err)
	//res = ph.Sum(nil)
	//assert.Equal(t, "cdde1eda231566cf3d59e81967227f1d775cdbf0c1ee2c0df30ef32050fd2913", hex.EncodeToString(res))
	//ph.Reset()
	//_, err = ph.Write(pwb)
	//assert.NoError(t, err)
	//res = ph.Sum(nil)
	//assert.Equal(t, "cdde1eda231566cf3d59e81967227f1d775cdbf0c1ee2c0df30ef32050fd2913", hex.EncodeToString(res))
	//ph.Reset()
	//_, err = ph.Write(pwb)
	//assert.NoError(t, err)
	//res = ph.Sum(nil)
	//assert.Equal(t, "cdde1eda231566cf3d59e81967227f1d775cdbf0c1ee2c0df30ef32050fd2913", hex.EncodeToString(res))
	//ph.Reset()
	//_, err = ph.Write(pwb)
	//assert.NoError(t, err)
	//res = ph.Sum(nil)
	//assert.Equal(t, "cdde1eda231566cf3d59e81967227f1d775cdbf0c1ee2c0df30ef32050fd2913", hex.EncodeToString(res))
	//ph.Reset()
}
