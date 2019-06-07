package purchaseorder

import (
	"github.com/centrifuge/centrifuge-protobufs/gen/go/purchaseorder"
	"github.com/centrifuge/go-centrifuge/documents"
	clientpb "github.com/centrifuge/go-centrifuge/protobufs/gen/go/purchaseorder"
	"github.com/centrifuge/go-centrifuge/utils/timeutils"
)

func toClientLineItems(items []*LineItem) ([]*clientpb.LineItem, error) {
	var citems []*clientpb.LineItem
	for _, i := range items {
		decs := documents.DecimalsToStrings(i.UnitOfMeasure, i.Quantity, i.PricePerUnit, i.AmountInvoiced, i.AmountTotal, i.ReceivedQuantity)
		pts, err := timeutils.ToProtoTimestamps(i.DateCreated, i.DateUpdated)
		if err != nil {
			return nil, err
		}

		activities, err := toClientActivities(i.Activities)
		if err != nil {
			return nil, err
		}

		citems = append(citems, &clientpb.LineItem{
			Status:            i.Status,
			Description:       i.Description,
			ItemNumber:        i.ItemNumber,
			UnitOfMeasure:     decs[0],
			Quantity:          decs[1],
			PricePerUnit:      decs[2],
			AmountInvoiced:    decs[3],
			AmountTotal:       decs[4],
			ReceivedQuantity:  decs[5],
			DateCreated:       pts[0],
			DateUpdated:       pts[1],
			PartNo:            i.PartNumber,
			RequisitionItem:   i.RequisitionItem,
			RequisitionNumber: i.RequisitionNumber,
			RevisionNumber:    int64(i.RevisionNumber),
			Activities:        activities,
			TaxItems:          toClientTaxItems(i.TaxItems),
		})
	}

	return citems, nil
}

func toClientActivities(activities []*LineItemActivity) ([]*clientpb.LineItemActivity, error) {
	var cactivities []*clientpb.LineItemActivity
	for _, a := range activities {
		pts, err := timeutils.ToProtoTimestamps(a.Date)
		if err != nil {
			return nil, err
		}

		decs := documents.DecimalsToStrings(a.Quantity, a.Amount)
		cactivities = append(cactivities, &clientpb.LineItemActivity{
			Quantity:              decs[0],
			Amount:                decs[1],
			ItemNumber:            a.ItemNumber,
			Status:                a.Status,
			Date:                  pts[0],
			ReferenceDocumentId:   a.ReferenceDocumentID,
			ReferenceDocumentItem: a.ReferenceDocumentItem,
		})
	}

	return cactivities, nil
}

func toClientTaxItems(items []*TaxItem) []*clientpb.TaxItem {
	var citems []*clientpb.TaxItem
	for _, i := range items {
		decs := documents.DecimalsToStrings(i.TaxAmount, i.TaxBaseAmount, i.TaxCode, i.TaxRate)
		citems = append(citems, &clientpb.TaxItem{
			ItemNumber:              i.ItemNumber,
			PurchaseOrderItemNumber: i.PurchaseOrderItemNumber,
			TaxAmount:               decs[0],
			TaxBaseAmount:           decs[1],
			TaxCode:                 decs[2],
			TaxRate:                 decs[3],
		})
	}

	return citems
}

func toP2PLineItems(items []*LineItem) ([]*purchaseorderpb.LineItem, error) {
	var pitems []*purchaseorderpb.LineItem
	for _, i := range items {
		decs, err := documents.DecimalsToBytes(i.UnitOfMeasure, i.Quantity, i.PricePerUnit, i.AmountInvoiced, i.AmountTotal, i.ReceivedQuantity)
		if err != nil {
			return nil, err
		}

		patts, err := toP2PActivities(i.Activities)
		if err != nil {
			return nil, err
		}

		pti, err := toP2PTaxItems(i.TaxItems)
		if err != nil {
			return nil, err
		}

		pts, err := timeutils.ToProtoTimestamps(i.DateCreated, i.DateUpdated)
		if err != nil {
			return nil, err
		}

		pitems = append(pitems, &purchaseorderpb.LineItem{
			Status:            i.Status,
			Description:       i.Description,
			ItemNumber:        i.ItemNumber,
			UnitOfMeasure:     decs[0],
			Quantity:          decs[1],
			PricePerUnit:      decs[2],
			AmountInvoiced:    decs[3],
			AmountTotal:       decs[4],
			ReceivedQuantity:  decs[5],
			DateCreated:       pts[0],
			DateUpdated:       pts[1],
			PartNo:            i.PartNumber,
			RequisitionItem:   i.RequisitionItem,
			RequisitionNumber: i.RequisitionNumber,
			RevisionNumber:    int64(i.RevisionNumber),
			Activities:        patts,
			TaxItems:          pti,
		})
	}

	return pitems, nil
}

