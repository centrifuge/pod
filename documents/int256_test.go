package documents

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewInt256(t *testing.T) {
	tests := []struct {
		name  string
		val   string
		isErr bool
	}{
		{
			"zero",
			"0",
			false,
		},
		{
			"int256 max",
			"57896044618658097711785492504343953926634992332820282019728792003956564819967",
			false,
		},
		{
			"int256 min",
			"-57896044618658097711785492504343953926634992332820282019728792003956564819967",
			false,
		},
		{
			"allow bit length 256, but fails because max",
			"115792089237316195423570985008687907853269984665640564039457584007913129639935",
			true,
		},
		{
			"not allow bit length more than 256 = 257",
			"115792089237316195423570985008687907853269984665640564039457584007913129639936",
			true,
		},
		{
			"more than allowed int256 (+1)",
			"57896044618658097711785492504343953926634992332820282019728792003956564819968",
			true,
		},
		{
			"less than allowed int256 (-1)",
			"-57896044618658097711785492504343953926634992332820282019728792003956564819968",
			true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			i, err := NewInt256(test.val)
			if test.isErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, i)
				exp := new(big.Int)
				exp.SetString(test.val, 10)
				assert.True(t, i.v.Cmp(exp) == 0)
			}
		})
	}
}
