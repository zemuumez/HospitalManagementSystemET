package postgres

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5/pgxpool"
	"hms.local/api/internal/application"
	"hms.local/api/internal/domain"
	"testing"
)

func testMergePreservesHistory(t *testing.T, db *pgxpool.Pool, store Store, actors []domain.Actor, patients []domain.Patient) {
	ctx := context.Background()
	var before, after int
	if err := db.QueryRow(ctx, `SELECT count(*) FROM encounter WHERE patient_id=$1`, patients[1].ID).Scan(&before); err != nil {
		t.Fatal(err)
	}
	app := application.PatientExtensionsService{Store: store}
	err := app.MergePatients(ctx, actors[0], domain.MergePatientInput{PrimaryPatientID: patients[0].ID, MergedPatientID: patients[1].ID, Reason: "Review test"})
	if !errors.Is(err, domain.ErrUnavailable) {
		t.Fatalf("unsafe merge must be explicitly unavailable: %v", err)
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
}
