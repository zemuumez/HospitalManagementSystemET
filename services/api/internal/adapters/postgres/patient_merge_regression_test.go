package postgres

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5/pgxpool"
	"hms.local/api/internal/application"
	"hms.local/api/internal/domain"
	"sync"
	"testing"
	"time"
)

func testMergePreservesHistory(t *testing.T, db *pgxpool.Pool, store Store, actors []domain.Actor, patients []domain.Patient) {
	ctx := context.Background()
	var before, after int
	if err := db.QueryRow(ctx, `SELECT count(*) FROM encounter WHERE patient_id=$1`, patients[1].ID).Scan(&before); err != nil {
		t.Fatal(err)
	}
	app := application.PatientExtensionsService{Store: store}
	err := app.MergePatients(ctx, actors[0], domain.MergePatientInput{PrimaryPatientID: patients[0].ID, MergedPatientID: patients[1].ID, Reason: "Review test"})
	if !errors.Is(err, domain.ErrStale) {
		t.Fatalf("merge with active care must conflict: %v", err)
	}
	if err = db.QueryRow(ctx, `SELECT count(*) FROM encounter WHERE patient_id=$1`, patients[1].ID).Scan(&after); err != nil {
		t.Fatal(err)
	}
	if before != after {
		t.Fatal("merge changed historical encounters")
	}
	var events int
	if err = db.QueryRow(ctx, `SELECT count(*) FROM patient_merge_event WHERE merged_patient_id=$1`, patients[1].ID).Scan(&events); err != nil {
		t.Fatal(err)
	}
	if events != 0 {
		t.Fatal("merge reported success through an audit event")
	}
	hospital := application.Hospital{Store: store, Now: time.Now}
	primary, e := hospital.Register(ctx, actors[0], domain.PatientInput{GivenName: "Primary", FamilyName: "Merge", DateOfBirth: "2000-01-01"})
	if e != nil {
		t.Fatal(e)
	}
	source, e := hospital.Register(ctx, actors[0], domain.PatientInput{GivenName: "Duplicate", FamilyName: "Merge", DateOfBirth: "2000-01-01"})
	if e != nil {
		t.Fatal(e)
	}
	if _, e = db.Exec(ctx, `INSERT INTO "user"(id,name,email) VALUES('merge-owner','Merge owner','merge-owner@example.test'); INSERT INTO staff_access(user_id,role) VALUES('merge-owner','patient')`); e != nil {
		t.Fatal(e)
	}
	owner := domain.Actor{ID: "merge-owner", Role: "patient"}
	if e = hospital.LinkPatient(ctx, actors[0], source.ID, domain.PatientAccess{UserID: owner.ID, ClinicianID: actors[1].ID}); e != nil {
		t.Fatal(e)
	}
	clinic := application.Clinical{Store: store, Now: time.Now}
	caseRecord, e := clinic.CreateCase(ctx, actors[0], domain.CaseInput{PatientID: source.ID, DoctorID: actors[1].ID})
	if e != nil {
		t.Fatal(e)
	}
	var encounterID string
	e = db.QueryRow(ctx, `INSERT INTO encounter(kind,case_id,patient_id,doctor_id,admitted_at,status,discharged_at,discharge_summary,created_by,request_key) VALUES('opd',$1,$2,$3,now()-interval '1 hour','discharged',now(),'Retained summary',$4,'merge-fixture-encounter') RETURNING id`, caseRecord.ID, source.ID, actors[1].ID, actors[0].ID).Scan(&encounterID)
	if e != nil {
		t.Fatal(e)
	}
	billing := application.Billing{Store: store, Now: time.Now}
	account, e := billing.CreateAccount(ctx, actors[0], "Merge fixture charge")
	if e != nil {
		t.Fatal(e)
	}
	invoice, e := billing.CreateInvoice(ctx, actors[0], domain.InvoiceInput{PatientID: source.ID, InvoiceDate: time.Now().In(domain.HospitalLocation).Format("2006-01-02"), Lines: []domain.InvoiceLine{{AccountID: account.ID, Quantity: 1, UnitPriceMinor: 100}}}, "merge-fixture-invoice")
	if e != nil {
		t.Fatal(e)
	}
	merge := domain.MergePatientInput{PrimaryPatientID: primary.ID, MergedPatientID: source.ID, Reason: "Verified duplicate identity"}
	var wg sync.WaitGroup
	for n := 0; n < 4; n++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if e := app.MergePatients(ctx, actors[0], merge); e != nil {
				t.Errorf("concurrent merge: %v", e)
			}
		}()
	}
	wg.Wait()
	var originalID, canonical string
	if e = db.QueryRow(ctx, `SELECT patient_id::text,canonical_patient_id(patient_id)::text FROM encounter WHERE id=$1`, encounterID).Scan(&originalID, &canonical); e != nil {
		t.Fatal(e)
	}
	if originalID != source.ID || canonical != primary.ID {
		t.Fatal("historical encounter identity rewritten or not resolved")
	}
	resolved, e := store.PatientProfile(ctx, owner, source.ID)
	if e != nil || resolved.ID != primary.ID {
		t.Fatalf("alias profile resolution: %v", e)
	}
	historical, e := billing.Invoice(ctx, owner, invoice.ID)
	if e != nil || historical.PatientID != source.ID {
		t.Fatalf("retained invoice access: %v", e)
	}
	if e = db.QueryRow(ctx, `SELECT count(*) FROM patient_merge_event WHERE merged_patient_id=$1`, source.ID).Scan(&events); e != nil || events != 1 {
		t.Fatalf("duplicate merge audit: %d %v", events, e)
	}
	if _, e = clinic.CreateCase(ctx, actors[0], domain.CaseInput{PatientID: source.ID, DoctorID: actors[1].ID}); e == nil {
		t.Fatal("new work allowed against retired identity")
	}
	// Existing distinct portal owners may never be silently consolidated.
	merge.PrimaryPatientID = patients[0].ID
	merge.MergedPatientID = primary.ID
	if e = app.MergePatients(ctx, actors[0], merge); !errors.Is(e, domain.ErrStale) {
		t.Fatalf("portal ownership conflict accepted: %v", e)
	}

}
