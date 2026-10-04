package application

import (
	"context"
	"fmt"
	"time"

	"hms.local/api/internal/domain"
)

type DiagnosticReportsRepository interface {
	// Categories & Units
	DiagnosticCategories(ctx context.Context, kind string) ([]domain.DiagnosticCategory, error)
	CreateDiagnosticCategory(ctx context.Context, a domain.Actor, input domain.DiagnosticCategoryInput) (domain.DiagnosticCategory, error)
	DiagnosticUnits(ctx context.Context) ([]domain.DiagnosticUnit, error)
	CreateDiagnosticUnit(ctx context.Context, a domain.Actor, input domain.DiagnosticUnitInput) (domain.DiagnosticUnit, error)

	// Report Files & Portal Release
	DiagnosticReportFiles(ctx context.Context, orderID string, patientOnly bool) ([]domain.DiagnosticReportFile, error)
	UploadDiagnosticReportFile(ctx context.Context, a domain.Actor, orderID string, input domain.UploadDiagnosticReportFileInput) (domain.DiagnosticReportFile, error)
	ReleaseReportFileToPortal(ctx context.Context, a domain.Actor, fileID string) error

	// Diagnosis Templates
	DiagnosisTemplates(ctx context.Context) ([]domain.DiagnosisTemplate, error)
	CreateDiagnosisTemplate(ctx context.Context, a domain.Actor, input domain.DiagnosisTemplateInput) (domain.DiagnosisTemplate, error)

	// Vaccines & Vaccinations
	Vaccines(ctx context.Context) ([]domain.Vaccine, error)
	CreateVaccine(ctx context.Context, a domain.Actor, input domain.VaccineInput) (domain.Vaccine, error)
	PatientVaccinations(ctx context.Context, patientID string) ([]domain.PatientVaccination, error)
	AdministerVaccine(ctx context.Context, a domain.Actor, input domain.AdministerVaccineInput) (domain.PatientVaccination, error)

	// Vital Reports
	BirthReports(ctx context.Context, page int) ([]domain.BirthReport, error)
	CreateBirthReport(ctx context.Context, a domain.Actor, reportNumber string, input domain.CreateBirthReportInput) (domain.BirthReport, error)
	DeathReports(ctx context.Context, page int) ([]domain.DeathReport, error)
	CreateDeathReport(ctx context.Context, a domain.Actor, reportNumber string, input domain.CreateDeathReportInput) (domain.DeathReport, error)
	OperationReports(ctx context.Context, page int) ([]domain.OperationReport, error)
	CreateOperationReport(ctx context.Context, a domain.Actor, reportNumber string, input domain.CreateOperationReportInput) (domain.OperationReport, error)
	InvestigationReports(ctx context.Context, patientID string) ([]domain.InvestigationReport, error)
	CreateInvestigationReport(ctx context.Context, a domain.Actor, reportNumber string, input domain.CreateInvestigationReportInput) (domain.InvestigationReport, error)
}

type DiagnosticReportsService struct {
	Store DiagnosticReportsRepository
}

// ─── Categories & Units ───────────────────────────────────────────────────────

func (s DiagnosticReportsService) DiagnosticCategories(ctx context.Context, a domain.Actor, kind string) ([]domain.DiagnosticCategory, error) {
	return s.Store.DiagnosticCategories(ctx, kind)
}

func (s DiagnosticReportsService) CreateDiagnosticCategory(ctx context.Context, a domain.Actor, input domain.DiagnosticCategoryInput) (domain.DiagnosticCategory, error) {
	if a.Role != "admin" && a.Role != "lab_technician" {
		return domain.DiagnosticCategory{}, domain.ErrForbidden
	}
	if err := input.Validate(); err != nil {
		return domain.DiagnosticCategory{}, err
	}
	return s.Store.CreateDiagnosticCategory(ctx, a, input)
}

func (s DiagnosticReportsService) DiagnosticUnits(ctx context.Context, a domain.Actor) ([]domain.DiagnosticUnit, error) {
	return s.Store.DiagnosticUnits(ctx)
}

