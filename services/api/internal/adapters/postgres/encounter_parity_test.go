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

func setupEncounterTestDB(t *testing.T, throughMigration int) (*pgxpool.Pool, Store, string, func()) {
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
	schema := "hms_test_enc_" + hex.EncodeToString(random)
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

	if _, err := db.Exec(ctx, `CREATE TABLE IF NOT EXISTS schema_migration(name text PRIMARY KEY, applied_at timestamptz NOT NULL DEFAULT now())`); err != nil {
		db.Close()
		adminDB.Close()
		t.Fatal(err)
	}

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
		if _, e = db.Exec(ctx, `INSERT INTO schema_migration(name) VALUES($1) ON CONFLICT DO NOTHING`, base); e != nil {
			db.Close()
			adminDB.Close()
			t.Fatalf("failed recording %s in schema_migration: %v", base, e)
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

func seedEncounterActors(ctx context.Context, t *testing.T, db *pgxpool.Pool) (domain.Actor, domain.Actor, domain.Actor, domain.Actor, domain.Actor, string, string) {
	t.Helper()
	admin := domain.Actor{ID: "usr-admin-enc", Role: "admin"}
	doc1 := domain.Actor{ID: "usr-doc1-enc", Role: "doctor"}
	doc2 := domain.Actor{ID: "usr-doc2-enc", Role: "doctor"}
	nurse := domain.Actor{ID: "usr-nurse-enc", Role: "nurse"}
	patientUser := domain.Actor{ID: "usr-pat-enc", Role: "patient"}

	_, err := db.Exec(ctx, `
		INSERT INTO "user"(id, email, name)
		VALUES
			($1, 'admin_enc@hospital.test', 'Admin Encounter'),
			($2, 'doc1_enc@hospital.test', 'Doctor Primary'),
			($3, 'doc2_enc@hospital.test', 'Doctor Secondary'),
			($4, 'nurse_enc@hospital.test', 'Nurse Encounter'),
			($5, 'pat_enc@hospital.test', 'Patient Encounter')
		ON CONFLICT(id) DO NOTHING
	`, admin.ID, doc1.ID, doc2.ID, nurse.ID, patientUser.ID)
	if err != nil {
		t.Fatalf("failed to insert users: %v", err)
	}

	_, err = db.Exec(ctx, `
		INSERT INTO staff_access(user_id, role, active)
		VALUES 
			($1, 'admin', true),
			($2, 'doctor', true),
			($3, 'doctor', true),
			($4, 'nurse', true)
		ON CONFLICT(user_id) DO NOTHING
	`, admin.ID, doc1.ID, doc2.ID, nurse.ID)
	if err != nil {
		t.Fatalf("failed to insert staff access: %v", err)
	}

	var patientID string
	err = db.QueryRow(ctx, `
		INSERT INTO patient(given_name, family_name, phone, user_id)
		VALUES('Abebe', 'Bikila', '+251911223344', $1)
		RETURNING id
	`, patientUser.ID).Scan(&patientID)
	if err != nil {
		t.Fatalf("failed to create patient: %v", err)
	}

	_, err = db.Exec(ctx, `
		UPDATE patient_profile SET gender='male' WHERE patient_id=$1
	`, patientID)
	if err != nil {
		t.Fatalf("failed to update patient profile: %v", err)
	}

	var caseID string
	err = db.QueryRow(ctx, `
		INSERT INTO patient_case(patient_id, doctor_id, description, created_by)
		VALUES($1, $2, 'Cardiac observation', $3)
		RETURNING id
	`, patientID, doc1.ID, admin.ID).Scan(&caseID)
	if err != nil {
		t.Fatalf("failed to create case: %v", err)
	}

	return admin, doc1, doc2, nurse, patientUser, patientID, caseID
}

// 1. Verify populated upgrade preserves maintenance/unavailable states and 068 recovers prior states
func TestEncounter_PopulatedUpgrade_PreservesBedStates(t *testing.T) {
	db, store, _, cleanup := setupEncounterTestDB(t, 66)
	defer cleanup()
	ctx := context.Background()

	admin, doc1, _, _, _, patientID, caseID := seedEncounterActors(ctx, t, db)

	// Create bed type
	bedType, err := store.SaveBedType(ctx, admin, "", domain.BedTypeInput{Name: "Trauma Ward", Active: true})
	if err != nil {
		t.Fatalf("SaveBedType: %v", err)
	}

	// Bed 1: Intentionally in 'maintenance' (unoccupied)
	bedMaint, err := store.CreateBed(ctx, admin, domain.BedInput{
		Name: "BED-MAINT", TypeID: bedType.ID, WardName: "Trauma Wing A", ChargeMinor: 50000,
	})
	if err != nil {
		t.Fatalf("CreateBed 1: %v", err)
	}
	_, err = db.Exec(ctx, `UPDATE hospital_bed SET state='maintenance' WHERE id=$1`, bedMaint.ID)
	if err != nil {
		t.Fatalf("set maintenance: %v", err)
	}
	_, err = db.Exec(ctx, `INSERT INTO bed_state_event(bed_id, from_state, to_state, actor_id, reason, version) VALUES($1, 'ready', 'maintenance', $2, 'Telemetry calibration', 2)`, bedMaint.ID, admin.ID)
	if err != nil {
		t.Fatalf("record state event maint: %v", err)
	}

	// Bed 2: Intentionally in 'unavailable' (unoccupied)
	bedUnavail, err := store.CreateBed(ctx, admin, domain.BedInput{
		Name: "BED-UNAVAIL", TypeID: bedType.ID, WardName: "Trauma Wing B", ChargeMinor: 50000,
	})
	if err != nil {
		t.Fatalf("CreateBed 2: %v", err)
	}
	_, err = db.Exec(ctx, `UPDATE hospital_bed SET state='unavailable' WHERE id=$1`, bedUnavail.ID)
	if err != nil {
		t.Fatalf("set unavailable: %v", err)
	}
	_, err = db.Exec(ctx, `INSERT INTO bed_state_event(bed_id, from_state, to_state, actor_id, reason, version) VALUES($1, 'ready', 'unavailable', $2, 'Sanitization underway', 2)`, bedUnavail.ID, admin.ID)
	if err != nil {
		t.Fatalf("record state event unavail: %v", err)
	}

	// Bed 3: Normal 'ready' bed (unoccupied)
	bedReady, err := store.CreateBed(ctx, admin, domain.BedInput{
		Name: "BED-READY", TypeID: bedType.ID, WardName: "Trauma Wing C", ChargeMinor: 50000,
	})
	if err != nil {
		t.Fatalf("CreateBed 3: %v", err)
	}

	// Bed 4: Active inpatient bed (occupied)
	bedOccupied, err := store.CreateBed(ctx, admin, domain.BedInput{
		Name: "BED-OCCUPIED", TypeID: bedType.ID, WardName: "Trauma Wing D", ChargeMinor: 50000,
	})
	if err != nil {
		t.Fatalf("CreateBed 4: %v", err)
	}
	now := time.Now().UTC().Truncate(time.Microsecond)
	activeEnc, err := store.Admit(ctx, admin, domain.EncounterInput{
		Kind: "ipd", CaseID: caseID, BedID: bedOccupied.ID, AdmittedAt: now, Symptoms: "Acute trauma",
	}, "admit-occupied-key-1")
	if err != nil {
		t.Fatalf("Admit active encounter: %v", err)
	}

	// Bed 5: Historical discharged encounter with orphaned open bed_assignment
	var patient2ID string
	err = db.QueryRow(ctx, `
		INSERT INTO patient(given_name, family_name, phone)
		VALUES('Tirunesh', 'Dibaba', '+251922334455')
		RETURNING id
	`).Scan(&patient2ID)
	if err != nil {
		t.Fatalf("failed to create patient 2: %v", err)
	}

	var case2ID string
	err = db.QueryRow(ctx, `
		INSERT INTO patient_case(patient_id, doctor_id, description, created_by)
		VALUES($1, $2, 'Prior fracture repair', $3)
		RETURNING id
	`, patient2ID, doc1.ID, admin.ID).Scan(&case2ID)
	if err != nil {
		t.Fatalf("failed to create case 2: %v", err)
	}

	bedOrphaned, err := store.CreateBed(ctx, admin, domain.BedInput{
		Name: "BED-ORPHANED", TypeID: bedType.ID, WardName: "Trauma Wing E", ChargeMinor: 50000,
	})
	if err != nil {
		t.Fatalf("CreateBed 5: %v", err)
	}
	dischargedEnc, err := store.Admit(ctx, admin, domain.EncounterInput{
		Kind: "ipd", CaseID: case2ID, BedID: bedOrphaned.ID, AdmittedAt: now.Add(-48 * time.Hour), Symptoms: "Prior trauma",
	}, "admit-discharged-key-2")
	if err != nil {
		t.Fatalf("Admit discharged encounter: %v", err)
	}
	// Simulate historical unclosed assignment
	_, err = db.Exec(ctx, `UPDATE encounter SET status='discharged', discharged_at=clock_timestamp(), discharge_summary='Healed fully', bed_id=NULL WHERE id=$1`, dischargedEnc.ID)
	if err != nil {
		t.Fatalf("mark discharged: %v", err)
	}

	// Bed 6: Regression scenario where buggy 067 ran previously and set a maintenance bed to ready
	bedBuggy067, err := store.CreateBed(ctx, admin, domain.BedInput{
		Name: "BED-BUGGY-067", TypeID: bedType.ID, WardName: "Trauma Wing F", ChargeMinor: 50000,
	})
	if err != nil {
		t.Fatalf("CreateBed 6: %v", err)
	}
	// Record that staff placed Bed 6 into 'maintenance'
	_, err = db.Exec(ctx, `INSERT INTO bed_state_event(bed_id, from_state, to_state, actor_id, reason, version) VALUES($1, 'ready', 'maintenance', $2, 'Physical damage repair', 2)`, bedBuggy067.ID, admin.ID)
	if err != nil {
		t.Fatalf("record state event for bed 6: %v", err)
	}
	// Simulate buggy 067 erroneously setting Bed 6 state to 'ready'
	_, err = db.Exec(ctx, `UPDATE hospital_bed SET state='ready' WHERE id=$1`, bedBuggy067.ID)
	if err != nil {
		t.Fatalf("simulate buggy 067: %v", err)
	}

	// Apply migration 067
	mig067Path := "../../../../../db/migrations/067_clinical_encounter_parity_and_discharge_repair.sql"
	sql067, err := os.ReadFile(mig067Path)
	if err != nil {
		t.Fatalf("read 067: %v", err)
	}
	if _, err = db.Exec(ctx, string(sql067)); err != nil {
		t.Fatalf("apply 067: %v", err)
	}

	// ASSERTION 1: 067 did NOT flip Bed 1 ('maintenance') or Bed 2 ('unavailable')
	var state1, state2, state3, state4, state5 string
	_ = db.QueryRow(ctx, `SELECT state FROM hospital_bed WHERE id=$1`, bedMaint.ID).Scan(&state1)
	_ = db.QueryRow(ctx, `SELECT state FROM hospital_bed WHERE id=$1`, bedUnavail.ID).Scan(&state2)
	_ = db.QueryRow(ctx, `SELECT state FROM hospital_bed WHERE id=$1`, bedReady.ID).Scan(&state3)
	_ = db.QueryRow(ctx, `SELECT state FROM hospital_bed WHERE id=$1`, bedOccupied.ID).Scan(&state4)
	_ = db.QueryRow(ctx, `SELECT state FROM hospital_bed WHERE id=$1`, bedOrphaned.ID).Scan(&state5)

	if state1 != "maintenance" {
		t.Fatalf("migration 067 regressed maintenance bed to %q, expected 'maintenance'", state1)
	}
	if state2 != "unavailable" {
		t.Fatalf("migration 067 regressed unavailable bed to %q, expected 'unavailable'", state2)
	}
	if state3 != "ready" {
		t.Fatalf("expected ready bed to remain ready, got %q", state3)
	}
	if state4 != "ready" {
		t.Fatalf("expected occupied bed to remain ready, got %q", state4)
	}
	if state5 != "ready" {
		t.Fatalf("expected orphaned discharged bed to remain ready, got %q", state5)
	}

	// ASSERTION 2: Orphaned open assignment on Bed 5 was closed by 067
	var openAssignments int
	_ = db.QueryRow(ctx, `SELECT count(*) FROM bed_assignment WHERE encounter_id=$1 AND assigned_to IS NULL`, dischargedEnc.ID).Scan(&openAssignments)
	if openAssignments != 0 {
		t.Fatalf("expected 0 open assignments for discharged encounter, got %d", openAssignments)
	}

	// Apply corrective migration 068
	mig068Path := "../../../../../db/migrations/068_bed_state_preservation_and_repair.sql"
	sql068, err := os.ReadFile(mig068Path)
	if err != nil {
		t.Fatalf("read 068: %v", err)
	}
	if _, err = db.Exec(ctx, string(sql068)); err != nil {
		t.Fatalf("apply 068: %v", err)
	}

	// ASSERTION 3: Migration 068 recovered Bed 6 back to 'maintenance' from reliable evidence
	var state6 string
	err = db.QueryRow(ctx, `SELECT state FROM hospital_bed WHERE id=$1`, bedBuggy067.ID).Scan(&state6)
	if err != nil {
		t.Fatalf("query bed 6: %v", err)
	}
	if state6 != "maintenance" {
		t.Fatalf("migration 068 failed to recover bed 6 to 'maintenance' from bed_state_event, got %q", state6)
	}

	_ = activeEnc
	_ = patientID
}

// 2. Verify Single Encounter Retrieval & Scoped Access
func TestEncounter_GetByID_ScopingAndCareTeam(t *testing.T) {
	db, store, _, cleanup := setupEncounterTestDB(t, 0)
	defer cleanup()
	ctx := context.Background()

	admin, doc1, doc2, nurse, patientUser, _, caseID := seedEncounterActors(ctx, t, db)

	// Create bed & admit
	bedType, _ := store.SaveBedType(ctx, admin, "", domain.BedTypeInput{Name: "General Ward", Active: true})
	bed, _ := store.CreateBed(ctx, admin, domain.BedInput{Name: "B-101", TypeID: bedType.ID, ChargeMinor: 50000})

	now := time.Now().UTC().Truncate(time.Microsecond)
	enc, err := store.Admit(ctx, admin, domain.EncounterInput{
		Kind:       "ipd",
		CaseID:     caseID,
		BedID:      bed.ID,
		AdmittedAt: now,
		Symptoms:   "Acute abdomen",
		Intake: domain.EncounterIntake{
			Telephone:     "+251911223344",
			BloodPressure: "120/80",
			PaymentMode:   "cash",
		},
	}, "test-key-scoping-12345")
	if err != nil {
		t.Fatalf("Admit failed: %v", err)
	}

	// 1. Admin can read
	adminEnc, err := store.Encounter(ctx, admin, enc.ID)
	if err != nil {
		t.Fatalf("admin should read encounter: %v", err)
	}
	if adminEnc.ID != enc.ID || adminEnc.Symptoms != "Acute abdomen" {
		t.Fatalf("unexpected encounter data: %+v", adminEnc)
	}

	// 2. Primary Doctor can read
	doc1Enc, err := store.Encounter(ctx, doc1, enc.ID)
	if err != nil {
		t.Fatalf("primary doctor should read encounter: %v", err)
	}
	if doc1Enc.ID != enc.ID {
		t.Fatalf("primary doctor read wrong ID: %s", doc1Enc.ID)
	}

	// 3. Unassigned Doctor (doc2) CANNOT read (returns ErrNotFound)
	_, err = store.Encounter(ctx, doc2, enc.ID)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("unassigned doctor must get ErrNotFound, got %v", err)
	}

	// 4. Add doc2 to Care Team
	teamMember, err := store.AddCareTeamMember(ctx, doc1, enc.ID, domain.AddCareTeamMemberInput{
		StaffID:   doc2.ID,
		RoleTitle: "Consultant Cardiologist",
		Notes:     "Requested cardiac consultation",
	})
	if err != nil {
		t.Fatalf("AddCareTeamMember: %v", err)
	}

	// Now doc2 CAN read
	doc2Enc, err := store.Encounter(ctx, doc2, enc.ID)
	if err != nil {
		t.Fatalf("care team doctor should read encounter: %v", err)
	}
	if doc2Enc.ID != enc.ID {
		t.Fatalf("care team doctor read wrong ID: %s", doc2Enc.ID)
	}

	// 5. Revoke doc2 from Care Team
	err = store.RevokeCareTeamMember(ctx, doc1, teamMember.ID)
	if err != nil {
		t.Fatalf("RevokeCareTeamMember: %v", err)
	}

	// Now doc2 CANNOT read again
	_, err = store.Encounter(ctx, doc2, enc.ID)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("revoked doctor must get ErrNotFound, got %v", err)
	}

	// 6. Patient portal owner can read
	patEnc, err := store.Encounter(ctx, patientUser, enc.ID)
	if err != nil {
		t.Fatalf("patient portal owner should read encounter: %v", err)
	}
	if patEnc.ID != enc.ID {
		t.Fatalf("patient portal owner read wrong ID: %s", patEnc.ID)
	}

	// 7. Nurse on care team can read
	nurseMember, err := store.AddCareTeamMember(ctx, admin, enc.ID, domain.AddCareTeamMemberInput{
		StaffID:   nurse.ID,
		RoleTitle: "Primary Nurse",
	})
	if err != nil {
		t.Fatalf("AddCareTeamMember for nurse: %v", err)
	}
	nurseEnc, err := store.Encounter(ctx, nurse, enc.ID)
	if err != nil {
		t.Fatalf("care team nurse should read encounter: %v", err)
	}
	if nurseEnc.ID != enc.ID {
		t.Fatalf("nurse read wrong ID: %s", nurseEnc.ID)
	}
	_ = nurseMember
}

