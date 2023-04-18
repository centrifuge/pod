package permissions

import (
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types/codec"
)

func deriveAccountID(poolID types.U64) (*types.AccountID, error) {
	typeID := []byte("pool")

	var t [4]byte

	copy(t[:], typeID)

	encodedTypeID, err := codec.Encode(t)

	if err != nil {
		return nil, err
	}

	encodedPoolID, err := codec.Encode(poolID)

	if err != nil {
		return nil, err
	}

	encodedData := appendTrailingZeroes(append(encodedTypeID, encodedPoolID...), 32)

	var accountID types.AccountID

	if err := codec.Decode(encodedData, &accountID); err != nil {
		return nil, err
	}

	return &accountID, nil
}

func appendTrailingZeroes(s []byte, maxSize uint) []byte {
	sliceLen := uint(len(s))

	if sliceLen >= maxSize {
		return nil
	}

	count := maxSize - sliceLen

	for i := uint(0); i < count; i++ {
		s = append(s, byte(0))
	}

	return s
}
