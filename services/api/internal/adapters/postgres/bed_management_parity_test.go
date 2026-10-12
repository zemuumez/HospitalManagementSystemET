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
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"hms.local/api/internal/domain"
)

func setupBedTestDB(t *testing.T, throughMigration int) (*pgxpool.Pool, Store, string, func()) {
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
	schema := "hms_test_bed_" + hex.EncodeToString(random)
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

	store := Store{DB: db}
	cleanup := func() {
		db.Close()
		_, _ = adminDB.Exec(ctx, "DROP SCHEMA IF EXISTS "+quoted+" CASCADE")
		adminDB.Close()
	}

	return db, store, schema, cleanup
}

func seedBedStaffAndPatient(ctx context.Context, t *testing.T, db *pgxpool.Pool) (domain.Actor, domain.Actor, domain.Actor, string, string, string) {
	t.Helper()
	admin := domain.Actor{ID: "usr-admin-bed", Role: "admin"}
	doctor := domain.Actor{ID: "usr-doc-bed", Role: "doctor"}
	nurse := domain.Actor{ID: "usr-nurse-bed", Role: "nurse"}

	_, err := db.Exec(ctx, `
		INSERT INTO "user"(id, email, name)
		VALUES
			($1, 'admin_bed@hospital.test', 'Admin Bed'),
			($2, 'doc_bed@hospital.test', 'Doctor Bed'),
			($3, 'nurse_bed@hospital.test', 'Nurse Bed')
		ON CONFLICT(id) DO NOTHING
	`, admin.ID, doctor.ID, nurse.ID)
	if err != nil {
		t.Fatalf("failed to insert users: %v", err)
	}

	_, err = db.Exec(ctx, `
		INSERT INTO staff_access(user_id, role, active)
		VALUES 
			($1, 'admin', true),
			($2, 'doctor', true),
			($3, 'nurse', true)
		ON CONFLICT(user_id) DO NOTHING
	`, admin.ID, doctor.ID, nurse.ID)
	if err != nil {
		t.Fatalf("failed to insert staff_access: %v", err)
	}

	var patientID, encounterID, caseID string
	err = db.QueryRow(ctx, `
		INSERT INTO patient(given_name, family_name, date_of_birth, phone)
		VALUES ('Abebe', 'Bikila', '1990-01-01', '+251911223344')
		RETURNING id::text
	`).Scan(&patientID)
	if err != nil {
		t.Fatalf("failed to insert patient: %v", err)
	}
	_, _ = db.Exec(ctx, `UPDATE patient_profile SET email = 'abebe@hospital.test', gender = 'Male' WHERE patient_id = $1`, patientID)

	err = db.QueryRow(ctx, `
		INSERT INTO patient_case(patient_id, doctor_id, description, created_by)
		VALUES ($1, $2, 'Inpatient general case', $3)
		RETURNING id::text
	`, patientID, doctor.ID, admin.ID).Scan(&caseID)
	if err != nil {
		t.Fatalf("failed to insert patient_case: %v", err)
	}

	err = db.QueryRow(ctx, `
		INSERT INTO encounter(kind, case_id, patient_id, doctor_id, admitted_at, status, symptoms, created_by, request_key)
		VALUES ('ipd', $1, $2, $3, clock_timestamp(), 'active', 'Chest pain', $4, 'req-key-bed-test-1')
		RETURNING id::text
	`, caseID, patientID, doctor.ID, admin.ID).Scan(&encounterID)
	if err != nil {
		t.Fatalf("failed to insert encounter: %v", err)
	}

	return admin, doctor, nurse, patientID, encounterID, caseID
}

