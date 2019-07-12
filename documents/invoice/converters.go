package invoice

import (
	"github.com/centrifuge/centrifuge-protobufs/gen/go/invoice"
	"github.com/centrifuge/go-centrifuge/documents"
)

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