func (s DiagnosticReportsService) CreateDiagnosticUnit(ctx context.Context, a domain.Actor, input domain.DiagnosticUnitInput) (domain.DiagnosticUnit, error) {
	if a.Role != "admin" && a.Role != "lab_technician" {
		return domain.DiagnosticUnit{}, domain.ErrForbidden
	}
	if err := input.Validate(); err != nil {
		return domain.DiagnosticUnit{}, err
	}
	return s.Store.CreateDiagnosticUnit(ctx, a, input)
}

// ─── Report Files & Portal Release ───────────────────────────────────────────

func (s DiagnosticReportsService) DiagnosticReportFiles(ctx context.Context, a domain.Actor, orderID string) ([]domain.DiagnosticReportFile, error) {
	if !domain.UUIDPattern.MatchString(orderID) {
		return nil, domain.ErrValidation
	}
	patientOnly := a.Role == "patient"
	return s.Store.DiagnosticReportFiles(ctx, orderID, patientOnly)
}

func (s DiagnosticReportsService) UploadDiagnosticReportFile(ctx context.Context, a domain.Actor, orderID string, input domain.UploadDiagnosticReportFileInput) (domain.DiagnosticReportFile, error) {
	if a.Role != "admin" && a.Role != "lab_technician" && a.Role != "doctor" {
		return domain.DiagnosticReportFile{}, domain.ErrForbidden
	}
	if !domain.UUIDPattern.MatchString(orderID) {
		return domain.DiagnosticReportFile{}, domain.ErrValidation
	}
	if err := input.Validate(); err != nil {
		return domain.DiagnosticReportFile{}, err
	}
	return s.Store.UploadDiagnosticReportFile(ctx, a, orderID, input)
}

func (s DiagnosticReportsService) ReleaseReportFileToPortal(ctx context.Context, a domain.Actor, fileID string) error {
	if a.Role != "admin" && a.Role != "doctor" && a.Role != "lab_technician" {
		return domain.ErrForbidden
	}
	if !domain.UUIDPattern.MatchString(fileID) {
		return domain.ErrValidation
	}
	return s.Store.ReleaseReportFileToPortal(ctx, a, fileID)
}

// ─── Diagnosis Templates ──────────────────────────────────────────────────────

func (s DiagnosticReportsService) DiagnosisTemplates(ctx context.Context, a domain.Actor) ([]domain.DiagnosisTemplate, error) {
	return s.Store.DiagnosisTemplates(ctx)
}

func (s DiagnosticReportsService) CreateDiagnosisTemplate(ctx context.Context, a domain.Actor, input domain.DiagnosisTemplateInput) (domain.DiagnosisTemplate, error) {
	if a.Role != "admin" && a.Role != "doctor" {
		return domain.DiagnosisTemplate{}, domain.ErrForbidden
	}
	if err := input.Validate(); err != nil {
		return domain.DiagnosisTemplate{}, err
	}
	return s.Store.CreateDiagnosisTemplate(ctx, a, input)
}

// ─── Vaccines & Vaccinations ──────────────────────────────────────────────────

func (s DiagnosticReportsService) Vaccines(ctx context.Context, a domain.Actor) ([]domain.Vaccine, error) {
	return s.Store.Vaccines(ctx)
}

func (s DiagnosticReportsService) CreateVaccine(ctx context.Context, a domain.Actor, input domain.VaccineInput) (domain.Vaccine, error) {
	if a.Role != "admin" && a.Role != "doctor" {
		return domain.Vaccine{}, domain.ErrForbidden
	}
	if err := input.Validate(); err != nil {
		return domain.Vaccine{}, err
	}
	return s.Store.CreateVaccine(ctx, a, input)
}

func (s DiagnosticReportsService) PatientVaccinations(ctx context.Context, a domain.Actor, patientID string) ([]domain.PatientVaccination, error) {
	if !domain.UUIDPattern.MatchString(patientID) {
		return nil, domain.ErrValidation
	}
	return s.Store.PatientVaccinations(ctx, patientID)
}

