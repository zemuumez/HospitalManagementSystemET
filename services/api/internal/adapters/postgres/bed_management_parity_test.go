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
		{BedID: bed.ID, EncounterID: enc1ID, PatientID: patient1ID, Notes: "Patient 1 admission", Version: 1},
		{BedID: bed.ID, EncounterID: enc2ID, PatientID: patient2ID, Notes: "Patient 2 admission", Version: 1},
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
		Version:     1,
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
		Version:     1,
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
		Version:     1,
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
		Version:     1,
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
		Version:     1,
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
		Version:     assignA.EncounterVersion,
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
	beds, total, err := store.SearchBeds(ctx, admin, "", "", "", nil, 1, 10)
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
		Version:     1,
	})
	if err != nil {
		t.Fatalf("failed to assign bed: %v", err)
	}

	repBefore, _ := store.BedOccupancyReport(ctx)
	if repBefore.OccupiedBeds != 1 {
		t.Fatalf("expected 1 occupied bed before discharge, got: %d", repBefore.OccupiedBeds)
	}

	// Discharge bed assignment
	dischargedAssign, err := store.DischargeBedAssignment(ctx, admin, assign.ID, domain.DischargeBedAssignmentInput{
		Version:        assign.EncounterVersion,
		DischargeNotes: "Patient stabilized, discharged home",
	})
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
	_, err = store.DischargeBedAssignment(ctx, admin, assign.ID, domain.DischargeBedAssignmentInput{
		Version:        dischargedAssign.EncounterVersion,
		DischargeNotes: "Duplicate discharge attempt",
	})
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

// ─────────────────────────────────────────────────────────────────────────────
// Independent Review Findings Regressions (R1 - R5)
// ─────────────────────────────────────────────────────────────────────────────

