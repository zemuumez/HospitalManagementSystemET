package application

import (
	"context"
	"strings"

	"hms.local/api/internal/domain"
)

type ClinicalCareRepository interface {
	AuthorizeClinicalRecord(context.Context, domain.Actor, string, string, string) error
	// Bed assignments & occupancy
	BedOccupancyReport(ctx context.Context) (domain.BedOccupancyReport, error)
	AssignBed(ctx context.Context, a domain.Actor, input domain.AssignBedInput) (domain.BedAssignment, error)
	ListBedAssignments(ctx context.Context, encounterID string) ([]domain.BedAssignment, error)
	ListAllBedAssignments(ctx context.Context, search, status string, page, limit int) ([]domain.BedAssignment, int, error)
	DischargeBedAssignment(ctx context.Context, a domain.Actor, id, notes string) (domain.BedAssignment, error)

	// Care team delegation
	CareTeam(ctx context.Context, encounterID string) ([]domain.CareTeamMember, error)
	AddCareTeamMember(ctx context.Context, a domain.Actor, encounterID string, input domain.AddCareTeamMemberInput) (domain.CareTeamMember, error)
	RevokeCareTeamMember(ctx context.Context, a domain.Actor, memberID string) error

	// Diagnoses, procedures, attachments
	Diagnoses(ctx context.Context, encounterID string) ([]domain.EncounterDiagnosis, error)
	AddDiagnosis(ctx context.Context, a domain.Actor, encounterID string, input domain.AddDiagnosisInput) (domain.EncounterDiagnosis, error)
	Procedures(ctx context.Context, encounterID string) ([]domain.EncounterProcedure, error)
	AddProcedure(ctx context.Context, a domain.Actor, encounterID string, input domain.AddProcedureInput) (domain.EncounterProcedure, error)
	Attachments(ctx context.Context, encounterID string) ([]domain.EncounterAttachment, error)
	AddAttachment(ctx context.Context, a domain.Actor, encounterID string, input domain.AddAttachmentInput) (domain.EncounterAttachment, error)

	// IPD admission details
	IPDAdmissionDetails(ctx context.Context, encounterID string) (domain.IPDAdmissionDetails, error)
	SaveIPDAdmissionDetails(ctx context.Context, a domain.Actor, details domain.IPDAdmissionDetails) error

	// Encounter billing & clearance
	EncounterBilling(ctx context.Context, encounterID string) (domain.EncounterBilling, error)
	UpdateEncounterBilling(ctx context.Context, a domain.Actor, encounterID string, input domain.UpdateEncounterBillingInput) (domain.EncounterBilling, error)
	GrantFinancialClearance(ctx context.Context, a domain.Actor, encounterID string, waiverReason string) (domain.EncounterBilling, error)
	LinkEncounterInvoice(ctx context.Context, a domain.Actor, encounterID string, invoiceID string) (domain.EncounterBilling, error)

	// Discharge summaries
	DischargeSummary(ctx context.Context, encounterID string) (domain.DischargeSummary, error)
	SaveDischargeSummary(ctx context.Context, a domain.Actor, summary domain.DischargeSummary) error

	// OPD follow-ups & referrals
	OPDFollowUps(ctx context.Context, patientID string) ([]domain.OPDFollowUp, error)
	CreateOPDFollowUp(ctx context.Context, a domain.Actor, input domain.CreateOPDFollowUpInput) (domain.OPDFollowUp, error)
	PatientReferrals(ctx context.Context, patientID string) ([]domain.PatientReferral, error)
	CreatePatientReferral(ctx context.Context, a domain.Actor, input domain.CreatePatientReferralInput) (domain.PatientReferral, error)

	// Odontogram
	Odontogram(ctx context.Context, patientID string) ([]domain.OdontogramEntry, error)
	SetToothCondition(ctx context.Context, a domain.Actor, patientID string, input domain.SetToothConditionInput) (domain.OdontogramEntry, error)
	CreateOdontogramChart(ctx context.Context, a domain.Actor, input domain.SaveOdontogramChartInput) (domain.OdontogramChart, error)
	UpdateOdontogramChart(ctx context.Context, a domain.Actor, id string, input domain.SaveOdontogramChartInput) (domain.OdontogramChart, error)
	GetOdontogramChart(ctx context.Context, id string) (domain.OdontogramChart, error)
	ListOdontogramCharts(ctx context.Context, a domain.Actor, patientID, search string, page, limit int) ([]domain.OdontogramChart, int, error)
	DeleteOdontogramChart(ctx context.Context, id string) error
	ValidatePatientEncounter(ctx context.Context, patientID, encounterID string) error

	// Odontogram Legends
	ListOdontogramLegends(ctx context.Context) ([]domain.OdontogramLegend, error)
	CreateOdontogramLegend(ctx context.Context, a domain.Actor, input domain.SaveOdontogramLegendInput) (domain.OdontogramLegend, error)
	UpdateOdontogramLegend(ctx context.Context, a domain.Actor, id string, input domain.SaveOdontogramLegendInput) (domain.OdontogramLegend, error)
	DeleteOdontogramLegend(ctx context.Context, a domain.Actor, id string) error
}

