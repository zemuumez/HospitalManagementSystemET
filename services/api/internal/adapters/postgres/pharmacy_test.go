package postgres

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"hms.local/api/internal/adapters/httpapi"
	"hms.local/api/internal/application"
	"hms.local/api/internal/domain"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

func testPharmacy(t *testing.T, db *pgxpool.Pool, store Store, actors []domain.Actor, cases []domain.Case) {
	t.Helper()
	ctx := context.Background()
	admin, doctor, other, patient := actors[0], actors[1], actors[2], actors[3]
	pharmacist := domain.Actor{ID: "pharmacist", Role: "pharmacist"}
	if _, e := db.Exec(ctx, `INSERT INTO "user"(id,name,email) VALUES('pharmacist','Pharmacist','pharmacist@example.test');INSERT INTO staff_access(user_id,role) VALUES('pharmacist','pharmacist')`); e != nil {
		t.Fatal(e)
	}
	now := time.Now()
	ph := application.Pharmacy{Store: store, Now: func() time.Time { return now }}
	clinic := application.Clinical{Store: store, Now: func() time.Time { return now }}
	enc, e := clinic.Admit(ctx, admin, domain.EncounterInput{Kind: "opd", CaseID: cases[0].ID, AdmittedAt: now.Add(-time.Minute)}, "pharmacy-encounter-0001")
	if e != nil {
		t.Fatal(e)
	}
	medInput := domain.MedicineInput{Name: "Synthetic medicine", Category: "Test category", Brand: "Test brand", Unit: "tablet", SellingPriceMinor: 100}
	med, e := ph.CreateMedicine(ctx, pharmacist, medInput)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = ph.CreateMedicine(ctx, doctor, medInput); !errors.Is(e, domain.ErrForbidden) {
		t.Fatal("doctor changed stock catalog", e)
	}
	batchInput := domain.BatchInput{MedicineID: med.ID, Lot: "LOT-1", ExpiryDate: now.AddDate(0, 1, 0).Format("2006-01-02"), Supplier: "Synthetic supplier", PurchaseReference: "TEST-PO-1", Quantity: 5, UnitCostMinor: 50}
	batch, e := ph.Receive(ctx, pharmacist, batchInput, "receipt-key-00000001")
	if e != nil || batch.Balance != 5 {
		t.Fatal("receipt", e, batch)
	}
	retry, e := ph.Receive(ctx, pharmacist, batchInput, "receipt-key-00000001")
	if e != nil || retry.ID != batch.ID {
		t.Fatal("receipt retry", e)
	}
	changed := batchInput
	changed.Quantity = 6
	if _, e = ph.Receive(ctx, pharmacist, changed, "receipt-key-00000001"); !errors.Is(e, domain.ErrConflict) {
		t.Fatal("changed receipt retry", e)
	}
	orderInput := domain.MedicationInput{EncounterID: enc.ID, MedicineID: med.ID, Quantity: 10, Dose: "1 tablet", Route: "oral", Frequency: "twice daily", DurationDays: 5, Instructions: "Synthetic test only"}
	if _, e = ph.Sign(ctx, other, orderInput, "order-key-00000001"); !errors.Is(e, domain.ErrNotFound) {
		t.Fatal("unassigned doctor prescribed", e)
	}
	if _, e = ph.Sign(ctx, admin, orderInput, "order-key-00000001"); !errors.Is(e, domain.ErrForbidden) {
		t.Fatal("admin prescribed", e)
	}
	order, e := ph.Sign(ctx, doctor, orderInput, "order-key-00000001")
	if e != nil {
		t.Fatal(e)
	}
	orderRetry, e := ph.Sign(ctx, doctor, orderInput, "order-key-00000001")
	if e != nil || orderRetry.ID != order.ID {
		t.Fatal("prescription retry", e)
	}
	for _, statement := range []string{`UPDATE medication_order SET dose='changed' WHERE id=$1`, `DELETE FROM medication_order WHERE id=$1`} {
		if _, e = db.Exec(ctx, statement, order.ID); e == nil {
			t.Fatal("signed prescription mutated")
		}
	}
	own, e := ph.Orders(ctx, patient, enc.ID, 1)
	if e != nil || len(own) != 1 {
		t.Fatal("own prescription visibility", e)
	}
	if _, e = ph.Orders(ctx, other, enc.ID, 1); !errors.Is(e, domain.ErrNotFound) {
		t.Fatal("other doctor saw prescription", e)
	}
	if _, e = ph.Orders(ctx, actors[4], enc.ID, 1); !errors.Is(e, domain.ErrForbidden) {
		t.Fatal("reception saw prescription", e)
	}
	issue := domain.StockInput{BatchID: batch.ID, OrderID: order.ID, Kind: "dispense", Quantity: 4}
	var wg sync.WaitGroup
	var issues [2]domain.StockMovement
	var errs [2]error
	for n := 0; n < 2; n++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			issues[n], errs[n] = ph.Move(ctx, pharmacist, issue, fmt.Sprintf("dispense-key-%08d", n))
		}(n)
	}
	wg.Wait()
	winner := 0
	if errs[0] != nil {
		winner = 1
	}
	if errs[winner] != nil || !errors.Is(errs[1-winner], domain.ErrStale) {
		t.Fatal("concurrent stock depletion", errs)
	}
	issued := issues[winner]
	issueRetry, e := ph.Move(ctx, pharmacist, issue, fmt.Sprintf("dispense-key-%08d", winner))
	if e != nil || issueRetry.ID != issued.ID {
		t.Fatal("duplicate dispensing", e)
	}
	returned := domain.StockInput{BatchID: batch.ID, OrderID: order.ID, Kind: "return", Quantity: 2, OriginalID: issued.ID, Reason: "Synthetic return"}
	ret, e := ph.Move(ctx, pharmacist, returned, "return-key-00000001")
	if e != nil {
		t.Fatal(e)
	}
	retRetry, e := ph.Move(ctx, pharmacist, returned, "return-key-00000001")
	if e != nil || retRetry.ID != ret.ID {
		t.Fatal("return retry", e)
	}
	returned.Quantity = 3
	if _, e = ph.Move(ctx, pharmacist, returned, "return-key-00000002"); !errors.Is(e, domain.ErrStale) {
		t.Fatal("excess return", e)
	}
	batches, e := ph.Batches(ctx, pharmacist, med.ID, 1)
	if e != nil || batches[0].Balance != 1 {
		t.Fatal("return incorrectly restocked", e, batches)
	}
	if _, e = db.Exec(ctx, `UPDATE medicine_batch SET balance=5 WHERE id=$1`, batch.ID); e == nil {
		t.Fatal("unreconciled balance accepted")
	}
	if _, e = db.Exec(ctx, `UPDATE medicine_batch SET expiry_date=CURRENT_DATE WHERE id=$1`, batch.ID); e == nil {
		t.Fatal("receipt snapshot mutated")
	}
	for _, statement := range []string{`UPDATE pharmacy_movement SET quantity=1 WHERE id=$1`, `DELETE FROM pharmacy_movement WHERE id=$1`} {
		if _, e = db.Exec(ctx, statement, issued.ID); e == nil {
			t.Fatal("movement history mutated")
		}
	}
	// A second batch must not bypass the total prescribed quantity.
	secondInput := batchInput
	secondInput.Lot = "LOT-2"
	secondInput.Quantity = 20
	second, e := ph.Receive(ctx, pharmacist, secondInput, "receipt-key-00000002")
	if e != nil {
		t.Fatal(e)
	}
	excess := domain.StockInput{BatchID: second.ID, OrderID: order.ID, Kind: "dispense", Quantity: 7}
	if _, e = ph.Move(ctx, pharmacist, excess, "excess-prescription-key"); !errors.Is(e, domain.ErrStale) {
		t.Fatal("over-prescription across batches", e)
	}
	// Returned units do not authorize additional supply.
	excess.Quantity = 8
	if _, e = ph.Move(ctx, pharmacist, excess, "return-refill-key-0001"); !errors.Is(e, domain.ErrStale) {
		t.Fatal("return reset prescribed allowance", e)
	}
	late := ph
	late.Now = func() time.Time { return now.AddDate(1, 0, 0) }
	issue.Quantity = 1
	if _, e = late.Move(ctx, pharmacist, issue, "expired-dispense-key"); !errors.Is(e, domain.ErrStale) {
		t.Fatal("expired stock dispensed", e)
	}
	if e = ph.Cancel(ctx, other, order.ID, "not my order"); !errors.Is(e, domain.ErrNotFound) {
		t.Fatal("other doctor cancelled", e)
	}
	if e = ph.Cancel(ctx, doctor, order.ID, "Stop remaining supply"); e != nil {
		t.Fatal(e)
	}
	if e = ph.Cancel(ctx, doctor, order.ID, "Stop remaining supply"); e != nil {
		t.Fatal("cancel retry", e)
	}
	if _, e = ph.Move(ctx, pharmacist, issue, "cancelled-dispense-key"); !errors.Is(e, domain.ErrStale) {
		t.Fatal("cancelled order dispensed", e)
	}
	if _, e = ph.Move(ctx, patient, issue, "patient-dispense-key"); !errors.Is(e, domain.ErrForbidden) {
		t.Fatal("patient moved stock", e)
	}
	// Disposal may remove expired/unusable stock without altering its signed history.
	if _, e = ph.Move(ctx, pharmacist, domain.StockInput{BatchID: batch.ID, Kind: "disposal", Quantity: 1, Reason: "Damaged"}, "disposal-key-00000001"); e != nil {
		t.Fatal(e)
	}
	movements, e := ph.Movements(ctx, pharmacist, batch.ID, 1)
	if e != nil || len(movements) != 4 {
		t.Fatal("movement audit count", e, len(movements))
	}
	auth := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		who := strings.TrimPrefix(r.Header.Get("Cookie"), "session=")
		fmt.Fprintf(w, `{"user":{"id":%q},"session":{"userId":%q,"expiresAt":%q}}`, who, who, time.Now().Add(time.Hour).Format(time.RFC3339))
	}))
	defer auth.Close()
	handler := httpapi.Server{Pharmacy: ph, Actors: store, AuthURL: auth.URL, Origin: "http://hospital.test", Client: auth.Client()}.Handler()
	for _, tc := range []struct {
		actor, method, path, body, origin string
		want                              int
	}{
		{"patient", "GET", "/v1/medicine-batches?medicineId=" + med.ID, "", "", 403},
		{"pharmacist", "GET", "/v1/medicine-batches?medicineId=" + med.ID, "", "", 200},
		{"pharmacist", "POST", "/v1/medicines", `{"name":"test","extra":true}`, "http://hospital.test", 400},
		{"pharmacist", "POST", "/v1/medicines", `{}`, "http://evil.test", 403},
		{"pharmacist", "POST", "/v1/pharmacy-movements", `{}`, "http://hospital.test", 422},
	} {
		req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
		req.Header.Set("Cookie", "session="+tc.actor)
		req.Header.Set("Origin", tc.origin)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		if w.Code != tc.want {
			t.Fatalf("%s want %d got %d: %s", tc.path, tc.want, w.Code, w.Body.String())
		}
	}
}
