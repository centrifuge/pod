package entity

import entitypb "github.com/centrifuge/centrifuge-protobufs/gen/go/entity"

func toProtoAddresses(addrs []Address) (paddrs []*entitypb.Address) {
	for _, addr := range addrs {
		paddrs = append(paddrs, &entitypb.Address{
			State:         addr.State,
			AddressLine1:  addr.AddressLine1,
			AddressLine2:  addr.AddressLine2,
			ContactPerson: addr.ContactPerson,
			Country:       addr.Country,
			IsMain:        addr.IsMain,
			IsPayTo:       addr.IsPayTo,
			IsRemitTo:     addr.IsRemitTo,
			IsShipTo:      addr.IsShipTo,
			Label:         addr.Label,
			Zip:           addr.Zip,
		})
	}

	return paddrs
}

func fromProtoAddresses(paddrs []*entitypb.Address) (addrs []Address) {
	for _, addr := range paddrs {
		addrs = append(addrs, Address{
			State:         addr.State,
			AddressLine1:  addr.AddressLine1,
			AddressLine2:  addr.AddressLine2,
			ContactPerson: addr.ContactPerson,
			Country:       addr.Country,
			IsMain:        addr.IsMain,
			IsPayTo:       addr.IsPayTo,
			IsRemitTo:     addr.IsRemitTo,
			IsShipTo:      addr.IsShipTo,
			Label:         addr.Label,
			Zip:           addr.Zip,
		})
	}

	return addrs
}

func toProtoContacts(contacts []Contact) (pcontacts []*entitypb.Contact) {
	for _, contact := range contacts {
		pcontacts = append(pcontacts, &entitypb.Contact{
			Name:  contact.Name,
			Email: contact.Email,
			Fax:   contact.Fax,
			Phone: contact.Phone,
			Title: contact.Title,
		})
	}

	return pcontacts
}

func fromProtoContacts(pcontacts []*entitypb.Contact) (contacts []Contact) {
	for _, contact := range pcontacts {
		contacts = append(contacts, Contact{
			Name:  contact.Name,
			Email: contact.Email,
			Fax:   contact.Fax,
			Phone: contact.Phone,
			Title: contact.Title,
		})
	}

	return contacts
}

func toProtoBankPaymentMethod(bpm *BankPaymentMethod) *entitypb.BankPaymentMethod {
	if bpm == nil {
		return nil
	}

	return &entitypb.BankPaymentMethod{
		Address:           toProtoAddresses([]Address{bpm.Address})[0],
		Identifier:        bpm.Identifier,
		BankAccountNumber: bpm.BankAccountNumber,
		BankKey:           bpm.BankKey,
		HolderName:        bpm.HolderName,
		SupportedCurrency: bpm.SupportedCurrency,
	}
}

func fromProtoBankPaymentMethod(pbpm *entitypb.BankPaymentMethod) *BankPaymentMethod {
	if pbpm == nil {
		return nil
	}

	return &BankPaymentMethod{
		SupportedCurrency: pbpm.SupportedCurrency,
		HolderName:        pbpm.HolderName,
		BankKey:           pbpm.BankKey,
		BankAccountNumber: pbpm.BankAccountNumber,
		Identifier:        pbpm.Identifier,
		Address:           fromProtoAddresses([]*entitypb.Address{pbpm.Address})[0],
	}
}

func toProtoCryptoPaymentMethod(cpm *CryptoPaymentMethod) *entitypb.CryptoPaymentMethod {
	if cpm == nil {
		return nil
	}

	return &entitypb.CryptoPaymentMethod{
		Identifier:        cpm.Identifier,
		SupportedCurrency: cpm.SupportedCurrency,
		ChainUri:          cpm.ChainURI,
		To:                cpm.To,
	}
}

func fromProtoCryptoPaymentMethod(pcpm *entitypb.CryptoPaymentMethod) *CryptoPaymentMethod {
	if pcpm == nil {
		return nil
	}

	return &CryptoPaymentMethod{
		Identifier:        pcpm.Identifier,
		SupportedCurrency: pcpm.SupportedCurrency,
		ChainURI:          pcpm.ChainUri,
		To:                pcpm.To,
	}
}

func toProtoOtherPaymentMethod(opm *OtherPaymentMethod) *entitypb.OtherPayment {
	if opm == nil {
		return nil
	}

	return &entitypb.OtherPayment{
		SupportedCurrency: opm.SupportedCurrency,
		Identifier:        opm.Identifier,
		Type:              opm.Type,
		PayTo:             opm.PayTo,
	}
}

func fromProtoOtherPaymentMethod(popm *entitypb.OtherPayment) *OtherPaymentMethod {
	if popm == nil {
		return nil
	}

	return &OtherPaymentMethod{
		PayTo:             popm.PayTo,
		Type:              popm.Type,
		Identifier:        popm.Identifier,
		SupportedCurrency: popm.SupportedCurrency,
	}
}

func toProtoPaymentDetail(pd PaymentDetail) *entitypb.PaymentDetail {
	ppd := &entitypb.PaymentDetail{Predefined: pd.Predefined}
	if pd.BankPaymentMethod != nil {
		ppd.PaymentMethod = &entitypb.PaymentDetail_BankPaymentMethod{BankPaymentMethod: toProtoBankPaymentMethod(pd.BankPaymentMethod)}
		return ppd
	}

	if pd.CryptoPaymentMethod != nil {
		ppd.PaymentMethod = &entitypb.PaymentDetail_CryptoPaymentMethod{CryptoPaymentMethod: toProtoCryptoPaymentMethod(pd.CryptoPaymentMethod)}
		return ppd
	}

	ppd.PaymentMethod = &entitypb.PaymentDetail_OtherMethod{OtherMethod: toProtoOtherPaymentMethod(pd.OtherPaymentMethod)}
	return ppd
}

func fromProtoPaymentDetail(ppd *entitypb.PaymentDetail) PaymentDetail {
	pd := PaymentDetail{Predefined: ppd.Predefined}
	if ppd.GetBankPaymentMethod() != nil {
		pd.BankPaymentMethod = fromProtoBankPaymentMethod(ppd.GetBankPaymentMethod())
		return pd
	}

	if ppd.GetCryptoPaymentMethod() != nil {
		pd.CryptoPaymentMethod = fromProtoCryptoPaymentMethod(ppd.GetCryptoPaymentMethod())
		return pd
	}

	pd.OtherPaymentMethod = fromProtoOtherPaymentMethod(ppd.GetOtherMethod())
	return pd
}

func toProtoPaymentDetails(pds []PaymentDetail) (ppds []*entitypb.PaymentDetail) {
	for _, pd := range pds {
		ppds = append(ppds, toProtoPaymentDetail(pd))
	}

	return ppds
}

func fromProtoPaymentDetails(ppds []*entitypb.PaymentDetail) (pds []PaymentDetail) {
	for _, ppd := range ppds {
		pds = append(pds, fromProtoPaymentDetail(ppd))
	}

	return pds
}
