package domain

import (
	"strings"
	"time"
)

// ─── Bed Assignment & Occupancy Report ────────────────────────────────────────

type BedAssignment struct {
	ID           string     `json:"id"`
	BedID        string     `json:"bedId"`
	BedName      string     `json:"bedName,omitempty"`
	BedType      string     `json:"bedType,omitempty"`
	EncounterID  string     `json:"encounterId"`
	PatientID    string     `json:"patientId"`
	PatientName  string     `json:"patientName,omitempty"`
	AssignedFrom time.Time  `json:"assignedFrom"`
	AssignedTo   *time.Time `json:"assignedTo,omitempty"`
	Notes        string     `json:"notes"`
	AssignedBy   string     `json:"assignedBy"`
	CreatedAt    time.Time  `json:"createdAt"`
}

type BedOccupancyReport struct {
	TotalBeds        int     `json:"totalBeds"`
	OccupiedBeds     int     `json:"occupiedBeds"`
	AvailableBeds    int     `json:"availableBeds"`
	OccupancyRate    float64 `json:"occupancyRate"`
	ActiveAdmissions int     `json:"activeAdmissions"`
}

type AssignBedInput struct {
	Version     int    `json:"version"`
	BedID       string `json:"bedId"`
	EncounterID string `json:"encounterId"`
	PatientID   string `json:"patientId"`
	Notes       string `json:"notes"`
}

func (i *AssignBedInput) Validate() error {
	i.Notes = strings.TrimSpace(i.Notes)
	if i.Version < 1 || !UUIDPattern.MatchString(i.BedID) || !UUIDPattern.MatchString(i.EncounterID) || !UUIDPattern.MatchString(i.PatientID) {
		return ErrValidation
	}
	if len(i.Notes) > 1000 {
		return ErrValidation
	}
	return nil
}

// ─── Care Team Delegation ─────────────────────────────────────────────────────

type CareTeamMember struct {
	ID          string     `json:"id"`
	EncounterID string     `json:"encounterId"`
	StaffID     string     `json:"staffId"`
	StaffName   string     `json:"staffName,omitempty"`
	RoleTitle   string     `json:"roleTitle"`
	AssignedAt  time.Time  `json:"assignedAt"`
	RevokedAt   *time.Time `json:"revokedAt,omitempty"`
	AssignedBy  string     `json:"assignedBy"`
	Notes       string     `json:"notes"`
}

type AddCareTeamMemberInput struct {
	StaffID   string `json:"staffId"`
	RoleTitle string `json:"roleTitle"`
	Notes     string `json:"notes"`
}

func (i *AddCareTeamMemberInput) Validate() error {
	i.StaffID = strings.TrimSpace(i.StaffID)
	i.RoleTitle = strings.TrimSpace(i.RoleTitle)
	i.Notes = strings.TrimSpace(i.Notes)
	if i.StaffID == "" || len(i.RoleTitle) < 1 || len(i.RoleTitle) > 80 || len(i.Notes) > 500 {
		return ErrValidation
	}
	return nil
}

// ─── Consultation Registers: Diagnoses, Procedures, Attachments ──────────────

type EncounterDiagnosis struct {
	ID          string    `json:"id"`
	EncounterID string    `json:"encounterId"`
	ICD10Code   string    `json:"icd10Code"`
	Description string    `json:"description"`
	Category    string    `json:"category"` // provisional, final, differential
	Status      string    `json:"status"`   // active, resolved, ruled_out
	DiagnosedBy string    `json:"diagnosedBy"`
	DiagnosedAt time.Time `json:"diagnosedAt"`
	CreatedAt   time.Time `json:"createdAt"`
}

type AddDiagnosisInput struct {
	ICD10Code   string `json:"icd10Code"`
	Description string `json:"description"`
	Category    string `json:"category"`
	Status      string `json:"status"`
}

func (i *AddDiagnosisInput) Validate() error {
	i.ICD10Code = strings.TrimSpace(i.ICD10Code)
	i.Description = strings.TrimSpace(i.Description)
	i.Category = strings.TrimSpace(i.Category)
	i.Status = strings.TrimSpace(i.Status)
	if len([]rune(i.Description)) < 1 || len(i.Description) > 2000 {
		return ErrValidation
	}
	switch i.Category {
	case "provisional", "final", "differential":
	default:
		i.Category = "provisional"
	}
	switch i.Status {
	case "active", "resolved", "ruled_out":
	default:
		i.Status = "active"
	}
	return nil
}

