package postgres

import (
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

func testPharmacyBloodBank(t *testing.T, db *pgxpool.Pool, store Store, actors []domain.Actor, patients []domain.Patient, cases []domain.Case) {
	t.Helper()
	ctx := context.Background()

	admin := actors[0]
	doctor := actors[1]
	otherDoctor := actors[2]
	patientUser := actors[3]
	receptionist := actors[4]

	srv := application.PharmacyBloodBankService{Store: store, Now: time.Now}

	// 1. Authorization checks
	// Patient cannot create medicine category
	_, err := srv.CreateMedicineCategory(ctx, patientUser, domain.MedicineCategoryInput{
		Name:     "Unauthorized Category",
		IsActive: true,
	})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected ErrForbidden for patient creating medicine category, got %v", err)
	}

	// Receptionist cannot create medicine brand
	_, err = srv.CreateMedicineBrand(ctx, receptionist, domain.MedicineBrandInput{
		Name:  "Unauthorized Brand",
		Email: "brand@test.com",
	})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected ErrForbidden for receptionist creating medicine brand, got %v", err)
	}

	// Patient cannot create blood donor
	now := time.Now()
	_, err = srv.CreateBloodDonor(ctx, patientUser, domain.BloodDonorInput{
		Name:           "Unauthorized Donor",
		Age:            25,
		Gender:         0,
		BloodGroup:     "A+",
		LastDonateDate: &now,
	})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected ErrForbidden for patient creating blood donor, got %v", err)
	}

	// Patient cannot record blood donation
	_, err = srv.RecordBloodDonation(ctx, patientUser, domain.BloodDonationInput{
		DonorID: "dummy-id",
		Bags:    1,
	})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected ErrForbidden for patient recording blood donation, got %v", err)
	}

	// Patient cannot create blood issue
	_, err = srv.CreateBloodIssue(ctx, patientUser, domain.BloodIssueInput{
		DoctorID:   doctor.ID,
		PatientID:  patients[0].ID,
		BloodGroup: "A+",
		Bags:       1,
		IssueDate:  &now,
	})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected ErrForbidden for patient creating blood issue, got %v", err)
	}

	// Patient cannot create prescription
	_, err = srv.CreatePrescription(ctx, patientUser, domain.PrescriptionInput{
		PatientID: patients[0].ID,
		Medicines: []domain.PrescriptionMedicineInput{
			{MedicineName: "Amoxicillin", Dosage: "500mg"},
		},
	})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected ErrForbidden for patient creating prescription, got %v", err)
	}

	// 2. Validation checks
	// Empty category name
	_, err = srv.CreateMedicineCategory(ctx, admin, domain.MedicineCategoryInput{
		Name: "",
	})
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected ErrValidation for empty category name, got %v", err)
	}

	// Invalid blood group
	_, err = srv.CreateBloodDonor(ctx, doctor, domain.BloodDonorInput{
		Name:           "Kidus Test",
		Age:            30,
		Gender:         0,
		BloodGroup:     "INVALID_GROUP",
		LastDonateDate: &now,
	})
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected ErrValidation for invalid blood group, got %v", err)
	}

	// Prescription with empty medicines
	_, err = srv.CreatePrescription(ctx, doctor, domain.PrescriptionInput{
		PatientID: patients[0].ID,
		Medicines: []domain.PrescriptionMedicineInput{},
	})
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected ErrValidation for prescription without medicines, got %v", err)
	}

	// 3. Medicine Categories & Brands CRUD
	cat, err := srv.CreateMedicineCategory(ctx, admin, domain.MedicineCategoryInput{
		Name:     "Cardiovascular Agents",
		IsActive: true,
	})
	if err != nil {
		t.Fatalf("failed to create medicine category: %v", err)
	}

	categories, cTotal, err := srv.MedicineCategories(ctx, doctor, 1, "Cardio")
	if err != nil || cTotal == 0 || len(categories) == 0 {
		t.Fatalf("failed to list medicine categories: %v, total: %d", err, cTotal)
	}

	updatedCat, err := srv.UpdateMedicineCategory(ctx, admin, cat.ID, domain.MedicineCategoryInput{
		Name:     "Cardiovascular & Antihypertensive Agents",
		IsActive: true,
	})
	if err != nil || updatedCat.Name != "Cardiovascular & Antihypertensive Agents" {
		t.Fatalf("failed to update medicine category: %v", err)
	}

	brand, err := srv.CreateMedicineBrand(ctx, admin, domain.MedicineBrandInput{
		Name:  "Ethiopian Pharmaceuticals Manufacturing Sh.Co.",
		Email: "info@epharm.et",
		Phone: "+251114421111",
	})
	if err != nil || brand.ID == "" {
		t.Fatalf("failed to create medicine brand: %v", err)
	}

	brands, bTotal, err := srv.MedicineBrands(ctx, doctor, 1, "EPHARM")
	if err != nil || bTotal == 0 || len(brands) == 0 {
		t.Fatalf("failed to list medicine brands: %v, total: %d", err, bTotal)
	}

	// 4. Blood Bank, Donors, Donations & Issues Lifecycle
	// Initial blood bank query: all 8 standard groups exist with 0 bags
	bankItems, err := srv.BloodBank(ctx, doctor)
	if err != nil || len(bankItems) != 8 {
		t.Fatalf("expected 8 blood bank groups, got %d, err: %v", len(bankItems), err)
	}

	// Create a blood donor
	donor, err := srv.CreateBloodDonor(ctx, doctor, domain.BloodDonorInput{
		Name:           "Mulugeta Tadesse",
		Age:            29,
		Gender:         0,
		BloodGroup:     "O+",
		LastDonateDate: &now,
	})
	if err != nil {
		t.Fatalf("failed to create blood donor: %v", err)
	}

	// Record a donation of 3 bags for O+
	donation, err := srv.RecordBloodDonation(ctx, doctor, domain.BloodDonationInput{
		DonorID: donor.ID,
		Bags:    3,
	})
	if err != nil {
		t.Fatalf("failed to record blood donation: %v", err)
	}
	if donation.Bags != 3 || donation.BloodGroup != "O+" {
		t.Fatalf("unexpected donation result: %+v", donation)
	}

	// Verify Blood Bank inventory updated to 3 bags for O+
	updatedBank, err := srv.BloodBank(ctx, doctor)
	if err != nil {
		t.Fatalf("failed to fetch updated blood bank: %v", err)
	}
	var oPosBags int
	for _, b := range updatedBank {
		if b.BloodGroup == "O+" {
			oPosBags = b.RemainedBags
		}
	}
	if oPosBags != 3 {
		t.Fatalf("expected 3 remained bags for O+, got %d", oPosBags)
	}

	// Attempt to issue 5 bags for O+ (more than available) -> fails with ErrValidation
	_, err = srv.CreateBloodIssue(ctx, doctor, domain.BloodIssueInput{
		DoctorID:    doctor.ID,
		PatientID:   patients[0].ID,
		BloodGroup:  "O+",
		Bags:        5,
		AmountMinor: 50000,
		IssueDate:   &now,
		Remarks:     "Emergency transfusion request",
	})
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected ErrValidation for issuing more blood bags than available, got %v", err)
	}

	// Issue 1 bag for O+ -> succeeds
	issue, err := srv.CreateBloodIssue(ctx, doctor, domain.BloodIssueInput{
		DoctorID:    doctor.ID,
		DonorID:     &donor.ID,
		PatientID:   patients[0].ID,
		BloodGroup:  "O+",
		Bags:        1,
		AmountMinor: 50000,
		IssueDate:   &now,
		Remarks:     "Transfusion for elective surgery",
	})
	if err != nil {
		t.Fatalf("failed to create blood issue: %v", err)
	}
	if issue.Bags != 1 || issue.BloodGroup != "O+" {
		t.Fatalf("unexpected blood issue result: %+v", issue)
	}

	// Verify inventory decremented to 2 bags
	bankAfterIssue, err := srv.BloodBank(ctx, doctor)
	if err != nil {
		t.Fatalf("failed to check blood bank after issue: %v", err)
	}
	for _, b := range bankAfterIssue {
		if b.BloodGroup == "O+" && b.RemainedBags != 2 {
			t.Fatalf("expected 2 remained bags for O+ after issue, got %d", b.RemainedBags)
		}
	}

	// Patient (linked to patients[0]) queries blood issues and sees their issue
	patientIssues, pCount, err := srv.BloodIssues(ctx, patientUser, 1)
	if err != nil || pCount == 0 || len(patientIssues) == 0 {
		t.Fatalf("patient failed to list own blood issues: %v, count: %d", err, pCount)
	}

	// 5. Prescriptions Lifecycle
	prescription, err := srv.CreatePrescription(ctx, doctor, domain.PrescriptionInput{
		PatientID:          patients[0].ID,
		HighBloodPressure:  "130/85",
		Diabetic:           "No",
		ProblemDescription: "Acute bacterial pharyngitis and mild fever",
		Advice:             "Take medications after meals with full glass of water. Rest 3 days.",
		NextVisitQty:       "7",
		NextVisitTime:      "days",
		Medicines: []domain.PrescriptionMedicineInput{
			{
				MedicineName: "Amoxicillin 500mg Capsule",
				Dosage:       "1 cap three times daily",
				Day:          "7",
				Time:         "After Meals",
				Comment:      "Complete entire course",
			},
			{
				MedicineName: "Paracetamol 500mg Tablet",
				Dosage:       "1 tab every 8 hours as needed",
				Day:          "3",
				Time:         "After Meals",
				Comment:      "For fever and pain",
			},
		},
	})
	if err != nil {
		t.Fatalf("failed to create prescription: %v", err)
	}
	if prescription.Status != 0 || len(prescription.Medicines) != 2 {
		t.Fatalf("unexpected prescription created: %+v", prescription)
	}

	// Patient lists prescriptions
	prescriptions, prCount, err := srv.Prescriptions(ctx, patientUser, 1)
	if err != nil || prCount == 0 || len(prescriptions) == 0 {
		t.Fatalf("patient failed to list prescriptions: %v, count: %d", err, prCount)
	}

	// Read prescription details
	prDetails, err := srv.Prescription(ctx, doctor, prescription.ID)
	if err != nil || len(prDetails.Medicines) != 2 {
		t.Fatalf("failed to read prescription details: %v, medicines: %v", err, prDetails.Medicines)
	}

	// Unrelated user cannot read prescription details
	_, err = srv.Prescription(ctx, otherDoctor, prescription.ID)
	// (other doctor is staff, so prescriptions.read allows staff, but patientUser cannot read other patients' prescriptions)
	// Test patient isolation: if patientUser accesses prescription not linked to them -> ErrForbidden
	// Create another prescription for patients[1] (not linked to patientUser)
	otherPrescription, err := srv.CreatePrescription(ctx, doctor, domain.PrescriptionInput{
		PatientID: patients[1].ID,
		Medicines: []domain.PrescriptionMedicineInput{
			{MedicineName: "Ibuprofen 400mg", Dosage: "1 tab twice daily"},
		},
	})
	if err != nil {
		t.Fatalf("failed to create prescription for second patient: %v", err)
	}
	_, err = srv.Prescription(ctx, patientUser, otherPrescription.ID)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected ErrForbidden for patientUser accessing other patient's prescription, got %v", err)
	}

	// Doctor / pharmacist updates prescription status to 1 (dispensed)
	dispensed, err := srv.UpdatePrescriptionStatus(ctx, doctor, prescription.ID, 1)
	if err != nil || dispensed.Status != 1 {
		t.Fatalf("failed to update prescription status: %v", err)
	}

	// 6. HTTP API Integration
	authSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"user": map[string]any{
				"id": doctor.ID,
			},
			"session": map[string]any{
				"userId":    doctor.ID,
				"expiresAt": time.Now().Add(time.Hour),
			},
		})
	}))
	defer authSrv.Close()

	httpHandler := httpapi.Server{
		PharmacyBloodBank: srv,
		Actors:            store,
		AuthURL:           authSrv.URL,
		Origin:            "http://hospital.test",
		Client:            authSrv.Client(),
	}.Handler()

	// GET /v1/blood-bank
	req := httptest.NewRequest("GET", "/v1/blood-bank", nil)
	req.Header.Set("Cookie", "session=doctor-token")
	req.Header.Set("Origin", "http://hospital.test")
	rec := httptest.NewRecorder()
	httpHandler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for GET /v1/blood-bank, got %d: %s", rec.Code, rec.Body.String())
	}

	// GET /v1/blood-donors
	req = httptest.NewRequest("GET", "/v1/blood-donors", nil)
	req.Header.Set("Cookie", "session=doctor-token")
	req.Header.Set("Origin", "http://hospital.test")
	rec = httptest.NewRecorder()
	httpHandler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for GET /v1/blood-donors, got %d: %s", rec.Code, rec.Body.String())
	}

	// GET /v1/medicine-categories
	req = httptest.NewRequest("GET", "/v1/medicine-categories", nil)
	req.Header.Set("Cookie", "session=doctor-token")
	req.Header.Set("Origin", "http://hospital.test")
	rec = httptest.NewRecorder()
	httpHandler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for GET /v1/medicine-categories, got %d: %s", rec.Code, rec.Body.String())
	}

	// GET /v1/medicine-brands
	req = httptest.NewRequest("GET", "/v1/medicine-brands", nil)
	req.Header.Set("Cookie", "session=doctor-token")
	req.Header.Set("Origin", "http://hospital.test")
	rec = httptest.NewRecorder()
	httpHandler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for GET /v1/medicine-brands, got %d: %s", rec.Code, rec.Body.String())
	}

	// GET /v1/prescriptions
	req = httptest.NewRequest("GET", "/v1/prescriptions", nil)
	req.Header.Set("Cookie", "session=doctor-token")
	req.Header.Set("Origin", "http://hospital.test")
	rec = httptest.NewRecorder()
	httpHandler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for GET /v1/prescriptions, got %d: %s", rec.Code, rec.Body.String())
	}
}
