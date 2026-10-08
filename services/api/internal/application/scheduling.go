package application

import (
	"context"
	"hms.local/api/internal/domain"
	"strings"
	"time"
)

type SchedulingRepository interface {
	AppointmentStatusHistory(context.Context, domain.Actor, string, int) ([]domain.AppointmentStatusEvent, error)
	Absences(context.Context, domain.Actor, string, int) ([]domain.DoctorAbsence, error)
	CreateAbsence(context.Context, domain.Actor, domain.AbsenceInput) (domain.DoctorAbsence, error)
	CancelAbsence(context.Context, domain.Actor, string, domain.AbsenceCancel) (domain.DoctorAbsence, error)
	Reschedule(context.Context, domain.Actor, string, domain.RescheduleInput, time.Time) (domain.Appointment, error)
	RescheduleHistory(context.Context, domain.Actor, string, int) ([]domain.RescheduleEvent, error)
	Doctors(context.Context) ([]domain.Doctor, error)
	Doctor(context.Context, string) (domain.Doctor, error)
	CreateDoctorProfile(context.Context, domain.Actor, domain.CreateDoctorInput) (domain.Doctor, error)
	UpdateDoctorProfile(context.Context, domain.Actor, string, domain.UpdateDoctorInput) (domain.Doctor, error)
	SetDoctorStatus(context.Context, domain.Actor, string, bool) error
	DeleteDoctor(context.Context, domain.Actor, string) error
	ListDoctors(context.Context, string, string, string) ([]domain.Doctor, error)
	SaveDoctor(context.Context, domain.Actor, domain.Doctor) (domain.Doctor, error)
	Slots(context.Context, string, time.Time, time.Time) ([]time.Time, error)
	Appointments(context.Context, domain.Actor, int) ([]domain.Appointment, error)
	Book(context.Context, domain.Actor, domain.AppointmentInput, string, time.Time) (domain.Appointment, error)
	ChangeAppointment(context.Context, domain.Actor, string, domain.AppointmentChange, time.Time) (domain.Appointment, error)
	DoctorSchedules(context.Context, string) ([]domain.DoctorSchedule, error)
	DoctorSchedule(context.Context, string) (domain.DoctorSchedule, error)
	SaveDoctorSchedule(context.Context, domain.Actor, domain.SaveDoctorScheduleInput) (domain.DoctorSchedule, error)
	DeleteDoctorSchedule(context.Context, domain.Actor, string) error
	DoctorHolidays(context.Context, string) ([]domain.DoctorHoliday, error)
	CreateDoctorHoliday(context.Context, domain.Actor, domain.CreateDoctorHolidayInput) (domain.DoctorHoliday, error)
	DeleteDoctorHoliday(context.Context, domain.Actor, string) error
	DoctorBreaks(context.Context, string) ([]domain.DoctorLunchBreak, error)
	CreateDoctorBreak(context.Context, domain.Actor, domain.CreateDoctorBreakInput) (domain.DoctorLunchBreak, error)
	DeleteDoctorBreak(context.Context, domain.Actor, string) error
	DoctorOPDCharges(context.Context, string) ([]domain.DoctorOPDCharge, error)
	DoctorOPDCharge(context.Context, string) (domain.DoctorOPDCharge, error)
	SaveDoctorOPDCharge(context.Context, domain.Actor, domain.SaveDoctorOPDChargeInput) (domain.DoctorOPDCharge, error)
	DeleteDoctorOPDCharge(context.Context, domain.Actor, string) error
}
type Scheduling struct {
	Store SchedulingRepository
	Now   func() time.Time
}

func (s Scheduling) Doctors(ctx context.Context, a domain.Actor) ([]domain.Doctor, error) {
	if !a.Can("appointments.read") {
		return nil, domain.ErrForbidden
	}
	docs, err := s.Store.Doctors(ctx)
	if err != nil {
		return nil, err
	}
	if a.Role != "admin" {
		for i := range docs {
			if a.Role == "doctor" && a.ID == docs[i].ID {
				continue
			}
			docs[i] = docs[i].PublicDirectory(a.Role)
		}
	}
	return docs, nil
}

func (s Scheduling) Doctor(ctx context.Context, a domain.Actor, id string) (domain.Doctor, error) {
	if a.Role == "doctor" && a.ID != id {
		return domain.Doctor{}, domain.ErrForbidden
	}
	if a.Role != "admin" && a.Role != "receptionist" && a.Role != "nurse" && a.Role != "doctor" {
		return domain.Doctor{}, domain.ErrForbidden
	}
	return s.Store.Doctor(ctx, id)
}

func (s Scheduling) CreateDoctorProfile(ctx context.Context, a domain.Actor, input domain.CreateDoctorInput) (domain.Doctor, error) {
	if a.Role != "admin" {
		return domain.Doctor{}, domain.ErrForbidden
	}
	if err := input.Validate(); err != nil {
		return domain.Doctor{}, err
	}
	return s.Store.CreateDoctorProfile(ctx, a, input)
}