// Test 1: Bed Type & Bed CRUD, unique constraints, and in-use deletion protection
func TestBed_CRUDAndUniqueConstraints(t *testing.T) {
	_, store, _, cleanup := setupBedTestDB(t, 0)
	defer cleanup()
	ctx := context.Background()

	admin, _, _, _, _, _ := seedBedStaffAndPatient(ctx, t, store.DB)

	// Create Bed Type
	bt1, err := store.SaveBedType(ctx, admin, "", domain.BedTypeInput{
		Name:        "Intensive Care Unit",
		Description: "ICU level-3 monitoring",
		Active:      true,
	})
	if err != nil {
		t.Fatalf("failed to create bed type: %v", err)
	}
	if bt1.ID == "" || bt1.Name != "Intensive Care Unit" {
		t.Fatalf("unexpected bed type: %+v", bt1)
	}

	// Unique Bed Type name constraint
	_, err = store.SaveBedType(ctx, admin, "", domain.BedTypeInput{
		Name:        "Intensive Care Unit",
		Description: "Duplicate name",
		Active:      true,
	})
	if err == nil {
		t.Fatal("expected error on duplicate bed type name, got nil")
	}

	// Create Bed
	bed1, err := store.CreateBed(ctx, admin, domain.BedInput{
		TypeID:      bt1.ID,
		Name:        "ICU-101",
		WardName:    "ICU Ward",
		ChargeMinor: 150000, // 1,500.00 ETB
		Description: "Bed next to vital monitors",
	})
	if err != nil {
		t.Fatalf("failed to create bed: %v", err)
	}
	if bed1.ID == "" || bed1.Name != "ICU-101" || bed1.ChargeMinor != 150000 || bed1.WardName != "ICU Ward" {
		t.Fatalf("unexpected bed: %+v", bed1)
	}

	// Unique Bed name constraint
	_, err = store.CreateBed(ctx, admin, domain.BedInput{
		TypeID:      bt1.ID,
		Name:        "ICU-101",
		WardName:    "ICU Ward",
		ChargeMinor: 200000,
	})
	if err == nil {
		t.Fatal("expected error on duplicate bed name, got nil")
	}

	// In-use deletion protection for Bed Type
	err = store.DeleteBedType(ctx, admin, bt1.ID, bt1.Version)
	if !errors.Is(err, domain.ErrInUse) {
		t.Fatalf("expected ErrInUse when deleting bed type with existing beds, got: %v", err)
	}

	// Update Bed
	updatedBed, err := store.UpdateBed(ctx, admin, bed1.ID, domain.UpdateBedInput{
		TypeID:      bt1.ID,
		Name:        "ICU-101A",
		WardName:    "ICU Critical",
		ChargeMinor: 175000,
		Description: "Upgraded ICU bed",
		Version:     bed1.Version,
	})
	if err != nil {
		t.Fatalf("failed to update bed: %v", err)
	}
	if updatedBed.Name != "ICU-101A" || updatedBed.ChargeMinor != 175000 || updatedBed.WardName != "ICU Critical" {
		t.Fatalf("unexpected updated bed: %+v", updatedBed)
	}

	// Stale edit rejection on bed update
	_, err = store.UpdateBed(ctx, admin, bed1.ID, domain.UpdateBedInput{
		TypeID:      bt1.ID,
		Name:        "ICU-101B",
		WardName:    "ICU Critical",
		ChargeMinor: 180000,
		Version:     bed1.Version, // Stale version!
	})
	if !errors.Is(err, domain.ErrStale) {
		t.Fatalf("expected ErrStale on stale version edit, got: %v", err)
	}
}

