package postgres

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"hms.local/api/internal/application"
	"hms.local/api/internal/domain"
	"net/url"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

// Tests use a fresh schema, never patient tables in the public development schema.
func TestClinicalTransactions(t *testing.T) {
	dsn := os.Getenv("HMS_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set HMS_TEST_DATABASE_URL to run isolated-schema PostgreSQL tests")
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
	defer adminDB.Close()
	random := make([]byte, 12)
	if _, err = rand.Read(random); err != nil {
		t.Fatal(err)
	}
	schema := "hms_test_" + hex.EncodeToString(random)
	quoted := pgx.Identifier{schema}.Sanitize()
	if _, err = adminDB.Exec(ctx, "CREATE SCHEMA "+quoted); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if _, e := adminDB.Exec(ctx, "DROP SCHEMA "+quoted+" CASCADE"); e != nil {
			t.Error(e)
		}
	}()
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		t.Fatal(err)
	}
	config.ConnConfig.RuntimeParams["search_path"] = schema
	db, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	migrations, err := filepath.Glob("../../../../../db/migrations/*.sql")
	if err != nil || len(migrations) < 3 {
		t.Fatal("migration files unavailable")
	}
	for _, file := range migrations {
		sql, e := os.ReadFile(file)
		if e != nil {
			t.Fatal(e)
		}
		if _, e = db.Exec(ctx, string(sql)); e != nil {
			t.Fatal(file, e)
		}
	}
	actors := []domain.Actor{{ID: "admin", Role: "admin"}, {ID: "doctor", Role: "doctor", Name: "Doctor"}, {ID: "other-doctor", Role: "doctor"}, {ID: "patient", Role: "patient"}, {ID: "reception", Role: "receptionist"}}
	for _, a := range actors {
		if _, err = db.Exec(ctx, `INSERT INTO "user"(id,name,email) VALUES($1,$1,$1||'@example.test')`, a.ID); err != nil {
			t.Fatal(err)
		}
		if _, err = db.Exec(ctx, `INSERT INTO staff_access(user_id,role) VALUES($1,$2)`, a.ID, a.Role); err != nil {
			t.Fatal(err)
		}
	}
	store := Store{DB: db}
	clinic := application.Clinical{Store: store, Now: time.Now}
	hospital := application.Hospital{Store: store, Now: time.Now}
	patients := []domain.Patient{}
	cases := []domain.Case{}
	for n := 0; n < 2; n++ {
		p, e := hospital.Register(ctx, actors[0], domain.PatientInput{GivenName: "Synthetic", FamilyName: "Test", DateOfBirth: "2000-01-01"})
		if e != nil {
			t.Fatal(e)
		}
		patients = append(patients, p)
		c, e := clinic.CreateCase(ctx, actors[0], domain.CaseInput{PatientID: p.ID, DoctorID: "doctor"})
		if e != nil {
			t.Fatal(e)
		}
		cases = append(cases, c)
	}
	if err = hospital.LinkPatient(ctx, actors[0], patients[0].ID, domain.PatientAccess{UserID: "patient", ClinicianID: "doctor"}); err != nil {
		t.Fatal(err)
	}
	bed, err := clinic.CreateBed(ctx, actors[0], domain.BedInput{Name: "A1", Type: "General", ChargeMinor: 12345})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = clinic.CreateBed(ctx, actors[3], domain.BedInput{Name: "Blocked", Type: "General"}); !errors.Is(err, domain.ErrForbidden) {
		t.Fatal("patient created bed", err)
	}
	inputs := []domain.EncounterInput{{Kind: "ipd", CaseID: cases[0].ID, BedID: bed.ID, AdmittedAt: time.Now().Add(-time.Minute)}, {Kind: "ipd", CaseID: cases[1].ID, BedID: bed.ID, AdmittedAt: time.Now().Add(-time.Minute)}}
	results := make([]domain.Encounter, 2)
	errs := make([]error, 2)
	var wg sync.WaitGroup
	for n := 0; n < 2; n++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			results[n], errs[n] = clinic.Admit(ctx, actors[0], inputs[n], "admission-key-000"+string(rune('0'+n)))
		}(n)
	}
	wg.Wait()
	winner := 0
	if errs[0] != nil {
		winner = 1
	}
	loser := 1 - winner
	if errs[winner] != nil || !errors.Is(errs[loser], domain.ErrStale) {
		t.Fatalf("bed race: %v", errs)
	}
	enc := results[winner]
	retry, e := clinic.Admit(ctx, actors[0], inputs[winner], "admission-key-000"+string(rune('0'+winner)))
	if e != nil || retry.ID != enc.ID {
		t.Fatal("retry did not return same admission", e)
	}
	beds, e := clinic.Beds(ctx, actors[0], 1)
	if e != nil || beds[0].Available {
		t.Fatal("bed occupancy not reflected", e)
	}
	if _, e = clinic.SignNote(ctx, actors[2], enc.ID, domain.NoteInput{Body: "Unauthorized"}, "note-key-00000001"); !errors.Is(e, domain.ErrNotFound) {
		t.Fatal("unassigned doctor wrote note", e)
	}
	if _, e = clinic.Notes(ctx, actors[4], enc.ID); !errors.Is(e, domain.ErrForbidden) {
		t.Fatal("reception read clinical notes", e)
	}
	note, e := clinic.SignNote(ctx, actors[1], enc.ID, domain.NoteInput{Body: "Signed synthetic clinical assessment"}, "note-key-00000002")
	if e != nil {
		t.Fatal(e)
	}
	duplicate, e := clinic.SignNote(ctx, actors[1], enc.ID, domain.NoteInput{Body: note.Body}, "note-key-00000002")
	if e != nil || duplicate.ID != note.ID {
		t.Fatal("note retry duplicated", e)
	}
	if _, e = db.Exec(ctx, `UPDATE clinical_note SET body='changed' WHERE id=$1`, note.ID); e == nil {
		t.Fatal("signed note was changed")
	}
	if _, e = db.Exec(ctx, `DELETE FROM clinical_note WHERE id=$1`, note.ID); e == nil {
		t.Fatal("signed note was deleted")
	}
	own, e := clinic.Encounters(ctx, actors[3], "ipd", 1)
	if e != nil {
		t.Fatal(e)
	}
	for _, row := range own {
		if row.PatientID != patients[0].ID {
			t.Fatal("patient saw someone else's admission")
		}
	}
	if _, e = clinic.Discharge(ctx, actors[0], enc.ID, domain.Discharge{Version: 1, Summary: "Admin cannot sign discharge"}); !errors.Is(e, domain.ErrForbidden) {
		t.Fatal("admin signed discharge", e)
	}
	out, e := clinic.Discharge(ctx, actors[1], enc.ID, domain.Discharge{Version: 1, Summary: "Synthetic discharge summary"})
	if e != nil || out.Status != "discharged" {
		t.Fatal(e)
	}
	if _, e = clinic.SignNote(ctx, actors[1], enc.ID, domain.NoteInput{Body: "Late note"}, "note-key-00000003"); !errors.Is(e, domain.ErrStale) {
		t.Fatal("closed encounter accepted note", e)
	}
	beds, e = clinic.Beds(ctx, actors[0], 1)
	if e != nil || !beds[0].Available {
		t.Fatal("discharge did not free bed", e)
	}
	if _, e = clinic.Admit(ctx, actors[0], inputs[loser], "admission-key-rebook"); e != nil {
		t.Fatal("released bed could not be allocated", e)
	}
}
