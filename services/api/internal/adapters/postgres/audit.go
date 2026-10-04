package postgres

import (
	"context"
	"hms.local/api/internal/domain"
)

func (s Store) AuditEvents(ctx context.Context, a domain.Actor, f domain.AuditFilter) ([]domain.AuditEvent, error) {
	if a.Role != "admin" {
		return nil, domain.ErrForbidden
	}
	tx, e := s.DB.Begin(ctx)
	if e != nil {
		return nil, e
	}
	defer tx.Rollback(ctx)
	rows, e := tx.Query(ctx, `SELECT id,actor_id,action,resource_id,created_at FROM audit_event WHERE ($1::bigint=0 OR id<$1) AND ($2='' OR actor_id=$2) AND ($3='' OR action=$3) AND ($4='' OR resource_id=$4) ORDER BY id DESC LIMIT 50`, f.Before, f.ActorID, f.Action, f.ResourceID)
	if e != nil {
		return nil, e
	}
	result := []domain.AuditEvent{}
	for rows.Next() {
		var event domain.AuditEvent
		if e = rows.Scan(&event.ID, &event.ActorID, &event.Action, &event.ResourceID, &event.CreatedAt); e != nil {
			rows.Close()
			return nil, e
		}
		result = append(result, event)
	}
	e = rows.Err()
	rows.Close()
	if e != nil {
		return nil, e
	}
	// Record the review without persisting sensitive arbitrary filter input.
	if _, e = tx.Exec(ctx, `INSERT INTO audit_event(actor_id,action) VALUES($1,'audit.reviewed')`, a.ID); e != nil {
		return nil, e
	}
	return result, tx.Commit(ctx)
}