// Test 2: Concurrent double-assignment race condition protection
func TestBed_ConcurrentAssignmentRace(t *testing.T) {
	_, store, _, cleanup := setupBedTestDB(t, 0)
	defer cleanup()
	ctx := context.Background()

	admin, doctor, _, patient1ID, enc1ID, _ := seedBedStaffAndPatient(ctx, t, store.DB)

	// Create second patient and second inpatient encounter
	var patient2ID, enc2ID string
	err := store.DB.QueryRow(ctx, `
		INSERT INTO patient(given_name, family_name, date_of_birth, phone)
		VALUES ('Tirunesh', 'Dibaba', '1992-05-05', '+251922334455')
		RETURNING id::text
	`).Scan(&patient2ID)
	if err != nil {
		t.Fatalf("failed to insert second patient: %v", err)
	}
	var case2ID string
	err = store.DB.QueryRow(ctx, `
		INSERT INTO patient_case(patient_id, doctor_id, description, created_by)
		VALUES ($1, $2, 'Inpatient general case 2', $3)
		RETURNING id::text
	`, patient2ID, doctor.ID, admin.ID).Scan(&case2ID)
	if err != nil {
		t.Fatalf("failed to insert second patient_case: %v", err)
	}

	err = store.DB.QueryRow(ctx, `
		INSERT INTO encounter(kind, case_id, patient_id, doctor_id, admitted_at, status, symptoms, created_by, request_key)
		VALUES ('ipd', $1, $2, $3, clock_timestamp(), 'active', 'High fever', $4, 'req-key-bed-test-2')
		RETURNING id::text
	`, case2ID, patient2ID, doctor.ID, admin.ID).Scan(&enc2ID)
	if err != nil {
		t.Fatalf("failed to insert second encounter: %v", err)
	}

	bt, err := store.SaveBedType(ctx, admin, "", domain.BedTypeInput{Name: "General Bed Type", Active: true})
	if err != nil {
		t.Fatalf("failed to create bed type: %v", err)
	}

	bed, err := store.CreateBed(ctx, admin, domain.BedInput{
		TypeID:      bt.ID,
		Name:        "GEN-01",
		WardName:    "General Ward",
		ChargeMinor: 50000,
	})
	if err != nil {
		t.Fatalf("failed to create bed: %v", err)
	}

	// Execute concurrent assignment race
	concurrency := 2
	var wg sync.WaitGroup
	wg.Add(concurrency)

	results := make([]error, concurrency)

	assignInputs := []domain.AssignBedInput{
		{BedID: bed.ID, EncounterID: enc1ID, PatientID: patient1ID, Notes: "Patient 1 admission"},
		{BedID: bed.ID, EncounterID: enc2ID, PatientID: patient2ID, Notes: "Patient 2 admission"},
	}

	startGate := make(chan struct{})

	for i := 0; i < concurrency; i++ {
		idx := i
		go func() {
			defer wg.Done()
			<-startGate
			_, err := store.AssignBed(ctx, admin, assignInputs[idx])
			results[idx] = err
		}()
	}

	close(startGate)
	wg.Wait()

	successCount := 0
	conflictCount := 0

	for _, res := range results {
		if res == nil {
			successCount++
		} else if errors.Is(res, domain.ErrConflict) || errors.Is(res, domain.ErrStale) {
			conflictCount++
		} else {
			t.Errorf("unexpected error in concurrent assignment: %v", res)
		}
	}

	if successCount != 1 || conflictCount != 1 {
		t.Fatalf("expected exactly 1 success and 1 conflict in concurrent assignment race, got %d successes and %d conflicts", successCount, conflictCount)
	}

	// Verify database record has exactly 1 active assignment
	var activeCount int
	err = store.DB.QueryRow(ctx, `SELECT COUNT(*) FROM bed_assignment WHERE bed_id = $1 AND assigned_to IS NULL`, bed.ID).Scan(&activeCount)
	if err != nil {
		t.Fatalf("failed to query active assignments: %v", err)
	}
	if activeCount != 1 {
		t.Fatalf("expected exactly 1 active assignment in database, got: %d", activeCount)
	}
}

