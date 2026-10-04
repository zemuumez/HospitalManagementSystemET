package postgres

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"hms.local/api/internal/domain"
)

const addendumFields = `id,encounter_id,COALESCE(note_id::text,''),kind,author_id,body,reason,signed_at`

func scanAddendum(row pgx.Row) (domain.Addendum, error) {
	var o domain.Addendum
	e := row.Scan(&o.ID, &o.EncounterID, &o.NoteID, &o.Kind, &o.AuthorID, &o.Body, &o.Reason, &o.SignedAt)
	return o, clinicalError(e)
}
func (s Store) AddAddendum(ctx context.Context, a domain.Actor, id string, i domain.AddendumInput, key string) (domain.Addendum, error) {
	out := domain.Addendum{}
	tx, e := s.DB.Begin(ctx)
	if e != nil {
		return out, e
	}
	defer tx.Rollback(ctx)
	if _, e = tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1,0))`, "addendum:"+a.ID+":"+key); e != nil {
		return out, e
	}
	// Always recheck record authorization, including idempotent retries.
	var status string
	if e = tx.QueryRow(ctx, `SELECT status FROM encounter WHERE id=$1 AND doctor_id=$2 FOR UPDATE`, id, a.ID).Scan(&status); e != nil {
		return out, clinicalError(e)
	}
	old, e := scanAddendum(tx.QueryRow(ctx, `SELECT `+addendumFields+` FROM clinical_addendum WHERE author_id=$1 AND request_key=$2`, a.ID, key))
	if e == nil {
		if old.EncounterID != id || old.AddendumInput != i {
			return out, domain.ErrConflict
		}
		return old, tx.Commit(ctx)
	}
	if !errors.Is(e, domain.ErrNotFound) {
		return out, e
	}
	if i.Kind == "discharge" && status != "discharged" {
		return out, domain.ErrStale
	}
	if i.Kind == "note" {
		var exists bool
		if e = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM clinical_note WHERE id=$1 AND encounter_id=$2)`, i.NoteID, id).Scan(&exists); e != nil {
			return out, e
		}
		if !exists {
			return out, domain.ErrNotFound
		}
	}
	out, e = scanAddendum(tx.QueryRow(ctx, `INSERT INTO clinical_addendum(encounter_id,note_id,kind,author_id,body,reason,request_key) VALUES($1,NULLIF($2,'')::uuid,$3,$4,$5,$6,$7) RETURNING `+addendumFields, id, i.NoteID, i.Kind, a.ID, i.Body, i.Reason, key))
	if e != nil {
		return out, e
	}
	if e = pharmacyAudit(ctx, tx, a, "clinical.addendum_signed", out.ID); e != nil {
		return out, e
	}
	return out, tx.Commit(ctx)
}
func (s Store) Addenda(ctx context.Context, a domain.Actor, id string, page int) ([]domain.Addendum, error) {
	tx, e := s.DB.Begin(ctx)
	if e != nil {
		return nil, e
	}
	defer tx.Rollback(ctx)
	var found string
	if e = tx.QueryRow(ctx, `SELECT e.id FROM encounter e JOIN patient p ON p.id=e.patient_id WHERE e.id=$3 AND ($1='admin' OR ($1='doctor' AND e.doctor_id=$2) OR ($1='patient' AND p.user_id=$2))`, a.Role, a.ID, id).Scan(&found); e != nil {
		return nil, clinicalError(e)
	}
	rows, e := tx.Query(ctx, `SELECT `+addendumFields+` FROM clinical_addendum WHERE encounter_id=$1 ORDER BY signed_at DESC,id LIMIT 25 OFFSET $2`, id, (page-1)*25)
	if e != nil {
		return nil, e
	}
	out := []domain.Addendum{}
	for rows.Next() {
		v, err := scanAddendum(rows)
		if err != nil {
			rows.Close()
			return nil, err
		}
		out = append(out, v)
	}
	e = rows.Err()
	rows.Close()
	if e != nil {
		return nil, e
	}
	if e = pharmacyAudit(ctx, tx, a, "clinical.addenda_viewed", id); e != nil {
		return nil, e
	}
	return out, tx.Commit(ctx)
}