// 3. Verify Encounter Update & Concurrency Protection
func TestEncounter_Update_ConcurrencyAndOwnership(t *testing.T) {
	db, store, _, cleanup := setupEncounterTestDB(t, 0)
	defer cleanup()
	ctx := context.Background()

	admin, doc1, doc2, _, _, _, caseID := seedEncounterActors(ctx, t, db)

	bedType, _ := store.SaveBedType(ctx, admin, "", domain.BedTypeInput{Name: "General Ward", Active: true})
	bed, _ := store.CreateBed(ctx, admin, domain.BedInput{Name: "B-201", TypeID: bedType.ID, ChargeMinor: 50000})

	now := time.Now().UTC().Truncate(time.Microsecond)
	enc, err := store.Admit(ctx, admin, domain.EncounterInput{
		Kind:       "ipd",
		CaseID:     caseID,
		BedID:      bed.ID,
		AdmittedAt: now,
		Symptoms:   "Initial headache",
		Intake: domain.EncounterIntake{
			Telephone:     "+251911223344",
			BloodPressure: "118/76",
			PaymentMode:   "cash",
		},
	}, "test-key-update-lock-12345")
	if err != nil {
		t.Fatalf("Admit failed: %v", err)
	}

	// 1. Unassigned doctor cannot update (ErrForbidden)
	_, err = store.UpdateEncounter(ctx, doc2, enc.ID, domain.UpdateEncounterInput{
		Symptoms: "Tampered symptoms",
		Version:  enc.Version,
		Intake:   enc.Intake,
	})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("unassigned doctor must receive ErrForbidden, got %v", err)
	}

	// 2. Assigned doctor updates successfully
	updated, err := store.UpdateEncounter(ctx, doc1, enc.ID, domain.UpdateEncounterInput{
		Symptoms: "Migraine with aura; nausea resolved",
		Version:  enc.Version,
		Intake: domain.EncounterIntake{
			Telephone:     "+251911223344",
			BloodPressure: "125/82",
			Height:        175,
			Weight:        70,
			PaymentMode:   "cash",
		},
	})
	if err != nil {
		t.Fatalf("doctor update failed: %v", err)
	}
	if updated.Symptoms != "Migraine with aura; nausea resolved" {
		t.Fatalf("expected updated symptoms, got %q", updated.Symptoms)
	}
	if updated.Version != enc.Version+1 {
		t.Fatalf("expected version %d, got %d", enc.Version+1, updated.Version)
	}

	// 3. Stale edit rejection: updating with old version returns ErrStale
	_, err = store.UpdateEncounter(ctx, doc1, enc.ID, domain.UpdateEncounterInput{
		Symptoms: "Stale edit attempt",
		Version:  enc.Version, // stale!
		Intake:   updated.Intake,
	})
	if !errors.Is(err, domain.ErrStale) {
		t.Fatalf("stale edit must return ErrStale, got %v", err)
	}

	// 4. Admin reassigns doctor
	reassigned, err := store.UpdateEncounter(ctx, admin, enc.ID, domain.UpdateEncounterInput{
		Symptoms: updated.Symptoms,
		DoctorID: doc2.ID,
		Version:  updated.Version,
		Intake:   updated.Intake,
	})
	if err != nil {
		t.Fatalf("admin doctor reassignment failed: %v", err)
	}
	if reassigned.DoctorID != doc2.ID {
		t.Fatalf("expected doctor %s, got %s", doc2.ID, reassigned.DoctorID)
	}
	if reassigned.Version != updated.Version+1 {
		t.Fatalf("expected version %d, got %d", updated.Version+1, reassigned.Version)
	}

	// Now doc2 is primary doctor and can update; doc1 is no longer primary
	_, err = store.UpdateEncounter(ctx, doc1, enc.ID, domain.UpdateEncounterInput{
		Symptoms: "Doc1 trying after reassignment",
		Version:  reassigned.Version,
		Intake:   reassigned.Intake,
	})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("doc1 should now be forbidden, got %v", err)
	}
}