// Test 3: Encounter ownership validation server-side
func TestBed_EncounterOwnershipValidation(t *testing.T) {
	_, store, _, cleanup := setupBedTestDB(t, 0)
	defer cleanup()
	ctx := context.Background()

	admin, _, _, patient1ID, enc1ID, _ := seedBedStaffAndPatient(ctx, t, store.DB)

	// Create second patient
	var patient2ID string
	err := store.DB.QueryRow(ctx, `
		INSERT INTO patient(given_name, family_name, date_of_birth, phone)
		VALUES ('Kenenisa', 'Bekele', '1982-06-13', '+251933445566')
		RETURNING id::text
	`).Scan(&patient2ID)
	if err != nil {
		t.Fatalf("failed to insert second patient: %v", err)
	}
	_, _ = store.DB.Exec(ctx, `UPDATE patient_profile SET email = 'kenenisa@hospital.test', gender = 'Male' WHERE patient_id = $1`, patient2ID)

	bt, _ := store.SaveBedType(ctx, admin, "", domain.BedTypeInput{Name: "Type Ownership", Active: true})
	bed, _ := store.CreateBed(ctx, admin, domain.BedInput{TypeID: bt.ID, Name: "OWN-01", WardName: "General", ChargeMinor: 50000})

	// Attempt assignment of Encounter 1 with Patient 2 ID (mismatched ownership)
	_, err = store.AssignBed(ctx, admin, domain.AssignBedInput{
		BedID:       bed.ID,
		EncounterID: enc1ID,
		PatientID:   patient2ID, // WRONG patient!
		Notes:       "Forged ownership test",
	})

	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected ErrValidation when encounter does not belong to patient, got: %v", err)
	}

	// Correct ownership succeeds
	assign, err := store.AssignBed(ctx, admin, domain.AssignBedInput{
		BedID:       bed.ID,
		EncounterID: enc1ID,
		PatientID:   patient1ID, // Correct patient!
		Notes:       "Genuine ownership",
	})
	if err != nil {
		t.Fatalf("expected valid assignment to succeed, got: %v", err)
	}
	if assign.ID == "" || assign.PatientID != patient1ID || assign.BedID != bed.ID {
		t.Fatalf("unexpected assignment result: %+v", assign)
	}
}

// Test 4: Maintenance state workflow and assignment prevention
func TestBed_MaintenanceStateWorkflow(t *testing.T) {
	_, store, _, cleanup := setupBedTestDB(t, 0)
	defer cleanup()
	ctx := context.Background()

	admin, _, _, patientID, encID, _ := seedBedStaffAndPatient(ctx, t, store.DB)

	bt, _ := store.SaveBedType(ctx, admin, "", domain.BedTypeInput{Name: "Maint Type", Active: true})
	bed, _ := store.CreateBed(ctx, admin, domain.BedInput{TypeID: bt.ID, Name: "MAINT-01", WardName: "General", ChargeMinor: 60000})

	// Put bed into maintenance
	maintBed, err := store.SetBedState(ctx, admin, bed.ID, domain.BedStateInput{
		State:   "maintenance",
		Version: bed.Version,
		Reason:  "Motor malfunction repair",
	})
	if err != nil {
		t.Fatalf("failed to set bed to maintenance: %v", err)
	}
	if maintBed.State != "maintenance" || maintBed.Available {
		t.Fatalf("expected maintenance bed to be unavailable, got: %+v", maintBed)
	}

	// Attempt to assign bed in maintenance -> must fail with conflict
	_, err = store.AssignBed(ctx, admin, domain.AssignBedInput{
		BedID:       bed.ID,
		EncounterID: encID,
		PatientID:   patientID,
		Notes:       "Should fail on maintenance bed",
	})
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("expected ErrConflict when assigning bed under maintenance, got: %v", err)
	}

	// Restore bed to ready
	readyBed, err := store.SetBedState(ctx, admin, bed.ID, domain.BedStateInput{
		State:   "ready",
		Version: maintBed.Version,
		Reason:  "Repairs completed",
	})
	if err != nil {
		t.Fatalf("failed to restore bed to ready: %v", err)
	}
	if readyBed.State != "ready" || !readyBed.Available {
		t.Fatalf("expected ready bed to be available, got: %+v", readyBed)
	}

	// Now assignment succeeds
	_, err = store.AssignBed(ctx, admin, domain.AssignBedInput{
		BedID:       bed.ID,
		EncounterID: encID,
		PatientID:   patientID,
		Notes:       "Assignment after repair",
	})
	if err != nil {
		t.Fatalf("assignment failed on ready bed: %v", err)
	}

	// Attempting to put an OCCUPIED bed into maintenance must fail with conflict
	_, err = store.SetBedState(ctx, admin, bed.ID, domain.BedStateInput{
		State:   "maintenance",
		Version: readyBed.Version,
		Reason:  "Cannot service occupied bed",
	})
	if !errors.Is(err, domain.ErrConflict) && !errors.Is(err, domain.ErrStale) {
		t.Fatalf("expected conflict when setting occupied bed to maintenance, got: %v", err)
	}
}

