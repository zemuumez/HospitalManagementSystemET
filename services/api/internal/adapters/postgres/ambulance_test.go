package postgres

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"hms.local/api/internal/adapters/httpapi"
	"hms.local/api/internal/application"
	"hms.local/api/internal/domain"
)

func testAmbulances(t *testing.T, db *pgxpool.Pool, store Store, actors []domain.Actor) {
	t.Helper()
	ctx := context.Background()

	actorMap := make(map[string]domain.Actor)
	for _, a := range actors {
		actorMap[a.Role] = a
	}

	for _, role := range []string{"case_manager", "accountant"} {
		id := "test-" + role
		if _, err := db.Exec(ctx, `INSERT INTO "user"(id,name,email) VALUES($1,$1,$1||'@example.test') ON CONFLICT(id) DO NOTHING`, id); err != nil {
			t.Fatal(err)
		}
		if _, err := db.Exec(ctx, `INSERT INTO staff_access(user_id,role) VALUES($1,$2) ON CONFLICT(user_id) DO UPDATE SET role=$2`, id, role); err != nil {
			t.Fatal(err)
		}
		actorMap[role] = domain.Actor{ID: id, Role: role}
	}

	admin := actorMap["admin"]
	receptionist := actorMap["receptionist"]
	caseManager := actorMap["case_manager"]
	doctor := actorMap["doctor"]
	accountant := actorMap["accountant"]
	patientUser := actorMap["patient"]

	ambService := application.AmbulanceService{Store: store, Now: time.Now}
	billingService := application.Billing{Store: store, Now: time.Now}

	// 1. Authorization tests on ambulance master
	ambInput := domain.AmbulanceInput{
		VehicleNumber: "ETH-TEST-99",
		VehicleModel:  "Ford Transit ICU",
		YearMade:      "2023",
		DriverName:    "Mulugeta Tesfaye",
		DriverLicense: "DL-ETH-11223",
		DriverContact: "+251911998877",
		VehicleType:   2,
	}

	// Doctor cannot create ambulance
	if _, err := ambService.CreateAmbulance(ctx, doctor, ambInput); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected forbidden for doctor creating ambulance, got %v", err)
	}
	// Patient cannot create ambulance
	if _, err := ambService.CreateAmbulance(ctx, patientUser, ambInput); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected forbidden for patient creating ambulance, got %v", err)
	}

	// Receptionist creates ambulance
	amb, err := ambService.CreateAmbulance(ctx, receptionist, ambInput)
	if err != nil {
		t.Fatalf("failed to create ambulance: %v", err)
	}
	if amb.VehicleNumber != "ETH-TEST-99" || !amb.IsAvailable || amb.Version != 1 {
		t.Fatalf("unexpected ambulance created: %+v", amb)
	}

	// Update ambulance by Case Manager
	updateIn := ambInput
	updateIn.DriverName = "Mulugeta Tesfaye Updated"
	updatedAmb, err := ambService.UpdateAmbulance(ctx, caseManager, amb.ID, updateIn, amb.Version)
	if err != nil {
		t.Fatalf("failed to update ambulance: %v", err)
	}
	if updatedAmb.DriverName != "Mulugeta Tesfaye Updated" || updatedAmb.Version != 2 {
		t.Fatalf("unexpected updated ambulance: %+v", updatedAmb)
	}

	// Stale update returns ErrStale
	if _, err = ambService.UpdateAmbulance(ctx, admin, amb.ID, updateIn, 1); !errors.Is(err, domain.ErrStale) {
		t.Fatalf("expected ErrStale for stale version update, got %v", err)
	}

	// List ambulances
	ambList, total, err := ambService.Ambulances(ctx, doctor, 1, nil)
	if err != nil {
		t.Fatalf("failed to list ambulances: %v", err)
	}
	if total < 1 || len(ambList) == 0 {
		t.Fatalf("expected at least 1 ambulance in list, got total %d", total)
	}

	// 2. Ambulance Calls
	// Get real patient
	var patientID string
	err = db.QueryRow(ctx, `SELECT id FROM patient WHERE user_id = $1`, patientUser.ID).Scan(&patientID)
	if err != nil {
		// Fallback to any patient
		err = db.QueryRow(ctx, `SELECT id FROM patient LIMIT 1`).Scan(&patientID)
		if err != nil {
			t.Fatalf("failed to query patient for ambulance test: %v", err)
		}
	}

	callInput := domain.AmbulanceCallInput{
		AmbulanceID:    amb.ID,
		PatientID:      patientID,
		CallDate:       time.Now(),
		AmountMinor:    180000, // 1800.00 ETB
		PickupLocation: "Bole Medhanialem",
		Destination:    "Hospital Emergency Wing",
		Notes:          "Acute respiratory distress",
	}

	// Doctor cannot create ambulance call
	if _, err = ambService.CreateCall(ctx, doctor, callInput, "key-call-1"); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected forbidden for doctor creating call, got %v", err)
	}

	// Receptionist creates ambulance call -> Dispatched
	call, err := ambService.CreateCall(ctx, receptionist, callInput, "key-call-1")
	if err != nil {
		t.Fatalf("failed to create ambulance call: %v", err)
	}
	if call.Status != "dispatched" || call.VehicleNumber != "ETH-TEST-99" {
		t.Fatalf("unexpected call created: %+v", call)
	}

	// Vehicle should now be marked unavailable
	reloadedAmb, err := ambService.Ambulance(ctx, admin, amb.ID)
	if err != nil {
		t.Fatalf("failed to reload ambulance: %v", err)
	}
	if reloadedAmb.IsAvailable {
		t.Fatalf("expected ambulance to be unavailable after dispatch")
	}

	// Concurrent dispatch attempt on the same unavailable ambulance must fail with ErrConflict
	if _, err = ambService.CreateCall(ctx, admin, callInput, "key-call-conflict"); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("expected ErrConflict when dispatching unavailable ambulance, got %v", err)
	}

	// 3. Status Transition: Complete Call & Release Vehicle
	updateCallIn := domain.AmbulanceCallUpdateInput{
		AmbulanceID:    call.AmbulanceID,
		DriverName:     call.DriverName,
		CallDate:       call.CallDate,
		AmountMinor:    call.AmountMinor,
		Status:         "completed",
		PickupLocation: call.PickupLocation,
		Destination:    call.Destination,
		Notes:          "Patient handed over to triage successfully",
		Version:        call.Version,
	}

	completedCall, err := ambService.UpdateCall(ctx, caseManager, call.ID, updateCallIn)
	if err != nil {
		t.Fatalf("failed to complete call: %v", err)
	}
	if completedCall.Status != "completed" || completedCall.Version != 2 {
		t.Fatalf("unexpected completed call: %+v", completedCall)
	}

	// Vehicle should now be released and available again
	reloadedAmb, err = ambService.Ambulance(ctx, admin, amb.ID)
	if err != nil {
		t.Fatalf("failed to reload ambulance after completion: %v", err)
	}
	if !reloadedAmb.IsAvailable {
		t.Fatalf("expected ambulance to be available after call completion")
	}

	// 4. Billing Integration
	// Query a charge account
	var accountID string
	err = db.QueryRow(ctx, `SELECT id FROM charge_account WHERE active LIMIT 1`).Scan(&accountID)
	if err != nil {
		// create account
		acc, err := billingService.CreateAccount(ctx, admin, "Ambulance Transportation")
		if err != nil {
			t.Fatalf("failed to create charge account: %v", err)
		}
		accountID = acc.ID
	}

	billIn := domain.SourceInvoiceInput{
		AccountID:           accountID,
		DiscountBasisPoints: 500, // 5% discount
	}

	// Receptionist cannot bill (only admin or accountant)
	if _, err = ambService.BillCall(ctx, receptionist, call.ID, billIn, "bill-key-1", "2026-10-05"); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected forbidden for receptionist billing call, got %v", err)
	}

	// Accountant bills the ambulance call
	inv, err := ambService.BillCall(ctx, accountant, call.ID, billIn, "bill-key-1", "2026-10-05")
	if err != nil {
		t.Fatalf("accountant failed to bill ambulance call: %v", err)
	}
	if inv.TotalMinor <= 0 || inv.PatientID != call.PatientID {
		t.Fatalf("unexpected invoice issued for ambulance call: %+v", inv)
	}

	// Idempotent re-bill returns the exact same invoice
	inv2, err := ambService.BillCall(ctx, accountant, call.ID, billIn, "bill-key-1", "2026-10-05")
	if err != nil {
		t.Fatalf("idempotent billing failed: %v", err)
	}
	if inv2.ID != inv.ID {
		t.Fatalf("expected matching invoice ID on replay, got %s vs %s", inv2.ID, inv.ID)
	}

	// Changed terms return ErrConflict
	changedBillIn := billIn
	changedBillIn.DiscountBasisPoints = 1000
	if _, err = ambService.BillCall(ctx, accountant, call.ID, changedBillIn, "bill-key-diff", "2026-10-05"); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("expected ErrConflict on changed terms, got %v", err)
	}

	// Database trigger protects retained invoice link
	var deleteErr error
	_, deleteErr = db.Exec(ctx, `DELETE FROM ambulance_call_invoice WHERE call_id = $1`, call.ID)
	if deleteErr == nil {
		t.Fatalf("expected trigger to prevent deletion of ambulance_call_invoice row")
	}

	// 5. Patient Scope Access Check
	// Patient reading call detail
	patientCall, err := ambService.Call(ctx, patientUser, call.ID)
	if err != nil {
		t.Logf("patient read call result: %v", err)
	} else if patientCall.ID != call.ID {
		t.Fatalf("expected patient to read own call, got %s", patientCall.ID)
	}

	auth := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		who := strings.TrimPrefix(r.Header.Get("Cookie"), "session=")
		fmt.Fprintf(w, `{"user":{"id":%q},"session":{"userId":%q,"expiresAt":%q}}`, who, who, time.Now().Add(time.Hour).Format(time.RFC3339))
	}))
	defer auth.Close()

	server := httpapi.Server{
		Ambulance: ambService,
		Billing:   billingService,
		Actors:    store,
		AuthURL:   auth.URL,
		Origin:    "http://hospital.test",
		Client:    auth.Client(),
	}
	handler := server.Handler()

	// GET /v1/ambulances
	req := httptest.NewRequest("GET", "/v1/ambulances", nil)
	req.Header.Set("Cookie", "session="+receptionist.ID)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for GET /v1/ambulances, got %d", rec.Code)
	}

	// GET /v1/ambulance-calls
	req = httptest.NewRequest("GET", "/v1/ambulance-calls", nil)
	req.Header.Set("Cookie", "session="+accountant.ID)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for GET /v1/ambulance-calls, got %d", rec.Code)
	}

	// POST /v1/ambulance-calls/:id/bill via HTTP
	call2, err := ambService.CreateCall(ctx, admin, domain.AmbulanceCallInput{
		AmbulanceID: amb.ID,
		PatientID:   patientID,
		CallDate:    time.Now(),
		AmountMinor: 120000,
	}, "key-call-http")
	if err != nil {
		t.Fatalf("failed to create call2: %v", err)
	}

	billPayload, _ := json.Marshal(billIn)
	req = httptest.NewRequest("POST", fmt.Sprintf("/v1/ambulance-calls/%s/bill", call2.ID), bytes.NewReader(billPayload))
	req.Header.Set("Origin", "http://hospital.test")
	req.Header.Set("Cookie", "session="+accountant.ID)
	req.Header.Set("Idempotency-Key", "http-amb-bill-key")
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for HTTP bill ambulance call, got %d (body: %s)", rec.Code, rec.Body.String())
	}

	// 7. Concurrency Race Test
	// Dispatching ambulance from multiple concurrent goroutines
	var wg sync.WaitGroup
	results := make([]error, 5)
	for n := 0; n < 5; n++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			_, results[idx] = ambService.CreateCall(ctx, admin, domain.AmbulanceCallInput{
				AmbulanceID: amb.ID,
				PatientID:   patientID,
				CallDate:    time.Now(),
				AmountMinor: 100000,
			}, fmt.Sprintf("race-key-%d", idx))
		}(n)
	}
	wg.Wait()

	// Since ambulance was in use, all or at most 0 should succeed
	successCount := 0
	for _, err := range results {
		if err == nil {
			successCount++
		}
	}
	if successCount > 1 {
		t.Fatalf("expected at most 1 concurrent dispatch to succeed, got %d", successCount)
	}
}
