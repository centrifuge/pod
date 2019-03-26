package invoice

import (
	"github.com/centrifuge/centrifuge-protobufs/gen/go/invoice"
	"github.com/centrifuge/go-centrifuge/documents"
	clientpb "github.com/centrifuge/go-centrifuge/protobufs/gen/go/invoice"
)

func toClientLineItems(items []*LineItem) []*clientpb.LineItem {
	var citems []*clientpb.LineItem
	for _, item := range items {
		decs := documents.DecimalsToStrings(
			item.PricePerUnit, item.Quantity, item.NetWeight, item.TaxAmount, item.TaxRate, item.TaxCode, item.TotalAmount)

		citems = append(citems, &clientpb.LineItem{
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

func toP2PLineItems(items []*LineItem) ([]*invoicepb.LineItem, error) {
	var pitems []*invoicepb.LineItem
	for _, item := range items {
		decs, err := documents.DecimalsToBytes(
			item.PricePerUnit, item.Quantity, item.NetWeight, item.TaxAmount, item.TaxRate, item.TaxCode, item.TotalAmount)
		if err != nil {
			return nil, err
		}

		pitems = append(pitems, &invoicepb.LineItem{
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

func fromClientLineItems(citems []*clientpb.LineItem) ([]*LineItem, error) {
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

func fromP2PLineItems(pitems []*invoicepb.LineItem) ([]*LineItem, error) {
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

func toClientTaxItems(items []*TaxItem) []*clientpb.TaxItem {
	var citems []*clientpb.TaxItem
	for _, item := range items {
		decs := documents.DecimalsToStrings(item.TaxAmount, item.TaxRate, item.TaxCode, item.TaxBaseAmount)
		citems = append(citems, &clientpb.TaxItem{
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

func toP2PTaxItems(items []*TaxItem) ([]*invoicepb.TaxItem, error) {
	var pitems []*invoicepb.TaxItem
	for _, item := range items {
		decs, err := documents.DecimalsToBytes(item.TaxAmount, item.TaxRate, item.TaxCode, item.TaxBaseAmount)
		if err != nil {
			return nil, err
		}

		pitems = append(pitems, &invoicepb.TaxItem{
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

func fromClientTaxItems(citems []*clientpb.TaxItem) ([]*TaxItem, error) {
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

func fromP2PTaxItems(pitems []*invoicepb.TaxItem) ([]*TaxItem, error) {
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