// 4. Verify Discharge, Atomic Bed Release, and Bed Availability Recalculation
func TestEncounter_Discharge_AtomicBedReleaseAndConcurrency(t *testing.T) {
	db, store, _, cleanup := setupEncounterTestDB(t, 0)
	defer cleanup()
	ctx := context.Background()

	admin, doc1, doc2, _, _, _, caseID := seedEncounterActors(ctx, t, db)

	bedType, _ := store.SaveBedType(ctx, admin, "", domain.BedTypeInput{Name: "General Ward", Active: true})
	bed, err := store.CreateBed(ctx, admin, domain.BedInput{
		Name:        "B-301",
		TypeID:      bedType.ID,
		WardName:    "Medical Ward",
		ChargeMinor: 75000,
	})
	if err != nil {
		t.Fatalf("CreateBed: %v", err)
	}

	// Verify bed is initially available
	beds, _, err := store.SearchBeds(ctx, admin, "B-301", "", "", nil, 1, 10)
	if err != nil || len(beds) == 0 {
		t.Fatalf("SearchBeds initial: %v", err)
	}
	if !beds[0].Available || beds[0].State != "ready" {
		t.Fatalf("expected available bed initially, got available=%v state=%s", beds[0].Available, beds[0].State)
	}

	// Admit patient to bed
	admittedAt := time.Now().UTC().Add(-2 * time.Hour).Truncate(time.Microsecond)
	enc, err := store.Admit(ctx, admin, domain.EncounterInput{
		Kind:       "ipd",
		CaseID:     caseID,
		BedID:      bed.ID,
		AdmittedAt: admittedAt,
		Symptoms:   "Pneumonia",
	}, "test-key-discharge-release-123")
	if err != nil {
		t.Fatalf("Admit failed: %v", err)
	}

	// Bed is now occupied and unavailable
	beds, _, err = store.SearchBeds(ctx, admin, "B-301", "", "", nil, 1, 10)
	if err != nil || len(beds) == 0 {
		t.Fatalf("SearchBeds after admit: %v", err)
	}
	if beds[0].Available {
		t.Fatalf("expected bed to be occupied/unavailable after admit, got available=%v", beds[0].Available)
	}

	// Attempting to admit another patient to same bed fails
	case2ID := caseID
	_, err = store.Admit(ctx, admin, domain.EncounterInput{
		Kind:       "ipd",
		CaseID:     case2ID,
		BedID:      bed.ID,
		AdmittedAt: time.Now().UTC(),
		Symptoms:   "Another illness",
	}, "conflict-admit-key-1234")
	if err == nil {
		t.Fatal("admitting to occupied bed must fail, but succeeded")
	}

	// Unassigned doctor cannot discharge
	clinic := application.Clinical{Store: store, Now: time.Now}
	_, err = clinic.Discharge(ctx, doc2, enc.ID, domain.Discharge{
		Version: enc.Version,
		Summary: "Premature discharge",
	})
	if !errors.Is(err, domain.ErrNotFound) && !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("unassigned doctor must not discharge, got %v", err)
	}

	// Stale discharge attempt (invalid version)
	_, err = clinic.Discharge(ctx, doc1, enc.ID, domain.Discharge{
		Version: enc.Version + 99,
		Summary: "Discharged with bad version",
	})
	if !errors.Is(err, domain.ErrStale) {
		t.Fatalf("stale discharge must return ErrStale, got %v", err)
	}

	// Primary doctor discharges patient
	dischargedAt := time.Now().UTC()
	discharged, err := store.Discharge(ctx, doc1, enc.ID, domain.Discharge{
		Version: enc.Version,
		Summary: "Patient recovered fully. Prescribed oral antibiotics.",
	}, dischargedAt)
	if err != nil {
		t.Fatalf("Discharge failed: %v", err)
	}

	if discharged.Status != "discharged" {
		t.Fatalf("expected status 'discharged', got %s", discharged.Status)
	}
	if discharged.BedID != "" {
		t.Fatalf("expected bed_id to be cleared, got %s", discharged.BedID)
	}
	if discharged.DischargeSummary != "Patient recovered fully. Prescribed oral antibiotics." {
		t.Fatalf("unexpected summary: %s", discharged.DischargeSummary)
	}

	// Verify hospital_bed state was ATOMICALLY released to 'ready'
	var finalBedState string
	err = db.QueryRow(ctx, `SELECT state FROM hospital_bed WHERE id=$1`, bed.ID).Scan(&finalBedState)
	if err != nil {
		t.Fatalf("query bed: %v", err)
	}
	if finalBedState != "ready" {
		t.Fatalf("expected bed state 'ready' after discharge, got %q", finalBedState)
	}

	// Verify SearchBeds confirms bed is available and occupancy is cleared
	beds, _, err = store.SearchBeds(ctx, admin, "B-301", "", "", nil, 1, 10)
	if err != nil || len(beds) == 0 {
		t.Fatalf("SearchBeds after discharge: %v", err)
	}
	if !beds[0].Available || beds[0].State != "ready" {
		t.Fatalf("expected bed to be available after discharge, got available=%v state=%s", beds[0].Available, beds[0].State)
	}

	// Verify bed_assignment was closed
	var closedCount int
	err = db.QueryRow(ctx, `SELECT count(*) FROM bed_assignment WHERE encounter_id=$1 AND assigned_to IS NOT NULL`, enc.ID).Scan(&closedCount)
	if err != nil {
		t.Fatalf("query bed_assignment: %v", err)
	}
	if closedCount == 0 {
		t.Fatal("expected closed bed assignment record")
	}

	// Verify double discharge fails with ErrStale
	_, err = store.Discharge(ctx, doc1, enc.ID, domain.Discharge{
		Version: discharged.Version,
		Summary: "Second discharge",
	}, time.Now().UTC())
	if !errors.Is(err, domain.ErrStale) {
		t.Fatalf("double discharge must return ErrStale, got %v", err)
	}
}

