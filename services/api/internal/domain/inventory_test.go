package domain

import "testing"

func TestInventoryValidation(t *testing.T) {
	id := "00000000-0000-4000-8000-000000000001"
	for _, input := range []InventoryMovementInput{
		{ItemID: id, Kind: "receive", QuantityMilli: 1, Reason: "Test"},
		{ItemID: id, Kind: "issue", QuantityMilli: 1, Reason: "Test"},
		{ItemID: id, Kind: "return", QuantityMilli: 1, OriginalID: id, RecipientID: "forged", Reason: "Test"},
		{ItemID: id, Kind: "writeoff", QuantityMilli: 0, Reason: "Test"},
		{ItemID: id, Kind: "receive", QuantityMilli: 1000000000001, Reference: "x", Reason: "Test"},
	} {
		if input.Validate() == nil {
			t.Fatal("invalid inventory movement accepted", input)
		}
	}
	valid := InventoryMovementInput{ItemID: id, Kind: "receive", QuantityMilli: 1250, Reference: "PO-1", Reason: "Test"}
	if e := valid.Validate(); e != nil {
		t.Fatal(e)
	}
}
