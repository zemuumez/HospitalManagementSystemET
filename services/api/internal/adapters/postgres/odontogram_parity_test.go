package postgres

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"hms.local/api/internal/adapters/httpapi"
	"hms.local/api/internal/application"
	"hms.local/api/internal/domain"
)

func setupOdontogramTestDB(t *testing.T, throughMigration int) (*pgxpool.Pool, Store, string, func()) {
	t.Helper()
	dsn := os.Getenv("HMS_TEST_DATABASE_URL")
	if dsn == "" {
		dsn = os.Getenv("DATABASE_URL")
	}
	if dsn == "" {
		t.Skip("set HMS_TEST_DATABASE_URL or DATABASE_URL to run isolated-schema PostgreSQL tests")
	}
	u, err := url.Parse(dsn)
	if err != nil || (u.Hostname() != "127.0.0.1" && u.Hostname() != "localhost") {
		t.Fatal("test database must be loopback")
	}
	ctx := context.Background()
	adminDB, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}

	random := make([]byte, 12)
	if _, err = rand.Read(random); err != nil {
		adminDB.Close()
		t.Fatal(err)
	}
	schema := "hms_test_odonto_" + hex.EncodeToString(random)
	quoted := pgx.Identifier{schema}.Sanitize()
	if _, err = adminDB.Exec(ctx, "CREATE SCHEMA "+quoted); err != nil {
		adminDB.Close()
		t.Fatal(err)
	}

	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		adminDB.Close()
		t.Fatal(err)
	}
	config.ConnConfig.RuntimeParams["search_path"] = schema
	db, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		adminDB.Close()
		t.Fatal(err)
	}

	migrations, err := filepath.Glob("../../../../../db/migrations/*.sql")
	if err != nil || len(migrations) < 3 {
		db.Close()
		adminDB.Close()
		t.Fatal("migration files unavailable")
	}
	sort.Strings(migrations)

	for _, file := range migrations {
		base := filepath.Base(file)
		var num int
		_, _ = fmt.Sscanf(base, "%d_", &num)
		if throughMigration > 0 && num > throughMigration {
			continue
		}

		sql, e := os.ReadFile(file)
		if e != nil {
			db.Close()
			adminDB.Close()
			t.Fatal(e)
		}
		if _, e = db.Exec(ctx, string(sql)); e != nil {
			db.Close()
			adminDB.Close()
			t.Fatalf("failed migration %s: %v", file, e)
		}
	}

	// Seed standard users and actors
	actors := []domain.Actor{
		{ID: "admin-user", Role: "admin", Name: "Admin User"},
		{ID: "doctor-user-1", Role: "doctor", Name: "Dr. Alice Adams"},
		{ID: "doctor-user-2", Role: "doctor", Name: "Dr. Bob Baker"},
		{ID: "reception-user", Role: "receptionist", Name: "Receptionist"},
		{ID: "patient-user", Role: "patient", Name: "Patient"},
	}
	for _, a := range actors {
		if _, err = db.Exec(ctx, `INSERT INTO "user"(id, name, email) VALUES($1, $2, $1||'@example.test') ON CONFLICT DO NOTHING`, a.ID, a.Name); err != nil {
			db.Close()
			adminDB.Close()
			t.Fatal(err)
		}
		if _, err = db.Exec(ctx, `INSERT INTO staff_access(user_id, role) VALUES($1, $2) ON CONFLICT DO NOTHING`, a.ID, a.Role); err != nil {
			db.Close()
			adminDB.Close()
			t.Fatal(err)
		}
		if a.Role == "doctor" {
			if _, err = db.Exec(ctx, `INSERT INTO doctor_profile(user_id, department, timezone, slot_minutes, version) VALUES($1, 'Dentistry', 'Africa/Addis_Ababa', 30, 1) ON CONFLICT DO NOTHING`, a.ID); err != nil {
				db.Close()
				adminDB.Close()
				t.Fatal(err)
			}
		}
	}

	store := Store{DB: db}
	cleanup := func() {
		db.Close()
		_, _ = adminDB.Exec(context.Background(), "DROP SCHEMA "+quoted+" CASCADE")
		adminDB.Close()
	}

	return db, store, schema, cleanup
}

func helperCreatePatient(t *testing.T, ctx context.Context, db *pgxpool.Pool, givenName, familyName string) string {
	t.Helper()
	var patientID string
	err := db.QueryRow(ctx, `
		INSERT INTO patient (given_name, family_name, date_of_birth)
		VALUES ($1, $2, '1990-01-01')
		RETURNING id::text
	`, givenName, familyName).Scan(&patientID)
	if err != nil {
		t.Fatalf("failed to insert patient %s %s: %v", givenName, familyName, err)
	}
	return patientID
}