type ClinicalCareService struct {
	Store ClinicalCareRepository
}

// ─── Bed Occupancy & Assignments ──────────────────────────────────────────────

func (s ClinicalCareService) BedOccupancyReport(ctx context.Context, a domain.Actor) (domain.BedOccupancyReport, error) {
	if a.Role != "admin" && a.Role != "doctor" && a.Role != "nurse" && a.Role != "receptionist" {
		return domain.BedOccupancyReport{}, domain.ErrForbidden
	}
	return s.Store.BedOccupancyReport(ctx)
}

func (s ClinicalCareService) AssignBed(ctx context.Context, a domain.Actor, input domain.AssignBedInput) (domain.BedAssignment, error) {
	if a.Role != "admin" && a.Role != "doctor" && a.Role != "nurse" && a.Role != "receptionist" {
		return domain.BedAssignment{}, domain.ErrForbidden
	}
	if err := input.Validate(); err != nil {
		return domain.BedAssignment{}, err
	}
	if err := s.Store.AuthorizeClinicalRecord(ctx, a, input.PatientID, input.EncounterID, "admission"); err != nil {
		return domain.BedAssignment{}, err
	}
	return s.Store.AssignBed(ctx, a, input)
}

func (s ClinicalCareService) ListBedAssignments(ctx context.Context, a domain.Actor, encounterID string) ([]domain.BedAssignment, error) {
	if !domain.UUIDPattern.MatchString(encounterID) {
		return nil, domain.ErrValidation
	}
	if err := s.Store.AuthorizeClinicalRecord(ctx, a, "", encounterID, "admission"); err != nil {
		return nil, err
	}
	return s.Store.ListBedAssignments(ctx, encounterID)
}

func (s ClinicalCareService) ListAllBedAssignments(ctx context.Context, a domain.Actor, search, status string, page, pageSize int) ([]domain.BedAssignment, int, error) {
	if a.Role != "admin" && a.Role != "doctor" && a.Role != "nurse" && a.Role != "receptionist" {
		return nil, 0, domain.ErrForbidden
	}
	if page < 1 || page > 1000 || pageSize < 1 || pageSize > 100 {
		return nil, 0, domain.ErrValidation
	}
	return s.Store.ListAllBedAssignments(ctx, search, status, page, pageSize)
}

func (s ClinicalCareService) DischargeBedAssignment(ctx context.Context, a domain.Actor, id, notes string) (domain.BedAssignment, error) {
	if a.Role != "admin" && a.Role != "doctor" && a.Role != "nurse" && a.Role != "receptionist" {
		return domain.BedAssignment{}, domain.ErrForbidden
	}
	if !domain.UUIDPattern.MatchString(id) {
		return domain.BedAssignment{}, domain.ErrValidation
	}
	return s.Store.DischargeBedAssignment(ctx, a, id, notes)
}

// ─── Care Team ────────────────────────────────────────────────────────────────

