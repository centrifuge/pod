// +build tools

package build

import (
	_ "github.com/ethereum/go-ethereum/cmd/abigen"
	_ "github.com/goware/modvendor"
	_ "github.com/jteeuwen/go-bindata"
	_ "github.com/karalabe/xgo"
	_ "github.com/swaggo/swag/cmd/swag"
)