// R1: Role access scoping & capacity-only bed redaction
func TestBed_ReviewR1_RoleAccessScopeAndRedaction(t *testing.T) {
	_, store, _, cleanup := setupBedTestDB(t, 0)
	defer cleanup()
	ctx := context.Background()

	admin, doctor, _, patientID, encID, _ := seedBedStaffAndPatient(ctx, t, store.DB)

	// Create unrelated doctor and unrelated nurse
	unrelatedDoc := domain.Actor{ID: "usr-unrelated-doc", Role: "doctor"}
	unrelatedNurse := domain.Actor{ID: "usr-unrelated-nurse", Role: "nurse"}
	receptionist := domain.Actor{ID: "usr-receptionist-test", Role: "receptionist"}

	_, err := store.DB.Exec(ctx, `
		INSERT INTO "user"(id, email, name)
		VALUES
			($1, 'unrelated_doc@hospital.test', 'Unrelated Doctor'),
			($2, 'unrelated_nurse@hospital.test', 'Unrelated Nurse'),
			($3, 'receptionist_bed@hospital.test', 'Receptionist Bed')
		ON CONFLICT(id) DO NOTHING
	`, unrelatedDoc.ID, unrelatedNurse.ID, receptionist.ID)
	if err != nil {
		t.Fatalf("failed to insert unrelated users: %v", err)
	}
	_, err = store.DB.Exec(ctx, `
		INSERT INTO staff_access(user_id, role, active)
		VALUES
			($1, 'doctor', true),
			($2, 'nurse', true),
			($3, 'receptionist', true)
		ON CONFLICT(user_id) DO NOTHING
	`, unrelatedDoc.ID, unrelatedNurse.ID, receptionist.ID)
	if err != nil {
		t.Fatalf("failed to insert unrelated staff_access: %v", err)
	}

	bt, _ := store.SaveBedType(ctx, admin, "", domain.BedTypeInput{Name: "R1 Bed Type", Active: true})
	bed, _ := store.CreateBed(ctx, admin, domain.BedInput{TypeID: bt.ID, Name: "R1-BED-01", WardName: "Private ICU", ChargeMinor: 120000})

	assign, err := store.AssignBed(ctx, admin, domain.AssignBedInput{
		BedID:       bed.ID,
		EncounterID: encID,
		PatientID:   patientID,
		Notes:       "R1 Protected Patient Admission",
		Version:     1,
	})
	if err != nil {
		t.Fatalf("failed to assign bed: %v", err)
	}

	// 1. Unrelated nurse must NOT see this assignment in ListAllBedAssignments
	unrelatedNurseList, total, err := store.ListAllBedAssignments(ctx, unrelatedNurse, "", "active", 1, 25)
	if err != nil {
		t.Fatalf("unrelated nurse ListAllBedAssignments failed: %v", err)
	}
	if total != 0 || len(unrelatedNurseList) != 0 {
		t.Fatalf("security violation: unrelated nurse saw %d assignments (expected 0)", total)
	}

	// 2. Unrelated doctor must NOT see this assignment in ListAllBedAssignments
	unrelatedDocList, total, err := store.ListAllBedAssignments(ctx, unrelatedDoc, "", "active", 1, 25)
	if err != nil {
		t.Fatalf("unrelated doctor ListAllBedAssignments failed: %v", err)
	}
	if total != 0 || len(unrelatedDocList) != 0 {
		t.Fatalf("security violation: unrelated doctor saw %d assignments (expected 0)", total)
	}

	// 2b. Doctor B (who is patient.clinician_user_id but NOT on doctor A's encounter)
	// Canonical encounter access denies Doctor B:
	// - AuthorizeClinicalRecord returns ErrForbidden
	// - ListAllBedAssignments returns 0 rows
	// - SearchBeds returns bed as occupied but REDACTS patient clinical details
	// - SearchBeds with patient search returns 0 results
	patientClinicianDoc := domain.Actor{ID: "usr-patient-clinician-b", Role: "doctor"}
	_, err = store.DB.Exec(ctx, `
		INSERT INTO "user"(id, email, name) VALUES ($1, 'doc_b@hospital.test', 'Doctor B') ON CONFLICT(id) DO NOTHING
	`, patientClinicianDoc.ID)
	if err != nil {
		t.Fatalf("failed to insert doc B: %v", err)
	}
	_, err = store.DB.Exec(ctx, `
		INSERT INTO staff_access(user_id, role, active) VALUES ($1, 'doctor', true) ON CONFLICT(user_id) DO NOTHING
	`, patientClinicianDoc.ID)
	if err != nil {
		t.Fatalf("failed to insert staff_access for doc B: %v", err)
	}
	_, err = store.DB.Exec(ctx, `
		UPDATE patient SET clinician_user_id = $1 WHERE id = $2
	`, patientClinicianDoc.ID, patientID)
	if err != nil {
		t.Fatalf("failed to update patient clinician_user_id: %v", err)
	}

	// Canonical admission authorization must deny Doctor B for Doctor A's encounter
	if err := store.AuthorizeClinicalRecord(ctx, patientClinicianDoc, patientID, encID, "admission"); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected ErrForbidden for patient-level clinician B accessing doctor A encounter, got: %v", err)
	}

	// ListAllBedAssignments must NOT return Doctor A's encounter assignment to Doctor B
	docBList, totalB, err := store.ListAllBedAssignments(ctx, patientClinicianDoc, "", "active", 1, 25)
	if err != nil {
		t.Fatalf("doctor B ListAllBedAssignments failed: %v", err)
	}
	if totalB != 0 || len(docBList) != 0 {
		t.Fatalf("security violation: doctor B saw %d assignments belonging to doctor A encounter", totalB)
	}

	// SearchBeds must redact patient details for Doctor B
	docBBeds, _, err := store.SearchBeds(ctx, patientClinicianDoc, "", "", "", nil, 1, 10)
	if err != nil {
		t.Fatalf("doctor B SearchBeds failed: %v", err)
	}
	for _, b := range docBBeds {
		if b.ID == bed.ID {
			if b.PatientName != "" || b.PatientPhone != "" || b.PatientGender != "" || b.EncounterID != "" {
				t.Fatalf("security violation: doctor B saw patient clinical details on doctor A encounter bed: %+v", b)
			}
		}
	}

	// SearchBeds with patient name must return 0 results for Doctor B, but 1 for Doctor A
	_, totalSearchB, err := store.SearchBeds(ctx, patientClinicianDoc, "Derartu", "", "", nil, 1, 10)
	if err != nil {
		t.Fatalf("doctor B SearchBeds by name failed: %v", err)
	}
	if totalSearchB != 0 {
		t.Fatalf("security violation: doctor B matched bed by patient name, got %d", totalSearchB)
	}

	_, totalSearchA, err := store.SearchBeds(ctx, doctor, "Derartu", "", "", nil, 1, 10)
	if err != nil {
		t.Fatalf("doctor A SearchBeds by name failed: %v", err)
	}
	if totalSearchA != 1 {
		t.Fatalf("doctor A expected to match 1 bed by patient name, got %d", totalSearchA)
	}

	// 3. Admin, receptionist, and assigned doctor MUST see the assignment
	adminList, total, err := store.ListAllBedAssignments(ctx, admin, "", "active", 1, 25)
	if err != nil || total != 1 || len(adminList) != 1 {
		t.Fatalf("admin expected 1 assignment, got %d, err: %v", total, err)
	}
	recList, total, err := store.ListAllBedAssignments(ctx, receptionist, "", "active", 1, 25)
	if err != nil || total != 1 || len(recList) != 1 {
		t.Fatalf("receptionist expected 1 assignment, got %d, err: %v", total, err)
	}
	docList, total, err := store.ListAllBedAssignments(ctx, doctor, "", "active", 1, 25)
	if err != nil || total != 1 || len(docList) != 1 {
		t.Fatalf("assigned doctor expected 1 assignment, got %d, err: %v", total, err)
	}
	if docList[0].EncounterVersion <= 0 {
		t.Fatalf("expected positive encounter version in listing, got: %d", docList[0].EncounterVersion)
	}

	// 4. Unrelated nurse attempting to discharge assignment must be rejected with ErrForbidden
	_, err = store.DischargeBedAssignment(ctx, unrelatedNurse, assign.ID, domain.DischargeBedAssignmentInput{
		Version:        assign.EncounterVersion,
		DischargeNotes: "Malicious discharge by unrelated nurse",
	})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected ErrForbidden when unrelated nurse attempts discharge, got: %v", err)
	}

	// 5. Redaction test in SearchBeds:
	// Capacity-only viewer (unrelated nurse) sees bed occupied, but patient clinical details REDACTED
	unrelatedBeds, _, err := store.SearchBeds(ctx, unrelatedNurse, "", "", "", nil, 1, 10)
	if err != nil {
		t.Fatalf("unrelated nurse SearchBeds failed: %v", err)
	}
	foundRedacted := false
	for _, b := range unrelatedBeds {
		if b.ID == bed.ID {
			foundRedacted = true
			if b.Available {
				t.Fatalf("expected bed to be occupied (not available)")
			}
			if b.PatientName != "" || b.PatientPhone != "" || b.PatientGender != "" || b.EncounterID != "" {
				t.Fatalf("security violation: patient details not redacted for capacity viewer: %+v", b)
			}
		}
	}
	if !foundRedacted {
		t.Fatal("bed not found in capacity listing")
	}

	// Authorized viewer (admin) sees patient clinical details
	adminBeds, _, err := store.SearchBeds(ctx, admin, "", "", "", nil, 1, 10)
	if err != nil {
		t.Fatalf("admin SearchBeds failed: %v", err)
	}
	foundAuthorized := false
	for _, b := range adminBeds {
		if b.ID == bed.ID {
			foundAuthorized = true
			if b.PatientName == "" || b.EncounterID == "" {
				t.Fatalf("expected authorized viewer to see patient details, got empty: %+v", b)
			}
		}
	}
	if !foundAuthorized {
		t.Fatal("bed not found in admin listing")
	}
}

