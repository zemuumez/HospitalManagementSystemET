package application

import (
	"context"
	"hms.local/api/internal/domain"
	"strings"
)

type AuditRepository interface {
	AuditEvents(context.Context, domain.Actor, domain.AuditFilter) ([]domain.AuditEvent, error)
}
type Audit struct{ Store AuditRepository }

func (s Audit) Events(ctx context.Context, a domain.Actor, f domain.AuditFilter) ([]domain.AuditEvent, error) {
	if a.Role != "admin" {
		return nil, domain.ErrForbidden
	}
	if f.Before < 0 || len(f.ActorID) > 128 || len(f.Action) > 128 || len(f.ResourceID) > 128 || strings.ContainsAny(f.ActorID+f.Action+f.ResourceID, "\x00\r\n") {
		return nil, domain.ErrValidation
	}
	return s.Store.AuditEvents(ctx, a, f)
}