func toP2PActivities(activities []*LineItemActivity) ([]*purchaseorderpb.LineItemActivity, error) {
	var pactivities []*purchaseorderpb.LineItemActivity
	for _, a := range activities {
		decs, err := documents.DecimalsToBytes(a.Quantity, a.Amount)
		if err != nil {
			return nil, err
		}

		pts, err := timeutils.ToProtoTimestamps(a.Date)
		if err != nil {
			return nil, err
		}

		pactivities = append(pactivities, &purchaseorderpb.LineItemActivity{
			Quantity:              decs[0],
			Amount:                decs[1],
			ItemNumber:            a.ItemNumber,
			Status:                a.Status,
			Date:                  pts[0],
			ReferenceDocumentId:   a.ReferenceDocumentID,
			ReferenceDocumentItem: a.ReferenceDocumentItem,
		})
	}

	return pactivities, nil
}

func toP2PTaxItems(items []*TaxItem) ([]*purchaseorderpb.TaxItem, error) {
	var pitems []*purchaseorderpb.TaxItem
	for _, i := range items {
		decs, err := documents.DecimalsToBytes(i.TaxAmount, i.TaxBaseAmount, i.TaxCode, i.TaxRate)
		if err != nil {
			return nil, err
		}
		pitems = append(pitems, &purchaseorderpb.TaxItem{
			ItemNumber:              i.ItemNumber,
			PurchaseOrderItemNumber: i.PurchaseOrderItemNumber,
			TaxAmount:               decs[0],
			TaxBaseAmount:           decs[1],
			TaxCode:                 decs[2],
			TaxRate:                 decs[3],
		})
	}

	return pitems, nil
}

func fromClientLineItems(citems []*clientpb.LineItem) ([]*LineItem, error) {
	var items []*LineItem
	for _, ci := range citems {
		decs, err := documents.StringsToDecimals(
			ci.AmountInvoiced,
			ci.AmountTotal,
			ci.PricePerUnit,
			ci.UnitOfMeasure,
			ci.Quantity,
			ci.ReceivedQuantity)
		if err != nil {
			return nil, err
		}

		ti, err := fromClientTaxItems(ci.TaxItems)
		if err != nil {
			return nil, err
		}

		la, err := fromClientLineItemActivities(ci.Activities)
		if err != nil {
			return nil, err
		}

		tms, err := timeutils.FromProtoTimestamps(ci.DateCreated, ci.DateUpdated)
		if err != nil {
			return nil, err
		}

		items = append(items, &LineItem{
			Status:            ci.Status,
			ItemNumber:        ci.ItemNumber,
			Description:       ci.Description,
			AmountInvoiced:    decs[0],
			AmountTotal:       decs[1],
			RequisitionNumber: ci.RequisitionNumber,
			RequisitionItem:   ci.RequisitionItem,
			RevisionNumber:    int(ci.RevisionNumber),
			PricePerUnit:      decs[2],
			UnitOfMeasure:     decs[3],
			Quantity:          decs[4],
			ReceivedQuantity:  decs[5],
			DateCreated:       tms[0],
			DateUpdated:       tms[1],
			PartNumber:        ci.PartNo,
			TaxItems:          ti,
			Activities:        la,
		})
	}

	return items, nil
}

func fromClientTaxItems(citems []*clientpb.TaxItem) ([]*TaxItem, error) {
	var items []*TaxItem
	for _, ci := range citems {
		decs, err := documents.StringsToDecimals(ci.TaxAmount, ci.TaxRate, ci.TaxCode, ci.TaxBaseAmount)
		if err != nil {
			return nil, err
		}

		items = append(items, &TaxItem{
			ItemNumber:              ci.ItemNumber,
			PurchaseOrderItemNumber: ci.PurchaseOrderItemNumber,
			TaxAmount:               decs[0],
			TaxRate:                 decs[1],
			TaxCode:                 decs[2],
			TaxBaseAmount:           decs[3],
		})
	}

	return items, nil
}

