package postgres

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"hms.local/api/internal/adapters/httpapi"
	"hms.local/api/internal/application"
	"hms.local/api/internal/domain"
)

func testDoctors(t *testing.T, db *pgxpool.Pool, store Store, actors []domain.Actor) {
	t.Helper()
	ctx := context.Background()

	var admin, docActor, otherDocActor, receptionist, patientUser domain.Actor
	for _, a := range actors {
		switch a.ID {
		case "admin":
			admin = a
		case "doctor":
			docActor = a
		case "other-doctor":
			otherDocActor = a
		case "reception":
			receptionist = a
		case "patient":
			patientUser = a
		}
	}
	if receptionist.ID == "" {
		receptionist = domain.Actor{ID: "reception", Role: "receptionist"}
	}
	nurse := domain.Actor{ID: "nurse-user", Role: "nurse"}
	accountant := domain.Actor{ID: "accountant-user", Role: "accountant"}

	// Insert nurse and accountant in user and staff_access for auth tests
	for _, a := range []domain.Actor{nurse, accountant} {
		_, _ = db.Exec(ctx, `INSERT INTO "user"(id,name,email) VALUES($1,$1,$1||'@example.test') ON CONFLICT DO NOTHING`, a.ID)
		_, _ = db.Exec(ctx, `INSERT INTO staff_access(user_id,role,active) VALUES($1,$2,true) ON CONFLICT DO NOTHING`, a.ID, a.Role)
	}

	// 1. Seed clinical departments: active and archived
	var activeDeptID, archivedDeptID string
	err := db.QueryRow(ctx, `
		INSERT INTO doctor_department (title, description, archived)
		VALUES ('Cardiology Unit', 'Cardiovascular diseases', false)
		RETURNING id::text
	`).Scan(&activeDeptID)
	if err != nil {
		t.Fatalf("failed to insert active doctor department: %v", err)
	}

	err = db.QueryRow(ctx, `
		INSERT INTO doctor_department (title, description, archived)
		VALUES ('Deprecated Surgery', 'Old surgical wing', true)
		RETURNING id::text
	`).Scan(&archivedDeptID)
	if err != nil {
		t.Fatalf("failed to insert archived doctor department: %v", err)
	}

	// 2. Seed test doctor users in user and staff_access
	docUser1 := "doctor-test-1"
	docUser2 := "doctor-test-2"
	nonDocUser := "nondoctor-candidate"

	for _, u := range []struct {
		id, name, role string
	}{
		{docUser1, "Dr. Alice Smith", "doctor"},
		{docUser2, "Dr. Bob Jones", "doctor"},
		{nonDocUser, "Nurse Candidate", "nurse"},
	} {
		_, err = db.Exec(ctx, `INSERT INTO "user"(id,name,email) VALUES($1,$2,$1||'@hospital.test') ON CONFLICT DO NOTHING`, u.id, u.name)
		if err != nil {
			t.Fatalf("failed to seed user %s: %v", u.id, err)
		}
		_, err = db.Exec(ctx, `INSERT INTO staff_access(user_id,role,active) VALUES($1,$2,true) ON CONFLICT DO NOTHING`, u.id, u.role)
		if err != nil {
			t.Fatalf("failed to seed staff_access for %s: %v", u.id, err)
		}
	}

	sched := application.Scheduling{Store: store, Now: time.Now}

	// -------------------------------------------------------------
	// Test A: Doctor creation flow with full fields and default schedule
	// -------------------------------------------------------------
	createIn := domain.CreateDoctorInput{
		UserID:            docUser1,
		DepartmentID:      activeDeptID,
		Specialist:        "Senior Cardiologist",
		Designation:       "Head of Department",
		Qualification:     "MD, FACC, FSCAI",
		Phone:             "+251911223344",
		Gender:            "female",
		DateOfBirth:       "1980-05-12",
		BloodGroup:        "O+",
		Address1:          "Bole Road 123",
		Address2:          "Suite 401",
		City:              "Addis Ababa",
		Zip:               "1000",
		Description:       "Cardiology specialist with 15 years experience.",
		PhotoURL:          "https://cdn.hospital.test/photos/dr-alice.jpg",
		AppointmentCharge: 500,
		OpdCharge:         350,
		// SlotMinutes omitted to verify D2 default of 60 minutes
	}

	createdDoc, err := sched.CreateDoctorProfile(ctx, admin, createIn)
	if err != nil {
		t.Fatalf("failed to create doctor profile: %v", err)
	}
	if createdDoc.ID != docUser1 {
		t.Fatalf("expected doctor ID %s, got %s", docUser1, createdDoc.ID)
	}
	if createdDoc.Specialist != "Senior Cardiologist" {
		t.Fatalf("expected specialist Senior Cardiologist, got %s", createdDoc.Specialist)
	}
	if createdDoc.Designation != "Head of Department" || createdDoc.Qualification != "MD, FACC, FSCAI" {
		t.Fatalf("expected designation and qualification to match, got %s, %s", createdDoc.Designation, createdDoc.Qualification)
	}
	if createdDoc.Phone != "+251911223344" || createdDoc.BloodGroup != "O+" || createdDoc.City != "Addis Ababa" {
		t.Fatalf("expected staff details to match: %+v", createdDoc)
	}
	if createdDoc.AppointmentCharge != 500 || createdDoc.OpdCharge != 350 {
		t.Fatalf("expected charges 500 and 350, got %f and %f", createdDoc.AppointmentCharge, createdDoc.OpdCharge)
	}
	if !createdDoc.Active {
		t.Fatal("expected newly created doctor to be active")
	}
	if createdDoc.Version != 1 {
		t.Fatalf("expected version 1, got %d", createdDoc.Version)
	}
	// D2: Verify default 60-minute slot duration
	if createdDoc.SlotMinutes != 60 {
		t.Fatalf("expected default slot duration 60 minutes, got %d", createdDoc.SlotMinutes)
	}
	// D2: Verify default 7-day hours (0..6, Sunday through Saturday, 10:00 to 19:30 / 600 to 1170)
	if len(createdDoc.Hours) != 7 {
		t.Fatalf("expected 7 default working hour blocks, got %d", len(createdDoc.Hours))
	}
	for _, h := range createdDoc.Hours {
		if h.StartMinute != 600 || h.EndMinute != 1170 {
			t.Fatalf("expected default hours 10:00 (600) to 19:30 (1170), got %d to %d for weekday %d", h.StartMinute, h.EndMinute, h.Weekday)
		}
	}

	// Verify audit event
	var auditCount int
	err = db.QueryRow(ctx, `SELECT count(*) FROM audit_event WHERE action='doctor.created' AND resource_id=$1`, docUser1).Scan(&auditCount)
	if err != nil || auditCount != 1 {
		t.Fatalf("expected 1 audit event for doctor.created, got %d, err: %v", auditCount, err)
	}

	// -------------------------------------------------------------
	// Test B: Doctor details retrieval and role scoping
	// -------------------------------------------------------------
	fetchedByAdmin, err := sched.Doctor(ctx, admin, docUser1)
	if err != nil || fetchedByAdmin.ID != docUser1 {
		t.Fatalf("admin failed to fetch doctor: %v", err)
	}

	// Receptionist can fetch doctor
	fetchedByRecep, err := sched.Doctor(ctx, receptionist, docUser1)
	if err != nil || fetchedByRecep.ID != docUser1 {
		t.Fatalf("receptionist failed to fetch doctor: %v", err)
	}

	// Nurse can fetch doctor
	fetchedByNurse, err := sched.Doctor(ctx, nurse, docUser1)
	if err != nil || fetchedByNurse.ID != docUser1 {
		t.Fatalf("nurse failed to fetch doctor: %v", err)
	}

	// Patient cannot fetch individual doctor via internal route
	_, err = sched.Doctor(ctx, patientUser, docUser1)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected ErrForbidden for patient fetching doctor, got %v", err)
	}

	// Doctor viewing another doctor is blocked
	_, err = sched.Doctor(ctx, otherDocActor, docUser1)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected ErrForbidden for doctor viewing another doctor, got %v", err)
	}

	// Doctor viewing self succeeds
	selfDoctorActor := domain.Actor{ID: docUser1, Role: "doctor"}
	fetchedBySelf, err := sched.Doctor(ctx, selfDoctorActor, docUser1)
	if err != nil || fetchedBySelf.ID != docUser1 {
		t.Fatalf("doctor viewing self failed: %v", err)
	}

	// -------------------------------------------------------------
	// Test C: Update doctor profile with optimistic concurrency
	// -------------------------------------------------------------
	newSpec := "Lead Interventional Cardiologist"
	newCharge := 600.0
	newPhone := "+251922334455"
	newHours := []domain.DoctorHours{
		{Weekday: 1, StartMinute: 480, EndMinute: 960},
		{Weekday: 3, StartMinute: 480, EndMinute: 960},
	}
	updateIn := domain.UpdateDoctorInput{
		Specialist:        &newSpec,
		AppointmentCharge: &newCharge,
		Phone:             &newPhone,
		Hours:             newHours,
		Version:           1, // matches current version
	}

	updatedDoc, err := sched.UpdateDoctorProfile(ctx, admin, docUser1, updateIn)
	if err != nil {
		t.Fatalf("failed to update doctor profile: %v", err)
	}
	if updatedDoc.Specialist != newSpec {
		t.Fatalf("expected updated specialist %s, got %s", newSpec, updatedDoc.Specialist)
	}
	if updatedDoc.AppointmentCharge != 600 {
		t.Fatalf("expected updated appointment charge 600, got %f", updatedDoc.AppointmentCharge)
	}
	if updatedDoc.Phone != newPhone {
		t.Fatalf("expected updated phone %s, got %s", newPhone, updatedDoc.Phone)
	}
	if updatedDoc.Version != 2 {
		t.Fatalf("expected version 2 after update, got %d", updatedDoc.Version)
	}
	if len(updatedDoc.Hours) != 2 {
		t.Fatalf("expected 2 updated hours, got %d", len(updatedDoc.Hours))
	}

	// -------------------------------------------------------------
	// Test D: Stale version rejection (409 Conflict)
	// -------------------------------------------------------------
	staleIn := domain.UpdateDoctorInput{
		Specialist: &newSpec,
		Version:    1, // Stale! Current version is 2
	}
	_, err = sched.UpdateDoctorProfile(ctx, admin, docUser1, staleIn)
	if !errors.Is(err, domain.ErrStale) {
		t.Fatalf("expected ErrStale on update with old version, got %v", err)
	}

	// -------------------------------------------------------------
	// Test E: Department validation on create and update
	// -------------------------------------------------------------
	// 1. Nonexistent department on create
	nonExistentIn := domain.CreateDoctorInput{
		UserID:        docUser2,
		DepartmentID:  "00000000-0000-0000-0000-000000000000",
		Specialist:    "Neurologist",
		Designation:   "Consultant",
		Qualification: "MBBS",
		Gender:        "male",
	}
	_, err = sched.CreateDoctorProfile(ctx, admin, nonExistentIn)
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected ErrValidation for nonexistent department on create, got %v", err)
	}

	// 2. Archived department on create
	archivedIn := domain.CreateDoctorInput{
		UserID:        docUser2,
		DepartmentID:  archivedDeptID,
		Specialist:    "Neurologist",
		Designation:   "Consultant",
		Qualification: "MBBS",
		Gender:        "male",
	}
	_, err = sched.CreateDoctorProfile(ctx, admin, archivedIn)
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected ErrValidation for archived department on create, got %v", err)
	}

	// 3. Archived department on update
	archivedUpdateIn := domain.UpdateDoctorInput{
		DepartmentID: &archivedDeptID,
		Version:      2,
	}
	_, err = sched.UpdateDoctorProfile(ctx, admin, docUser1, archivedUpdateIn)
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected ErrValidation for changing to archived department, got %v", err)
	}

	// 4. Doctor attempting to change own department is forbidden
	doctorUpdateDeptIn := domain.UpdateDoctorInput{
		DepartmentID: &activeDeptID,
		Version:      2,
	}
	_, err = sched.UpdateDoctorProfile(ctx, selfDoctorActor, docUser1, doctorUpdateDeptIn)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected ErrForbidden when doctor attempts to change own department, got %v", err)
	}

	// 5. D3 Regression: missing required fields on create
	for _, tc := range []struct {
		name string
		in   domain.CreateDoctorInput
	}{
		{"missing designation", domain.CreateDoctorInput{UserID: docUser2, DepartmentID: activeDeptID, Specialist: "Neuro", Qualification: "MBBS", Gender: "male"}},
		{"blank designation", domain.CreateDoctorInput{UserID: docUser2, DepartmentID: activeDeptID, Specialist: "Neuro", Designation: "   ", Qualification: "MBBS", Gender: "male"}},
		{"missing qualification", domain.CreateDoctorInput{UserID: docUser2, DepartmentID: activeDeptID, Specialist: "Neuro", Designation: "Cons", Gender: "male"}},
		{"blank qualification", domain.CreateDoctorInput{UserID: docUser2, DepartmentID: activeDeptID, Specialist: "Neuro", Designation: "Cons", Qualification: "  ", Gender: "male"}},
		{"missing gender", domain.CreateDoctorInput{UserID: docUser2, DepartmentID: activeDeptID, Specialist: "Neuro", Designation: "Cons", Qualification: "MBBS"}},
		{"invalid gender", domain.CreateDoctorInput{UserID: docUser2, DepartmentID: activeDeptID, Specialist: "Neuro", Designation: "Cons", Qualification: "MBBS", Gender: "nonbinary"}},
	} {
		_, err := sched.CreateDoctorProfile(ctx, admin, tc.in)
		if !errors.Is(err, domain.ErrValidation) {
			t.Fatalf("expected ErrValidation for %s on create, got %v", tc.name, err)
		}
	}

	// 6. D3 Regression: blank required fields on update
	blankStr := "   "
	for _, tc := range []struct {
		name string
		in   domain.UpdateDoctorInput
	}{
		{"blank designation", domain.UpdateDoctorInput{Designation: &blankStr, Version: 2}},
		{"blank qualification", domain.UpdateDoctorInput{Qualification: &blankStr, Version: 2}},
		{"blank specialist", domain.UpdateDoctorInput{Specialist: &blankStr, Version: 2}},
		{"blank gender", domain.UpdateDoctorInput{Gender: &blankStr, Version: 2}},
	} {
		_, err := sched.UpdateDoctorProfile(ctx, admin, docUser1, tc.in)
		if !errors.Is(err, domain.ErrValidation) {
			t.Fatalf("expected ErrValidation for %s on update, got %v", tc.name, err)
		}
	}

	// -------------------------------------------------------------
	// Test F: Identity checks: non-doctor candidate and conflict
	// -------------------------------------------------------------
	// Candidate has role='nurse', not 'doctor'
	invalidRoleIn := domain.CreateDoctorInput{
		UserID:        nonDocUser,
		DepartmentID:  activeDeptID,
		Specialist:    "Pediatrician",
		Designation:   "Consultant",
		Qualification: "MBBS",
		Gender:        "female",
	}
	_, err = sched.CreateDoctorProfile(ctx, admin, invalidRoleIn)
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected ErrValidation for candidate without role='doctor', got %v", err)
	}

	// Conflict check: doctor profile already exists for docUser1
	dupIn := domain.CreateDoctorInput{
		UserID:        docUser1,
		DepartmentID:  activeDeptID,
		Specialist:    "Duplicate Doctor",
		Designation:   "Consultant",
		Qualification: "MBBS",
		Gender:        "female",
	}
	_, err = sched.CreateDoctorProfile(ctx, admin, dupIn)
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("expected ErrConflict for duplicate doctor creation, got %v", err)
	}

	// -------------------------------------------------------------
	// Test G: Status toggle (deactivation & reactivation)
	// -------------------------------------------------------------
	// Seed active session for docUser1
	_, err = db.Exec(ctx, `
		INSERT INTO session(id, "userId", token, "expiresAt", "createdAt", "updatedAt")
		VALUES ('test-sess-1', $1, 'tok-1', now() + interval '1 hour', now(), now())
	`, docUser1)
	if err != nil {
		t.Fatalf("failed to insert test session: %v", err)
	}

	// Deactivate doctor
	err = sched.SetDoctorStatus(ctx, admin, docUser1, false)
	if err != nil {
		t.Fatalf("failed to deactivate doctor: %v", err)
	}

	// Verify doctor is inactive
	docAfterDeact, err := sched.Doctor(ctx, admin, docUser1)
	if err != nil || docAfterDeact.Active {
		t.Fatalf("expected doctor active=false, got %v, active: %t", err, docAfterDeact.Active)
	}

	// Verify session was revoked
	var sessCount int
	_ = db.QueryRow(ctx, `SELECT count(*) FROM session WHERE "userId"=$1`, docUser1).Scan(&sessCount)
	if sessCount != 0 {
		t.Fatalf("expected session to be revoked on doctor deactivation, found %d", sessCount)
	}

	// Verify status audit event
	var statusAuditCount int
	_ = db.QueryRow(ctx, `SELECT count(*) FROM audit_event WHERE action='doctor.status_changed' AND resource_id=$1`, docUser1).Scan(&statusAuditCount)
	if statusAuditCount != 1 {
		t.Fatalf("expected audit event for doctor.status_changed, got %d", statusAuditCount)
	}

	// Reactivate doctor
	err = sched.SetDoctorStatus(ctx, admin, docUser1, true)
	if err != nil {
		t.Fatalf("failed to reactivate doctor: %v", err)
	}
	docAfterReact, _ := sched.Doctor(ctx, admin, docUser1)
	if !docAfterReact.Active {
		t.Fatal("expected doctor to be active after reactivation")
	}

	// -------------------------------------------------------------
	// Test H: List filtering (status, department, search)
	// -------------------------------------------------------------
	// Create second doctor for listing tests
	createDoc2In := domain.CreateDoctorInput{
		UserID:        docUser2,
		DepartmentID:  activeDeptID,
		Specialist:    "Pediatric Neurologist",
		Designation:   "Consultant",
		Qualification: "MBBS, MD",
		Gender:        "male",
		SlotMinutes:   20,
	}
	_, err = sched.CreateDoctorProfile(ctx, admin, createDoc2In)
	if err != nil {
		t.Fatalf("failed to create second doctor: %v", err)
	}

	// Deactivate second doctor
	err = sched.SetDoctorStatus(ctx, admin, docUser2, false)
	if err != nil {
		t.Fatalf("failed to deactivate second doctor: %v", err)
	}

	// Admin list with status="active" returns docUser1, but NOT docUser2
	activeList, err := sched.ListDoctors(ctx, admin, "active", "", "")
	if err != nil {
		t.Fatalf("failed to list active doctors: %v", err)
	}
	foundDoc1 := false
	foundDoc2 := false
	for _, d := range activeList {
		if d.ID == docUser1 {
			foundDoc1 = true
		}
		if d.ID == docUser2 {
			foundDoc2 = true
		}
	}
	if !foundDoc1 || foundDoc2 {
		t.Fatalf("expected only active doctors in active list (doc1=%t, doc2=%t)", foundDoc1, foundDoc2)
	}

	// Admin list with status="inactive" returns docUser2
	inactiveList, err := sched.ListDoctors(ctx, admin, "inactive", "", "")
	if err != nil {
		t.Fatalf("failed to list inactive doctors: %v", err)
	}
	foundInactiveDoc2 := false
	for _, d := range inactiveList {
		if d.ID == docUser2 {
			foundInactiveDoc2 = true
		}
	}
	if !foundInactiveDoc2 {
		t.Fatal("expected inactive doctor in inactive list")
	}

	// Non-admin requesting status="inactive" is coerced to "active"
	nonAdminList, err := sched.ListDoctors(ctx, receptionist, "inactive", "", "")
	if err != nil {
		t.Fatalf("failed to list doctors as receptionist: %v", err)
	}
	for _, d := range nonAdminList {
		if d.ID == docUser2 {
			t.Fatal("receptionist must not see inactive doctors even when requesting inactive")
		}
	}

	// Search filter
	searchList, err := sched.ListDoctors(ctx, admin, "all", "", "Pediatric")
	if err != nil {
		t.Fatalf("failed to search doctors: %v", err)
	}
	if len(searchList) != 1 || searchList[0].ID != docUser2 {
		t.Fatalf("expected search for 'Pediatric' to return docUser2, got %+v", searchList)
	}

	// -------------------------------------------------------------
	// Test I: HTTP API endpoint integration
	// -------------------------------------------------------------
	authSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		cookie := r.Header.Get("Cookie")
		switch cookie {
		case "session=admin":
			_ = json.NewEncoder(w).Encode(map[string]any{"user": map[string]string{"id": admin.ID}, "session": map[string]any{"userId": admin.ID, "expiresAt": time.Now().Add(time.Hour)}})
		case "session=doctor":
			_ = json.NewEncoder(w).Encode(map[string]any{"user": map[string]string{"id": docActor.ID}, "session": map[string]any{"userId": docActor.ID, "expiresAt": time.Now().Add(time.Hour)}})
		case "session=other-doctor":
			_ = json.NewEncoder(w).Encode(map[string]any{"user": map[string]string{"id": otherDocActor.ID}, "session": map[string]any{"userId": otherDocActor.ID, "expiresAt": time.Now().Add(time.Hour)}})
		case "session=nurse":
			_ = json.NewEncoder(w).Encode(map[string]any{"user": map[string]string{"id": nurse.ID}, "session": map[string]any{"userId": nurse.ID, "expiresAt": time.Now().Add(time.Hour)}})
		case "session=receptionist":
			_ = json.NewEncoder(w).Encode(map[string]any{"user": map[string]string{"id": receptionist.ID}, "session": map[string]any{"userId": receptionist.ID, "expiresAt": time.Now().Add(time.Hour)}})
		case "session=patient":
			_ = json.NewEncoder(w).Encode(map[string]any{"user": map[string]string{"id": patientUser.ID}, "session": map[string]any{"userId": patientUser.ID, "expiresAt": time.Now().Add(time.Hour)}})
		default:
			w.WriteHeader(http.StatusUnauthorized)
		}
	}))
	defer authSrv.Close()

	httpHandler := httpapi.Server{
		Scheduling: sched,
		Actors:     store,
		AuthURL:    authSrv.URL,
		Origin:     "http://hospital.test",
		Client:     authSrv.Client(),
	}.Handler()

	// 1. Anonymous GET /v1/doctors -> 401
	anonReq := httptest.NewRequest("GET", "/v1/doctors", nil)
	anonRec := httptest.NewRecorder()
	httpHandler.ServeHTTP(anonRec, anonReq)
	if anonRec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for anonymous GET /v1/doctors, got %d", anonRec.Code)
	}

	// 2. Admin GET /v1/doctors?status=all -> 200
	adminListReq := httptest.NewRequest("GET", "/v1/doctors?status=all", nil)
	adminListReq.Header.Set("Cookie", "session=admin")
	adminListReq.Header.Set("Origin", "http://hospital.test")
	adminListRec := httptest.NewRecorder()
	httpHandler.ServeHTTP(adminListRec, adminListReq)
	if adminListRec.Code != http.StatusOK {
		t.Fatalf("expected 200 for admin GET /v1/doctors, got %d", adminListRec.Code)
	}

	// 3. Patient POST /v1/doctors -> 403 Forbidden
	patPostReq := httptest.NewRequest("POST", "/v1/doctors", bytes.NewReader([]byte(`{"userId":"pat","specialist":"x"}`)))
	patPostReq.Header.Set("Cookie", "session=patient")
	patPostReq.Header.Set("Origin", "http://hospital.test")
	patPostReq.Header.Set("Content-Type", "application/json")
	patPostRec := httptest.NewRecorder()
	httpHandler.ServeHTTP(patPostRec, patPostReq)
	if patPostRec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for patient POST /v1/doctors, got %d", patPostRec.Code)
	}

	// 4. Admin GET /v1/doctors/{id} -> 200
	adminGetReq := httptest.NewRequest("GET", "/v1/doctors/"+docUser1, nil)
	adminGetReq.Header.Set("Cookie", "session=admin")
	adminGetReq.Header.Set("Origin", "http://hospital.test")
	adminGetRec := httptest.NewRecorder()
	httpHandler.ServeHTTP(adminGetRec, adminGetReq)
	if adminGetRec.Code != http.StatusOK {
		t.Fatalf("expected 200 for admin GET /v1/doctors/{id}, got %d: %s", adminGetRec.Code, adminGetRec.Body.String())
	}

	// 5. Admin PUT /v1/doctors/{id} with stale version -> 409 Conflict
	stalePutReq := httptest.NewRequest("PUT", "/v1/doctors/"+docUser1, bytes.NewReader([]byte(`{"specialist":"Conflict Test","version":1}`)))
	stalePutReq.Header.Set("Cookie", "session=admin")
	stalePutReq.Header.Set("Origin", "http://hospital.test")
	stalePutReq.Header.Set("Content-Type", "application/json")
	stalePutRec := httptest.NewRecorder()
	httpHandler.ServeHTTP(stalePutRec, stalePutReq)
	if stalePutRec.Code != http.StatusConflict {
		t.Fatalf("expected 409 for stale version PUT /v1/doctors/{id}, got %d", stalePutRec.Code)
	}

	// 6. D5 Regression: Invalid / conflicting status payloads MUST return 422 with ZERO mutations
	auditBeforeInvalid := 0
	_ = db.QueryRow(ctx, `SELECT count(*) FROM audit_event WHERE action='doctor.status_changed' AND resource_id=$1`, docUser1).Scan(&auditBeforeInvalid)

	for _, invalidBody := range []string{
		`{"status":99}`,
		`{"status":-1}`,
		`{"status":2}`,
		`{"active":true,"status":0}`,
		`{"active":false,"status":1}`,
		`{}`,
	} {
		req := httptest.NewRequest("PATCH", "/v1/doctors/"+docUser1+"/status", bytes.NewReader([]byte(invalidBody)))
		req.Header.Set("Cookie", "session=admin")
		req.Header.Set("Origin", "http://hospital.test")
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		httpHandler.ServeHTTP(rec, req)
		if rec.Code != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422 for invalid status body %s, got %d: %s", invalidBody, rec.Code, rec.Body.String())
		}
	}

	// Verify no mutations happened from invalid status requests
	auditAfterInvalid := 0
	_ = db.QueryRow(ctx, `SELECT count(*) FROM audit_event WHERE action='doctor.status_changed' AND resource_id=$1`, docUser1).Scan(&auditAfterInvalid)
	if auditAfterInvalid != auditBeforeInvalid {
		t.Fatalf("audit count mutated on invalid status requests: before=%d, after=%d", auditBeforeInvalid, auditAfterInvalid)
	}

	// 7. Valid status patch with status=0
	validPatchReq := httptest.NewRequest("PATCH", "/v1/doctors/"+docUser1+"/status", bytes.NewReader([]byte(`{"status":0}`)))
	validPatchReq.Header.Set("Cookie", "session=admin")
	validPatchReq.Header.Set("Origin", "http://hospital.test")
	validPatchReq.Header.Set("Content-Type", "application/json")
	validPatchRec := httptest.NewRecorder()
	httpHandler.ServeHTTP(validPatchRec, validPatchReq)
	if validPatchRec.Code != http.StatusOK {
		t.Fatalf("expected 200 for PATCH /v1/doctors/{id}/status with status=0, got %d: %s", validPatchRec.Code, validPatchRec.Body.String())
	}

	// Idempotent repeat: calling again with status=0
	auditBeforeIdempotent := 0
	_ = db.QueryRow(ctx, `SELECT count(*) FROM audit_event WHERE action='doctor.status_changed' AND resource_id=$1`, docUser1).Scan(&auditBeforeIdempotent)
	repeatPatchReq := httptest.NewRequest("PATCH", "/v1/doctors/"+docUser1+"/status", bytes.NewReader([]byte(`{"status":0}`)))
	repeatPatchReq.Header.Set("Cookie", "session=admin")
	repeatPatchReq.Header.Set("Origin", "http://hospital.test")
	repeatPatchReq.Header.Set("Content-Type", "application/json")
	repeatPatchRec := httptest.NewRecorder()
	httpHandler.ServeHTTP(repeatPatchRec, repeatPatchReq)
	if repeatPatchRec.Code != http.StatusOK {
		t.Fatalf("expected 200 for idempotent repeat PATCH, got %d: %s", repeatPatchRec.Code, repeatPatchRec.Body.String())
	}
	auditAfterIdempotent := 0
	_ = db.QueryRow(ctx, `SELECT count(*) FROM audit_event WHERE action='doctor.status_changed' AND resource_id=$1`, docUser1).Scan(&auditAfterIdempotent)
	if auditAfterIdempotent != auditBeforeIdempotent {
		t.Fatalf("audit event created on idempotent no-op status update")
	}

	// Restore doctor1 to active for subsequent tests
	reactReq := httptest.NewRequest("PATCH", "/v1/doctors/"+docUser1+"/status", bytes.NewReader([]byte(`{"active":true}`)))
	reactReq.Header.Set("Cookie", "session=admin")
	reactReq.Header.Set("Origin", "http://hospital.test")
	reactReq.Header.Set("Content-Type", "application/json")
	reactRec := httptest.NewRecorder()
	httpHandler.ServeHTTP(reactRec, reactReq)
	if reactRec.Code != http.StatusOK {
		t.Fatalf("expected 200 for restoring doctor active status, got %d", reactRec.Code)
	}

	// 8. D1 Regression: Patient listing MUST NOT receive doctor DOB, blood group, or address
	patListReq := httptest.NewRequest("GET", "/v1/doctors", nil)
	patListReq.Header.Set("Cookie", "session=patient")
	patListRec := httptest.NewRecorder()
	httpHandler.ServeHTTP(patListRec, patListReq)
	if patListRec.Code != http.StatusOK {
		t.Fatalf("expected 200 for patient GET /v1/doctors, got %d: %s", patListRec.Code, patListRec.Body.String())
	}
	var patListResp struct {
		Doctors []domain.Doctor `json:"doctors"`
	}
	if err := json.Unmarshal(patListRec.Body.Bytes(), &patListResp); err != nil {
		t.Fatalf("failed to decode patient doctors list: %v", err)
	}
	if len(patListResp.Doctors) == 0 {
		t.Fatal("expected active doctors in patient directory")
	}
	for _, d := range patListResp.Doctors {
		if d.DateOfBirth != "" {
			t.Fatalf("D1 violation: patient received doctor %s dateOfBirth: %s", d.ID, d.DateOfBirth)
		}
		if d.BloodGroup != "" {
			t.Fatalf("D1 violation: patient received doctor %s bloodGroup: %s", d.ID, d.BloodGroup)
		}
		if d.Address1 != "" || d.Address2 != "" || d.City != "" || d.Zip != "" {
			t.Fatalf("D1 violation: patient received doctor %s home address: %s %s %s", d.ID, d.Address1, d.City, d.Zip)
		}
		if d.Phone != "" {
			t.Fatalf("D1 violation: patient received doctor %s personal phone: %s", d.ID, d.Phone)
		}
	}

	// 9. D1 Regression: Doctor listing other doctors MUST NOT receive other doctors' DOB, blood group, address
	docListReq := httptest.NewRequest("GET", "/v1/doctors", nil)
	docListReq.Header.Set("Cookie", "session=other-doctor")
	docListRec := httptest.NewRecorder()
	httpHandler.ServeHTTP(docListRec, docListReq)
	if docListRec.Code != http.StatusOK {
		t.Fatalf("expected 200 for doctor GET /v1/doctors, got %d: %s", docListRec.Code, docListRec.Body.String())
	}
	var docListResp struct {
		Doctors []domain.Doctor `json:"doctors"`
	}
	if err := json.Unmarshal(docListRec.Body.Bytes(), &docListResp); err != nil {
		t.Fatalf("failed to decode doctor list: %v", err)
	}
	for _, d := range docListResp.Doctors {
		if d.ID == docUser1 {
			if d.DateOfBirth != "" || d.BloodGroup != "" || d.Address1 != "" {
				t.Fatalf("D1 violation: doctor %s received other doctor private fields: dob=%s bg=%s addr=%s", otherDocActor.ID, d.DateOfBirth, d.BloodGroup, d.Address1)
			}
		}
	}

	// 10. Admin listing DOES receive private profile fields for administration
	admListReq := httptest.NewRequest("GET", "/v1/doctors", nil)
	admListReq.Header.Set("Cookie", "session=admin")
	admListRec := httptest.NewRecorder()
	httpHandler.ServeHTTP(admListRec, admListReq)
	if admListRec.Code != http.StatusOK {
		t.Fatalf("expected 200 for admin GET /v1/doctors, got %d: %s", admListRec.Code, admListRec.Body.String())
	}
	var admListResp struct {
		Doctors []domain.Doctor `json:"doctors"`
	}
	if err := json.Unmarshal(admListRec.Body.Bytes(), &admListResp); err != nil {
		t.Fatalf("failed to decode admin doctor list: %v", err)
	}
	foundDoc1WithPrivate := false
	for _, d := range admListResp.Doctors {
		if d.ID == docUser1 && d.DateOfBirth != "" && d.BloodGroup != "" && d.Address1 != "" {
			foundDoc1WithPrivate = true
		}
	}
	if !foundDoc1WithPrivate {
		t.Fatal("expected admin to receive doctor1 with populated private profile fields")
	}

	// 11. D4: Name and email editing on doctor profile with validation and conflict checks
	newName := "Dr. Alice M. Smith"
	newEmail := "alice.smith.updated@hospital.test"
	updatedIdentityDoc, err := sched.UpdateDoctorProfile(ctx, admin, docUser1, domain.UpdateDoctorInput{
		Name:    &newName,
		Email:   &newEmail,
		Version: 2,
	})
	if err != nil {
		t.Fatalf("expected successful name and email update, got %v", err)
	}
	if updatedIdentityDoc.Name != newName {
		t.Fatalf("expected updated name %q, got %q", newName, updatedIdentityDoc.Name)
	}
	if updatedIdentityDoc.Email != newEmail {
		t.Fatalf("expected updated email %q, got %q", newEmail, updatedIdentityDoc.Email)
	}
	if updatedIdentityDoc.Version != 3 {
		t.Fatalf("expected version 3 after identity update, got %d", updatedIdentityDoc.Version)
	}

	// Verify database persistence in "user" table
	var dbUserName, dbUserEmail string
	err = db.QueryRow(ctx, `SELECT name, email FROM "user" WHERE id=$1`, docUser1).Scan(&dbUserName, &dbUserEmail)
	if err != nil || dbUserName != newName || dbUserEmail != newEmail {
		t.Fatalf("expected user table to have updated name/email, got name=%q email=%q err=%v", dbUserName, dbUserEmail, err)
	}

	// Validation: blank name rejected
	blankName := "   "
	_, err = sched.UpdateDoctorProfile(ctx, admin, docUser1, domain.UpdateDoctorInput{
		Name:    &blankName,
		Version: 3,
	})
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected ErrValidation for blank name update, got %v", err)
	}

	// Validation: invalid email rejected
	invalidEmail := "not-an-email"
	_, err = sched.UpdateDoctorProfile(ctx, admin, docUser1, domain.UpdateDoctorInput{
		Email:   &invalidEmail,
		Version: 3,
	})
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected ErrValidation for invalid email update, got %v", err)
	}

	// Conflict: duplicate email of doctor-test-2 rejected
	dupEmail := docUser2 + "@hospital.test"
	_, err = sched.UpdateDoctorProfile(ctx, admin, docUser1, domain.UpdateDoctorInput{
		Email:   &dupEmail,
		Version: 3,
	})
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("expected ErrConflict for duplicate email update, got %v", err)
	}

	// HTTP PUT /v1/doctors/{id} with name and email
	httpPutBody := `{"name":"Dr. Alice Parity","email":"alice.parity@hospital.test","version":3}`
	putReq := httptest.NewRequest("PUT", "/v1/doctors/"+docUser1, bytes.NewReader([]byte(httpPutBody)))
	putReq.Header.Set("Cookie", "session=admin")
	putReq.Header.Set("Origin", "http://hospital.test")
	putReq.Header.Set("Content-Type", "application/json")
	putRec := httptest.NewRecorder()
	httpHandler.ServeHTTP(putRec, putReq)
	if putRec.Code != http.StatusOK {
		t.Fatalf("expected 200 for HTTP PUT /v1/doctors with name/email, got %d: %s", putRec.Code, putRec.Body.String())
	}
	var putDoc domain.Doctor
	if err := json.Unmarshal(putRec.Body.Bytes(), &putDoc); err != nil {
		t.Fatalf("failed to decode PUT response: %v", err)
	}
	if putDoc.Name != "Dr. Alice Parity" || putDoc.Email != "alice.parity@hospital.test" || putDoc.Version != 4 {
		t.Fatalf("unexpected PUT response doc: %+v", putDoc)
	}

	// HTTP PUT duplicate email returns 409
	dupPutBody := `{"email":"` + docUser2 + `@hospital.test","version":4}`
	dupPutReq := httptest.NewRequest("PUT", "/v1/doctors/"+docUser1, bytes.NewReader([]byte(dupPutBody)))
	dupPutReq.Header.Set("Cookie", "session=admin")
	dupPutReq.Header.Set("Origin", "http://hospital.test")
	dupPutReq.Header.Set("Content-Type", "application/json")
	dupPutRec := httptest.NewRecorder()
	httpHandler.ServeHTTP(dupPutRec, dupPutReq)
	if dupPutRec.Code != http.StatusConflict {
		t.Fatalf("expected 409 for HTTP PUT duplicate email, got %d: %s", dupPutRec.Code, dupPutRec.Body.String())
	}

	// 12. Strict email validation: reject malformed emails with 422 using shared fixtures without mutating account, profile version, or audit
	var beforeUserName, beforeUserEmail string
	err = db.QueryRow(ctx, `SELECT name, email FROM "user" WHERE id=$1`, docUser1).Scan(&beforeUserName, &beforeUserEmail)
	if err != nil {
		t.Fatalf("failed to query user before email test: %v", err)
	}
	var beforeDocVersion int
	err = db.QueryRow(ctx, `SELECT version FROM doctor_profile WHERE user_id=$1`, docUser1).Scan(&beforeDocVersion)
	if err != nil {
		t.Fatalf("failed to query doctor_profile before email test: %v", err)
	}
	var beforeAuditCount int
	err = db.QueryRow(ctx, `SELECT count(*) FROM audit_event WHERE resource_id=$1`, docUser1).Scan(&beforeAuditCount)
	if err != nil {
		t.Fatalf("failed to query audit_event before email test: %v", err)
	}

	var fixtureBytes []byte
	for _, p := range []string{
		"../../../../../docs/module-audit/doctor-email-fixtures.json",
		"../../../../docs/module-audit/doctor-email-fixtures.json",
		"docs/module-audit/doctor-email-fixtures.json",
	} {
		if data, readErr := os.ReadFile(p); readErr == nil {
			fixtureBytes = data
			break
		}
	}
	if len(fixtureBytes) == 0 {
		t.Fatalf("failed to read shared email fixtures from candidate paths")
	}
	var emailFixtures struct {
		Valid   []string `json:"valid"`
		Invalid []string `json:"invalid"`
	}
	if err := json.Unmarshal(fixtureBytes, &emailFixtures); err != nil {
		t.Fatalf("failed to unmarshal shared email fixtures: %v", err)
	}

	for _, malformedEmail := range emailFixtures.Invalid {
		body := fmt.Sprintf(`{"email":%q,"version":%d}`, malformedEmail, beforeDocVersion)
		badReq := httptest.NewRequest("PUT", "/v1/doctors/"+docUser1, bytes.NewReader([]byte(body)))
		badReq.Header.Set("Cookie", "session=admin")
		badReq.Header.Set("Origin", "http://hospital.test")
		badReq.Header.Set("Content-Type", "application/json")
		badRec := httptest.NewRecorder()
		httpHandler.ServeHTTP(badRec, badReq)
		if badRec.Code != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422 for malformed email %q, got %d: %s", malformedEmail, badRec.Code, badRec.Body.String())
		}
	}

	var afterUserName, afterUserEmail string
	err = db.QueryRow(ctx, `SELECT name, email FROM "user" WHERE id=$1`, docUser1).Scan(&afterUserName, &afterUserEmail)
	if err != nil {
		t.Fatalf("failed to query user after email test: %v", err)
	}
	var afterDocVersion int
	err = db.QueryRow(ctx, `SELECT version FROM doctor_profile WHERE user_id=$1`, docUser1).Scan(&afterDocVersion)
	if err != nil {
		t.Fatalf("failed to query doctor_profile after email test: %v", err)
	}
	var afterAuditCount int
	err = db.QueryRow(ctx, `SELECT count(*) FROM audit_event WHERE resource_id=$1`, docUser1).Scan(&afterAuditCount)
	if err != nil {
		t.Fatalf("failed to query audit_event after email test: %v", err)
	}

	if afterUserName != beforeUserName || afterUserEmail != beforeUserEmail {
		t.Fatalf("user account data changed on invalid email: before=(%s,%s), after=(%s,%s)", beforeUserName, beforeUserEmail, afterUserName, afterUserEmail)
	}
	if afterDocVersion != beforeDocVersion {
		t.Fatalf("doctor profile version changed on invalid email: before=%d, after=%d", beforeDocVersion, afterDocVersion)
	}
	if afterAuditCount != beforeAuditCount {
		t.Fatalf("audit event count changed on invalid email: before=%d, after=%d", beforeAuditCount, afterAuditCount)
	}

	// Verify valid email from shared fixtures succeeds and increments version
	validSharedEmail := emailFixtures.Valid[0]
	validBody := fmt.Sprintf(`{"email":%q,"version":%d}`, validSharedEmail, afterDocVersion)
	validReq := httptest.NewRequest("PUT", "/v1/doctors/"+docUser1, bytes.NewReader([]byte(validBody)))
	validReq.Header.Set("Cookie", "session=admin")
	validReq.Header.Set("Origin", "http://hospital.test")
	validReq.Header.Set("Content-Type", "application/json")
	validRec := httptest.NewRecorder()
	httpHandler.ServeHTTP(validRec, validReq)
	if validRec.Code != http.StatusOK {
		t.Fatalf("expected 200 for valid email %q, got %d: %s", validSharedEmail, validRec.Code, validRec.Body.String())
	}
	var updatedSharedEmail string
	var updatedSharedVersion int
	err = db.QueryRow(ctx, `SELECT email FROM "user" WHERE id=$1`, docUser1).Scan(&updatedSharedEmail)
	if err != nil || updatedSharedEmail != validSharedEmail {
		t.Fatalf("expected updated email %q, got %q (err=%v)", validSharedEmail, updatedSharedEmail, err)
	}
	err = db.QueryRow(ctx, `SELECT version FROM doctor_profile WHERE user_id=$1`, docUser1).Scan(&updatedSharedVersion)
	if err != nil || updatedSharedVersion != afterDocVersion+1 {
		t.Fatalf("expected incremented version %d, got %d (err=%v)", afterDocVersion+1, updatedSharedVersion, err)
	}

	// 13. Transaction rollback tests:
	// 13a. Pre-insert validation failure (archived department): verify ErrValidation and 0 partial records
	deliberateDocID := "deliberate-fail-doctor"
	_, err = db.Exec(ctx, `INSERT INTO "user"(id,name,email) VALUES($1,$2,$1||'@hospital.test')`, deliberateDocID, "Dr. Deliberate Fail")
	if err != nil {
		t.Fatalf("failed to insert test user: %v", err)
	}
	_, err = db.Exec(ctx, `INSERT INTO staff_access(user_id,role,active) VALUES($1,'doctor',true)`, deliberateDocID)
	if err != nil {
		t.Fatalf("failed to insert test staff_access: %v", err)
	}

	deliberateIn := domain.CreateDoctorInput{
		UserID:        deliberateDocID,
		DepartmentID:  archivedDeptID,
		Specialist:    "Cardiology",
		Designation:   "Consultant",
		Qualification: "MBBS",
		Gender:        "male",
	}
	_, err = sched.CreateDoctorProfile(ctx, admin, deliberateIn)
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected ErrValidation for archived department, got %v", err)
	}

	var profileCount, hoursCount, auditDoctorCount int
	if err := db.QueryRow(ctx, `SELECT count(*) FROM doctor_profile WHERE user_id=$1`, deliberateDocID).Scan(&profileCount); err != nil {
		t.Fatalf("failed to query doctor_profile count: %v", err)
	}
	if err := db.QueryRow(ctx, `SELECT count(*) FROM doctor_hours WHERE doctor_id=$1`, deliberateDocID).Scan(&hoursCount); err != nil {
		t.Fatalf("failed to query doctor_hours count: %v", err)
	}
	if err := db.QueryRow(ctx, `SELECT count(*) FROM audit_event WHERE action='doctor.created' AND resource_id=$1`, deliberateDocID).Scan(&auditDoctorCount); err != nil {
		t.Fatalf("failed to query audit_event count: %v", err)
	}

	if profileCount != 0 {
		t.Fatalf("expected 0 doctor_profile records after archived-dept rejection, got %d", profileCount)
	}
	if hoursCount != 0 {
		t.Fatalf("expected 0 doctor_hours records after archived-dept rejection, got %d", hoursCount)
	}
	if auditDoctorCount != 0 {
		t.Fatalf("expected 0 audit_event records after archived-dept rejection, got %d", auditDoctorCount)
	}

	// 13b. Post-profile-write transaction rollback:
	// Inject a failure on the audit_event insert (which executes AFTER doctor_profile, staff_profile, and doctor_hours within CreateDoctorProfile)
	// and verify that the transaction rollback leaves zero partial records behind.
	postWriteDocID := "fail-audit-doctor"
	_, err = db.Exec(ctx, `INSERT INTO "user"(id,name,email) VALUES($1,$2,$1||'@hospital.test')`, postWriteDocID, "Dr. Post Write Rollback")
	if err != nil {
		t.Fatalf("failed to insert test user for post-write rollback: %v", err)
	}
	_, err = db.Exec(ctx, `INSERT INTO staff_access(user_id,role,active) VALUES($1,'doctor',true)`, postWriteDocID)
	if err != nil {
		t.Fatalf("failed to insert staff_access for post-write rollback: %v", err)
	}

	// Create temporary trigger on audit_event to raise exception when inserting for postWriteDocID
	_, err = db.Exec(ctx, `
		CREATE OR REPLACE FUNCTION test_fail_audit_insert() RETURNS trigger AS $$
		BEGIN
			IF NEW.resource_id = '`+postWriteDocID+`' THEN
				RAISE EXCEPTION 'injected test failure on audit_event insert';
			END IF;
			RETURN NEW;
		END;
		$$ LANGUAGE plpgsql;
		DROP TRIGGER IF EXISTS trg_test_fail_audit_insert ON audit_event;
		CREATE TRIGGER trg_test_fail_audit_insert
		BEFORE INSERT ON audit_event
		FOR EACH ROW EXECUTE FUNCTION test_fail_audit_insert();
	`)
	if err != nil {
		t.Fatalf("failed to install test trigger: %v", err)
	}
	defer func() {
		_, _ = db.Exec(ctx, `DROP TRIGGER IF EXISTS trg_test_fail_audit_insert ON audit_event`)
		_, _ = db.Exec(ctx, `DROP FUNCTION IF EXISTS test_fail_audit_insert()`)
		_, _ = db.Exec(ctx, `DELETE FROM staff_access WHERE user_id=$1`, postWriteDocID)
		_, _ = db.Exec(ctx, `DELETE FROM "user" WHERE id=$1`, postWriteDocID)
	}()

	var beforeAllProfiles, beforeAllHours, beforeAllAudits int
	if err := db.QueryRow(ctx, `SELECT count(*) FROM doctor_profile`).Scan(&beforeAllProfiles); err != nil {
		t.Fatalf("failed to count doctor_profile before post-write failure: %v", err)
	}
	if err := db.QueryRow(ctx, `SELECT count(*) FROM doctor_hours`).Scan(&beforeAllHours); err != nil {
		t.Fatalf("failed to count doctor_hours before post-write failure: %v", err)
	}
	if err := db.QueryRow(ctx, `SELECT count(*) FROM audit_event`).Scan(&beforeAllAudits); err != nil {
		t.Fatalf("failed to count audit_event before post-write failure: %v", err)
	}

	postWriteIn := domain.CreateDoctorInput{
		UserID:        postWriteDocID,
		DepartmentID:  activeDeptID,
		Specialist:    "Neurology",
		Designation:   "Consultant",
		Qualification: "MBBS, MD",
		Gender:        "female",
	}
	_, err = sched.CreateDoctorProfile(ctx, admin, postWriteIn)
	if err == nil {
		t.Fatalf("expected error from injected audit failure, but CreateDoctorProfile succeeded")
	}

	// Verify 0 rows exist for postWriteDocID across doctor_profile, doctor_hours, audit_event
	var postProfileCount, postHoursCount, postAuditCount int
	if err := db.QueryRow(ctx, `SELECT count(*) FROM doctor_profile WHERE user_id=$1`, postWriteDocID).Scan(&postProfileCount); err != nil {
		t.Fatalf("failed to query doctor_profile for postWriteDocID: %v", err)
	}
	if err := db.QueryRow(ctx, `SELECT count(*) FROM doctor_hours WHERE doctor_id=$1`, postWriteDocID).Scan(&postHoursCount); err != nil {
		t.Fatalf("failed to query doctor_hours for postWriteDocID: %v", err)
	}
	if err := db.QueryRow(ctx, `SELECT count(*) FROM audit_event WHERE resource_id=$1`, postWriteDocID).Scan(&postAuditCount); err != nil {
		t.Fatalf("failed to query audit_event for postWriteDocID: %v", err)
	}
	if postProfileCount != 0 {
		t.Fatalf("expected 0 doctor_profile records after post-write rollback, got %d", postProfileCount)
	}
	if postHoursCount != 0 {
		t.Fatalf("expected 0 doctor_hours records after post-write rollback, got %d", postHoursCount)
	}
	if postAuditCount != 0 {
		t.Fatalf("expected 0 audit_event records after post-write rollback, got %d", postAuditCount)
	}

	// Compare total table row counts before and after failure
	var afterAllProfiles, afterAllHours, afterAllAudits int
	if err := db.QueryRow(ctx, `SELECT count(*) FROM doctor_profile`).Scan(&afterAllProfiles); err != nil {
		t.Fatalf("failed to count doctor_profile after post-write failure: %v", err)
	}
	if err := db.QueryRow(ctx, `SELECT count(*) FROM doctor_hours`).Scan(&afterAllHours); err != nil {
		t.Fatalf("failed to count doctor_hours after post-write failure: %v", err)
	}
	if err := db.QueryRow(ctx, `SELECT count(*) FROM audit_event`).Scan(&afterAllAudits); err != nil {
		t.Fatalf("failed to count audit_event after post-write failure: %v", err)
	}
	if afterAllProfiles != beforeAllProfiles {
		t.Fatalf("doctor_profile table count changed after post-write rollback: before=%d, after=%d", beforeAllProfiles, afterAllProfiles)
	}
	if afterAllHours != beforeAllHours {
		t.Fatalf("doctor_hours table count changed after post-write rollback: before=%d, after=%d", beforeAllHours, afterAllHours)
	}
	if afterAllAudits != beforeAllAudits {
		t.Fatalf("audit_event table count changed after post-write rollback: before=%d, after=%d", beforeAllAudits, afterAllAudits)
	}

	// -------------------------------------------------------------
	// 14. Doctor Deletion Protection (Step 2 of Parity Plan)
	// -------------------------------------------------------------
	seedTestDoctor := func(id, name string) {
		t.Helper()
		_, err := db.Exec(ctx, `INSERT INTO "user"(id,name,email) VALUES($1,$2,$1||'@hospital.test') ON CONFLICT DO NOTHING`, id, name)
		if err != nil {
			t.Fatalf("failed to insert user %s: %v", id, err)
		}
		_, err = db.Exec(ctx, `INSERT INTO staff_access(user_id,role,active) VALUES($1,'doctor',true) ON CONFLICT DO NOTHING`, id)
		if err != nil {
			t.Fatalf("failed to insert staff_access %s: %v", id, err)
		}
		_, err = db.Exec(ctx, `INSERT INTO doctor_profile(user_id,department,timezone,slot_minutes,version) VALUES($1,'General medicine','Africa/Addis_Ababa',60,1) ON CONFLICT DO NOTHING`, id)
		if err != nil {
			t.Fatalf("failed to insert doctor_profile %s: %v", id, err)
		}
		for wd := 0; wd < 7; wd++ {
			_, err = db.Exec(ctx, `INSERT INTO doctor_hours(doctor_id,weekday,start_minute,end_minute) VALUES($1,$2,600,1170) ON CONFLICT DO NOTHING`, id, wd)
			if err != nil {
				t.Fatalf("failed to insert doctor_hours %s wd %d: %v", id, wd, err)
			}
		}
	}

	deleteDoctorHTTP := func(cookie, docID string) *httptest.ResponseRecorder {
		t.Helper()
		req := httptest.NewRequest("DELETE", "/v1/doctors/"+docID, nil)
		if cookie != "" {
			req.Header.Set("Cookie", cookie)
		}
		req.Header.Set("Origin", "http://hospital.test")
		rec := httptest.NewRecorder()
		httpHandler.ServeHTTP(rec, req)
		return rec
	}

	var testPatID string
	err = db.QueryRow(ctx, `INSERT INTO patient(given_name,family_name,date_of_birth,phone) VALUES('Test','Deletions','1990-01-01','+251911999999') RETURNING id::text`).Scan(&testPatID)
	if err != nil {
		t.Fatalf("failed to seed test patient: %v", err)
	}

	// 14a. Unauthorized doctor deletion requests
	unauthDocID := "doc-unauth-test"
	seedTestDoctor(unauthDocID, "Dr. Unauthorized Target")

	// No session cookie -> 401
	rec := deleteDoctorHTTP("", unauthDocID)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for unauthenticated DELETE /v1/doctors, got %d: %s", rec.Code, rec.Body.String())
	}
	// Doctor session -> 403
	rec = deleteDoctorHTTP("session=doctor", unauthDocID)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for doctor session DELETE /v1/doctors, got %d: %s", rec.Code, rec.Body.String())
	}
	// Nurse session -> 403
	rec = deleteDoctorHTTP("session=nurse", unauthDocID)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for nurse session DELETE /v1/doctors, got %d: %s", rec.Code, rec.Body.String())
	}
	// Receptionist session -> 403
	rec = deleteDoctorHTTP("session=receptionist", unauthDocID)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for receptionist session DELETE /v1/doctors, got %d: %s", rec.Code, rec.Body.String())
	}
	// Patient session -> 403
	rec = deleteDoctorHTTP("session=patient", unauthDocID)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for patient session DELETE /v1/doctors, got %d: %s", rec.Code, rec.Body.String())
	}

	// 14b. Admin self-deletion prevention -> 409
	rec = deleteDoctorHTTP("session=admin", admin.ID)
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409 for admin self-deletion, got %d: %s", rec.Code, rec.Body.String())
	}

	// 14c. Non-existent doctor -> 404
	rec = deleteDoctorHTTP("session=admin", "non-existent-doctor-999")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for non-existent doctor deletion, got %d: %s", rec.Code, rec.Body.String())
	}

	// 14d. Dependency 1: PatientCase (patient_case.doctor_id)
	depDoc1 := "doc-dep-case"
	seedTestDoctor(depDoc1, "Dr. Case Dep")
	var case1ID string
	err = db.QueryRow(ctx, `INSERT INTO patient_case(patient_id, doctor_id, description, created_by) VALUES($1, $2, 'Case Dep Test', $3) RETURNING id::text`, testPatID, depDoc1, admin.ID).Scan(&case1ID)
	if err != nil {
		t.Fatalf("failed to insert patient_case fixture: %v", err)
	}
	rec = deleteDoctorHTTP("session=admin", depDoc1)
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409 for doctor with patient_case, got %d: %s", rec.Code, rec.Body.String())
	}
	var docCount int
	_ = db.QueryRow(ctx, `SELECT count(*) FROM doctor_profile WHERE user_id=$1`, depDoc1).Scan(&docCount)
	if docCount != 1 {
		t.Fatalf("expected doctor %s preserved after conflict, got count %d", depDoc1, docCount)
	}
	var caseCount int
	_ = db.QueryRow(ctx, `SELECT count(*) FROM patient_case WHERE id=$1`, case1ID).Scan(&caseCount)
	if caseCount != 1 {
		t.Fatalf("expected clinical history preserved for patient_case %s, got count %d", case1ID, caseCount)
	}

	// 14e. Dependency 2: Encounter (encounter.doctor_id)
	depDoc2 := "doc-dep-encounter"
	seedTestDoctor(depDoc2, "Dr. Encounter Dep")
	var case2ID string
	err = db.QueryRow(ctx, `INSERT INTO patient_case(patient_id, doctor_id, description, created_by) VALUES($1, $2, 'Case 2', $3) RETURNING id::text`, testPatID, depDoc2, admin.ID).Scan(&case2ID)
	if err != nil {
		t.Fatalf("failed to insert patient_case for encounter: %v", err)
	}
	var enc2ID string
	err = db.QueryRow(ctx, `
		INSERT INTO encounter(kind, case_id, patient_id, doctor_id, admitted_at, status, created_by, request_key)
		VALUES('opd', $1, $2, $3, now(), 'active', $4, 'req-key-enc-2')
		RETURNING id::text
	`, case2ID, testPatID, depDoc2, admin.ID).Scan(&enc2ID)
	if err != nil {
		t.Fatalf("failed to insert encounter fixture: %v", err)
	}
	rec = deleteDoctorHTTP("session=admin", depDoc2)
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409 for doctor with encounter, got %d: %s", rec.Code, rec.Body.String())
	}
	_ = db.QueryRow(ctx, `SELECT count(*) FROM doctor_profile WHERE user_id=$1`, depDoc2).Scan(&docCount)
	if docCount != 1 {
		t.Fatalf("expected doctor %s preserved after conflict, got count %d", depDoc2, docCount)
	}
	var encCount int
	_ = db.QueryRow(ctx, `SELECT count(*) FROM encounter WHERE id=$1`, enc2ID).Scan(&encCount)
	if encCount != 1 {
		t.Fatalf("expected encounter %s preserved, got count %d", enc2ID, encCount)
	}

	// 14f. Dependency 3: Appointment (appointment.doctor_id)
	depDoc3 := "doc-dep-appt"
	seedTestDoctor(depDoc3, "Dr. Appt Dep")
	var appt3ID string
	err = db.QueryRow(ctx, `
		INSERT INTO appointment(patient_id, doctor_id, starts_at, ends_at, status, problem, created_by, request_key)
		VALUES($1, $2, now() + interval '1 hour', now() + interval '2 hours', 'booked', 'Dep appt test', $3, 'req-key-appt-test-3')
		RETURNING id::text
	`, testPatID, depDoc3, admin.ID).Scan(&appt3ID)
	if err != nil {
		t.Fatalf("failed to insert appointment fixture: %v", err)
	}
	rec = deleteDoctorHTTP("session=admin", depDoc3)
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409 for doctor with appointment, got %d: %s", rec.Code, rec.Body.String())
	}
	_ = db.QueryRow(ctx, `SELECT count(*) FROM doctor_profile WHERE user_id=$1`, depDoc3).Scan(&docCount)
	if docCount != 1 {
		t.Fatalf("expected doctor %s preserved after conflict, got count %d", depDoc3, docCount)
	}
	var apptCount int
	_ = db.QueryRow(ctx, `SELECT count(*) FROM appointment WHERE id=$1`, appt3ID).Scan(&apptCount)
	if apptCount != 1 {
		t.Fatalf("expected appointment %s preserved, got count %d", appt3ID, apptCount)
	}

	// 14g. Dependency 4: BirthReport (birth_report.delivered_by)
	depDoc4 := "doc-dep-birth"
	seedTestDoctor(depDoc4, "Dr. Birth Dep")
	var birth4ID string
	err = db.QueryRow(ctx, `
		INSERT INTO birth_report(report_number, child_name, gender, birth_date, weight_kg, mother_name, delivered_by, created_by)
		VALUES('BR-TEST-004', 'Baby Dep', 'female', now(), 3.10, 'Mother Dep', $1, $2)
		RETURNING id::text
	`, depDoc4, admin.ID).Scan(&birth4ID)
	if err != nil {
		t.Fatalf("failed to insert birth_report fixture: %v", err)
	}
	rec = deleteDoctorHTTP("session=admin", depDoc4)
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409 for doctor with birth_report, got %d: %s", rec.Code, rec.Body.String())
	}
	_ = db.QueryRow(ctx, `SELECT count(*) FROM doctor_profile WHERE user_id=$1`, depDoc4).Scan(&docCount)
	if docCount != 1 {
		t.Fatalf("expected doctor %s preserved after conflict, got count %d", depDoc4, docCount)
	}
	var birthCount int
	_ = db.QueryRow(ctx, `SELECT count(*) FROM birth_report WHERE id=$1`, birth4ID).Scan(&birthCount)
	if birthCount != 1 {
		t.Fatalf("expected birth_report %s preserved, got count %d", birth4ID, birthCount)
	}

	// 14h. Dependency 5: DeathReport (death_report.certified_by)
	depDoc5 := "doc-dep-death"
	seedTestDoctor(depDoc5, "Dr. Death Dep")
	var death5ID string
	err = db.QueryRow(ctx, `
		INSERT INTO death_report(report_number, patient_id, death_date, cause_of_death, certified_by, created_by)
		VALUES('DR-TEST-005', $1, now(), 'Cardiac Arrest', $2, $3)
		RETURNING id::text
	`, testPatID, depDoc5, admin.ID).Scan(&death5ID)
	if err != nil {
		t.Fatalf("failed to insert death_report fixture: %v", err)
	}
	rec = deleteDoctorHTTP("session=admin", depDoc5)
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409 for doctor with death_report, got %d: %s", rec.Code, rec.Body.String())
	}
	_ = db.QueryRow(ctx, `SELECT count(*) FROM doctor_profile WHERE user_id=$1`, depDoc5).Scan(&docCount)
	if docCount != 1 {
		t.Fatalf("expected doctor %s preserved after conflict, got count %d", depDoc5, docCount)
	}
	var deathCount int
	_ = db.QueryRow(ctx, `SELECT count(*) FROM death_report WHERE id=$1`, death5ID).Scan(&deathCount)
	if deathCount != 1 {
		t.Fatalf("expected death_report %s preserved, got count %d", death5ID, deathCount)
	}

	// 14i. Dependency 6: InvestigationReport (investigation_report.investigated_by)
	depDoc6 := "doc-dep-inv"
	seedTestDoctor(depDoc6, "Dr. Inv Dep")
	var inv6ID string
	err = db.QueryRow(ctx, `
		INSERT INTO investigation_report(report_number, patient_id, title, investigation_type, investigated_by)
		VALUES('IR-TEST-006', $1, 'Complete Blood Count', 'pathology', $2)
		RETURNING id::text
	`, testPatID, depDoc6).Scan(&inv6ID)
	if err != nil {
		t.Fatalf("failed to insert investigation_report fixture: %v", err)
	}
	rec = deleteDoctorHTTP("session=admin", depDoc6)
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409 for doctor with investigation_report, got %d: %s", rec.Code, rec.Body.String())
	}
	_ = db.QueryRow(ctx, `SELECT count(*) FROM doctor_profile WHERE user_id=$1`, depDoc6).Scan(&docCount)
	if docCount != 1 {
		t.Fatalf("expected doctor %s preserved after conflict, got count %d", depDoc6, docCount)
	}
	var invCount int
	_ = db.QueryRow(ctx, `SELECT count(*) FROM investigation_report WHERE id=$1`, inv6ID).Scan(&invCount)
	if invCount != 1 {
		t.Fatalf("expected investigation_report %s preserved, got count %d", inv6ID, invCount)
	}

	// 14j. Dependency 7: OperationReport (operation_report.surgeon_id)
	depDoc7 := "doc-dep-op"
	seedTestDoctor(depDoc7, "Dr. Surgeon Dep")
	// Use enc2ID (where encounter.doctor_id is depDoc2, not depDoc7) to isolate operation_report
	var op7ID string
	err = db.QueryRow(ctx, `
		INSERT INTO operation_report(report_number, encounter_id, patient_id, operation_name, surgeon_id, operation_date, created_by)
		VALUES('OR-TEST-007', $1, $2, 'Appendectomy', $3, now(), $4)
		RETURNING id::text
	`, enc2ID, testPatID, depDoc7, admin.ID).Scan(&op7ID)
	if err != nil {
		t.Fatalf("failed to insert operation_report fixture: %v", err)
	}
	rec = deleteDoctorHTTP("session=admin", depDoc7)
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409 for doctor with operation_report, got %d: %s", rec.Code, rec.Body.String())
	}
	_ = db.QueryRow(ctx, `SELECT count(*) FROM doctor_profile WHERE user_id=$1`, depDoc7).Scan(&docCount)
	if docCount != 1 {
		t.Fatalf("expected doctor %s preserved after conflict, got count %d", depDoc7, docCount)
	}
	var opCount int
	_ = db.QueryRow(ctx, `SELECT count(*) FROM operation_report WHERE id=$1`, op7ID).Scan(&opCount)
	if opCount != 1 {
		t.Fatalf("expected operation_report %s preserved, got count %d", op7ID, opCount)
	}

	// 14k. Dependency 8: Prescription (prescription.doctor_id)
	depDoc8 := "doc-dep-rx"
	seedTestDoctor(depDoc8, "Dr. Rx Dep")
	var rx8ID string
	err = db.QueryRow(ctx, `
		INSERT INTO prescription(patient_id, doctor_id)
		VALUES($1, $2)
		RETURNING id::text
	`, testPatID, depDoc8).Scan(&rx8ID)
	if err != nil {
		t.Fatalf("failed to insert prescription fixture: %v", err)
	}
	rec = deleteDoctorHTTP("session=admin", depDoc8)
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409 for doctor with prescription, got %d: %s", rec.Code, rec.Body.String())
	}
	_ = db.QueryRow(ctx, `SELECT count(*) FROM doctor_profile WHERE user_id=$1`, depDoc8).Scan(&docCount)
	if docCount != 1 {
		t.Fatalf("expected doctor %s preserved after conflict, got count %d", depDoc8, docCount)
	}
	var rxCount int
	_ = db.QueryRow(ctx, `SELECT count(*) FROM prescription WHERE id=$1`, rx8ID).Scan(&rxCount)
	if rxCount != 1 {
		t.Fatalf("expected prescription %s preserved, got count %d", rx8ID, rxCount)
	}

	// 14l. Dependency 9: IpdPatientDepartment (ipd_admission_details via encounter.doctor_id)
	depDoc9 := "doc-dep-ipd"
	seedTestDoctor(depDoc9, "Dr. IPD Dep")
	var case9ID, bedType9ID, bed9ID, enc9ID string
	err = db.QueryRow(ctx, `INSERT INTO patient_case(patient_id, doctor_id, description, created_by) VALUES($1, $2, 'Case 9', $3) RETURNING id::text`, testPatID, depDoc9, admin.ID).Scan(&case9ID)
	if err != nil {
		t.Fatalf("failed to insert patient_case for ipd: %v", err)
	}
	err = db.QueryRow(ctx, `INSERT INTO bed_type(name) VALUES('Dep Bed Type 9') ON CONFLICT (name) DO UPDATE SET active=true RETURNING id::text`).Scan(&bedType9ID)
	if err != nil {
		t.Fatalf("failed to insert bed_type: %v", err)
	}
	err = db.QueryRow(ctx, `INSERT INTO hospital_bed(name, bed_type, type_id, charge_minor, active, created_by) VALUES('IPD Bed 9', 'Dep Bed Type 9', $1, 5000, true, $2) RETURNING id::text`, bedType9ID, admin.ID).Scan(&bed9ID)
	if err != nil {
		t.Fatalf("failed to insert hospital_bed for ipd: %v", err)
	}
	err = db.QueryRow(ctx, `
		INSERT INTO encounter(kind, case_id, patient_id, doctor_id, bed_id, admitted_at, status, created_by, request_key)
		VALUES('ipd', $1, $2, $3, $4, now(), 'active', $5, 'req-key-enc-test-9')
		RETURNING id::text
	`, case9ID, testPatID, depDoc9, bed9ID, admin.ID).Scan(&enc9ID)
	if err != nil {
		t.Fatalf("failed to insert ipd encounter: %v", err)
	}
	_, err = db.Exec(ctx, `
		INSERT INTO ipd_admission_details(encounter_id, package_name, package_charge_minor, guardian_name)
		VALUES($1, 'Surgical IPD Package', 75000, 'Guardian Test')
	`, enc9ID)
	if err != nil {
		t.Fatalf("failed to insert ipd_admission_details: %v", err)
	}
	rec = deleteDoctorHTTP("session=admin", depDoc9)
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409 for doctor with ipd_admission_details, got %d: %s", rec.Code, rec.Body.String())
	}
	_ = db.QueryRow(ctx, `SELECT count(*) FROM doctor_profile WHERE user_id=$1`, depDoc9).Scan(&docCount)
	if docCount != 1 {
		t.Fatalf("expected doctor %s preserved after conflict, got count %d", depDoc9, docCount)
	}
	var ipdDetailCount int
	_ = db.QueryRow(ctx, `SELECT count(*) FROM ipd_admission_details WHERE encounter_id=$1`, enc9ID).Scan(&ipdDetailCount)
	if ipdDetailCount != 1 {
		t.Fatalf("expected ipd_admission_details preserved, got count %d", ipdDetailCount)
	}

	// 14m. Dependency 10: EmployeePayroll (employee_payroll.user_id)
	depDoc10 := "doc-dep-payroll"
	seedTestDoctor(depDoc10, "Dr. Payroll Dep")
	var payroll10ID string
	err = db.QueryRow(ctx, `
		INSERT INTO employee_payroll(payroll_number, user_id, role, month, year, basic_salary_minor, net_salary_minor, created_by)
		VALUES('PAY-TEST-010', $1, 'doctor', 'October', 2026, 6000000, 5500000, $2)
		RETURNING id::text
	`, depDoc10, admin.ID).Scan(&payroll10ID)
	if err != nil {
		t.Fatalf("failed to insert employee_payroll fixture: %v", err)
	}
	rec = deleteDoctorHTTP("session=admin", depDoc10)
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409 for doctor with employee_payroll, got %d: %s", rec.Code, rec.Body.String())
	}
	_ = db.QueryRow(ctx, `SELECT count(*) FROM doctor_profile WHERE user_id=$1`, depDoc10).Scan(&docCount)
	if docCount != 1 {
		t.Fatalf("expected doctor %s preserved after conflict, got count %d", depDoc10, docCount)
	}
	var payrollCount int
	_ = db.QueryRow(ctx, `SELECT count(*) FROM employee_payroll WHERE id=$1`, payroll10ID).Scan(&payrollCount)
	if payrollCount != 1 {
		t.Fatalf("expected employee_payroll preserved, got count %d", payrollCount)
	}

	// 14n. Successful unreferenced doctor deletion
	cleanDocID := "doc-unreferenced-clean"
	seedTestDoctor(cleanDocID, "Dr. Clean Unreferenced")
	_, err = db.Exec(ctx, `INSERT INTO session(id, "expiresAt", token, "userId") VALUES('clean-sess-1', now() + interval '1 hour', 'clean-tok-1', $1)`, cleanDocID)
	if err != nil {
		t.Fatalf("failed to seed session for clean doc: %v", err)
	}
	_, err = db.Exec(ctx, `INSERT INTO account(id, "accountId", "providerId", "userId") VALUES('clean-acc-1', $1, 'credential', $1)`, cleanDocID)
	if err != nil {
		t.Fatalf("failed to seed account for clean doc: %v", err)
	}

	rec = deleteDoctorHTTP("session=admin", cleanDocID)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for clean unreferenced doctor deletion, got %d: %s", rec.Code, rec.Body.String())
	}
	var delResp struct {
		Deleted bool   `json:"deleted"`
		ID      string `json:"id"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &delResp); err != nil {
		t.Fatalf("failed to unmarshal deletion response: %v", err)
	}
	if !delResp.Deleted || delResp.ID != cleanDocID {
		t.Fatalf("unexpected deletion response: %+v", delResp)
	}

	// Verify all records deleted
	for table, query := range map[string]string{
		"user":           `SELECT count(*) FROM "user" WHERE id=$1`,
		"staff_access":   `SELECT count(*) FROM staff_access WHERE user_id=$1`,
		"staff_profile":  `SELECT count(*) FROM staff_profile WHERE user_id=$1`,
		"doctor_profile": `SELECT count(*) FROM doctor_profile WHERE user_id=$1`,
		"doctor_hours":   `SELECT count(*) FROM doctor_hours WHERE doctor_id=$1`,
		"session":        `SELECT count(*) FROM session WHERE "userId"=$1`,
		"account":        `SELECT count(*) FROM account WHERE "userId"=$1`,
	} {
		var cnt int
		if err := db.QueryRow(ctx, query, cleanDocID).Scan(&cnt); err != nil {
			t.Fatalf("failed to query count for %s: %v", table, err)
		}
		if cnt != 0 {
			t.Fatalf("expected 0 rows in %s for deleted doctor %s, got %d", table, cleanDocID, cnt)
		}
	}

	// Verify audit_event logged doctor.deleted
	var delAuditCount int
	err = db.QueryRow(ctx, `SELECT count(*) FROM audit_event WHERE action='doctor.deleted' AND resource_id=$1 AND actor_id=$2`, cleanDocID, admin.ID).Scan(&delAuditCount)
	if err != nil || delAuditCount != 1 {
		t.Fatalf("expected 1 audit_event for doctor.deleted, got %d (err: %v)", delAuditCount, err)
	}

	// 14o. Doctor deletion transaction rollback test
	rollbackDocID := "doc-rollback-test"
	seedTestDoctor(rollbackDocID, "Dr. Rollback Test")
	_, err = db.Exec(ctx, `INSERT INTO session(id, "expiresAt", token, "userId") VALUES('rb-sess-1', now() + interval '1 hour', 'rb-tok-1', $1)`, rollbackDocID)
	if err != nil {
		t.Fatalf("failed to seed session for rollback doc: %v", err)
	}
	_, err = db.Exec(ctx, `INSERT INTO account(id, "accountId", "providerId", "userId") VALUES('rb-acc-1', $1, 'credential', $1)`, rollbackDocID)
	if err != nil {
		t.Fatalf("failed to seed account for rollback doc: %v", err)
	}

	// Install trigger on "user" before delete that fails for rollbackDocID
	_, err = db.Exec(ctx, `
		CREATE OR REPLACE FUNCTION test_fail_user_delete() RETURNS trigger AS $$
		BEGIN
			IF OLD.id = '`+rollbackDocID+`' THEN
				RAISE EXCEPTION 'injected test failure on user deletion';
			END IF;
			RETURN OLD;
		END;
		$$ LANGUAGE plpgsql;
		DROP TRIGGER IF EXISTS trg_test_fail_user_delete ON "user";
		CREATE TRIGGER trg_test_fail_user_delete
		BEFORE DELETE ON "user"
		FOR EACH ROW EXECUTE FUNCTION test_fail_user_delete();
	`)
	if err != nil {
		t.Fatalf("failed to install test rollback trigger: %v", err)
	}

	rec = deleteDoctorHTTP("session=admin", rollbackDocID)
	if rec.Code == http.StatusOK {
		t.Fatalf("expected failure from injected trigger, but got 200 OK")
	}

	// Verify all records survived intact
	for table, query := range map[string]string{
		"user":           `SELECT count(*) FROM "user" WHERE id=$1`,
		"staff_access":   `SELECT count(*) FROM staff_access WHERE user_id=$1`,
		"doctor_profile": `SELECT count(*) FROM doctor_profile WHERE user_id=$1`,
		"session":        `SELECT count(*) FROM session WHERE "userId"=$1`,
		"account":        `SELECT count(*) FROM account WHERE "userId"=$1`,
	} {
		var cnt int
		if err := db.QueryRow(ctx, query, rollbackDocID).Scan(&cnt); err != nil {
			t.Fatalf("failed to query count for %s after rollback: %v", table, err)
		}
		if cnt != 1 {
			t.Fatalf("expected 1 row in %s after rollback for %s, got %d", table, rollbackDocID, cnt)
		}
	}
	var hoursCnt int
	if err := db.QueryRow(ctx, `SELECT count(*) FROM doctor_hours WHERE doctor_id=$1`, rollbackDocID).Scan(&hoursCnt); err != nil || hoursCnt != 7 {
		t.Fatalf("expected 7 doctor_hours after rollback, got %d (err: %v)", hoursCnt, err)
	}
	var auditCnt int
	if err := db.QueryRow(ctx, `SELECT count(*) FROM audit_event WHERE action='doctor.deleted' AND resource_id=$1`, rollbackDocID).Scan(&auditCnt); err != nil || auditCnt != 0 {
		t.Fatalf("expected 0 doctor.deleted audit events after rollback, got %d (err: %v)", auditCnt, err)
	}

	// Drop trigger and retry deletion successfully
	_, _ = db.Exec(ctx, `DROP TRIGGER IF EXISTS trg_test_fail_user_delete ON "user"`)
	_, _ = db.Exec(ctx, `DROP FUNCTION IF EXISTS test_fail_user_delete()`)

	rec = deleteDoctorHTTP("session=admin", rollbackDocID)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on retry after removing trigger, got %d: %s", rec.Code, rec.Body.String())
	}
	var finalUserCnt int
	_ = db.QueryRow(ctx, `SELECT count(*) FROM "user" WHERE id=$1`, rollbackDocID).Scan(&finalUserCnt)
	if finalUserCnt != 0 {
		t.Fatalf("expected user %s removed on successful retry, got count %d", rollbackDocID, finalUserCnt)
	}

	// -------------------------------------------------------------
	// Test K: Doctor Schedules, Holidays, and Breaks Parity (Step 3)
	// -------------------------------------------------------------
	docSchedUser := "doctor-sched-test"
	_, _ = db.Exec(ctx, `INSERT INTO "user"(id,name,email) VALUES($1,'Dr. Schedule Test',$1||'@hospital.test') ON CONFLICT DO NOTHING`, docSchedUser)
	_, _ = db.Exec(ctx, `INSERT INTO staff_access(user_id,role,active) VALUES($1,'doctor',true) ON CONFLICT DO NOTHING`, docSchedUser)

	_, err = sched.CreateDoctorProfile(ctx, admin, domain.CreateDoctorInput{
		UserID:            docSchedUser,
		DepartmentID:      activeDeptID,
		Specialist:        "Scheduling Specialist",
		Designation:       "Associate Professor",
		Qualification:     "MD",
		Gender:            "male",
		SlotMinutes:       60,
		AppointmentCharge: 200,
		OpdCharge:         150,
	})
	if err != nil {
		t.Fatalf("failed to create docSchedUser profile: %v", err)
	}

	docSchedActor := domain.Actor{ID: docSchedUser, Role: "doctor"}

	// 1. Fetch default schedule
	schedulesList, err := sched.DoctorSchedules(ctx, admin, docSchedUser)
	if err != nil || len(schedulesList) != 1 {
		t.Fatalf("expected 1 schedule for doctor, got %d (err: %v)", len(schedulesList), err)
	}
	if len(schedulesList[0].Days) != 7 {
		t.Fatalf("expected default 7 days in schedule, got %d", len(schedulesList[0].Days))
	}

	// Doctor viewing own schedule allowed; doctor viewing another doctor's schedule forbidden
	_, err = sched.DoctorSchedules(ctx, docSchedActor, docSchedUser)
	if err != nil {
		t.Fatalf("doctor viewing own schedule should succeed, got %v", err)
	}
	_, err = sched.DoctorSchedule(ctx, docSchedActor, docUser1)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("doctor viewing other doctor's schedule must be forbidden, got %v", err)
	}

	// 2. Save custom schedule
	newScheduleDays := []domain.ScheduleDayRow{
		{Day: "Monday", From: "09:00:00", To: "17:00:00"},
		{Day: "Wednesday", From: "09:00:00", To: "17:00:00"},
		{Day: "Friday", From: "09:00:00", To: "17:00:00"},
	}
	savedSched, err := sched.SaveDoctorSchedule(ctx, admin, domain.SaveDoctorScheduleInput{
		DoctorID:    docSchedUser,
		SlotMinutes: 30,
		Days:        newScheduleDays,
	})
	if err != nil {
		t.Fatalf("failed to save custom doctor schedule: %v", err)
	}
	if savedSched.SlotMinutes != 30 || len(savedSched.Days) != 3 {
		t.Fatalf("saved schedule mismatch: slotMinutes=%d days=%d", savedSched.SlotMinutes, len(savedSched.Days))
	}

	// 3. Schedule change conflict protection with future active appointment
	// Find the next upcoming Wednesday at 10:00
	nowTime := time.Now().In(domain.HospitalLocation)
	daysUntilWed := (int(time.Wednesday) - int(nowTime.Weekday()) + 7) % 7
	if daysUntilWed == 0 {
		daysUntilWed = 7
	}
	targetWed := nowTime.AddDate(0, 0, daysUntilWed)
	apptStartTime := time.Date(targetWed.Year(), targetWed.Month(), targetWed.Day(), 10, 0, 0, 0, domain.HospitalLocation).UTC()

	// Seed patient for appointment
	schedPatUser := "sched-pat-test"
	_, _ = db.Exec(ctx, `INSERT INTO "user"(id,name,email) VALUES($1,'Sched Pat',$1||'@patient.test') ON CONFLICT DO NOTHING`, schedPatUser)
	var schedPatID string
	_ = db.QueryRow(ctx, `INSERT INTO patient(given_name,family_name,date_of_birth,phone) VALUES('Sched','Pat','1995-05-05','+251911999999') RETURNING id`).Scan(&schedPatID)

	bookedAppt, err := sched.Book(ctx, admin, domain.AppointmentInput{
		PatientID: schedPatID,
		DoctorID:  docSchedUser,
		StartsAt:  apptStartTime,
		Problem:   "Routine Checkup",
	}, "test-sched-book-key-01")
	if err != nil {
		t.Fatalf("failed to book appointment on Wednesday: %v", err)
	}

	// Now attempt to change doctor schedule to ONLY Tuesday & Thursday (dropping Wednesday)
	// Must fail with ErrConflict because active appointment on Wednesday falls outside new schedule!
	_, err = sched.SaveDoctorSchedule(ctx, admin, domain.SaveDoctorScheduleInput{
		DoctorID:    docSchedUser,
		SlotMinutes: 30,
		Days: []domain.ScheduleDayRow{
			{Day: "Tuesday", From: "09:00:00", To: "17:00:00"},
			{Day: "Thursday", From: "09:00:00", To: "17:00:00"},
		},
	})
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("expected ErrConflict when changing schedule drops active appointment day, got %v", err)
	}

	// DeleteDoctorSchedule must fail with ErrInUse because doctor has active appointments
	err = sched.DeleteDoctorSchedule(ctx, admin, docSchedUser)
	if !errors.Is(err, domain.ErrInUse) {
		t.Fatalf("expected ErrInUse when deleting schedule of doctor with appointments, got %v", err)
	}

	// 4. Doctor Holidays: creation, duplicates, active appointment conflict, and slot exclusion
	holidayDateStr := targetWed.AddDate(0, 0, 7).Format("2006-01-02") // Next Wednesday
	createdHol, err := sched.CreateDoctorHoliday(ctx, admin, domain.CreateDoctorHolidayInput{
		DoctorID: docSchedUser,
		Date:     holidayDateStr,
		Reason:   "Medical Conference",
	})
	if err != nil {
		t.Fatalf("failed to create doctor holiday: %v", err)
	}
	if createdHol.DoctorID != docSchedUser || createdHol.Date != holidayDateStr {
		t.Fatalf("unexpected holiday output: %+v", createdHol)
	}

	// Duplicate holiday on same date must be rejected
	_, err = sched.CreateDoctorHoliday(ctx, admin, domain.CreateDoctorHolidayInput{
		DoctorID: docSchedUser,
		Date:     holidayDateStr,
		Reason:   "Duplicate attempt",
	})
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("expected ErrConflict for duplicate holiday, got %v", err)
	}

	// Attempt to create holiday on date of booked appointment (targetWed) must be rejected
	wedDateStr := targetWed.Format("2006-01-02")
	_, err = sched.CreateDoctorHoliday(ctx, admin, domain.CreateDoctorHolidayInput{
		DoctorID: docSchedUser,
		Date:     wedDateStr,
		Reason:   "Conflict attempt",
	})
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("expected ErrConflict when creating holiday on date with active appointment, got %v", err)
	}

	// Slot generation on holiday date must return 0 slots
	holSlots, err := sched.Slots(ctx, admin, docSchedUser, holidayDateStr)
	if err != nil {
		t.Fatalf("failed to fetch slots on holiday date: %v", err)
	}
	if len(holSlots) != 0 {
		t.Fatalf("expected 0 slots on doctor holiday date, got %d", len(holSlots))
	}

	// Booking on holiday date must be rejected
	holDayTime := time.Date(targetWed.AddDate(0, 0, 7).Year(), targetWed.AddDate(0, 0, 7).Month(), targetWed.AddDate(0, 0, 7).Day(), 10, 0, 0, 0, domain.HospitalLocation).UTC()
	_, err = sched.Book(ctx, admin, domain.AppointmentInput{
		PatientID: schedPatID,
		DoctorID:  docSchedUser,
		StartsAt:  holDayTime,
		Problem:   "Book on holiday",
	}, "test-book-on-holiday-key")
	if !errors.Is(err, domain.ErrConflict) && !errors.Is(err, domain.ErrStale) {
		t.Fatalf("expected conflict booking on holiday date, got %v", err)
	}

	// Rescheduling to holiday date must be rejected
	_, err = sched.Reschedule(ctx, admin, bookedAppt.ID, domain.RescheduleInput{
		StartsAt: holDayTime,
		Reason:   "Reschedule to holiday",
		Version:  bookedAppt.Version,
	})
	if !errors.Is(err, domain.ErrConflict) && !errors.Is(err, domain.ErrStale) {
		t.Fatalf("expected conflict rescheduling to holiday date, got %v", err)
	}

	// Delete doctor holiday
	err = sched.DeleteDoctorHoliday(ctx, admin, createdHol.ID)
	if err != nil {
		t.Fatalf("failed to delete doctor holiday: %v", err)
	}

	// Slots after deleting holiday should now be available
	afterHolSlots, err := sched.Slots(ctx, admin, docSchedUser, holidayDateStr)
	if err != nil {
		t.Fatalf("failed to fetch slots after holiday deletion: %v", err)
	}
	if len(afterHolSlots) == 0 {
		t.Fatalf("expected slots available after holiday deletion, got 0")
	}

	// 5. Lunch Breaks: creation, duplicate, appointment overlap conflict, slot exclusion
	createdBreak, err := sched.CreateDoctorBreak(ctx, admin, domain.CreateDoctorBreakInput{
		DoctorID:  docSchedUser,
		BreakFrom: "12:00:00",
		BreakTo:   "13:00:00",
		EveryDay:  true,
	})
	if err != nil {
		t.Fatalf("failed to create lunch break: %v", err)
	}
	if !createdBreak.EveryDay || createdBreak.BreakFrom != "12:00:00" {
		t.Fatalf("unexpected break output: %+v", createdBreak)
	}

	// Duplicate break must be rejected
	_, err = sched.CreateDoctorBreak(ctx, admin, domain.CreateDoctorBreakInput{
		DoctorID:  docSchedUser,
		BreakFrom: "12:00:00",
		BreakTo:   "13:00:00",
		EveryDay:  true,
	})
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("expected ErrConflict for duplicate break, got %v", err)
	}

	// Slots must not include lunch break hours (12:00 - 13:00)
	checkSlots, err := sched.Slots(ctx, admin, docSchedUser, holidayDateStr)
	if err != nil {
		t.Fatalf("failed to query slots: %v", err)
	}
	for _, sl := range checkSlots {
		eatTime := sl.In(domain.HospitalLocation)
		if eatTime.Hour() == 12 {
			t.Fatalf("slot %v generated during lunch break (12:00-13:00)!", eatTime)
		}
	}

	// Attempting to book during lunch break must be rejected
	lunchSlotTime := time.Date(targetWed.AddDate(0, 0, 7).Year(), targetWed.AddDate(0, 0, 7).Month(), targetWed.AddDate(0, 0, 7).Day(), 12, 0, 0, 0, domain.HospitalLocation).UTC()
	_, err = sched.Book(ctx, admin, domain.AppointmentInput{
		PatientID: schedPatID,
		DoctorID:  docSchedUser,
		StartsAt:  lunchSlotTime,
		Problem:   "Book during lunch break",
	}, "test-book-during-break-key")
	if !errors.Is(err, domain.ErrConflict) && !errors.Is(err, domain.ErrStale) {
		t.Fatalf("expected conflict booking during lunch break, got %v", err)
	}

	// Attempting to create lunch break overlapping an active appointment (at 10:00) must be rejected
	_, err = sched.CreateDoctorBreak(ctx, admin, domain.CreateDoctorBreakInput{
		DoctorID:  docSchedUser,
		BreakFrom: "09:30:00",
		BreakTo:   "10:30:00",
		EveryDay:  true,
	})
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("expected ErrConflict when break overlaps active appointment, got %v", err)
	}

	// Delete lunch break
	err = sched.DeleteDoctorBreak(ctx, admin, createdBreak.ID)
	if err != nil {
		t.Fatalf("failed to delete lunch break: %v", err)
	}

	// 6. Test HTTP endpoints for schedules, holidays, and breaks
	serverHandler := httpHandler

	callHTTP := func(method, path, sessionCookie string, body []byte) *httptest.ResponseRecorder {
		var req *http.Request
		if len(body) > 0 {
			req = httptest.NewRequest(method, path, bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
		} else {
			req = httptest.NewRequest(method, path, nil)
		}
		req.Header.Set("Origin", "http://hospital.test")
		req.Header.Set("Cookie", sessionCookie)
		w := httptest.NewRecorder()
		serverHandler.ServeHTTP(w, req)
		return w
	}

	// Test GET /v1/doctor-schedules
	respRec := callHTTP("GET", "/v1/doctor-schedules?doctorId="+docSchedUser, "session=admin", nil)
	if respRec.Code != http.StatusOK {
		t.Fatalf("GET /v1/doctor-schedules failed: %d %s", respRec.Code, respRec.Body.String())
	}

	// Test POST /v1/doctor-holidays
	holBody, _ := json.Marshal(map[string]any{
		"doctorId": docSchedUser,
		"date":     "2026-12-25",
		"reason":   "Christmas Holiday",
	})
	respRec = callHTTP("POST", "/v1/doctor-holidays", "session=admin", holBody)
	if respRec.Code != http.StatusCreated {
		t.Fatalf("POST /v1/doctor-holidays failed: %d %s", respRec.Code, respRec.Body.String())
	}
	var createdHolHTTP domain.DoctorHoliday
	_ = json.Unmarshal(respRec.Body.Bytes(), &createdHolHTTP)

	// Test DELETE /v1/doctor-holidays/{id}
	respRec = callHTTP("DELETE", "/v1/doctor-holidays/"+createdHolHTTP.ID, "session=admin", nil)
	if respRec.Code != http.StatusOK {
		t.Fatalf("DELETE /v1/doctor-holidays failed: %d %s", respRec.Code, respRec.Body.String())
	}

	// Test POST /v1/doctor-breaks
	breakBody, _ := json.Marshal(map[string]any{
		"doctorId":  docSchedUser,
		"breakFrom": "13:00:00",
		"breakTo":   "14:00:00",
		"everyDay":  true,
	})
	respRec = callHTTP("POST", "/v1/doctor-breaks", "session=admin", breakBody)
	if respRec.Code != http.StatusCreated {
		t.Fatalf("POST /v1/doctor-breaks failed: %d %s", respRec.Code, respRec.Body.String())
	}
	var createdBreakHTTP domain.DoctorLunchBreak
	_ = json.Unmarshal(respRec.Body.Bytes(), &createdBreakHTTP)

	// Test DELETE /v1/doctor-breaks/{id}
	respRec = callHTTP("DELETE", "/v1/doctor-breaks/"+createdBreakHTTP.ID, "session=admin", nil)
	if respRec.Code != http.StatusOK {
		t.Fatalf("DELETE /v1/doctor-breaks failed: %d %s", respRec.Code, respRec.Body.String())
	}

	// -------------------------------------------------------------
	// Test L: Doctor OPD Charges Master & Flow Connection (Step 4)
	// -------------------------------------------------------------
	// 1. Save doctor OPD charge
	savedCharge, err := sched.SaveDoctorOPDCharge(ctx, admin, domain.SaveDoctorOPDChargeInput{
		DoctorID:       docSchedUser,
		StandardCharge: 350.00,
		CurrencySymbol: "ETB",
	})
	if err != nil {
		t.Fatalf("failed to save doctor OPD charge: %v", err)
	}
	if savedCharge.StandardCharge != 350.00 || savedCharge.CurrencySymbol != "ETB" {
		t.Fatalf("unexpected saved OPD charge: %+v", savedCharge)
	}

	// Verify database record in doctor_opd_charge
	var dbStdCharge float64
	var dbCurr string
	err = db.QueryRow(ctx, `SELECT standard_charge, currency_symbol FROM doctor_opd_charge WHERE doctor_id=$1`, docSchedUser).Scan(&dbStdCharge, &dbCurr)
	if err != nil || dbStdCharge != 350.00 || dbCurr != "ETB" {
		t.Fatalf("database verification failed for doctor_opd_charge: charge=%v curr=%s err=%v", dbStdCharge, dbCurr, err)
	}

	// Verify synchronization with doctor_profile.opd_charge
	var profileOpdCharge float64
	err = db.QueryRow(ctx, `SELECT opd_charge FROM doctor_profile WHERE user_id=$1`, docSchedUser).Scan(&profileOpdCharge)
	if err != nil || profileOpdCharge != 350.00 {
		t.Fatalf("doctor_profile.opd_charge not synchronized: expected 350.00 got %v (err: %v)", profileOpdCharge, err)
	}

	// 2. Fetch single OPD charge
	fetchedCharge, err := sched.DoctorOPDCharge(ctx, admin, docSchedUser)
	if err != nil || fetchedCharge.StandardCharge != 350.00 {
		t.Fatalf("failed to fetch doctor OPD charge: %+v (err: %v)", fetchedCharge, err)
	}

	// Doctor can view own OPD charge
	_, err = sched.DoctorOPDCharge(ctx, docSchedActor, docSchedUser)
	if err != nil {
		t.Fatalf("doctor should be allowed to view own OPD charge, got %v", err)
	}

	// 3. List OPD charges with search
	chargesList, err := sched.DoctorOPDCharges(ctx, admin, "Schedule")
	if err != nil || len(chargesList) == 0 {
		t.Fatalf("expected doctor in OPD charges list search, got %d (err: %v)", len(chargesList), err)
	}

	// 4. Role restrictions: non-admin cannot save or delete OPD charge
	_, err = sched.SaveDoctorOPDCharge(ctx, docSchedActor, domain.SaveDoctorOPDChargeInput{
		DoctorID:       docSchedUser,
		StandardCharge: 500.00,
	})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("non-admin save OPD charge must be forbidden, got %v", err)
	}
	err = sched.DeleteDoctorOPDCharge(ctx, docSchedActor, docSchedUser)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("non-admin delete OPD charge must be forbidden, got %v", err)
	}

	// 5. Delete OPD charge
	err = sched.DeleteDoctorOPDCharge(ctx, admin, docSchedUser)
	if err != nil {
		t.Fatalf("failed to delete OPD charge: %v", err)
	}
	var postDelCnt int
	_ = db.QueryRow(ctx, `SELECT count(*) FROM doctor_opd_charge WHERE doctor_id=$1`, docSchedUser).Scan(&postDelCnt)
	if postDelCnt != 0 {
		t.Fatalf("expected doctor_opd_charge removed after deletion, got count %d", postDelCnt)
	}
	_ = db.QueryRow(ctx, `SELECT opd_charge FROM doctor_profile WHERE user_id=$1`, docSchedUser).Scan(&profileOpdCharge)
	if profileOpdCharge != 0 {
		t.Fatalf("expected doctor_profile.opd_charge reset to 0 after deletion, got %v", profileOpdCharge)
	}

	// Test POST /v1/doctor-opd-charges
	opdBody, _ := json.Marshal(map[string]any{
		"doctorId":       docSchedUser,
		"standardCharge": 420.00,
		"currencySymbol": "ETB",
	})
	respRec = callHTTP("POST", "/v1/doctor-opd-charges", "session=admin", opdBody)
	if respRec.Code != http.StatusOK {
		t.Fatalf("POST /v1/doctor-opd-charges failed: %d %s", respRec.Code, respRec.Body.String())
	}

	// Test GET /v1/doctor-opd-charges/{doctorId}
	respRec = callHTTP("GET", "/v1/doctor-opd-charges/"+docSchedUser, "session=admin", nil)
	if respRec.Code != http.StatusOK {
		t.Fatalf("GET /v1/doctor-opd-charges/{id} failed: %d %s", respRec.Code, respRec.Body.String())
	}

	// Test DELETE /v1/doctor-opd-charges/{doctorId}
	respRec = callHTTP("DELETE", "/v1/doctor-opd-charges/"+docSchedUser, "session=admin", nil)
	if respRec.Code != http.StatusOK {
		t.Fatalf("DELETE /v1/doctor-opd-charges/{id} failed: %d %s", respRec.Code, respRec.Body.String())
	}
}
