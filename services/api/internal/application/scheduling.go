package application

import (
	"context"
	"hms.local/api/internal/domain"
	"strings"
	"time"
)

type SchedulingRepository interface {
	Absences(context.Context, domain.Actor, string, int) ([]domain.DoctorAbsence, error)
	CreateAbsence(context.Context, domain.Actor, domain.AbsenceInput) (domain.DoctorAbsence, error)
	CancelAbsence(context.Context, domain.Actor, string, domain.AbsenceCancel) (domain.DoctorAbsence, error)
	Reschedule(context.Context, domain.Actor, string, domain.RescheduleInput, time.Time) (domain.Appointment, error)
	RescheduleHistory(context.Context, domain.Actor, string, int) ([]domain.RescheduleEvent, error)
	Doctors(context.Context) ([]domain.Doctor, error)
	SaveDoctor(context.Context, domain.Actor, domain.Doctor) (domain.Doctor, error)
	Slots(context.Context, string, time.Time, time.Time) ([]time.Time, error)
	Appointments(context.Context, domain.Actor, int) ([]domain.Appointment, error)
	Book(context.Context, domain.Actor, domain.AppointmentInput, string, time.Time) (domain.Appointment, error)
	ChangeAppointment(context.Context, domain.Actor, string, domain.AppointmentChange, time.Time) (domain.Appointment, error)
}
type Scheduling struct {
	Store SchedulingRepository
	Now   func() time.Time
}

func (s Scheduling) Doctors(ctx context.Context, a domain.Actor) ([]domain.Doctor, error) {
	if !a.Can("appointments.read") {
		return nil, domain.ErrForbidden
	}
	return s.Store.Doctors(ctx)
}
func (s Scheduling) SaveDoctor(ctx context.Context, a domain.Actor, d domain.Doctor) (domain.Doctor, error) {
	if a.Role != "admin" && !(a.Role == "doctor" && a.ID == d.ID) {
		return d, domain.ErrForbidden
	}
	if err := d.Validate(); err != nil {
		return d, err
	}
	return s.Store.SaveDoctor(ctx, a, d)
}
func (s Scheduling) Slots(ctx context.Context, a domain.Actor, id, date string) ([]time.Time, error) {
	if !a.Can("appointments.read") {
		return nil, domain.ErrForbidden
	}
	day, err := time.ParseInLocation("2006-01-02", date, domain.HospitalLocation)
	if err != nil || id == "" || len(id) > 128 || day.Before(s.Now().In(domain.HospitalLocation).Truncate(24*time.Hour).Add(-24*time.Hour)) || day.After(s.Now().AddDate(1, 0, 0)) {
		return nil, domain.ErrValidation
	}
	return s.Store.Slots(ctx, id, day, s.Now())
}
func (s Scheduling) Appointments(ctx context.Context, a domain.Actor, page int) ([]domain.Appointment, error) {
	if !a.Can("appointments.read") {
		return nil, domain.ErrForbidden
	}
	if page < 1 || page > 1000 {
		return nil, domain.ErrValidation
	}
	return s.Store.Appointments(ctx, a, page)
}
func (s Scheduling) Book(ctx context.Context, a domain.Actor, i domain.AppointmentInput, key string) (domain.Appointment, error) {
	if !a.Can("appointments.book") {
		return domain.Appointment{}, domain.ErrForbidden
	}
	if len(key) < 16 || len(key) > 80 {
		return domain.Appointment{}, domain.ErrValidation
	}
	if err := i.Validate(s.Now()); err != nil {
		return domain.Appointment{}, err
	}
	return s.Store.Book(ctx, a, i, key, s.Now())
}
func (s Scheduling) Change(ctx context.Context, a domain.Actor, id string, c domain.AppointmentChange) (domain.Appointment, error) {
	if !a.Can("appointments.read") {
		return domain.Appointment{}, domain.ErrForbidden
	}
	if !domain.UUIDPattern.MatchString(id) || c.Version < 1 {
		return domain.Appointment{}, domain.ErrValidation
	}
	return s.Store.ChangeAppointment(ctx, a, id, c, s.Now())
}

func (s Scheduling) Absences(ctx context.Context, a domain.Actor, doctorID string, page int) ([]domain.DoctorAbsence, error) {
	if a.Role != "admin" && !(a.Role == "doctor" && a.ID == doctorID) {
		return nil, domain.ErrForbidden
	}
	if doctorID == "" || len(doctorID) > 128 || page < 1 || page > 1000 {
		return nil, domain.ErrValidation
	}
	return s.Store.Absences(ctx, a, doctorID, page)
}
func (s Scheduling) CreateAbsence(ctx context.Context, a domain.Actor, i domain.AbsenceInput) (domain.DoctorAbsence, error) {
	if a.Role != "admin" && !(a.Role == "doctor" && a.ID == i.DoctorID) {
		return domain.DoctorAbsence{}, domain.ErrForbidden
	}
	i.Reason = strings.TrimSpace(i.Reason)
	i.StartsAt = i.StartsAt.UTC().Truncate(time.Microsecond)
	i.EndsAt = i.EndsAt.UTC().Truncate(time.Microsecond)
	if i.DoctorID == "" || len(i.DoctorID) > 128 || len([]rune(i.Reason)) < 1 || len([]rune(i.Reason)) > 200 || !i.StartsAt.After(s.Now()) || !i.EndsAt.After(i.StartsAt) || i.EndsAt.After(s.Now().AddDate(1, 0, 0)) {
		return domain.DoctorAbsence{}, domain.ErrValidation
	}
	return s.Store.CreateAbsence(ctx, a, i)
}
func (s Scheduling) CancelAbsence(ctx context.Context, a domain.Actor, id string, i domain.AbsenceCancel) (domain.DoctorAbsence, error) {
	if a.Role != "admin" && a.Role != "doctor" {
		return domain.DoctorAbsence{}, domain.ErrForbidden
	}
	i.Reason = strings.TrimSpace(i.Reason)
	if !domain.UUIDPattern.MatchString(id) || i.Version < 1 || len([]rune(i.Reason)) < 1 || len([]rune(i.Reason)) > 200 {
		return domain.DoctorAbsence{}, domain.ErrValidation
	}
	return s.Store.CancelAbsence(ctx, a, id, i)
}
func (s Scheduling) Reschedule(ctx context.Context, a domain.Actor, id string, i domain.RescheduleInput) (domain.Appointment, error) {
	if !a.Can("appointments.book") {
		return domain.Appointment{}, domain.ErrForbidden
	}
	i.StartsAt = i.StartsAt.UTC().Truncate(time.Microsecond)
	i.Reason = strings.TrimSpace(i.Reason)
	if !domain.UUIDPattern.MatchString(id) || i.Version < 1 || len([]rune(i.Reason)) < 1 || len([]rune(i.Reason)) > 2000 || !i.StartsAt.After(s.Now()) || i.StartsAt.After(s.Now().AddDate(1, 0, 0)) {
		return domain.Appointment{}, domain.ErrValidation
	}
	return s.Store.Reschedule(ctx, a, id, i, s.Now())
}
func (s Scheduling) RescheduleHistory(ctx context.Context, a domain.Actor, id string, page int) ([]domain.RescheduleEvent, error) {
	if !a.Can("appointments.read") {
		return nil, domain.ErrForbidden
	}
	if !domain.UUIDPattern.MatchString(id) || page < 1 || page > 1000 {
		return nil, domain.ErrValidation
	}
	return s.Store.RescheduleHistory(ctx, a, id, page)
}
