package domain

import (
	"strings"
	"time"
)

// ─── Diagnostic Categories & Units ───────────────────────────────────────────

type DiagnosticCategory struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Kind        string    `json:"kind"` // pathology, radiology
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
}

type DiagnosticCategoryInput struct {
	Name        string `json:"name"`
	Kind        string `json:"kind"`
	Description string `json:"description"`
}

func (i *DiagnosticCategoryInput) Validate() error {
	i.Name = strings.TrimSpace(i.Name)
	i.Kind = strings.TrimSpace(i.Kind)
	i.Description = strings.TrimSpace(i.Description)
	if len(i.Name) < 1 || len(i.Name) > 100 || (i.Kind != "pathology" && i.Kind != "radiology") || len(i.Description) > 1000 {
		return ErrValidation
	}
	return nil
}

type DiagnosticUnit struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
}

type DiagnosticUnitInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (i *DiagnosticUnitInput) Validate() error {
	i.Name = strings.TrimSpace(i.Name)
	i.Description = strings.TrimSpace(i.Description)
	if len(i.Name) < 1 || len(i.Name) > 50 || len(i.Description) > 500 {
		return ErrValidation
	}
	return nil
}

// ─── Diagnostic Report Files & Portal Release ─────────────────────────────────

type DiagnosticReportFile struct {
	ID              string     `json:"id"`
	OrderID         string     `json:"orderId"`
	FileName        string     `json:"fileName"`
	FileURL         string     `json:"fileUrl"`
	FileSizeBytes   int64      `json:"fileSizeBytes"`
	MimeType        string     `json:"mimeType"`
	PatientReleased bool       `json:"patientReleased"`
	ReleasedAt      *time.Time `json:"releasedAt,omitempty"`
	ReleasedBy      *string    `json:"releasedBy,omitempty"`
	UploadedBy      string     `json:"uploadedBy"`
	UploadedAt      time.Time  `json:"uploadedAt"`
}

type UploadDiagnosticReportFileInput struct {
	FileName      string `json:"fileName"`
	FileURL       string `json:"fileUrl"`
	FileSizeBytes int64  `json:"fileSizeBytes"`
	MimeType      string `json:"mimeType"`
}

func (i *UploadDiagnosticReportFileInput) Validate() error {
	i.FileName = strings.TrimSpace(i.FileName)
	i.FileURL = strings.TrimSpace(i.FileURL)
	i.MimeType = strings.TrimSpace(i.MimeType)
	if len(i.FileName) < 1 || len(i.FileName) > 255 || len(i.FileURL) < 1 || i.FileSizeBytes < 0 {
		return ErrValidation
	}
	if i.MimeType == "" {
		i.MimeType = "application/pdf"
	}
	return nil
}

// ─── Diagnosis Templates ──────────────────────────────────────────────────────

