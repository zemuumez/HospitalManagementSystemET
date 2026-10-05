package postgres

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5/pgxpool"
	"hms.local/api/internal/application"
	"hms.local/api/internal/domain"
	"testing"
)

func testClinicalCareScope(t *testing.T, db *pgxpool.Pool, store Store, actors []domain.Actor, patients []domain.Patient) {
	ctx := context.Background()
	var encounterID string
	if e := db.QueryRow(ctx, `SELECT id FROM encounter WHERE patient_id=$1 LIMIT 1`, patients[1].ID).Scan(&encounterID); e != nil {
		t.Fatal(e)
	}
	app := application.ClinicalCareService{Store: store}
	for _, actor := range []domain.Actor{actors[2], actors[3], {ID: "unassigned-nurse", Role: "nurse"}} {
		checks := map[string]func() error{
			"diagnoses":   func() error { _, e := app.Diagnoses(ctx, actor, encounterID); return e },
			"procedures":  func() error { _, e := app.Procedures(ctx, actor, encounterID); return e },
			"care team":   func() error { _, e := app.CareTeam(ctx, actor, encounterID); return e },
			"attachments": func() error { _, e := app.Attachments(ctx, actor, encounterID); return e },
			"beds":        func() error { _, e := app.ListBedAssignments(ctx, actor, encounterID); return e },
			"odontogram":  func() error { _, e := app.Odontogram(ctx, actor, patients[1].ID); return e },
			"follow ups":  func() error { _, e := app.OPDFollowUps(ctx, actor, patients[1].ID); return e },
			"referrals":   func() error { _, e := app.PatientReferrals(ctx, actor, patients[1].ID); return e },
		}
		for name, check := range checks {
			t.Run(actor.Role+"/"+name, func(t *testing.T) {
				if e := check(); !errors.Is(e, domain.ErrForbidden) {
					t.Fatalf("unrelated record accessible: %v", e)
				}
			})
		}
	}
	if _, e := app.Diagnoses(ctx, actors[0], encounterID); e != nil {
		t.Fatalf("admin read denied: %v", e)
	}
	if _, e := app.AddDiagnosis(ctx, actors[2], encounterID, domain.AddDiagnosisInput{Description: "Unauthorized diagnosis"}); !errors.Is(e, domain.ErrForbidden) {
		t.Fatalf("unassigned doctor wrote diagnosis: %v", e)
	}
	if _, e := app.AddDiagnosis(ctx, actors[1], encounterID, domain.AddDiagnosisInput{Description: "Synthetic diagnosis"}); e != nil {
		t.Fatalf("assigned doctor denied: %v", e)
	}
	var active string
	if e := db.QueryRow(ctx, `SELECT id FROM encounter WHERE patient_id=$1 AND status='active' LIMIT 1`, patients[1].ID).Scan(&active); e != nil {
		t.Fatal(e)
	}
	if _, e := app.AddCareTeamMember(ctx, actors[0], active, domain.AddCareTeamMemberInput{StaffID: actors[3].ID, RoleTitle: "Forged clinician"}); e == nil {
		t.Fatal("patient assigned as clinician")
	}
	member, e := app.AddCareTeamMember(ctx, actors[0], active, domain.AddCareTeamMemberInput{StaffID: actors[2].ID, RoleTitle: "Covering doctor"})
	if e != nil {
		t.Fatal(e)
	}
	if _, e = app.Diagnoses(ctx, actors[2], active); e != nil {
		t.Fatal("delegated clinician denied", e)
	}
	if e = app.RevokeCareTeamMember(ctx, actors[0], member.ID); e != nil {
		t.Fatal(e)
	}
	if _, e = app.Diagnoses(ctx, actors[2], active); !errors.Is(e, domain.ErrForbidden) {
		t.Fatal("revoked delegation retained access", e)
	}
	if e := store.AuthorizeClinicalRecord(ctx, actors[0], patients[0].ID, encounterID, "clinical"); !errors.Is(e, domain.ErrForbidden) {
		t.Fatalf("mismatched patient and encounter accepted: %v", e)
	}
}