func (s Scheduling) UpdateDoctorProfile(ctx context.Context, a domain.Actor, id string, input domain.UpdateDoctorInput) (domain.Doctor, error) {
	if a.Role != "admin" && !(a.Role == "doctor" && a.ID == id) {
		return domain.Doctor{}, domain.ErrForbidden
	}
	if a.Role == "doctor" && input.DepartmentID != nil {
		return domain.Doctor{}, domain.ErrForbidden
	}
	if err := input.Validate(); err != nil {
		return domain.Doctor{}, err
	}
	return s.Store.UpdateDoctorProfile(ctx, a, id, input)
}

func (s Scheduling) SetDoctorStatus(ctx context.Context, a domain.Actor, id string, active bool) error {
	if a.Role != "admin" {
		return domain.ErrForbidden
	}
	if a.ID == id && !active {
		return domain.ErrConflict
	}
	return s.Store.SetDoctorStatus(ctx, a, id, active)
}

func (s Scheduling) DeleteDoctor(ctx context.Context, a domain.Actor, id string) error {
	if a.Role != "admin" {
		return domain.ErrForbidden
	}
	return s.Store.DeleteDoctor(ctx, a, id)
}

func (s Scheduling) ListDoctors(ctx context.Context, a domain.Actor, statusFilter string, deptFilter string, search string) ([]domain.Doctor, error) {
	if !a.Can("appointments.read") {
		return nil, domain.ErrForbidden
	}
	if a.Role != "admin" || (statusFilter != "inactive" && statusFilter != "all") {
		statusFilter = "active"
	}
	docs, err := s.Store.ListDoctors(ctx, statusFilter, deptFilter, search)
	if err != nil {
		return nil, err
	}
	if a.Role != "admin" {
		for i := range docs {
			if a.Role == "doctor" && a.ID == docs[i].ID {
				continue
			}
			docs[i] = docs[i].PublicDirectory(a.Role)
		}
	}
	return docs, nil
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
	if !domain.UUIDPattern.MatchString(id) || c.Version < 1 || c.Version > 1000000000 || len([]rune(strings.TrimSpace(c.Reason))) > 1000 || strings.ContainsRune(c.Reason, 0) {
		return domain.Appointment{}, domain.ErrValidation
	}
	c.Reason = strings.TrimSpace(c.Reason)
	return s.Store.ChangeAppointment(ctx, a, id, c, s.Now())
}