func (s ClinicalCareService) CareTeam(ctx context.Context, a domain.Actor, encounterID string) ([]domain.CareTeamMember, error) {
	if !domain.UUIDPattern.MatchString(encounterID) {
		return nil, domain.ErrValidation
	}
	if err := s.Store.AuthorizeClinicalRecord(ctx, a, "", encounterID, "clinical"); err != nil {
		return nil, err
	}
	return s.Store.CareTeam(ctx, encounterID)
}

func (s ClinicalCareService) AddCareTeamMember(ctx context.Context, a domain.Actor, encounterID string, input domain.AddCareTeamMemberInput) (domain.CareTeamMember, error) {
	if a.Role != "admin" && a.Role != "doctor" && a.Role != "nurse" {
		return domain.CareTeamMember{}, domain.ErrForbidden
	}
	if !domain.UUIDPattern.MatchString(encounterID) {
		return domain.CareTeamMember{}, domain.ErrValidation
	}
	if err := input.Validate(); err != nil {
		return domain.CareTeamMember{}, err
	}
	if err := s.Store.AuthorizeClinicalRecord(ctx, a, "", encounterID, "clinical"); err != nil {
		return domain.CareTeamMember{}, err
	}
	return s.Store.AddCareTeamMember(ctx, a, encounterID, input)
}

func (s ClinicalCareService) RevokeCareTeamMember(ctx context.Context, a domain.Actor, memberID string) error {
	if a.Role != "admin" && a.Role != "doctor" {
		return domain.ErrForbidden
	}
	if !domain.UUIDPattern.MatchString(memberID) {
		return domain.ErrValidation
	}
	return s.Store.RevokeCareTeamMember(ctx, a, memberID)
}

// ─── Diagnoses, Procedures, Attachments ───────────────────────────────────────

func (s ClinicalCareService) Diagnoses(ctx context.Context, a domain.Actor, encounterID string) ([]domain.EncounterDiagnosis, error) {
	if !domain.UUIDPattern.MatchString(encounterID) {
		return nil, domain.ErrValidation
	}
	if err := s.Store.AuthorizeClinicalRecord(ctx, a, "", encounterID, "clinical"); err != nil {
		return nil, err
	}
	return s.Store.Diagnoses(ctx, encounterID)
}

func (s ClinicalCareService) AddDiagnosis(ctx context.Context, a domain.Actor, encounterID string, input domain.AddDiagnosisInput) (domain.EncounterDiagnosis, error) {
	if a.Role != "admin" && a.Role != "doctor" {
		return domain.EncounterDiagnosis{}, domain.ErrForbidden
	}
	if !domain.UUIDPattern.MatchString(encounterID) {
		return domain.EncounterDiagnosis{}, domain.ErrValidation
	}
	if err := input.Validate(); err != nil {
		return domain.EncounterDiagnosis{}, err
	}
	if err := s.Store.AuthorizeClinicalRecord(ctx, a, "", encounterID, "clinical"); err != nil {
		return domain.EncounterDiagnosis{}, err
	}
	return s.Store.AddDiagnosis(ctx, a, encounterID, input)
}

func (s ClinicalCareService) Procedures(ctx context.Context, a domain.Actor, encounterID string) ([]domain.EncounterProcedure, error) {
	if !domain.UUIDPattern.MatchString(encounterID) {
		return nil, domain.ErrValidation
	}
	if err := s.Store.AuthorizeClinicalRecord(ctx, a, "", encounterID, "clinical"); err != nil {
		return nil, err
	}
	return s.Store.Procedures(ctx, encounterID)
}

func (s ClinicalCareService) AddProcedure(ctx context.Context, a domain.Actor, encounterID string, input domain.AddProcedureInput) (domain.EncounterProcedure, error) {
	if a.Role != "admin" && a.Role != "doctor" {
		return domain.EncounterProcedure{}, domain.ErrForbidden
	}
	if !domain.UUIDPattern.MatchString(encounterID) {
		return domain.EncounterProcedure{}, domain.ErrValidation
	}
	if err := input.Validate(); err != nil {
		return domain.EncounterProcedure{}, err
	}
	if err := s.Store.AuthorizeClinicalRecord(ctx, a, "", encounterID, "clinical"); err != nil {
		return domain.EncounterProcedure{}, err
	}
	return s.Store.AddProcedure(ctx, a, encounterID, input)
}

