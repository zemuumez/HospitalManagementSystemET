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

func testAddenda(t *testing.T, db *pgxpool.Pool, s Store, actors []domain.Actor, cases []domain.Case) {
	t.Helper()
	ctx := context.Background()
	app := application.Clinical{Store: s, Now: time.Now}
	enc, e := app.Admit(ctx, actors[0], domain.EncounterInput{Kind: "opd", CaseID: cases[0].ID, AdmittedAt: time.Now().Add(-time.Hour)}, "addendum-admission")
	if e != nil {
		t.Fatal(e)
	}
	note, e := app.SignNote(ctx, actors[1], enc.ID, domain.NoteInput{Body: "Original signed observation"}, "addendum-note-original")
	if e != nil {
		t.Fatal(e)
	}
	i := domain.AddendumInput{Kind: "note", NoteID: note.ID, Body: "Corrected observation", Reason: "Transcription correction"}
	out, e := app.AddAddendum(ctx, actors[1], enc.ID, i, "addendum-note-correction")
	if e != nil {
		t.Fatal(e)
	}
	retry, e := app.AddAddendum(ctx, actors[1], enc.ID, i, "addendum-note-correction")
	if e != nil || retry.ID != out.ID {
		t.Fatal(e)
	}
	changed := i
	changed.Body = "Changed retry"
	if _, e = app.AddAddendum(ctx, actors[1], enc.ID, changed, "addendum-note-correction"); !errors.Is(e, domain.ErrConflict) {
		t.Fatal(e)
	}
	if _, e = app.AddAddendum(ctx, actors[2], enc.ID, i, "addendum-forged-doctor"); !errors.Is(e, domain.ErrNotFound) {
		t.Fatal(e)
	}
	if _, e = app.AddAddendum(ctx, actors[0], enc.ID, i, "addendum-forged-admin"); !errors.Is(e, domain.ErrForbidden) {
		t.Fatal(e)
	}
	discharge := domain.AddendumInput{Kind: "discharge", Body: "Follow-up clarification", Reason: "Clarify signed discharge instructions"}
	if _, e = app.AddAddendum(ctx, actors[1], enc.ID, discharge, "addendum-discharge-correction"); !errors.Is(e, domain.ErrStale) {
		t.Fatal(e)
	}
	if _, e = app.Discharge(ctx, actors[1], enc.ID, domain.Discharge{Version: 1, Summary: "Original signed discharge"}); e != nil {
		t.Fatal(e)
	}
	var wg sync.WaitGroup
	ids := make(chan string, 2)
	errs := make(chan error, 2)
	for n := 0; n < 2; n++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			v, err := app.AddAddendum(ctx, actors[1], enc.ID, discharge, "addendum-discharge-correction")
			ids <- v.ID
			errs <- err
		}()
	}
	wg.Wait()
	close(ids)
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	first := ""
	for id := range ids {
		if first != "" && first != id {
			t.Fatal("duplicate correction")
		}
		first = id
	}
	list, e := app.Addenda(ctx, actors[3], enc.ID, 1)
	if e != nil || len(list) != 2 {
		t.Fatal(list, e)
	}
	if _, e = app.Addenda(ctx, actors[2], enc.ID, 1); !errors.Is(e, domain.ErrNotFound) {
		t.Fatal(e)
	}
	if _, e = app.Addenda(ctx, actors[4], enc.ID, 1); !errors.Is(e, domain.ErrForbidden) {
		t.Fatal(e)
	}
	other, e := app.Admit(ctx, actors[0], domain.EncounterInput{Kind: "opd", CaseID: cases[1].ID, AdmittedAt: time.Now().Add(-time.Hour)}, "addendum-other-encounter")
	if e != nil {
		t.Fatal(e)
	}
	if _, e = app.AddAddendum(ctx, actors[1], other.ID, i, "addendum-cross-note"); !errors.Is(e, domain.ErrNotFound) {
		t.Fatal("cross-encounter note", e)
	}
	for _, sql := range []string{`UPDATE clinical_addendum SET body='tampered'`, `DELETE FROM clinical_addendum`, `UPDATE encounter SET discharge_summary='tampered' WHERE status='discharged'`, `UPDATE encounter SET status='active',discharged_at=NULL WHERE status='discharged'`} {
		if _, e = db.Exec(ctx, sql); e == nil {
			t.Fatal("retained record changed", sql)
		}
	}
	var summary string
	if e = db.QueryRow(ctx, `SELECT discharge_summary FROM encounter WHERE id=$1`, enc.ID).Scan(&summary); e != nil || summary != "Original signed discharge" {
		t.Fatal(summary, e)
	}
}