// R2: Canonical Store.TransferBed and Clinical.Discharge consistency
func TestBed_ReviewR2_CanonicalTransferAndDischargeConsistency(t *testing.T) {
	_, store, _, cleanup := setupBedTestDB(t, 0)
	defer cleanup()
	ctx := context.Background()

	admin, doctor, _, patientID, encID, _ := seedBedStaffAndPatient(ctx, t, store.DB)

	bt, _ := store.SaveBedType(ctx, admin, "", domain.BedTypeInput{Name: "R2 Type", Active: true})
	bedA, _ := store.CreateBed(ctx, admin, domain.BedInput{TypeID: bt.ID, Name: "R2-BED-A", WardName: "Ward A", ChargeMinor: 40000})
	bedB, _ := store.CreateBed(ctx, admin, domain.BedInput{TypeID: bt.ID, Name: "R2-BED-B", WardName: "Ward B", ChargeMinor: 70000})

	// Initial assignment to Bed A
	assignA, err := store.AssignBed(ctx, admin, domain.AssignBedInput{
		BedID:       bedA.ID,
		EncounterID: encID,
		PatientID:   patientID,
		Notes:       "Initial Bed A",
		Version:     1,
	})
	if err != nil {
		t.Fatalf("failed initial assignment: %v", err)
	}

	// Query current encounter to get current version
	var encVersion int
	err = store.DB.QueryRow(ctx, `SELECT version FROM encounter WHERE id = $1`, encID).Scan(&encVersion)
	if err != nil {
		t.Fatalf("failed to query encounter version: %v", err)
	}

	// Transfer using canonical Store.TransferBed
	transferredEnc, err := store.TransferBed(ctx, admin, encID, domain.BedTransfer{
		BedID:   bedB.ID,
		Version: encVersion,
		Reason:  "Condition escalated, transferred to Ward B",
	})
	if err != nil {
		t.Fatalf("Store.TransferBed failed: %v", err)
	}
	if transferredEnc.BedID != bedB.ID {
		t.Fatalf("expected encounter bed_id to be Bed B, got: %s", transferredEnc.BedID)
	}

	// Verify Bed A assignment is now closed
	var assignATo *time.Time
	err = store.DB.QueryRow(ctx, `SELECT assigned_to FROM bed_assignment WHERE id = $1`, assignA.ID).Scan(&assignATo)
	if err != nil || assignATo == nil {
		t.Fatalf("expected Bed A assignment to be closed with assigned_to, got: %v", assignATo)
	}

	// Verify new assignment exists for Bed B and is open
	var assignBID string
	var assignBTo *time.Time
	err = store.DB.QueryRow(ctx, `
		SELECT id, assigned_to FROM bed_assignment 
		WHERE encounter_id = $1 AND bed_id = $2
	`, encID, bedB.ID).Scan(&assignBID, &assignBTo)
	if err != nil {
		t.Fatalf("failed to find new assignment for Bed B: %v", err)
	}
	if assignBTo != nil {
		t.Fatalf("expected Bed B assignment to be active (assigned_to NULL), got: %v", assignBTo)
	}

	// Verify occupancy report: Bed A is available, Bed B is occupied
	rep, _ := store.BedOccupancyReport(ctx)
	if rep.OccupiedBeds != 1 || rep.AvailableBeds != 1 {
		t.Fatalf("unexpected occupancy after transfer: %+v", rep)
	}

	// Now perform clinical discharge via clinical.go Discharge
	now := time.Now().UTC()
	dischargedEnc, err := store.Discharge(ctx, doctor, encID, domain.Discharge{
		Version: transferredEnc.Version,
		Summary: "Patient discharged successfully",
	}, now)
	if err != nil {
		t.Fatalf("clinical.Discharge failed: %v", err)
	}
	if dischargedEnc.Status != "discharged" || dischargedEnc.BedID != "" {
		t.Fatalf("expected discharged encounter with cleared bed_id, got: %+v", dischargedEnc)
	}

	// Verify Bed B assignment is now closed
	err = store.DB.QueryRow(ctx, `SELECT assigned_to FROM bed_assignment WHERE id = $1`, assignBID).Scan(&assignBTo)
	if err != nil || assignBTo == nil {
		t.Fatalf("expected Bed B assignment to be closed on clinical discharge, got: %v", assignBTo)
	}

	// Verify Bed B is now available
	repAfter, _ := store.BedOccupancyReport(ctx)
	if repAfter.OccupiedBeds != 0 || repAfter.AvailableBeds != 2 {
		t.Fatalf("expected 0 occupied and 2 available after clinical discharge, got: %+v", repAfter)
	}
}