func (s ClinicalCareService) Attachments(ctx context.Context, a domain.Actor, encounterID string) ([]domain.EncounterAttachment, error) {
	if !domain.UUIDPattern.MatchString(encounterID) {
		return nil, domain.ErrValidation
	}
	if err := s.Store.AuthorizeClinicalRecord(ctx, a, "", encounterID, "clinical"); err != nil {
		return nil, err
	}
	return s.Store.Attachments(ctx, encounterID)
}

func (s ClinicalCareService) AddAttachment(ctx context.Context, a domain.Actor, encounterID string, input domain.AddAttachmentInput) (domain.EncounterAttachment, error) {
	if a.Role != "admin" && a.Role != "doctor" && a.Role != "nurse" {
		return domain.EncounterAttachment{}, domain.ErrForbidden
	}
	if !domain.UUIDPattern.MatchString(encounterID) {
		return domain.EncounterAttachment{}, domain.ErrValidation
	}
	if err := input.Validate(); err != nil {
		return domain.EncounterAttachment{}, err
	}
	if err := s.Store.AuthorizeClinicalRecord(ctx, a, "", encounterID, "clinical"); err != nil {
		return domain.EncounterAttachment{}, err
	}
	return s.Store.AddAttachment(ctx, a, encounterID, input)
}

// ─── IPD Admission Details ────────────────────────────────────────────────────

func (s ClinicalCareService) IPDAdmissionDetails(ctx context.Context, a domain.Actor, encounterID string) (domain.IPDAdmissionDetails, error) {
	if !domain.UUIDPattern.MatchString(encounterID) {
		return domain.IPDAdmissionDetails{}, domain.ErrValidation
	}
	if err := s.Store.AuthorizeClinicalRecord(ctx, a, "", encounterID, "admission"); err != nil {
		return domain.IPDAdmissionDetails{}, err
	}
	return s.Store.IPDAdmissionDetails(ctx, encounterID)
}

func (s ClinicalCareService) SaveIPDAdmissionDetails(ctx context.Context, a domain.Actor, details domain.IPDAdmissionDetails) error {
	if a.Role != "admin" && a.Role != "receptionist" && a.Role != "doctor" {
		return domain.ErrForbidden
	}
	if err := details.Validate(); err != nil {
		return err
	}
	if err := s.Store.AuthorizeClinicalRecord(ctx, a, "", details.EncounterID, "admission"); err != nil {
		return err
	}
	return s.Store.SaveIPDAdmissionDetails(ctx, a, details)
}

// ─── Encounter Billing & Financial Clearance ──────────────────────────────────

func (s ClinicalCareService) EncounterBilling(ctx context.Context, a domain.Actor, encounterID string) (domain.EncounterBilling, error) {
	if !domain.UUIDPattern.MatchString(encounterID) {
		return domain.EncounterBilling{}, domain.ErrValidation
	}
	if err := s.Store.AuthorizeClinicalRecord(ctx, a, "", encounterID, "billing"); err != nil {
		return domain.EncounterBilling{}, err
	}
	return s.Store.EncounterBilling(ctx, encounterID)
}

func (s ClinicalCareService) UpdateEncounterBilling(ctx context.Context, a domain.Actor, encounterID string, input domain.UpdateEncounterBillingInput) (domain.EncounterBilling, error) {
	if a.Role != "admin" && a.Role != "accountant" {
		return domain.EncounterBilling{}, domain.ErrForbidden
	}
	if !domain.UUIDPattern.MatchString(encounterID) {
		return domain.EncounterBilling{}, domain.ErrValidation
	}
	if err := input.Validate(); err != nil {
		return domain.EncounterBilling{}, err
	}
	if err := s.Store.AuthorizeClinicalRecord(ctx, a, "", encounterID, "billing"); err != nil {
		return domain.EncounterBilling{}, err
	}
	return s.Store.UpdateEncounterBilling(ctx, a, encounterID, input)
}