// 5. Verify Admin Discharge Permission
func TestEncounter_Discharge_AdminPermission(t *testing.T) {
	db, store, _, cleanup := setupEncounterTestDB(t, 0)
	defer cleanup()
	ctx := context.Background()

	admin, _, _, _, _, _, caseID := seedEncounterActors(ctx, t, db)

	bedType, _ := store.SaveBedType(ctx, admin, "", domain.BedTypeInput{Name: "General Ward", Active: true})
	bed, _ := store.CreateBed(ctx, admin, domain.BedInput{Name: "B-Admin-01", TypeID: bedType.ID, ChargeMinor: 40000})

	enc, err := store.Admit(ctx, admin, domain.EncounterInput{
		Kind:       "ipd",
		CaseID:     caseID,
		BedID:      bed.ID,
		AdmittedAt: time.Now().UTC().Add(-1 * time.Hour),
		Symptoms:   "Observation",
	}, "test-admin-discharge-key-1234")
	if err != nil {
		t.Fatalf("Admit: %v", err)
	}

	// Admin executes discharge
	clinic := application.Clinical{Store: store, Now: time.Now}
	discharged, err := clinic.Discharge(ctx, admin, enc.ID, domain.Discharge{
		Version: enc.Version,
		Summary: "Administrative discharge authorized by CMO",
	})
	if err != nil {
		t.Fatalf("admin discharge failed: %v", err)
	}
	if discharged.Status != "discharged" {
		t.Fatalf("expected discharged status, got %s", discharged.Status)
	}

	var bedState string
	_ = db.QueryRow(ctx, `SELECT state FROM hospital_bed WHERE id=$1`, bed.ID).Scan(&bedState)
	if bedState != "ready" {
		t.Fatalf("expected bed state ready after admin discharge, got %s", bedState)
	}
}