// R3: Mandatory positive optimistic version contracts
func TestBed_ReviewR3_MandatoryPositiveVersionContracts(t *testing.T) {
	_, store, _, cleanup := setupBedTestDB(t, 0)
	defer cleanup()
	ctx := context.Background()

	admin, _, _, patientID, encID, _ := seedBedStaffAndPatient(ctx, t, store.DB)

	bt, _ := store.SaveBedType(ctx, admin, "", domain.BedTypeInput{Name: "R3 Type", Active: true})
	bed, _ := store.CreateBed(ctx, admin, domain.BedInput{TypeID: bt.ID, Name: "R3-BED-01", WardName: "General", ChargeMinor: 50000})

	// 1. AssignBed with Version <= 0 must fail validation
	inv1 := domain.AssignBedInput{BedID: bed.ID, EncounterID: encID, PatientID: patientID, Version: 0}
	if err := inv1.Validate(); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected ErrValidation on AssignBed with version 0, got: %v", err)
	}
	invNeg := domain.AssignBedInput{BedID: bed.ID, EncounterID: encID, PatientID: patientID, Version: -1}
	if err := invNeg.Validate(); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected ErrValidation on AssignBed with version -1, got: %v", err)
	}

	// Stale version in AssignBed must return ErrStale
	_, err := store.AssignBed(ctx, admin, domain.AssignBedInput{
		BedID:       bed.ID,
		EncounterID: encID,
		PatientID:   patientID,
		Version:     999, // Stale!
	})
	if !errors.Is(err, domain.ErrStale) {
		t.Fatalf("expected ErrStale on mismatched version in AssignBed, got: %v", err)
	}

	// Correct version succeeds
	assign, err := store.AssignBed(ctx, admin, domain.AssignBedInput{
		BedID:       bed.ID,
		EncounterID: encID,
		PatientID:   patientID,
		Version:     1,
	})
	if err != nil {
		t.Fatalf("expected success with version 1, got: %v", err)
	}

	// 2. DischargeBedAssignmentInput with Version <= 0 must fail validation
	discInv := domain.DischargeBedAssignmentInput{Version: 0}
	if err := discInv.Validate(); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected ErrValidation on DischargeBedAssignmentInput with version 0, got: %v", err)
	}

	// Stale version in DischargeBedAssignment must return ErrStale
	_, err = store.DischargeBedAssignment(ctx, admin, assign.ID, domain.DischargeBedAssignmentInput{
		Version: 999, // Stale!
	})
	if !errors.Is(err, domain.ErrStale) {
		t.Fatalf("expected ErrStale on mismatched version in DischargeBedAssignment, got: %v", err)
	}

	// 3. DeleteBedType with version <= 0 or stale version must return ErrStale
	err = store.DeleteBedType(ctx, admin, bt.ID, 0)
	if !errors.Is(err, domain.ErrStale) && !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected ErrStale or ErrValidation on DeleteBedType with version 0, got: %v", err)
	}
	err = store.DeleteBedType(ctx, admin, bt.ID, 999)
	if !errors.Is(err, domain.ErrStale) {
		t.Fatalf("expected ErrStale on DeleteBedType with mismatched version, got: %v", err)
	}
}

