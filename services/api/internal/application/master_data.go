package application

import (
	"context"
	"hms.local/api/internal/domain"
	"strings"
)

// MasterDataRepository is implemented by the postgres store.
type MasterDataRepository interface {
	AuthorizePatientRecord(context.Context, domain.Actor, string) error
	// Doctor departments
	DoctorDepartments(ctx context.Context, includeArchived bool) ([]domain.DoctorDepartment, error)
	DoctorDepartment(ctx context.Context, id string) (domain.DoctorDepartment, error)
	CreateDoctorDepartment(ctx context.Context, a domain.Actor, i domain.DoctorDepartmentInput) (domain.DoctorDepartment, error)
	UpdateDoctorDepartment(ctx context.Context, a domain.Actor, id string, i domain.DoctorDepartmentInput) (domain.DoctorDepartment, error)
	ArchiveDoctorDepartment(ctx context.Context, a domain.Actor, id string, version int, reason string) (domain.DoctorDepartment, error)
	DoctorDepartmentRevisions(ctx context.Context, a domain.Actor, id string, page int) ([]domain.DoctorDepartmentRevision, error)

	// Doctor profile extensions
	DoctorExt(ctx context.Context, doctorID string) (domain.DoctorExt, error)
	SaveDoctorExt(ctx context.Context, a domain.Actor, doctorID string, i domain.DoctorExtInput) (domain.DoctorExt, error)

	// Hospital hours
	HospitalHours(ctx context.Context) ([]domain.HospitalHoursEntry, error)
	SaveHospitalHours(ctx context.Context, a domain.Actor, i domain.HospitalHoursInput) ([]domain.HospitalHoursEntry, error)

	// Hospital date overrides
	HospitalDateOverrides(ctx context.Context, page int) ([]domain.HospitalDateOverride, error)
	CreateHospitalDateOverride(ctx context.Context, a domain.Actor, i domain.HospitalDateOverrideInput) (domain.HospitalDateOverride, error)
	DeleteHospitalDateOverride(ctx context.Context, a domain.Actor, id string) error

	// Patient profile ext
	PatientProfileExt(ctx context.Context, patientID string) (domain.PatientProfileExt, error)
	SavePatientProfileExt(ctx context.Context, a domain.Actor, patientID string, i domain.PatientProfileExt) error
}

type MasterDataService struct {
	Store MasterDataRepository
}

// ─── Doctor departments ───────────────────────────────────────────────────────

func (s MasterDataService) DoctorDepartments(ctx context.Context, a domain.Actor, includeArchived bool) ([]domain.DoctorDepartment, error) {
	if !a.Can("appointments.read") && !a.Can("staff.manage") {
		return nil, domain.ErrForbidden
	}
	return s.Store.DoctorDepartments(ctx, includeArchived)
}

func (s MasterDataService) DoctorDepartment(ctx context.Context, a domain.Actor, id string) (domain.DoctorDepartment, error) {
	if !a.Can("appointments.read") && !a.Can("staff.manage") {
		return domain.DoctorDepartment{}, domain.ErrForbidden
	}
	if !domain.UUIDPattern.MatchString(id) {
		return domain.DoctorDepartment{}, domain.ErrValidation
	}
	return s.Store.DoctorDepartment(ctx, id)
}

func (s MasterDataService) CreateDoctorDepartment(ctx context.Context, a domain.Actor, i domain.DoctorDepartmentInput) (domain.DoctorDepartment, error) {
	if a.Role != "admin" {
		return domain.DoctorDepartment{}, domain.ErrForbidden
	}
	if err := i.Validate(true); err != nil {
		return domain.DoctorDepartment{}, err
	}
	return s.Store.CreateDoctorDepartment(ctx, a, i)
}

func (s MasterDataService) UpdateDoctorDepartment(ctx context.Context, a domain.Actor, id string, i domain.DoctorDepartmentInput) (domain.DoctorDepartment, error) {
	if a.Role != "admin" {
		return domain.DoctorDepartment{}, domain.ErrForbidden
	}
	if !domain.UUIDPattern.MatchString(id) {
		return domain.DoctorDepartment{}, domain.ErrValidation
	}
	if err := i.Validate(false); err != nil {
		return domain.DoctorDepartment{}, err
	}
	return s.Store.UpdateDoctorDepartment(ctx, a, id, i)
}

func (s MasterDataService) ArchiveDoctorDepartment(ctx context.Context, a domain.Actor, id string, version int, reason string) (domain.DoctorDepartment, error) {
	if a.Role != "admin" {
		return domain.DoctorDepartment{}, domain.ErrForbidden
	}
	if !domain.UUIDPattern.MatchString(id) || version < 1 || version > 1000000000 {
		return domain.DoctorDepartment{}, domain.ErrValidation
	}
	reason = strings.TrimSpace(reason)
	if len([]rune(reason)) < 1 || len([]rune(reason)) > 500 {
		return domain.DoctorDepartment{}, domain.ErrValidation
	}
	return s.Store.ArchiveDoctorDepartment(ctx, a, id, version, reason)
}

