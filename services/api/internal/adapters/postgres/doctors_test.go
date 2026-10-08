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
}