func fromClientLineItemActivities(catts []*clientpb.LineItemActivity) ([]*LineItemActivity, error) {
	var atts []*LineItemActivity
	for _, ca := range catts {
		decs, err := documents.StringsToDecimals(ca.Quantity, ca.Amount)
		if err != nil {
			return nil, err
		}

		tms, err := timeutils.FromProtoTimestamps(ca.Date)
		if err != nil {
			return nil, err
		}

		atts = append(atts, &LineItemActivity{
			ItemNumber:            ca.ItemNumber,
			Status:                ca.Status,
			Quantity:              decs[0],
			Amount:                decs[1],
			Date:                  tms[0],
			ReferenceDocumentItem: ca.ReferenceDocumentItem,
			ReferenceDocumentID:   ca.ReferenceDocumentId,
		})
	}

	return atts, nil
}

func fromP2PLineItemActivities(patts []*purchaseorderpb.LineItemActivity) ([]*LineItemActivity, error) {
	var atts []*LineItemActivity
	for _, ca := range patts {
		decs, err := documents.BytesToDecimals(ca.Quantity, ca.Amount)
		if err != nil {
			return nil, err
		}

		tms, err := timeutils.FromProtoTimestamps(ca.Date)
		if err != nil {
			return nil, err
		}

		atts = append(atts, &LineItemActivity{
			ItemNumber:            ca.ItemNumber,
			Status:                ca.Status,
			Quantity:              decs[0],
			Amount:                decs[1],
			Date:                  tms[0],
			ReferenceDocumentItem: ca.ReferenceDocumentItem,
			ReferenceDocumentID:   ca.ReferenceDocumentId,
		})
	}

	return atts, nil
}

func fromP2PTaxItems(pitems []*purchaseorderpb.TaxItem) ([]*TaxItem, error) {
	var items []*TaxItem
	for _, ci := range pitems {
		decs, err := documents.BytesToDecimals(ci.TaxAmount, ci.TaxRate, ci.TaxCode, ci.TaxBaseAmount)
		if err != nil {
			return nil, err
		}

		items = append(items, &TaxItem{
			ItemNumber:              ci.ItemNumber,
			PurchaseOrderItemNumber: ci.PurchaseOrderItemNumber,
			TaxAmount:               decs[0],
			TaxRate:                 decs[1],
			TaxCode:                 decs[2],
			TaxBaseAmount:           decs[3],
		})
	}

	return items, nil
}

func fromP2PLineItems(pitems []*purchaseorderpb.LineItem) ([]*LineItem, error) {
	var items []*LineItem
	for _, ci := range pitems {
		decs, err := documents.BytesToDecimals(
			ci.AmountInvoiced,
			ci.AmountTotal,
			ci.PricePerUnit,
			ci.UnitOfMeasure,
			ci.Quantity,
			ci.ReceivedQuantity)
		if err != nil {
			return nil, err
		}

		ti, err := fromP2PTaxItems(ci.TaxItems)
		if err != nil {
			return nil, err
		}

		la, err := fromP2PLineItemActivities(ci.Activities)
		if err != nil {
			return nil, err
		}

		tms, err := timeutils.FromProtoTimestamps(ci.DateCreated, ci.DateUpdated)
		if err != nil {
			return nil, err
		}

		items = append(items, &LineItem{
			Status:            ci.Status,
			ItemNumber:        ci.ItemNumber,
			Description:       ci.Description,
			AmountInvoiced:    decs[0],
			AmountTotal:       decs[1],
			RequisitionNumber: ci.RequisitionNumber,
			RequisitionItem:   ci.RequisitionItem,
			RevisionNumber:    int(ci.RevisionNumber),
			PricePerUnit:      decs[2],
			UnitOfMeasure:     decs[3],
			Quantity:          decs[4],
			ReceivedQuantity:  decs[5],
			DateCreated:       tms[0],
			DateUpdated:       tms[1],
			PartNumber:        ci.PartNo,
			TaxItems:          ti,
			Activities:        la,
		})
	}

	return items, nil
}