func (s Scheduling) Absences(ctx context.Context, a domain.Actor, doctorID string, page int) ([]domain.DoctorAbsence, error) {
	if a.Role != "admin" && !(a.Role == "doctor" && a.ID == doctorID) {
		return nil, domain.ErrForbidden
	}
	if (doctorID == "" && a.Role != "admin") || len(doctorID) > 128 || page < 1 || page > 1000 {
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

func (s Scheduling) StatusHistory(ctx context.Context, a domain.Actor, id string, page int) ([]domain.AppointmentStatusEvent, error) {
	if !a.Can("appointments.read") {
		return nil, domain.ErrForbidden
	}
	if !domain.UUIDPattern.MatchString(id) || page < 1 || page > 1000 {
		return nil, domain.ErrValidation
	}
	return s.Store.AppointmentStatusHistory(ctx, a, id, page)
}

func (s Scheduling) DoctorSchedules(ctx context.Context, a domain.Actor, doctorID string) ([]domain.DoctorSchedule, error) {
	if a.Role == "doctor" {
		if doctorID != "" && doctorID != a.ID {
			return nil, domain.ErrForbidden
		}
		doctorID = a.ID
	} else if a.Role != "admin" && a.Role != "receptionist" && a.Role != "nurse" {
		return nil, domain.ErrForbidden
	}
	return s.Store.DoctorSchedules(ctx, doctorID)
}

func (s Scheduling) DoctorSchedule(ctx context.Context, a domain.Actor, doctorID string) (domain.DoctorSchedule, error) {
	if a.Role == "doctor" && doctorID != a.ID {
		return domain.DoctorSchedule{}, domain.ErrForbidden
	} else if a.Role != "admin" && a.Role != "receptionist" && a.Role != "nurse" && a.Role != "doctor" {
		return domain.DoctorSchedule{}, domain.ErrForbidden
	}
	return s.Store.DoctorSchedule(ctx, doctorID)
}

func (s Scheduling) SaveDoctorSchedule(ctx context.Context, a domain.Actor, input domain.SaveDoctorScheduleInput) (domain.DoctorSchedule, error) {
	if a.Role == "doctor" && input.DoctorID != a.ID {
		return domain.DoctorSchedule{}, domain.ErrForbidden
	} else if a.Role != "admin" && a.Role != "doctor" {
		return domain.DoctorSchedule{}, domain.ErrForbidden
	}
	if err := input.Validate(); err != nil {
		return domain.DoctorSchedule{}, err
	}
	return s.Store.SaveDoctorSchedule(ctx, a, input)
}

func (s Scheduling) DeleteDoctorSchedule(ctx context.Context, a domain.Actor, doctorID string) error {
	if a.Role != "admin" {
		return domain.ErrForbidden
	}
	return s.Store.DeleteDoctorSchedule(ctx, a, doctorID)
}

func (s Scheduling) DoctorHolidays(ctx context.Context, a domain.Actor, doctorID string) ([]domain.DoctorHoliday, error) {
	if a.Role == "doctor" {
		if doctorID != "" && doctorID != a.ID {
			return nil, domain.ErrForbidden
		}
		doctorID = a.ID
	} else if a.Role != "admin" && a.Role != "receptionist" && a.Role != "nurse" {
		return nil, domain.ErrForbidden
	}
	return s.Store.DoctorHolidays(ctx, doctorID)
}

func (s Scheduling) CreateDoctorHoliday(ctx context.Context, a domain.Actor, input domain.CreateDoctorHolidayInput) (domain.DoctorHoliday, error) {
	if a.Role == "doctor" && input.DoctorID != a.ID {
		return domain.DoctorHoliday{}, domain.ErrForbidden
	} else if a.Role != "admin" && a.Role != "doctor" {
		return domain.DoctorHoliday{}, domain.ErrForbidden
	}
	if err := input.Validate(); err != nil {
		return domain.DoctorHoliday{}, err
	}
	return s.Store.CreateDoctorHoliday(ctx, a, input)
}

func (s Scheduling) DeleteDoctorHoliday(ctx context.Context, a domain.Actor, id string) error {
	if a.Role != "admin" && a.Role != "doctor" {
		return domain.ErrForbidden
	}
	return s.Store.DeleteDoctorHoliday(ctx, a, id)
}

func (s Scheduling) DoctorBreaks(ctx context.Context, a domain.Actor, doctorID string) ([]domain.DoctorLunchBreak, error) {
	if a.Role == "doctor" {
		if doctorID != "" && doctorID != a.ID {
			return nil, domain.ErrForbidden
		}
		doctorID = a.ID
	} else if a.Role != "admin" && a.Role != "receptionist" && a.Role != "nurse" {
		return nil, domain.ErrForbidden
	}
	return s.Store.DoctorBreaks(ctx, doctorID)
}

func (s Scheduling) CreateDoctorBreak(ctx context.Context, a domain.Actor, input domain.CreateDoctorBreakInput) (domain.DoctorLunchBreak, error) {
	if a.Role == "doctor" && input.DoctorID != a.ID {
		return domain.DoctorLunchBreak{}, domain.ErrForbidden
	} else if a.Role != "admin" && a.Role != "doctor" {
		return domain.DoctorLunchBreak{}, domain.ErrForbidden
	}
	if err := input.Validate(); err != nil {
		return domain.DoctorLunchBreak{}, err
	}
	return s.Store.CreateDoctorBreak(ctx, a, input)
}

func (s Scheduling) DeleteDoctorBreak(ctx context.Context, a domain.Actor, id string) error {
	if a.Role != "admin" && a.Role != "doctor" {
		return domain.ErrForbidden
	}
	return s.Store.DeleteDoctorBreak(ctx, a, id)
}

func (s Scheduling) DoctorOPDCharges(ctx context.Context, a domain.Actor, search string) ([]domain.DoctorOPDCharge, error) {
	if !a.Can("appointments.read") && a.Role != "accountant" {
		return nil, domain.ErrForbidden
	}
	return s.Store.DoctorOPDCharges(ctx, search)
}

func (s Scheduling) DoctorOPDCharge(ctx context.Context, a domain.Actor, doctorID string) (domain.DoctorOPDCharge, error) {
	if a.Role == "doctor" && doctorID != a.ID {
		return domain.DoctorOPDCharge{}, domain.ErrForbidden
	} else if !a.Can("appointments.read") && a.Role != "accountant" {
		return domain.DoctorOPDCharge{}, domain.ErrForbidden
	}
	return s.Store.DoctorOPDCharge(ctx, doctorID)
}

func (s Scheduling) SaveDoctorOPDCharge(ctx context.Context, a domain.Actor, input domain.SaveDoctorOPDChargeInput) (domain.DoctorOPDCharge, error) {
	if a.Role != "admin" {
		return domain.DoctorOPDCharge{}, domain.ErrForbidden
	}
	if err := input.Validate(); err != nil {
		return domain.DoctorOPDCharge{}, err
	}
	return s.Store.SaveDoctorOPDCharge(ctx, a, input)
}

func (s Scheduling) DeleteDoctorOPDCharge(ctx context.Context, a domain.Actor, doctorID string) error {
	if a.Role != "admin" {
		return domain.ErrForbidden
	}
	return s.Store.DeleteDoctorOPDCharge(ctx, a, doctorID)
}

