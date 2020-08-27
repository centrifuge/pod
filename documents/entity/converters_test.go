// +build unit

package entity

import (
	"testing"

	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"
)

func TestConverters_ToAddress(t *testing.T) {
	addresses := []Address{
		{
			Zip:   "12345",
			Label: "test",
		},

		{
			Zip:   "123455",
			Label: "test2",
		},
	}

	paddrs := toProtoAddresses(addresses)
	assert.Len(t, paddrs, 2)
	addrs := fromProtoAddresses(paddrs)
	assert.Equal(t, addresses, addrs)
}

func TestConverters_contacts(t *testing.T) {
	contacts := []Contact{
		{
			Title: "Home",
			Phone: "+123456789",
		},
	}

	pcontacts := toProtoContacts(contacts)
	assert.Len(t, pcontacts, 1)
	cts := fromProtoContacts(pcontacts)
	assert.Equal(t, contacts, cts)
}

func TestConverters_paymentDetails(t *testing.T) {
	pds := []PaymentDetail{
		{
			Predefined: true,
			BankPaymentMethod: &BankPaymentMethod{
				Identifier: utils.RandomSlice(32),
				Address: Address{
					Zip:   "12345",
					Label: "test",
				},
				HolderName: "John Doe",
			},
		},

		{
			CryptoPaymentMethod: &CryptoPaymentMethod{
				Identifier:        utils.RandomSlice(32),
				SupportedCurrency: "ERC20",
				To:                hexutil.Encode(utils.RandomSlice(32)),
			},
		},

		{
			OtherPaymentMethod: &OtherPaymentMethod{
				Identifier:        utils.RandomSlice(32),
				SupportedCurrency: "some currency",
			},
		},
	}

	ppds := toProtoPaymentDetails(pds)
	assert.Len(t, ppds, 3)
	apds := fromProtoPaymentDetails(ppds)
	assert.Equal(t, pds, apds)
}