func (s MasterDataService) DoctorDepartmentRevisions(ctx context.Context, a domain.Actor, id string, page int) ([]domain.DoctorDepartmentRevision, error) {
	if a.Role != "admin" {
		return nil, domain.ErrForbidden
	}
	if !domain.UUIDPattern.MatchString(id) || page < 1 || page > 1000 {
		return nil, domain.ErrValidation
	}
	return s.Store.DoctorDepartmentRevisions(ctx, a, id, page)
}

// ─── Doctor profile extensions ────────────────────────────────────────────────

func (s MasterDataService) DoctorExt(ctx context.Context, a domain.Actor, doctorID string) (domain.DoctorExt, error) {
	if !a.Can("appointments.read") && !a.Can("staff.manage") {
		return domain.DoctorExt{}, domain.ErrForbidden
	}
	if doctorID == "" || len(doctorID) > 128 {
		return domain.DoctorExt{}, domain.ErrValidation
	}
	return s.Store.DoctorExt(ctx, doctorID)
}

func (s MasterDataService) SaveDoctorExt(ctx context.Context, a domain.Actor, doctorID string, i domain.DoctorExtInput) (domain.DoctorExt, error) {
	if a.Role != "admin" && !(a.Role == "doctor" && a.ID == doctorID) {
		return domain.DoctorExt{}, domain.ErrForbidden
	}
	if doctorID == "" || len(doctorID) > 128 {
		return domain.DoctorExt{}, domain.ErrValidation
	}
	if err := i.Validate(); err != nil {
		return domain.DoctorExt{}, err
	}
	return s.Store.SaveDoctorExt(ctx, a, doctorID, i)
}

// ─── Hospital hours ───────────────────────────────────────────────────────────

func (s MasterDataService) HospitalHours(ctx context.Context, a domain.Actor) ([]domain.HospitalHoursEntry, error) {
	// Any authenticated user may read opening hours.
	return s.Store.HospitalHours(ctx)
}

func (s MasterDataService) SaveHospitalHours(ctx context.Context, a domain.Actor, i domain.HospitalHoursInput) ([]domain.HospitalHoursEntry, error) {
	if a.Role != "admin" {
		return nil, domain.ErrForbidden
	}
	if err := i.Validate(); err != nil {
		return nil, err
	}
	return s.Store.SaveHospitalHours(ctx, a, i)
}

// ─── Hospital date overrides ──────────────────────────────────────────────────

func (s MasterDataService) HospitalDateOverrides(ctx context.Context, a domain.Actor, page int) ([]domain.HospitalDateOverride, error) {
	if page < 1 || page > 1000 {
		return nil, domain.ErrValidation
	}
	return s.Store.HospitalDateOverrides(ctx, page)
}

func (s MasterDataService) CreateHospitalDateOverride(ctx context.Context, a domain.Actor, i domain.HospitalDateOverrideInput) (domain.HospitalDateOverride, error) {
	if a.Role != "admin" {
		return domain.HospitalDateOverride{}, domain.ErrForbidden
	}
	if err := i.Validate(); err != nil {
		return domain.HospitalDateOverride{}, err
	}
	return s.Store.CreateHospitalDateOverride(ctx, a, i)
}

func (s MasterDataService) DeleteHospitalDateOverride(ctx context.Context, a domain.Actor, id string) error {
	if a.Role != "admin" {
		return domain.ErrForbidden
	}
	if !domain.UUIDPattern.MatchString(id) {
		return domain.ErrValidation
	}
	return s.Store.DeleteHospitalDateOverride(ctx, a, id)
}

// ─── Patient profile extensions ───────────────────────────────────────────────

func (s MasterDataService) PatientProfileExt(ctx context.Context, a domain.Actor, patientID string) (domain.PatientProfileExt, error) {
	if !a.Can("patients.read") {
		return domain.PatientProfileExt{}, domain.ErrForbidden
	}
	if !domain.UUIDPattern.MatchString(patientID) {
		return domain.PatientProfileExt{}, domain.ErrValidation
	}
	if err := s.Store.AuthorizePatientRecord(ctx, a, patientID); err != nil {
		return domain.PatientProfileExt{}, err
	}
	return s.Store.PatientProfileExt(ctx, patientID)
}

func (s MasterDataService) SavePatientProfileExt(ctx context.Context, a domain.Actor, patientID string, i domain.PatientProfileExt) error {
	if !a.Can("patients.create") {
		return domain.ErrForbidden
	}
	if !domain.UUIDPattern.MatchString(patientID) {
		return domain.ErrValidation
	}
	if err := i.Validate(); err != nil {
		return err
	}
	if err := s.Store.AuthorizePatientRecord(ctx, a, patientID); err != nil {
		return err
	}
	return s.Store.SavePatientProfileExt(ctx, a, patientID, i)
}
