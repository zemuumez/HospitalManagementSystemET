package postgres

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"hms.local/api/internal/adapters/httpapi"
	"hms.local/api/internal/application"
	"hms.local/api/internal/domain"
)

func testServicesOperations(t *testing.T, db *pgxpool.Pool, store Store, actors []domain.Actor) {
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

	srv := application.ServicesOperationsService{Store: store, Now: time.Now}

	// 1. Authorization tests
	// Patient cannot manage services
	_, err := srv.CreateService(ctx, patientUser, domain.HospitalServiceInput{
		Name:      "Unauthorized Service",
		RateMinor: 5000,
	})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected ErrForbidden for patient creating service, got %v", err)
	}

	// Receptionist cannot manage operations
	_, err = srv.CreateOperationCategory(ctx, receptionist, domain.OperationCategoryInput{
		Name: "Forbidden Op Category",
	})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected ErrForbidden for receptionist creating op category, got %v", err)
	}

	// Doctor can read services and operations
	svcList, _, err := srv.Services(ctx, doctor, 1, nil)
	if err != nil {
		t.Fatalf("expected doctor to read services, got %v", err)
	}
	if len(svcList) == 0 {
		t.Fatalf("expected seeded services to be returned")
	}

	// 2. Charge Categories and Charges CRUD & Validation
	// Validation error: invalid charge_type
	_, err = srv.CreateChargeCategory(ctx, admin, domain.ChargeCategoryInput{
		Name:       "Invalid Type Cat",
		ChargeType: 99,
	})
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected ErrValidation for invalid charge_type, got %v", err)
	}

	// Create valid charge category
	cc, err := srv.CreateChargeCategory(ctx, admin, domain.ChargeCategoryInput{
		Name:        "Test Dental Procedures",
		Description: "General and cosmetic dental work",
		ChargeType:  5,
	})
	if err != nil {
		t.Fatalf("failed to create charge category: %v", err)
	}
	if cc.ID == "" || cc.Name != "Test Dental Procedures" {
		t.Fatalf("unexpected charge category: %+v", cc)
	}

	// Update charge category
	ccUpdated, err := srv.UpdateChargeCategory(ctx, admin, cc.ID, domain.ChargeCategoryInput{
		Name:        "Test Dental & Maxillofacial",
		Description: "Updated description",
		ChargeType:  5,
	})
	if err != nil {
		t.Fatalf("failed to update charge category: %v", err)
	}
	if ccUpdated.Name != "Test Dental & Maxillofacial" {
		t.Fatalf("expected updated name, got %s", ccUpdated.Name)
	}

	// Create hospital charge under this category
	chg, err := srv.CreateCharge(ctx, admin, domain.HospitalChargeInput{
		ChargeType:          5,
		ChargeCategoryID:    cc.ID,
		Code:                "DENT-001",
		StandardChargeMinor: 45000,
		Description:         "Tooth Extraction Simple",
	})
	if err != nil {
		t.Fatalf("failed to create hospital charge: %v", err)
	}
	if chg.ID == "" || chg.Code != "DENT-001" {
		t.Fatalf("unexpected charge: %+v", chg)
	}

	// Read charges list
	chgList, totalChg, err := srv.Charges(ctx, admin, 1, "")
	if err != nil || totalChg == 0 || len(chgList) == 0 {
		t.Fatalf("failed to list charges: %v, total: %d", err, totalChg)
	}

	// 3. Hospital Services CRUD & Validation
	// Validation error: negative rate
	_, err = srv.CreateService(ctx, admin, domain.HospitalServiceInput{
		Name:      "Negative Rate Service",
		RateMinor: -100,
	})
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected ErrValidation for negative rate, got %v", err)
	}

	// Create valid service
	newSvc, err := srv.CreateService(ctx, admin, domain.HospitalServiceInput{
		Name:        "Physiotherapy Session",
		Description: "45-minute musculoskeletal rehabilitation",
		Quantity:    1,
		RateMinor:   30000,
		Status:      1,
	})
	if err != nil {
		t.Fatalf("failed to create service: %v", err)
	}

	// Update service
	updatedSvc, err := srv.UpdateService(ctx, admin, newSvc.ID, domain.HospitalServiceInput{
		Name:        "Physiotherapy Session (Extended)",
		Description: "60-minute rehabilitation",
		Quantity:    1,
		RateMinor:   38000,
		Status:      1,
	})
	if err != nil {
		t.Fatalf("failed to update service: %v", err)
	}
	if updatedSvc.RateMinor != 38000 {
		t.Fatalf("expected rate minor 38000, got %d", updatedSvc.RateMinor)
	}

	// 4. Operations & Categories
	opCat, err := srv.CreateOperationCategory(ctx, admin, domain.OperationCategoryInput{
		Name: "Ophthalmology",
	})
	if err != nil {
		t.Fatalf("failed to create operation category: %v", err)
	}

	op, err := srv.CreateOperation(ctx, admin, domain.HospitalOperationInput{
		OperationCategoryID: opCat.ID,
		Name:                "Cataract Phacoemulsification",
		Description:         "Day-case cataract extraction with IOL implant",
		Status:              1,
	})
	if err != nil {
		t.Fatalf("failed to create operation: %v", err)
	}
	if op.OperationCategoryName != "Ophthalmology" {
		t.Fatalf("expected category name Ophthalmology, got %s", op.OperationCategoryName)
	}

	// Update operation
	upOp, err := srv.UpdateOperation(ctx, admin, op.ID, domain.HospitalOperationInput{
		OperationCategoryID: opCat.ID,
		Name:                "Cataract Phacoemulsification with Premium IOL",
		Description:         "Updated",
		Status:              1,
	})
	if err != nil {
		t.Fatalf("failed to update operation: %v", err)
	}
	if upOp.Name != "Cataract Phacoemulsification with Premium IOL" {
		t.Fatalf("unexpected operation name: %s", upOp.Name)
	}

	// 5. Custom Fields
	// Validation error: invalid grid
	_, err = srv.CreateCustomField(ctx, admin, domain.CustomFieldInput{
		ModuleName: "patient",
		FieldType:  "text",
		FieldName:  "Tax Identification Number",
		Grid:       15,
	})
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected ErrValidation for grid > 12, got %v", err)
	}

	cf, err := srv.CreateCustomField(ctx, admin, domain.CustomFieldInput{
		ModuleName: "patient",
		FieldType:  "text",
		FieldName:  "National ID / Kebele ID",
		IsRequired: true,
		Grid:       6,
	})
	if err != nil {
		t.Fatalf("failed to create custom field: %v", err)
	}

	cfs, err := srv.CustomFields(ctx, admin, "patient")
	if err != nil || len(cfs) == 0 {
		t.Fatalf("failed to list custom fields: %v", err)
	}

	// Delete custom field
	if err := srv.DeleteCustomField(ctx, admin, cf.ID); err != nil {
		t.Fatalf("failed to delete custom field: %v", err)
	}

	// 6. Module Settings
	settings, err := srv.ModuleSettings(ctx, admin)
	if err != nil || len(settings) == 0 {
		t.Fatalf("failed to get module settings: %v", err)
	}

	updatedMod, err := srv.UpdateModuleSetting(ctx, admin, "ambulances", false)
	if err != nil {
		t.Fatalf("failed to toggle module setting: %v", err)
	}
	if updatedMod.IsActive {
		t.Fatalf("expected ambulances module to be deactivated")
	}

	// Re-activate
	_, _ = srv.UpdateModuleSetting(ctx, admin, "ambulances", true)

	// 7. HTTP API Integration tests
	authSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"user": map[string]any{
				"id": admin.ID,
			},
			"session": map[string]any{
				"userId":    admin.ID,
				"expiresAt": time.Now().Add(time.Hour),
			},
		})
	}))
	defer authSrv.Close()

	httpHandler := httpapi.Server{
		ServicesOperations: srv,
		Actors:             store,
		AuthURL:            authSrv.URL,
		Origin:             "http://hospital.test",
		Client:             authSrv.Client(),
	}.Handler()

	// GET /v1/services
	req := httptest.NewRequest("GET", "/v1/services", nil)
	req.Header.Set("Cookie", "session=admin-token")
	req.Header.Set("Origin", "http://hospital.test")
	rec := httptest.NewRecorder()
	httpHandler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for GET /v1/services, got %d: %s", rec.Code, rec.Body.String())
	}

	// POST /v1/services
	svcPayload, _ := json.Marshal(domain.HospitalServiceInput{
		Name:      "HTTP Created Service",
		Quantity:  1,
		RateMinor: 12000,
		Status:    1,
	})
	req = httptest.NewRequest("POST", "/v1/services", bytes.NewReader(svcPayload))
	req.Header.Set("Cookie", "session=admin-token")
	req.Header.Set("Origin", "http://hospital.test")
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	httpHandler.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 for POST /v1/services, got %d: %s", rec.Code, rec.Body.String())
	}

	// GET /v1/module-settings
	req = httptest.NewRequest("GET", "/v1/module-settings", nil)
	req.Header.Set("Cookie", "session=admin-token")
	req.Header.Set("Origin", "http://hospital.test")
	rec = httptest.NewRecorder()
	httpHandler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for GET /v1/module-settings, got %d: %s", rec.Code, rec.Body.String())
	}
}