func (s ClinicalCareService) GrantFinancialClearance(ctx context.Context, a domain.Actor, encounterID string, waiverReason string) (domain.EncounterBilling, error) {
	if a.Role != "admin" && a.Role != "accountant" {
		return domain.EncounterBilling{}, domain.ErrForbidden
	}
	if !domain.UUIDPattern.MatchString(encounterID) {
		return domain.EncounterBilling{}, domain.ErrValidation
	}
	waiverReason = strings.TrimSpace(waiverReason)
	if err := s.Store.AuthorizeClinicalRecord(ctx, a, "", encounterID, "billing"); err != nil {
		return domain.EncounterBilling{}, err
	}
	return s.Store.GrantFinancialClearance(ctx, a, encounterID, waiverReason)
}

func (s ClinicalCareService) LinkEncounterInvoice(ctx context.Context, a domain.Actor, encounterID string, invoiceID string) (domain.EncounterBilling, error) {
	if a.Role != "admin" && a.Role != "accountant" {
		return domain.EncounterBilling{}, domain.ErrForbidden
	}
	if !domain.UUIDPattern.MatchString(encounterID) || !domain.UUIDPattern.MatchString(invoiceID) {
		return domain.EncounterBilling{}, domain.ErrValidation
	}
	if err := s.Store.AuthorizeClinicalRecord(ctx, a, "", encounterID, "billing"); err != nil {
		return domain.EncounterBilling{}, err
	}
	return s.Store.LinkEncounterInvoice(ctx, a, encounterID, invoiceID)
}

// ─── Discharge Summary ────────────────────────────────────────────────────────

func (s ClinicalCareService) DischargeSummary(ctx context.Context, a domain.Actor, encounterID string) (domain.DischargeSummary, error) {
	if !domain.UUIDPattern.MatchString(encounterID) {
		return domain.DischargeSummary{}, domain.ErrValidation
	}
	if err := s.Store.AuthorizeClinicalRecord(ctx, a, "", encounterID, "clinical"); err != nil {
		return domain.DischargeSummary{}, err
	}
	return s.Store.DischargeSummary(ctx, encounterID)
}

func (s ClinicalCareService) SaveDischargeSummary(ctx context.Context, a domain.Actor, summary domain.DischargeSummary) error {
	if a.Role != "admin" && a.Role != "doctor" {
		return domain.ErrForbidden
	}
	if err := summary.Validate(); err != nil {
		return err
	}
	summary.SignedBy = a.ID
	if err := s.Store.AuthorizeClinicalRecord(ctx, a, "", summary.EncounterID, "clinical"); err != nil {
		return err
	}
	return s.Store.SaveDischargeSummary(ctx, a, summary)
}

// ─── OPD Follow-ups & Referrals ───────────────────────────────────────────────

func (s ClinicalCareService) OPDFollowUps(ctx context.Context, a domain.Actor, patientID string) ([]domain.OPDFollowUp, error) {
	if !domain.UUIDPattern.MatchString(patientID) {
		return nil, domain.ErrValidation
	}
	if err := s.Store.AuthorizeClinicalRecord(ctx, a, patientID, "", "clinical"); err != nil {
		return nil, err
	}
	return s.Store.OPDFollowUps(ctx, patientID)
}

func (s ClinicalCareService) CreateOPDFollowUp(ctx context.Context, a domain.Actor, input domain.CreateOPDFollowUpInput) (domain.OPDFollowUp, error) {
	if a.Role != "admin" && a.Role != "doctor" {
		return domain.OPDFollowUp{}, domain.ErrForbidden
	}
	if err := input.Validate(); err != nil {
		return domain.OPDFollowUp{}, err
	}
	if err := s.Store.AuthorizeClinicalRecord(ctx, a, input.PatientID, input.EncounterID, "clinical"); err != nil {
		return domain.OPDFollowUp{}, err
	}
	return s.Store.CreateOPDFollowUp(ctx, a, input)
}

func (s ClinicalCareService) PatientReferrals(ctx context.Context, a domain.Actor, patientID string) ([]domain.PatientReferral, error) {
	if !domain.UUIDPattern.MatchString(patientID) {
		return nil, domain.ErrValidation
	}
	if err := s.Store.AuthorizeClinicalRecord(ctx, a, patientID, "", "clinical"); err != nil {
		return nil, err
	}
	return s.Store.PatientReferrals(ctx, patientID)
}

