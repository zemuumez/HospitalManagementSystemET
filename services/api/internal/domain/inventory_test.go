package domain

import (
	"encoding/json"
	"testing"
)

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

func TestInventoryJSONUnmarshaling(t *testing.T) {
	// Floating-point values from client-side multiplication should round cleanly
	payload := `{"itemId":"00000000-0000-4000-8000-000000000001","kind":"receive","quantityMilli":1140.0000000000002,"costMinor":2550.0000000000005,"reference":"PO-1","reason":"Replenishment"}`
	var in InventoryMovementInput
	if err := json.Unmarshal([]byte(payload), &in); err != nil {
		t.Fatalf("unexpected unmarshal error: %v", err)
	}
	if in.QuantityMilli != 1140 {
		t.Fatalf("expected quantityMilli=1140, got %d", in.QuantityMilli)
	}
	if in.CostMinor != 2550 {
		t.Fatalf("expected costMinor=2550, got %d", in.CostMinor)
	}

	// Unknown fields must be rejected
	badPayload := `{"itemId":"00000000-0000-4000-8000-000000000001","unknown":true}`
	var bad InventoryMovementInput
	if err := json.Unmarshal([]byte(badPayload), &bad); err == nil {
		t.Fatal("expected error on unknown field, got nil")
	}

	// InventoryItemInput floating-point reorderMilli
	itemPayload := `{"categoryId":"00000000-0000-4000-8000-000000000001","name":"Syringe","unit":"piece","description":"","reorderMilli":2560.0000000000005,"active":true,"version":1}`
	var itemIn InventoryItemInput
	if err := json.Unmarshal([]byte(itemPayload), &itemIn); err != nil {
		t.Fatalf("unexpected unmarshal error on item: %v", err)
	}
	if itemIn.ReorderMilli != 2560 {
		t.Fatalf("expected reorderMilli=2560, got %d", itemIn.ReorderMilli)
	}
}