// Test 5: Transfer atomicity (releasing source bed and reserving target bed)
func TestBed_TransferAtomicity(t *testing.T) {
	_, store, _, cleanup := setupBedTestDB(t, 0)
	defer cleanup()
	ctx := context.Background()

	admin, _, _, patientID, encID, _ := seedBedStaffAndPatient(ctx, t, store.DB)

	bt, _ := store.SaveBedType(ctx, admin, "", domain.BedTypeInput{Name: "Transfer Ward Type", Active: true})
	bedA, _ := store.CreateBed(ctx, admin, domain.BedInput{TypeID: bt.ID, Name: "BED-A", WardName: "Ward 1", ChargeMinor: 50000})
	bedB, _ := store.CreateBed(ctx, admin, domain.BedInput{TypeID: bt.ID, Name: "BED-B", WardName: "Ward 2", ChargeMinor: 80000})

	// Initial assignment to Bed A
	assignA, err := store.AssignBed(ctx, admin, domain.AssignBedInput{
		BedID:       bedA.ID,
		EncounterID: encID,
		PatientID:   patientID,
		Notes:       "Initial Bed A",
	})
	if err != nil {
		t.Fatalf("failed initial assignment to Bed A: %v", err)
	}

	// Verify occupancy report: 1 occupied, 1 available
	rep1, err := store.BedOccupancyReport(ctx)
	if err != nil {
		t.Fatalf("failed to get occupancy report: %v", err)
	}
	if rep1.OccupiedBeds != 1 || rep1.AvailableBeds != 1 {
		t.Fatalf("unexpected occupancy before transfer: %+v", rep1)
	}

	// Transfer patient to Bed B
	assignB, err := store.AssignBed(ctx, admin, domain.AssignBedInput{
		BedID:       bedB.ID,
		EncounterID: encID,
		PatientID:   patientID,
		Notes:       "Transfer from Bed A to Bed B",
	})
	if err != nil {
		t.Fatalf("failed transfer to Bed B: %v", err)
	}
	if assignB.BedID != bedB.ID {
		t.Fatalf("expected transfer assignment bed to be Bed B, got: %+v", assignB)
	}

	// Verify Bed A assignment now has assigned_to timestamp
	var assignATo *time.Time
	err = store.DB.QueryRow(ctx, `SELECT assigned_to FROM bed_assignment WHERE id = $1`, assignA.ID).Scan(&assignATo)
	if err != nil {
		t.Fatalf("failed to query assign A: %v", err)
	}
	if assignATo == nil {
		t.Fatal("expected Bed A assignment to have assigned_to set upon transfer, got nil")
	}

	// Verify Bed A is now available in listing
	beds, total, err := store.SearchBeds(ctx, "", "", "", nil, 1, 10)
	if err != nil {
		t.Fatalf("failed to list beds: %v", err)
	}
	if total != 2 {
		t.Fatalf("expected 2 total beds, got: %d", total)
	}

	for _, b := range beds {
		if b.ID == bedA.ID && !b.Available {
			t.Fatalf("expected Bed A to be available after transfer, got unavailable: %+v", b)
		}
		if b.ID == bedB.ID && b.Available {
			t.Fatalf("expected Bed B to be occupied after transfer, got available: %+v", b)
		}
	}
}

