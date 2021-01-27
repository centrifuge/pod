// +build unit integration testworld

package identity

import "github.com/stretchr/testify/mock"

type MockFactory struct {
	mock.Mock
	FactoryInterface
}
