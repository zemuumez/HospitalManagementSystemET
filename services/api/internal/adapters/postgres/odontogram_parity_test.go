package postgres

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
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