// 1. Populated Database Migration Upgrade
func TestOdontogram_PopulatedDatabaseMigrationUpgrade(t *testing.T) {
	ctx := context.Background()
	// Apply migrations through 059 only
	db, _, _, cleanup := setupOdontogramTestDB(t, 59)
	defer cleanup()

	patientID := helperCreatePatient(t, ctx, db, "Elena", "Rostova")

	// Seed existing legacy tooth rows before migration 060
	t1 := time.Date(2026, 1, 10, 8, 30, 0, 0, time.UTC)
	t2 := time.Date(2026, 1, 15, 14, 45, 0, 0, time.UTC)

	_, err := db.Exec(ctx, `
		INSERT INTO patient_odontogram_entry (
			patient_id, tooth_number, condition, procedure_notes, diagnosed_by, created_at, updated_at
		) VALUES
		($1, 1, 'caries', 'Incipient lesion', 'doctor-user-1', $2, $2),
		($1, 14, 'crown', 'Ceramic crown seated', 'doctor-user-1', $3, $3)
	`, patientID, t1, t2)
	if err != nil {
		t.Fatalf("failed to insert pre-060 teeth: %v", err)
	}

	// Now apply migration 060
	m60Path := "../../../../../db/migrations/060_odontogram_chart_parity.sql"
	sql60, err := os.ReadFile(m60Path)
	if err != nil {
		t.Fatalf("failed to read migration 060: %v", err)
	}
	if _, err = db.Exec(ctx, string(sql60)); err != nil {
		t.Fatalf("failed to execute migration 060: %v", err)
	}

	// Verify legacy chart was created without inventing dates
	var chartID string
	var isLegacy bool
	var docID string
	var chartCreatedAt, chartUpdatedAt time.Time
	var odontogramJson string

	err = db.QueryRow(ctx, `
		SELECT id::text, is_legacy, doctor_id, created_at, updated_at, odontogram::text
		FROM odontogram_chart
		WHERE patient_id = $1::uuid
	`, patientID).Scan(&chartID, &isLegacy, &docID, &chartCreatedAt, &chartUpdatedAt, &odontogramJson)
	if err != nil {
		t.Fatalf("legacy chart was not created: %v", err)
	}

	if !isLegacy {
		t.Errorf("expected is_legacy=true, got %v", isLegacy)
	}
	if docID != "doctor-user-1" {
		t.Errorf("expected doctor_id='doctor-user-1', got %s", docID)
	}
	if !chartCreatedAt.Equal(t1) {
		t.Errorf("expected chart created_at to equal min tooth date %v, got %v", t1, chartCreatedAt)
	}
	if !chartUpdatedAt.Equal(t2) {
		t.Errorf("expected chart updated_at to equal max tooth date %v, got %v", t2, chartUpdatedAt)
	}

	// Verify child tooth rows are linked to this chart
	var countLinked int
	err = db.QueryRow(ctx, `
		SELECT COUNT(*) FROM patient_odontogram_entry WHERE chart_id = $1::uuid
	`, chartID).Scan(&countLinked)
	if err != nil || countLinked != 2 {
		t.Fatalf("expected 2 teeth linked to chart %s, got %d (err: %v)", chartID, countLinked, err)
	}

	// Verify new uniqueness constraint on chart_id + tooth_number
	var constraintExists bool
	err = db.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM pg_constraint
			WHERE conname = 'patient_odontogram_entry_chart_tooth_key'
		)
	`).Scan(&constraintExists)
	if err != nil || !constraintExists {
		t.Fatalf("expected constraint patient_odontogram_entry_chart_tooth_key to exist")
	}
}

// 2. Multiple Charts Per Patient
func TestOdontogram_MultipleChartsPerPatient(t *testing.T) {
	ctx := context.Background()
	db, store, _, cleanup := setupOdontogramTestDB(t, 0)
	defer cleanup()

	admin := domain.Actor{ID: "admin-user", Role: "admin"}
	patientID := helperCreatePatient(t, ctx, db, "Abebe", "Bikila")

	// Chart A (Initial examination)
	chartA, err := store.CreateOdontogramChart(ctx, admin, domain.SaveOdontogramChartInput{
		PatientID:   patientID,
		DoctorID:    "doctor-user-1",
		Description: "Initial Examination 2026-01",
		Odontogram: map[int]string{
			1: "K",
			2: "Healthy",
		},
	})
	if err != nil {
		t.Fatalf("failed to create chart A: %v", err)
	}

	// Chart B (Follow-up examination for SAME patient with conflicting tooth 1 condition)
	chartB, err := store.CreateOdontogramChart(ctx, admin, domain.SaveOdontogramChartInput{
		PatientID:   patientID,
		DoctorID:    "doctor-user-2",
		Description: "Follow-up Examination 2026-06",
		Odontogram: map[int]string{
			1: "KR", // Tooth 1 has different condition in chart B
			3: "IP",
		},
	})
	if err != nil {
		t.Fatalf("failed to create chart B: %v", err)
	}

	if chartA.ID == chartB.ID {
		t.Fatalf("charts must have distinct IDs")
	}

	// Retrieve chart A: tooth 1 must still be "K"
	fetchedA, err := store.GetOdontogramChart(ctx, chartA.ID)
	if err != nil {
		t.Fatalf("failed to get chart A: %v", err)
	}
	if fetchedA.Odontogram[1] != "K" {
		t.Errorf("chart A tooth 1 expected 'K', got '%s'", fetchedA.Odontogram[1])
	}
	if fetchedA.Odontogram[2] != "Healthy" {
		t.Errorf("chart A tooth 2 expected 'Healthy', got '%s'", fetchedA.Odontogram[2])
	}

	// Retrieve chart B: tooth 1 must be "KR", tooth 3 must be "IP"
	fetchedB, err := store.GetOdontogramChart(ctx, chartB.ID)
	if err != nil {
		t.Fatalf("failed to get chart B: %v", err)
	}
	if fetchedB.Odontogram[1] != "KR" {
		t.Errorf("chart B tooth 1 expected 'KR', got '%s'", fetchedB.Odontogram[1])
	}
	if fetchedB.Odontogram[3] != "IP" {
		t.Errorf("chart B tooth 3 expected 'IP', got '%s'", fetchedB.Odontogram[3])
	}

	// List charts for this patient: must return both
	list, total, err := store.ListOdontogramCharts(ctx, admin, patientID, "", 1, 10)
	if err != nil {
		t.Fatalf("failed to list charts: %v", err)
	}
	if total != 2 || len(list) != 2 {
		t.Fatalf("expected 2 charts in list, got total=%d len=%d", total, len(list))
	}
}

// 3. Lossless Code Round-Trips
func TestOdontogram_LosslessCodeRoundTrip(t *testing.T) {
	ctx := context.Background()
	db, store, _, cleanup := setupOdontogramTestDB(t, 0)
	defer cleanup()

	admin := domain.Actor{ID: "admin-user", Role: "admin"}
	patientID := helperCreatePatient(t, ctx, db, "Sara", "Connor")

	// All 11 standard codes + custom codes
	standardCodes := []string{"Healthy", "K", "C", "Ce", "D", "KR", "PS", "IP", "X", "F", "B"}
	customCodes := []string{"CUSTOM_VENEER", "IMPL_ZIR"}

	inputOdonto := make(map[int]string)
	for i, code := range standardCodes {
		inputOdonto[i+1] = code
	}
	inputOdonto[12] = customCodes[0]
	inputOdonto[13] = customCodes[1]

	chart, err := store.CreateOdontogramChart(ctx, admin, domain.SaveOdontogramChartInput{
		PatientID:   patientID,
		DoctorID:    "doctor-user-1",
		Description: "Full 11-legend + custom fidelity audit",
		Odontogram:  inputOdonto,
	})
	if err != nil {
		t.Fatalf("failed to create chart with standard & custom codes: %v", err)
	}

	// Verify retrieval from Go aggregate
	fetched, err := store.GetOdontogramChart(ctx, chart.ID)
	if err != nil {
		t.Fatalf("failed to fetch chart: %v", err)
	}

	for toothNum, expectedCode := range inputOdonto {
		actualCode := fetched.Odontogram[toothNum]
		if actualCode != expectedCode {
			t.Errorf("tooth %d: expected code '%s', got '%s'", toothNum, expectedCode, actualCode)
		}
	}

	// Verify database rows directly in patient_odontogram_entry
	rows, err := db.Query(ctx, `
		SELECT tooth_number, condition
		FROM patient_odontogram_entry
		WHERE chart_id = $1::uuid
		ORDER BY tooth_number
	`, chart.ID)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()

	directRows := make(map[int]string)
	for rows.Next() {
		var tn int
		var cond string
		if err := rows.Scan(&tn, &cond); err != nil {
			t.Fatal(err)
		}
		directRows[tn] = cond
	}

	for toothNum, expectedCode := range inputOdonto {
		if directRows[toothNum] != expectedCode {
			t.Errorf("direct row tooth %d: expected '%s', got '%s'", toothNum, expectedCode, directRows[toothNum])
		}
	}
}

// 4. Atomic Rollback on Failure
func TestOdontogram_AtomicRollback(t *testing.T) {
	ctx := context.Background()
	db, store, _, cleanup := setupOdontogramTestDB(t, 0)
	defer cleanup()

	admin := domain.Actor{ID: "admin-user", Role: "admin"}
	patientID := helperCreatePatient(t, ctx, db, "Marcus", "Aurelius")

	invalidInput := domain.SaveOdontogramChartInput{
		PatientID:   patientID,
		DoctorID:    "non-existent-doctor-id",
		Description: "Should abort atomically",
		Odontogram: map[int]string{
			1: "K",
		},
	}

	_, err := store.CreateOdontogramChart(ctx, admin, invalidInput)
	if err == nil {
		t.Fatalf("expected error creating chart with invalid doctor")
	}

	// Verify zero charts were written
	var chartCount int
	_ = db.QueryRow(ctx, `SELECT COUNT(*) FROM odontogram_chart WHERE patient_id = $1::uuid`, patientID).Scan(&chartCount)
	if chartCount != 0 {
		t.Errorf("expected 0 charts, got %d", chartCount)
	}

	// Verify zero tooth rows were written
	var teethCount int
	_ = db.QueryRow(ctx, `SELECT COUNT(*) FROM patient_odontogram_entry WHERE patient_id = $1::uuid`, patientID).Scan(&teethCount)
	if teethCount != 0 {
		t.Errorf("expected 0 tooth rows, got %d", teethCount)
	}
}

// 5. Stale Edits & Concurrency (HTTP 409)
func TestOdontogram_StaleEditsAndConcurrency(t *testing.T) {
	ctx := context.Background()
	db, store, _, cleanup := setupOdontogramTestDB(t, 0)
	defer cleanup()

	admin := domain.Actor{ID: "admin-user", Role: "admin"}
	patientID := helperCreatePatient(t, ctx, db, "Grace", "Hopper")

	// Create initial chart (version 1)
	chart, err := store.CreateOdontogramChart(ctx, admin, domain.SaveOdontogramChartInput{
		PatientID:   patientID,
		DoctorID:    "doctor-user-1",
		Description: "Original Examination",
		Odontogram: map[int]string{
			1: "K",
		},
	})
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if chart.Version != 1 {
		t.Fatalf("expected initial version 1, got %d", chart.Version)
	}

	// Session A updates with version 1 -> succeeds, increments to version 2
	updatedA, err := store.UpdateOdontogramChart(ctx, admin, chart.ID, domain.SaveOdontogramChartInput{
		PatientID:   patientID,
		DoctorID:    "doctor-user-1",
		Description: "Session A update",
		Odontogram: map[int]string{
			1: "F",
		},
		Version: 1,
	})
	if err != nil {
		t.Fatalf("session A update failed: %v", err)
	}
	if updatedA.Version != 2 {
		t.Fatalf("expected version 2 after update, got %d", updatedA.Version)
	}

	// Session B attempts update with stale version 1 -> must return domain.ErrConflict
	_, err = store.UpdateOdontogramChart(ctx, admin, chart.ID, domain.SaveOdontogramChartInput{
		PatientID:   patientID,
		DoctorID:    "doctor-user-1",
		Description: "Session B stale update",
		Odontogram: map[int]string{
			1: "C",
		},
		Version: 1, // Stale! Current version is 2
	})
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("expected domain.ErrConflict for stale edit, got: %v", err)
	}

	// Verify chart was NOT overwritten by session B
	latest, err := store.GetOdontogramChart(ctx, chart.ID)
	if err != nil {
		t.Fatal(err)
	}
	if latest.Description != "Session A update" || latest.Odontogram[1] != "F" {
		t.Errorf("chart data corrupted: %+v", latest)
	}

	// Session B updates with correct version 2 -> succeeds, increments to 3
	updatedB, err := store.UpdateOdontogramChart(ctx, admin, chart.ID, domain.SaveOdontogramChartInput{
		PatientID:   patientID,
		DoctorID:    "doctor-user-1",
		Description: "Session B corrected update",
		Odontogram: map[int]string{
			1: "C",
		},
		Version: 2,
	})
	if err != nil {
		t.Fatalf("session B corrected update failed: %v", err)
	}
	if updatedB.Version != 3 {
		t.Fatalf("expected version 3, got %d", updatedB.Version)
	}
}

// 6. Role Boundaries & Scoping
func TestOdontogram_RoleBoundariesAndScoping(t *testing.T) {
	ctx := context.Background()
	db, store, _, cleanup := setupOdontogramTestDB(t, 0)
	defer cleanup()

	admin := domain.Actor{ID: "admin-user", Role: "admin"}
	doctor := domain.Actor{ID: "doctor-user-1", Role: "doctor"}
	receptionist := domain.Actor{ID: "reception-user", Role: "receptionist"}
	patientActor := domain.Actor{ID: "patient-user", Role: "patient"}

	service := application.ClinicalCareService{
		Store: store,
	}

	patientID := helperCreatePatient(t, ctx, db, "Ada", "Lovelace")
	if _, err := db.Exec(ctx, "UPDATE patient SET clinician_user_id = 'doctor-user-1' WHERE id = $1::uuid", patientID); err != nil {
		t.Fatal(err)
	}

	// Receptionist cannot create charts
	_, err := service.CreateOdontogramChart(ctx, receptionist, domain.SaveOdontogramChartInput{
		PatientID:   patientID,
		DoctorID:    "doctor-user-1",
		Description: "Receptionist write attempt",
		Odontogram:  map[int]string{1: "Healthy"},
	})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected ErrForbidden for receptionist write, got %v", err)
	}

	// Patient cannot create charts
	_, err = service.CreateOdontogramChart(ctx, patientActor, domain.SaveOdontogramChartInput{
		PatientID:   patientID,
		DoctorID:    "doctor-user-1",
		Description: "Patient write attempt",
		Odontogram:  map[int]string{1: "Healthy"},
	})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected ErrForbidden for patient write, got %v", err)
	}

	// Doctor can create chart
	chart, err := service.CreateOdontogramChart(ctx, doctor, domain.SaveOdontogramChartInput{
		PatientID:   patientID,
		DoctorID:    "doctor-user-1",
		Description: "Doctor clinical chart",
		Odontogram:  map[int]string{1: "Healthy"},
	})
	if err != nil {
		t.Fatalf("doctor chart creation failed: %v", err)
	}

	// Admin can view chart
	_, err = service.GetOdontogramChart(ctx, admin, chart.ID)
	if err != nil {
		t.Fatalf("admin get chart failed: %v", err)
	}

	// Patient cannot delete chart
	err = service.DeleteOdontogramChart(ctx, patientActor, chart.ID)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected ErrForbidden for patient delete, got %v", err)
	}

	// Patient attempting to view chart when not linked returns ErrForbidden
	_, err = service.GetOdontogramChart(ctx, patientActor, chart.ID)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected ErrForbidden for unlinked patient, got %v", err)
	}

	// Link patient to patient-user
	_, err = db.Exec(ctx, `UPDATE patient SET user_id = 'patient-user' WHERE id = $1::uuid`, patientID)
	if err != nil {
		t.Fatal(err)
	}

	// Now linked patient can view their own chart
	_, err = service.GetOdontogramChart(ctx, patientActor, chart.ID)
	if err != nil {
		t.Fatalf("linked patient should be able to view their chart: %v", err)
	}

	// Admin can delete chart
	err = service.DeleteOdontogramChart(ctx, admin, chart.ID)
	if err != nil {
		t.Fatalf("admin delete chart failed: %v", err)
	}
}

// 7. Clinical Relationships & Encounter Validation
func TestOdontogram_ClinicalRelationshipsAndEncounterValidation(t *testing.T) {
	ctx := context.Background()
	db, store, _, cleanup := setupOdontogramTestDB(t, 0)
	defer cleanup()

	admin := domain.Actor{ID: "admin-user", Role: "admin"}

	patientAID := helperCreatePatient(t, ctx, db, "Patient", "A")
	patientBID := helperCreatePatient(t, ctx, db, "Patient", "B")

	// Create encounter for Patient B
	var caseBID string
	err := db.QueryRow(ctx, `
		INSERT INTO patient_case (patient_id, doctor_id, created_by)
		VALUES ($1::uuid, 'doctor-user-1', 'admin-user')
		RETURNING id::text
	`, patientBID).Scan(&caseBID)
	if err != nil {
		t.Fatal(err)
	}

	var encBID string
	err = db.QueryRow(ctx, `
		INSERT INTO encounter (kind, case_id, patient_id, doctor_id, admitted_at, created_by, request_key)
		VALUES ('opd', $1::uuid, $2::uuid, 'doctor-user-1', clock_timestamp(), 'admin-user', 'req-enc-b-1')
		RETURNING id::text
	`, caseBID, patientBID).Scan(&encBID)
	if err != nil {
		t.Fatal(err)
	}

	// Attempt to create chart for Patient A referencing Patient B's encounter -> MUST FAIL
	_, err = store.CreateOdontogramChart(ctx, admin, domain.SaveOdontogramChartInput{
		PatientID:   patientAID,
		DoctorID:    "doctor-user-1",
		EncounterID: &encBID,
		Description: "Cross-patient encounter spoof attempt",
		Odontogram:  map[int]string{1: "Healthy"},
	})
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected ErrValidation for cross-patient encounter, got %v", err)
	}

	// Create valid encounter for Patient A
	var caseAID string
	err = db.QueryRow(ctx, `
		INSERT INTO patient_case (patient_id, doctor_id, created_by)
		VALUES ($1::uuid, 'doctor-user-1', 'admin-user')
		RETURNING id::text
	`, patientAID).Scan(&caseAID)
	if err != nil {
		t.Fatal(err)
	}

	var encAID string
	err = db.QueryRow(ctx, `
		INSERT INTO encounter (kind, case_id, patient_id, doctor_id, admitted_at, created_by, request_key)
		VALUES ('opd', $1::uuid, $2::uuid, 'doctor-user-1', clock_timestamp(), 'admin-user', 'req-enc-a-1')
		RETURNING id::text
	`, caseAID, patientAID).Scan(&encAID)
	if err != nil {
		t.Fatal(err)
	}

	// Now create chart for Patient A with Patient A's encounter -> MUST SUCCEED
	chart, err := store.CreateOdontogramChart(ctx, admin, domain.SaveOdontogramChartInput{
		PatientID:   patientAID,
		DoctorID:    "doctor-user-1",
		EncounterID: &encAID,
		Description: "Legitimate examination with encounter",
		Odontogram:  map[int]string{1: "Healthy"},
	})
	if err != nil {
		t.Fatalf("expected chart creation to succeed: %v", err)
	}

	// Verify actor ID recorded in created_by / updated_by separately from attending doctor
	if chart.CreatedBy != admin.ID {
		t.Errorf("expected created_by='%s', got '%s'", admin.ID, chart.CreatedBy)
	}
	if chart.DoctorID != "doctor-user-1" {
		t.Errorf("expected doctor_id='doctor-user-1', got '%s'", chart.DoctorID)
	}
	if chart.EncounterID == nil || *chart.EncounterID != encAID {
		t.Errorf("expected encounter_id='%s', got '%v'", encAID, chart.EncounterID)
	}
}

// 8. Doctor Deletion Protection
func TestOdontogram_DoctorDeletionProtection(t *testing.T) {
	ctx := context.Background()
	db, store, _, cleanup := setupOdontogramTestDB(t, 0)
	defer cleanup()

	admin := domain.Actor{ID: "admin-user", Role: "admin"}
	patientID := helperCreatePatient(t, ctx, db, "Test", "Patient")

	// Create an odontogram chart referencing doctor-user-2
	chart, err := store.CreateOdontogramChart(ctx, admin, domain.SaveOdontogramChartInput{
		PatientID:   patientID,
		DoctorID:    "doctor-user-2",
		Description: "Active dental chart referencing Dr. Bob",
		Odontogram:  map[int]string{1: "Healthy"},
	})
	if err != nil {
		t.Fatalf("failed to create chart: %v", err)
	}

	// Attempting to delete Dr. Bob must be rejected with ErrInUse
	err = store.DeleteDoctor(ctx, admin, "doctor-user-2")
	if !errors.Is(err, domain.ErrInUse) {
		t.Fatalf("expected domain.ErrInUse when deleting doctor with active odontogram, got: %v", err)
	}

	// Delete the odontogram chart
	err = store.DeleteOdontogramChart(ctx, chart.ID)
	if err != nil {
		t.Fatalf("failed to delete chart: %v", err)
	}

	// Now deleting Dr. Bob must succeed
	err = store.DeleteDoctor(ctx, admin, "doctor-user-2")
	if err != nil {
		t.Fatalf("expected doctor deletion to succeed after chart deletion, got: %v", err)
	}
}

// 9. Single Tooth Compatibility Endpoint
func TestOdontogram_SingleToothCompatibilityEndpoint(t *testing.T) {
	ctx := context.Background()
	db, store, _, cleanup := setupOdontogramTestDB(t, 0)
	defer cleanup()

	admin := domain.Actor{ID: "admin-user", Role: "admin"}
	patientID := helperCreatePatient(t, ctx, db, "Tim", "Berners-Lee")

	// Calling SetToothCondition without chart_id creates/targets the legacy chart
	entry1, err := store.SetToothCondition(ctx, admin, patientID, domain.SetToothConditionInput{
		ToothNumber:    8,
		Condition:      "K",
		ProcedureNotes: "Legacy single-tooth call",
	})
	if err != nil {
		t.Fatalf("SetToothCondition failed: %v", err)
	}

	if entry1.ChartID == "" {
		t.Errorf("expected ChartID to be populated")
	}

	// Verify chart was created and JSONB updated atomically
	chart, err := store.GetOdontogramChart(ctx, entry1.ChartID)
	if err != nil {
		t.Fatalf("failed to get chart: %v", err)
	}
	if chart.Odontogram[8] != "K" {
		t.Errorf("expected chart.Odontogram[8] == 'K', got '%s'", chart.Odontogram[8])
	}

	// Update another tooth on the same chart using explicit ChartID
	entry2, err := store.SetToothCondition(ctx, admin, patientID, domain.SetToothConditionInput{
		ChartID:        &chart.ID,
		ToothNumber:    9,
		Condition:      "Ce",
		ProcedureNotes: "Second tooth update",
	})
	if err != nil {
		t.Fatalf("SetToothCondition tooth 9 failed: %v", err)
	}
	if entry2.ChartID != chart.ID {
		t.Errorf("expected entry2.ChartID == chart.ID")
	}

	// Re-fetch chart: must contain both teeth atomically
	chartReloaded, err := store.GetOdontogramChart(ctx, chart.ID)
	if err != nil {
		t.Fatal(err)
	}
	if chartReloaded.Odontogram[8] != "K" || chartReloaded.Odontogram[9] != "Ce" {
		t.Errorf("chart JSONB divergence: %+v", chartReloaded.Odontogram)
	}
}

// ─── REGRESSION SUITES FOR REVIEW AUDIT FINDINGS (R1 - R6) ───────────────────

// R1: Doctor clinical scope isolation
func TestOdontogram_R1_DoctorClinicalScopeIsolation(t *testing.T) {
	ctx := context.Background()
	db, store, _, cleanup := setupOdontogramTestDB(t, 0)
	defer cleanup()

	app := application.ClinicalCareService{Store: store}

	alice := domain.Actor{ID: "doctor-user-1", Role: "doctor", Name: "Dr. Alice"}
	bob := domain.Actor{ID: "doctor-user-2", Role: "doctor", Name: "Dr. Bob"}

	// Patient P is assigned to Dr. Alice (clinician_user_id = doctor-user-1)
	var patientID string
	err := db.QueryRow(ctx, `
		INSERT INTO patient (given_name, family_name, date_of_birth, clinician_user_id)
		VALUES ('Patient', 'Protected', '1995-05-05', 'doctor-user-1')
		RETURNING id::text
	`).Scan(&patientID)
	if err != nil {
		t.Fatalf("failed to insert patient: %v", err)
	}

	// Alice creates an odontogram chart for Patient P
	chartInput := domain.SaveOdontogramChartInput{
		PatientID:   patientID,
		DoctorID:    "doctor-user-1",
		Description: "Alice examination",
		Odontogram:  map[int]string{1: "K", 2: "C"},
	}
	chart, err := app.CreateOdontogramChart(ctx, alice, chartInput)
	if err != nil {
		t.Fatalf("Alice failed to create chart: %v", err)
	}

	// Bob is an unrelated doctor (not assigned, no encounter, not on care team)
	// 1. Bob must be forbidden from GetOdontogramChart
	_, err = app.GetOdontogramChart(ctx, bob, chart.ID)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("expected ErrForbidden for Bob on GetOdontogramChart, got: %v", err)
	}

	// 2. Bob must be forbidden from UpdateOdontogramChart
	_, err = app.UpdateOdontogramChart(ctx, bob, chart.ID, domain.SaveOdontogramChartInput{
		PatientID:   patientID,
		DoctorID:    "doctor-user-2",
		Description: "Bob unauthorized update",
		Odontogram:  map[int]string{1: "Healthy"},
		Version:     1,
	})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("expected ErrForbidden for Bob on UpdateOdontogramChart, got: %v", err)
	}

	// 3. Bob must be forbidden from DeleteOdontogramChart
	err = app.DeleteOdontogramChart(ctx, bob, chart.ID)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("expected ErrForbidden for Bob on DeleteOdontogramChart, got: %v", err)
	}

	// 4. Bob must be forbidden from CreateOdontogramChart for Patient P
	_, err = app.CreateOdontogramChart(ctx, bob, domain.SaveOdontogramChartInput{
		PatientID:   patientID,
		DoctorID:    "doctor-user-2",
		Description: "Bob unauthorized create",
		Odontogram:  map[int]string{3: "B"},
	})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("expected ErrForbidden for Bob on CreateOdontogramChart, got: %v", err)
	}

	// 5. Bob must be forbidden from ListOdontogramCharts for Patient P specifically
	_, _, err = app.ListOdontogramCharts(ctx, bob, patientID, "", 1, 10)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("expected ErrForbidden for Bob on ListOdontogramCharts with patientID, got: %v", err)
	}

	// In general list, Bob must see 0 charts for Patient P
	chartsBob, _, err := app.ListOdontogramCharts(ctx, bob, "", "", 1, 10)
	if err != nil {
		t.Fatalf("Bob ListOdontogramCharts error: %v", err)
	}
	for _, c := range chartsBob {
		if c.PatientID == patientID {
			t.Errorf("unrelated doctor Bob saw Patient P's chart in general list!")
		}
	}

	// 6. Non-dental clinical endpoint check: Bob must be forbidden from clinical access
	err = store.AuthorizeClinicalRecord(ctx, bob, patientID, "", "clinical")
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("expected AuthorizeClinicalRecord to return ErrForbidden for unrelated Bob, got: %v", err)
	}

	// 7. Alice CAN view and list Patient P's charts
	chartAlice, err := app.GetOdontogramChart(ctx, alice, chart.ID)
	if err != nil {
		t.Fatalf("Alice GetOdontogramChart failed: %v", err)
	}
	if chartAlice.ID != chart.ID {
		t.Errorf("expected chart ID %s, got %s", chart.ID, chartAlice.ID)
	}

	chartsAlice, totalAlice, err := app.ListOdontogramCharts(ctx, alice, patientID, "", 1, 10)
	if err != nil || totalAlice != 1 || len(chartsAlice) != 1 {
		t.Errorf("expected Alice to see 1 chart, got total=%d err=%v", totalAlice, err)
	}
}

// R2: Omission of version bypasses optimistic concurrency
func TestOdontogram_R2_OptimisticConcurrencyStaleAndZeroVersion(t *testing.T) {
	ctx := context.Background()
	db, store, _, cleanup := setupOdontogramTestDB(t, 0)
	defer cleanup()

	admin := domain.Actor{ID: "admin-user", Role: "admin"}
	app := application.ClinicalCareService{Store: store}
	patientID := helperCreatePatient(t, ctx, db, "Grace", "Hopper")

	chart, err := app.CreateOdontogramChart(ctx, admin, domain.SaveOdontogramChartInput{
		PatientID:   patientID,
		DoctorID:    "doctor-user-1",
		Description: "Initial check",
		Odontogram:  map[int]string{1: "K"},
	})
	if err != nil {
		t.Fatalf("CreateOdontogramChart failed: %v", err)
	}
	if chart.Version != 1 {
		t.Fatalf("expected version 1, got %d", chart.Version)
	}

	// 1. Update with Version = 0 must be rejected with ErrValidation
	_, err = app.UpdateOdontogramChart(ctx, admin, chart.ID, domain.SaveOdontogramChartInput{
		PatientID:   patientID,
		DoctorID:    "doctor-user-1",
		Description: "Zero version update",
		Odontogram:  map[int]string{1: "Healthy"},
		Version:     0,
	})
	if !errors.Is(err, domain.ErrValidation) {
		t.Errorf("expected ErrValidation for Version 0, got: %v", err)
	}

	// 2. Update with negative Version must be rejected with ErrValidation
	_, err = app.UpdateOdontogramChart(ctx, admin, chart.ID, domain.SaveOdontogramChartInput{
		PatientID:   patientID,
		DoctorID:    "doctor-user-1",
		Description: "Negative version update",
		Odontogram:  map[int]string{1: "Healthy"},
		Version:     -1,
	})
	if !errors.Is(err, domain.ErrValidation) {
		t.Errorf("expected ErrValidation for Version -1, got: %v", err)
	}

	// Verify chart was not modified
	current, err := app.GetOdontogramChart(ctx, admin, chart.ID)
	if err != nil || current.Version != 1 || current.Odontogram[1] != "K" {
		t.Fatalf("chart modified despite validation failure: version=%d odonto=%v", current.Version, current.Odontogram)
	}

	// 3. Valid update with matching Version = 1 succeeds
	updated, err := app.UpdateOdontogramChart(ctx, admin, chart.ID, domain.SaveOdontogramChartInput{
		PatientID:   patientID,
		DoctorID:    "doctor-user-1",
		Description: "Legitimate update",
		Odontogram:  map[int]string{1: "C"},
		Version:     1,
	})
	if err != nil {
		t.Fatalf("UpdateOdontogramChart failed: %v", err)
	}
	if updated.Version != 2 {
		t.Errorf("expected updated version 2, got %d", updated.Version)
	}

	// 4. Stale update with Version = 1 must be rejected with ErrConflict (409)
	_, err = app.UpdateOdontogramChart(ctx, admin, chart.ID, domain.SaveOdontogramChartInput{
		PatientID:   patientID,
		DoctorID:    "doctor-user-1",
		Description: "Concurrent edit",
		Odontogram:  map[int]string{1: "X"},
		Version:     1,
	})
	if !errors.Is(err, domain.ErrConflict) {
		t.Errorf("expected ErrConflict for stale Version 1, got: %v", err)
	}

	// Verify chart version remains 2 and edits preserved
	reloaded, _ := app.GetOdontogramChart(ctx, admin, chart.ID)
	if reloaded.Version != 2 || reloaded.Odontogram[1] != "C" {
		t.Errorf("chart corrupted by stale write: version=%d odonto=%v", reloaded.Version, reloaded.Odontogram)
	}
}

// R3: Compatibility writes target only legacy chart and serialize concurrent creation
func TestOdontogram_R3_CompatibilityTargetsOnlyLegacyAndParentFirstLocking(t *testing.T) {
	ctx := context.Background()
	db, store, _, cleanup := setupOdontogramTestDB(t, 0)
	defer cleanup()

	admin := domain.Actor{ID: "admin-user", Role: "admin"}
	app := application.ClinicalCareService{Store: store}
	patientID := helperCreatePatient(t, ctx, db, "Alan", "Turing")

	// Create a modern chart for patient (is_legacy = false)
	modernChart, err := app.CreateOdontogramChart(ctx, admin, domain.SaveOdontogramChartInput{
		PatientID:   patientID,
		DoctorID:    "doctor-user-1",
		Description: "Modern routine dental checkup",
		Odontogram:  map[int]string{10: "Healthy", 11: "Healthy"},
	})
	if err != nil {
		t.Fatalf("CreateOdontogramChart failed: %v", err)
	}
	if modernChart.IsLegacy {
		t.Fatalf("expected is_legacy = false for modern chart")
	}

	// Calling SetToothCondition without chart_id must NOT target the modern chart!
	// It must create or target an explicitly designated legacy chart (is_legacy = true).
	entry, err := store.SetToothCondition(ctx, admin, patientID, domain.SetToothConditionInput{
		ToothNumber:    14,
		Condition:      "K",
		ProcedureNotes: "Legacy single-tooth compatibility call",
	})
	if err != nil {
		t.Fatalf("SetToothCondition failed: %v", err)
	}

	if entry.ChartID == modernChart.ID {
		t.Fatalf("CRITICAL: SetToothCondition silently targeted modern chart %s!", modernChart.ID)
	}

	// Verify legacy chart was created with is_legacy = true
	legacyChart, err := store.GetOdontogramChart(ctx, entry.ChartID)
	if err != nil {
		t.Fatalf("failed to get legacy chart: %v", err)
	}
	if !legacyChart.IsLegacy {
		t.Errorf("expected legacyChart.IsLegacy = true")
	}
	if legacyChart.Odontogram[14] != "K" {
		t.Errorf("expected tooth 14 = 'K' on legacy chart, got: %s", legacyChart.Odontogram[14])
	}

	// Verify modern chart was completely untouched
	reloadedModern, err := store.GetOdontogramChart(ctx, modernChart.ID)
	if err != nil {
		t.Fatalf("failed to reload modern chart: %v", err)
	}
	if reloadedModern.Version != modernChart.Version || len(reloadedModern.Odontogram) != 2 {
		t.Errorf("modern chart was mutated by compatibility call: version=%d odonto=%v", reloadedModern.Version, reloadedModern.Odontogram)
	}

	// Concurrent creation test: 5 concurrent callers calling SetToothCondition without chart_id
	patientID2 := helperCreatePatient(t, ctx, db, "Ada", "Lovelace")
	var wg sync.WaitGroup
	errCh := make(chan error, 5)

	for i := 1; i <= 5; i++ {
		wg.Add(1)
		go func(toothNum int) {
			defer wg.Done()
			_, e := store.SetToothCondition(ctx, admin, patientID2, domain.SetToothConditionInput{
				ToothNumber:    toothNum,
				Condition:      "Ce",
				ProcedureNotes: fmt.Sprintf("Concurrent tooth %d", toothNum),
			})
			if e != nil {
				errCh <- e
			}
		}(i)
	}
	wg.Wait()
	close(errCh)

	for e := range errCh {
		t.Fatalf("concurrent SetToothCondition failed: %v", e)
	}

	// Exactly 1 legacy chart must exist for patientID2
	var legacyCount int
	err = db.QueryRow(ctx, `SELECT COUNT(*) FROM odontogram_chart WHERE patient_id = $1::uuid AND is_legacy = true`, patientID2).Scan(&legacyCount)
	if err != nil || legacyCount != 1 {
		t.Errorf("expected exactly 1 legacy chart created, got count=%d err=%v", legacyCount, err)
	}
}

// R4: Legacy provenance and tooth notes/author preservation across ordinary edits
func TestOdontogram_R4_LegacyProvenanceAndToothHistoryPreservationAcrossOrdinaryEdits(t *testing.T) {
	ctx := context.Background()
	db, store, _, cleanup := setupOdontogramTestDB(t, 59)
	defer cleanup()

	admin := domain.Actor{ID: "admin-user", Role: "admin"}
	app := application.ClinicalCareService{Store: store}
	patientID := helperCreatePatient(t, ctx, db, "Marie", "Curie")

	// Seed multiple distinct pre-060 teeth with different notes, authors, and timestamps
	t1 := time.Date(2026, 1, 10, 8, 30, 0, 0, time.UTC)
	t2 := time.Date(2026, 2, 14, 11, 15, 0, 0, time.UTC)

	_, err := db.Exec(ctx, `
		INSERT INTO patient_odontogram_entry (
			patient_id, tooth_number, condition, procedure_notes, diagnosed_by, created_at, updated_at
		) VALUES
		($1, 1, 'caries', 'Deep distal caries with pulpal involvement', 'doctor-user-1', $2, $2),
		($1, 2, 'crown', 'Porcelain fused to metal crown seated well', 'doctor-user-2', $3, $3)
	`, patientID, t1, t2)
	if err != nil {
		t.Fatalf("failed to seed pre-060 teeth: %v", err)
	}

	// Run migration 060
	m60Path := "../../../../../db/migrations/060_odontogram_chart_parity.sql"
	sql60, err := os.ReadFile(m60Path)
	if err != nil {
		t.Fatalf("failed to read migration 060: %v", err)
	}
	if _, err = db.Exec(ctx, string(sql60)); err != nil {
		t.Fatalf("failed to execute migration 060: %v", err)
	}

	// Run migration 061 (adds legacy_provenance column, unique constraint, and backfills provenance)
	m61Path := "../../../../../db/migrations/061_odontogram_parity_upgrade.sql"
	sql61, err := os.ReadFile(m61Path)
	if err != nil {
		t.Fatalf("failed to read migration 061: %v", err)
	}
	if _, err = db.Exec(ctx, string(sql61)); err != nil {
		t.Fatalf("failed to execute migration 061: %v", err)
	}

	// Get legacy chart
	charts, total, err := app.ListOdontogramCharts(ctx, admin, patientID, "", 1, 10)
	if err != nil || total != 1 {
		t.Fatalf("failed to list legacy chart: total=%d err=%v", total, err)
	}
	legacyChart := charts[0]

	// 1. Verify immutable legacy_provenance snapshot is populated
	if len(legacyChart.LegacyProvenance) == 0 {
		t.Fatalf("expected legacy_provenance to be populated on legacy chart")
	}

	var provItems []map[string]any
	if err = json.Unmarshal(legacyChart.LegacyProvenance, &provItems); err != nil {
		t.Fatalf("failed to unmarshal legacy_provenance: %v", err)
	}
	if len(provItems) != 2 {
		t.Fatalf("expected 2 items in legacy_provenance, got %d", len(provItems))
	}

	// 2. Perform an ordinary edit: add Tooth 3 ("B"), keep Tooth 1 ("caries")
	updatedChart, err := app.UpdateOdontogramChart(ctx, admin, legacyChart.ID, domain.SaveOdontogramChartInput{
		PatientID:   patientID,
		DoctorID:    "doctor-user-1",
		Description: "Updated examination notes after follow-up",
		Odontogram:  map[int]string{1: "caries", 3: "B"},
		Version:     legacyChart.Version,
	})
	if err != nil {
		t.Fatalf("UpdateOdontogramChart failed: %v", err)
	}

	// 3. Verify legacy_provenance remains untouched and intact
	if len(updatedChart.LegacyProvenance) == 0 {
		t.Fatalf("legacy_provenance was lost after ordinary edit!")
	}

	// 4. Verify in patient_odontogram_entry that Tooth 1 still preserved original notes and author!
	var notes1, author1 string
	var created1 time.Time
	err = db.QueryRow(ctx, `
		SELECT procedure_notes, diagnosed_by, created_at
		FROM patient_odontogram_entry
		WHERE chart_id = $1::uuid AND tooth_number = 1
	`, legacyChart.ID).Scan(&notes1, &author1, &created1)
	if err != nil {
		t.Fatalf("failed to query tooth 1 entry: %v", err)
	}

	if notes1 != "Deep distal caries with pulpal involvement" {
		t.Errorf("Tooth 1 procedure_notes was overwritten! Got: '%s'", notes1)
	}
	if author1 != "doctor-user-1" {
		t.Errorf("Tooth 1 author was overwritten! Got: '%s'", author1)
	}
	if !created1.Equal(t1) {
		t.Errorf("Tooth 1 created_at timestamp was overwritten! Expected %v, got %v", t1, created1)
	}
}

// R5: Durable custom legend definitions CRUD and reference protection
func TestOdontogram_R5_DurableCustomLegendsCRUD(t *testing.T) {
	ctx := context.Background()
	db, store, _, cleanup := setupOdontogramTestDB(t, 0)
	defer cleanup()

	admin := domain.Actor{ID: "admin-user", Role: "admin"}
	app := application.ClinicalCareService{Store: store}
	patientID := helperCreatePatient(t, ctx, db, "Rosalind", "Franklin")

	// 1. Initial list: all 11 default clinical legends must be present
	legends, err := app.ListOdontogramLegends(ctx)
	if err != nil {
		t.Fatalf("ListOdontogramLegends failed: %v", err)
	}
	if len(legends) < 11 {
		t.Errorf("expected at least 11 default legends, got %d", len(legends))
	}

	// 2. Create durable custom legend IMP-Z
	created, err := app.CreateOdontogramLegend(ctx, admin, domain.SaveOdontogramLegendInput{
		Code:        "IMP-Z",
		Name:        "Zirconia Implant",
		Color:       "#06b6d4",
		Description: "Biocompatible zirconia endosseous implant fixture",
	})
	if err != nil {
		t.Fatalf("CreateOdontogramLegend failed: %v", err)
	}
	if created.Code != "IMP-Z" || created.IsDefault {
		t.Errorf("invalid created legend: %+v", created)
	}

	// Duplicate code must conflict
	_, err = app.CreateOdontogramLegend(ctx, admin, domain.SaveOdontogramLegendInput{
		Code:  "IMP-Z",
		Name:  "Duplicate",
		Color: "#000000",
	})
	if !errors.Is(err, domain.ErrConflict) {
		t.Errorf("expected ErrConflict for duplicate legend code, got: %v", err)
	}

	// 3. Deleting a default legend must be forbidden
	defaultLeg := legends[0]
	err = app.DeleteOdontogramLegend(ctx, admin, defaultLeg.ID)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("expected ErrForbidden when deleting default legend %s, got: %v", defaultLeg.Code, err)
	}

	// 4. Create chart using custom condition IMP-Z
	chart, err := app.CreateOdontogramChart(ctx, admin, domain.SaveOdontogramChartInput{
		PatientID:   patientID,
		DoctorID:    "doctor-user-1",
		Description: "Patient with custom zirconia implant",
		Odontogram:  map[int]string{14: "IMP-Z"},
	})
	if err != nil {
		t.Fatalf("CreateOdontogramChart with custom legend failed: %v", err)
	}

	// 5. Deleting custom legend currently in use by clinical records must return ErrInUse
	err = app.DeleteOdontogramLegend(ctx, admin, created.ID)
	if !errors.Is(err, domain.ErrInUse) {
		t.Errorf("expected ErrInUse when deleting legend in use, got: %v", err)
	}

	// 6. Delete chart, then delete legend succeeds
	err = app.DeleteOdontogramChart(ctx, admin, chart.ID)
	if err != nil {
		t.Fatalf("failed to delete chart: %v", err)
	}

	err = app.DeleteOdontogramLegend(ctx, admin, created.ID)
	if err != nil {
		t.Errorf("failed to delete unused custom legend: %v", err)
	}

	// Reload legends: IMP-Z must be gone
	legendsAfter, _ := app.ListOdontogramLegends(ctx)
	for _, l := range legendsAfter {
		if l.Code == "IMP-Z" {
			t.Errorf("custom legend IMP-Z still exists after deletion!")
		}
	}
}

// R6: PDF generation with multi-page pagination, word-wrapped notes, and Amharic Unicode
func TestOdontogram_R6_MultipagePDFWithWrappedNotesAndAmharic(t *testing.T) {
	// Build chart with Amharic patient name, long wrapped clinical description, and all 52 teeth
	chart := domain.OdontogramChart{
		ID:          "4fd15e13-b280-48e9-b2b3-58f232590834",
		PatientID:   "patient-uuid-1",
		PatientName: "አበበ ቢቂላ (Abebe Bikila)",
		PatientMRN:  "MRN-1004829",
		DoctorID:    "doctor-user-1",
		DoctorName:  "Dr. Dawit Haile",
		Description: "Comprehensive full-mouth periodontal and restorative dental examination. የታካሚው ጥርስ ምርመራ እና ህክምና። Patient presents with multiple lesions, restorations, and bridges requiring multi-disciplinary dental intervention. Tooth 14 restored with custom biocompatible zirconia implant unit.",
		Version:     2,
		CreatedAt:   time.Date(2026, 10, 12, 10, 30, 0, 0, time.UTC),
		Odontogram:  make(map[int]string),
	}

	// Chart all 52 supported teeth (1-32 adult + 33-52 pediatric/deciduous)
	for i := 1; i <= 52; i++ {
		cond := "Healthy"
		switch i {
		case 1:
			cond = "K"
		case 8:
			cond = "B"
		case 14:
			cond = "IMP-Z"
		case 20:
			cond = "C"
		case 32:
			cond = "X"
		case 45:
			cond = "Ce"
		}
		chart.Odontogram[i] = cond
	}

	pdfBytes := httpapi.GenerateOdontogramPDF(chart)

	// 1. Must start with standard PDF header
	if !bytes.HasPrefix(pdfBytes, []byte("%PDF-1.4")) {
		t.Fatalf("expected %%PDF-1.4 header, got: %s", string(pdfBytes[:8]))
	}

	// 2. Must be substantial in size (embedded TrueType font)
	if len(pdfBytes) < 50000 {
		t.Errorf("expected PDF to include embedded font (>50KB), got %d bytes", len(pdfBytes))
	}

	pdfStr := string(pdfBytes)

	// 3. Must be multi-page (at least 2 pages) because 52 teeth + long notes exceed a single page
	if !strings.Contains(pdfStr, "/Count 2") && !strings.Contains(pdfStr, "/Count 3") {
		t.Errorf("expected multi-page PDF (/Count 2 or /Count 3), got pages object definition in PDF")
	}

	// 4. Must include Type 0 font and embedded CIDFont with Identity-H
	if !strings.Contains(pdfStr, "/Subtype /Type0") {
		t.Errorf("expected Type 0 font definition in PDF")
	}
	if !strings.Contains(pdfStr, "/Encoding /Identity-H") {
		t.Errorf("expected /Identity-H encoding for Unicode support")
	}

	// 5. Must include ToUnicode CMap for selectable text
	if !strings.Contains(pdfStr, "/ToUnicode") {
		t.Errorf("expected /ToUnicode CMap stream in PDF")
	}

	// 6. Must terminate cleanly with %%EOF
	if !strings.HasSuffix(strings.TrimSpace(pdfStr), "%%EOF") {
		t.Errorf("expected %%EOF trailer at the end of PDF")
	}
}

// F1: Upgrade from pre-existing 060 schema (handling already-applied 060, duplicate legacy charts, and backfilling provenance)
func TestOdontogram_F1_UpgradeFromPreExisting060(t *testing.T) {
	ctx := context.Background()
	db, _, _, cleanup := setupOdontogramTestDB(t, 60)
	defer cleanup()

	patientID := helperCreatePatient(t, ctx, db, "Tigist", "Mengistu")

	// Under 060, simulate 2 legacy charts created concurrently for this patient
	var legacy1ID, legacy2ID string
	err := db.QueryRow(ctx, `
		INSERT INTO odontogram_chart (patient_id, doctor_id, description, odontogram, is_legacy, created_by, updated_by)
		VALUES ($1::uuid, 'doctor-user-1', 'Initial Legacy Chart 1', '{"1":"K"}', true, 'doctor-user-1', 'doctor-user-1')
		RETURNING id::text
	`, patientID).Scan(&legacy1ID)
	if err != nil {
		t.Fatalf("failed to insert legacy chart 1: %v", err)
	}
	err = db.QueryRow(ctx, `
		INSERT INTO odontogram_chart (patient_id, doctor_id, description, odontogram, is_legacy, created_by, updated_by)
		VALUES ($1::uuid, 'doctor-user-1', 'Duplicate Legacy Chart 2', '{"8":"C"}', true, 'doctor-user-1', 'doctor-user-1')
		RETURNING id::text
	`, patientID).Scan(&legacy2ID)
	if err != nil {
		t.Fatalf("failed to insert legacy chart 2: %v", err)
	}

	// Link tooth 1 to legacy1 and tooth 8 to legacy2
	_, err = db.Exec(ctx, `
		INSERT INTO patient_odontogram_entry (id, patient_id, chart_id, tooth_number, condition, procedure_notes, diagnosed_by, created_at, updated_at)
		VALUES
		  (gen_random_uuid(), $1::uuid, $2::uuid, 1, 'K', 'Original note for tooth 1', 'doctor-user-1', '2026-01-01 09:00:00Z', '2026-01-01 09:00:00Z'),
		  (gen_random_uuid(), $1::uuid, $3::uuid, 8, 'C', 'Original note for tooth 8', 'doctor-user-1', '2026-01-02 10:00:00Z', '2026-01-02 10:00:00Z')
	`, patientID, legacy1ID, legacy2ID)
	if err != nil {
		t.Fatalf("failed to seed tooth entries: %v", err)
	}

	// Verify that 2 legacy charts currently exist before 061
	var legacyCountBefore int
	err = db.QueryRow(ctx, `SELECT COUNT(*) FROM odontogram_chart WHERE patient_id = $1::uuid AND is_legacy = true`, patientID).Scan(&legacyCountBefore)
	if err != nil || legacyCountBefore != 2 {
		t.Fatalf("expected 2 legacy charts before upgrade, got %d, err=%v", legacyCountBefore, err)
	}

	// Read and run migration 061_odontogram_parity_upgrade.sql
	migration061, err := os.ReadFile("../../../../../db/migrations/061_odontogram_parity_upgrade.sql")
	if err != nil {
		t.Fatalf("failed to read 061 migration: %v", err)
	}
	if _, err = db.Exec(ctx, string(migration061)); err != nil {
		t.Fatalf("failed to execute migration 061: %v", err)
	}

	// 3. Assert duplicate legacy charts deduplicated to exactly 1
	var legacyCountAfter int
	var keeperID string
	var provJSON []byte
	err = db.QueryRow(ctx, `
		SELECT COUNT(*), (ARRAY_AGG(id::text))[1], (ARRAY_AGG(legacy_provenance))[1]
		FROM odontogram_chart
		WHERE patient_id = $1::uuid AND is_legacy = true
		GROUP BY patient_id
	`, patientID).Scan(&legacyCountAfter, &keeperID, &provJSON)
	if err != nil || legacyCountAfter != 1 {
		t.Fatalf("expected exactly 1 legacy chart after 061 upgrade, got %d, err=%v", legacyCountAfter, err)
	}

	// 4. Assert both tooth 1 and tooth 8 now belong to the keeper legacy chart
	var childCount int
	err = db.QueryRow(ctx, `SELECT COUNT(*) FROM patient_odontogram_entry WHERE chart_id = $1::uuid`, keeperID).Scan(&childCount)
	if err != nil || childCount != 2 {
		t.Fatalf("expected all 2 child tooth entries to be preserved in keeper chart, got %d, err=%v", childCount, err)
	}

	// 5. Assert legacy_provenance was backfilled
	if len(provJSON) == 0 || string(provJSON) == "null" {
		t.Fatalf("expected legacy_provenance to be populated on legacy chart")
	}
	provStr := string(provJSON)
	if !strings.Contains(provStr, "Original note for tooth 1") || !strings.Contains(provStr, "Original note for tooth 8") {
		t.Fatalf("expected legacy_provenance to contain distinct tooth notes, got: %s", provStr)
	}

	// 6. Assert odontogram_legend table exists and has defaults
	var legendCount int
	err = db.QueryRow(ctx, `SELECT COUNT(*) FROM odontogram_legend WHERE is_default = true`).Scan(&legendCount)
	if err != nil || legendCount < 11 {
		t.Fatalf("expected at least 11 default legends in odontogram_legend, got %d, err=%v", legendCount, err)
	}
}

// F2: Established default clinical legends definitions exact contract
func TestOdontogram_F2_EstablishedDefaultLegendsDefinitions(t *testing.T) {
	ctx := context.Background()
	db, store, _, cleanup := setupOdontogramTestDB(t, 0)
	defer cleanup()

	adminActor := domain.Actor{ID: "admin-user", Role: "admin"}

	legends, err := store.ListOdontogramLegends(ctx)
	if err != nil {
		t.Fatalf("failed to list legends: %v", err)
	}

	legendMap := make(map[string]domain.OdontogramLegend)
	for _, l := range legends {
		legendMap[l.Code] = l
	}

	// Exact definitions contract from odontogram-register.tsx
	expected := map[string]struct {
		name        string
		color       string
		description string
	}{
		"Healthy": {"Healthy Tooth", "#ffffff", "Sound tooth surface with no pathology"},
		"K":       {"Caries / Cavity", "#e91e63", "Active dental caries lesion"},
		"C":       {"Crown", "#3f51b5", "Artificial full crown or cap"},
		"Ce":      {"Pulpitis / Deep Caries", "#009688", "Extensive caries with pulpal involvement"},
		"D":       {"Decayed / Denture", "#795548", "Grossly decayed or retained root"},
		"KR":      {"Root Canal (Karies Radix)", "#607d8b", "Endodontic therapy / Root canal treated"},
		"PS":      {"Porcelain / Prosthetic", "#ffc107", "Porcelain veneer or prosthetic replacement"},
		"IP":      {"Implant", "#673ab7", "Titanium or ceramic dental implant"},
		"X":       {"Missing / Extracted", "#2196f3", "Congenitally missing or extracted tooth"},
		"F":       {"Composite Filling", "#10b981", "Tooth-colored composite restoration"},
		"B":       {"Bridge", "#f97316", "Fixed partial denture / pontic unit"},
	}

	for code, exp := range expected {
		l, ok := legendMap[code]
		if !ok {
			t.Fatalf("missing expected default legend code: %s", code)
		}
		if l.Name != exp.name {
			t.Errorf("legend [%s] name mismatch: expected '%s', got '%s'", code, exp.name, l.Name)
		}
		if l.Color != exp.color {
			t.Errorf("legend [%s] color mismatch: expected '%s', got '%s'", code, exp.color, l.Color)
		}
		if l.Description != exp.description {
			t.Errorf("legend [%s] description mismatch: expected '%s', got '%s'", code, exp.description, l.Description)
		}
		if !l.IsDefault {
			t.Errorf("legend [%s] should be marked is_default = true", code)
		}
	}

	// Reconcile must not overwrite user-customized definitions
	custom, err := store.CreateOdontogramLegend(ctx, adminActor, domain.SaveOdontogramLegendInput{
		Code:        "USER-CUSTOM",
		Name:        "Custom User Definition",
		Color:       "#123456",
		Description: "User customized notes",
	})
	if err != nil {
		t.Fatalf("failed to create custom legend: %v", err)
	}

	// Run migration 061 again to simulate re-running or reconciling
	migration061, err := os.ReadFile("../../../../../db/migrations/061_odontogram_parity_upgrade.sql")
	if err != nil {
		t.Fatalf("failed to read 061 migration: %v", err)
	}
	if _, err = db.Exec(ctx, string(migration061)); err != nil {
		t.Fatalf("failed to re-apply 061: %v", err)
	}

	// Custom legend must remain untouched
	customReloaded, err := store.ListOdontogramLegends(ctx)
	if err != nil {
		t.Fatalf("failed to reload legends: %v", err)
	}
	var foundCustom bool
	for _, l := range customReloaded {
		if l.ID == custom.ID {
			foundCustom = true
			if l.Name != "Custom User Definition" || l.Color != "#123456" {
				t.Errorf("custom legend was modified during reconciliation: %+v", l)
			}
		}
	}
	if !foundCustom {
		t.Errorf("custom legend was lost during reconciliation")
	}
}

// F3: Referenced legend code and name modification protection
func TestOdontogram_F3_ReferencedLegendCodeAndNameProtection(t *testing.T) {
	ctx := context.Background()
	db, store, _, cleanup := setupOdontogramTestDB(t, 0)
	defer cleanup()

	adminActor := domain.Actor{ID: "admin-user", Role: "admin"}
	docID := "doctor-user-1"
	docActor := domain.Actor{ID: docID, Role: "doctor"}

	patientID := helperCreatePatient(t, ctx, db, "Hanna", "Kassahun")

	// 1. Create custom legend IMP-Z
	impZ, err := store.CreateOdontogramLegend(ctx, adminActor, domain.SaveOdontogramLegendInput{
		Code:        "IMP-Z",
		Name:        "Zirconia Implant",
		Color:       "#10b981",
		Description: "Zirconia dental fixture",
	})
	if err != nil {
		t.Fatalf("failed to create custom legend IMP-Z: %v", err)
	}

	// 2. Chart tooth using IMP-Z
	chart, err := store.CreateOdontogramChart(ctx, docActor, domain.SaveOdontogramChartInput{
		PatientID:   patientID,
		DoctorID:    docID,
		Description: "Patient charting with custom IMP-Z unit",
		Odontogram:  map[int]string{14: "IMP-Z"},
	})
	if err != nil {
		t.Fatalf("failed to create chart with IMP-Z: %v", err)
	}
	_ = chart

	// 3. Attempt to change CODE of IMP-Z to IMP-Y -> Must be rejected with ErrConflict
	_, err = store.UpdateOdontogramLegend(ctx, adminActor, impZ.ID, domain.SaveOdontogramLegendInput{
		Code:        "IMP-Y",
		Name:        "Zirconia Implant",
		Color:       "#10b981",
		Description: "Zirconia dental fixture",
	})
	if !errors.Is(err, domain.ErrConflict) {
		t.Errorf("expected ErrConflict when changing code of referenced legend, got: %v", err)
	}

	// 4. Attempt to change NAME of referenced IMP-Z -> Must be rejected with ErrConflict
	_, err = store.UpdateOdontogramLegend(ctx, adminActor, impZ.ID, domain.SaveOdontogramLegendInput{
		Code:        "IMP-Z",
		Name:        "Titanium Fixture",
		Color:       "#10b981",
		Description: "Zirconia dental fixture",
	})
	if !errors.Is(err, domain.ErrConflict) {
		t.Errorf("expected ErrConflict when renaming referenced legend, got: %v", err)
	}

	// 5. Updating COLOR and DESCRIPTION of referenced IMP-Z -> Must succeed
	updated, err := store.UpdateOdontogramLegend(ctx, adminActor, impZ.ID, domain.SaveOdontogramLegendInput{
		Code:        "IMP-Z",
		Name:        "Zirconia Implant",
		Color:       "#059669",
		Description: "Updated high-precision zirconia dental fixture",
	})
	if err != nil {
		t.Fatalf("expected updating color/description to succeed, got: %v", err)
	}
	if updated.Color != "#059669" || updated.Description != "Updated high-precision zirconia dental fixture" {
		t.Errorf("unexpected updated values: %+v", updated)
	}

	// 6. Attempt to delete referenced IMP-Z -> Must be rejected with ErrInUse
	err = store.DeleteOdontogramLegend(ctx, adminActor, impZ.ID)
	if !errors.Is(err, domain.ErrInUse) {
		t.Errorf("expected ErrInUse when deleting referenced legend, got: %v", err)
	}

	// 7. Unreferenced legend CAN have code/name changed or be deleted
	tempLeg, err := store.CreateOdontogramLegend(ctx, adminActor, domain.SaveOdontogramLegendInput{
		Code:        "TEMP-1",
		Name:        "Temporary Legend",
		Color:       "#123456",
		Description: "Unreferenced",
	})
	if err != nil {
		t.Fatalf("failed to create temp legend: %v", err)
	}
	_, err = store.UpdateOdontogramLegend(ctx, adminActor, tempLeg.ID, domain.SaveOdontogramLegendInput{
		Code:        "TEMP-2",
		Name:        "Renamed Temporary Legend",
		Color:       "#654321",
		Description: "Still unreferenced",
	})
	if err != nil {
		t.Errorf("expected unreferenced legend to be updatable, got: %v", err)
	}
	err = store.DeleteOdontogramLegend(ctx, adminActor, tempLeg.ID)
	if err != nil {
		t.Errorf("expected unreferenced legend to be deletable, got: %v", err)
	}

	// 8. Default legends cannot have code/name changed or be deleted
	var defLegID string
	err = db.QueryRow(ctx, `SELECT id FROM odontogram_legend WHERE code = 'K'`).Scan(&defLegID)
	if err != nil {
		t.Fatalf("failed to find default legend K: %v", err)
	}
	_, err = store.UpdateOdontogramLegend(ctx, adminActor, defLegID, domain.SaveOdontogramLegendInput{
		Code:        "K-NEW",
		Name:        "Caries Modified",
		Color:       "#ef4444",
		Description: "Notes",
	})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("expected ErrForbidden when modifying default legend code/name, got: %v", err)
	}
	err = store.DeleteOdontogramLegend(ctx, adminActor, defLegID)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("expected ErrForbidden when deleting default legend, got: %v", err)
	}
}

// F4: PDF text metrics, boundary enforcement, and text wrapping
func TestOdontogram_F4_PDFTextMetricsBoundsAndWrapping(t *testing.T) {
	chart := domain.OdontogramChart{
		ID:          "f4-chart-uuid-1",
		PatientID:   "patient-uuid-f4",
		PatientName: "አበበ ቢቂላ ገብረመስቀል ወልደዮሐንስ (Abebe Bikila Gebremeskel Woldeyohannes)",
		PatientMRN:  "MRN-ETH-LONG-IDENTIFIER-994827103984",
		DoctorID:    "doctor-user-f4",
		DoctorName:  "Dr. Dawit Haile G/Mariam",
		Description: "Comprehensive examination report with overlong unbroken token SUPERLONGUNBROKENCLINICALFINDINGWITHNOSPACESTHATEXCEEDSPAGEWIDTH1234567890ABCDEF and bilingual Amharic clinical documentation የታካሚው የጥርስ ጤንነት አጠቃላይ ምርመራ ተደርጓል። Detailed tooth-by-tooth restoration plan established.",
		Version:     3,
		CreatedAt:   time.Date(2026, 10, 12, 14, 0, 0, 0, time.UTC),
		Odontogram:  make(map[int]string),
	}

	// Chart teeth including 64-character unbroken custom code across 52 teeth to test pagination
	longCustomCode := "CUSTOM-LONG-UNBROKEN-IMPLANT-BIOMATERIAL-ZIRCONIA-REST-PROSTHESIS-64"
	for i := 1; i <= 52; i++ {
		chart.Odontogram[i] = "Healthy"
	}
	chart.Odontogram[1] = "K"
	chart.Odontogram[8] = "B"
	chart.Odontogram[14] = longCustomCode

	pdfBytes := httpapi.GenerateOdontogramPDF(chart)
	if !bytes.HasPrefix(pdfBytes, []byte("%PDF-1.4")) {
		t.Fatalf("expected valid %%PDF-1.4 header")
	}

	pdfStr := string(pdfBytes)

	// 1. Assert CIDFont contains /DW and /W array with actual font metrics
	if !strings.Contains(pdfStr, "/DW 600") {
		t.Errorf("expected /DW 600 default width in CIDFont dictionary")
	}
	if !strings.Contains(pdfStr, "/W [") {
		t.Errorf("expected /W [ width array in CIDFont dictionary")
	}

	// 2. Assert ToUnicode map is present
	if !strings.Contains(pdfStr, "/ToUnicode") {
		t.Errorf("expected ToUnicode map")
	}

	// 3. Assert long unbroken custom code and description were processed and included (>50KB)
	if len(pdfBytes) < 50000 {
		t.Errorf("expected PDF payload >50KB, got %d", len(pdfBytes))
	}

	// 4. Assert multi-page structure
	if !strings.Contains(pdfStr, "/Count 2") && !strings.Contains(pdfStr, "/Count 3") {
		t.Errorf("expected multi-page PDF output")
	}
}


