package domain

import (
	"testing"
	"time"
)

func TestPharmacyValidation(t *testing.T) {
	id := "00000000-0000-4000-8000-000000000001"
	for _, i := range []StockInput{
		{BatchID: id, Kind: "receipt", Quantity: 1},
		{BatchID: id, Kind: "dispense", Quantity: 1},
		{BatchID: id, Kind: "disposal", Quantity: 1},
		{BatchID: id, Kind: "return", OrderID: id, OriginalID: id, Quantity: 1},
		{BatchID: id, Kind: "dispense", OrderID: id, Quantity: -1},
		{BatchID: id, Kind: "dispense", OrderID: id, Quantity: 1000001},
	} {
		if i.Validate() == nil {
			t.Errorf("invalid stock input accepted: %+v", i)
		}
	}
	valid := StockInput{BatchID: id, OrderID: id, Kind: "dispense", Quantity: 1}
	if e := valid.Validate(); e != nil {
		t.Fatal(e)
	}
	order := MedicationInput{EncounterID: id, MedicineID: id, Quantity: 4, Dose: "1 tablet", Route: "oral", Frequency: "twice daily", DurationDays: 2}
	if e := order.Validate(); e != nil {
		t.Fatal(e)
	}
	order.DurationDays = 0
	if order.Validate() == nil {
		t.Fatal("zero duration accepted")
	}
	batch := BatchInput{MedicineID: id, Lot: "a", ExpiryDate: "2026-10-03", Supplier: "supplier", PurchaseReference: "ref", Quantity: 1}
	if batch.Validate(time.Date(2026, 10, 4, 0, 0, 0, 0, HospitalLocation)) == nil {
		t.Fatal("expired receipt accepted")
	}
	for _, role := range []string{"admin", "doctor", "patient", "nurse", "receptionist", "pharmacist", "accountant", "case_manager", "lab_technician"} {
		a := Actor{Role: role}
		if a.Can("pharmacy.manage") != (role == "admin" || role == "pharmacist") {
			t.Fatal("stock permission", role)
		}
		if a.Can("medication.read") != (role == "admin" || role == "pharmacist" || role == "doctor" || role == "patient") {
			t.Fatal("prescription permission", role)
		}
	}
}