func (s DiagnosticReportsService) AdministerVaccine(ctx context.Context, a domain.Actor, input domain.AdministerVaccineInput) (domain.PatientVaccination, error) {
	if a.Role != "admin" && a.Role != "doctor" && a.Role != "nurse" {
		return domain.PatientVaccination{}, domain.ErrForbidden
	}
	if err := input.Validate(); err != nil {
		return domain.PatientVaccination{}, err
	}
	return s.Store.AdministerVaccine(ctx, a, input)
}

// ─── Vital Reports ────────────────────────────────────────────────────────────

func (s DiagnosticReportsService) BirthReports(ctx context.Context, a domain.Actor, page int) ([]domain.BirthReport, error) {
	if page < 1 {
		page = 1
	}
	return s.Store.BirthReports(ctx, page)
}

func (s DiagnosticReportsService) CreateBirthReport(ctx context.Context, a domain.Actor, input domain.CreateBirthReportInput) (domain.BirthReport, error) {
	if a.Role != "admin" && a.Role != "doctor" && a.Role != "nurse" {
		return domain.BirthReport{}, domain.ErrForbidden
	}
	if err := input.Validate(); err != nil {
		return domain.BirthReport{}, err
	}
	reportNumber := fmt.Sprintf("BR-%d-%04d", time.Now().Year(), time.Now().UnixNano()%10000)
	return s.Store.CreateBirthReport(ctx, a, reportNumber, input)
}

func (s DiagnosticReportsService) DeathReports(ctx context.Context, a domain.Actor, page int) ([]domain.DeathReport, error) {
	if page < 1 {
		page = 1
	}
	return s.Store.DeathReports(ctx, page)
}

func (s DiagnosticReportsService) CreateDeathReport(ctx context.Context, a domain.Actor, input domain.CreateDeathReportInput) (domain.DeathReport, error) {
	if a.Role != "admin" && a.Role != "doctor" {
		return domain.DeathReport{}, domain.ErrForbidden
	}
	if err := input.Validate(); err != nil {
		return domain.DeathReport{}, err
	}
	reportNumber := fmt.Sprintf("DR-%d-%04d", time.Now().Year(), time.Now().UnixNano()%10000)
	return s.Store.CreateDeathReport(ctx, a, reportNumber, input)
}

func (s DiagnosticReportsService) OperationReports(ctx context.Context, a domain.Actor, page int) ([]domain.OperationReport, error) {
	if page < 1 {
		page = 1
	}
	return s.Store.OperationReports(ctx, page)
}

func (s DiagnosticReportsService) CreateOperationReport(ctx context.Context, a domain.Actor, input domain.CreateOperationReportInput) (domain.OperationReport, error) {
	if a.Role != "admin" && a.Role != "doctor" {
		return domain.OperationReport{}, domain.ErrForbidden
	}
	if err := input.Validate(); err != nil {
		return domain.OperationReport{}, err
	}
	reportNumber := fmt.Sprintf("OR-%d-%04d", time.Now().Year(), time.Now().UnixNano()%10000)
	return s.Store.CreateOperationReport(ctx, a, reportNumber, input)
}

func (s DiagnosticReportsService) InvestigationReports(ctx context.Context, a domain.Actor, patientID string) ([]domain.InvestigationReport, error) {
	if !domain.UUIDPattern.MatchString(patientID) {
		return nil, domain.ErrValidation
	}
	return s.Store.InvestigationReports(ctx, patientID)
}

func (s DiagnosticReportsService) CreateInvestigationReport(ctx context.Context, a domain.Actor, input domain.CreateInvestigationReportInput) (domain.InvestigationReport, error) {
	if a.Role != "admin" && a.Role != "doctor" && a.Role != "lab_technician" {
		return domain.InvestigationReport{}, domain.ErrForbidden
	}
	if err := input.Validate(); err != nil {
		return domain.InvestigationReport{}, err
	}
	reportNumber := fmt.Sprintf("IR-%d-%04d", time.Now().Year(), time.Now().UnixNano()%10000)
	return s.Store.CreateInvestigationReport(ctx, a, reportNumber, input)
}