type DiagnosisTemplate struct {
	ID              string    `json:"id"`
	Title           string    `json:"title"`
	Category        string    `json:"category"`
	TemplateContent string    `json:"templateContent"`
	CreatedBy       string    `json:"createdBy"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

type DiagnosisTemplateInput struct {
	Title           string `json:"title"`
	Category        string `json:"category"`
	TemplateContent string `json:"templateContent"`
}

func (i *DiagnosisTemplateInput) Validate() error {
	i.Title = strings.TrimSpace(i.Title)
	i.Category = strings.TrimSpace(i.Category)
	i.TemplateContent = strings.TrimSpace(i.TemplateContent)
	if len([]rune(i.Title)) < 1 || len(i.Title) > 160 || len([]rune(i.TemplateContent)) < 1 {
		return ErrValidation
	}
	return nil
}

// ─── Vaccines & Vaccinations ──────────────────────────────────────────────────

type Vaccine struct {
	ID                string    `json:"id"`
	Name              string    `json:"name"`
	TargetDisease     string    `json:"targetDisease"`
	RecommendedDoses  int       `json:"recommendedDoses"`
	MinAgeMonths      int       `json:"minAgeMonths"`
	Instructions      string    `json:"instructions"`
	CreatedAt         time.Time `json:"createdAt"`
}

type VaccineInput struct {
	Name             string `json:"name"`
	TargetDisease    string `json:"targetDisease"`
	RecommendedDoses int    `json:"recommendedDoses"`
	MinAgeMonths     int    `json:"minAgeMonths"`
	Instructions     string `json:"instructions"`
}

func (i *VaccineInput) Validate() error {
	i.Name = strings.TrimSpace(i.Name)
	i.TargetDisease = strings.TrimSpace(i.TargetDisease)
	i.Instructions = strings.TrimSpace(i.Instructions)
	if len([]rune(i.Name)) < 1 || len(i.Name) > 160 || i.RecommendedDoses < 1 || i.MinAgeMonths < 0 {
		return ErrValidation
	}
	return nil
}

type PatientVaccination struct {
	ID             string     `json:"id"`
	PatientID      string     `json:"patientId"`
	PatientName    string     `json:"patientName,omitempty"`
	VaccineID      string     `json:"vaccineId"`
	VaccineName    string     `json:"vaccineName,omitempty"`
	DoseNumber     int        `json:"doseNumber"`
	LotNumber      string     `json:"lotNumber"`
	ExpiryDate     *string    `json:"expiryDate,omitempty"`
	AdministeredAt time.Time  `json:"administeredAt"`
	AdministeredBy string     `json:"administeredBy"`
	NextDueDate    *string    `json:"nextDueDate,omitempty"`
	Notes          string     `json:"notes"`
	CreatedAt      time.Time  `json:"createdAt"`
}

type AdministerVaccineInput struct {
	PatientID   string  `json:"patientId"`
	VaccineID   string  `json:"vaccineId"`
	DoseNumber  int     `json:"doseNumber"`
	LotNumber   string  `json:"lotNumber"`
	ExpiryDate  *string `json:"expiryDate,omitempty"`
	NextDueDate *string `json:"nextDueDate,omitempty"`
	Notes       string  `json:"notes"`
}

func (i *AdministerVaccineInput) Validate() error {
	i.LotNumber = strings.TrimSpace(i.LotNumber)
	i.Notes = strings.TrimSpace(i.Notes)
	if !UUIDPattern.MatchString(i.PatientID) || !UUIDPattern.MatchString(i.VaccineID) || i.DoseNumber < 1 {
		return ErrValidation
	}
	if i.ExpiryDate != nil && *i.ExpiryDate != "" {
		if _, err := time.Parse("2006-01-02", *i.ExpiryDate); err != nil {
			return ErrValidation
		}
	}
	if i.NextDueDate != nil && *i.NextDueDate != "" {
		if _, err := time.Parse("2006-01-02", *i.NextDueDate); err != nil {
			return ErrValidation
		}
	}
	return nil
}

// ─── Birth, Death, Operation, and Investigation Vital Reports ─────────────────

type BirthReport struct {
	ID              string    `json:"id"`
	ReportNumber    string    `json:"reportNumber"`
	ChildName       string    `json:"childName"`
	Gender          string    `json:"gender"`
	BirthDate       time.Time `json:"birthDate"`
	WeightKg        float64   `json:"weightKg"`
	MotherPatientID *string   `json:"motherPatientId,omitempty"`
	MotherName      string    `json:"motherName"`
	FatherName      string    `json:"fatherName"`
	DeliveredBy     string    `json:"deliveredBy"`
	DeliveredByName string    `json:"deliveredByName,omitempty"`
	Notes           string    `json:"notes"`
	CreatedBy       string    `json:"createdBy"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

type CreateBirthReportInput struct {
	ChildName       string    `json:"childName"`
	Gender          string    `json:"gender"`
	BirthDate       time.Time `json:"birthDate"`
	WeightKg        float64   `json:"weightKg"`
	MotherPatientID *string   `json:"motherPatientId,omitempty"`
	MotherName      string    `json:"motherName"`
	FatherName      string    `json:"fatherName"`
	DeliveredBy     string    `json:"deliveredBy"`
	Notes           string    `json:"notes"`
}

func (i *CreateBirthReportInput) Validate() error {
	i.ChildName = strings.TrimSpace(i.ChildName)
	i.Gender = strings.TrimSpace(i.Gender)
	i.MotherName = strings.TrimSpace(i.MotherName)
	i.FatherName = strings.TrimSpace(i.FatherName)
	i.DeliveredBy = strings.TrimSpace(i.DeliveredBy)
	if len([]rune(i.ChildName)) < 1 || len(i.ChildName) > 160 || len([]rune(i.MotherName)) < 1 || len(i.MotherName) > 160 {
		return ErrValidation
	}
	if i.Gender != "male" && i.Gender != "female" && i.Gender != "other" {
		return ErrValidation
	}
	if i.WeightKg <= 0 || i.WeightKg > 15 || i.DeliveredBy == "" {
		return ErrValidation
	}
	if i.MotherPatientID != nil && *i.MotherPatientID != "" && !UUIDPattern.MatchString(*i.MotherPatientID) {
		return ErrValidation
	}
	return nil
}

type DeathReport struct {
	ID                   string    `json:"id"`
	ReportNumber         string    `json:"reportNumber"`
	PatientID            string    `json:"patientId"`
	PatientName          string    `json:"patientName,omitempty"`
	DeathDate            time.Time `json:"deathDate"`
	CauseOfDeath         string    `json:"causeOfDeath"`
	CertifiedBy          string    `json:"certifiedBy"`
	CertifiedByName      string    `json:"certifiedByName,omitempty"`
	GuardianAcknowledged string    `json:"guardianAcknowledged"`
	Notes                string    `json:"notes"`
	CreatedBy            string    `json:"createdBy"`
	CreatedAt            time.Time `json:"createdAt"`
	UpdatedAt            time.Time `json:"updatedAt"`
}

type CreateDeathReportInput struct {
	PatientID            string    `json:"patientId"`
	DeathDate            time.Time `json:"deathDate"`
	CauseOfDeath         string    `json:"causeOfDeath"`
	CertifiedBy          string    `json:"certifiedBy"`
	GuardianAcknowledged string    `json:"guardianAcknowledged"`
	Notes                string    `json:"notes"`
}

func (i *CreateDeathReportInput) Validate() error {
	i.CauseOfDeath = strings.TrimSpace(i.CauseOfDeath)
	i.CertifiedBy = strings.TrimSpace(i.CertifiedBy)
	i.GuardianAcknowledged = strings.TrimSpace(i.GuardianAcknowledged)
	if !UUIDPattern.MatchString(i.PatientID) || len([]rune(i.CauseOfDeath)) < 1 || len(i.CauseOfDeath) > 2000 || i.CertifiedBy == "" {
		return ErrValidation
	}
	return nil
}

type OperationReport struct {
	ID                     string    `json:"id"`
	ReportNumber           string    `json:"reportNumber"`
	EncounterID            string    `json:"encounterId"`
	PatientID              string    `json:"patientId"`
	PatientName            string    `json:"patientName,omitempty"`
	OperationName          string    `json:"operationName"`
	SurgeonID              string    `json:"surgeonId"`
	SurgeonName            string    `json:"surgeonName,omitempty"`
	AssistantSurgeon       string    `json:"assistantSurgeon"`
	Anesthetist            string    `json:"anesthetist"`
	AnesthesiaType         string    `json:"anesthesiaType"`
	OperationDate          time.Time `json:"operationDate"`
	PreOperativeDiagnosis  string    `json:"preOperativeDiagnosis"`
	PostOperativeDiagnosis string    `json:"postOperativeDiagnosis"`
	ProcedureTechnique     string    `json:"procedureTechnique"`
	Findings               string    `json:"findings"`
	Complications          string    `json:"complications"`
	CreatedBy              string    `json:"createdBy"`
	CreatedAt              time.Time `json:"createdAt"`
	UpdatedAt              time.Time `json:"updatedAt"`
}

type CreateOperationReportInput struct {
	EncounterID            string    `json:"encounterId"`
	PatientID              string    `json:"patientId"`
	OperationName          string    `json:"operationName"`
	SurgeonID              string    `json:"surgeonId"`
	AssistantSurgeon       string    `json:"assistantSurgeon"`
	Anesthetist            string    `json:"anesthetist"`
	AnesthesiaType         string    `json:"anesthesiaType"`
	OperationDate          time.Time `json:"operationDate"`
	PreOperativeDiagnosis  string    `json:"preOperativeDiagnosis"`
	PostOperativeDiagnosis string    `json:"postOperativeDiagnosis"`
	ProcedureTechnique     string    `json:"procedureTechnique"`
	Findings               string    `json:"findings"`
	Complications          string    `json:"complications"`
}

func (i *CreateOperationReportInput) Validate() error {
	i.OperationName = strings.TrimSpace(i.OperationName)
	i.SurgeonID = strings.TrimSpace(i.SurgeonID)
	if !UUIDPattern.MatchString(i.EncounterID) || !UUIDPattern.MatchString(i.PatientID) ||
		len([]rune(i.OperationName)) < 1 || len(i.OperationName) > 160 || i.SurgeonID == "" {
		return ErrValidation
	}
	return nil
}

type InvestigationReport struct {
	ID                string    `json:"id"`
	ReportNumber      string    `json:"reportNumber"`
	PatientID         string    `json:"patientId"`
	PatientName       string    `json:"patientName,omitempty"`
	EncounterID       *string   `json:"encounterId,omitempty"`
	Title             string    `json:"title"`
	InvestigationType string    `json:"investigationType"`
	ClinicalNotes     string    `json:"clinicalNotes"`
	Conclusion        string    `json:"conclusion"`
	InvestigatedBy    string    `json:"investigatedBy"`
	InvestigatedByName string   `json:"investigatedByName,omitempty"`
	CreatedAt         time.Time `json:"createdAt"`
}

type CreateInvestigationReportInput struct {
	PatientID         string  `json:"patientId"`
	EncounterID       *string `json:"encounterId,omitempty"`
	Title             string  `json:"title"`
	InvestigationType string  `json:"investigationType"`
	ClinicalNotes     string  `json:"clinicalNotes"`
	Conclusion        string  `json:"conclusion"`
	InvestigatedBy    string  `json:"investigatedBy"`
}

func (i *CreateInvestigationReportInput) Validate() error {
	i.Title = strings.TrimSpace(i.Title)
	i.InvestigationType = strings.TrimSpace(i.InvestigationType)
	i.InvestigatedBy = strings.TrimSpace(i.InvestigatedBy)
	if !UUIDPattern.MatchString(i.PatientID) || len([]rune(i.Title)) < 1 || len(i.Title) > 160 ||
		len([]rune(i.InvestigationType)) < 1 || i.InvestigatedBy == "" {
		return ErrValidation
	}
	if i.EncounterID != nil && *i.EncounterID != "" && !UUIDPattern.MatchString(*i.EncounterID) {
		return ErrValidation
	}
	return nil
}
