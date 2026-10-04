package postgres

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"hms.local/api/internal/domain"
)

func (s Store) LinkPatient(ctx context.Context, a domain.Actor, id string, access domain.PatientAccess) error {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	// Serialize with provisioning/disablement; no email or phone auto-claiming.
	if _, err = tx.Exec(ctx, `SELECT pg_advisory_xact_lock(72841022)`); err != nil {
		return err
	}
	for _, target := range []struct{ id, role string }{{access.UserID, "patient"}, {access.ClinicianID, "doctor"}} {
		if target.id == "" {
			continue
		}
		var ok bool
		err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM staff_access WHERE user_id=$1 AND role=$2 AND active)`, target.id, target.role).Scan(&ok)
		if err != nil {
			return err
		}
		if !ok {
			return domain.ErrValidation
		}
	}
	var current string
	err = tx.QueryRow(ctx, `SELECT COALESCE(user_id,'') FROM patient WHERE id=$1 FOR UPDATE`, id).Scan(&current)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrNotFound
	}
	if err != nil {
		return err
	}
	// Existing portal ownership cannot be reassigned by a routine edit.
	if current != "" && current != access.UserID {
		return domain.ErrStale
	}
	_, err = tx.Exec(ctx, `UPDATE patient SET user_id=NULLIF($2,''),clinician_user_id=NULLIF($3,'') WHERE id=$1`, id, access.UserID, access.ClinicianID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrStale
		}
		return err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO audit_event(actor_id,action,resource_id) VALUES($1,'patient.access_linked',$2)`, a.ID, id); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
