package application

import (
	"context"
	"hms.local/api/internal/domain"
	"time"
)

type PharmacyBloodBankStore interface {
	ValidateClinicalAttribution(context.Context, domain.Actor, string, string, *string) error
	// Categories & Brands
	MedicineCategories(context.Context, int, string) ([]domain.MedicineCategory, int, error)
	CreateMedicineCategory(context.Context, domain.Actor, domain.MedicineCategoryInput) (domain.MedicineCategory, error)
	UpdateMedicineCategory(context.Context, domain.Actor, string, domain.MedicineCategoryInput) (domain.MedicineCategory, error)

	MedicineBrands(context.Context, int, string) ([]domain.MedicineBrand, int, error)
	CreateMedicineBrand(context.Context, domain.Actor, domain.MedicineBrandInput) (domain.MedicineBrand, error)
	UpdateMedicineBrand(context.Context, domain.Actor, string, domain.MedicineBrandInput) (domain.MedicineBrand, error)

	// Blood Bank
	BloodBank(context.Context) ([]domain.BloodBankItem, error)
	BloodDonors(context.Context, int, string) ([]domain.BloodDonor, int, error)
	CreateBloodDonor(context.Context, domain.Actor, domain.BloodDonorInput) (domain.BloodDonor, error)

	BloodDonations(context.Context, int) ([]domain.BloodDonation, int, error)
	RecordBloodDonation(context.Context, domain.Actor, domain.BloodDonationInput) (domain.BloodDonation, error)

	BloodIssues(context.Context, string, int) ([]domain.BloodIssue, int, error)
	BloodIssue(context.Context, string) (domain.BloodIssue, error)
	CreateBloodIssue(context.Context, domain.Actor, domain.BloodIssueInput) (domain.BloodIssue, error)

	// Prescriptions
	Prescriptions(context.Context, string, string, int) ([]domain.Prescription, int, error)
	Prescription(context.Context, string) (domain.Prescription, error)
	CreatePrescription(context.Context, domain.Actor, domain.PrescriptionInput) (domain.Prescription, error)
	UpdatePrescriptionStatus(context.Context, domain.Actor, string, int) (domain.Prescription, error)
}

type PharmacyBloodBankService struct {
	Store PharmacyBloodBankStore
	Now   func() time.Time
}

// --- Categories & Brands ---

func (s PharmacyBloodBankService) MedicineCategories(ctx context.Context, a domain.Actor, page int, search string) ([]domain.MedicineCategory, int, error) {
	if !a.Can("pharmacy.catalog") {
		return nil, 0, domain.ErrForbidden
	}
	return s.Store.MedicineCategories(ctx, page, search)
}

func (s PharmacyBloodBankService) CreateMedicineCategory(ctx context.Context, a domain.Actor, in domain.MedicineCategoryInput) (domain.MedicineCategory, error) {
	if !a.Can("pharmacy.manage") {
		return domain.MedicineCategory{}, domain.ErrForbidden
	}
	if err := in.Validate(); err != nil {
		return domain.MedicineCategory{}, err
	}
	return s.Store.CreateMedicineCategory(ctx, a, in)
}

func (s PharmacyBloodBankService) UpdateMedicineCategory(ctx context.Context, a domain.Actor, id string, in domain.MedicineCategoryInput) (domain.MedicineCategory, error) {
	if !a.Can("pharmacy.manage") {
		return domain.MedicineCategory{}, domain.ErrForbidden
	}
	if err := in.Validate(); err != nil {
		return domain.MedicineCategory{}, err
	}
	return s.Store.UpdateMedicineCategory(ctx, a, id, in)
}

func (s PharmacyBloodBankService) MedicineBrands(ctx context.Context, a domain.Actor, page int, search string) ([]domain.MedicineBrand, int, error) {
	if !a.Can("pharmacy.catalog") {
		return nil, 0, domain.ErrForbidden
	}
	return s.Store.MedicineBrands(ctx, page, search)
}

func (s PharmacyBloodBankService) CreateMedicineBrand(ctx context.Context, a domain.Actor, in domain.MedicineBrandInput) (domain.MedicineBrand, error) {
	if !a.Can("pharmacy.manage") {
		return domain.MedicineBrand{}, domain.ErrForbidden
	}
	if err := in.Validate(); err != nil {
		return domain.MedicineBrand{}, err
	}
	return s.Store.CreateMedicineBrand(ctx, a, in)
}

func (s PharmacyBloodBankService) UpdateMedicineBrand(ctx context.Context, a domain.Actor, id string, in domain.MedicineBrandInput) (domain.MedicineBrand, error) {
	if !a.Can("pharmacy.manage") {
		return domain.MedicineBrand{}, domain.ErrForbidden
	}
	if err := in.Validate(); err != nil {
		return domain.MedicineBrand{}, err
	}
	return s.Store.UpdateMedicineBrand(ctx, a, id, in)
}

// --- Blood Bank ---

func (s PharmacyBloodBankService) BloodBank(ctx context.Context, a domain.Actor) ([]domain.BloodBankItem, error) {
	if !a.Can("blood_bank.read") || a.Role == "patient" {
		return nil, domain.ErrForbidden
	}
	return s.Store.BloodBank(ctx)
}

func (s PharmacyBloodBankService) BloodDonors(ctx context.Context, a domain.Actor, page int, search string) ([]domain.BloodDonor, int, error) {
	if !a.Can("blood_bank.read") || a.Role == "patient" {
		return nil, 0, domain.ErrForbidden
	}
	return s.Store.BloodDonors(ctx, page, search)
}