// Test 6: Discharge lifecycle (releasing bed, updating report and status)
func TestBed_DischargeLifecycle(t *testing.T) {
	_, store, _, cleanup := setupBedTestDB(t, 0)
	defer cleanup()
	ctx := context.Background()

	admin, _, _, patientID, encID, _ := seedBedStaffAndPatient(ctx, t, store.DB)

	bt, _ := store.SaveBedType(ctx, admin, "", domain.BedTypeInput{Name: "Discharge Type", Active: true})
	bed, _ := store.CreateBed(ctx, admin, domain.BedInput{TypeID: bt.ID, Name: "DISC-01", WardName: "General", ChargeMinor: 60000})

	assign, err := store.AssignBed(ctx, admin, domain.AssignBedInput{
		BedID:       bed.ID,
		EncounterID: encID,
		PatientID:   patientID,
		Notes:       "Admitted for observation",
	})
	if err != nil {
		t.Fatalf("failed to assign bed: %v", err)
	}

	repBefore, _ := store.BedOccupancyReport(ctx)
	if repBefore.OccupiedBeds != 1 {
		t.Fatalf("expected 1 occupied bed before discharge, got: %d", repBefore.OccupiedBeds)
	}

	// Discharge bed assignment
	dischargedAssign, err := store.DischargeBedAssignment(ctx, admin, assign.ID, "Patient stabilized, discharged home")
	if err != nil {
		t.Fatalf("failed to discharge bed assignment: %v", err)
	}
	if dischargedAssign.Status {
		t.Fatalf("expected discharged assignment status to be false (inactive), got true")
	}
	if dischargedAssign.AssignedTo == nil {
		t.Fatal("expected discharged assignment to have assignedTo timestamp, got nil")
	}

	// Second discharge attempt must fail with conflict
	_, err = store.DischargeBedAssignment(ctx, admin, assign.ID, "Duplicate discharge attempt")
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("expected ErrConflict on duplicate discharge, got: %v", err)
	}

	// Bed occupancy report must reflect 0 occupied and 1 available
	repAfter, _ := store.BedOccupancyReport(ctx)
	if repAfter.OccupiedBeds != 0 || repAfter.AvailableBeds != 1 {
		t.Fatalf("expected 0 occupied and 1 available after discharge, got: %+v", repAfter)
	}

	// In-use delete protection:
	// Since bed has historical assignment, deleting it soft-deletes (active = false)
	err = store.DeleteBed(ctx, admin, bed.ID)
	if err != nil {
		t.Fatalf("expected soft-delete of historical bed to succeed, got: %v", err)
	}

	var isActive bool
	err = store.DB.QueryRow(ctx, `SELECT active FROM hospital_bed WHERE id = $1`, bed.ID).Scan(&isActive)
	if err != nil {
		t.Fatalf("failed to query bed active status: %v", err)
	}
	if isActive {
		t.Fatal("expected historical bed to be soft-deleted (active = false), got active = true")
	}
}

