// +build unit

package networks

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestViperNetworkConfigurationLoader_LoadNetworkConfig(t *testing.T) {
	cl := NewViperNetworkConfigurationLoader()
	err := cl.LoadNetworkConfig()
	assert.Nil(t, err)
	fmt.Println(cl.networksConfig)

	assert.Equal(t, 4, cl.networksConfig.GetInt("centrifuge-russianhill-eth-rinkeby.ethereumNetworkId"))
}
