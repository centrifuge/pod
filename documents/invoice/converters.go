package invoice

import (
	"strings"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/invoice"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/identity"
	clientpb "github.com/centrifuge/go-centrifuge/protobufs/gen/go/invoice"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

func toClientAttachments(atts []*BinaryAttachment) []*clientpb.BinaryAttachment {
	var catts []*clientpb.BinaryAttachment
	for _, att := range atts {
		var data, checksum string
		if len(att.Data) > 0 {
			data = hexutil.Encode(att.Data)
		}

		if len(att.Checksum) > 0 {
			checksum = hexutil.Encode(att.Checksum)
		}

		catts = append(catts, &clientpb.BinaryAttachment{
			Name:     att.Name,
			FileType: att.FileType,
			Size:     att.Size,
			Data:     data,
			Checksum: checksum,
		})
	}

	return catts
}

func toP2PAttachments(atts []*BinaryAttachment) []*invoicepb.BinaryAttachment {
	var patts []*invoicepb.BinaryAttachment
	for _, att := range atts {
		patts = append(patts, &invoicepb.BinaryAttachment{
			Name:     att.Name,
			FileType: att.FileType,
			Size:     att.Size,
			Data:     att.Data,
			Checksum: att.Checksum,
		})
	}

	return patts
}

func fromClientAttachments(catts []*clientpb.BinaryAttachment) ([]*BinaryAttachment, error) {
	var atts []*BinaryAttachment
	for _, att := range catts {
		var data, checksum []byte
		var err error
		if s := strings.TrimSpace(att.Data); s != "" {
			data, err = hexutil.Decode(s)
			if err != nil {
				return nil, err
			}
		}

		if s := strings.TrimSpace(att.Checksum); s != "" {
			checksum, err = hexutil.Decode(s)
			if err != nil {
				return nil, err
			}
		}

		atts = append(atts, &BinaryAttachment{
			Name:     att.Name,
			FileType: att.FileType,
			Size:     att.Size,
			Data:     data,
			Checksum: checksum,
		})
	}

	return atts, nil
}

func fromP2PAttachments(patts []*invoicepb.BinaryAttachment) []*BinaryAttachment {
	var atts []*BinaryAttachment
	for _, att := range patts {
		atts = append(atts, &BinaryAttachment{
			Name:     att.Name,
			FileType: att.FileType,
			Size:     att.Size,
			Data:     att.Data,
			Checksum: att.Checksum,
		})
	}

	return atts
}

func toClientLineItems(items []*LineItem) []*clientpb.InvoiceLineItem {
	var citems []*clientpb.InvoiceLineItem
	for _, item := range items {
		decs := documents.DecimalsToStrings(
			item.PricePerUnit, item.Quantity, item.NetWeight, item.TaxAmount, item.TaxRate, item.TaxCode, item.TotalAmount)

		citems = append(citems, &clientpb.InvoiceLineItem{
			ItemNumber:              item.ItemNumber,
			Description:             item.Description,
			SenderPartNo:            item.SenderPartNo,
			PricePerUnit:            decs[0],
			Quantity:                decs[1],
			UnitOfMeasure:           item.UnitOfMeasure,
			NetWeight:               decs[2],
			TaxAmount:               decs[3],
			TaxRate:                 decs[4],
			TaxCode:                 decs[5],
			TotalAmount:             decs[6],
			PurchaseOrderNumber:     item.PurchaseOrderNumber,
			PurchaseOrderItemNumber: item.PurchaseOrderItemNumber,
			DeliveryNoteNumber:      item.DeliveryNoteNumber,
		})
	}

	return citems
}

func toP2PLineItems(items []*LineItem) ([]*invoicepb.InvoiceLineItem, error) {
	var pitems []*invoicepb.InvoiceLineItem
	for _, item := range items {
		decs, err := documents.DecimalsToBytes(
			item.PricePerUnit, item.Quantity, item.NetWeight, item.TaxAmount, item.TaxRate, item.TaxCode, item.TotalAmount)
		if err != nil {
			return nil, err
		}

		pitems = append(pitems, &invoicepb.InvoiceLineItem{
			ItemNumber:              item.ItemNumber,
			Description:             item.Description,
			SenderPartNo:            item.SenderPartNo,
			PricePerUnit:            decs[0],
			Quantity:                decs[1],
			UnitOfMeasure:           item.UnitOfMeasure,
			NetWeight:               decs[2],
			TaxAmount:               decs[3],
			TaxRate:                 decs[4],
			TaxCode:                 decs[5],
			TotalAmount:             decs[6],
			PurchaseOrderNumber:     item.PurchaseOrderNumber,
			PurchaseOrderItemNumber: item.PurchaseOrderItemNumber,
			DeliveryNoteNumber:      item.DeliveryNoteNumber,
		})
	}

	return pitems, nil
}

func fromClientLineItems(citems []*clientpb.InvoiceLineItem) ([]*LineItem, error) {
	var items []*LineItem
	for _, item := range citems {
		decs, err := documents.StringsToDecimals(
			item.PricePerUnit, item.Quantity, item.NetWeight, item.TaxAmount, item.TaxRate, item.TaxCode, item.TotalAmount)
		if err != nil {
			return nil, err
		}

		items = append(items, &LineItem{
			ItemNumber:              item.ItemNumber,
			Description:             item.Description,
			SenderPartNo:            item.SenderPartNo,
			PricePerUnit:            decs[0],
			Quantity:                decs[1],
			UnitOfMeasure:           item.UnitOfMeasure,
			NetWeight:               decs[2],
			TaxAmount:               decs[3],
			TaxRate:                 decs[4],
			TaxCode:                 decs[5],
			TotalAmount:             decs[6],
			PurchaseOrderNumber:     item.PurchaseOrderNumber,
			PurchaseOrderItemNumber: item.PurchaseOrderItemNumber,
			DeliveryNoteNumber:      item.DeliveryNoteNumber,
		})
	}

	return items, nil
}

func fromP2PLineItems(pitems []*invoicepb.InvoiceLineItem) ([]*LineItem, error) {
	var items []*LineItem
	for _, item := range pitems {
		decs, err := documents.BytesToDecimals(
			item.PricePerUnit, item.Quantity, item.NetWeight, item.TaxAmount, item.TaxRate, item.TaxCode, item.TotalAmount)
		if err != nil {
			return nil, err
		}

		items = append(items, &LineItem{
			ItemNumber:              item.ItemNumber,
			Description:             item.Description,
			SenderPartNo:            item.SenderPartNo,
			PricePerUnit:            decs[0],
			Quantity:                decs[1],
			UnitOfMeasure:           item.UnitOfMeasure,
			NetWeight:               decs[2],
			TaxAmount:               decs[3],
			TaxRate:                 decs[4],
			TaxCode:                 decs[5],
			TotalAmount:             decs[6],
			PurchaseOrderNumber:     item.PurchaseOrderNumber,
			PurchaseOrderItemNumber: item.PurchaseOrderItemNumber,
			DeliveryNoteNumber:      item.DeliveryNoteNumber,
		})
	}

	return items, nil
}

func toClientPaymentDetails(details []*PaymentDetails) []*clientpb.PaymentDetails {
	var cdetails []*clientpb.PaymentDetails
	for _, detail := range details {
		decs := documents.DecimalsToStrings(detail.Amount)
		dids := identity.DIDsToStrings(detail.Payee, detail.Payer)
		cdetails = append(cdetails, &clientpb.PaymentDetails{
			Id:                    detail.ID,
			DateExecuted:          detail.DateExecuted,
			Payee:                 dids[0],
			Payer:                 dids[1],
			Amount:                decs[0],
			Currency:              detail.Currency,
			Reference:             detail.Reference,
			BankName:              detail.BankName,
			BankAddress:           detail.BankAddress,
			BankAccountCurrency:   detail.BankAccountCurrency,
			BankAccountHolderName: detail.BankAccountHolderName,
			BankAccountNumber:     detail.BankAccountNumber,
			BankCountry:           detail.BankCountry,
			BankKey:               detail.BankKey,
			CryptoChainUri:        detail.CryptoChainURI,
			CryptoFrom:            detail.CryptoFrom,
			CryptoTo:              detail.CryptoTo,
			CryptoTransactionId:   detail.CryptoTransactionID,
		})
	}

	return cdetails
}

func toP2PPaymentDetails(details []*PaymentDetails) ([]*invoicepb.PaymentDetails, error) {
	var pdetails []*invoicepb.PaymentDetails
	for _, detail := range details {
		decs, err := documents.DecimalsToBytes(detail.Amount)
		if err != nil {
			return nil, err
		}
		dids := identity.DIDsToBytes(detail.Payee, detail.Payer)
		pdetails = append(pdetails, &invoicepb.PaymentDetails{
			Id:                    detail.ID,
			DateExecuted:          detail.DateExecuted,
			Payee:                 dids[0],
			Payer:                 dids[1],
			Amount:                decs[0],
			Currency:              detail.Currency,
			Reference:             detail.Reference,
			BankName:              detail.BankName,
			BankAddress:           detail.BankAddress,
			BankAccountCurrency:   detail.BankAccountCurrency,
			BankAccountHolderName: detail.BankAccountHolderName,
			BankAccountNumber:     detail.BankAccountNumber,
			BankCountry:           detail.BankCountry,
			BankKey:               detail.BankKey,
			CryptoChainUri:        detail.CryptoChainURI,
			CryptoFrom:            detail.CryptoFrom,
			CryptoTo:              detail.CryptoTo,
			CryptoTransactionId:   detail.CryptoTransactionID,
		})
	}

	return pdetails, nil
}

func fromClientPaymentDetails(cdetails []*clientpb.PaymentDetails) ([]*PaymentDetails, error) {
	var details []*PaymentDetails
	for _, detail := range cdetails {
		decs, err := documents.StringsToDecimals(detail.Amount)
		if err != nil {
			return nil, err
		}

		dids, err := identity.StringsToDIDs(detail.Payee, detail.Payer)
		if err != nil {
			return nil, err
		}

		details = append(details, &PaymentDetails{
			ID:                    detail.Id,
			DateExecuted:          detail.DateExecuted,
			Payee:                 dids[0],
			Payer:                 dids[1],
			Amount:                decs[0],
			Currency:              detail.Currency,
			Reference:             detail.Reference,
			BankName:              detail.BankName,
			BankAddress:           detail.BankAddress,
			BankAccountCurrency:   detail.BankAccountCurrency,
			BankAccountHolderName: detail.BankAccountHolderName,
			BankAccountNumber:     detail.BankAccountNumber,
			BankCountry:           detail.BankCountry,
			BankKey:               detail.BankKey,
			CryptoChainURI:        detail.CryptoChainUri,
			CryptoFrom:            detail.CryptoFrom,
			CryptoTo:              detail.CryptoTo,
			CryptoTransactionID:   detail.CryptoTransactionId,
		})
	}

	return details, nil
}

func fromP2PPaymentDetails(pdetails []*invoicepb.PaymentDetails) ([]*PaymentDetails, error) {
	var details []*PaymentDetails
	for _, detail := range pdetails {
		decs, err := documents.BytesToDecimals(detail.Amount)
		if err != nil {
			return nil, err
		}
		dids := identity.BytesToDIDs(detail.Payee, detail.Payer)
		details = append(details, &PaymentDetails{
			ID:                    detail.Id,
			DateExecuted:          detail.DateExecuted,
			Payee:                 dids[0],
			Payer:                 dids[1],
			Amount:                decs[0],
			Currency:              detail.Currency,
			Reference:             detail.Reference,
			BankName:              detail.BankName,
			BankAddress:           detail.BankAddress,
			BankAccountCurrency:   detail.BankAccountCurrency,
			BankAccountHolderName: detail.BankAccountHolderName,
			BankAccountNumber:     detail.BankAccountNumber,
			BankCountry:           detail.BankCountry,
			BankKey:               detail.BankKey,
			CryptoChainURI:        detail.CryptoChainUri,
			CryptoFrom:            detail.CryptoFrom,
			CryptoTo:              detail.CryptoTo,
			CryptoTransactionID:   detail.CryptoTransactionId,
		})
	}

	return details, nil
}

func toClientTaxItems(items []*TaxItem) []*clientpb.InvoiceTaxItem {
	var citems []*clientpb.InvoiceTaxItem
	for _, item := range items {
		decs := documents.DecimalsToStrings(item.TaxAmount, item.TaxRate, item.TaxCode, item.TaxBaseAmount)
		citems = append(citems, &clientpb.InvoiceTaxItem{
			ItemNumber:        item.ItemNumber,
			InvoiceItemNumber: item.InvoiceItemNumber,
			TaxAmount:         decs[0],
			TaxRate:           decs[1],
			TaxCode:           decs[2],
			TaxBaseAmount:     decs[3],
		})
	}

	return citems
}

func toP2PTaxItems(items []*TaxItem) ([]*invoicepb.InvoiceTaxItem, error) {
	var pitems []*invoicepb.InvoiceTaxItem
	for _, item := range items {
		decs, err := documents.DecimalsToBytes(item.TaxAmount, item.TaxRate, item.TaxCode, item.TaxBaseAmount)
		if err != nil {
			return nil, err
		}

		pitems = append(pitems, &invoicepb.InvoiceTaxItem{
			ItemNumber:        item.ItemNumber,
			InvoiceItemNumber: item.InvoiceItemNumber,
			TaxAmount:         decs[0],
			TaxRate:           decs[1],
			TaxCode:           decs[2],
			TaxBaseAmount:     decs[3],
		})
	}

	return pitems, nil
}

func fromClientTaxItems(citems []*clientpb.InvoiceTaxItem) ([]*TaxItem, error) {
	var items []*TaxItem
	for _, item := range citems {
		decs, err := documents.StringsToDecimals(item.TaxAmount, item.TaxRate, item.TaxCode, item.TaxBaseAmount)
		if err != nil {
			return nil, err
		}

		items = append(items, &TaxItem{
			ItemNumber:        item.ItemNumber,
			InvoiceItemNumber: item.InvoiceItemNumber,
			TaxAmount:         decs[0],
			TaxRate:           decs[1],
			TaxCode:           decs[2],
			TaxBaseAmount:     decs[3],
		})
	}

	return items, nil
}

func fromP2PTaxItems(pitems []*invoicepb.InvoiceTaxItem) ([]*TaxItem, error) {
	var items []*TaxItem
	for _, item := range pitems {
		decs, err := documents.BytesToDecimals(item.TaxAmount, item.TaxRate, item.TaxCode, item.TaxBaseAmount)
		if err != nil {
			return nil, err
		}

		items = append(items, &TaxItem{
			ItemNumber:        item.ItemNumber,
			InvoiceItemNumber: item.InvoiceItemNumber,
			TaxAmount:         decs[0],
			TaxRate:           decs[1],
			TaxCode:           decs[2],
			TaxBaseAmount:     decs[3],
		})
	}

	return items, nil
}