func (s ClinicalCareService) CreatePatientReferral(ctx context.Context, a domain.Actor, input domain.CreatePatientReferralInput) (domain.PatientReferral, error) {
	if a.Role != "admin" && a.Role != "doctor" {
		return domain.PatientReferral{}, domain.ErrForbidden
	}
	if err := input.Validate(); err != nil {
		return domain.PatientReferral{}, err
	}
	if err := s.Store.AuthorizeClinicalRecord(ctx, a, input.PatientID, input.EncounterID, "clinical"); err != nil {
		return domain.PatientReferral{}, err
	}
	return s.Store.CreatePatientReferral(ctx, a, input)
}

// ─── Odontogram ───────────────────────────────────────────────────────────────

func (s ClinicalCareService) Odontogram(ctx context.Context, a domain.Actor, patientID string) ([]domain.OdontogramEntry, error) {
	if !domain.UUIDPattern.MatchString(patientID) {
		return nil, domain.ErrValidation
	}
	if err := s.Store.AuthorizeClinicalRecord(ctx, a, patientID, "", "clinical"); err != nil {
		return nil, err
	}
	return s.Store.Odontogram(ctx, patientID)
}

func (s ClinicalCareService) SetToothCondition(ctx context.Context, a domain.Actor, patientID string, input domain.SetToothConditionInput) (domain.OdontogramEntry, error) {
	if a.Role != "admin" && a.Role != "doctor" {
		return domain.OdontogramEntry{}, domain.ErrForbidden
	}
	if !domain.UUIDPattern.MatchString(patientID) {
		return domain.OdontogramEntry{}, domain.ErrValidation
	}
	if err := input.Validate(); err != nil {
		return domain.OdontogramEntry{}, err
	}
	if err := s.Store.AuthorizeClinicalRecord(ctx, a, patientID, pointerValue(input.EncounterID), "clinical"); err != nil {
		return domain.OdontogramEntry{}, err
	}
	return s.Store.SetToothCondition(ctx, a, patientID, input)
}

func (s ClinicalCareService) CreateOdontogramChart(ctx context.Context, a domain.Actor, input domain.SaveOdontogramChartInput) (domain.OdontogramChart, error) {
	if a.Role != "admin" && a.Role != "doctor" {
		return domain.OdontogramChart{}, domain.ErrForbidden
	}
	if err := input.Validate(); err != nil {
		return domain.OdontogramChart{}, err
	}
	if err := s.Store.AuthorizeClinicalRecord(ctx, a, input.PatientID, pointerValue(input.EncounterID), "clinical"); err != nil {
		return domain.OdontogramChart{}, err
	}
	return s.Store.CreateOdontogramChart(ctx, a, input)
}

func (s ClinicalCareService) UpdateOdontogramChart(ctx context.Context, a domain.Actor, id string, input domain.SaveOdontogramChartInput) (domain.OdontogramChart, error) {
	if a.Role != "admin" && a.Role != "doctor" {
		return domain.OdontogramChart{}, domain.ErrForbidden
	}
	if !domain.UUIDPattern.MatchString(id) {
		return domain.OdontogramChart{}, domain.ErrValidation
	}
	if err := input.ValidateUpdate(); err != nil {
		return domain.OdontogramChart{}, err
	}
	existing, err := s.Store.GetOdontogramChart(ctx, id)
	if err != nil {
		return domain.OdontogramChart{}, err
	}
	if a.Role == "doctor" && existing.CreatedBy != a.ID && existing.DoctorID != a.ID {
		if err := s.Store.AuthorizeClinicalRecord(ctx, a, existing.PatientID, pointerValue(input.EncounterID), "clinical"); err != nil {
			return domain.OdontogramChart{}, err
		}
	}
	return s.Store.UpdateOdontogramChart(ctx, a, id, input)
}