// Test 7: Populated Upgrade Safety Test Across Migration 063
func TestBed_PopulatedUpgradeSafety(t *testing.T) {
	// Setup database only through migration 062
	db, store, schema, cleanup := setupBedTestDB(t, 62)
	defer cleanup()
	ctx := context.Background()

	admin := domain.Actor{ID: "usr-admin-bed", Role: "admin"}
	doctor := domain.Actor{ID: "usr-doc-bed", Role: "doctor"}

	_, err := db.Exec(ctx, `
		INSERT INTO "user"(id, email, name)
		VALUES
			($1, 'admin_bed@hospital.test', 'Admin Bed'),
			($2, 'doc_bed@hospital.test', 'Doctor Bed')
		ON CONFLICT(id) DO NOTHING
	`, admin.ID, doctor.ID)
	if err != nil {
		t.Fatalf("failed to insert users: %v", err)
	}

	_, err = db.Exec(ctx, `
		INSERT INTO staff_access(user_id, role, active)
		VALUES 
			($1, 'admin', true),
			($2, 'doctor', true)
		ON CONFLICT(user_id) DO NOTHING
	`, admin.ID, doctor.ID)
	if err != nil {
		t.Fatalf("failed to insert staff_access: %v", err)
	}

	var patientID, caseID, encID string
	err = db.QueryRow(ctx, `
		INSERT INTO patient(given_name, family_name, date_of_birth, phone)
		VALUES ('Abebe', 'Bikila', '1990-01-01', '+251911223344')
		RETURNING id::text
	`).Scan(&patientID)
	if err != nil {
		t.Fatalf("failed to insert patient: %v", err)
	}

	err = db.QueryRow(ctx, `
		INSERT INTO patient_case(patient_id, doctor_id, description, created_by)
		VALUES ($1, $2, 'Inpatient general case', $3)
		RETURNING id::text
	`, patientID, doctor.ID, admin.ID).Scan(&caseID)
	if err != nil {
		t.Fatalf("failed to insert patient_case: %v", err)
	}

	// Populate historical data under 062 schema
	var typeID, bedID, assignID string
	err = db.QueryRow(ctx, `
		INSERT INTO bed_type(name, description, active)
		VALUES ('Pre-063 Type', 'Created before migration 063', true)
		RETURNING id::text
	`).Scan(&typeID)
	if err != nil {
		t.Fatalf("failed to insert pre-063 bed type: %v", err)
	}

	err = db.QueryRow(ctx, `
		INSERT INTO hospital_bed(name, bed_type, type_id, charge_minor, active, created_by, ward_name)
		VALUES ('PRE-63-BED', 'Pre-063 Type', $1, 75000, true, $2, 'Pre-063 Ward')
		RETURNING id::text
	`, typeID, admin.ID).Scan(&bedID)
	if err != nil {
		t.Fatalf("failed to insert pre-063 bed: %v", err)
	}

	err = db.QueryRow(ctx, `
		INSERT INTO encounter(kind, case_id, patient_id, doctor_id, bed_id, admitted_at, status, symptoms, created_by, request_key)
		VALUES ('ipd', $1, $2, $3, $4, clock_timestamp(), 'active', 'Pre-063 admission', $5, 'req-key-bed-pre63')
		RETURNING id::text
	`, caseID, patientID, doctor.ID, bedID, admin.ID).Scan(&encID)
	if err != nil {
		t.Fatalf("failed to insert pre-063 encounter: %v", err)
	}

	err = db.QueryRow(ctx, `
		INSERT INTO bed_assignment(bed_id, encounter_id, patient_id, notes, assigned_by)
		VALUES ($1, $2, $3, 'Pre-upgrade active assignment', $4)
		RETURNING id::text
	`, bedID, encID, patientID, doctor.ID).Scan(&assignID)
	if err != nil {
		t.Fatalf("failed to insert pre-063 bed assignment: %v", err)
	}

	// Apply migration 063 onto populated database
	m63Path := "../../../../../db/migrations/063_bed_management_parity.sql"
	sql63, err := os.ReadFile(m63Path)
	if err != nil {
		t.Fatalf("failed to read migration 063: %v", err)
	}

	_, err = db.Exec(ctx, string(sql63))
	if err != nil {
		t.Fatalf("migration 063 failed on populated database: %v", err)
	}

	// Verify all pre-existing records and values are 100% preserved
	var survivingBedName, survivingWard, survivingDesc string
	var survivingCharge int64
	err = db.QueryRow(ctx, `
		SELECT name, ward_name, charge_minor, description 
		FROM hospital_bed WHERE id = $1
	`, bedID).Scan(&survivingBedName, &survivingWard, &survivingCharge, &survivingDesc)
	if err != nil {
		t.Fatalf("failed to query surviving bed: %v", err)
	}
	if survivingBedName != "PRE-63-BED" || survivingWard != "Pre-063 Ward" || survivingCharge != 75000 || survivingDesc != "" {
		t.Fatalf("pre-existing bed corrupted by upgrade: name=%s ward=%s charge=%d desc=%s", survivingBedName, survivingWard, survivingCharge, survivingDesc)
	}

	var survivingAssignNotes string
	var survivingAssignedTo *time.Time
	err = db.QueryRow(ctx, `
		SELECT notes, assigned_to FROM bed_assignment WHERE id = $1
	`, assignID).Scan(&survivingAssignNotes, &survivingAssignedTo)
	if err != nil {
		t.Fatalf("failed to query surviving assignment: %v", err)
	}
	if survivingAssignNotes != "Pre-upgrade active assignment" || survivingAssignedTo != nil {
		t.Fatalf("pre-existing assignment corrupted by upgrade: notes=%s assignedTo=%v", survivingAssignNotes, survivingAssignedTo)
	}

	// Verify that the new unique index on bed_assignment enforces conflict rejection
	_, err = db.Exec(ctx, `
		INSERT INTO bed_assignment(bed_id, encounter_id, patient_id, notes, assigned_by)
		VALUES ($1, $2, $3, 'Colliding duplicate open assignment', $4)
	`, bedID, encID, patientID, admin.ID)
	if err == nil {
		t.Fatal("expected unique constraint violation on colliding open bed assignment under migration 063, got nil")
	}

	_ = schema
	_ = store
}
