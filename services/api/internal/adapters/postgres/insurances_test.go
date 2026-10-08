package postgres

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"hms.local/api/internal/adapters/httpapi"
	"hms.local/api/internal/application"
	"hms.local/api/internal/domain"
)

func testInsurances(t *testing.T, db *pgxpool.Pool, store Store, actors []domain.Actor, cases []domain.Case) {
	t.Helper()
	ctx := context.Background()

	actorMap := make(map[string]domain.Actor)
	for _, a := range actors {
		actorMap[a.Role] = a
	}

	admin := actorMap["admin"]
	receptionist := actorMap["receptionist"]
	doctor := actorMap["doctor"]
	patientUser := actorMap["patient"]
	nurse := domain.Actor{ID: "nurse", Role: "nurse"}
	accountant := domain.Actor{ID: "accountant", Role: "accountant"}

	insurancesService := application.InsurancesService{Store: store, Now: time.Now}

	// 1. Create Insurance with multiple diseases and exact totals
	insIn := domain.InsuranceInput{
		Name:              "National Health Insurance Corp",
		ServiceTaxMinor:   5000,  // 50.00 ETB
		HospitalRateMinor: 15000, // 150.00 ETB
		Discount:          10,    // 10%
		Remark:            "National insurance policy",
		InsuranceNo:       "NHIC-00918",
		InsuranceCode:     "NAT-HLTH-01",
		Diseases: []domain.InsuranceDiseaseLineInput{
			{DiseaseName: "Malaria Treatment", DiseaseChargeMinor: 4500}, // 45.00 ETB
			{DiseaseName: "Typhoid Care", DiseaseChargeMinor: 5500},      // 55.00 ETB
		},
	}

	createdIns, err := insurancesService.CreateInsurance(ctx, admin, insIn)
	if err != nil {
		t.Fatalf("failed to create insurance as admin: %v", err)
	}

	if createdIns.ID == "" {
		t.Fatal("expected non-empty insurance ID")
	}
	if createdIns.TotalMinor != 27000 || createdIns.Total != 270.00 {
		t.Fatalf("expected total 27000 minor (270.00 ETB), got minor=%d total=%.2f", createdIns.TotalMinor, createdIns.Total)
	}
	if len(createdIns.Diseases) != 2 {
		t.Fatalf("expected 2 disease lines, got %d", len(createdIns.Diseases))
	}
	if createdIns.Status != 1 {
		t.Fatalf("expected status 1 (active), got %d", createdIns.Status)
	}

	// 2. Fetch created insurance and verify persistence
	fetchedIns, err := insurancesService.Insurance(ctx, doctor, createdIns.ID)
	if err != nil {
		t.Fatalf("failed to fetch insurance as doctor: %v", err)
	}
	if fetchedIns.Name != createdIns.Name || fetchedIns.TotalMinor != 27000 {
		t.Fatalf("fetched insurance mismatch: %+v", fetchedIns)
	}
	if len(fetchedIns.Diseases) != 2 {
		t.Fatalf("expected 2 disease lines in fetched insurance, got %d", len(fetchedIns.Diseases))
	}

	// 3. Update Insurance (modify discount, hospital rate, and disease lines)
	updateIn := domain.InsuranceInput{
		Name:              "National Health Insurance Corp",
		ServiceTaxMinor:   5000,  // 50.00 ETB
		HospitalRateMinor: 20000, // 200.00 ETB
		Discount:          20,    // 20%
		Remark:            "Updated national corporate policy",
		InsuranceNo:       "NHIC-00918-B",
		InsuranceCode:     "NAT-HLTH-02",
		Diseases: []domain.InsuranceDiseaseLineInput{
			{DiseaseName: "Comprehensive Inpatient Care", DiseaseChargeMinor: 10000}, // 100.00 ETB
		},
	}
	// Base = 5000 + 20000 + 10000 = 35000. Discount 20% = 7000. Total = 28000.
	updatedIns, err := insurancesService.UpdateInsurance(ctx, receptionist, createdIns.ID, updateIn)
	if err != nil {
		t.Fatalf("failed to update insurance as receptionist: %v", err)
	}
	if updatedIns.TotalMinor != 28000 || updatedIns.Total != 280.00 {
		t.Fatalf("expected updated total 28000 minor, got %d", updatedIns.TotalMinor)
	}
	if len(updatedIns.Diseases) != 1 || updatedIns.Diseases[0].DiseaseName != "Comprehensive Inpatient Care" {
		t.Fatalf("expected 1 updated disease line, got %+v", updatedIns.Diseases)
	}

	// 4. Toggle Insurance Status (active <-> inactive)
	toggledIns, err := insurancesService.ToggleStatus(ctx, admin, createdIns.ID)
	if err != nil {
		t.Fatalf("failed to toggle insurance status: %v", err)
	}
	if toggledIns.Status != 0 {
		t.Fatalf("expected toggled status 0 (inactive), got %d", toggledIns.Status)
	}
	toggledBackIns, err := insurancesService.ToggleStatus(ctx, receptionist, createdIns.ID)
	if err != nil {
		t.Fatalf("failed to toggle back insurance status: %v", err)
	}
	if toggledBackIns.Status != 1 {
		t.Fatalf("expected toggled back status 1 (active), got %d", toggledBackIns.Status)
	}

	// 5. Test Duplicate Name Rejection
	dupIn := insIn
	dupIn.InsuranceNo = "DIFF-NO"
	dupIn.InsuranceCode = "DIFF-CODE"
	_, err = insurancesService.CreateInsurance(ctx, admin, dupIn)
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("expected ErrConflict for duplicate insurance name, got: %v", err)
	}

	// 6. Test Validation Rejections
	for _, badInput := range []domain.InsuranceInput{
		{Name: "", InsuranceNo: "NO", InsuranceCode: "CODE", Diseases: []domain.InsuranceDiseaseLineInput{{DiseaseName: "D", DiseaseChargeMinor: 100}}},
		{Name: "Valid", InsuranceNo: "", InsuranceCode: "CODE", Diseases: []domain.InsuranceDiseaseLineInput{{DiseaseName: "D", DiseaseChargeMinor: 100}}},
		{Name: "Valid", InsuranceNo: "NO", InsuranceCode: "", Diseases: []domain.InsuranceDiseaseLineInput{{DiseaseName: "D", DiseaseChargeMinor: 100}}},
		{Name: "Valid", InsuranceNo: "NO", InsuranceCode: "CODE", Discount: 101, Diseases: []domain.InsuranceDiseaseLineInput{{DiseaseName: "D", DiseaseChargeMinor: 100}}},
		{Name: "Valid", InsuranceNo: "NO", InsuranceCode: "CODE", Diseases: nil},
		{Name: "Valid", InsuranceNo: "NO", InsuranceCode: "CODE", Diseases: []domain.InsuranceDiseaseLineInput{{DiseaseName: "", DiseaseChargeMinor: 100}}},
	} {
		_, err = insurancesService.CreateInsurance(ctx, admin, badInput)
		if !errors.Is(err, domain.ErrValidation) {
			t.Fatalf("expected ErrValidation for bad input %+v, got: %v", badInput, err)
		}
	}

	// 7. Test In-Use Deletion Rejection (when referenced by ipd_admission_details)
	cleanInsIn := domain.InsuranceInput{
		Name:              "Temporary Insurance Policy",
		InsuranceNo:       "TEMP-001",
		InsuranceCode:     "TEMP-01",
		HospitalRateMinor: 5000,
		Diseases: []domain.InsuranceDiseaseLineInput{
			{DiseaseName: "General Check", DiseaseChargeMinor: 2000},
		},
	}
	cleanIns, err := insurancesService.CreateInsurance(ctx, admin, cleanInsIn)
	if err != nil {
		t.Fatalf("failed to create temporary insurance: %v", err)
	}

	// Reference createdIns in ipd_admission_details
	if len(cases) > 0 {
		var encID string
		err = db.QueryRow(ctx, `
			INSERT INTO encounter (case_id, patient_id, doctor_id, kind, status)
			VALUES ($1, $2, $3, 'ipd', 'admitted')
			RETURNING id
		`, cases[0].ID, cases[0].PatientID, cases[0].DoctorID).Scan(&encID)
		if err == nil {
			_, err = db.Exec(ctx, `
				INSERT INTO ipd_admission_details (encounter_id, insurance_id)
				VALUES ($1, $2)
			`, encID, createdIns.ID)
			if err != nil {
				t.Fatalf("failed to seed ipd_admission_details reference: %v", err)
			}

			// Deletion must be blocked
			err = insurancesService.DeleteInsurance(ctx, admin, createdIns.ID)
			if !errors.Is(err, domain.ErrInUse) {
				t.Fatalf("expected ErrInUse for referenced insurance deletion, got: %v", err)
			}

			// Verify record still exists
			_, err = insurancesService.Insurance(ctx, admin, createdIns.ID)
			if err != nil {
				t.Fatalf("referenced insurance was deleted despite being in use: %v", err)
			}

			// Clean up reference
			_, _ = db.Exec(ctx, `DELETE FROM ipd_admission_details WHERE encounter_id = $1`, encID)
			_, _ = db.Exec(ctx, `DELETE FROM encounter WHERE id = $1`, encID)
		}
	}

	// 8. Test Unreferenced Deletion
	err = insurancesService.DeleteInsurance(ctx, admin, cleanIns.ID)
	if err != nil {
		t.Fatalf("expected unreferenced insurance deletion to succeed, got: %v", err)
	}
	_, err = insurancesService.Insurance(ctx, admin, cleanIns.ID)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound after deletion, got: %v", err)
	}

	// 9. Test RBAC Permissions
	// Doctor cannot create
	_, err = insurancesService.CreateInsurance(ctx, doctor, cleanInsIn)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected ErrForbidden when doctor creates insurance, got: %v", err)
	}
	// Doctor cannot delete
	err = insurancesService.DeleteInsurance(ctx, doctor, createdIns.ID)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected ErrForbidden when doctor deletes insurance, got: %v", err)
	}
	// Patient cannot update
	_, err = insurancesService.UpdateInsurance(ctx, patientUser, createdIns.ID, updateIn)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected ErrForbidden when patient updates insurance, got: %v", err)
	}
	// Patient and Doctor can read
	_, err = insurancesService.Insurance(ctx, doctor, createdIns.ID)
	if err != nil {
		t.Fatalf("doctor should be permitted to read insurance: %v", err)
	}
	_, err = insurancesService.Insurance(ctx, patientUser, createdIns.ID)
	if err != nil {
		t.Fatalf("patient should be permitted to read insurance: %v", err)
	}
	// Nurse and Accountant cannot read insurance catalog
	_, err = insurancesService.Insurance(ctx, nurse, createdIns.ID)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("nurse should not be permitted to read insurance: %v", err)
	}
	_, err = insurancesService.Insurance(ctx, accountant, createdIns.ID)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("accountant should not be permitted to read insurance: %v", err)
	}

	// 10. Test HTTP API Handlers
	authSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie := r.Header.Get("Cookie")
		switch cookie {
		case "session=admin":
			json.NewEncoder(w).Encode(map[string]any{"user": map[string]string{"id": admin.ID}, "session": map[string]any{"userId": admin.ID, "expiresAt": time.Now().Add(time.Hour)}})
		case "session=doctor":
			json.NewEncoder(w).Encode(map[string]any{"user": map[string]string{"id": doctor.ID}, "session": map[string]any{"userId": doctor.ID, "expiresAt": time.Now().Add(time.Hour)}})
		case "session=nurse":
			json.NewEncoder(w).Encode(map[string]any{"user": map[string]string{"id": nurse.ID}, "session": map[string]any{"userId": nurse.ID, "expiresAt": time.Now().Add(time.Hour)}})
		default:
			w.WriteHeader(http.StatusUnauthorized)
		}
	}))
	defer authSrv.Close()

	httpHandler := httpapi.Server{
		Insurances: insurancesService,
		Actors:     store,
		AuthURL:    authSrv.URL,
		Origin:     "http://hospital.test",
		Client:     authSrv.Client(),
	}.Handler()

	// Anonymous access to GET /v1/insurances -> 401
	anonReq := httptest.NewRequest("GET", "/v1/insurances", nil)
	anonRec := httptest.NewRecorder()
	httpHandler.ServeHTTP(anonRec, anonReq)
	if anonRec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for anonymous GET /v1/insurances, got %d", anonRec.Code)
	}

	// Admin GET /v1/insurances -> 200
	adminReq := httptest.NewRequest("GET", "/v1/insurances", nil)
	adminReq.Header.Set("Cookie", "session=admin")
	adminReq.Header.Set("Origin", "http://hospital.test")
	adminRec := httptest.NewRecorder()
	httpHandler.ServeHTTP(adminRec, adminReq)
	if adminRec.Code != http.StatusOK {
		t.Fatalf("expected 200 for admin GET /v1/insurances, got %d: %s", adminRec.Code, adminRec.Body.String())
	}

	// Doctor POST /v1/insurances -> 403 Forbidden
	docPayload, _ := json.Marshal(domain.InsuranceInput{
		Name:          "Doctor Created Insurance",
		InsuranceNo:   "DOC-001",
		InsuranceCode: "DOC-01",
		Diseases:      []domain.InsuranceDiseaseLineInput{{DiseaseName: "Flu", DiseaseChargeMinor: 1000}},
	})
	docReq := httptest.NewRequest("POST", "/v1/insurances", bytes.NewReader(docPayload))
	docReq.Header.Set("Cookie", "session=doctor")
	docReq.Header.Set("Origin", "http://hospital.test")
	docReq.Header.Set("Content-Type", "application/json")
	docRec := httptest.NewRecorder()
	httpHandler.ServeHTTP(docRec, docReq)
	if docRec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for doctor POST /v1/insurances, got %d", docRec.Code)
	}

	// Nurse GET /v1/insurances -> 403 Forbidden
	nurseReq := httptest.NewRequest("GET", "/v1/insurances", nil)
	nurseReq.Header.Set("Cookie", "session=nurse")
	nurseReq.Header.Set("Origin", "http://hospital.test")
	nurseRec := httptest.NewRecorder()
	httpHandler.ServeHTTP(nurseRec, nurseReq)
	if nurseRec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for nurse GET /v1/insurances, got %d", nurseRec.Code)
	}

	// Export insurances endpoint
	exportReq := httptest.NewRequest("GET", "/v1/insurances-export", nil)
	exportReq.Header.Set("Cookie", "session=admin")
	exportReq.Header.Set("Origin", "http://hospital.test")
	exportRec := httptest.NewRecorder()
	httpHandler.ServeHTTP(exportRec, exportReq)
	if exportRec.Code != http.StatusOK {
		t.Fatalf("expected 200 for GET /v1/insurances-export, got %d", exportRec.Code)
	}

	// -------------------------------------------------------------
	// Regression R2: Money overflow with 100% discount must be rejected
	// -------------------------------------------------------------
	overflowInsIn := domain.InsuranceInput{
		Name:          "Overflow Insurance",
		InsuranceNo:   "INS-OVF",
		InsuranceCode: "OVF",
		Discount:      100,
		Diseases: []domain.InsuranceDiseaseLineInput{
			{DiseaseName: "Overflow Disease", DiseaseChargeMinor: 100000000000000000},
		},
	}
	_, err = insurancesService.CreateInsurance(ctx, admin, overflowInsIn)
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected ErrValidation for insurance money overflow, got %v", err)
	}

	// -------------------------------------------------------------
	// Regression R4: Export must return complete results (> 25 rows)
	// -------------------------------------------------------------
	for iIdx := 1; iIdx <= 35; iIdx++ {
		_, err = insurancesService.CreateInsurance(ctx, admin, domain.InsuranceInput{
			Name:              fmt.Sprintf("Export Bulk Insurance %03d", iIdx),
			ServiceTaxMinor:   1000,
			HospitalRateMinor: 2000,
			Discount:          5,
			InsuranceNo:       fmt.Sprintf("EXP-NO-%03d", iIdx),
			InsuranceCode:     fmt.Sprintf("EXP-CODE-%03d", iIdx),
			Diseases: []domain.InsuranceDiseaseLineInput{
				{DiseaseName: "Routine Coverage", DiseaseChargeMinor: 1000},
			},
		})
		if err != nil {
			t.Fatalf("failed to seed bulk insurance %d: %v", iIdx, err)
		}
	}

	expInsList, expInsTotal, err := insurancesService.ExportInsurances(ctx, admin, "Export Bulk Insurance")
	if err != nil {
		t.Fatalf("ExportInsurances failed: %v", err)
	}
	if expInsTotal != 35 || len(expInsList) != 35 {
		t.Fatalf("expected export to return all 35 bulk insurances, got len=%d total=%d", len(expInsList), expInsTotal)
	}

	bulkInsExportReq := httptest.NewRequest("GET", "/v1/insurances-export?search=Export+Bulk+Insurance", nil)
	bulkInsExportReq.Header.Set("Cookie", "session=admin")
	bulkInsExportReq.Header.Set("Origin", "http://hospital.test")
	bulkInsExportRec := httptest.NewRecorder()
	httpHandler.ServeHTTP(bulkInsExportRec, bulkInsExportReq)
	if bulkInsExportRec.Code != http.StatusOK {
		t.Fatalf("expected 200 for bulk insurances export, got %d", bulkInsExportRec.Code)
	}
	var insExportPayload struct {
		Insurances []domain.Insurance `json:"insurances"`
		Total      int                `json:"total"`
	}
	if err := json.Unmarshal(bulkInsExportRec.Body.Bytes(), &insExportPayload); err != nil {
		t.Fatalf("failed to decode insurances export response: %v", err)
	}
	if insExportPayload.Total != 35 || len(insExportPayload.Insurances) != 35 {
		t.Fatalf("HTTP export truncated: expected 35 insurances, got total=%d len=%d", insExportPayload.Total, len(insExportPayload.Insurances))
	}

	// -------------------------------------------------------------
	// Regression R7: Idempotent status updates, duplicate/retry safety, concurrent updates
	// -------------------------------------------------------------
	// 1) Test duplicate/retry setting status to 0 (must not oscillate/flip)
	st0 := 0
	for retry := 1; retry <= 3; retry++ {
		res, err := insurancesService.SetStatus(ctx, admin, createdIns.ID, &st0)
		if err != nil {
			t.Fatalf("SetStatus(0) failed on retry %d: %v", retry, err)
		}
		if res.Status != 0 {
			t.Fatalf("expected status 0 on retry %d, got %d", retry, res.Status)
		}
	}
	checkIns, err := insurancesService.Insurance(ctx, admin, createdIns.ID)
	if err != nil || checkIns.Status != 0 {
		t.Fatalf("expected persisted status 0, got %d (err: %v)", checkIns.Status, err)
	}

	// 2) Test duplicate/retry setting status to 1 (must not oscillate/flip)
	st1 := 1
	for retry := 1; retry <= 3; retry++ {
		res, err := insurancesService.SetStatus(ctx, admin, createdIns.ID, &st1)
		if err != nil {
			t.Fatalf("SetStatus(1) failed on retry %d: %v", retry, err)
		}
		if res.Status != 1 {
			t.Fatalf("expected status 1 on retry %d, got %d", retry, res.Status)
		}
	}

	// 3) Test invalid status rejection in application/service
	badSt := 2
	_, err = insurancesService.SetStatus(ctx, admin, createdIns.ID, &badSt)
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected ErrValidation for status=2, got: %v", err)
	}
	badStNeg := -1
	_, err = insurancesService.SetStatus(ctx, admin, createdIns.ID, &badStNeg)
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected ErrValidation for status=-1, got: %v", err)
	}

	// 4) Test invalid status rejection in InsuranceInput.Validate
	badStatusInput := domain.InsuranceInput{
		Name:          "Bad Status Insurance",
		InsuranceNo:   "BAD-ST-001",
		InsuranceCode: "BAD-ST",
		Status:        &badSt,
		Diseases: []domain.InsuranceDiseaseLineInput{
			{DiseaseName: "Care", DiseaseChargeMinor: 1000},
		},
	}
	_, err = insurancesService.CreateInsurance(ctx, admin, badStatusInput)
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected ErrValidation creating insurance with status=2, got: %v", err)
	}

	// 5) Test concurrent status requests for the same target status (all succeed, remain stable)
	var wg sync.WaitGroup
	concurrentErrors := make(chan error, 10)
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			target := 0
			r, e := insurancesService.SetStatus(ctx, admin, createdIns.ID, &target)
			if e != nil {
				concurrentErrors <- e
				return
			}
			if r.Status != 0 {
				concurrentErrors <- fmt.Errorf("unexpected status %d in concurrent update", r.Status)
			}
		}()
	}
	wg.Wait()
	close(concurrentErrors)
	for e := range concurrentErrors {
		t.Fatalf("concurrent SetStatus error: %v", e)
	}
	finalIns, err := insurancesService.Insurance(ctx, admin, createdIns.ID)
	if err != nil || finalIns.Status != 0 {
		t.Fatalf("expected final status 0 after concurrent updates, got %d (err: %v)", finalIns.Status, err)
	}

	// 6) Test HTTP PATCH /v1/insurances/{id}/status endpoint with explicit status & retries
	for retry := 1; retry <= 2; retry++ {
		patchPayload, _ := json.Marshal(map[string]int{"status": 1})
		patchReq := httptest.NewRequest("PATCH", fmt.Sprintf("/v1/insurances/%s/status", createdIns.ID), bytes.NewReader(patchPayload))
		patchReq.Header.Set("Cookie", "session=admin")
		patchReq.Header.Set("Origin", "http://hospital.test")
		patchReq.Header.Set("Content-Type", "application/json")
		patchRec := httptest.NewRecorder()
		httpHandler.ServeHTTP(patchRec, patchReq)
		if patchRec.Code != http.StatusOK {
			t.Fatalf("expected 200 for HTTP PATCH status on retry %d, got %d: %s", retry, patchRec.Code, patchRec.Body.String())
		}
		var patchResp domain.Insurance
		if err := json.Unmarshal(patchRec.Body.Bytes(), &patchResp); err != nil {
			t.Fatalf("failed to decode PATCH response: %v", err)
		}
		if patchResp.Status != 1 {
			t.Fatalf("expected status 1 in HTTP response on retry %d, got %d", retry, patchResp.Status)
		}
	}

	// 7) Test HTTP PATCH with invalid status (returns 422 Unprocessable Entity)
	badPatchPayload, _ := json.Marshal(map[string]int{"status": 99})
	badPatchReq := httptest.NewRequest("PATCH", fmt.Sprintf("/v1/insurances/%s/status", createdIns.ID), bytes.NewReader(badPatchPayload))
	badPatchReq.Header.Set("Cookie", "session=admin")
	badPatchReq.Header.Set("Origin", "http://hospital.test")
	badPatchReq.Header.Set("Content-Type", "application/json")
	badPatchRec := httptest.NewRecorder()
	httpHandler.ServeHTTP(badPatchRec, badPatchReq)
	if badPatchRec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422 Unprocessable Entity for status=99, got %d", badPatchRec.Code)
	}

	// 8) Comprehensive 9-role authorization coverage for insurances
	allRoles := []string{"admin", "receptionist", "doctor", "case_manager", "patient", "nurse", "accountant", "pharmacist", "laboratorian"}
	for _, rName := range allRoles {
		act := actorMap[rName]
		if act.ID == "" {
			act = domain.Actor{ID: rName, Role: rName}
		}
		canManageExpected := (rName == "admin" || rName == "receptionist")
		canReadExpected := (rName == "admin" || rName == "receptionist" || rName == "doctor" || rName == "case_manager" || rName == "patient")

		// Test read
		_, readErr := insurancesService.Insurance(ctx, act, createdIns.ID)
		if canReadExpected && readErr != nil {
			t.Errorf("role %s expected read allowed, got error: %v", rName, readErr)
		} else if !canReadExpected && !errors.Is(readErr, domain.ErrForbidden) {
			t.Errorf("role %s expected read forbidden, got: %v", rName, readErr)
		}

		// Test status update
		stVal := 1
		_, stErr := insurancesService.SetStatus(ctx, act, createdIns.ID, &stVal)
		if canManageExpected && stErr != nil {
			t.Errorf("role %s expected status manage allowed, got error: %v", rName, stErr)
		} else if !canManageExpected && !errors.Is(stErr, domain.ErrForbidden) {
			t.Errorf("role %s expected status manage forbidden, got: %v", rName, stErr)
		}
	}
}