// Helper function: standard migration runner with schema_migration tracking table
func runBedMigrationRunner(t *testing.T, db *pgxpool.Pool, ctx context.Context) {
	t.Helper()
	if _, err := db.Exec(ctx, `CREATE TABLE IF NOT EXISTS schema_migration(name text PRIMARY KEY, applied_at timestamptz NOT NULL DEFAULT now())`); err != nil {
		t.Fatalf("failed to create schema_migration: %v", err)
	}
	files, err := filepath.Glob("../../../../../db/migrations/*.sql")
	if err != nil || len(files) == 0 {
		t.Fatalf("failed to glob migration files: %v", err)
	}
	sort.Strings(files)
	for _, f := range files {
		base := filepath.Base(f)
		var exists bool
		err = db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM schema_migration WHERE name=$1)`, base).Scan(&exists)
		if err != nil {
			t.Fatalf("failed checking schema_migration for %s: %v", base, err)
		}
		if exists {
			continue
		}
		sqlBytes, err := os.ReadFile(f)
		if err != nil {
			t.Fatalf("failed reading migration %s: %v", f, err)
		}
		if _, err = db.Exec(ctx, string(sqlBytes)); err != nil {
			t.Fatalf("normal migration runner failed applying %s: %v", base, err)
		}
		if _, err = db.Exec(ctx, `INSERT INTO schema_migration(name) VALUES($1)`, base); err != nil {
			t.Fatalf("failed recording %s in schema_migration: %v", base, err)
		}
	}
}

// R4 / F1: Migration 064 safely handles surviving open duplicate without unique constraint collision (SQLSTATE 23505)
func TestBed_ReviewR4_SurvivingOpenDuplicateCollision(t *testing.T) {
	db, _, _, cleanup := setupBedTestDB(t, 62)
	defer cleanup()
	ctx := context.Background()

	admin := domain.Actor{ID: "usr-admin-f1", Role: "admin"}
	doctor := domain.Actor{ID: "usr-doc-f1", Role: "doctor"}

	_, _ = db.Exec(ctx, `
		INSERT INTO "user"(id, email, name) VALUES ($1, 'admin_f1@hospital.test', 'Admin F1'), ($2, 'doc_f1@hospital.test', 'Doc F1')
		ON CONFLICT(id) DO NOTHING
	`, admin.ID, doctor.ID)
	_, _ = db.Exec(ctx, `
		INSERT INTO staff_access(user_id, role, active) VALUES ($1, 'admin', true), ($2, 'doctor', true)
		ON CONFLICT(user_id) DO NOTHING
	`, admin.ID, doctor.ID)

	var patientID, caseID, encID string
	_ = db.QueryRow(ctx, `INSERT INTO patient(given_name, family_name, date_of_birth, phone) VALUES ('Abebe', 'Bikila', '1960-01-01', '+251911112233') RETURNING id::text`).Scan(&patientID)
	_ = db.QueryRow(ctx, `INSERT INTO patient_case(patient_id, doctor_id, description, created_by) VALUES ($1, $2, 'Case F1', $3) RETURNING id::text`, patientID, doctor.ID, admin.ID).Scan(&caseID)

	var typeID, bedID string
	_ = db.QueryRow(ctx, `INSERT INTO bed_type(name, description, active) VALUES ('Type F1', 'Desc', true) RETURNING id::text`).Scan(&typeID)
	_ = db.QueryRow(ctx, `INSERT INTO hospital_bed(name, bed_type, type_id, charge_minor, active, created_by, ward_name) VALUES ('F1-BED', 'Type F1', $1, 60000, true, $2, 'Ward F1') RETURNING id::text`, typeID, admin.ID).Scan(&bedID)
	_ = db.QueryRow(ctx, `INSERT INTO encounter(kind, case_id, patient_id, doctor_id, bed_id, admitted_at, status, symptoms, created_by, request_key) VALUES ('ipd', $1, $2, $3, $4, clock_timestamp(), 'active', 'Symptoms', $5, 'req-f1') RETURNING id::text`, caseID, patientID, doctor.ID, bedID, admin.ID).Scan(&encID)

	// Seed TWO assignments for the same active encounter & bed:
	// Older assignment 1 (closed by original 063 with assigned_to = created_at)
	// Newer assignment 2 (surviving open assignment: assigned_to IS NULL)
	var assign1ID, assign2ID string
	t1 := time.Now().Add(-5 * time.Hour).UTC()
	t2 := time.Now().Add(-2 * time.Hour).UTC()
	_ = db.QueryRow(ctx, `
		INSERT INTO bed_assignment(bed_id, encounter_id, patient_id, notes, assigned_by, assigned_from, assigned_to, created_at)
		VALUES ($1, $2, $3, 'Older duplicate assignment', $4, $5, $5, $5)
		RETURNING id::text
	`, bedID, encID, patientID, doctor.ID, t1).Scan(&assign1ID)

	_ = db.QueryRow(ctx, `
		INSERT INTO bed_assignment(bed_id, encounter_id, patient_id, notes, assigned_by, assigned_from, assigned_to, created_at)
		VALUES ($1, $2, $3, 'Surviving open assignment', $4, $5, NULL, $5)
		RETURNING id::text
	`, bedID, encID, patientID, doctor.ID, t2).Scan(&assign2ID)

	// In an original-063 database, the unique indexes are already active on bed_assignment
	_, err := db.Exec(ctx, `
		CREATE UNIQUE INDEX IF NOT EXISTS one_active_bed_assignment ON bed_assignment(bed_id) WHERE assigned_to IS NULL;
		CREATE UNIQUE INDEX IF NOT EXISTS one_active_encounter_bed_assignment ON bed_assignment(encounter_id) WHERE assigned_to IS NULL;
	`)
	if err != nil {
		t.Fatalf("failed to create unique indexes: %v", err)
	}

	// Apply migration 064: MUST NOT CRASH with SQLSTATE 23505!
	sql64, err := os.ReadFile("../../../../../db/migrations/064_bed_assignment_reconciliation_audit.sql")
	if err != nil {
		t.Fatalf("failed to read migration 064: %v", err)
	}
	if _, err = db.Exec(ctx, string(sql64)); err != nil {
		t.Fatalf("migration 064 failed on surviving open duplicate: %v", err)
	}

	// Apply migration 065: repair must also run cleanly
	sql65, err := os.ReadFile("../../../../../db/migrations/065_bed_assignment_reconciliation_repair.sql")
	if err != nil {
		t.Fatalf("failed to read migration 065: %v", err)
	}
	if _, err = db.Exec(ctx, string(sql65)); err != nil {
		t.Fatalf("migration 065 failed: %v", err)
	}

	// Verify surviving assignment 2 remains active
	var a2AssignedTo *time.Time
	err = db.QueryRow(ctx, `SELECT assigned_to FROM bed_assignment WHERE id = $1`, assign2ID).Scan(&a2AssignedTo)
	if err != nil || a2AssignedTo != nil {
		t.Fatalf("expected assignment 2 to remain active with assigned_to NULL, got err=%v val=%v", err, a2AssignedTo)
	}

	// Verify older assignment 1 was archived and quarantined, NOT reopened to cause collision
	var a1AssignedTo *time.Time
	err = db.QueryRow(ctx, `SELECT assigned_to FROM bed_assignment WHERE id = $1`, assign1ID).Scan(&a1AssignedTo)
	if err != nil {
		t.Fatalf("failed to query assignment 1: %v", err)
	}
	if a1AssignedTo == nil {
		t.Fatalf("security violation: assignment 1 was erroneously reopened, violating unique slot occupancy")
	}

	// Verify audit table captured full original provenance
	var auditCount int
	var action, origNotes string
	var origRecord []byte
	err = db.QueryRow(ctx, `
		SELECT COUNT(*), COALESCE(MAX(resolution_action), ''), COALESCE(MAX(original_notes), ''), COALESCE(MAX(original_record::text), '')
		FROM bed_assignment_migration_audit
		WHERE bed_assignment_id = $1
	`, assign1ID).Scan(&auditCount, &action, &origNotes, &origRecord)
	if err != nil || auditCount != 1 || action != "quarantined_superseded_duplicate" {
		t.Fatalf("expected audit record with quarantined_superseded_duplicate, count=%d action='%s', err=%v", auditCount, action, err)
	}
	if origNotes != "Older duplicate assignment" || len(origRecord) == 0 {
		t.Fatalf("expected audit table to capture complete original notes and JSON record")
	}
}

// R4 / F2: Fresh installation archives duplicate open assignments before mutation without fabricating timestamps
func TestBed_ReviewR4_FreshInstallAndProvenanceArchiving(t *testing.T) {
	db, _, _, cleanup := setupBedTestDB(t, 62)
	defer cleanup()
	ctx := context.Background()

	admin := domain.Actor{ID: "usr-admin-f2", Role: "admin"}
	doctor := domain.Actor{ID: "usr-doc-f2", Role: "doctor"}

	_, _ = db.Exec(ctx, `
		INSERT INTO "user"(id, email, name) VALUES ($1, 'admin_f2@hospital.test', 'Admin F2'), ($2, 'doc_f2@hospital.test', 'Doc F2')
		ON CONFLICT(id) DO NOTHING
	`, admin.ID, doctor.ID)
	_, _ = db.Exec(ctx, `
		INSERT INTO staff_access(user_id, role, active) VALUES ($1, 'admin', true), ($2, 'doctor', true)
		ON CONFLICT(user_id) DO NOTHING
	`, admin.ID, doctor.ID)

	var patientID, caseID, encID string
	_ = db.QueryRow(ctx, `INSERT INTO patient(given_name, family_name, date_of_birth, phone) VALUES ('Haile', 'Gebrselassie', '1973-04-18', '+251911445566') RETURNING id::text`).Scan(&patientID)
	_ = db.QueryRow(ctx, `INSERT INTO patient_case(patient_id, doctor_id, description, created_by) VALUES ($1, $2, 'Case F2', $3) RETURNING id::text`, patientID, doctor.ID, admin.ID).Scan(&caseID)

	var typeID, bedID string
	_ = db.QueryRow(ctx, `INSERT INTO bed_type(name, description, active) VALUES ('Type F2', 'Desc', true) RETURNING id::text`).Scan(&typeID)
	_ = db.QueryRow(ctx, `INSERT INTO hospital_bed(name, bed_type, type_id, charge_minor, active, created_by, ward_name) VALUES ('F2-BED', 'Type F2', $1, 60000, true, $2, 'Ward F2') RETURNING id::text`, typeID, admin.ID).Scan(&bedID)
	_ = db.QueryRow(ctx, `INSERT INTO encounter(kind, case_id, patient_id, doctor_id, bed_id, admitted_at, status, symptoms, created_by, request_key) VALUES ('ipd', $1, $2, $3, $4, clock_timestamp(), 'active', 'Symptoms', $5, 'req-f2') RETURNING id::text`, caseID, patientID, doctor.ID, bedID, admin.ID).Scan(&encID)

	// Seed two concurrently open assignments (assigned_to IS NULL) before running migrations 063-065
	var assign1ID, assign2ID string
	t1 := time.Now().Add(-4 * time.Hour).UTC()
	t2 := time.Now().Add(-1 * time.Hour).UTC()
	_ = db.QueryRow(ctx, `
		INSERT INTO bed_assignment(bed_id, encounter_id, patient_id, notes, assigned_by, assigned_from, assigned_to, created_at)
		VALUES ($1, $2, $3, 'First open duplicate', $4, $5, NULL, $5)
		RETURNING id::text
	`, bedID, encID, patientID, doctor.ID, t1).Scan(&assign1ID)

	_ = db.QueryRow(ctx, `
		INSERT INTO bed_assignment(bed_id, encounter_id, patient_id, notes, assigned_by, assigned_from, assigned_to, created_at)
		VALUES ($1, $2, $3, 'Second open duplicate', $4, $5, NULL, $5)
		RETURNING id::text
	`, bedID, encID, patientID, doctor.ID, t2).Scan(&assign2ID)

	// Run migration runner through 063, 064, and 065
	runBedMigrationRunner(t, db, ctx)

	// Exactly ONE assignment must survive as active in bed_assignment
	var openCount int
	err := db.QueryRow(ctx, `SELECT COUNT(*) FROM bed_assignment WHERE bed_id = $1 AND assigned_to IS NULL`, bedID).Scan(&openCount)
	if err != nil || openCount != 1 {
		t.Fatalf("expected exactly 1 active assignment for bed, got %d, err: %v", openCount, err)
	}

	// Verify the older duplicate was archived with complete provenance and WITHOUT fabricated timestamps
	var auditCount int
	var resolutionAction, origNotes string
	var reconciledAssignedTo, adminQuarantineAt *time.Time
	err = db.QueryRow(ctx, `
		SELECT COUNT(*), COALESCE(MAX(resolution_action), ''), COALESCE(MAX(original_notes), ''),
		       MAX(reconciled_assigned_to), MAX(administrative_quarantine_at)
		FROM bed_assignment_migration_audit
		WHERE bed_assignment_id = $1
	`, assign1ID).Scan(&auditCount, &resolutionAction, &origNotes, &reconciledAssignedTo, &adminQuarantineAt)
	if err != nil || auditCount != 1 {
		t.Fatalf("expected 1 audit record for duplicate assignment 1, got %d, err: %v", auditCount, err)
	}
	if resolutionAction != "quarantined_duplicate" {
		t.Fatalf("expected resolution_action 'quarantined_duplicate', got: %s", resolutionAction)
	}
	if origNotes != "First open duplicate" {
		t.Fatalf("expected complete original notes captured in audit table, got: %s", origNotes)
	}
	if reconciledAssignedTo != nil {
		t.Fatalf("clinical end time must remain NULL (unknown) without fabricating durations, got: %v", reconciledAssignedTo)
	}
	if adminQuarantineAt == nil {
		t.Fatalf("administrative quarantine time must be separately recorded")
	}
}

// R4: Restoration of legitimately active assignment when slot is vacant
func TestBed_ReviewR4_VacantSlotRestoration(t *testing.T) {
	db, _, _, cleanup := setupBedTestDB(t, 62)
	defer cleanup()
	ctx := context.Background()

	admin := domain.Actor{ID: "usr-admin-vac", Role: "admin"}
	doctor := domain.Actor{ID: "usr-doc-vac", Role: "doctor"}

	_, _ = db.Exec(ctx, `
		INSERT INTO "user"(id, email, name) VALUES ($1, 'admin_vac@hospital.test', 'Admin Vac'), ($2, 'doc_vac@hospital.test', 'Doc Vac')
		ON CONFLICT(id) DO NOTHING
	`, admin.ID, doctor.ID)
	_, _ = db.Exec(ctx, `
		INSERT INTO staff_access(user_id, role, active) VALUES ($1, 'admin', true), ($2, 'doctor', true)
		ON CONFLICT(user_id) DO NOTHING
	`, admin.ID, doctor.ID)

	var patientID, caseID, encID string
	_ = db.QueryRow(ctx, `INSERT INTO patient(given_name, family_name, date_of_birth, phone) VALUES ('Kenenisa', 'Bekele', '1982-06-13', '+251911778899') RETURNING id::text`).Scan(&patientID)
	_ = db.QueryRow(ctx, `INSERT INTO patient_case(patient_id, doctor_id, description, created_by) VALUES ($1, $2, 'Case Vac', $3) RETURNING id::text`, patientID, doctor.ID, admin.ID).Scan(&caseID)

	var typeID, bedID string
	_ = db.QueryRow(ctx, `INSERT INTO bed_type(name, description, active) VALUES ('Type Vac', 'Desc', true) RETURNING id::text`).Scan(&typeID)
	_ = db.QueryRow(ctx, `INSERT INTO hospital_bed(name, bed_type, type_id, charge_minor, active, created_by, ward_name) VALUES ('VAC-BED', 'Type Vac', $1, 60000, true, $2, 'Ward Vac') RETURNING id::text`, typeID, admin.ID).Scan(&bedID)
	_ = db.QueryRow(ctx, `INSERT INTO encounter(kind, case_id, patient_id, doctor_id, bed_id, admitted_at, status, symptoms, created_by, request_key) VALUES ('ipd', $1, $2, $3, $4, clock_timestamp(), 'active', 'Symptoms', $5, 'req-vac') RETURNING id::text`, caseID, patientID, doctor.ID, bedID, admin.ID).Scan(&encID)

	// Seed single assignment closed with fabricated created_at and NO other assignment
	var assignID string
	createdAt := time.Now().Add(-2 * time.Hour).UTC()
	_ = db.QueryRow(ctx, `
		INSERT INTO bed_assignment(bed_id, encounter_id, patient_id, notes, assigned_by, assigned_from, assigned_to, created_at)
		VALUES ($1, $2, $3, 'Erroneously closed assignment', $4, $5, $5, $5)
		RETURNING id::text
	`, bedID, encID, patientID, doctor.ID, createdAt).Scan(&assignID)

	// Run migration runner through 063, 064, and 065
	runBedMigrationRunner(t, db, ctx)

	// Verify the vacant assignment is safely restored to active (assigned_to IS NULL)
	var assignedTo *time.Time
	err := db.QueryRow(ctx, `SELECT assigned_to FROM bed_assignment WHERE id = $1`, assignID).Scan(&assignedTo)
	if err != nil || assignedTo != nil {
		t.Fatalf("expected assignment to be safely restored with assigned_to NULL, got err=%v val=%v", err, assignedTo)
	}

	var action string
	_ = db.QueryRow(ctx, `SELECT resolution_action FROM bed_assignment_migration_audit WHERE bed_assignment_id = $1`, assignID).Scan(&action)
	if action != "restored_active" {
		t.Fatalf("expected resolution_action 'restored_active', got: %s", action)
	}
}

// R5: Pagination beyond 25 encounters and 100 beds
func TestBed_ReviewR5_PaginationAndSelectors(t *testing.T) {
	_, store, _, cleanup := setupBedTestDB(t, 0)
	defer cleanup()
	ctx := context.Background()

	admin, doctor, _, patientID, _, _ := seedBedStaffAndPatient(ctx, t, store.DB)

	var caseID string
	_ = store.DB.QueryRow(ctx, `INSERT INTO patient_case(patient_id, doctor_id, description, created_by) VALUES ($1, $2, 'Case R5', $3) RETURNING id::text`, patientID, doctor.ID, admin.ID).Scan(&caseID)

	// Seed 30 active encounters across distinct patients
	for i := 1; i <= 30; i++ {
		var pID, cID string
		err := store.DB.QueryRow(ctx, `
			INSERT INTO patient(given_name, family_name, date_of_birth, phone)
			VALUES ($1, 'Patient', '1990-01-01', $2)
			RETURNING id::text
		`, fmt.Sprintf("Bulk-%d", i), fmt.Sprintf("+251911%06d", i)).Scan(&pID)
		if err != nil {
			t.Fatalf("failed to insert bulk patient %d: %v", i, err)
		}
		err = store.DB.QueryRow(ctx, `INSERT INTO patient_case(patient_id, doctor_id, description, created_by) VALUES ($1, $2, 'Case R5', $3) RETURNING id::text`, pID, doctor.ID, admin.ID).Scan(&cID)
		if err != nil {
			t.Fatalf("failed to insert bulk case %d: %v", i, err)
		}

		reqKey := fmt.Sprintf("req-page-enc-%d", i)
		_, err = store.DB.Exec(ctx, `
			INSERT INTO encounter(kind, case_id, patient_id, doctor_id, admitted_at, status, symptoms, created_by, request_key)
			VALUES ('ipd', $1, $2, $3, clock_timestamp() - make_interval(mins => $4), 'active', 'Symptoms', $5, $6)
		`, cID, pID, doctor.ID, i, admin.ID, reqKey)
		if err != nil {
			t.Fatalf("failed to insert encounter %d: %v", i, err)
		}
	}

	// Fetch encounters with pageSize = 50 -> must return all 31 (including seeded initial)
	encList, err := store.Encounters(ctx, admin, "ipd", 1, 50)
	if err != nil {
		t.Fatalf("failed to query encounters: %v", err)
	}
	if len(encList) < 30 {
		t.Fatalf("expected >= 30 encounters across page 1 with pageSize 50, got: %d", len(encList))
	}

	// Seed 110 beds across multiple bed types
	bt, _ := store.SaveBedType(ctx, admin, "", domain.BedTypeInput{Name: "Bulk Type", Active: true})
	for i := 1; i <= 110; i++ {
		bedName := fmt.Sprintf("BULK-%03d", i)
		_, err := store.CreateBed(ctx, admin, domain.BedInput{
			TypeID:      bt.ID,
			Name:        bedName,
			WardName:    "Large Medical Ward",
			ChargeMinor: 25000,
		})
		if err != nil {
			t.Fatalf("failed to create bulk bed %d: %v", i, err)
		}
	}

	// SearchBeds with pageSize = 150 must return all >100 beds in single query
	allBeds, totalBeds, err := store.SearchBeds(ctx, admin, "", "", "", nil, 1, 150)
	if err != nil {
		t.Fatalf("failed to search beds: %v", err)
	}
	if totalBeds < 110 || len(allBeds) < 110 {
		t.Fatalf("expected >= 110 beds with pageSize 150, got len=%d total=%d", len(allBeds), totalBeds)
	}
}

// R4 / F1: Corrective migration 065 executes on databases where 064 was already recorded in schema_migration
func TestBed_ReviewR4_AlreadyApplied064Runner(t *testing.T) {
	db, _, _, cleanup := setupBedTestDB(t, 62)
	defer cleanup()
	ctx := context.Background()

	admin := domain.Actor{ID: "usr-admin-rep", Role: "admin"}
	doctor := domain.Actor{ID: "usr-doc-rep", Role: "doctor"}

	_, _ = db.Exec(ctx, `
		INSERT INTO "user"(id, email, name) VALUES ($1, 'admin_rep@hospital.test', 'Admin Rep'), ($2, 'doc_rep@hospital.test', 'Doc Rep')
		ON CONFLICT(id) DO NOTHING
	`, admin.ID, doctor.ID)
	_, _ = db.Exec(ctx, `
		INSERT INTO staff_access(user_id, role, active) VALUES ($1, 'admin', true), ($2, 'doctor', true)
		ON CONFLICT(user_id) DO NOTHING
	`, admin.ID, doctor.ID)

	var patientID, caseID, encID string
	_ = db.QueryRow(ctx, `INSERT INTO patient(given_name, family_name, date_of_birth, phone) VALUES ('Tirunesh', 'Dibaba', '1985-06-01', '+251911334455') RETURNING id::text`).Scan(&patientID)
	_ = db.QueryRow(ctx, `INSERT INTO patient_case(patient_id, doctor_id, description, created_by) VALUES ($1, $2, 'Case Rep', $3) RETURNING id::text`, patientID, doctor.ID, admin.ID).Scan(&caseID)

	var typeID, bedID string
	_ = db.QueryRow(ctx, `INSERT INTO bed_type(name, description, active) VALUES ('Type Rep', 'Desc', true) RETURNING id::text`).Scan(&typeID)
	_ = db.QueryRow(ctx, `INSERT INTO hospital_bed(name, bed_type, type_id, charge_minor, active, created_by, ward_name) VALUES ('REP-BED', 'Type Rep', $1, 60000, true, $2, 'Ward Rep') RETURNING id::text`, typeID, admin.ID).Scan(&bedID)
	_ = db.QueryRow(ctx, `INSERT INTO encounter(kind, case_id, patient_id, doctor_id, bed_id, admitted_at, status, symptoms, created_by, request_key) VALUES ('ipd', $1, $2, $3, $4, clock_timestamp(), 'active', 'Symptoms', $5, 'req-rep') RETURNING id::text`, caseID, patientID, doctor.ID, bedID, admin.ID).Scan(&encID)

	// Apply 063 and 064 manually and record them in schema_migration
	if _, err := db.Exec(ctx, `CREATE TABLE IF NOT EXISTS schema_migration(name text PRIMARY KEY, applied_at timestamptz NOT NULL DEFAULT now())`); err != nil {
		t.Fatalf("failed to create schema_migration: %v", err)
	}
	sql63, _ := os.ReadFile("../../../../../db/migrations/063_bed_management_parity.sql")
	if _, err := db.Exec(ctx, string(sql63)); err != nil {
		t.Fatalf("failed applying 063: %v", err)
	}
	_, _ = db.Exec(ctx, `INSERT INTO schema_migration(name) VALUES ('063_bed_management_parity.sql') ON CONFLICT DO NOTHING`)

	sql64, _ := os.ReadFile("../../../../../db/migrations/064_bed_assignment_reconciliation_audit.sql")
	if _, err := db.Exec(ctx, string(sql64)); err != nil {
		t.Fatalf("failed applying 064: %v", err)
	}
	_, _ = db.Exec(ctx, `INSERT INTO schema_migration(name) VALUES ('064_bed_assignment_reconciliation_audit.sql') ON CONFLICT DO NOTHING`)

	// Normal migration runner runs: MUST skip 063, 064 and apply 065 cleanly!
	runBedMigrationRunner(t, db, ctx)

	var count065 int
	_ = db.QueryRow(ctx, `SELECT count(*) FROM schema_migration WHERE name='065_bed_assignment_reconciliation_repair.sql'`).Scan(&count065)
	if count065 != 1 {
		t.Fatalf("expected 065 in schema_migration, got: %d", count065)
	}
}

// F4: Multiple active encounters per patient and pagination beyond cutoffs
func TestBed_F4_MultipleEncountersPerPatient(t *testing.T) {
	_, store, _, cleanup := setupBedTestDB(t, 0)
	defer cleanup()
	ctx := context.Background()

	admin, doctor, _, patientID, _, _ := seedBedStaffAndPatient(ctx, t, store.DB)

	var case1ID, case2ID string
	_ = store.DB.QueryRow(ctx, `INSERT INTO patient_case(patient_id, doctor_id, description, created_by) VALUES ($1, $2, 'Case 1', $3) RETURNING id::text`, patientID, doctor.ID, admin.ID).Scan(&case1ID)
	_ = store.DB.QueryRow(ctx, `INSERT INTO patient_case(patient_id, doctor_id, description, created_by) VALUES ($1, $2, 'Case 2', $3) RETURNING id::text`, patientID, doctor.ID, admin.ID).Scan(&case2ID)

	// Create TWO active IPD encounters for the SAME patient
	var enc1ID, enc2ID string
	_ = store.DB.QueryRow(ctx, `
		INSERT INTO encounter(kind, case_id, patient_id, doctor_id, admitted_at, status, symptoms, created_by, request_key)
		VALUES ('ipd', $1, $2, $3, clock_timestamp() - interval '2 days', 'active', 'Admission 1', $4, 'req-menc-1')
		RETURNING id::text
	`, case1ID, patientID, doctor.ID, admin.ID).Scan(&enc1ID)

	_ = store.DB.QueryRow(ctx, `
		INSERT INTO encounter(kind, case_id, patient_id, doctor_id, admitted_at, status, symptoms, created_by, request_key)
		VALUES ('ipd', $1, $2, $3, clock_timestamp() - interval '1 hour', 'active', 'Admission 2', $4, 'req-menc-2')
		RETURNING id::text
	`, case2ID, patientID, doctor.ID, admin.ID).Scan(&enc2ID)

	encs, err := store.Encounters(ctx, admin, "ipd", 1, 50)
	if err != nil {
		t.Fatalf("failed to query encounters: %v", err)
	}

	found1, found2 := false, false
	for _, e := range encs {
		if e.ID == enc1ID {
			found1 = true
		}
		if e.ID == enc2ID {
			found2 = true
		}
	}
	if !found1 || !found2 {
		t.Fatalf("expected both active encounters for same patient to be returned, found1=%v, found2=%v", found1, found2)
	}
}