func (s ClinicalCareService) GetOdontogramChart(ctx context.Context, a domain.Actor, id string) (domain.OdontogramChart, error) {
	if !domain.UUIDPattern.MatchString(id) {
		return domain.OdontogramChart{}, domain.ErrValidation
	}
	chart, err := s.Store.GetOdontogramChart(ctx, id)
	if err != nil {
		return domain.OdontogramChart{}, err
	}
	if a.Role == "patient" {
		if err := s.Store.AuthorizeClinicalRecord(ctx, a, chart.PatientID, pointerValue(chart.EncounterID), "odontogram_read"); err != nil {
			return domain.OdontogramChart{}, err
		}
	} else if a.Role == "doctor" {
		if chart.CreatedBy != a.ID && chart.DoctorID != a.ID {
			if err := s.Store.AuthorizeClinicalRecord(ctx, a, chart.PatientID, pointerValue(chart.EncounterID), "odontogram_read"); err != nil {
				return domain.OdontogramChart{}, err
			}
		}
	} else if a.Role != "admin" {
		return domain.OdontogramChart{}, domain.ErrForbidden
	}
	return chart, nil
}

func (s ClinicalCareService) ListOdontogramCharts(ctx context.Context, a domain.Actor, patientID, search string, page, limit int) ([]domain.OdontogramChart, int, error) {
	if a.Role != "admin" && a.Role != "doctor" && a.Role != "patient" {
		return nil, 0, domain.ErrForbidden
	}
	if patientID != "" && !domain.UUIDPattern.MatchString(patientID) {
		return nil, 0, domain.ErrValidation
	}
	if patientID != "" {
		if err := s.Store.AuthorizeClinicalRecord(ctx, a, patientID, "", "odontogram_read"); err != nil {
			return nil, 0, err
		}
	}
	return s.Store.ListOdontogramCharts(ctx, a, patientID, search, page, limit)
}

func (s ClinicalCareService) DeleteOdontogramChart(ctx context.Context, a domain.Actor, id string) error {
	if a.Role != "admin" && a.Role != "doctor" {
		return domain.ErrForbidden
	}
	if !domain.UUIDPattern.MatchString(id) {
		return domain.ErrValidation
	}
	chart, err := s.Store.GetOdontogramChart(ctx, id)
	if err != nil {
		return err
	}
	if a.Role == "doctor" && chart.CreatedBy != a.ID && chart.DoctorID != a.ID {
		if err := s.Store.AuthorizeClinicalRecord(ctx, a, chart.PatientID, pointerValue(chart.EncounterID), "clinical"); err != nil {
			return err
		}
	}
	return s.Store.DeleteOdontogramChart(ctx, id)
}

func (s ClinicalCareService) ListOdontogramLegends(ctx context.Context) ([]domain.OdontogramLegend, error) {
	return s.Store.ListOdontogramLegends(ctx)
}

func (s ClinicalCareService) CreateOdontogramLegend(ctx context.Context, a domain.Actor, input domain.SaveOdontogramLegendInput) (domain.OdontogramLegend, error) {
	if a.Role != "admin" && a.Role != "doctor" {
		return domain.OdontogramLegend{}, domain.ErrForbidden
	}
	if err := input.Validate(); err != nil {
		return domain.OdontogramLegend{}, err
	}
	return s.Store.CreateOdontogramLegend(ctx, a, input)
}

func (s ClinicalCareService) UpdateOdontogramLegend(ctx context.Context, a domain.Actor, id string, input domain.SaveOdontogramLegendInput) (domain.OdontogramLegend, error) {
	if a.Role != "admin" && a.Role != "doctor" {
		return domain.OdontogramLegend{}, domain.ErrForbidden
	}
	if id == "" {
		return domain.OdontogramLegend{}, domain.ErrValidation
	}
	if err := input.Validate(); err != nil {
		return domain.OdontogramLegend{}, err
	}
	return s.Store.UpdateOdontogramLegend(ctx, a, id, input)
}

func (s ClinicalCareService) DeleteOdontogramLegend(ctx context.Context, a domain.Actor, id string) error {
	if a.Role != "admin" && a.Role != "doctor" {
		return domain.ErrForbidden
	}
	if id == "" {
		return domain.ErrValidation
	}
	return s.Store.DeleteOdontogramLegend(ctx, a, id)
}

func pointerValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
