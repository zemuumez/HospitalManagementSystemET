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

func testPackages(t *testing.T, db *pgxpool.Pool, store Store, actors []domain.Actor, cases []domain.Case) {
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

	// Insert active services
	var svc1ID, svc2ID, svc3ID, inactiveSvcID string
	err := db.QueryRow(ctx, `
		INSERT INTO hospital_service (name, description, quantity, rate_minor, status)
		VALUES ('Package Test Service 1', 'Test consultation', 1, 25000, 1)
		RETURNING id
	`).Scan(&svc1ID)
	if err != nil {
		t.Fatalf("failed to insert test service 1: %v", err)
	}

	err = db.QueryRow(ctx, `
		INSERT INTO hospital_service (name, description, quantity, rate_minor, status)
		VALUES ('Package Test Service 2', 'Test lab investigation', 1, 15000, 1)
		RETURNING id
	`).Scan(&svc2ID)
	if err != nil {
		t.Fatalf("failed to insert test service 2: %v", err)
	}

	err = db.QueryRow(ctx, `
		INSERT INTO hospital_service (name, description, quantity, rate_minor, status)
		VALUES ('Package Test Service 3', 'Test therapy', 1, 10000, 1)
		RETURNING id
	`).Scan(&svc3ID)
	if err != nil {
		t.Fatalf("failed to insert test service 3: %v", err)
	}

	err = db.QueryRow(ctx, `
		INSERT INTO hospital_service (name, description, quantity, rate_minor, status)
		VALUES ('Package Inactive Service', 'Disabled service', 1, 5000, 0)
		RETURNING id
	`).Scan(&inactiveSvcID)
	if err != nil {
		t.Fatalf("failed to insert inactive service: %v", err)
	}

	packagesService := application.PackagesService{Store: store, Now: time.Now}

	// -------------------------------------------------------------
	// 1. Valid multi-line creation and persisted totals
	// -------------------------------------------------------------
	pkgIn := domain.PackageInput{
		Name:        "Executive Health Check",
		Description: "Comprehensive health checkup",
		Discount:    10,
		Services: []domain.PackageServiceLineInput{
			{ServiceID: svc1ID, Quantity: 1, RateMinor: 25000},
			{ServiceID: svc2ID, Quantity: 2, RateMinor: 15000},
		},
	}
	createdPkg, err := packagesService.CreatePackage(ctx, admin, pkgIn)
	if err != nil {
		t.Fatalf("failed to create valid package: %v", err)
	}
	if createdPkg.ID == "" {
		t.Fatal("expected created package to have an ID")
	}
	// Subtotal: 25000 + 30000 = 55000. 10% discount = 5500. Total = 49500.
	if createdPkg.TotalAmountMinor != 49500 {
		t.Fatalf("expected total 49500, got %d", createdPkg.TotalAmountMinor)
	}
	if len(createdPkg.Services) != 2 {
		t.Fatalf("expected 2 service lines, got %d", len(createdPkg.Services))
	}
	if createdPkg.Services[0].AmountMinor != 25000 || createdPkg.Services[1].AmountMinor != 30000 {
		t.Fatalf("unexpected line amounts: %v", createdPkg.Services)
	}

	// Verify persistence in DB
	fetchedPkg, err := packagesService.Package(ctx, admin, createdPkg.ID)
	if err != nil {
		t.Fatalf("failed to fetch package: %v", err)
	}
	if fetchedPkg.Name != "Executive Health Check" || fetchedPkg.TotalAmountMinor != 49500 {
		t.Fatalf("fetched package mismatch: %+v", fetchedPkg)
	}
	if len(fetchedPkg.Services) != 2 {
		t.Fatalf("fetched package line count mismatch: %d", len(fetchedPkg.Services))
	}

	// Verify audit event
	var auditCount int
	err = db.QueryRow(ctx, `SELECT count(*) FROM audit_event WHERE action = 'package.created' AND resource_id = $1`, createdPkg.ID).Scan(&auditCount)
	if err != nil || auditCount != 1 {
		t.Fatalf("expected 1 audit event for package.created, got %d (err: %v)", auditCount, err)
	}

	// -------------------------------------------------------------
	// 2. Update/add/remove lines
	// -------------------------------------------------------------
	// Modify line 1 quantity to 2, delete line 2, add line 3, change discount to 20%
	line1ID := fetchedPkg.Services[0].ID
	updateIn := domain.PackageInput{
		Name:        "Executive Health Check Updated",
		Description: "Updated description",
		Discount:    20,
		Services: []domain.PackageServiceLineInput{
			{ID: line1ID, ServiceID: svc1ID, Quantity: 2, RateMinor: 25000},
			{ServiceID: svc3ID, Quantity: 1, RateMinor: 10000},
		},
	}
	updatedPkg, err := packagesService.UpdatePackage(ctx, receptionist, createdPkg.ID, updateIn)
	if err != nil {
		t.Fatalf("receptionist failed to update package: %v", err)
	}
	// Subtotal: 2*25000 + 1*10000 = 60000. 20% discount = 12000. Total = 48000.
	if updatedPkg.TotalAmountMinor != 48000 {
		t.Fatalf("expected updated total 48000, got %d", updatedPkg.TotalAmountMinor)
	}
	if len(updatedPkg.Services) != 2 {
		t.Fatalf("expected 2 service lines after update, got %d", len(updatedPkg.Services))
	}

	// Verify DB state: line 2 was deleted, line 1 quantity was updated
	var remainingLines int
	_ = db.QueryRow(ctx, `SELECT count(*) FROM package_service WHERE package_id = $1`, createdPkg.ID).Scan(&remainingLines)
	if remainingLines != 2 {
		t.Fatalf("expected 2 lines in DB after update, got %d", remainingLines)
	}

	// Verify audit event for update
	err = db.QueryRow(ctx, `SELECT count(*) FROM audit_event WHERE action = 'package.updated' AND resource_id = $1`, createdPkg.ID).Scan(&auditCount)
	if err != nil || auditCount != 1 {
		t.Fatalf("expected 1 audit event for package.updated, got %d (err: %v)", auditCount, err)
	}

	// -------------------------------------------------------------
	// 3. Duplicate names rejection
	// -------------------------------------------------------------
	dupIn := domain.PackageInput{
		Name:        "Executive Health Check Updated",
		Description: "Duplicate name package",
		Discount:    5,
		Services: []domain.PackageServiceLineInput{
			{ServiceID: svc1ID, Quantity: 1, RateMinor: 25000},
		},
	}
	_, err = packagesService.CreatePackage(ctx, admin, dupIn)
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("expected ErrConflict for duplicate package name, got %v", err)
	}

	// -------------------------------------------------------------
	// 4. Invalid references
	// -------------------------------------------------------------
	// Non-existent service ID
	badRef := domain.PackageInput{
		Name:     "Non-existent Service Package",
		Discount: 0,
		Services: []domain.PackageServiceLineInput{
			{ServiceID: "00000000-0000-0000-0000-000000000000", Quantity: 1, RateMinor: 1000},
		},
	}
	_, err = packagesService.CreatePackage(ctx, admin, badRef)
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected ErrValidation for non-existent service, got %v", err)
	}

	// Inactive service
	inactiveRefIn := domain.PackageInput{
		Name:     "Inactive Service Package",
		Discount: 0,
		Services: []domain.PackageServiceLineInput{
			{ServiceID: inactiveSvcID, Quantity: 1, RateMinor: 5000},
		},
	}
	_, err = packagesService.CreatePackage(ctx, admin, inactiveRefIn)
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected ErrValidation for inactive service, got %v", err)
	}

	// -------------------------------------------------------------
	// 5. Invalid quantities/rates/discounts
	// -------------------------------------------------------------
	for _, badInput := range []domain.PackageInput{
		{Name: "Bad Qty", Discount: 10, Services: []domain.PackageServiceLineInput{{ServiceID: svc1ID, Quantity: 0, RateMinor: 25000}}},
		{Name: "Bad Rate", Discount: 10, Services: []domain.PackageServiceLineInput{{ServiceID: svc1ID, Quantity: 1, RateMinor: -100}}},
		{Name: "Bad Disc", Discount: -5, Services: []domain.PackageServiceLineInput{{ServiceID: svc1ID, Quantity: 1, RateMinor: 25000}}},
		{Name: "Bad Disc High", Discount: 105, Services: []domain.PackageServiceLineInput{{ServiceID: svc1ID, Quantity: 1, RateMinor: 25000}}},
	} {
		_, err = packagesService.CreatePackage(ctx, admin, badInput)
		if !errors.Is(err, domain.ErrValidation) {
			t.Fatalf("expected ErrValidation for bad input %+v, got %v", badInput, err)
		}
	}

	// -------------------------------------------------------------
	// 6. Cross-parent child-ID rejection
	// -------------------------------------------------------------
	// Create Package A and Package B
	pkgA, err := packagesService.CreatePackage(ctx, admin, domain.PackageInput{
		Name:     "Package A",
		Discount: 5,
		Services: []domain.PackageServiceLineInput{{ServiceID: svc1ID, Quantity: 1, RateMinor: 25000}},
	})
	if err != nil {
		t.Fatalf("failed to create pkgA: %v", err)
	}

	pkgB, err := packagesService.CreatePackage(ctx, admin, domain.PackageInput{
		Name:     "Package B",
		Discount: 5,
		Services: []domain.PackageServiceLineInput{{ServiceID: svc2ID, Quantity: 1, RateMinor: 15000}},
	})
	if err != nil {
		t.Fatalf("failed to create pkgB: %v", err)
	}

	// Attempt to update Package A by passing Package B's line ID
	lineB_ID := pkgB.Services[0].ID
	crossParentUpdate := domain.PackageInput{
		Name:     "Package A Hijack",
		Discount: 5,
		Services: []domain.PackageServiceLineInput{
			{ID: lineB_ID, ServiceID: svc1ID, Quantity: 2, RateMinor: 25000},
		},
	}
	_, err = packagesService.UpdatePackage(ctx, admin, pkgA.ID, crossParentUpdate)
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("expected ErrConflict for cross-parent line ID, got %v", err)
	}

	// Verify Package B's line still belongs to Package B
	var checkOwner string
	_ = db.QueryRow(ctx, `SELECT package_id FROM package_service WHERE id = $1`, lineB_ID).Scan(&checkOwner)
	if checkOwner != pkgB.ID {
		t.Fatalf("package B line corrupted! Owner is %s, expected %s", checkOwner, pkgB.ID)
	}

	// -------------------------------------------------------------
	// 7. Rollback after a forced child-write failure
	// -------------------------------------------------------------
	// Create an input where the first service is valid, but the second service is invalid in DB
	txErrInput := domain.PackageInput{
		Name:     "Rollback Test Package",
		Discount: 10,
		Services: []domain.PackageServiceLineInput{
			{ServiceID: svc1ID, Quantity: 1, RateMinor: 25000},
			{ServiceID: "00000000-0000-0000-0000-000000000000", Quantity: 1, RateMinor: 1000},
		},
	}
	_, err = packagesService.CreatePackage(ctx, admin, txErrInput)
	if err == nil {
		t.Fatal("expected error on invalid second service line")
	}
	// Verify that "Rollback Test Package" was NOT created in DB
	var rollbackExists bool
	_ = db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM package WHERE name = 'Rollback Test Package')`).Scan(&rollbackExists)
	if rollbackExists {
		t.Fatal("transaction failed to roll back! 'Rollback Test Package' found in database")
	}

	// -------------------------------------------------------------
	// 8. Admission-linked deletion rejection
	// -------------------------------------------------------------
	// If cases is available, link package to an encounter in ipd_admission_details
	if len(cases) > 0 {
		var encID string
		err = db.QueryRow(ctx, `
			INSERT INTO encounter (kind, case_id, patient_id, doctor_id, admitted_at, status, request_key, created_by)
			VALUES ('opd', $1, $2, $3, now(), 'active', 'enc-pkg-test', $3)
			RETURNING id
		`, cases[0].ID, cases[0].PatientID, cases[0].DoctorID).Scan(&encID)
		if err == nil {
			_, err = db.Exec(ctx, `
				INSERT INTO ipd_admission_details (encounter_id, package_id, package_name, package_charge_minor)
				VALUES ($1, $2, 'Package A Linked', 25000)
			`, encID, pkgA.ID)
			if err != nil {
				t.Fatalf("failed to insert ipd_admission_details: %v", err)
			}

			// Attempt to delete pkgA -> must fail with ErrInUse!
			err = packagesService.DeletePackage(ctx, admin, pkgA.ID)
			if !errors.Is(err, domain.ErrInUse) {
				t.Fatalf("expected ErrInUse for admission-linked package deletion, got %v", err)
			}

			// Verify pkgA still exists
			_, err = packagesService.Package(ctx, admin, pkgA.ID)
			if err != nil {
				t.Fatalf("admission-linked package was improperly deleted: %v", err)
			}

			// Remove admission link
			_, _ = db.Exec(ctx, `DELETE FROM ipd_admission_details WHERE encounter_id = $1`, encID)
			_, _ = db.Exec(ctx, `DELETE FROM encounter WHERE id = $1`, encID)
		}
	}

	// Delete unlinked package pkgA -> should succeed
	err = packagesService.DeletePackage(ctx, admin, pkgA.ID)
	if err != nil {
		t.Fatalf("failed to delete unlinked package pkgA: %v", err)
	}
	// Verify lines cascaded and package gone
	var pkgACount int
	_ = db.QueryRow(ctx, `SELECT count(*) FROM package WHERE id = $1`, pkgA.ID).Scan(&pkgACount)
	if pkgACount != 0 {
		t.Fatal("package pkgA still exists in DB after deletion")
	}
	var linesACount int
	_ = db.QueryRow(ctx, `SELECT count(*) FROM package_service WHERE package_id = $1`, pkgA.ID).Scan(&linesACount)
	if linesACount != 0 {
		t.Fatal("package_service lines for pkgA were not cascaded on deletion")
	}

	// Verify audit event for deletion
	err = db.QueryRow(ctx, `SELECT count(*) FROM audit_event WHERE action = 'package.deleted' AND resource_id = $1`, pkgA.ID).Scan(&auditCount)
	if err != nil || auditCount != 1 {
		t.Fatalf("expected 1 audit event for package.deleted, got %d (err: %v)", auditCount, err)
	}

	// -------------------------------------------------------------
	// 9. Role permissions and record scope
	// -------------------------------------------------------------
	// Doctor cannot create or delete packages
	_, err = packagesService.CreatePackage(ctx, doctor, pkgIn)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected ErrForbidden for doctor creating package, got %v", err)
	}
	err = packagesService.DeletePackage(ctx, doctor, pkgB.ID)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected ErrForbidden for doctor deleting package, got %v", err)
	}

	// Patient cannot manage packages
	_, err = packagesService.UpdatePackage(ctx, patientUser, pkgB.ID, pkgIn)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected ErrForbidden for patient updating package, got %v", err)
	}

	// Doctor, patient, and receptionist CAN read packages
	_, err = packagesService.Package(ctx, doctor, pkgB.ID)
	if err != nil {
		t.Fatalf("doctor should be allowed to read package, got %v", err)
	}
	_, err = packagesService.Package(ctx, patientUser, pkgB.ID)
	if err != nil {
		t.Fatalf("patient should be allowed to read package, got %v", err)
	}

	// Nurse and accountant CANNOT read packages
	_, err = packagesService.Package(ctx, nurse, pkgB.ID)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected ErrForbidden for nurse reading package, got %v", err)
	}
	_, err = packagesService.Package(ctx, accountant, pkgB.ID)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected ErrForbidden for accountant reading package, got %v", err)
	}

	// -------------------------------------------------------------
	// 10. Concurrency & Pagination
	// -------------------------------------------------------------
	// Test concurrent updates on pkgB
	var wg sync.WaitGroup
	var updateErrors int
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(qty int) {
			defer wg.Done()
			_, uErr := packagesService.UpdatePackage(ctx, admin, pkgB.ID, domain.PackageInput{
				Name:     "Package B",
				Discount: 5,
				Services: []domain.PackageServiceLineInput{
					{ServiceID: svc2ID, Quantity: qty + 1, RateMinor: 15000},
				},
			})
			if uErr != nil {
				updateErrors++
			}
		}(i)
	}
	wg.Wait()
	// Check package B state is consistent
	pkgBFinal, err := packagesService.Package(ctx, admin, pkgB.ID)
	if err != nil {
		t.Fatalf("failed to fetch pkgB after concurrent updates: %v", err)
	}
	if len(pkgBFinal.Services) != 1 {
		t.Fatalf("expected exactly 1 service line on pkgB, got %d", len(pkgBFinal.Services))
	}

	// Pagination & search
	list, total, err := packagesService.Packages(ctx, admin, 1, 10, "Executive")
	if err != nil {
		t.Fatalf("failed to search packages: %v", err)
	}
	if total < 1 || len(list) < 1 {
		t.Fatalf("expected search for 'Executive' to find results, got total=%d, len=%d", total, len(list))
	}

	// -------------------------------------------------------------
	// 11. HTTP API integration & anonymous access
	// -------------------------------------------------------------
	authSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
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
		Packages: packagesService,
		Actors:   store,
		AuthURL:  authSrv.URL,
		Origin:   "http://hospital.test",
		Client:   authSrv.Client(),
	}.Handler()

	// Anonymous access to GET /v1/packages -> 401
	anonReq := httptest.NewRequest("GET", "/v1/packages", nil)
	anonRec := httptest.NewRecorder()
	httpHandler.ServeHTTP(anonRec, anonReq)
	if anonRec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for anonymous GET /v1/packages, got %d", anonRec.Code)
	}

	// Admin GET /v1/packages -> 200
	adminReq := httptest.NewRequest("GET", "/v1/packages", nil)
	adminReq.Header.Set("Cookie", "session=admin")
	adminReq.Header.Set("Origin", "http://hospital.test")
	adminRec := httptest.NewRecorder()
	httpHandler.ServeHTTP(adminRec, adminReq)
	if adminRec.Code != http.StatusOK {
		t.Fatalf("expected 200 for admin GET /v1/packages, got %d: %s", adminRec.Code, adminRec.Body.String())
	}

	// Doctor POST /v1/packages -> 403 Forbidden
	docPayload, _ := json.Marshal(domain.PackageInput{
		Name:     "Doctor Created Package",
		Discount: 5,
		Services: []domain.PackageServiceLineInput{{ServiceID: svc1ID, Quantity: 1, RateMinor: 25000}},
	})
	docReq := httptest.NewRequest("POST", "/v1/packages", bytes.NewReader(docPayload))
	docReq.Header.Set("Cookie", "session=doctor")
	docReq.Header.Set("Origin", "http://hospital.test")
	docReq.Header.Set("Content-Type", "application/json")
	docRec := httptest.NewRecorder()
	httpHandler.ServeHTTP(docRec, docReq)
	if docRec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for doctor POST /v1/packages, got %d", docRec.Code)
	}

	// Nurse GET /v1/packages -> 403 Forbidden
	nurseReq := httptest.NewRequest("GET", "/v1/packages", nil)
	nurseReq.Header.Set("Cookie", "session=nurse")
	nurseReq.Header.Set("Origin", "http://hospital.test")
	nurseRec := httptest.NewRecorder()
	httpHandler.ServeHTTP(nurseRec, nurseReq)
	if nurseRec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for nurse GET /v1/packages, got %d", nurseRec.Code)
	}

	// Export packages endpoint
	exportReq := httptest.NewRequest("GET", "/v1/packages-export", nil)
	exportReq.Header.Set("Cookie", "session=admin")
	exportReq.Header.Set("Origin", "http://hospital.test")
	exportRec := httptest.NewRecorder()
	httpHandler.ServeHTTP(exportRec, exportReq)
	if exportRec.Code != http.StatusOK {
		t.Fatalf("expected 200 for GET /v1/packages-export, got %d", exportRec.Code)
	}

	// -------------------------------------------------------------
	// 12. Regression R1: Duplicate child line ID on update must be rejected
	// -------------------------------------------------------------
	pkgR1In := domain.PackageInput{
		Name:        "R1 Test Package",
		Description: "Testing duplicate child ID rejection",
		Discount:    0,
		Services: []domain.PackageServiceLineInput{
			{ServiceID: svc1ID, Quantity: 1, RateMinor: 10000, HasRate: true},
		},
	}
	pkgR1Created, err := packagesService.CreatePackage(ctx, admin, pkgR1In)
	if err != nil {
		t.Fatalf("failed to create R1 package: %v", err)
	}
	childID := pkgR1Created.Services[0].ID

	var auditCountBefore int
	_ = db.QueryRow(ctx, `SELECT count(*) FROM audit_event WHERE action = 'package.updated' AND resource_id = $1`, pkgR1Created.ID).Scan(&auditCountBefore)

	updateDupChild := domain.PackageInput{
		Name:     "R1 Test Package Mutated",
		Discount: 0,
		Services: []domain.PackageServiceLineInput{
			{ID: childID, ServiceID: svc1ID, Quantity: 1, RateMinor: 100, HasRate: true},
			{ID: childID, ServiceID: svc2ID, Quantity: 1, RateMinor: 200, HasRate: true},
		},
	}
	_, err = packagesService.UpdatePackage(ctx, admin, pkgR1Created.ID, updateDupChild)
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected ErrValidation for duplicate child ID, got %v", err)
	}

	pkgR1Reloaded, err := packagesService.Package(ctx, admin, pkgR1Created.ID)
	if err != nil {
		t.Fatalf("failed to reload R1 package: %v", err)
	}
	if len(pkgR1Reloaded.Services) != 1 || pkgR1Reloaded.Services[0].RateMinor != 10000 {
		t.Fatalf("original rows were mutated despite validation failure: %+v", pkgR1Reloaded)
	}
	if pkgR1Reloaded.TotalAmountMinor != 10000 {
		t.Fatalf("parent total corrupted: expected 10000, got %d", pkgR1Reloaded.TotalAmountMinor)
	}

	var auditCountAfter int
	_ = db.QueryRow(ctx, `SELECT count(*) FROM audit_event WHERE action = 'package.updated' AND resource_id = $1`, pkgR1Created.ID).Scan(&auditCountAfter)
	if auditCountAfter != auditCountBefore {
		t.Fatalf("audit event was recorded for failed duplicate child ID update")
	}

	// -------------------------------------------------------------
	// 13. Regression R2: Money overflow with 100% discount must be rejected
	// -------------------------------------------------------------
	overflowIn := domain.PackageInput{
		Name:     "Overflow Package",
		Discount: 100,
		Services: []domain.PackageServiceLineInput{
			{ServiceID: svc1ID, Quantity: 1, RateMinor: 100000000000000000, HasRate: true},
		},
	}
	_, err = packagesService.CreatePackage(ctx, admin, overflowIn)
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected ErrValidation for money overflow, got %v", err)
	}

	// -------------------------------------------------------------
	// 14. Regression R3: Explicit zero price must NOT be overwritten by catalog price
	// -------------------------------------------------------------
	explicitZeroIn := domain.PackageInput{
		Name:     "Explicit Zero Rate Package",
		Discount: 0,
		Services: []domain.PackageServiceLineInput{
			{ServiceID: svc1ID, Quantity: 1, RateMinor: 0, HasRate: true},
		},
	}
	pkgZeroCreated, err := packagesService.CreatePackage(ctx, admin, explicitZeroIn)
	if err != nil {
		t.Fatalf("failed to create package with explicit zero rate: %v", err)
	}
	if len(pkgZeroCreated.Services) != 1 || pkgZeroCreated.Services[0].RateMinor != 0 {
		t.Fatalf("expected line rate 0, got %d", pkgZeroCreated.Services[0].RateMinor)
	}
	if pkgZeroCreated.TotalAmountMinor != 0 {
		t.Fatalf("expected total 0 for explicit zero rate, got %d", pkgZeroCreated.TotalAmountMinor)
	}

	updateZeroIn := domain.PackageInput{
		Name:     "Explicit Zero Rate Package Updated",
		Discount: 0,
		Services: []domain.PackageServiceLineInput{
			{ID: pkgZeroCreated.Services[0].ID, ServiceID: svc1ID, Quantity: 2, RateMinor: 0, HasRate: true},
		},
	}
	pkgZeroUpdated, err := packagesService.UpdatePackage(ctx, admin, pkgZeroCreated.ID, updateZeroIn)
	if err != nil {
		t.Fatalf("failed to update package with explicit zero rate: %v", err)
	}
	if pkgZeroUpdated.Services[0].RateMinor != 0 || pkgZeroUpdated.TotalAmountMinor != 0 {
		t.Fatalf("expected line rate 0 and total 0 on update, got rate=%d total=%d", pkgZeroUpdated.Services[0].RateMinor, pkgZeroUpdated.TotalAmountMinor)
	}

	omittedRateIn := domain.PackageInput{
		Name:     "Omitted Rate Package",
		Discount: 0,
		Services: []domain.PackageServiceLineInput{
			{ServiceID: svc1ID, Quantity: 1}, // HasRate: false
		},
	}
	pkgOmittedCreated, err := packagesService.CreatePackage(ctx, admin, omittedRateIn)
	if err != nil {
		t.Fatalf("failed to create package with omitted rate: %v", err)
	}
	if pkgOmittedCreated.Services[0].RateMinor != 25000 || pkgOmittedCreated.TotalAmountMinor != 25000 {
		t.Fatalf("expected omitted rate to default to 25000, got rate=%d total=%d", pkgOmittedCreated.Services[0].RateMinor, pkgOmittedCreated.TotalAmountMinor)
	}

	// -------------------------------------------------------------
	// 15. Regression R4: Export must NOT truncate at 25 or 100 rows
	// -------------------------------------------------------------
	for pIdx := 1; pIdx <= 35; pIdx++ {
		_, err = packagesService.CreatePackage(ctx, admin, domain.PackageInput{
			Name:     fmt.Sprintf("Export Bulk Package %03d", pIdx),
			Discount: 5,
			Services: []domain.PackageServiceLineInput{
				{ServiceID: svc1ID, Quantity: 1, RateMinor: 10000, HasRate: true},
			},
		})
		if err != nil {
			t.Fatalf("failed to seed bulk package %d: %v", pIdx, err)
		}
	}
	expList, expTotal, err := packagesService.ExportPackages(ctx, admin, "Export Bulk Package")
	if err != nil {
		t.Fatalf("ExportPackages failed: %v", err)
	}
	if expTotal != 35 || len(expList) != 35 {
		t.Fatalf("expected export to return all 35 bulk packages, got len=%d total=%d", len(expList), expTotal)
	}

	bulkExportReq := httptest.NewRequest("GET", "/v1/packages-export?search=Export+Bulk+Package", nil)
	bulkExportReq.Header.Set("Cookie", "session=admin")
	bulkExportReq.Header.Set("Origin", "http://hospital.test")
	bulkExportRec := httptest.NewRecorder()
	httpHandler.ServeHTTP(bulkExportRec, bulkExportReq)
	if bulkExportRec.Code != http.StatusOK {
		t.Fatalf("expected 200 for bulk packages export, got %d", bulkExportRec.Code)
	}
	var exportPayload struct {
		Packages []domain.Package `json:"packages"`
		Total    int              `json:"total"`
	}
	if err := json.Unmarshal(bulkExportRec.Body.Bytes(), &exportPayload); err != nil {
		t.Fatalf("failed to decode export response: %v", err)
	}
	if exportPayload.Total != 35 || len(exportPayload.Packages) != 35 {
		t.Fatalf("HTTP export truncated: expected 35 records, got total=%d len=%d", exportPayload.Total, len(exportPayload.Packages))
	}
}
