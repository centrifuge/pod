package nft

import "github.com/centrifuge/go-centrifuge/errors"

const (
	// ErrInvalidParameter must be used if passed parameters are invalid
	ErrInvalidParameter = errors.Error("error with passed parameters.")

	// ErrInvalidAddress must be used if passed address is not an address
	ErrInvalidAddress = errors.Error("passed parameter is not an address")

	// ErrTokenTransfer must be used if token transfer transaction fails
	ErrTokenTransfer = errors.Error("token transfer transaction failed")

	// ErrOwnerOf must be used if an ownerOf calls to a NFT registry fails
	ErrOwnerOf = errors.Error("ownerOf call on NFT registry failed")
)
