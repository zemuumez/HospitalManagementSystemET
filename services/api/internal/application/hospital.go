package application

import (
	"context"
	"hms.local/api/internal/domain"
	"time"
)

type Repository interface {
	Patients(context.Context, domain.Actor, string, int) ([]domain.Patient, error)
	RegisterPatient(context.Context, domain.Actor, domain.PatientInput) (domain.Patient, error)
	Overview(context.Context, domain.Actor) (domain.Overview, error)
	Enqueue(context.Context, domain.Actor, domain.MessageInput, string) (domain.Message, error)
	Messages(context.Context) ([]domain.Message, error)
}
type Hospital struct {
	Store Repository
	Now   func() time.Time
}

func (h Hospital) Patients(ctx context.Context, a domain.Actor, search string, page int) ([]domain.Patient, error) {
	if !a.Can("patients.read") {
		return nil, domain.ErrForbidden
	}
	if page < 1 || page > 1000 || len(search) > 80 {
		return nil, domain.ErrValidation
	}
	return h.Store.Patients(ctx, a, search, page)
}
func (h Hospital) Register(ctx context.Context, a domain.Actor, p domain.PatientInput) (domain.Patient, error) {
	if !a.Can("patients.create") {
		return domain.Patient{}, domain.ErrForbidden
	}
	if err := p.Validate(h.Now()); err != nil {
		return domain.Patient{}, err
	}
	return h.Store.RegisterPatient(ctx, a, p)
}
func (h Hospital) Overview(ctx context.Context, a domain.Actor) (domain.Overview, error) {
	if !a.Can("patients.read") {
		return domain.Overview{}, nil
	}
	return h.Store.Overview(ctx, a)
}
func (h Hospital) Enqueue(ctx context.Context, a domain.Actor, m domain.MessageInput, key string) (domain.Message, error) {
	if !a.Can("messages.manage") {
		return domain.Message{}, domain.ErrForbidden
	}
	if len(key) < 16 || len(key) > 80 {
		return domain.Message{}, domain.ErrValidation
	}
	if err := m.Validate(); err != nil {
		return domain.Message{}, err
	}
	return h.Store.Enqueue(ctx, a, m, key)
}
func (h Hospital) Messages(ctx context.Context, a domain.Actor) ([]domain.Message, error) {
	if !a.Can("messages.manage") {
		return nil, domain.ErrForbidden
	}
	return h.Store.Messages(ctx)
}