func (s PharmacyBloodBankService) CreateBloodDonor(ctx context.Context, a domain.Actor, in domain.BloodDonorInput) (domain.BloodDonor, error) {
	if !a.Can("blood_bank.manage") {
		return domain.BloodDonor{}, domain.ErrForbidden
	}
	if err := in.Validate(); err != nil {
		return domain.BloodDonor{}, err
	}
	return s.Store.CreateBloodDonor(ctx, a, in)
}

func (s PharmacyBloodBankService) BloodDonations(ctx context.Context, a domain.Actor, page int) ([]domain.BloodDonation, int, error) {
	if !a.Can("blood_bank.read") || a.Role == "patient" {
		return nil, 0, domain.ErrForbidden
	}
	return s.Store.BloodDonations(ctx, page)
}

func (s PharmacyBloodBankService) RecordBloodDonation(ctx context.Context, a domain.Actor, in domain.BloodDonationInput) (domain.BloodDonation, error) {
	if !a.Can("blood_bank.manage") {
		return domain.BloodDonation{}, domain.ErrForbidden
	}
	if err := in.Validate(); err != nil {
		return domain.BloodDonation{}, err
	}
	return s.Store.RecordBloodDonation(ctx, a, in)
}

func (s PharmacyBloodBankService) BloodIssues(ctx context.Context, a domain.Actor, page int) ([]domain.BloodIssue, int, error) {
	if !a.Can("blood_bank.read") {
		return nil, 0, domain.ErrForbidden
	}
	patientFilter := ""
	if a.Role == "patient" {
		patientFilter = a.ID
	}
	return s.Store.BloodIssues(ctx, patientFilter, page)
}

func (s PharmacyBloodBankService) BloodIssue(ctx context.Context, a domain.Actor, id string) (domain.BloodIssue, error) {
	if !a.Can("blood_bank.read") {
		return domain.BloodIssue{}, domain.ErrForbidden
	}
	issue, err := s.Store.BloodIssue(ctx, id)
	if err != nil {
		return domain.BloodIssue{}, err
	}
	if a.Role == "patient" && issue.PatientID != a.ID && (issue.PatientUserID == nil || *issue.PatientUserID != a.ID) {
		return domain.BloodIssue{}, domain.ErrForbidden
	}
	return issue, nil
}

func (s PharmacyBloodBankService) CreateBloodIssue(ctx context.Context, a domain.Actor, in domain.BloodIssueInput) (domain.BloodIssue, error) {
	if !a.Can("blood_bank.manage") {
		return domain.BloodIssue{}, domain.ErrForbidden
	}
	if err := in.Validate(); err != nil {
		return domain.BloodIssue{}, err
	}
	if err := s.Store.ValidateClinicalAttribution(ctx, a, in.PatientID, in.DoctorID, nil); err != nil {
		return domain.BloodIssue{}, err
	}
	return s.Store.CreateBloodIssue(ctx, a, in)
}

// --- Prescriptions ---

func (s PharmacyBloodBankService) Prescriptions(ctx context.Context, a domain.Actor, page int) ([]domain.Prescription, int, error) {
	if !a.Can("prescriptions.read") {
		return nil, 0, domain.ErrForbidden
	}
	patientFilter := ""
	if a.Role == "patient" {
		patientFilter = a.ID
	}
	doctorFilter := ""
	if a.Role == "doctor" {
		doctorFilter = a.ID
	}
	return s.Store.Prescriptions(ctx, patientFilter, doctorFilter, page)
}

func (s PharmacyBloodBankService) Prescription(ctx context.Context, a domain.Actor, id string) (domain.Prescription, error) {
	if !a.Can("prescriptions.read") {
		return domain.Prescription{}, domain.ErrForbidden
	}
	p, err := s.Store.Prescription(ctx, id)
	if err != nil {
		return domain.Prescription{}, err
	}
	if a.Role == "patient" && p.PatientID != a.ID && (p.PatientUserID == nil || *p.PatientUserID != a.ID) {
		return domain.Prescription{}, domain.ErrForbidden
	}
	if a.Role == "doctor" && p.DoctorID != a.ID {
		return domain.Prescription{}, domain.ErrForbidden
	}
	return p, nil
}

func (s PharmacyBloodBankService) CreatePrescription(ctx context.Context, a domain.Actor, in domain.PrescriptionInput) (domain.Prescription, error) {
	if !a.Can("prescriptions.manage") {
		return domain.Prescription{}, domain.ErrForbidden
	}
	if a.Role == "doctor" && in.DoctorID == "" {
		in.DoctorID = a.ID
	}
	if err := in.Validate(); err != nil {
		return domain.Prescription{}, err
	}
	if err := s.Store.ValidateClinicalAttribution(ctx, a, in.PatientID, in.DoctorID, in.EncounterID); err != nil {
		return domain.Prescription{}, err
	}
	return s.Store.CreatePrescription(ctx, a, in)
}

func (s PharmacyBloodBankService) UpdatePrescriptionStatus(ctx context.Context, a domain.Actor, id string, status int) (domain.Prescription, error) {
	// Legacy active/inactive flag; never represents dispensing or stock movement.
	if !a.Can("pharmacy.manage") && !a.Can("prescriptions.manage") {
		return domain.Prescription{}, domain.ErrForbidden
	}
	if status != 0 && status != 1 {
		return domain.Prescription{}, domain.ErrValidation
	}
	if a.Role == "doctor" {
		if _, err := s.Prescription(ctx, a, id); err != nil {
			return domain.Prescription{}, err
		}
	}
	return s.Store.UpdatePrescriptionStatus(ctx, a, id, status)
}