type EncounterProcedure struct {
	ID              string    `json:"id"`
	EncounterID     string    `json:"encounterId"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	PerformedBy     string    `json:"performedBy"`
	PerformedByName string    `json:"performedByName,omitempty"`
	PerformedAt     time.Time `json:"performedAt"`
	AnesthesiaType  string    `json:"anesthesiaType"`
	Findings        string    `json:"findings"`
	Complications   string    `json:"complications"`
	CreatedAt       time.Time `json:"createdAt"`
}

type AddProcedureInput struct {
	Name           string `json:"name"`
	Description    string `json:"description"`
	AnesthesiaType string `json:"anesthesiaType"`
	Findings       string `json:"findings"`
	Complications  string `json:"complications"`
}

func (i *AddProcedureInput) Validate() error {
	i.Name = strings.TrimSpace(i.Name)
	i.Description = strings.TrimSpace(i.Description)
	i.AnesthesiaType = strings.TrimSpace(i.AnesthesiaType)
	i.Findings = strings.TrimSpace(i.Findings)
	i.Complications = strings.TrimSpace(i.Complications)
	if len([]rune(i.Name)) < 1 || len(i.Name) > 160 || len(i.Description) > 2000 ||
		len(i.Findings) > 4000 || len(i.Complications) > 2000 {
		return ErrValidation
	}
	return nil
}

type EncounterAttachment struct {
	ID            string    `json:"id"`
	EncounterID   string    `json:"encounterId"`
	Title         string    `json:"title"`
	FileURL       string    `json:"fileUrl"`
	FileType      string    `json:"fileType"`
	FileSizeBytes int64     `json:"fileSizeBytes"`
	UploadedBy    string    `json:"uploadedBy"`
	UploadedAt    time.Time `json:"uploadedAt"`
}

type AddAttachmentInput struct {
	Title         string `json:"title"`
	FileURL       string `json:"fileUrl"`
	FileType      string `json:"fileType"`
	FileSizeBytes int64  `json:"fileSizeBytes"`
}

func (i *AddAttachmentInput) Validate() error {
	i.Title = strings.TrimSpace(i.Title)
	i.FileURL = strings.TrimSpace(i.FileURL)
	i.FileType = strings.TrimSpace(i.FileType)
	if len([]rune(i.Title)) < 1 || len(i.Title) > 160 || len(i.FileURL) < 1 || len(i.FileURL) > 1000 || i.FileSizeBytes < 0 {
		return ErrValidation
	}
	if i.FileType == "" {
		i.FileType = "document"
	}
	return nil
}

// ─── IPD Admission Packages & Insurance ───────────────────────────────────────

type IPDAdmissionDetails struct {
	EncounterID           string    `json:"encounterId"`
	PackageName           string    `json:"packageName"`
	PackageChargeMinor    int64     `json:"packageChargeMinor"`
	InsurancePolicyNumber string    `json:"insurancePolicyNumber"`
	InsuranceProvider     string    `json:"insuranceProvider"`
	GuardianName          string    `json:"guardianName"`
	GuardianRelation      string    `json:"guardianRelation"`
	GuardianPhone         string    `json:"guardianPhone"`
	GuardianAddress       string    `json:"guardianAddress"`
	CreatedAt             time.Time `json:"createdAt"`
	UpdatedAt             time.Time `json:"updatedAt"`
}

func (d *IPDAdmissionDetails) Validate() error {
	d.PackageName = strings.TrimSpace(d.PackageName)
	d.InsurancePolicyNumber = strings.TrimSpace(d.InsurancePolicyNumber)
	d.InsuranceProvider = strings.TrimSpace(d.InsuranceProvider)
	d.GuardianName = strings.TrimSpace(d.GuardianName)
	d.GuardianRelation = strings.TrimSpace(d.GuardianRelation)
	d.GuardianPhone = strings.TrimSpace(d.GuardianPhone)
	d.GuardianAddress = strings.TrimSpace(d.GuardianAddress)
	if !UUIDPattern.MatchString(d.EncounterID) || d.PackageChargeMinor < 0 {
		return ErrValidation
	}
	return nil
}

// ─── Encounter Billing & Financial Clearance ──────────────────────────────────

type EncounterBilling struct {
	EncounterID        string     `json:"encounterId"`
	BedDays            int        `json:"bedDays"`
	BedTotalMinor      int64      `json:"bedTotalMinor"`
	DoctorFeeMinor     int64      `json:"doctorFeeMinor"`
	ProcedureFeeMinor  int64      `json:"procedureFeeMinor"`
	OtherChargesMinor  int64      `json:"otherChargesMinor"`
	TotalMinor         int64      `json:"totalMinor"`
	InvoiceID          *string    `json:"invoiceId,omitempty"`
	FinancialClearance bool       `json:"financialClearance"`
	ClearedBy          *string    `json:"clearedBy,omitempty"`
	ClearedAt          *time.Time `json:"clearedAt,omitempty"`
	WaiverReason       string     `json:"waiverReason"`
	CreatedAt          time.Time  `json:"createdAt"`
	UpdatedAt          time.Time  `json:"updatedAt"`
}

type UpdateEncounterBillingInput struct {
	BedDays           int   `json:"bedDays"`
	BedTotalMinor     int64 `json:"bedTotalMinor"`
	DoctorFeeMinor    int64 `json:"doctorFeeMinor"`
	ProcedureFeeMinor int64 `json:"procedureFeeMinor"`
	OtherChargesMinor int64 `json:"otherChargesMinor"`
}

func (i *UpdateEncounterBillingInput) Validate() error {
	if i.BedDays > 36500 || i.BedTotalMinor > 100000000000 || i.DoctorFeeMinor > 100000000000 || i.ProcedureFeeMinor > 100000000000 || i.OtherChargesMinor > 100000000000 || i.BedDays < 0 || i.BedTotalMinor < 0 || i.DoctorFeeMinor < 0 || i.ProcedureFeeMinor < 0 || i.OtherChargesMinor < 0 {
		return ErrValidation
	}
	return nil
}

// ─── Structured Discharge Summary ─────────────────────────────────────────────

type DischargeSummary struct {
	EncounterID          string    `json:"encounterId"`
	AdmissionDiagnosis   string    `json:"admissionDiagnosis"`
	DischargeDiagnosis   string    `json:"dischargeDiagnosis"`
	ConditionAtDischarge string    `json:"conditionAtDischarge"` // recovered, improved, unchanged, referred, deceased
	HospitalCourse       string    `json:"hospitalCourse"`
	SurgicalProcedures   string    `json:"surgicalProcedures"`
	DischargeMedications string    `json:"dischargeMedications"`
	FollowUpAdvice       string    `json:"followUpAdvice"`
	FollowUpDate         *string   `json:"followUpDate,omitempty"`
	SignedBy             string    `json:"signedBy"`
	SignedByName         string    `json:"signedByName,omitempty"`
	SignedAt             time.Time `json:"signedAt"`
	CreatedAt            time.Time `json:"createdAt"`
	UpdatedAt            time.Time `json:"updatedAt"`
}

func (s *DischargeSummary) Validate() error {
	s.AdmissionDiagnosis = strings.TrimSpace(s.AdmissionDiagnosis)
	s.DischargeDiagnosis = strings.TrimSpace(s.DischargeDiagnosis)
	s.HospitalCourse = strings.TrimSpace(s.HospitalCourse)
	s.SurgicalProcedures = strings.TrimSpace(s.SurgicalProcedures)
	s.DischargeMedications = strings.TrimSpace(s.DischargeMedications)
	s.FollowUpAdvice = strings.TrimSpace(s.FollowUpAdvice)
	if !UUIDPattern.MatchString(s.EncounterID) {
		return ErrValidation
	}
	switch s.ConditionAtDischarge {
	case "recovered", "improved", "unchanged", "referred", "deceased":
	default:
		s.ConditionAtDischarge = "improved"
	}
	if s.FollowUpDate != nil && *s.FollowUpDate != "" {
		if _, err := time.Parse("2006-01-02", *s.FollowUpDate); err != nil {
			return ErrValidation
		}
	}
	return nil
}

// ─── OPD Follow-ups & Referrals ───────────────────────────────────────────────

type OPDFollowUp struct {
	ID           string    `json:"id"`
	EncounterID  string    `json:"encounterId"`
	PatientID    string    `json:"patientId"`
	PatientName  string    `json:"patientName,omitempty"`
	DoctorID     string    `json:"doctorId"`
	DoctorName   string    `json:"doctorName,omitempty"`
	FollowUpDate string    `json:"followUpDate"` // YYYY-MM-DD
	Notes        string    `json:"notes"`
	Status       string    `json:"status"` // scheduled, attended, cancelled
	CreatedAt    time.Time `json:"createdAt"`
}

type CreateOPDFollowUpInput struct {
	EncounterID  string `json:"encounterId"`
	PatientID    string `json:"patientId"`
	DoctorID     string `json:"doctorId"`
	FollowUpDate string `json:"followUpDate"`
	Notes        string `json:"notes"`
}

func (i *CreateOPDFollowUpInput) Validate() error {
	i.DoctorID = strings.TrimSpace(i.DoctorID)
	i.FollowUpDate = strings.TrimSpace(i.FollowUpDate)
	i.Notes = strings.TrimSpace(i.Notes)
	if !UUIDPattern.MatchString(i.EncounterID) || !UUIDPattern.MatchString(i.PatientID) || i.DoctorID == "" {
		return ErrValidation
	}
	if _, err := time.Parse("2006-01-02", i.FollowUpDate); err != nil {
		return ErrValidation
	}
	if len(i.Notes) > 2000 {
		return ErrValidation
	}
	return nil
}

type PatientReferral struct {
	ID               string    `json:"id"`
	EncounterID      string    `json:"encounterId"`
	PatientID        string    `json:"patientId"`
	PatientName      string    `json:"patientName,omitempty"`
	ReferralType     string    `json:"referralType"` // inward, outward
	ExternalFacility string    `json:"externalFacility"`
	Department       string    `json:"department"`
	Reason           string    `json:"reason"`
	ReferredBy       string    `json:"referredBy"`
	ReferredAt       time.Time `json:"referredAt"`
}

type CreatePatientReferralInput struct {
	EncounterID      string `json:"encounterId"`
	PatientID        string `json:"patientId"`
	ReferralType     string `json:"referralType"`
	ExternalFacility string `json:"externalFacility"`
	Department       string `json:"department"`
	Reason           string `json:"reason"`
}

func (i *CreatePatientReferralInput) Validate() error {
	i.ReferralType = strings.TrimSpace(i.ReferralType)
	i.ExternalFacility = strings.TrimSpace(i.ExternalFacility)
	i.Department = strings.TrimSpace(i.Department)
	i.Reason = strings.TrimSpace(i.Reason)
	if !UUIDPattern.MatchString(i.EncounterID) || !UUIDPattern.MatchString(i.PatientID) {
		return ErrValidation
	}
	if i.ReferralType != "inward" && i.ReferralType != "outward" {
		i.ReferralType = "outward"
	}
	if len([]rune(i.ExternalFacility)) < 1 || len(i.ExternalFacility) > 200 || len([]rune(i.Reason)) < 1 || len(i.Reason) > 2000 {
		return ErrValidation
	}
	return nil
}

// ─── Odontogram (Dental Chart) ────────────────────────────────────────────────

type OdontogramEntry struct {
	ID             string    `json:"id"`
	PatientID      string    `json:"patientId"`
	EncounterID    *string   `json:"encounterId,omitempty"`
	ToothNumber    int       `json:"toothNumber"`
	Condition      string    `json:"condition"` // healthy, caries, missing, filled, crown, implant, extracted, root_canal
	ProcedureNotes string    `json:"procedureNotes"`
	DiagnosedBy    string    `json:"diagnosedBy"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

type SetToothConditionInput struct {
	EncounterID    *string `json:"encounterId,omitempty"`
	ToothNumber    int     `json:"toothNumber"`
	Condition      string  `json:"condition"`
	ProcedureNotes string  `json:"procedureNotes"`
}

func (i *SetToothConditionInput) Validate() error {
	i.Condition = strings.TrimSpace(i.Condition)
	i.ProcedureNotes = strings.TrimSpace(i.ProcedureNotes)
	if i.ToothNumber < 1 || i.ToothNumber > 52 {
		return ErrValidation
	}
	switch i.Condition {
	case "healthy", "caries", "missing", "filled", "crown", "implant", "extracted", "root_canal":
	default:
		return ErrValidation
	}
	if i.EncounterID != nil && *i.EncounterID != "" && !UUIDPattern.MatchString(*i.EncounterID) {
		return ErrValidation
	}
	if len(i.ProcedureNotes) > 1000 {
		return ErrValidation
	}
	return nil
}
